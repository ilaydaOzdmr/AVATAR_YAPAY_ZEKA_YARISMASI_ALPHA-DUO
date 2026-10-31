package com.avatar.app

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.NavigationBar
import androidx.compose.material3.NavigationBarItem
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import com.avatar.app.data.AppDatabase
import com.avatar.app.data.PredictionView
import com.avatar.app.data.PushResult
import com.avatar.app.data.ReportView
import com.avatar.app.data.SyncRepository
import com.avatar.app.data.entities.LocalLot
import com.avatar.app.data.entities.LocalMarket
import com.avatar.app.data.entities.LocalPrediction
import com.avatar.app.data.entities.LocalProduct
import com.avatar.app.data.entities.LocalSale
import com.avatar.app.ui.screens.History
import com.avatar.app.ui.screens.Home
import com.avatar.app.ui.screens.Prediction
import com.avatar.app.ui.screens.ProductList
import com.avatar.app.ui.screens.Reports
import com.avatar.app.ui.screens.Sale
import com.avatar.app.ui.theme.AvatarAppTheme
import kotlinx.coroutines.launch

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        val prefs = getSharedPreferences("avatar", MODE_PRIVATE)
        val initialUrl = prefs.getString("base_url", SyncRepository.DEFAULT_BASE_URL)
            ?: SyncRepository.DEFAULT_BASE_URL
        val repository = SyncRepository(AppDatabase.get(this), baseUrl = initialUrl)
        setContent {
            AvatarAppTheme {
                AvatarApp(
                    repository = repository,
                    initialServerUrl = initialUrl,
                    onServerUrlChanged = { url ->
                        prefs.edit().putString("base_url", url).apply()
                    },
                )
            }
        }
    }
}

