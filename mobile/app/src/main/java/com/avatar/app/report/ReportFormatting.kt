package com.avatar.app.report

import java.text.NumberFormat
import java.util.Locale

/** Rapor kolonlarını Türkçe başlık ve okunabilir değerlere çevirir. */
object ReportFormatting {
    private val turkish: Locale = Locale("tr", "TR")

    private val labels = mapOf(
        "lot_code" to "Lot Kodu",
        "material_code" to "Malzeme Kodu",
        "material_name" to "Malzeme",
        "supplier_code" to "Tedarikçi",
        "qty" to "Miktar",
        "unit_cost" to "Birim Fiyat",
        "total_cost" to "Toplam Tutar",
        "received_at" to "Geliş Tarihi",
        "code" to "Kod",
        "status" to "Durum",
        "input_kg" to "Girdi (kg)",
        "finished_qty" to "Mamul (kg)",
        "ekp" to "EKP (€)",
        "animal_type" to "Hayvan",
        "input_weight" to "Giriş Ağırlığı (kg)",
        "parts_weight" to "Parça Ağırlığı (kg)",
        "fire_weight" to "Fire (kg)",
        "part_count" to "Parça Sayısı",
        "movement_type" to "Hareket Tipi",
        "occurred_at" to "Tarih",
        "market_code" to "Pazar",
        "sold_at" to "Satış Tarihi",
        "line_count" to "Kalem Sayısı",
        "total_price" to "Toplam Tutar",
        "reason" to "Neden",
        "created_at" to "Tarih",
        "blocked_lots" to "Bloklu Lotlar",
        "blocked_lot_count" to "Bloklu Lot Sayısı",
        "generated_at" to "Oluşturma",
    )

    private val moneyKeys = setOf("unit_cost", "total_cost", "total_price", "ekp")
    private val integerKeys = setOf("line_count", "part_count", "blocked_lot_count")
    private val dateKeys = setOf("received_at", "occurred_at", "sold_at", "created_at")

    private val statusValues = mapOf(
        "draft" to "Taslak",
        "planned" to "Planlandı",
        "in_progress" to "Devam ediyor",
        "completed" to "Tamamlandı",
        "rejected" to "Reddedildi",
        "cancelled" to "İptal",
        "canceled" to "İptal",
    )

    private val movementValues = mapOf(
        "inbound" to "Giriş",
        "outbound" to "Çıkış",
        "production_output" to "Üretim çıktısı",
        "production_consume" to "Üretim tüketimi",
        "butchery_output" to "Parçalama çıktısı",
        "butchery_consume" to "Parçalama tüketimi",
        "waste" to "Fire",
        "sale" to "Satış",
        "shipment" to "Sevkiyat",
        "return" to "İade",
        "transfer" to "Transfer",
        "adjust" to "Düzeltme",
        "adjustment" to "Düzeltme",
    )

    private val dateTimePattern = Regex("""^(\d{4})-(\d{2})-(\d{2})[ T](\d{2}):(\d{2})""")
    private val dateOnlyPattern = Regex("""^(\d{4})-(\d{2})-(\d{2})$""")

    fun label(key: String): String =
        labels[key] ?: key.replace('_', ' ').replaceFirstChar { it.uppercase(turkish) }

    fun isNumeric(key: String): Boolean =
        key in moneyKeys || key in integerKeys ||
            key == "qty" || key.endsWith("_kg") || key.endsWith("_weight") || key == "ekp"

    fun format(key: String, raw: String): String {
        val value = raw.trim()
        if (value.isEmpty() || value == "null" || value == "<nil>") return "—"
        if (key == "blocked_lots") return formatList(value)
        if (key in dateKeys) return formatDate(value) ?: value
        if (key == "status") return statusValues[value.lowercase(turkish)] ?: value
        if (key == "movement_type") return movementValues[value.lowercase(turkish)] ?: value
        value.toDoubleOrNull()?.let { return formatNumber(key, it) }
        return value
    }

    fun formatDate(raw: String): String? {
        dateTimePattern.find(raw.trim())?.let { match ->
            val (year, month, day, hour, minute) = match.destructured
            return "$day.$month.$year $hour:$minute"
        }
        dateOnlyPattern.find(raw.trim())?.let { match ->
            val (year, month, day) = match.destructured
            return "$day.$month.$year"
        }
        return null
    }

    private fun formatNumber(key: String, value: Double): String {
        val formatter = NumberFormat.getNumberInstance(turkish).apply {
            maximumFractionDigits = when {
                key in integerKeys -> 0
                key in moneyKeys -> 2
                else -> 3
            }
            minimumFractionDigits = if (key in moneyKeys) 2 else 0
        }
        return formatter.format(value)
    }

    private fun formatList(raw: String): String {
        val cleaned = raw.trim().removePrefix("[").removeSuffix("]")
        if (cleaned.isBlank()) return "—"
        return cleaned
            .split(',')
            .map { it.trim().trim('"') }
            .filter { it.isNotEmpty() }
            .joinToString(", ")
            .ifEmpty { "—" }
    }
}
