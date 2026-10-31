package com.avatar.app.ui.screens

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
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
import com.avatar.app.data.SyncRepository
import com.avatar.app.data.entities.LocalLot
import com.avatar.app.data.entities.LocalSale

private val Amber = Color(0xFFB97722)
private val AmberBg = Color(0xFFFFF4DF)
private val Red = Color(0xFFB83B35)
private val Green = Color(0xFF2C7058)

// AVATAR OPERASYON ana ekranı — prototipteki mobil önizlemeyle birebir.
@Composable
fun Home(
    blockedLots: List<LocalLot>,
    lastSale: LocalSale?,
    actions: HomeActions,
    connectionMessage: String? = null,
    serverUrl: String = SyncRepository.DEFAULT_BASE_URL,
) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(16.dp),
        verticalArrangement = Arrangement.spacedBy(14.dp),
    ) {
        // Başlık
        Text(
            text = "AVATAR OPERASYON",
            style = MaterialTheme.typography.labelSmall,
            color = MaterialTheme.colorScheme.primary,
            fontWeight = FontWeight.Bold,
        )
        Text(
            text = "Merhaba, Ayşe 👋",
            style = MaterialTheme.typography.headlineSmall,
        )
        Text(
            text = "Bugünkü operasyonların hazır.",
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
        )
        if (connectionMessage != null) {
            Text(
                text = connectionMessage,
                style = MaterialTheme.typography.labelSmall,
                color = if (connectionMessage.startsWith("Bağlı") || connectionMessage.startsWith("Sunucu:")) Green else Amber,
            )
        }
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(4.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Text(
                text = "Sunucu: $serverUrl",
                style = MaterialTheme.typography.labelSmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
                modifier = Modifier.weight(1f),
            )
            TextButton(onClick = actions.onRetry) { Text("Dene") }
            TextButton(onClick = actions.onServerSettings) { Text("Değiştir") }
        }

        // Kalite uyarısı
        if (blockedLots.isNotEmpty()) {
            Card(
                shape = RoundedCornerShape(14.dp),
                colors = CardDefaults.cardColors(containerColor = AmberBg),
            ) {
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(14.dp),
                    horizontalArrangement = Arrangement.spacedBy(10.dp),
                ) {
                    Text("⚠", style = MaterialTheme.typography.titleLarge, color = Amber)
                    Column(Modifier.weight(1f)) {
                        Text("Kalite uyarısı", fontWeight = FontWeight.Bold, color = Amber)
                        Text(
                            "${blockedLots.first().code} lotu bloke edildi. İlişkili ürünlerin satışı engellendi.",
                            style = MaterialTheme.typography.bodySmall,
                            color = Color(0xFF8A5A1B),
                        )
                        TextButton(onClick = actions.onViewAlerts, contentPadding = androidx.compose.foundation.layout.PaddingValues(0.dp)) {
                            Text("Detayları görüntüle →", color = Amber, fontWeight = FontWeight.Bold)
                        }
                    }
                }
            }
        }

        // Hızlı işlemler (2x2)
        Row(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
            QuickAction("▣", "Hızlı satış", "Lot seçerek satış yap", Modifier.weight(1f), actions.onQuickSale)
            QuickAction("⌗", "Lot okut", "Barkod / QR tara", Modifier.weight(1f), actions.onScan)
        }
        Row(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
            QuickAction("▤", "Stok sorgula", "Anlık stok durumu", Modifier.weight(1f), actions.onStock)
            QuickAction("↗", "Sevkiyatlar", "Hazır ve gönderilenler", Modifier.weight(1f), actions.onShipments)
        }

        // Son satış kartı
        if (lastSale != null) {
            Card(shape = RoundedCornerShape(14.dp)) {
                Column(Modifier.fillMaxWidth().padding(14.dp)) {
                    Row(
                        Modifier.fillMaxWidth(),
                        horizontalArrangement = Arrangement.SpaceBetween,
                        verticalAlignment = Alignment.CenterVertically,
                    ) {
                        Text("Son satış", fontWeight = FontWeight.Bold)
                        Text("Tamamlandı", style = MaterialTheme.typography.labelSmall, color = Green)
                    }
                    Spacer(Modifier.height(8.dp))
                    Text(
                        "${lastSale.code}",
                        style = MaterialTheme.typography.bodyMedium,
                    )
                    Spacer(Modifier.height(12.dp))
                    Text(
                        "Lot kontrolü",
                        style = MaterialTheme.typography.labelSmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                    )
                    Text("✓ Satışa uygun", style = MaterialTheme.typography.labelMedium, color = Green)
                }
            }
        }
    }
}

@Composable
private fun QuickAction(
    icon: String,
    title: String,
    subtitle: String,
    modifier: Modifier = Modifier,
    onClick: () -> Unit,
) {
    Card(
        onClick = onClick,
        modifier = modifier,
        shape = RoundedCornerShape(14.dp),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceContainerLow),
    ) {
        Column(Modifier.fillMaxWidth().padding(14.dp), verticalArrangement = Arrangement.spacedBy(6.dp)) {
            Text(icon, style = MaterialTheme.typography.titleLarge, color = MaterialTheme.colorScheme.primary)
            Text(title, fontWeight = FontWeight.Bold)
            Text(subtitle, style = MaterialTheme.typography.labelSmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
        }
    }
}

data class HomeActions(
    val onQuickSale: () -> Unit,
    val onScan: () -> Unit,
    val onStock: () -> Unit,
    val onShipments: () -> Unit,
    val onViewAlerts: () -> Unit,
    val onServerSettings: () -> Unit,
    val onRetry: () -> Unit,
)
