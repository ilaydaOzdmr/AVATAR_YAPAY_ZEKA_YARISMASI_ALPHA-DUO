package com.avatar.app.ui.screens

import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.Button
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.ListItem
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.avatar.app.data.entities.LocalLot
import com.avatar.app.data.entities.LocalProduct

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
    ) {
        Text("Products", style = MaterialTheme.typography.headlineSmall)
        LazyColumn(modifier = Modifier.weight(1f)) {
            items(products, key = { it.id }) { product ->
                ListItem(
                    headlineContent = { Text("${product.code} — ${product.name}") },
                    supportingContent = { Text("${product.unit} · ${product.list_price}") },
                )
            }
            item {
                HorizontalDivider()
                Text(
                    "Lots",
                    style = MaterialTheme.typography.titleMedium,
                    modifier = Modifier.padding(top = 12.dp, bottom = 4.dp),
                )
            }
            items(lots, key = { it.id }) { lot ->
                ListItem(
                    headlineContent = { Text("${lot.code} (${lot.item_type})") },
                    supportingContent = {
                        if (lot.blocked) {
                            Text("Blocked — sales disabled", color = MaterialTheme.colorScheme.error)
                        } else {
                            Text("${lot.qty} ${lot.unit}")
                        }
                    },
                )
            }
        }
        Button(
            onClick = onRefresh,
            modifier = Modifier
                .fillMaxWidth()
                .padding(top = 8.dp),
        ) {
            Text("Refresh")
        }
    }
}
