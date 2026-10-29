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
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import com.avatar.app.data.entities.LocalLot
import com.avatar.app.data.entities.LocalProduct

private val Amber = Color(0xFFB97722)
private val Red = Color(0xFFB83B35)
private val Green = Color(0xFF2C7058)

@Composable
fun ProductList(
    products: List<LocalProduct>,
    lots: List<LocalLot>,
    onRefresh: () -> Unit,
) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(16.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        Text("DEPO YÖNETİMİ", style = MaterialTheme.typography.labelSmall, color = Green, fontWeight = FontWeight.Bold)
        Text("Stok & Ürünler", style = MaterialTheme.typography.headlineSmall)

        // Metrikler
        Row(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
            MetricCard("Ürün", products.size.toString(), Modifier.weight(1f))
            MetricCard("Bloklu lot", lots.count { it.blocked }.toString(), Modifier.weight(1f))
        }

        LazyColumn(
            modifier = Modifier.weight(1f),
            verticalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            item { Text("Ürünler", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold) }
            items(products, key = { it.id }) { product ->
                Card(shape = RoundedCornerShape(14.dp)) {
                    Row(
                        Modifier.fillMaxWidth().padding(14.dp),
                        horizontalArrangement = Arrangement.SpaceBetween,
                        verticalAlignment = Alignment.CenterVertically,
                    ) {
                        Column {
                            Text("${product.code} · ${product.name}", fontWeight = FontWeight.Bold)
                            Text(product.unit, style = MaterialTheme.typography.labelSmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
                        }
                        Text("${product.list_price} ₺", color = Green, fontWeight = FontWeight.Bold)
                    }
                }
            }

            item { Text("Lotlar", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold) }
            items(lots, key = { it.id }) { lot ->
                val statusColor = if (lot.blocked) Red else Green
                val statusText = if (lot.blocked) "Bloke" else "Aktif"
                Card(shape = RoundedCornerShape(14.dp)) {
                    Row(
                        Modifier.fillMaxWidth().padding(14.dp),
                        horizontalArrangement = Arrangement.SpaceBetween,
                        verticalAlignment = Alignment.CenterVertically,
                    ) {
                        Column {
                            Text(lot.code, fontWeight = FontWeight.Bold)
                            Text("${lot.qty} ${lot.unit}", style = MaterialTheme.typography.labelSmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
                        }
                        Text(statusText, style = MaterialTheme.typography.labelMedium, color = statusColor, fontWeight = FontWeight.Bold)
                    }
                }
            }
        }

        TextButton(onClick = onRefresh, modifier = Modifier.fillMaxWidth()) {
            Text("Yenile", color = Green)
        }
    }
}

@Composable
private fun MetricCard(label: String, value: String, modifier: Modifier = Modifier) {
    Card(
        modifier = modifier,
        shape = RoundedCornerShape(14.dp),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceContainerLow),
    ) {
        Column(Modifier.fillMaxWidth().padding(14.dp)) {
            Text(label, style = MaterialTheme.typography.labelSmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
            Text(value, style = MaterialTheme.typography.headlineSmall, color = Green, fontWeight = FontWeight.Bold)
        }
    }
}
