package com.avatar.app.data.entities

import androidx.room.Entity
import androidx.room.PrimaryKey

@Entity(tableName = "local_sales")
data class LocalSale(
    @PrimaryKey(autoGenerate = true) val id: Long = 0,
    val code: String,
    val market_id: Long,
    val sold_at: String,
    val status: String,
    val synced: Boolean,
)
