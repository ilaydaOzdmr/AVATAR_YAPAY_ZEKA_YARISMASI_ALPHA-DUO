package com.avatar.app.ui.screens

import androidx.compose.foundation.ScrollState
import androidx.compose.foundation.background
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyListState
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import com.avatar.app.data.ReportSection
import com.avatar.app.data.ReportView
import com.avatar.app.report.ReportFormatting
import com.avatar.app.report.ReportPdfExporter
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext

private val Green = Color(0xFF2C7058)
private val GreenSoft = Color(0xFFE7F1EC)
private val Red = Color(0xFFB83B35)
private val RedSoft = Color(0xFFFDECEB)
private val Amber = Color(0xFFB97722)
private val RowAlt = Color(0xFFF1F6F3)
private val RowBase = Color(0xFFFFFFFF)
private val Ink = Color(0xFF1B1B1F)

/** Arayüzde gösterilen en fazla satır; tamamı PDF'e yazılır. */
private const val MAX_UI_ROWS = 100

/** Kolon genişlikleri anahtar bazlı sabittir; tablo her cihazda aynı düzende kalır. */
private val columnWidths: Map<String, Dp> = mapOf(
    "lot_code" to 132.dp,
    "material_code" to 128.dp,
    "material_name" to 150.dp,
    "supplier_code" to 112.dp,
    "code" to 140.dp,
    "status" to 112.dp,
    "animal_type" to 92.dp,
    "movement_type" to 124.dp,
    "received_at" to 132.dp,
    "occurred_at" to 132.dp,
    "sold_at" to 132.dp,
    "created_at" to 132.dp,
    "market_code" to 100.dp,
    "reason" to 220.dp,
    "blocked_lots" to 240.dp,
    "qty" to 92.dp,
    "unit_cost" to 104.dp,
    "total_cost" to 112.dp,
    "total_price" to 112.dp,
    "input_kg" to 104.dp,
    "finished_qty" to 104.dp,
    "input_weight" to 118.dp,
    "parts_weight" to 118.dp,
    "fire_weight" to 100.dp,
    "part_count" to 100.dp,
    "line_count" to 104.dp,
    "ekp" to 96.dp,
)

