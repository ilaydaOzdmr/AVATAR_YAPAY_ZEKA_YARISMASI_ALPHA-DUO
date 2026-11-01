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
class InsufficientStockSaleTest {
    @Test
    fun insufficientStockSaleIsRestoredAndLabelledSeparately() = runBlocking {
        val server = MockWebServer()
        server.dispatcher = object : Dispatcher() {
            override fun dispatch(request: RecordedRequest): MockResponse {
                if (request.path.orEmpty() == "/sales") {
                    return MockResponse()
                        .setResponseCode(409)
                        .setHeader("Content-Type", "application/json")
                        .setBody("""{"error":"insufficient stock"}""")
                }
                val body = if (request.path.orEmpty() == "/health") """{"status":"ok"}""" else """{}"""
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
                LocalLot(id = 2, code = "LOT", itemType = "product", itemId = 1, qty = 0.5, unit = "kg", blocked = false),
            ))
            val saleId = repo.saveSale(
                LocalSale(
                    code = "AND-LOW",
                    marketId = 3,
                    soldAt = "2026-11-01 12:00:00",
                    status = "completed",
                    synced = false,
                ),
                listOf(LocalSaleLine(saleId = 0, lotId = 2, qty = 1.0, unitPrice = 10.0, totalPrice = 10.0)),
            )

            assertTrue(repo.submitSale(saleId) is PushResult.InsufficientStock)
            val stored = db.saleDao().getById(saleId)
            assertEquals("stock_low", stored?.status)
            assertEquals(SaleDisplayStatus.StockLow, stored?.let { saleDisplayStatus(it) })
            assertEquals("Stok az olduğu için satış engellendi", SaleDisplayStatus.StockLow.label)
            // The rejected sale must not leave the local lot short.
            assertEquals(0.5, db.lotDao().getAll().first { it.id == 2L }.qty, 0.0)
        } finally {
            db.close()
            server.shutdown()
        }
    }
}
