package com.avatar.app.data.entities

import androidx.room.Entity
import androidx.room.PrimaryKey

@Entity(tableName = "outbox")
data class Outbox(
    @PrimaryKey(autoGenerate = true) val id: Long = 0,
    val entity: String,
    val entity_id: Long,
    val operation: String,
    val payload: String,
    val created_at: String,
)
