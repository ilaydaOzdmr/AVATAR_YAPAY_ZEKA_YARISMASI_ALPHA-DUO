package com.avatar.app.report

import android.graphics.Paint
import com.avatar.app.data.ReportSection
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner
import org.robolectric.annotation.Config

@RunWith(RobolectricTestRunner::class)
@Config(sdk = [34])
class ReportLayoutTest {
    @Test fun wrappedTextFitsColumnAndNeverExceedsLineLimit() {
        val paint = Paint().apply { textSize = 12f }
        val wrap = ReportPdfExporter::class.java.getDeclaredMethod("wrap", String::class.java, Float::class.javaPrimitiveType, Paint::class.java, Int::class.javaPrimitiveType)
        wrap.isAccessible = true
        for (text in listOf("", "short", "several words repeated several words repeated several words repeated", "X".repeat(200))) {
            @Suppress("UNCHECKED_CAST")
            val lines = wrap.invoke(ReportPdfExporter, text, 60f, paint, 2) as List<String>
            assertTrue(lines.size <= 2)
            assertTrue(lines.all { paint.measureText(it) <= 60.01f })
            if (text.isEmpty()) assertEquals(listOf(""), lines)
        }
    }

    @Test fun tableColumnsUseAvailablePageWidthWithoutNegativeSizes() {
        val widths = ReportPdfExporter::class.java.getDeclaredMethod("computeColumnWidths", ReportSection::class.java, Paint::class.java, Paint::class.java)
        widths.isAccessible = true
        val paint = Paint().apply { textSize = 12f }
        for (count in listOf(1, 3, 20)) {
            val section = ReportSection("Table", 1, List(count) { "column_$it" }, listOf(List(count) { "Long content ".repeat(count) }))
            val result = widths.invoke(ReportPdfExporter, section, paint, paint) as FloatArray
            assertEquals(count, result.size)
            assertTrue(result.all { it > 0 })
            assertEquals(790f, result.sum(), .01f)
        }
    }
}
