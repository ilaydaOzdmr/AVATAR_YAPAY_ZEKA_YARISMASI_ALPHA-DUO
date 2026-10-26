package com.avatar.app

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.NavigationBar
import androidx.compose.material3.NavigationBarItem
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import com.avatar.app.data.AppDatabase
import com.avatar.app.data.PushResult
import com.avatar.app.data.SyncRepository
import com.avatar.app.data.entities.LocalLot
import com.avatar.app.data.entities.LocalProduct
import com.avatar.app.data.entities.LocalSale
import com.avatar.app.ui.screens.History
import com.avatar.app.ui.screens.ProductList
import com.avatar.app.ui.screens.Sale
import kotlinx.coroutines.launch

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        val repository = SyncRepository(AppDatabase.get(this))
        setContent {
            MaterialTheme {
                AvatarApp(repository)
            }
        }
    }
}

private enum class Screen { Products, Sale, History }

@Composable
private fun AvatarApp(repository: SyncRepository) {
    var screen by remember { mutableStateOf(Screen.Products) }
    var products by remember { mutableStateOf(emptyList<LocalProduct>()) }
    var lots by remember { mutableStateOf(emptyList<LocalLot>()) }
    var sales by remember { mutableStateOf(emptyList<LocalSale>()) }
    var message by remember { mutableStateOf<String?>(null) }
    val scope = rememberCoroutineScope()

    suspend fun refresh() {
        products = repository.products()
        lots = repository.lots()
        sales = repository.sales()
    }

    suspend fun pullAndRefresh() {
        message = try {
            repository.pull()
            "Synced products and lots."
        } catch (error: Exception) {
            "Pull failed: ${error.message}"
        }
        refresh()
    }

    LaunchedEffect(Unit) {
        pullAndRefresh()
    }

    Scaffold(
        modifier = Modifier.fillMaxSize(),
        bottomBar = {
            NavigationBar {
                NavigationBarItem(
                    selected = screen == Screen.Products,
                    onClick = { screen = Screen.Products },
                    icon = {},
                    label = { Text("Products") },
                )
                NavigationBarItem(
                    selected = screen == Screen.Sale,
                    onClick = { screen = Screen.Sale },
                    icon = {},
                    label = { Text("Sale") },
                )
                NavigationBarItem(
                    selected = screen == Screen.History,
                    onClick = { screen = Screen.History },
                    icon = {},
                    label = { Text("History") },
                )
            }
        },
    ) { innerPadding ->
        Box(
            modifier = Modifier
                .fillMaxSize()
                .padding(innerPadding),
        ) {
            when (screen) {
                Screen.Products -> ProductList(
                    products = products,
                    lots = lots,
                    onRefresh = {
                        scope.launch { pullAndRefresh() }
                    },
                )
                Screen.Sale -> Sale(
                    products = products,
                    lots = lots,
                    message = message,
                    onCreateSale = { sale, line ->
                        scope.launch {
                            val saleId = repository.saveSale(sale, listOf(line))
                            message = when (val result = repository.submitSale(saleId)) {
                                is PushResult.Success -> "Sale synced."
                                is PushResult.Blocked -> "Blocked lot: sale rejected."
                                is PushResult.Failed -> "Saved offline: ${result.message}"
                            }
                            refresh()
                        }
                    },
                )
                Screen.History -> History(
                    sales = sales,
                    message = message,
                    onPush = {
                        scope.launch {
                            message = when (val result = repository.push()) {
                                is PushResult.Success -> "Pending sales synced."
                                is PushResult.Blocked -> "Blocked lot: sale rejected."
                                is PushResult.Failed -> "Sync failed: ${result.message}"
                            }
                            refresh()
                        }
                    },
                )
            }
        }
    }
}
