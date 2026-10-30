package com.avatar.app.data.entities

import androidx.room.Entity
import androidx.room.PrimaryKey

@Entity(tableName = "local_lots")
data class LocalLot(
    @PrimaryKey val id: Long,
    val code: String,
    val item_type: String,
    val item_id: Long,
    val qty: Double,
    val unit: String,
    val blocked: Boolean,
)
