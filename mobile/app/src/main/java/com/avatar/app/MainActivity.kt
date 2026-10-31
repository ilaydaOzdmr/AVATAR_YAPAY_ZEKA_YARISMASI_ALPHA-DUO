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
import com.avatar.app.ui.screens.latestAcceptedSale
import com.avatar.app.ui.theme.AvatarAppTheme
import kotlinx.coroutines.CoroutineScope
import com.avatar.app.ui.screens.HomeActions
import com.avatar.app.data.entities.LocalSaleLine
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        val prefs = getSharedPreferences("avatar", MODE_PRIVATE)
        // Eski sürümlerin otomatik "localhost/10.0.2.2" değerini kullanıcı
        // tercihi sayma; fiziksel cihaz için PC LAN IP'sini varsayılan yap.
        val legacyAuto = setOf("localhost", "127.0.0.1", "10.0.2.2")
        val storedUrl = prefs.getString("base_url_v2", null)
            ?: prefs.getString("base_url", null)?.takeUnless { saved ->
                legacyAuto.any { saved.contains(it) }
            }
        val initialUrl = storedUrl?.takeIf { it.isNotBlank() } ?: SyncRepository.DEFAULT_BASE_URL
        val repository = SyncRepository(AppDatabase.get(this), baseUrl = initialUrl)
        setContent {
            AvatarAppTheme {
                AvatarApp(
                    repository = repository,
                    initialServerUrl = initialUrl,
                    onServerUrlChanged = { url ->
                        prefs.edit().putString("base_url_v2", url).apply()
                    },
                )
            }
        }
    }
}

private enum class Screen { Home, Products, Sale, Prediction, Reports, History }

private class AvatarState(val repository: SyncRepository, initialServerUrl: String) {
    var screen by mutableStateOf(Screen.Home)
    var serverUrl by mutableStateOf(initialServerUrl)
    var showServerDialog by mutableStateOf(false)
    var serverInput by mutableStateOf(initialServerUrl)
    var products by mutableStateOf(emptyList<LocalProduct>())
    var lots by mutableStateOf(emptyList<LocalLot>())
    var markets by mutableStateOf(emptyList<LocalMarket>())
    var sales by mutableStateOf(emptyList<LocalSale>())
    var predictionHistory by mutableStateOf(emptyList<LocalPrediction>())
    var yieldResult by mutableStateOf<PredictionView?>(null)
    var butcheryResult by mutableStateOf<PredictionView?>(null)
    var reportView by mutableStateOf<ReportView?>(null)
    var message by mutableStateOf<String?>(null)
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
            val active = repository.activeBaseUrl()
            if (active != null) serverUrl = active
            "Sunucu: ${active ?: "-"} · ürün, lot ve market senkronize edildi."
        } catch (error: Exception) {
            "Bağlantı hatası: ${error.message}"
        }
        refresh()
    }


    suspend fun connectInitially() {
        var attempts = 0
        while (true) {
            pullAndRefresh()
            val failed = message?.startsWith("Bağlantı hatası") == true
            if (!failed || attempts >= 6) break
            attempts++
            delay(5_000)
        }
    }
    suspend fun refreshReports() {
        val result = repository.reports()
        reportView = result
        message = if (result.offline) result.message else "Raporlar MySQL'den alındı."
    }
    suspend fun createSale(sale: LocalSale, line: LocalSaleLine) {
        val saleId = repository.saveSale(sale, listOf(line))
        message = when (repository.submitSale(saleId)) {
            is PushResult.Success -> "Satış senkronize edildi."
            is PushResult.Blocked -> "Bloklu lot: satış reddedildi."
            is PushResult.Failed -> "Yerel kaydedildi (çevrimdışı)."
        }
        refresh()
    }
    suspend fun predictYield(raw: Double, spice: Double, casing: Double, service: Double, labor: Double) {
        val result = repository.predictYield(raw, spice, casing, service, labor)
        yieldResult = result
        predictionHistory = repository.predictions()
        message = predictionMessage(result, "Üretim tahmini alındı.")
    }
    suspend fun predictButchery(animal: String, live: Double, carcass: Double?) {
        val result = repository.predictButchery(animal, live, carcass)
        butcheryResult = result
        predictionHistory = repository.predictions()
        message = predictionMessage(result, "Parçalama tahmini alındı.")
    }
    private fun predictionMessage(result: PredictionView, success: String): String? =
        if (result.offline || result.modelStatus == "unavailable") result.message else success

    suspend fun pushPending() {
        message = when (val result = repository.push()) {
            is PushResult.Success -> "Pending sales synced."
            is PushResult.Blocked -> "Blocked lot: sale rejected."
            is PushResult.Failed -> "Sync failed: ${result.message}"
        }
        refresh()
    }
    fun saveServerUrl(onChanged: (String) -> Unit): Boolean {
        val url = normalizeServerUrl(serverInput) ?: return false
        repository.setBaseUrl(url)
        serverUrl = url
        onChanged(url)
        showServerDialog = false
        return true
    }
}

