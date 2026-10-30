package com.avatar.app.data.entities

import androidx.room.Entity
import androidx.room.PrimaryKey

@Entity(tableName = "prediction_history")
data class LocalPrediction(
    @PrimaryKey(autoGenerate = true) val id: Long = 0,
    val created_at: String,
    val kind: String,
    val input_summary: String,
    val result_summary: String,
    val model_version: String,
    val model_status: String,
    val warnings: String,
)
