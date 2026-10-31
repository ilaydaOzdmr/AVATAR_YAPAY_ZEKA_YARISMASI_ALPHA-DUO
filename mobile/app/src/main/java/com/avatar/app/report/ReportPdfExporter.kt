package com.avatar.app.report

import android.content.ContentValues
import android.content.Context
import android.graphics.Canvas
import android.graphics.Color
import android.graphics.Paint
import android.graphics.Typeface
import android.graphics.pdf.PdfDocument
import android.os.Build
import android.os.Environment
import android.provider.MediaStore
import com.avatar.app.data.ReportSection
import com.avatar.app.data.ReportView
import java.io.ByteArrayOutputStream
import java.io.File
import java.io.IOException
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale

/**
 * Raporu A4 yatay, çok sayfalı, tablo düzeninde bir PDF'e dönüştürür.
 * Android 10+ cihazlarda doğrudan İndirilenler/AVATAR Raporlar klasörüne yazar.
 */
object ReportPdfExporter {
    data class Result(val relativePath: String, val location: String)

    private const val PAGE_WIDTH = 842
    private const val PAGE_HEIGHT = 595
    private const val MARGIN = 26f
    private const val HEADER_BAND = 46f
    private const val CELL_PADDING = 5f
    private const val CELL_PADDING_V = 4f
    private const val LINE_HEIGHT = 10.5f
    private const val HEADER_ROW_HEIGHT = 22f
    private const val CONTENT_WIDTH = PAGE_WIDTH - 2 * MARGIN

    private const val DARK = 0xFF1B1B1F.toInt()
    private const val MUTED = 0xFF56605B.toInt()
    private const val GREEN = 0xFF1F5C4A.toInt()
    private const val HEADER_BG = 0xFFE7F1EC.toInt()
    private const val ALT_ROW_BG = 0xFFF7FAF8.toInt()
    private const val BORDER = 0xFFD9E2DD.toInt()

    fun saveToDownloads(context: Context, report: ReportView): Result {
        val bytes = render(report)
        val stamp = SimpleDateFormat("yyyy-MM-dd_HH-mm", Locale.US).format(Date())
        val fileName = "avatar-rapor-$stamp.pdf"

        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
            val values = ContentValues().apply {
                put(MediaStore.MediaColumns.DISPLAY_NAME, fileName)
                put(MediaStore.MediaColumns.MIME_TYPE, "application/pdf")
                put(
                    MediaStore.MediaColumns.RELATIVE_PATH,
                    "${Environment.DIRECTORY_DOWNLOADS}/AVATAR Raporlar",
                )
                put(MediaStore.MediaColumns.IS_PENDING, 1)
            }
            val resolver = context.contentResolver
            val uri = resolver.insert(MediaStore.Downloads.EXTERNAL_CONTENT_URI, values)
                ?: throw IOException("İndirilenler klasörü açılamadı")
            try {
                resolver.openOutputStream(uri)?.use { output -> output.write(bytes) }
                    ?: throw IOException("PDF dosyası yazılamadı")
                values.clear()
                values.put(MediaStore.MediaColumns.IS_PENDING, 0)
                resolver.update(uri, values, null, null)
            } catch (error: Exception) {
                resolver.delete(uri, null, null)
                throw error
            }
            return Result(
                relativePath = "Downloads/AVATAR Raporlar/$fileName",
                location = uri.toString(),
            )
        }

