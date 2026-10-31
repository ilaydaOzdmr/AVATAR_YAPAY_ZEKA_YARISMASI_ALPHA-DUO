package com.avatar.app.ui.screens

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Card
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import com.avatar.app.data.entities.LocalSale

private val Amber = Color(0xFFB97722)
private val Green = Color(0xFF2C7058)
private val Red = Color(0xFFB83B35)

@Composable
fun History(
    sales: List<LocalSale>,
    message: String?,
    onPush: () -> Unit,
) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(16.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        Text("SATIŞ & DAĞITIM", style = MaterialTheme.typography.labelSmall, color = Green, fontWeight = FontWeight.Bold)
        Text("Satış Geçmişi", style = MaterialTheme.typography.headlineSmall)

        LazyColumn(
            modifier = Modifier.weight(1f),
            verticalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            if (sales.isEmpty()) {
                item { Text("Henüz satış yok.", style = MaterialTheme.typography.bodyMedium, color = MaterialTheme.colorScheme.onSurfaceVariant) }
            }
            items(sales, key = { it.id }) { sale ->
                Card(shape = RoundedCornerShape(14.dp)) {
                    Row(
                        Modifier.fillMaxWidth().padding(14.dp),
                        horizontalArrangement = Arrangement.SpaceBetween,
                        verticalAlignment = Alignment.CenterVertically,
                    ) {
                        Column {
                            Text(sale.code, fontWeight = FontWeight.Bold)
                            Text(
                                "${sale.sold_at} · Market ${sale.market_id}",
                                style = MaterialTheme.typography.labelSmall,
                                color = MaterialTheme.colorScheme.onSurfaceVariant,
                            )
                        }
                        val (statusText, statusColor) = if (sale.synced) "Senkron" to Green else "Bekliyor" to Amber
                        Text(statusText, style = MaterialTheme.typography.labelMedium, color = statusColor, fontWeight = FontWeight.Bold)
                    }
                }
            }
        }

        TextButton(onClick = onPush, modifier = Modifier.fillMaxWidth()) {
            Text("Bekleyen satışları senkronla", color = Green)
        }
        if (message != null) {
            val isError = message.contains("offline", ignoreCase = true) ||
                message.contains("bağlantı", ignoreCase = true) ||
                message.contains("failed", ignoreCase = true) ||
                message.contains("reddedildi", ignoreCase = true)
            Text(message, style = MaterialTheme.typography.labelMedium, color = if (isError) Red else Green)
        }
    }
}
