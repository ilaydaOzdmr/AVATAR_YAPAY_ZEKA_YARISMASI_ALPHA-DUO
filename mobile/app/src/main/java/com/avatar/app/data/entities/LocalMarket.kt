package com.avatar.app.data.entities

import androidx.room.Entity
import androidx.room.PrimaryKey

@Entity(tableName = "local_markets")
data class LocalMarket(
    @PrimaryKey val id: Long,
    val code: String,
    val name: String,
)
