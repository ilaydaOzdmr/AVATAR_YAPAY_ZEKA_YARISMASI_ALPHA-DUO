package com.avatar.app.ui.screens

import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.Button
import androidx.compose.material3.ListItem
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.avatar.app.data.entities.LocalSale

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
    ) {
        Text("Sales History", style = MaterialTheme.typography.headlineSmall)
        LazyColumn(modifier = Modifier.weight(1f)) {
            items(sales, key = { it.id }) { sale ->
                ListItem(
                    headlineContent = { Text(sale.code) },
                    supportingContent = { Text("${sale.sold_at} · market ${sale.market_id}") },
                    trailingContent = {
                        Text(
                            if (sale.synced) "Synced" else "Pending",
                            color = if (sale.synced) {
                                MaterialTheme.colorScheme.primary
                            } else {
                                MaterialTheme.colorScheme.error
                            },
                        )
                    },
                )
            }
        }
        Button(
            onClick = onPush,
            modifier = Modifier
                .fillMaxWidth()
                .padding(top = 8.dp),
        ) {
            Text("Sync pending sales")
        }
        if (message != null) {
            Text(message, color = MaterialTheme.colorScheme.secondary)
        }
    }
}
