package com.avatar.app.ui.screens

import com.avatar.app.data.entities.LocalSale
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

class SaleStatusTest {
    private fun sale(
        status: String,
        synced: Boolean,
        id: Long = 1,
        code: String = "AND-1",
        soldAt: String = "2026-11-01 10:00:00",
    ) = LocalSale(id = id, code = code, marketId = 1, soldAt = soldAt, status = status, synced = synced)

    @Test
    fun rejectedSaleIsLabelledAsBlockedLotRejection() {
        assertEquals(SaleDisplayStatus.Rejected, saleDisplayStatus(sale("rejected", true)))
        assertEquals("Bloklu lot reddedildi", SaleDisplayStatus.Rejected.label)
    }

    @Test
    fun syncedAndPendingSalesKeepTheirLabels() {
        assertEquals(SaleDisplayStatus.Synced, saleDisplayStatus(sale("completed", true)))
        assertEquals(SaleDisplayStatus.Pending, saleDisplayStatus(sale("completed", false)))
    }

    @Test
    fun homeIgnoresRejectedAttemptWhenPickingLastSale() {
        val rejected = sale("rejected", true, id = 2, code = "AND-REJ", soldAt = "2026-11-01 11:00:00")
        val accepted = sale("completed", true, id = 1, code = "AND-OK", soldAt = "2026-11-01 10:00:00")

        assertEquals(accepted.code, latestAcceptedSale(listOf(rejected, accepted))?.code)
        assertNull(latestAcceptedSale(listOf(rejected)))
    }
}
