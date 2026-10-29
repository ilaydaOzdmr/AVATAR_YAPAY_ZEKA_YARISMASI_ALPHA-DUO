package com.avatar.app.data

import com.avatar.app.data.entities.LocalLot
import com.avatar.app.data.entities.LocalProduct
import com.avatar.app.data.entities.LocalSale
import com.avatar.app.data.entities.LocalSaleLine
import com.avatar.app.data.entities.Outbox
import java.io.IOException
import java.time.Instant
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

class SyncRepository(
    private val db: AppDatabase,
    private val client: OkHttpClient = OkHttpClient(),
    private val baseUrl: String = DEFAULT_BASE_URL,
) {
    companion object {
        // 10.0.2.2 = Android emulator host alias (127.0.0.1 of the dev machine).
        // On a physical device, replace with the host's LAN IP.
        const val DEFAULT_BASE_URL = "http://10.0.2.2:8080"
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
