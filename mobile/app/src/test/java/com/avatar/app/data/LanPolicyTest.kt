package com.avatar.app.data

import android.content.Context
import androidx.room.Room
import androidx.test.core.app.ApplicationProvider
import java.net.InetAddress
import java.net.InetSocketAddress
import java.net.Proxy
import java.util.concurrent.atomic.AtomicReference
import okhttp3.OkHttpClient
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertNull
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner
import org.robolectric.annotation.Config

@RunWith(RobolectricTestRunner::class)
@Config(sdk = [34])
class LanPolicyTest {
    private fun withRepository(test: (SyncRepository) -> Unit) {
        val db = Room.inMemoryDatabaseBuilder(ApplicationProvider.getApplicationContext<Context>(), AppDatabase::class.java).build()
        try { test(SyncRepository(db, baseUrl = "http://localhost:8080")) } finally { db.close() }
    }

    @Test fun discoveryAcceptsOnlyPrivateIpv4Interfaces() = withRepository { repository ->
        val probe = SyncRepository::class.java.getDeclaredMethod("probeAddress", InetAddress::class.java).apply { isAccessible = true }
        for (address in listOf("127.0.0.1", "169.254.1.2", "0.0.0.0", "8.8.8.8", "172.15.1.2", "172.32.1.2", "::1")) {
            assertNull(address, probe.invoke(repository, InetAddress.getByName(address)))
        }
        for (address in listOf("10.1.2.3", "172.16.1.2", "172.31.2.3", "192.168.2.3")) {
            assertNotNull(address, probe.invoke(repository, InetAddress.getByName(address)))
        }
    }

    @Test fun firstHealthyCandidateWinsWithoutContactingRealLanHosts() = withRepository { repository ->
        MockWebServer().use { server ->
            server.start()
            val proxy = Proxy(Proxy.Type.HTTP, InetSocketAddress("127.0.0.1", server.port))
            val client = OkHttpClient.Builder().proxy(proxy).build()
            val probe = SyncRepository::class.java.getDeclaredMethod("probeCandidate", String::class.java, OkHttpClient::class.java, AtomicReference::class.java).apply { isAccessible = true }
            val found = AtomicReference<String?>(null)
            server.enqueue(MockResponse().setResponseCode(503))
            probe.invoke(repository, "candidate.invalid", client, found)
            assertNull(found.get())
            server.enqueue(MockResponse().setResponseCode(200))
            probe.invoke(repository, "candidate.invalid", client, found)
            assertEquals("http://candidate.invalid:8080", found.get())
            probe.invoke(repository, "another.invalid", client, found)
            assertEquals(2, server.requestCount)
        }
    }
}
