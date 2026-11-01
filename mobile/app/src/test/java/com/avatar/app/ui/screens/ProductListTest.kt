package com.avatar.app.ui.screens

import androidx.compose.material3.MaterialTheme
import androidx.compose.ui.test.assertIsDisplayed
import androidx.compose.ui.test.junit4.createComposeRule
import androidx.compose.ui.test.onNodeWithText
import androidx.compose.ui.test.performScrollTo
import com.avatar.app.data.entities.LocalLot
import com.avatar.app.data.entities.LocalProduct
import org.junit.Rule
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner
import org.robolectric.annotation.Config

@RunWith(RobolectricTestRunner::class)
@Config(sdk = [34], qualifiers = "w800dp-h1280dp-mdpi")
class ProductListTest {
    @get:Rule val compose = createComposeRule()

    @Test
    fun productAndLotWithSameIdDoNotCollideInLazyColumn() {
        val product = LocalProduct(id = 3, code = "PRD-3", name = "Test Ürün", unit = "kg", listPrice = 10.0)
        val lot = LocalLot(id = 3, code = "LOT-3", itemType = "product", itemId = 3, qty = 5.0, unit = "kg", blocked = false)

        compose.setContent {
            MaterialTheme {
                ProductList(listOf(product), listOf(lot)) {}
            }
        }

        compose.onNodeWithText("PRD-3 · Test Ürün").performScrollTo().assertIsDisplayed()
        compose.onNodeWithText("LOT-3").performScrollTo().assertIsDisplayed()
    }
}
