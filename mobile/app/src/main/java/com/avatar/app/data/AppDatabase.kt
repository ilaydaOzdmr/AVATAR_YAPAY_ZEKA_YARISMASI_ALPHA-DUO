package com.avatar.app.data

import android.content.Context
import androidx.room.Dao
import androidx.room.Database
import androidx.room.Insert
import androidx.room.OnConflictStrategy
import androidx.room.Query
import androidx.room.Room
import androidx.room.RoomDatabase
import androidx.room.migration.Migration
import androidx.sqlite.db.SupportSQLiteDatabase
import com.avatar.app.data.entities.LocalLot
import com.avatar.app.data.entities.LocalMarket
import com.avatar.app.data.entities.LocalPrediction
import com.avatar.app.data.entities.LocalProduct
import com.avatar.app.data.entities.LocalSale
import com.avatar.app.data.entities.LocalSaleLine
import com.avatar.app.data.entities.Outbox

@Dao
interface ProductDao {
    @Query("SELECT * FROM local_products ORDER BY name")
    suspend fun getAll(): List<LocalProduct>

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun upsertAll(products: List<LocalProduct>)

    @Query("DELETE FROM local_products")
    suspend fun clear()
}

@Dao
interface LotDao {
    @Query("SELECT * FROM local_lots ORDER BY code")
    suspend fun getAll(): List<LocalLot>

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun upsertAll(lots: List<LocalLot>)

    @Query("DELETE FROM local_lots")
    suspend fun clear()

    @Query("UPDATE local_lots SET qty = qty + :delta WHERE id = :lotId")
    suspend fun adjustQty(lotId: Long, delta: Double)
}

@Dao
interface MarketDao {
    @Query("SELECT * FROM local_markets ORDER BY name")
    suspend fun getAll(): List<LocalMarket>

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun upsertAll(markets: List<LocalMarket>)

    @Query("DELETE FROM local_markets")
    suspend fun clear()
}

@Dao
interface SaleDao {
    @Query("SELECT * FROM local_sales ORDER BY sold_at DESC")
    suspend fun getAll(): List<LocalSale>

    @Query("SELECT * FROM local_sales WHERE id = :saleId")
    suspend fun getById(saleId: Long): LocalSale?

    @Insert
    suspend fun insert(sale: LocalSale): Long

    @Query("UPDATE local_sales SET synced = :synced WHERE id = :saleId")
    suspend fun setSynced(saleId: Long, synced: Boolean)

    @Query("SELECT * FROM local_sale_lines WHERE sale_id = :saleId")
    suspend fun getLines(saleId: Long): List<LocalSaleLine>

    @Insert
    suspend fun insertLines(lines: List<LocalSaleLine>)
}

@Dao
interface OutboxDao {
    @Query("SELECT * FROM outbox ORDER BY id")
    suspend fun getAll(): List<Outbox>

    @Insert
    suspend fun insert(entry: Outbox): Long

    @Query("DELETE FROM outbox WHERE entity = :entity AND entity_id = :entityId")
    suspend fun deleteByEntity(entity: String, entityId: Long)

    @Query("DELETE FROM outbox")
    suspend fun clear()
}

@Dao
interface PredictionDao {
    @Query("SELECT * FROM prediction_history ORDER BY created_at DESC LIMIT 20")
    suspend fun getRecent(): List<LocalPrediction>

    @Insert
    suspend fun insert(entry: LocalPrediction): Long
}

@Database(
    entities = [
        LocalProduct::class,
        LocalLot::class,
        LocalMarket::class,
        LocalSale::class,
        LocalSaleLine::class,
        Outbox::class,
        LocalPrediction::class,
    ],
    version = 3,
    exportSchema = false,
)
abstract class AppDatabase : RoomDatabase() {
    abstract fun productDao(): ProductDao
    abstract fun lotDao(): LotDao
    abstract fun marketDao(): MarketDao
    abstract fun saleDao(): SaleDao
    abstract fun outboxDao(): OutboxDao
    abstract fun predictionDao(): PredictionDao

    companion object {
        @Volatile
        private var instance: AppDatabase? = null

        private val MIGRATION_1_2 = object : Migration(1, 2) {
            override fun migrate(db: SupportSQLiteDatabase) {
                db.execSQL(
                    """
                    CREATE TABLE IF NOT EXISTS prediction_history (
                        id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
                        created_at TEXT NOT NULL,
                        kind TEXT NOT NULL,
                        input_summary TEXT NOT NULL,
                        result_summary TEXT NOT NULL,
                        model_version TEXT NOT NULL,
                        model_status TEXT NOT NULL,
                        warnings TEXT NOT NULL
                    )
                    """.trimIndent(),
                )
            }
        }

        private val MIGRATION_2_3 = object : Migration(2, 3) {
            override fun migrate(db: SupportSQLiteDatabase) {
                db.execSQL("ALTER TABLE local_lots ADD COLUMN item_id INTEGER NOT NULL DEFAULT 0")
                db.execSQL(
                    """
                    CREATE TABLE IF NOT EXISTS local_markets (
                        id INTEGER PRIMARY KEY NOT NULL,
                        code TEXT NOT NULL,
                        name TEXT NOT NULL
                    )
                    """.trimIndent(),
                )
            }
        }

        fun get(context: Context): AppDatabase =
            instance ?: synchronized(this) {
                instance ?: Room.databaseBuilder(
                    context.applicationContext,
                    AppDatabase::class.java,
                    "avatar.db",
                ).addMigrations(MIGRATION_1_2, MIGRATION_2_3).build().also { instance = it }
            }
    }
}