        val directory = File(
            context.getExternalFilesDir(Environment.DIRECTORY_DOWNLOADS),
            "AVATAR Raporlar",
        )
        if (!directory.exists() && !directory.mkdirs()) {
            throw IOException("Rapor klasörü oluşturulamadı")
        }
        val file = File(directory, fileName)
        file.writeBytes(bytes)
        return Result(relativePath = file.absolutePath, location = file.absolutePath)
    }

    fun render(report: ReportView): ByteArray {
        val document = PdfDocument()

        val headerTitle = paint(14f, Typeface.BOLD, Color.WHITE)
        val headerSub = paint(9f, Typeface.NORMAL, 0xFFD6E8E0.toInt())
        val headerInfo = paint(8f, Typeface.NORMAL, 0xFFD6E8E0.toInt())
        val sectionPaint = paint(11f, Typeface.BOLD, GREEN)
        val bodyPaint = paint(8.5f, Typeface.NORMAL, DARK)
        val headerPaint = paint(8.5f, Typeface.BOLD, GREEN)
        val mutedPaint = paint(8f, Typeface.NORMAL, MUTED)

        val bandPaint = fillPaint(GREEN)
        val headerBgPaint = fillPaint(HEADER_BG)
        val altRowPaint = fillPaint(ALT_ROW_BG)
        val dividerPaint = strokePaint(BORDER)
        val accentPaint = fillPaint(GREEN)
        val whitePaint = paint(8f, Typeface.NORMAL, Color.WHITE)

        var pageNumber = 0
        var page: PdfDocument.Page? = null
        var canvas: Canvas? = null
        var y = 0f

        fun drawPageHeader() {
            val c = canvas ?: return
            c.drawColor(Color.WHITE)
            c.drawRect(0f, 0f, PAGE_WIDTH.toFloat(), HEADER_BAND, bandPaint)
            c.drawText("AVATAR OPERASYON", MARGIN, 20f, headerTitle)
            c.drawText("Operasyon Raporları", MARGIN, 35f, headerSub)

            val generated = ReportFormatting.formatDate(report.generatedAt) ?: report.generatedAt
            val right = PAGE_WIDTH.toFloat() - MARGIN
            val generatedText = "Oluşturma: $generated"
            val pageText = "Sayfa $pageNumber"
            c.drawText(generatedText, right - headerInfo.measureText(generatedText), 20f, headerInfo)
            c.drawText(pageText, right - headerInfo.measureText(pageText), 35f, headerInfo)
        }

        fun startPage() {
            page?.let { document.finishPage(it) }
            pageNumber += 1
            page = document.startPage(
                PdfDocument.PageInfo.Builder(PAGE_WIDTH, PAGE_HEIGHT, pageNumber).create(),
            )
            canvas = page!!.canvas
            drawPageHeader()
            y = HEADER_BAND + 20f
        }

        fun ensureSpace(height: Float): Boolean {
            if (y + height > PAGE_HEIGHT - MARGIN) {
                startPage()
                return true
            }
            return false
        }

        fun drawSectionHeader(section: ReportSection) {
            val c = canvas!!
            c.drawRect(MARGIN, y + 1f, MARGIN + 3f, y + 13f, accentPaint)
            c.drawText(section.title, MARGIN + 9f, y + 11f, sectionPaint)
            val countText = "${section.count} kayıt"
            c.drawText(
                countText,
                PAGE_WIDTH - MARGIN - mutedPaint.measureText(countText),
                y + 10f,
                mutedPaint,
            )
            y += 18f
            c.drawLine(MARGIN, y, PAGE_WIDTH - MARGIN, y, dividerPaint)
            y += 5f
        }

        startPage()

        for (section in report.sections) {
            ensureSpace(52f)
            drawSectionHeader(section)

            if (section.rows.isEmpty()) {
                canvas!!.drawText("Kayıt bulunmuyor.", MARGIN + 2f, y + 9f, mutedPaint)
                y += 24f
                continue
            }

            val widths = computeColumnWidths(section, headerPaint, bodyPaint)
            val numeric = section.columns.map { ReportFormatting.isNumeric(it) }

            fun drawHeaderRow() {
                val c = canvas!!
                c.drawRect(MARGIN, y, PAGE_WIDTH - MARGIN, y + HEADER_ROW_HEIGHT, headerBgPaint)
                var x = MARGIN
                section.columns.forEachIndexed { index, key ->
                    val lines = wrap(
                        ReportFormatting.label(key),
                        widths[index] - 2 * CELL_PADDING,
                        headerPaint,
                        maxLines = 2,
                    )
                    var textY = y + CELL_PADDING_V + LINE_HEIGHT - 2f
                    lines.forEach { line ->
                        c.drawText(line, x + CELL_PADDING, textY, headerPaint)
                        textY += LINE_HEIGHT
                    }
                    x += widths[index]
                }
                y += HEADER_ROW_HEIGHT
            }

            fun rowLines(row: List<String>): List<List<String>> =
                section.columns.indices.map { index ->
                    wrap(
                        ReportFormatting.format(
                            section.columns[index],
                            row.getOrNull(index).orEmpty(),
                        ),
                        widths[index] - 2 * CELL_PADDING,
                        bodyPaint,
                        maxLines = 3,
                    )
                }

            val firstRowHeight = rowLines(section.rows.first()).maxOf { it.size } * LINE_HEIGHT +
                2 * CELL_PADDING_V
            ensureSpace(HEADER_ROW_HEIGHT + firstRowHeight)
            drawHeaderRow()

            section.rows.forEachIndexed { rowIndex, row ->
                val cellLines = rowLines(row)
                val rowHeight = cellLines.maxOf { it.size } * LINE_HEIGHT + 2 * CELL_PADDING_V
                if (ensureSpace(rowHeight)) {
                    drawHeaderRow()
                }
                val c = canvas!!
                if (rowIndex % 2 == 1) {
                    c.drawRect(MARGIN, y, PAGE_WIDTH - MARGIN, y + rowHeight, altRowPaint)
                }
                var x = MARGIN
                cellLines.forEachIndexed { index, lines ->
                    var textY = y + CELL_PADDING_V + LINE_HEIGHT - 2.5f
                    lines.forEach { line ->
                        val textX = if (numeric[index]) {
                            x + widths[index] - CELL_PADDING - bodyPaint.measureText(line)
                        } else {
                            x + CELL_PADDING
                        }
                        c.drawText(line, textX, textY, bodyPaint)
                        textY += LINE_HEIGHT
                    }
                    x += widths[index]
                }
                y += rowHeight
                c.drawLine(MARGIN, y, PAGE_WIDTH - MARGIN, y, dividerPaint)
            }
            y += 16f
        }

        val finalCanvas = canvas!!
        val footerText = "AVATAR OPERASYON · Operasyon Raporları · Sayfa $pageNumber"
        finalCanvas.drawText(
            footerText,
            PAGE_WIDTH - MARGIN - whitePaint.measureText(footerText),
            PAGE_HEIGHT - 10f,
            paint(7f, Typeface.NORMAL, MUTED),
        )

        page?.let { document.finishPage(it) }
        return ByteArrayOutputStream().use { output ->
            document.writeTo(output)
            document.close()
            output.toByteArray()
        }
    }

    private fun computeColumnWidths(
        section: ReportSection,
        headerPaint: Paint,
        bodyPaint: Paint,
    ): FloatArray {
        val natural = FloatArray(section.columns.size) { index ->
            val key = section.columns[index]
            var widest = headerPaint.measureText(ReportFormatting.label(key))
            section.rows.take(200).forEach { row ->
                val text = ReportFormatting.format(key, row.getOrNull(index).orEmpty())
                widest = maxOf(widest, bodyPaint.measureText(text))
            }
            widest + 2 * CELL_PADDING
        }

        val minWidth = 54f
        val total = natural.sum()
        if (total <= CONTENT_WIDTH) {
            val extra = (CONTENT_WIDTH - total) / natural.size
            return FloatArray(natural.size) { natural[it] + extra }
        }

        val clamped = FloatArray(natural.size) { maxOf(natural[it], minWidth) }
        val clampedTotal = clamped.sum()
        if (clampedTotal <= CONTENT_WIDTH) {
            val extra = (CONTENT_WIDTH - clampedTotal) / clamped.size
            return FloatArray(clamped.size) { clamped[it] + extra }
        }

        val scale = CONTENT_WIDTH / clampedTotal
        return FloatArray(clamped.size) { clamped[it] * scale }
    }

    private fun wrap(text: String, maxWidth: Float, paint: Paint, maxLines: Int): List<String> {
        if (text.isEmpty()) return listOf("")
        if (paint.measureText(text) <= maxWidth) return listOf(text)

        val words = text.split(' ')
        val lines = mutableListOf<String>()
        var current = StringBuilder()

        for (word in words) {
            val candidate = if (current.isEmpty()) word else "$current $word"
            if (paint.measureText(candidate) <= maxWidth) {
                current = StringBuilder(candidate)
                continue
            }
            if (current.isNotEmpty()) {
                lines += current.toString()
                current = StringBuilder()
                if (lines.size == maxLines) break
            }
            if (paint.measureText(word) <= maxWidth) {
                current = StringBuilder(word)
            } else {
                lines += ellipsize(word, maxWidth, paint)
                if (lines.size == maxLines) {
                    current = StringBuilder()
                    break
                }
            }
        }
        if (lines.size < maxLines && current.isNotEmpty()) {
            lines += current.toString()
        }

        return lines.take(maxLines).mapIndexed { index, line ->
            if (index == maxLines - 1 && paint.measureText(line) > maxWidth) {
                ellipsize(line, maxWidth, paint)
            } else {
                line
            }
        }
    }

    private fun ellipsize(text: String, maxWidth: Float, paint: Paint): String {
        val suffix = "…"
        var end = text.length
        while (end > 0 && paint.measureText(text.substring(0, end) + suffix) > maxWidth) {
            end -= 1
        }
        return text.substring(0, end) + suffix
    }

    private fun paint(size: Float, style: Int, color: Int): Paint =
        Paint(Paint.ANTI_ALIAS_FLAG).apply {
            textSize = size
            typeface = Typeface.create(Typeface.SANS_SERIF, style)
            this.color = color
        }

    private fun fillPaint(color: Int): Paint =
        Paint().apply {
            this.color = color
            style = Paint.Style.FILL
        }

    private fun strokePaint(color: Int): Paint =
        Paint().apply {
            this.color = color
            style = Paint.Style.STROKE
            strokeWidth = 0.6f
        }
}
