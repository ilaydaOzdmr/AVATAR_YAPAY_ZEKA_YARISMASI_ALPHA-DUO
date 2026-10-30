package com.avatar.app.data

import android.content.Context
import androidx.room.Dao
import androidx.room.Database
import androidx.room.Insert
import androidx.room.OnConflictStrategy
import androidx.room.Query
import androidx.room.Room
import androidx.room.RoomDatabase
import com.avatar.app.data.entities.LocalLot
import com.avatar.app.data.entities.LocalMarket
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

@Database(
    entities = [
        LocalProduct::class,
        LocalLot::class,
        LocalMarket::class,
        LocalSale::class,
        LocalSaleLine::class,
        Outbox::class,
    ],
    version = 2,
    exportSchema = false,
)
abstract class AppDatabase : RoomDatabase() {
    abstract fun productDao(): ProductDao
    abstract fun lotDao(): LotDao
    abstract fun marketDao(): MarketDao
    abstract fun saleDao(): SaleDao
    abstract fun outboxDao(): OutboxDao

    companion object {
        @Volatile
        private var instance: AppDatabase? = null

        fun get(context: Context): AppDatabase =
            instance ?: synchronized(this) {
                instance ?: Room.databaseBuilder(
                    context.applicationContext,
                    AppDatabase::class.java,
                    "avatar.db",
                ).fallbackToDestructiveMigration().build().also { instance = it }
            }
    }
}