internal fun normalizeServerUrl(input: String): String? {
    val url = input.trim()
    if (url.isEmpty()) return null
    if (url.startsWith("http://") || url.startsWith("https://")) return url
    return "http://$url"
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun AvatarApp(repository: SyncRepository, initialServerUrl: String, onServerUrlChanged: (String) -> Unit) {
    val state = remember { AvatarState(repository, initialServerUrl) }
    val scope = rememberCoroutineScope()
    LaunchedEffect(Unit) { state.connectInitially() }
    with(state) {
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
                            refreshReports()
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
            AvatarContent(state, scope)
        }
    }
        ServerDialog(state, scope, onServerUrlChanged)
    }
}

@Composable
private fun ServerDialog(state: AvatarState, scope: CoroutineScope, onServerUrlChanged: (String) -> Unit) {
    with(state) {

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
                        "Fiziksel cihaz: aynı Wi-Fi'da otomatik bulunur " +
                            "(varsayılan ${SyncRepository.DEFAULT_BASE_URL}).\n" +
                            "Emülatör: http://10.0.2.2:8080\n" +
                            "USB: `adb reverse tcp:8080 tcp:8080`",
                        style = MaterialTheme.typography.labelSmall,
                    )
                }
            },
            confirmButton = {
                TextButton(
                    onClick = {
                        if (saveServerUrl(onServerUrlChanged)) {
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
}

@Composable
private fun AvatarContent(state: AvatarState, scope: CoroutineScope) {
    with(state) {
        when (screen) {
            Screen.Home -> Home(
                blockedLots = lots.filter { it.blocked }, lastSale = latestAcceptedSale(sales),
                actions = HomeActions(
                    onQuickSale = { screen = Screen.Sale },
                    onScan = { message = "Barkod tarama simüle edildi." },
                    onStock = { screen = Screen.Products },
                    onShipments = { screen = Screen.History },
                    onViewAlerts = { screen = Screen.Products },
                    onServerSettings = { serverInput = serverUrl; showServerDialog = true },
                    onRetry = { repository.invalidateConnection(); scope.launch { pullAndRefresh() } },
                ), connectionMessage = message, serverUrl = serverUrl,
            )
            Screen.Products -> ProductList(products, lots) { scope.launch { pullAndRefresh() } }
            Screen.Sale -> Sale(products, lots, markets, message) { sale, line -> scope.launch { createSale(sale, line) } }
            Screen.Prediction -> Prediction(yieldResult, butcheryResult, predictionHistory, message,
                onPredictYield = { raw, spice, casing, service, labor -> scope.launch { predictYield(raw, spice, casing, service, labor) } },
                onPredictButchery = { animal, live, carcass -> scope.launch { predictButchery(animal, live, carcass) } },
            )
            Screen.Reports -> Reports(reportView, message) { scope.launch { refreshReports() } }
            Screen.History -> History(sales, message) { scope.launch { pushPending() } }
        }
    }
}
