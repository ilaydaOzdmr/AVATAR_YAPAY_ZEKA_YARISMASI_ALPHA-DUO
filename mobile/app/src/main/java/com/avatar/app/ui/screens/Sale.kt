package com.avatar.app.ui.screens

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Card
import androidx.compose.material3.DropdownMenu
import androidx.compose.material3.DropdownMenuItem
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import com.avatar.app.data.entities.LocalLot
import com.avatar.app.data.entities.LocalMarket
import com.avatar.app.data.entities.LocalProduct
import com.avatar.app.data.entities.LocalSale
import com.avatar.app.data.entities.LocalSaleLine
import java.time.LocalDateTime
import java.time.format.DateTimeFormatter

private val Green = Color(0xFF2C7058)
private val Red = Color(0xFFB83B35)

private val SOLD_AT_FORMAT = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss")

@Composable
fun Sale(
    products: List<LocalProduct>,
    lots: List<LocalLot>,
    markets: List<LocalMarket>,
    message: String?,
    onCreateSale: (LocalSale, LocalSaleLine) -> Unit,
) {
    var selectedProduct by remember { mutableStateOf<LocalProduct?>(null) }
    var selectedLot by remember { mutableStateOf<LocalLot?>(null) }
    var selectedMarket by remember { mutableStateOf<LocalMarket?>(null) }
    var quantity by remember { mutableStateOf("1") }
    var productMenu by remember { mutableStateOf(false) }
    var lotMenu by remember { mutableStateOf(false) }
    var marketMenu by remember { mutableStateOf(false) }

    // Yalnızca seçilen ürünle eşleşen (item_id) mamul lotları satılabilir.
    val sellableLots = lots.filter { lot ->
        val currentProduct = selectedProduct
        lot.itemType == "product" && (currentProduct == null || lot.itemId == currentProduct.id)
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(16.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        Text("SATIŞ & DAĞITIM", style = MaterialTheme.typography.labelSmall, color = Green, fontWeight = FontWeight.Bold)
        Text("Hızlı Satış", style = MaterialTheme.typography.headlineSmall)

        Card(shape = RoundedCornerShape(14.dp)) {
            Column(
                Modifier.fillMaxWidth().padding(14.dp),
                verticalArrangement = Arrangement.spacedBy(12.dp),
            ) {
                // Ürün seçimi
                Button(
                    onClick = { productMenu = true },
                    modifier = Modifier.fillMaxWidth(),
                    colors = ButtonDefaults.buttonColors(containerColor = MaterialTheme.colorScheme.surfaceContainerLow, contentColor = MaterialTheme.colorScheme.onSurface),
                ) {
                    Text(selectedProduct?.let { "${it.code} — ${it.name}" } ?: "Ürün seç")
                }
                DropdownMenu(expanded = productMenu, onDismissRequest = { productMenu = false }) {
                    products.forEach { product ->
                        DropdownMenuItem(
                            text = { Text("${product.code} — ${product.name}") },
                            onClick = {
                                selectedProduct = product
                                selectedLot = null
                                productMenu = false
                            },
                        )
                    }
                }

                // Lot seçimi (yalnızca ürünle eşleşen mamul lotları)
                Button(
                    onClick = { lotMenu = true },
                    modifier = Modifier.fillMaxWidth(),
                    colors = ButtonDefaults.buttonColors(containerColor = MaterialTheme.colorScheme.surfaceContainerLow, contentColor = MaterialTheme.colorScheme.onSurface),
                ) {
                    Text(lotLabel(selectedLot))
                }
                DropdownMenu(expanded = lotMenu, onDismissRequest = { lotMenu = false }) {
                    sellableLots.forEach { lot ->
                        DropdownMenuItem(
                            text = {
                                Text(
                                    "${lot.code} (${lot.qty} ${lot.unit})",
                                    color = if (lot.blocked) Red else MaterialTheme.colorScheme.onSurface,
                                )
                            },
                            enabled = !lot.blocked,
                            onClick = { selectedLot = lot; lotMenu = false },
                        )
                    }
                }

                // Market seçimi
                Button(
                    onClick = { marketMenu = true },
                    modifier = Modifier.fillMaxWidth(),
                    colors = ButtonDefaults.buttonColors(containerColor = MaterialTheme.colorScheme.surfaceContainerLow, contentColor = MaterialTheme.colorScheme.onSurface),
                ) {
                    Text(selectedMarket?.let { "${it.code} — ${it.name}" } ?: "Market seç")
                }
                DropdownMenu(expanded = marketMenu, onDismissRequest = { marketMenu = false }) {
                    markets.forEach { market ->
                        DropdownMenuItem(
                            text = { Text("${market.code} — ${market.name}") },
                            onClick = { selectedMarket = market; marketMenu = false },
                        )
                    }
                }

                // Miktar
                OutlinedTextField(
                    value = quantity,
                    onValueChange = { quantity = it },
                    label = { Text("Miktar") },
                    keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Decimal),
                    modifier = Modifier.fillMaxWidth(),
                )

                val product = selectedProduct
                val lot = selectedLot
                val market = selectedMarket
                val qty = quantity.toDoubleOrNull()
                val insufficient = lot != null && qty != null && qty > lot.qty + 0.0005
                val canSell = product != null && lot != null && market != null && !lot.blocked && !insufficient && qty != null && qty > 0

                if (insufficient && lot != null) {
                    Text(
                        "Stok yetersiz: mevcut ${lot.qty} ${lot.unit}",
                        style = MaterialTheme.typography.labelSmall,
                        color = Red,
                    )
                }

                Button(
                    onClick = {
                        buildSaleDraft(product, lot, market, qty)?.let { (sale, line) -> onCreateSale(sale, line) }
                    },
                    enabled = canSell,
                    modifier = Modifier.fillMaxWidth(),
                    colors = ButtonDefaults.buttonColors(containerColor = Green),
                ) {
                    Text("Satışı kaydet")
                }

                SaleMessage(message)
            }
        }
    }
}

private fun lotLabel(lot: LocalLot?): String {
    if (lot == null) return "Lot seç"
    return if (lot.blocked) "${lot.code} (bloke)" else lot.code
}

internal fun buildSaleDraft(product: LocalProduct?, lot: LocalLot?, market: LocalMarket?, qty: Double?): Pair<LocalSale, LocalSaleLine>? {
    if (product == null || lot == null || market == null || qty == null) return null
    if (lot.blocked || !qty.isFinite() || qty <= 0) return null
    if (qty > lot.qty + 0.0005) return null
    if (lot.itemType != "product" || lot.itemId != product.id) return null
    val sale = LocalSale(code = "AND-${System.currentTimeMillis()}", marketId = market.id,
        soldAt = LocalDateTime.now().format(SOLD_AT_FORMAT), status = "completed", synced = false)
    val line = LocalSaleLine(saleId = 0, lotId = lot.id, qty = qty,
        unitPrice = product.listPrice, totalPrice = product.listPrice * qty)
    return sale to line
}

@Composable
private fun SaleMessage(message: String?) {
    if (message != null) {
        val isError = message.contains("offline", ignoreCase = true) ||
            message.contains("yerel", ignoreCase = true) ||
            message.contains("bağlantı", ignoreCase = true) ||
            message.contains("reddedildi", ignoreCase = true) ||
            message.contains("yetersiz", ignoreCase = true) ||
            message.contains("failed", ignoreCase = true)
        Text(message, style = MaterialTheme.typography.labelMedium, color = if (isError) Red else Green)
    }
}
