package com.avatar.app.data.entities

import androidx.room.ColumnInfo
import androidx.room.Entity
import androidx.room.PrimaryKey

@Entity(tableName = "local_sales")
data class LocalSale(
    @PrimaryKey(autoGenerate = true) val id: Long = 0,
    val code: String,
    @ColumnInfo(name = "market_id") val marketId: Long,
    @ColumnInfo(name = "sold_at") val soldAt: String,
    val status: String,
    val synced: Boolean,
)
