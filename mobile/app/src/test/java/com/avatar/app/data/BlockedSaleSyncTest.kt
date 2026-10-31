package com.avatar.app.data

import android.content.Context
import androidx.room.Room
import androidx.test.core.app.ApplicationProvider
import com.avatar.app.data.entities.LocalLot
import com.avatar.app.data.entities.LocalSale
import com.avatar.app.data.entities.LocalSaleLine
import com.avatar.app.ui.screens.SaleDisplayStatus
import com.avatar.app.ui.screens.saleDisplayStatus
import kotlinx.coroutines.runBlocking
import okhttp3.mockwebserver.Dispatcher
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import okhttp3.mockwebserver.RecordedRequest
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner
import org.robolectric.annotation.Config

@RunWith(RobolectricTestRunner::class)
@Config(sdk = [34])
class BlockedSaleSyncTest {
    @Test
    fun blockedSaleStaysRejectedAndLabelledAfterPull() = runBlocking {
        val server = MockWebServer()
        server.dispatcher = object : Dispatcher() {
            override fun dispatch(request: RecordedRequest): MockResponse {
                val path = request.path.orEmpty()
                if (path == "/sales") {
                    return MockResponse()
                        .setResponseCode(409)
                        .setHeader("Content-Type", "application/json")
                        .setBody("""{"error":"lot is blocked by a recall"}""")
                }
                val body = when {
                    path == "/health" -> """{"status":"ok"}"""
                    path == "/sync/pull" -> """{"products":[{"id":1,"code":"PRD","name":"Product","unit":"kg","list_price":10}],"lots":[{"id":2,"code":"LOT","item_type":"product","item_id":1,"qty":10,"unit":"kg"}],"markets":[{"id":3,"code":"MKT","name":"Market"}],"blocked_lot_codes":["LOT"]}"""
                    else -> """{"status":"ok"}"""
                }
                return MockResponse().setHeader("Content-Type", "application/json").setBody(body)
            }
        }
        server.start()
        val context = ApplicationProvider.getApplicationContext<Context>()
        val db = Room.inMemoryDatabaseBuilder(context, AppDatabase::class.java)
            .allowMainThreadQueries()
            .build()
        try {
            val repo = SyncRepository(db, baseUrl = server.url("/").toString().trimEnd('/'))
            db.lotDao().upsertAll(listOf(
                LocalLot(id = 2, code = "LOT", itemType = "product", itemId = 1, qty = 10.0, unit = "kg", blocked = false),
            ))
            val saleId = repo.saveSale(
                LocalSale(
                    code = "AND-REJ",
                    marketId = 3,
                    soldAt = "2026-11-01 12:00:00",
                    status = "completed",
                    synced = false,
                ),
                listOf(LocalSaleLine(saleId = 0, lotId = 2, qty = 1.0, unitPrice = 10.0, totalPrice = 10.0)),
            )

            assertTrue(repo.submitSale(saleId) is PushResult.Blocked)
            assertEquals("rejected", db.saleDao().getById(saleId)?.status)
            // The rejected sale must not leave the local lot short.
            assertEquals(10.0, db.lotDao().getAll().first { it.id == 2L }.qty, 0.0)

            // Syncing (pull) must not erase or turn the blocked attempt back into
            // a normal sale; History relies on this status for its label.
            repo.pull()

            val stored = db.saleDao().getById(saleId)
            assertEquals("rejected", stored?.status)
            assertEquals(SaleDisplayStatus.Rejected, stored?.let { saleDisplayStatus(it) })
            assertEquals("Bloklandı", SaleDisplayStatus.Rejected.label)
        } finally {
            db.close()
            server.shutdown()
        }
    }
}