@Composable
fun Reports(
    report: ReportView?,
    message: String?,
    onRefresh: () -> Unit,
) {
    val context = LocalContext.current
    val scope = rememberCoroutineScope()
    var exporting by remember { mutableStateOf(false) }
    var exportMessage by remember { mutableStateOf<String?>(null) }
    var exportFailed by remember { mutableStateOf(false) }

    val sections = report?.sections.orEmpty()
    val totalRows = sections.sumOf { it.count }
    val canExport = report != null && !report.offline && sections.isNotEmpty()
    val listState = rememberLazyListState()

    LazyColumn(
        modifier = Modifier.fillMaxSize(),
        state = listState,
        contentPadding = PaddingValues(16.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        item {
            Text(
                "RAPORLAR",
                style = MaterialTheme.typography.labelSmall,
                color = Green,
                fontWeight = FontWeight.Bold,
            )
            Text("Operasyon Raporları", style = MaterialTheme.typography.headlineSmall)
            Text(
                "Alış, üretim, parçalama, lot hareketleri, satış ve geri çağırma kayıtları tablo halinde. PDF olarak indirilebilir.",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )
        }

        item {
            Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                OutlinedButton(
                    onClick = onRefresh,
                    modifier = Modifier.weight(1f),
                    enabled = !exporting,
                    shape = RoundedCornerShape(12.dp),
                ) { Text("Raporları yenile") }

                Button(
                    onClick = {
                        if (!canExport || exporting) return@Button
                        val current = report ?: return@Button
                        exporting = true
                        exportMessage = null
                        exportFailed = false
                        scope.launch {
                            val result = withContext(Dispatchers.IO) {
                                runCatching { ReportPdfExporter.saveToDownloads(context, current) }
                            }
                            exporting = false
                            result.fold(
                                onSuccess = { saved ->
                                    exportFailed = false
                                    exportMessage = "PDF indirildi: ${saved.relativePath}"
                                },
                                onFailure = { error ->
                                    exportFailed = true
                                    exportMessage =
                                        "PDF oluşturulamadı: ${error.message ?: "bilinmeyen hata"}"
                                },
                            )
                        }
                    },
                    modifier = Modifier.weight(1f),
                    enabled = canExport && !exporting,
                    shape = RoundedCornerShape(12.dp),
                    colors = ButtonDefaults.buttonColors(containerColor = Green),
                ) {
                    if (exporting) {
                        CircularProgressIndicator(
                            modifier = Modifier.size(16.dp),
                            strokeWidth = 2.dp,
                            color = Color.White,
                        )
                        Text("  PDF hazırlanıyor", maxLines = 1)
                    } else {
                        Text("PDF indir", maxLines = 1)
                    }
                }
            }
        }

        if (report != null && !report.offline) {
            item {
                Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    StatChip("Kayıt", "$totalRows", Modifier.weight(1f))
                    StatChip("Tablo", "${sections.size}", Modifier.weight(1f))
                    StatChip(
                        "Oluşturma",
                        ReportFormatting.formatDate(report.generatedAt)
                            ?.substringBefore(' ')
                            ?: report.generatedAt.ifBlank { "—" },
                        Modifier.weight(1f),
                    )
                }
            }
        }

        if (exportMessage != null) {
            item {
                Surface(
                    color = if (exportFailed) RedSoft else GreenSoft,
                    shape = RoundedCornerShape(12.dp),
                    modifier = Modifier.fillMaxWidth(),
                ) {
                    Text(
                        exportMessage.orEmpty(),
                        modifier = Modifier.padding(horizontal = 12.dp, vertical = 10.dp),
                        style = MaterialTheme.typography.labelMedium,
                        color = if (exportFailed) Red else Green,
                    )
                }
            }
        }

        if (report?.offline == true) {
            item {
                Card(
                    shape = RoundedCornerShape(14.dp),
                    colors = CardDefaults.cardColors(containerColor = RedSoft),
                ) {
                    Column(
                        Modifier.fillMaxWidth().padding(14.dp),
                        verticalArrangement = Arrangement.spacedBy(6.dp),
                    ) {
                        Text(
                            "Raporlar alınamadı",
                            fontWeight = FontWeight.Bold,
                            color = Red,
                        )
                        Text(
                            report.message ?: "Sunucuya bağlanılamadı.",
                            style = MaterialTheme.typography.bodySmall,
                            color = Red,
                        )
                        Text(
                            "Mağaza biriminde ağ bağlantısını kontrol edip yeniden deneyin.",
                            style = MaterialTheme.typography.labelSmall,
                            color = MaterialTheme.colorScheme.onSurfaceVariant,
                        )
                    }
                }
            }
        }

        if (message != null && report?.offline != true) {
            item {
                val isError = message.startsWith("Bağlantı") ||
                    message.contains("çevrimdışı", ignoreCase = true) ||
                    message.contains("alınamadı", ignoreCase = true)
                Text(
                    message,
                    color = if (isError) Red else MaterialTheme.colorScheme.onSurfaceVariant,
                    style = MaterialTheme.typography.labelMedium,
                )
            }
        }

        if (report == null) {
            item {
                Card(shape = RoundedCornerShape(14.dp)) {
                    Column(
                        Modifier.fillMaxWidth().padding(16.dp),
                        verticalArrangement = Arrangement.spacedBy(6.dp),
                    ) {
                        Text("Rapor henüz yüklenmedi", fontWeight = FontWeight.Bold)
                        Text(
                            "“Raporları yenile” düğmesine basarak sunucudan son verileri çekin.",
                            style = MaterialTheme.typography.bodySmall,
                            color = MaterialTheme.colorScheme.onSurfaceVariant,
                        )
                    }
                }
            }
        }

        items(sections.size) { index ->
            val section = sections[index]
            ReportSectionCard(section)
        }

        if (sections.isNotEmpty()) {
            item {
                Text(
                    "Arayüzde tablo başına en fazla $MAX_UI_ROWS satır gösterilir. PDF indirdiğinizde tüm kayıtlar dahil edilir.",
                    style = MaterialTheme.typography.labelSmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                )
            }
        }
    }
}

