package com.avatar.app.data

import com.avatar.app.BuildConfig
import com.avatar.app.data.entities.LocalLot
import com.avatar.app.data.entities.LocalMarket
import com.avatar.app.data.entities.LocalPrediction
import com.avatar.app.data.entities.LocalProduct
import com.avatar.app.data.entities.LocalSale
import com.avatar.app.data.entities.LocalSaleLine
import com.avatar.app.data.entities.Outbox
import androidx.room.withTransaction
import java.io.IOException
import java.net.Inet4Address
import java.net.NetworkInterface
import java.time.Instant
import java.time.LocalDateTime
import java.util.concurrent.TimeUnit
import java.util.concurrent.atomic.AtomicReference
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.async
import kotlinx.coroutines.awaitAll
import kotlinx.coroutines.coroutineScope
import kotlinx.coroutines.launch
import kotlinx.coroutines.sync.Semaphore
import kotlinx.coroutines.sync.withPermit
import kotlinx.coroutines.withContext
import kotlinx.coroutines.withTimeoutOrNull
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody.Companion.toRequestBody
import okhttp3.Response
import org.json.JSONArray
import org.json.JSONObject

sealed class PushResult {
    object Success : PushResult()
    data class Blocked(val message: String) : PushResult()
    data class InsufficientStock(val message: String) : PushResult()
    data class Failed(val message: String) : PushResult()
}

data class PredictionView(
    val rows: List<Pair<String, String>>,
    val warnings: List<String>,
    val modelVersion: String,
    val modelStatus: String,
    val offline: Boolean = false,
    val message: String? = null,
)

data class ReportSection(
    val title: String,
    val count: Int,
    val columns: List<String>,
    val rows: List<List<String>>,
)

data class ReportView(
    val generatedAt: String,
    val sections: List<ReportSection>,
    val offline: Boolean = false,
    val message: String? = null,
)

private val PREDICTION_LABELS = mapOf(
    "finished_qty" to "Beklenen mamul (kg)",
    "yield_pct" to "Verim (%)",
    "estimated_ekp" to "Tahmini EKP (€)",
    "leg_weight" to "But (kg)",
    "loin_weight" to "Sırt (kg)",
    "shoulder_weight" to "Omuz (kg)",
    "rib_weight" to "Kaburga (kg)",
    "mince_weight" to "Kıyma (kg)",
    "diced_meat_weight" to "Kuşbaşı (kg)",
    "bone_weight" to "Kemik (kg)",
    "waste_weight" to "Fire (kg)",
    "total_saleable_weight" to "Toplam satılabilir (kg)",
    "cutting_yield_pct" to "Parçalama verimi (%)",
)

