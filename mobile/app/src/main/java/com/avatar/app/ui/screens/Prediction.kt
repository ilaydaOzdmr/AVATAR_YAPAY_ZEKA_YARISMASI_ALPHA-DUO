package com.avatar.app.ui.screens

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Card
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
import com.avatar.app.data.PredictionView
import com.avatar.app.data.entities.LocalPrediction

private val Green = Color(0xFF2C7058)
private val Amber = Color(0xFFB97722)
private val Red = Color(0xFFB83B35)

@Composable
fun Prediction(
    yieldResult: PredictionView?,
    butcheryResult: PredictionView?,
    history: List<LocalPrediction>,
    message: String?,
    onPredictYield: (Double, Double, Double, Double, Double) -> Unit,
    onPredictButchery: (String, Double, Double?) -> Unit,
) {
    var rawKg by remember { mutableStateOf("100") }
    var spiceKg by remember { mutableStateOf("2") }
    var casingKg by remember { mutableStateOf("3") }
    var serviceCost by remember { mutableStateOf("400") }
    var laborCost by remember { mutableStateOf("80") }

    var animalType by remember { mutableStateOf("lamb") }
    var liveWeight by remember { mutableStateOf("45") }
    var carcassWeight by remember { mutableStateOf("21") }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .verticalScroll(rememberScrollState())
            .padding(16.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        Text("YAPAY ZEKÂ", style = MaterialTheme.typography.labelSmall, color = Green, fontWeight = FontWeight.Bold)
        Text("Tahmin Merkezi", style = MaterialTheme.typography.headlineSmall)
        Text(
            "Tahminler Go backend üzerinden alınır; çevrimdışıyken açık mesaj gösterilir.",
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
        )

        if (message != null) {
            Text(message, color = Amber, style = MaterialTheme.typography.labelMedium)
        }

        // Yield
        Card(shape = RoundedCornerShape(14.dp)) {
            Column(Modifier.fillMaxWidth().padding(14.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                Text("Üretim verimi ve EKP", fontWeight = FontWeight.Bold)
                OutlinedTextField(rawKg, { rawKg = it }, label = { Text("Hammadde kg") }, keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Decimal), modifier = Modifier.fillMaxWidth())
                Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    OutlinedTextField(spiceKg, { spiceKg = it }, label = { Text("Baharat kg") }, keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Decimal), modifier = Modifier.weight(1f))
                    OutlinedTextField(casingKg, { casingKg = it }, label = { Text("Bağırsak kg") }, keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Decimal), modifier = Modifier.weight(1f))
                }
                Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    OutlinedTextField(serviceCost, { serviceCost = it }, label = { Text("Hizmet €") }, keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Decimal), modifier = Modifier.weight(1f))
                    OutlinedTextField(laborCost, { laborCost = it }, label = { Text("İşçilik €") }, keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Decimal), modifier = Modifier.weight(1f))
                }
                Button(
                    onClick = {
                        val raw = rawKg.toDoubleOrNull()
                        if (raw != null && raw > 0) {
                            onPredictYield(
                                raw,
                                spiceKg.toDoubleOrNull() ?: 0.0,
                                casingKg.toDoubleOrNull() ?: 0.0,
                                serviceCost.toDoubleOrNull() ?: 0.0,
                                laborCost.toDoubleOrNull() ?: 0.0,
                            )
                        }
                    },
                    modifier = Modifier.fillMaxWidth(),
                    colors = ButtonDefaults.buttonColors(containerColor = Green),
                ) { Text("Verim / EKP tahmini") }
                PredictionResult(yieldResult)
            }
        }

        // Butchery
        Card(shape = RoundedCornerShape(14.dp)) {
            Column(Modifier.fillMaxWidth().padding(14.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                Text("Parçalama tahmini", fontWeight = FontWeight.Bold)
                Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    Button(
                        onClick = { animalType = "lamb" },
                        colors = ButtonDefaults.buttonColors(
                            containerColor = if (animalType == "lamb") Green else MaterialTheme.colorScheme.surfaceContainerLow,
                            contentColor = if (animalType == "lamb") Color.White else MaterialTheme.colorScheme.onSurface,
                        ),
                    ) { Text("Kuzu") }
                    Button(
                        onClick = { animalType = "cattle" },
                        colors = ButtonDefaults.buttonColors(
                            containerColor = if (animalType == "cattle") Green else MaterialTheme.colorScheme.surfaceContainerLow,
                            contentColor = if (animalType == "cattle") Color.White else MaterialTheme.colorScheme.onSurface,
                        ),
                    ) { Text("Sığır") }
                }
                OutlinedTextField(liveWeight, { liveWeight = it }, label = { Text("Canlı ağırlık kg") }, keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Decimal), modifier = Modifier.fillMaxWidth())
                OutlinedTextField(carcassWeight, { carcassWeight = it }, label = { Text("Karkas kg (opsiyonel)") }, keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Decimal), modifier = Modifier.fillMaxWidth())
                Button(
                    onClick = {
                        val live = liveWeight.toDoubleOrNull()
                        if (live != null && live > 0) {
                            onPredictButchery(animalType, live, carcassWeight.toDoubleOrNull())
                        }
                    },
                    modifier = Modifier.fillMaxWidth(),
                    colors = ButtonDefaults.buttonColors(containerColor = Green),
                ) { Text("Parça ağırlıkları tahmini") }
                PredictionResult(butcheryResult)
            }
        }

        // History
        Card(shape = RoundedCornerShape(14.dp)) {
            Column(Modifier.fillMaxWidth().padding(14.dp), verticalArrangement = Arrangement.spacedBy(6.dp)) {
                Text("Tahmin geçmişi", fontWeight = FontWeight.Bold)
                if (history.isEmpty()) {
                    Text("Henüz tahmin yok.", style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
                } else {
                    history.take(10).forEach { item ->
                        Column(Modifier.fillMaxWidth()) {
                            Text("${item.created_at} · ${item.kind}", style = MaterialTheme.typography.labelSmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
                            Text(item.result_summary, style = MaterialTheme.typography.bodySmall)
                            Text(
                                "Model ${item.model_version} · ${item.model_status}",
                                style = MaterialTheme.typography.labelSmall,
                                color = if (item.model_status == "synthetic_only") Amber else Green,
                            )
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun PredictionResult(result: PredictionView?) {
    if (result == null) return
    if (result.offline || result.message != null) {
        Text(result.message.orEmpty(), color = Red, style = MaterialTheme.typography.bodySmall)
    }
    result.rows.forEach { (label, value) ->
        Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
            Text(label, style = MaterialTheme.typography.bodySmall)
            Text(value, style = MaterialTheme.typography.bodySmall, fontWeight = FontWeight.Bold)
        }
    }
    if (result.rows.isNotEmpty()) {
        Text(
            "Model ${result.modelVersion} · ${result.modelStatus}",
            style = MaterialTheme.typography.labelSmall,
            color = Amber,
        )
        result.warnings.forEach { warning ->
            Text("⚠ $warning", style = MaterialTheme.typography.labelSmall, color = Amber)
        }
    }
}
