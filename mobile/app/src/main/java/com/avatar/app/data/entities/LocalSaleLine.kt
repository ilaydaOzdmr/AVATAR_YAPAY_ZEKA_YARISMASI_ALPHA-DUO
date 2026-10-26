package com.avatar.app.data.entities

import androidx.room.Entity
import androidx.room.PrimaryKey

@Entity(tableName = "local_sale_lines")
data class LocalSaleLine(
    @PrimaryKey(autoGenerate = true) val id: Long = 0,
    val sale_id: Long,
    val lot_id: Long,
    val qty: Double,
    val unit_price: Double,
    val total_price: Double,
)
