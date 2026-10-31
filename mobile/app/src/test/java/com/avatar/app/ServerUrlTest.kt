package com.avatar.app

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

class ServerUrlTest {
    @Test fun emptyInputDoesNotReplaceSavedServer() { assertNull(normalizeServerUrl("  ")) }
    @Test fun operatorHostGetsLocalDevelopmentScheme() {
        assertEquals("http://localhost:8080", normalizeServerUrl(" localhost:8080 "))
    }
    @Test fun explicitHttpsIsPreserved() {
        assertEquals("https://backend.example", normalizeServerUrl("https://backend.example"))
    }
}