private enum class Screen { Home, Products, Sale, Prediction, Reports, History }

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun AvatarApp(
    repository: SyncRepository,
    initialServerUrl: String,
    onServerUrlChanged: (String) -> Unit,
) {
    var screen by remember { mutableStateOf(Screen.Home) }
    var serverUrl by remember { mutableStateOf(initialServerUrl) }
    var showServerDialog by remember { mutableStateOf(false) }
    var serverInput by remember { mutableStateOf(initialServerUrl) }
    var products by remember { mutableStateOf(emptyList<LocalProduct>()) }
    var lots by remember { mutableStateOf(emptyList<LocalLot>()) }
    var markets by remember { mutableStateOf(emptyList<LocalMarket>()) }
    var sales by remember { mutableStateOf(emptyList<LocalSale>()) }
    var predictionHistory by remember { mutableStateOf(emptyList<LocalPrediction>()) }
    var yieldResult by remember { mutableStateOf<PredictionView?>(null) }
    var butcheryResult by remember { mutableStateOf<PredictionView?>(null) }
    var reportView by remember { mutableStateOf<ReportView?>(null) }
    var message by remember { mutableStateOf<String?>(null) }
    val scope = rememberCoroutineScope()

    suspend fun refresh() {
        products = repository.products()
        lots = repository.lots()
        markets = repository.markets()
        sales = repository.sales()
        predictionHistory = repository.predictions()
    }

    suspend fun pullAndRefresh() {
        message = try {
            repository.pull()
            "Sunucu: ${repository.activeBaseUrl() ?: "-"} · ürün, lot ve market senkronize edildi."
        } catch (error: Exception) {
            "Bağlantı hatası: ${error.message}"
        }
        refresh()
    }

    LaunchedEffect(Unit) {
        pullAndRefresh()
    }

    Scaffold(
        modifier = Modifier.fillMaxSize(),
        topBar = {
            TopAppBar(title = { Text("AVATAR OPERASYON") })
        },
        bottomBar = {
            NavigationBar {
                NavigationBarItem(
                    selected = screen == Screen.Home,
                    onClick = { screen = Screen.Home },
                    icon = {},
                    label = { Text("Ana Sayfa") },
                )
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
                    selected = screen == Screen.Prediction,
                    onClick = { screen = Screen.Prediction },
                    icon = {},
                    label = { Text("Tahmin") },
                )
                NavigationBarItem(
                    selected = screen == Screen.Reports,
                    onClick = {
                        screen = Screen.Reports
                        scope.launch {
                            val result = repository.reports()
                            reportView = result
                            message = if (result.offline) result.message else "Raporlar MySQL'den alındı."
                        }
                    },
                    icon = {},
                    label = { Text("Raporlar") },
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
                Screen.Home -> Home(
                    blockedLots = lots.filter { it.blocked },
                    lastSale = sales.firstOrNull(),
                    onQuickSale = { screen = Screen.Sale },
                    onScan = { message = "Barkod tarama simüle edildi." },
                    onStock = { screen = Screen.Products },
                    onShipments = { screen = Screen.History },
                    onViewAlerts = { screen = Screen.Products },
                    connectionMessage = message,
                    serverUrl = serverUrl,
                    onServerSettings = {
                        serverInput = serverUrl
                        showServerDialog = true
                    },
                    onRetry = { scope.launch { pullAndRefresh() } },
                )
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
                    markets = markets,
                    message = message,
                    onCreateSale = { sale, line ->
                        scope.launch {
                            val saleId = repository.saveSale(sale, listOf(line))
                            message = when (repository.submitSale(saleId)) {
                                is PushResult.Success -> "Satış senkronize edildi."
                                is PushResult.Blocked -> "Bloklu lot: satış reddedildi."
                                is PushResult.Failed -> "Yerel kaydedildi (çevrimdışı)."
                            }
                            refresh()
                        }
                    },
                )
                Screen.Prediction -> Prediction(
                    yieldResult = yieldResult,
                    butcheryResult = butcheryResult,
                    history = predictionHistory,
                    message = message,
                    onPredictYield = { raw, spice, casing, service, labor ->
                        scope.launch {
                            val result = repository.predictYield(raw, spice, casing, service, labor)
                            yieldResult = result
                            predictionHistory = repository.predictions()
                            message = if (result.offline || result.modelStatus == "unavailable") {
                                result.message
                            } else {
                                "Üretim tahmini alındı."
                            }
                        }
                    },
                    onPredictButchery = { animalType, liveWeight, carcassWeight ->
                        scope.launch {
                            val result = repository.predictButchery(animalType, liveWeight, carcassWeight)
                            butcheryResult = result
                            predictionHistory = repository.predictions()
                            message = if (result.offline || result.modelStatus == "unavailable") {
                                result.message
                            } else {
                                "Parçalama tahmini alındı."
                            }
                        }
                    },
                )
                Screen.Reports -> Reports(
                    report = reportView,
                    message = message,
                    onRefresh = {
                        scope.launch {
                            val result = repository.reports()
                            reportView = result
                            message = if (result.offline) result.message else "Raporlar MySQL'den alındı."
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

    if (showServerDialog) {
        AlertDialog(
            onDismissRequest = { showServerDialog = false },
            title = { Text("Sunucu adresi") },
            text = {
                Column {
                    OutlinedTextField(
                        value = serverInput,
                        onValueChange = { serverInput = it },
                        label = { Text("http://<PC-IP>:8080") },
                        modifier = Modifier.fillMaxWidth(),
                    )
                    Text(
                        "Emülatör: http://10.0.2.2:8080\n" +
                            "Fiziksel cihaz: USB’de `adb reverse tcp:8080 tcp:8080` " +
                            "veya aynı Wi-Fi’da PC LAN IP’si.",
                        style = MaterialTheme.typography.labelSmall,
                    )
                }
            },
            confirmButton = {
                TextButton(
                    onClick = {
                        var url = serverInput.trim()
                        if (url.isNotEmpty() && !url.startsWith("http://") && !url.startsWith("https://")) {
                            url = "http://$url"
                        }
                        if (url.isNotEmpty()) {
                            repository.setBaseUrl(url)
                            serverUrl = url
                            onServerUrlChanged(url)
                            showServerDialog = false
                            scope.launch { pullAndRefresh() }
                        }
                    },
                ) { Text("Kaydet ve dene") }
            },
            dismissButton = {
                TextButton(onClick = { showServerDialog = false }) { Text("Vazgeç") }
            },
        )
    }
}
