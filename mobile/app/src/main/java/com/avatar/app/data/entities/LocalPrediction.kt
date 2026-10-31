package com.avatar.app.data.entities

import androidx.room.ColumnInfo
import androidx.room.Entity
import androidx.room.PrimaryKey

@Entity(tableName = "prediction_history")
data class LocalPrediction(
    @PrimaryKey(autoGenerate = true) val id: Long = 0,
    @ColumnInfo(name = "created_at") val createdAt: String,
    val kind: String,
    @ColumnInfo(name = "input_summary") val inputSummary: String,
    @ColumnInfo(name = "result_summary") val resultSummary: String,
    @ColumnInfo(name = "model_version") val modelVersion: String,
    @ColumnInfo(name = "model_status") val modelStatus: String,
    val warnings: String,
)
