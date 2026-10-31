package com.avatar.app.data.entities

import androidx.room.ColumnInfo
import androidx.room.Entity
import androidx.room.PrimaryKey

@Entity(tableName = "outbox")
data class Outbox(
    @PrimaryKey(autoGenerate = true) val id: Long = 0,
    val entity: String,
    @ColumnInfo(name = "entity_id") val entityId: Long,
    val operation: String,
    val payload: String,
    @ColumnInfo(name = "created_at") val createdAt: String,
)