class SyncRepository(
    private val db: AppDatabase,
    private val client: OkHttpClient = OkHttpClient(),
    baseUrl: String = DEFAULT_BASE_URL,
) {
    private var configuredBaseUrl: String = baseUrl

    companion object {
        // Fiziksel cihaz varsayılanı: APK derlenirken tespit edilen PC LAN IP'si
        // (Gradle -> BuildConfig.DEFAULT_BASE_URL). Böylece telefon aynı Wi-Fi'da
        // olduğunda ekstra ayar gerekmez.
        // Yedekler: emülatörde 10.0.2.2, USB + `adb reverse` senaryosunda localhost.
        val DEFAULT_BASE_URL: String = BuildConfig.DEFAULT_BASE_URL
        const val PORT = 8080
        private const val CLIENT_ID = "android"
        private val JSON = "application/json; charset=utf-8".toMediaType()
    }

    private var resolvedBaseUrl: String? = null
    private val healthClient: OkHttpClient =
        client.newBuilder().callTimeout(2, TimeUnit.SECONDS).build()

    fun activeBaseUrl(): String? = resolvedBaseUrl

    // Persist the user's choice from the UI and force a fresh health probe.
    fun setBaseUrl(url: String) {
        configuredBaseUrl = normalizeBaseUrl(url) ?: url.trim()
        resolvedBaseUrl = null
    }

    // "Dene" butonu ve başarısız istekler için çözümlenmiş adresi sıfırla.
    fun invalidateConnection() {
        resolvedBaseUrl = null
    }

    private fun normalizeBaseUrl(url: String): String? {
        val trimmed = url.trim()
        if (trimmed.isEmpty()) return null
        val withScheme = when {
            trimmed.startsWith("http://") || trimmed.startsWith("https://") -> trimmed
            else -> "http://$trimmed"
        }
        return withScheme.trimEnd('/')
    }

    private fun candidateBaseUrls(): List<String> = buildList {
        normalizeBaseUrl(configuredBaseUrl)?.let { add(it) }
        normalizeBaseUrl(DEFAULT_BASE_URL)?.let { add(it) }
        // Emülatör: host makinenin localhost'u.
        add("http://10.0.2.2:$PORT")
        // Fiziksel cihaz + USB: `adb reverse tcp:8080 tcp:8080`.
        add("http://127.0.0.1:$PORT")
        add("http://localhost:$PORT")
    }.distinct()

    // Resolve the first reachable backend once and cache it. Bilinen adresler
    // paralel denenir (USB/adb reverse, Wi-Fi LAN IP ve emülatör aynı anda),
    // böylece biri zaman aşımına uğrasa bile açılış beklemez. Hiçbiri
    // cevaplamazsa yerel ağ /24 taraması yapılır.
    private suspend fun serverBaseUrl(): String {
        resolvedBaseUrl?.let { return it }
        val candidates = candidateBaseUrls()
        val failures = mutableListOf<String>()

        val results = withContext(Dispatchers.IO) {
            coroutineScope {
                candidates.map { candidate ->
                    async { candidate to probeHealth(candidate) }
                }.awaitAll()
            }
        }
        for ((candidate, failure) in results) {
            if (failure == null) {
                resolvedBaseUrl = candidate
                return candidate
            }
            failures.add("$candidate -> $failure")
        }

        val discovered = discoverLanServer()
        if (discovered != null) {
            resolvedBaseUrl = discovered
            return discovered
        }
        failures.add("yerel ağ taraması -> bulunamadı")

        throw IOException(
            "Sunucuya ulaşılamadı (port $PORT). Backend çalışıyor mu? " +
                "Telefon ve PC aynı Wi-Fi'da mı? Windows güvenlik duvarı $PORT portuna izin veriyor mu? " +
                "USB kullanıyorsan `adb reverse tcp:$PORT tcp:$PORT` çalıştır. " +
                "Denenen adresler: ${failures.joinToString("; ")}",
        )
    }

    // Returns null when /health is reachable, otherwise a short failure reason.
    private fun probeHealth(baseUrl: String, httpClient: OkHttpClient = healthClient): String? =
        try {
            get("$baseUrl/health", httpClient).use { response ->
                if (response.isSuccessful) null else "HTTP ${response.code}"
            }
        } catch (error: Exception) {
            error.message ?: error.javaClass.simpleName
        }

    // Known addresses tutmazsa telefonun bulunduğu /24 ağını tarar. Yalnızca
    // son çare olarak çalışır ve ilk cevap veren backend'i döndürür.
    private suspend fun discoverLanServer(): String? = withContext(Dispatchers.IO) {
        val hosts = lanProbeHosts()
        if (hosts.isEmpty()) return@withContext null

        val probeClient = client.newBuilder()
            .connectTimeout(500, TimeUnit.MILLISECONDS)
            .readTimeout(700, TimeUnit.MILLISECONDS)
            .callTimeout(1200, TimeUnit.MILLISECONDS)
            .retryOnConnectionFailure(false)
            .build()
        val found = AtomicReference<String?>(null)
        val semaphore = Semaphore(64)

        // Toplam tarama süresini sınırla: bağlantı yokken kullanıcıyı
        // saniyelerce bekletmeyelim. Bulunan ilk backend yeterli.
        withTimeoutOrNull(5_000) {
            coroutineScope {
                hosts.forEach { host ->
                    launch {
                        semaphore.withPermit {
                            probeCandidate(host, probeClient, found)
                        }
                    }
                }
            }
        }
        found.get()
    }

    private fun probeCandidate(host: String, probeClient: OkHttpClient, found: AtomicReference<String?>) {
    if (found.get() != null) return
    val url = "http://$host:$PORT"
    if (probeHealth(url, probeClient) == null) {
        found.compareAndSet(null, url)
    }
    }

    private fun lanProbeHosts(): List<String> {
        val hosts = LinkedHashSet<String>()
        // Hücresel/sanal arayüzlerde tarama yapma; sadece Wi-Fi, hotspot,
        // Ethernet ve USB tethering arayüzleri anlamlı.
        val skippedPrefixes = listOf(
            "ccmni", "rmnet", "pdp", "wwan", "clat", "dummy", "ifb",
            "sit", "ip6tnl", "ip_vti", "tun", "tap", "p2p", "bluetooth",
        )
        try {
            for (nif in NetworkInterface.getNetworkInterfaces()) {
                if (!nif.isUp || nif.isLoopback) continue
                val name = nif.name.lowercase()
                if (skippedPrefixes.any { name.startsWith(it) }) continue
                addInterfaceHosts(nif, hosts)
            }
        } catch (_: Exception) {
            // Arayüz listesi alınamazsa tarama atlanır; hata mesajı yine döner.
        }
        return hosts.toList()
    }

    private fun addInterfaceHosts(nif: NetworkInterface, hosts: MutableSet<String>) {
    for (interfaceAddress in nif.interfaceAddresses) {
        val ownIp = probeAddress(interfaceAddress.address) ?: continue
        // /24 varsayımı: ev/otel Wi-Fi ağlarında yaygın ve hızlı.
        val network = ownIp and 0xFFFFFF00L
        for (offset in 1..254) {
            val candidate = network + offset
            if (candidate == ownIp) continue
            hosts.add(longToIp(candidate))
        }
    }
    }

    private fun probeAddress(address: java.net.InetAddress): Long? {
        if (address !is Inet4Address) return null
    if (address.isLoopbackAddress || address.isLinkLocalAddress) return null
    val ownIp = ipToLong(address.hostAddress ?: return null) ?: return null
    if (ownIp == 0L) return null
    // Sadece özel ağlarda tara; mobil veride operatör /24'üne
    // istek yağdırmayalım.
    if (!isPrivateIpv4(ownIp)) return null
        return ownIp
    }

    private fun isPrivateIpv4(value: Long): Boolean {
        val first = ((value shr 24) and 0xFF).toInt()
        val second = ((value shr 16) and 0xFF).toInt()
        return first == 10 ||
            (first == 172 && second in 16..31) ||
            (first == 192 && second == 168)
    }

    private fun ipToLong(host: String): Long? {
        val parts = host.split('.')
        if (parts.size != 4) return null
        var value = 0L
        for (part in parts) {
            val octet = part.toIntOrNull() ?: return null
            if (octet !in 0..255) return null
            value = (value shl 8) or octet.toLong()
        }
        return value
    }

    private fun longToIp(value: Long): String =
        "${(value shr 24) and 0xFF}.${(value shr 16) and 0xFF}.${(value shr 8) and 0xFF}.${value and 0xFF}"

    suspend fun connectionStatus(): String = withContext(Dispatchers.IO) {
        try {
            "Bağlı: ${serverBaseUrl()}"
        } catch (error: IOException) {
            resolvedBaseUrl = null
            error.message ?: "Bağlantı yok"
        }
    }

    suspend fun products(): List<LocalProduct> = withContext(Dispatchers.IO) {
        db.productDao().getAll()
    }

    suspend fun lots(): List<LocalLot> = withContext(Dispatchers.IO) {
        db.lotDao().getAll()
    }

    suspend fun markets(): List<LocalMarket> = withContext(Dispatchers.IO) {
        db.marketDao().getAll()
    }

    suspend fun sales(): List<LocalSale> = withContext(Dispatchers.IO) {
        db.saleDao().getAll()
    }

    suspend fun predictions(): List<LocalPrediction> = withContext(Dispatchers.IO) {
        db.predictionDao().getRecent()
    }

    suspend fun predictYield(
        rawKg: Double,
        spiceKg: Double,
        casingKg: Double,
        serviceCost: Double,
        laborCost: Double,
    ): PredictionView = withContext(Dispatchers.IO) {
        val base = try {
            serverBaseUrl()
        } catch (error: IOException) {
            return@withContext offlinePrediction(error)
        }
        val payload = JSONObject()
            .put("raw_material_kg", rawKg)
            .put("spice_kg", spiceKg)
            .put("casing_kg", casingKg)
            .put("service_cost", serviceCost)
            .put("labor_cost", laborCost)
        postPrediction(
            url = "$base/ml/predict/v2/yield",
            payload = payload,
            kind = "yield",
            inputSummary = "%.1f kg hammadde".format(rawKg),
        )
    }

    suspend fun predictButchery(
        animalType: String,
        liveWeight: Double,
        carcassWeight: Double?,
    ): PredictionView = withContext(Dispatchers.IO) {
        val base = try {
            serverBaseUrl()
        } catch (error: IOException) {
            return@withContext offlinePrediction(error)
        }
        val payload = JSONObject()
            .put("animal_type", animalType)
            .put("live_weight_kg", liveWeight)
        if (carcassWeight != null) {
            payload.put("carcass_weight_kg", carcassWeight)
        }
        postPrediction(
            url = "$base/ml/predict/v2/butchery",
            payload = payload,
            kind = "butchery_$animalType",
            inputSummary = "%.1f kg canlı ağırlık".format(liveWeight),
        )
    }

    private suspend fun postPrediction(
        url: String,
        payload: JSONObject,
        kind: String,
        inputSummary: String,
    ): PredictionView {
        return try {
            post(url, payload.toString()).use { response ->
                val body = response.body?.string().orEmpty()
                when {
                    response.isSuccessful -> {
                        val view = parsePrediction(body)
                        if (!view.offline && view.modelStatus != "unavailable") {
                            db.predictionDao().insert(
                                LocalPrediction(
                                    createdAt = LocalDateTime.now().toString(),
                                    kind = kind,
                                    inputSummary = inputSummary,
                                    resultSummary = view.rows.joinToString(" | ") { "${it.first}=${it.second}" },
                                    modelVersion = view.modelVersion,
                                    modelStatus = view.modelStatus,
                                    warnings = view.warnings.joinToString(" | "),
                                ),
                            )
                        }
                        view
                    }
                    response.code == 503 -> PredictionView(
                        rows = emptyList(),
                        warnings = emptyList(),
                        modelVersion = "unknown",
                        modelStatus = "unavailable",
                        message = "Model hazır değil (503): $body",
                    )
                    else -> PredictionView(
                        rows = emptyList(),
                        warnings = emptyList(),
                        modelVersion = "unknown",
                        modelStatus = "error",
                        message = "Tahmin başarısız: HTTP ${response.code}",
                    )
                }
            }
        } catch (error: IOException) {
            resolvedBaseUrl = null
            offlinePrediction(error)
        }
    }

    private fun offlinePrediction(error: IOException): PredictionView = PredictionView(
        rows = emptyList(),
        warnings = emptyList(),
        modelVersion = "unknown",
        modelStatus = "offline",
        offline = true,
        message = "Çevrimdışı: tahmin için sunucuya ulaşılamadı (${error.message}).",
    )

    private fun parsePrediction(body: String): PredictionView {
        val json = JSONObject(body)
        val prediction = json.getJSONObject("prediction")
        val rows = buildList {
            val keys = prediction.keys()
            while (keys.hasNext()) {
                val key = keys.next()
                val value = prediction.optDouble(key, Double.NaN)
                if (!value.isNaN()) {
                    val formatted = if (key.endsWith("_pct")) "%.2f".format(value) else "%.3f".format(value)
                    add((PREDICTION_LABELS[key] ?: key) to formatted)
                }
            }
        }
        val warnings = json.optJSONArray("warnings")?.toStringSet()?.toList().orEmpty()
        return PredictionView(
            rows = rows,
            warnings = warnings,
            modelVersion = json.optString("model_version", "unknown"),
            modelStatus = json.optString("model_status", "unknown"),
        )
    }

    suspend fun pull(): Unit = withContext(Dispatchers.IO) {
        try {
            pullOnce()
        } catch (error: IOException) {
            resolvedBaseUrl = null
            throw error
        }
    }

    private suspend fun pullOnce() {
        val base = serverBaseUrl()
        val pullBody = get("$base/sync/pull").use { response ->
            if (!response.isSuccessful) {
                throw IOException("pull failed: HTTP ${response.code}")
            }
            JSONObject(response.body?.string().orEmpty())
        }
        val products = parseProducts(pullBody.optJSONArray("products") ?: JSONArray())
        val blockedCodes = if (pullBody.isNull("blocked_lot_codes")) {
            emptySet()
        } else {
            pullBody.getJSONArray("blocked_lot_codes").toStringSet()
        }
        val lots = parseLots(pullBody.optJSONArray("lots") ?: JSONArray(), blockedCodes)
        val markets = parseMarkets(pullBody.optJSONArray("markets") ?: JSONArray())

        // Replace the local cache atomically: a failed pull must not leave the
        // SQLite database half-empty.
        db.withTransaction {
            db.productDao().clear()
            db.productDao().upsertAll(products)
            db.lotDao().clear()
            db.lotDao().upsertAll(lots)
            db.marketDao().clear()
            db.marketDao().upsertAll(markets)
        }
    }

    suspend fun reports(): ReportView = withContext(Dispatchers.IO) {
        try {
            val base = serverBaseUrl()
            val body = get("$base/reports/summary").use { response ->
                if (!response.isSuccessful) {
                    throw IOException("reports failed: HTTP ${response.code}")
                }
                JSONObject(response.body?.string().orEmpty())
            }
            val sections = listOf(
                "purchases" to "Alış",
                "production" to "Üretim",
                "butchery" to "Parçalama",
                "lot_movements" to "Lot Hareketleri",
                "sales" to "Satış",
                "recalls" to "Geri Çağırma",
            ).map { (key, title) ->
                val array = body.optJSONArray(key) ?: JSONArray()
                val columns = if (array.length() > 0) {
                    array.getJSONObject(0).keys().asSequence().toList()
                } else {
                    emptyList()
                }
                val rows = buildList {
                    for (index in 0 until array.length()) {
                        val item = array.getJSONObject(index)
                        add(columns.map { column -> item.opt(column)?.toString().orEmpty() })
                    }
                }
                ReportSection(title = title, count = rows.size, columns = columns, rows = rows)
            }
            ReportView(
                generatedAt = body.optString("generated_at"),
                sections = sections,
            )
        } catch (error: IOException) {
            resolvedBaseUrl = null
            ReportView(
                generatedAt = "",
                sections = emptyList(),
                offline = true,
                message = error.message ?: "Raporlar alınamadı",
            )
        }
    }

    suspend fun saveSale(sale: LocalSale, lines: List<LocalSaleLine>): Long = withContext(Dispatchers.IO) {
        val saleId = db.saleDao().insert(sale.copy(synced = false))
        val savedLines = lines.map { it.copy(saleId = saleId) }
        db.saleDao().insertLines(savedLines)
        db.outboxDao().insert(
            Outbox(
                entity = "sale",
                entityId = saleId,
                operation = "create",
                payload = salePayload(sale.copy(id = saleId), savedLines).toString(),
                createdAt = Instant.now().toString(),
            ),
        )
        // Yerel stok hareketi: satılan miktar lot bakiyesinden düşülür.
        savedLines.forEach { line -> db.lotDao().adjustQty(line.lotId, -line.qty) }
        saleId
    }

    suspend fun submitSale(saleId: Long): PushResult = withContext(Dispatchers.IO) {
        val sale = db.saleDao().getById(saleId) ?: return@withContext PushResult.Failed("sale not found")
        if (sale.status == "rejected") {
            // Already handled; do not restore stock twice on a repeated submit.
            return@withContext PushResult.Blocked("blocked lot: sale rejected")
        }
        if (sale.status == "stock_low") {
            return@withContext PushResult.InsufficientStock("insufficient stock: sale rejected")
        }
        val lines = db.saleDao().getLines(saleId)
        try {
            val base = serverBaseUrl()
            post("$base/sales", salePayload(sale, lines).toString()).use { response ->
                when {
                    response.code == 409 -> {
                        // Server rejected the sale (blocked lot or insufficient
                        // stock). Undo the local stock deduction and do not retry
                        // forever.
                        val errorBody = response.body?.string().orEmpty()
                        val lowStock = errorBody.contains("insufficient stock", ignoreCase = true)
                        lines.forEach { line -> db.lotDao().adjustQty(line.lotId, line.qty) }
                        db.saleDao().setStatus(saleId, if (lowStock) "stock_low" else "rejected")
                        db.saleDao().setSynced(saleId, true)
                        db.outboxDao().deleteByEntity("sale", saleId)
                        if (lowStock) {
                            PushResult.InsufficientStock("insufficient stock: sale rejected")
                        } else {
                            PushResult.Blocked("blocked lot: sale rejected")
                        }
                    }
                    response.isSuccessful -> {
                        db.saleDao().setStatus(saleId, "completed")
                        db.saleDao().setSynced(saleId, true)
                        db.outboxDao().deleteByEntity("sale", saleId)
                        PushResult.Success
                    }
                    else -> PushResult.Failed("HTTP ${response.code}")
                }
            }
        } catch (error: IOException) {
            // Çevrimdışı: satış yerelde kalır, push() ile sonra gönderilir.
            resolvedBaseUrl = null
            PushResult.Failed("offline")
        }
    }

    suspend fun push(): PushResult = withContext(Dispatchers.IO) {
        val outbox = db.outboxDao().getAll()
        if (outbox.isEmpty()) {
            return@withContext PushResult.Success
        }

        var blocked = false
        var insufficient = false
        var failed = false
        // Push one sale at a time. The sale code is the idempotency key, so a
        // retry after a timeout cannot deduct stock twice.
        for (entry in outbox) {
            when (submitSale(entry.entityId)) {
                is PushResult.Success -> Unit
                is PushResult.Blocked -> blocked = true
                is PushResult.InsufficientStock -> insufficient = true
                is PushResult.Failed -> failed = true
            }
        }
        when {
            insufficient -> PushResult.InsufficientStock("insufficient stock: at least one sale rejected")
            blocked -> PushResult.Blocked("blocked lot: at least one sale rejected")
            failed -> PushResult.Failed("offline")
            else -> PushResult.Success
        }
    }

    private fun salePayload(sale: LocalSale, lines: List<LocalSaleLine>): JSONObject {
        val lineArray = JSONArray()
        lines.forEach { line ->
            lineArray.put(
                JSONObject()
                    .put("lot_id", line.lotId)
                    .put("qty", line.qty)
                    .put("unit_price", line.unitPrice)
                    .put("total_price", line.totalPrice),
            )
        }
        return JSONObject()
            .put("code", sale.code)
            .put("market_id", sale.marketId)
            .put("sold_at", sale.soldAt)
            .put("status", sale.status)
            .put("client_id", CLIENT_ID)
            .put("lines", lineArray)
    }

    private fun parseProducts(array: JSONArray): List<LocalProduct> = buildList {
        for (index in 0 until array.length()) {
            val item = array.getJSONObject(index)
            add(
                LocalProduct(
                    id = item.getLong("id"),
                    code = item.getString("code"),
                    name = item.getString("name"),
                    unit = item.getString("unit"),
                    listPrice = item.getDouble("list_price"),
                ),
            )
        }
    }

    private fun parseLots(array: JSONArray, blockedCodes: Set<String>): List<LocalLot> = buildList {
        for (index in 0 until array.length()) {
            val item = array.getJSONObject(index)
            val code = item.getString("code")
            add(
                LocalLot(
                    id = item.getLong("id"),
                    code = code,
                    itemType = item.getString("item_type"),
                    itemId = item.getLong("item_id"),
                    qty = item.getDouble("qty"),
                    unit = item.getString("unit"),
                    blocked = blockedCodes.contains(code),
                ),
            )
        }
    }

    private fun parseMarkets(array: JSONArray): List<LocalMarket> = buildList {
        for (index in 0 until array.length()) {
            val item = array.getJSONObject(index)
            add(
                LocalMarket(
                    id = item.getLong("id"),
                    code = item.getString("code"),
                    name = item.getString("name"),
                ),
            )
        }
    }

    private fun JSONArray.toStringSet(): Set<String> = buildSet {
        for (index in 0 until length()) {
            add(getString(index))
        }
    }

    private fun get(url: String, httpClient: OkHttpClient = client): Response =
        httpClient.newCall(Request.Builder().url(url).get().build()).execute()

    private fun post(url: String, body: String): Response =
        client.newCall(Request.Builder().url(url).post(body.toRequestBody(JSON)).build()).execute()
}
