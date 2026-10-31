package com.avatar.app.report

import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test

class ReportFormattingTest {
    @Test fun datesRetainBusinessDateAndMinutePrecision() {
        assertEquals("31.10.2026 18:05", ReportFormatting.formatDate("2026-10-31T18:05:22"))
        assertEquals("31.10.2026", ReportFormatting.formatDate("2026-10-31"))
        assertNull(ReportFormatting.formatDate("invalid"))
    }

    @Test fun missingValuesAndUnknownFieldsRemainReadable() {
        for (value in listOf("", "null", "<nil>")) assertEquals("—", ReportFormatting.format("qty", value))
        assertEquals("Yeni alan", ReportFormatting.label("yeni_alan"))
        assertEquals("unknown", ReportFormatting.format("status", "unknown"))
        assertEquals("Taslak", ReportFormatting.format("status", "draft"))
        assertEquals("Sevkiyat", ReportFormatting.format("movement_type", "shipment"))
    }

    @Test fun currencyAndWeightsUseTheirOwnPrecision() {
        assertEquals("1.234,50", ReportFormatting.format("total_cost", "1234.5"))
        assertEquals("1,235", ReportFormatting.format("input_weight", "1.23456"))
        assertEquals("2", ReportFormatting.format("line_count", "2"))
        assertTrue(ReportFormatting.isNumeric("input_weight"))
        assertFalse(ReportFormatting.isNumeric("lot_code"))
    }

    @Test fun blockedLotsAreDisplayedWithoutJsonSyntax() {
        assertEquals("LOT-1, LOT-2", ReportFormatting.format("blocked_lots", "[\"LOT-1\",\"LOT-2\"]"))
        assertEquals("—", ReportFormatting.format("blocked_lots", "[]"))
    }
}
