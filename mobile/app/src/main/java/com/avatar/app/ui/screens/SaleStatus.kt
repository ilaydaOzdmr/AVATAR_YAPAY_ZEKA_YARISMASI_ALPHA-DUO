package com.avatar.app.ui.screens

import com.avatar.app.data.entities.LocalSale

internal enum class SaleDisplayStatus(val label: String) {
    Rejected("Bloklu lot reddedildi"),
    Synced("Senkron"),
    Pending("Bekliyor"),
}

internal fun saleDisplayStatus(sale: LocalSale): SaleDisplayStatus = when {
    sale.status == "rejected" -> SaleDisplayStatus.Rejected
    sale.synced -> SaleDisplayStatus.Synced
    else -> SaleDisplayStatus.Pending
}

// Sales are stored newest-first. A blocked-lot attempt is a rejected record,
// not a completed sale, so Home must not present it as the last sale.
internal fun latestAcceptedSale(sales: List<LocalSale>): LocalSale? =
    sales.firstOrNull { it.status != "rejected" }
