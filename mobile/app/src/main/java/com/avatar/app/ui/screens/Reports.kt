package com.avatar.app.ui.screens

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Card
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import com.avatar.app.data.ReportView

private val Green = Color(0xFF2C7058)
private val Red = Color(0xFFB83B35)

@Composable
fun Reports(
    report: ReportView?,
    message: String?,
    onRefresh: () -> Unit,
) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .verticalScroll(rememberScrollState())
            .padding(16.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        Text("RAPORLAR", style = MaterialTheme.typography.labelSmall, color = Green, fontWeight = FontWeight.Bold)
        Text("Operasyon Raporları", style = MaterialTheme.typography.headlineSmall)
        Text(
            "Alış, üretim, parçalama, lot hareketleri, satış ve geri çağırma. Kaynak: MySQL (sunucu).",
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
        )

        Button(
            onClick = onRefresh,
            modifier = Modifier.fillMaxWidth(),
            colors = ButtonDefaults.buttonColors(containerColor = Green),
        ) { Text("Raporları yenile") }

        if (message != null) {
            Text(message, color = Red, style = MaterialTheme.typography.labelMedium)
        }
        if (report?.offline == true) {
            Text(
                report.message ?: "Çevrimdışı: raporlar alınamadı.",
                color = Red,
                style = MaterialTheme.typography.bodySmall,
            )
            return@Column
        }
        if (report == null) {
            Text("Rapor yüklenmedi. Yenile’ye basın.", style = MaterialTheme.typography.bodySmall)
            return@Column
        }

        report.sections.forEach { section ->
            Card(shape = RoundedCornerShape(14.dp)) {
                Column(
                    Modifier.fillMaxWidth().padding(14.dp),
                    verticalArrangement = Arrangement.spacedBy(6.dp),
                ) {
                    Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
                        Text(section.title, fontWeight = FontWeight.Bold)
                        Text("${section.count} kayıt", style = MaterialTheme.typography.labelSmall, color = Green)
                    }
                    if (section.rows.isEmpty()) {
                        Text("Kayıt yok.", style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
                    } else {
                        section.rows.take(25).forEach { row ->
                            Text(
                                row.joinToString(" · "),
                                style = MaterialTheme.typography.bodySmall,
                            )
                        }
                        if (section.rows.size > 25) {
                            Text(
                                "… ilk 25 kayıt gösteriliyor",
                                style = MaterialTheme.typography.labelSmall,
                                color = MaterialTheme.colorScheme.onSurfaceVariant,
                            )
                        }
                    }
                }
            }
        }
    }
}
