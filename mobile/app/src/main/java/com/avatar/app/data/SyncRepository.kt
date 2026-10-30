package com.avatar.app.data

import com.avatar.app.data.entities.LocalLot
import com.avatar.app.data.entities.LocalPrediction
import com.avatar.app.data.entities.LocalProduct
import com.avatar.app.data.entities.LocalSale
import com.avatar.app.data.entities.LocalSaleLine
import com.avatar.app.data.entities.Outbox
import java.io.IOException
import java.time.Instant
import java.time.LocalDateTime
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
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
    private val baseUrl: String = DEFAULT_BASE_URL,
) {
    companion object {
        // Fiziksel cihazda: adb reverse tcp:8080 tcp:8080 ile cihazın localhost:8080'i
        // host'un backend'ine yönlendirilir (weble aynı yer).
        // Emülatörde: adb reverse de çalışır; alternatif 10.0.2.2:8080.
        const val DEFAULT_BASE_URL = "http://localhost:8080"
        private const val CLIENT_ID = "android"
        private val JSON = "application/json; charset=utf-8".toMediaType()
    }

    suspend fun products(): List<LocalProduct> = withContext(Dispatchers.IO) {
        db.productDao().getAll()
    }

    suspend fun lots(): List<LocalLot> = withContext(Dispatchers.IO) {
        db.lotDao().getAll()
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
        val payload = JSONObject()
            .put("raw_material_kg", rawKg)
            .put("spice_kg", spiceKg)
            .put("casing_kg", casingKg)
            .put("service_cost", serviceCost)
            .put("labor_cost", laborCost)
        postPrediction(
            url = "$baseUrl/ml/predict/v2/yield",
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
        val payload = JSONObject()
            .put("animal_type", animalType)
            .put("live_weight_kg", liveWeight)
        if (carcassWeight != null) {
            payload.put("carcass_weight_kg", carcassWeight)
        }
        postPrediction(
            url = "$baseUrl/ml/predict/v2/butchery",
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
                                    created_at = LocalDateTime.now().toString(),
                                    kind = kind,
                                    input_summary = inputSummary,
                                    result_summary = view.rows.joinToString(" | ") { "${it.first}=${it.second}" },
                                    model_version = view.modelVersion,
                                    model_status = view.modelStatus,
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
            PredictionView(
                rows = emptyList(),
                warnings = emptyList(),
                modelVersion = "unknown",
                modelStatus = "offline",
                offline = true,
                message = "Çevrimdışı: tahmin için sunucuya ulaşılamadı (${error.message}).",
            )
        }
    }

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
        val pullBody = get("$baseUrl/sync/pull").use { response ->
            if (!response.isSuccessful) {
                throw IOException("pull failed: HTTP ${response.code}")
            }
            JSONObject(response.body?.string().orEmpty())
        }
        val products = parseProducts(pullBody.getJSONArray("products"))
        val blockedCodes = if (pullBody.isNull("blocked_lot_codes")) {
            emptySet()
        } else {
            pullBody.getJSONArray("blocked_lot_codes").toStringSet()
        }

        val lotsBody = get("$baseUrl/lots").use { response ->
            if (!response.isSuccessful) {
                throw IOException("lots failed: HTTP ${response.code}")
            }
            JSONArray(response.body?.string().orEmpty())
        }
        val lots = parseLots(lotsBody, blockedCodes)

        db.productDao().clear()
        db.productDao().upsertAll(products)
        db.lotDao().clear()
        db.lotDao().upsertAll(lots)
    }

    suspend fun saveSale(sale: LocalSale, lines: List<LocalSaleLine>): Long = withContext(Dispatchers.IO) {
        val saleId = db.saleDao().insert(sale.copy(synced = false))
        val savedLines = lines.map { it.copy(sale_id = saleId) }
        db.saleDao().insertLines(savedLines)
        db.outboxDao().insert(
            Outbox(
                entity = "sale",
                entity_id = saleId,
                operation = "create",
                payload = salePayload(sale.copy(id = saleId), savedLines).toString(),
                created_at = Instant.now().toString(),
            ),
        )
        saleId
    }

    suspend fun submitSale(saleId: Long): PushResult = withContext(Dispatchers.IO) {
        val sale = db.saleDao().getById(saleId) ?: return@withContext PushResult.Failed("sale not found")
        val lines = db.saleDao().getLines(saleId)
        post("$baseUrl/sales", salePayload(sale, lines).toString()).use { response ->
            when {
                response.code == 409 -> {
                    db.saleDao().setSynced(saleId, false)
                    PushResult.Blocked("blocked lot: sale rejected")
                }
                response.isSuccessful -> {
                    db.saleDao().setSynced(saleId, true)
                    db.outboxDao().deleteByEntity("sale", saleId)
                    PushResult.Success
                }
                else -> PushResult.Failed("HTTP ${response.code}")
            }
        }
    }

    suspend fun push(): PushResult = withContext(Dispatchers.IO) {
        val outbox = db.outboxDao().getAll()
        if (outbox.isEmpty()) {
            return@withContext PushResult.Success
        }

        val sales = JSONArray()
        outbox.forEach { sales.put(JSONObject(it.payload)) }
        val body = JSONObject().put("sales", sales).toString()

        post("$baseUrl/sync/push", body).use { response ->
            when {
                response.code == 409 -> {
                    outbox.forEach { db.saleDao().setSynced(it.entity_id, false) }
                    PushResult.Blocked("blocked lot: sale rejected")
                }
                response.isSuccessful -> {
                    outbox.forEach { db.saleDao().setSynced(it.entity_id, true) }
                    db.outboxDao().clear()
                    PushResult.Success
                }
                else -> PushResult.Failed("HTTP ${response.code}")
            }
        }
    }

    private fun salePayload(sale: LocalSale, lines: List<LocalSaleLine>): JSONObject {
        val lineArray = JSONArray()
        lines.forEach { line ->
            lineArray.put(
                JSONObject()
                    .put("lot_id", line.lot_id)
                    .put("qty", line.qty)
                    .put("unit_price", line.unit_price)
                    .put("total_price", line.total_price),
            )
        }
        return JSONObject()
            .put("code", sale.code)
            .put("market_id", sale.market_id)
            .put("sold_at", sale.sold_at)
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
                    list_price = item.getDouble("list_price"),
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
                    item_type = item.getString("item_type"),
                    qty = item.getDouble("qty"),
                    unit = item.getString("unit"),
                    blocked = blockedCodes.contains(code),
                ),
            )
        }
    }

    private fun JSONArray.toStringSet(): Set<String> = buildSet {
        for (index in 0 until length()) {
            add(getString(index))
        }
    }

    private fun get(url: String): Response =
        client.newCall(Request.Builder().url(url).get().build()).execute()

    private fun post(url: String, body: String): Response =
        client.newCall(Request.Builder().url(url).post(body.toRequestBody(JSON)).build()).execute()
}
