package com.avatar.app.data.entities

import androidx.room.Entity
import androidx.room.PrimaryKey

@Entity(tableName = "local_products")
data class LocalProduct(
    @PrimaryKey val id: Long,
    val code: String,
    val name: String,
    val unit: String,
    val list_price: Double,
)
