package com.avatar.app.ui.screens

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material3.Button
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
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import com.avatar.app.data.entities.LocalLot
import com.avatar.app.data.entities.LocalProduct
import com.avatar.app.data.entities.LocalSale
import com.avatar.app.data.entities.LocalSaleLine
import java.time.LocalDateTime
import java.time.format.DateTimeFormatter

private val SOLD_AT_FORMAT = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss")

@Composable
fun Sale(
    products: List<LocalProduct>,
    lots: List<LocalLot>,
    message: String?,
    onCreateSale: (LocalSale, LocalSaleLine) -> Unit,
) {
    var selectedProduct by remember { mutableStateOf<LocalProduct?>(null) }
    var selectedLot by remember { mutableStateOf<LocalLot?>(null) }
    var quantity by remember { mutableStateOf("1") }
    var productMenu by remember { mutableStateOf(false) }
    var lotMenu by remember { mutableStateOf(false) }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(16.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        Text("New Sale", style = MaterialTheme.typography.headlineSmall)

        Column {
            Button(onClick = { productMenu = true }, modifier = Modifier.fillMaxWidth()) {
                Text(selectedProduct?.let { "${it.code} — ${it.name}" } ?: "Select product")
            }
            DropdownMenu(expanded = productMenu, onDismissRequest = { productMenu = false }) {
                products.forEach { product ->
                    DropdownMenuItem(
                        text = { Text("${product.code} — ${product.name}") },
                        onClick = {
                            selectedProduct = product
                            productMenu = false
                        },
                    )
                }
            }
        }

        Column {
            Button(onClick = { lotMenu = true }, modifier = Modifier.fillMaxWidth()) {
                val label = selectedLot?.let { lot ->
                    if (lot.blocked) "${lot.code} (blocked)" else lot.code
                } ?: "Select lot"
                Text(label)
            }
            DropdownMenu(expanded = lotMenu, onDismissRequest = { lotMenu = false }) {
                lots.forEach { lot ->
                    DropdownMenuItem(
                        text = { Text("${lot.code} (${lot.item_type})") },
                        enabled = !lot.blocked,
                        onClick = {
                            selectedLot = lot
                            lotMenu = false
                        },
                    )
                }
            }
        }

        OutlinedTextField(
            value = quantity,
            onValueChange = { quantity = it },
            label = { Text("Quantity") },
            keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Decimal),
            modifier = Modifier.fillMaxWidth(),
        )

        val product = selectedProduct
        val lot = selectedLot
        val qty = quantity.toDoubleOrNull()
        val canSell = product != null && lot != null && !lot.blocked && qty != null && qty > 0

        Button(
            onClick = {
                if (product != null && lot != null && qty != null) {
                    val sale = LocalSale(
                        code = "AND-${System.currentTimeMillis()}",
                        market_id = 1L,
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
        ) {
            Text("Create sale")
        }

        if (message != null) {
            Text(message, color = MaterialTheme.colorScheme.primary)
        }
    }
}
