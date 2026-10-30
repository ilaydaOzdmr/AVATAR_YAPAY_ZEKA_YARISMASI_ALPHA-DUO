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
import androidx.compose.material3.CardDefaults
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
        lot.item_type == "product" && (currentProduct == null || lot.item_id == currentProduct.id)
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
                    Text(selectedLot?.let { if (it.blocked) "${it.code} (bloke)" else it.code } ?: "Lot seç")
                }
                DropdownMenu(expanded = lotMenu, onDismissRequest = { lotMenu = false }) {
                    sellableLots.forEach { lot ->
                        DropdownMenuItem(
                            text = { Text("${lot.code} (${lot.qty} ${lot.unit})") },
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
                val canSell = product != null && lot != null && market != null && !lot.blocked && qty != null && qty > 0

                Button(
                    onClick = {
                        if (product != null && lot != null && market != null && qty != null) {
                            val sale = LocalSale(
                                code = "AND-${System.currentTimeMillis()}",
                                market_id = market.id,
                                sold_at = LocalDateTime.now().format(SOLD_AT_FORMAT),
                                status = "completed",
                                synced = false,
                            )
                            val line = LocalSaleLine(
                                sale_id = 0,
                                lot_id = lot.id,
                                qty = qty,
                                unit_price = product.list_price,
                                total_price = product.list_price * qty,
                            )
                            onCreateSale(sale, line)
                        }
                    },
                    enabled = canSell,
                    modifier = Modifier.fillMaxWidth(),
                    colors = ButtonDefaults.buttonColors(containerColor = Green),
                ) {
                    Text("Satışı kaydet")
                }

                if (message != null) {
                    Text(message, style = MaterialTheme.typography.labelMedium, color = if (message.contains("offline") || message.contains("yerel")) Red else Green)
                }
            }
        }
    }
}
