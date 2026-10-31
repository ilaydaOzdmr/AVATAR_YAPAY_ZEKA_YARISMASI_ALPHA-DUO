package com.avatar.app.data.entities

import androidx.room.ColumnInfo
import androidx.room.Entity
import androidx.room.PrimaryKey

@Entity(tableName = "local_products")
data class LocalProduct(
    @PrimaryKey val id: Long,
    val code: String,
    val name: String,
    val unit: String,
    @ColumnInfo(name = "list_price") val listPrice: Double,
)