@Composable
private fun StatChip(label: String, value: String, modifier: Modifier = Modifier) {
    Surface(
        color = GreenSoft,
        shape = RoundedCornerShape(12.dp),
        modifier = modifier,
    ) {
        Column(
            Modifier.padding(horizontal = 10.dp, vertical = 8.dp),
            verticalArrangement = Arrangement.spacedBy(2.dp),
        ) {
            Text(
                label,
                style = MaterialTheme.typography.labelSmall,
                color = Green,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
            Text(
                value,
                style = MaterialTheme.typography.labelLarge,
                fontWeight = FontWeight.Bold,
                color = Ink,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
        }
    }
}

@Composable
private fun ReportSectionCard(section: ReportSection) {
    Card(
        shape = RoundedCornerShape(14.dp),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
    ) {
        Column(Modifier.fillMaxWidth()) {
            Row(
                Modifier.fillMaxWidth().padding(horizontal = 14.dp, vertical = 12.dp),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Text(
                    section.title,
                    fontWeight = FontWeight.Bold,
                    style = MaterialTheme.typography.titleSmall,
                )
                Surface(color = GreenSoft, shape = RoundedCornerShape(8.dp)) {
                    Text(
                        "${section.count} kayıt",
                        modifier = Modifier.padding(horizontal = 8.dp, vertical = 3.dp),
                        style = MaterialTheme.typography.labelSmall,
                        color = Green,
                        fontWeight = FontWeight.SemiBold,
                    )
                }
            }

            if (section.rows.isEmpty() || section.columns.isEmpty()) {
                Text(
                    "Bu tablo için kayıt bulunmuyor.",
                    modifier = Modifier.padding(horizontal = 14.dp, vertical = 14.dp),
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                )
            } else {
                ReportTable(section)
                if (section.count > MAX_UI_ROWS) {
                    Text(
                        "İlk $MAX_UI_ROWS / ${section.count} kayıt gösteriliyor · tamamı PDF'te.",
                        modifier = Modifier.padding(horizontal = 14.dp, vertical = 8.dp),
                        style = MaterialTheme.typography.labelSmall,
                        color = Amber,
                    )
                }
            }
        }
    }
}

@Composable
private fun ReportTable(section: ReportSection) {
    val widths = remember(section.columns) { section.columns.map { columnWidth(it) } }
    val totalWidth = widths.fold(0.dp) { acc, width -> acc + width }
    val numeric = remember(section.columns) { section.columns.map { ReportFormatting.isNumeric(it) } }

    Column(
        Modifier
            .fillMaxWidth()
            .horizontalScroll(remember { ScrollState(0) }),
    ) {
        Row(
            Modifier
                .width(totalWidth)
                .background(GreenSoft)
                .padding(vertical = 2.dp),
        ) {
            section.columns.forEachIndexed { index, key ->
                TableCell(
                    text = ReportFormatting.label(key),
                    width = widths[index],
                    isHeader = true,
                    numeric = false,
                )
            }
        }

        section.rows.take(MAX_UI_ROWS).forEachIndexed { rowIndex, row ->
            Row(
                Modifier
                    .width(totalWidth)
                    .background(if (rowIndex % 2 == 1) RowAlt else RowBase),
            ) {
                section.columns.forEachIndexed { index, key ->
                    TableCell(
                        text = ReportFormatting.format(key, row.getOrNull(index).orEmpty()),
                        width = widths[index],
                        isHeader = false,
                        numeric = numeric[index],
                    )
                }
            }
            Box(
                Modifier
                    .width(totalWidth)
                    .height(1.dp)
                    .background(MaterialTheme.colorScheme.outlineVariant.copy(alpha = 0.4f)),
            )
        }
    }
}

@Composable
private fun TableCell(
    text: String,
    width: Dp,
    isHeader: Boolean,
    numeric: Boolean,
) {
    Text(
        text = text,
        modifier = Modifier
            .width(width)
            .padding(horizontal = 8.dp, vertical = 7.dp),
        style = if (isHeader) {
            MaterialTheme.typography.labelSmall
        } else {
            MaterialTheme.typography.bodySmall
        },
        fontWeight = if (isHeader) FontWeight.Bold else FontWeight.Normal,
        color = if (isHeader) Green else Ink,
        textAlign = if (numeric) TextAlign.End else TextAlign.Start,
        maxLines = if (isHeader) 2 else 3,
        overflow = TextOverflow.Ellipsis,
    )
}

private fun columnWidth(key: String): Dp =
    columnWidths[key] ?: if (ReportFormatting.isNumeric(key)) 100.dp else 124.dp
