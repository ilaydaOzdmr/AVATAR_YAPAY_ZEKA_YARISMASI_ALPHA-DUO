package com.avatar.app.data.entities

import androidx.room.ColumnInfo
import androidx.room.Entity
import androidx.room.PrimaryKey

@Entity(tableName = "local_lots")
data class LocalLot(
    @PrimaryKey val id: Long,
    val code: String,
    @ColumnInfo(name = "item_type") val itemType: String,
    @ColumnInfo(name = "item_id") val itemId: Long,
    val qty: Double,
    val unit: String,
    val blocked: Boolean,
)
