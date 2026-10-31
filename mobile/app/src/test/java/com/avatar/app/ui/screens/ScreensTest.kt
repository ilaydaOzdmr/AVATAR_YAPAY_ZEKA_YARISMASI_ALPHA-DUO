package com.avatar.app.ui.screens

import androidx.compose.material3.MaterialTheme
import androidx.compose.ui.test.assertIsDisplayed
import androidx.compose.ui.test.assertIsEnabled
import androidx.compose.ui.test.junit4.createComposeRule
import androidx.compose.ui.test.onNodeWithText
import androidx.compose.ui.test.onAllNodesWithText
import androidx.compose.ui.test.performClick
import androidx.compose.ui.test.performScrollTo
import com.avatar.app.data.PredictionView
import com.avatar.app.data.ReportSection
import com.avatar.app.data.ReportView
import com.avatar.app.data.entities.LocalPrediction
import org.junit.Assert.assertEquals
import org.junit.Rule
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner
import org.robolectric.annotation.Config

@RunWith(RobolectricTestRunner::class)
@Config(sdk = [34], qualifiers = "w800dp-h1280dp-mdpi")
class ScreensTest {
    @get:Rule val compose = createComposeRule()

    @Test fun predictionFormsSendCurrentInputsAndAnimalSelection() {
        var production: List<Double>? = null
        var animal = ""
        compose.setContent {
            MaterialTheme {
                Prediction(null, null, emptyList(), null,
                    onPredictYield = { raw, spice, casing, service, labor -> production = listOf(raw, spice, casing, service, labor) },
                    onPredictButchery = { type, _, _ -> animal = type })
            }
        }
        compose.onNodeWithText("Verim / EKP tahmini").performScrollTo().performClick()
        assertEquals(listOf(100.0, 2.0, 3.0, 400.0, 80.0), production)
        compose.onNodeWithText("Sığır").performScrollTo().performClick()
        compose.onNodeWithText("Parça ağırlıkları tahmini").performScrollTo().performClick()
        assertEquals("cattle", animal)
        compose.onNodeWithText("Henüz tahmin yok.").performScrollTo().assertIsDisplayed()
    }

    @Test fun predictionResultAndHistoryAreRenderedWithWarnings() {
        val result = PredictionView(listOf("Mamul" to "84.5 kg"), listOf("Demo warning"), "1.1.0", "synthetic_only")
        val history = listOf(LocalPrediction(1, "2026-10-31", "yield", "100 kg", "History result", "1.1.0", "synthetic_only", ""))
        compose.setContent { MaterialTheme { Prediction(result, result.copy(offline = true, message = "Offline result"), history, "Status", { _, _, _, _, _ -> }, { _, _, _ -> }) } }
        compose.onAllNodesWithText("Mamul")[0].performScrollTo().assertIsDisplayed()
        compose.onNodeWithText("Offline result").performScrollTo().assertIsDisplayed()
        compose.onNodeWithText("History result").performScrollTo().assertIsDisplayed()
    }

    @Test fun reportsRenderMeasuredRowsAndEmptySections() {
        val sections = listOf(
            ReportSection("Purchases", 101, listOf("lot_code", "total_cost"), listOf(listOf("LOT-QA", "12.5"))),
            ReportSection("Empty section", 0, emptyList(), emptyList()))
        compose.setContent { MaterialTheme { Reports(ReportView("2026-10-31", sections), "Ready", {}) } }
        compose.onNodeWithText("LOT-QA").performScrollTo().assertIsDisplayed()
        compose.onNodeWithText("12,50").performScrollTo().assertIsDisplayed()
        compose.onNodeWithText("Bu tablo için kayıt bulunmuyor.").performScrollTo().assertIsDisplayed()
    }

    @Test fun unloadedAndOfflineReportsShowRecoveryMessages() {
        compose.setContent { MaterialTheme { Reports(null, "Bağlantı hatası", {}) } }
        compose.onNodeWithText("Rapor henüz yüklenmedi").assertIsDisplayed()
    }

    @Test fun offlineReportsShowFailureInsteadOfBusinessRows() {
        compose.setContent { MaterialTheme { Reports(ReportView("", emptyList(), true, "QA offline"), null, {}) } }
        compose.onNodeWithText("Raporlar alınamadı").assertIsDisplayed()
        compose.onNodeWithText("QA offline").assertIsDisplayed()
    }

    @Test fun failedPdfExportRestoresTheActionButton() {
        // Robolectric lacks PdfDocument native output; this exercises recovery.
        val report = ReportView("2026-10-31", listOf(ReportSection("Table", 0, emptyList(), emptyList())))
        compose.setContent { MaterialTheme { Reports(report, null, {}) } }
        compose.onNodeWithText("PDF indir").performClick()
        compose.waitUntil(10_000) { compose.onAllNodesWithText("PDF oluşturulamadı", substring = true).fetchSemanticsNodes().isNotEmpty() }
        compose.onNodeWithText("PDF indir").assertIsEnabled()
    }
}
