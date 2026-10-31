package com.avatar.app

import android.content.Context
import androidx.test.core.app.ActivityScenario
import androidx.test.core.app.ApplicationProvider
import androidx.compose.ui.test.assertIsDisplayed
import androidx.compose.ui.test.assertIsEnabled
import androidx.compose.ui.test.junit4.createEmptyComposeRule
import androidx.compose.ui.test.onNodeWithText
import androidx.compose.ui.test.onAllNodesWithText
import androidx.compose.ui.test.performClick
import androidx.compose.ui.test.performScrollTo
import androidx.compose.ui.test.performTextReplacement
import okhttp3.mockwebserver.Dispatcher
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import okhttp3.mockwebserver.RecordedRequest
import org.junit.Assert.assertTrue
import org.junit.Rule
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner
import org.robolectric.annotation.Config
import java.util.concurrent.ConcurrentLinkedQueue
import java.util.concurrent.atomic.AtomicBoolean

@RunWith(RobolectricTestRunner::class)
@Config(sdk = [34], qualifiers = "w800dp-h1280dp-mdpi")
class MainFlowTest {
    @get:Rule val compose = createEmptyComposeRule()

    @Test fun operatorCanSyncSellPredictReviewHistoryAndChangeServer() {
        val requests = ConcurrentLinkedQueue<String>()
        val modelUnavailable = AtomicBoolean(false)
        val reportsUnavailable = AtomicBoolean(false)
        val server = MockWebServer()
        server.dispatcher = object : Dispatcher() {
            override fun dispatch(request: RecordedRequest): MockResponse {
                val path = request.path.orEmpty()
                requests.add(path)
                if (path.startsWith("/ml/predict/") && modelUnavailable.get()) return MockResponse().setResponseCode(503).setBody("QA unavailable")
                if (path == "/reports/summary" && reportsUnavailable.get()) return MockResponse().setResponseCode(500).setBody("QA reports failed")
                val body = when {
                    path == "/health" -> """{"status":"ok"}"""
                    path == "/sync/pull" -> """{"products":[{"id":1,"code":"PRD","name":"Product","unit":"kg","list_price":12.5}],"lots":[{"id":2,"code":"LOT","item_type":"product","item_id":1,"qty":10,"unit":"kg"}],"markets":[{"id":3,"code":"MKT","name":"Market"}],"blocked_lot_codes":[]}"""
                    path.startsWith("/ml/predict/") -> """{"prediction":{"finished_qty":84.5,"yield_pct":0.845},"warnings":["QA model"],"model_version":"1.1.0","model_status":"synthetic_only"}"""
                    path == "/reports/summary" -> """{"generated_at":"2026-10-31","purchases":[],"production":[],"butchery":[],"lot_movements":[],"sales":[],"recalls":[]}"""
                    else -> """{"id":1,"status":"ok"}"""
                }
                return MockResponse().setHeader("Content-Type", "application/json").setBody(body)
            }
        }
        server.start()
        val context = ApplicationProvider.getApplicationContext<Context>()
        context.getSharedPreferences("avatar", Context.MODE_PRIVATE).edit()
            .putString("base_url_v2", server.url("/").toString().trimEnd('/')).commit()
        try {
            ActivityScenario.launch(MainActivity::class.java).use {
                compose.waitUntil(20_000) { compose.onAllNodesWithText("Sunucu:", substring = true).fetchSemanticsNodes().size >= 2 }
                compose.onNodeWithText("Değiştir").performClick()
                compose.onNodeWithText("http://<PC-IP>:8080").performTextReplacement(server.url("/").toString())
                compose.onNodeWithText("Kaydet ve dene").performClick()
                compose.onNodeWithText("Products").performClick()
                compose.onNodeWithText("PRD · Product").assertIsDisplayed()
                compose.onNodeWithText("Sale", useUnmergedTree = true).performClick()
                compose.onNodeWithText("Ürün seç").performClick()
                compose.onNodeWithText("PRD — Product").performClick()
                compose.onNodeWithText("Lot seç").performClick()
                compose.onNodeWithText("LOT (10.0 kg)").performClick()
                compose.onNodeWithText("Market seç").performClick()
                compose.onNodeWithText("MKT — Market").performClick()
                compose.onNodeWithText("Satışı kaydet").assertIsDisplayed().assertIsEnabled().performClick()
                compose.waitUntil(20_000) { requests.contains("/sales") }
                compose.onNodeWithText("Tahmin", useUnmergedTree = true).performClick()
                compose.onNodeWithText("Verim / EKP tahmini").performScrollTo().performClick()
                compose.waitUntil(20_000) { requests.contains("/ml/predict/v2/yield") }
                compose.onNodeWithText("Parça ağırlıkları tahmini").performScrollTo().performClick()
                compose.waitUntil(20_000) { requests.contains("/ml/predict/v2/butchery") }
                modelUnavailable.set(true)
                compose.onNodeWithText("Verim / EKP tahmini").performScrollTo().performClick()
                compose.waitUntil(20_000) { compose.onAllNodesWithText("Model hazır değil", substring = true).fetchSemanticsNodes().isNotEmpty() }
                compose.onNodeWithText("Parça ağırlıkları tahmini").performScrollTo().performClick()
                compose.waitUntil(20_000) { requests.count { path -> path == "/ml/predict/v2/butchery" } >= 2 }
                reportsUnavailable.set(true)
                compose.onNodeWithText("Raporlar", useUnmergedTree = true).performClick()
                compose.waitUntil(20_000) { requests.contains("/reports/summary") }
                compose.waitUntil(20_000) { compose.onAllNodesWithText("Raporlar alınamadı").fetchSemanticsNodes().isNotEmpty() }
                reportsUnavailable.set(false)
                compose.onNodeWithText("Raporları yenile").performClick()
                compose.onNodeWithText("History", useUnmergedTree = true).performClick()
                compose.onNodeWithText("Ana Sayfa", useUnmergedTree = true).performClick()
                compose.onNodeWithText("Dene").performClick()
                assertTrue(requests.contains("/sync/pull"))
                assertTrue(requests.contains("/sales"))
            }
        } finally { server.shutdown() }
    }
}
