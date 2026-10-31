package com.avatar.app.ui.screens

import com.avatar.app.data.entities.LocalLot
import com.avatar.app.data.entities.LocalMarket
import com.avatar.app.data.entities.LocalProduct
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertNull
import org.junit.Test

class SaleDraftTest {
    private val product = LocalProduct(1, "PRD", "Product", "kg", 12.5)
    private val lot = LocalLot(2, "LOT", "product", 1, 10.0, "kg", false)
    private val market = LocalMarket(3, "MKT", "Market")

    @Test fun validSelectionPreservesMarketLotPriceAndQuantity() {
        val draft = buildSaleDraft(product, lot, market, 2.0)
        assertNotNull(draft)
        assertEquals(market.id, draft!!.first.marketId)
        assertEquals(lot.id, draft.second.lotId)
        assertEquals(25.0, draft.second.totalPrice, 0.0001)
        assertEquals(12.5, draft.second.unitPrice, 0.0001)
    }

    @Test fun missingAndInvalidSelectionsCannotBeQueued() {
        assertNull(buildSaleDraft(null, lot, market, 1.0))
        assertNull(buildSaleDraft(product, null, market, 1.0))
        assertNull(buildSaleDraft(product, lot, null, 1.0))
        for (qty in listOf(null, 0.0, -1.0, Double.NaN, Double.POSITIVE_INFINITY)) {
            assertNull(buildSaleDraft(product, lot, market, qty))
        }
    }

    @Test fun blockedOrMismatchedLotCannotBeQueued() {
        assertNull(buildSaleDraft(product, lot.copy(blocked = true), market, 1.0))
        assertNull(buildSaleDraft(product, lot.copy(itemId = 99), market, 1.0))
        assertNull(buildSaleDraft(product, lot.copy(itemType = "material"), market, 1.0))
    }
}
