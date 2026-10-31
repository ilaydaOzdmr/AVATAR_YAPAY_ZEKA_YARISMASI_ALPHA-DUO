import java.net.DatagramSocket
import java.net.Inet4Address
import java.net.InetAddress
import java.net.NetworkInterface
import java.util.Properties
import org.gradle.testing.jacoco.tasks.JacocoReport
import org.gradle.testing.jacoco.plugins.JacocoTaskExtension

plugins {
    id("com.android.application")
    id("org.jetbrains.kotlin.android")
    id("com.google.devtools.ksp")
    jacoco
}

// Fiziksel cihaz varsayılanı: derleyen PC'nin LAN IPv4 adresi.
// Öncelik: -PavatarHostIp / AVATAR_HOST_IP / local.properties "avatar.hostIp"
// yoksa ağ arayüzlerinden otomatik seçilir. IP değişirse uygulama açılışta
// yerel ağı tarayıp backend'i kendisi de bulur (bkz. SyncRepository).
fun detectHostIpv4(): String {
    fun override(key: String): String? {
        val fromGradle = project.findProperty(key) as String?
        val fromEnv = System.getenv(key.replace('.', '_').uppercase())
        return (fromGradle ?: fromEnv)?.trim()?.takeIf { it.isNotEmpty() }
    }

    override("avatarHostIp")?.let { return it }
    override("avatar.hostIp")?.let { return it }

    val local = Properties().apply {
        val file = rootProject.file("local.properties")
        if (file.exists()) file.inputStream().use { load(it) }
    }
    local.getProperty("avatar.hostIp")?.trim()?.takeIf { it.isNotEmpty() }?.let { return it }

    // 1) Varsayılan çıkış arayüzü: PC'nin LAN'da kullandığı adres.
    runCatching {
        DatagramSocket().use { socket ->
            socket.connect(InetAddress.getByName("8.8.8.8"), 53)
            (socket.localAddress as? Inet4Address)?.hostAddress
        }
    }.getOrNull()?.takeIf { it != "0.0.0.0" }?.let { return it }

    // 2) Yedek: sanal adaptörleri eleyip Wi-Fi/Ethernet'i tercih et.
    val virtualHints = listOf(
        "vethernet", "vmware", "virtualbox", "vbox", "docker", "wsl", "hyper-v",
        "wan miniport", "tap", "tun", "loopback", "pseudo", "npcap", "bluetooth",
        "local area connection*", "vpn",
    )

    var best: Pair<Int, String>? = null
    runCatching {
        for (nif in NetworkInterface.getNetworkInterfaces()) {
            if (!nif.isUp || nif.isLoopback) continue
            val label = "${nif.name} ${nif.displayName}".lowercase()
            if (virtualHints.any { label.contains(it) }) continue
            val wireless = label.contains("wi-fi") || label.contains("wifi") ||
                label.contains("wireless") || label.contains("wlan")
            val ethernet = label.contains("ethernet") && !label.contains("virtual")
            if (!wireless && !ethernet) continue
            for (address in nif.inetAddresses) {
                if (address !is Inet4Address) continue
                if (address.isLoopbackAddress || address.isLinkLocalAddress) continue
                val host = address.hostAddress ?: continue
                var score = if (wireless) 200 else 150
                score += when {
                    host.startsWith("192.168.") -> 30
                    host.startsWith("10.") -> 20
                    host.startsWith("172.") -> 10
                    else -> 0
                }
                val current = best
                if (current == null || score > current.first) best = score to host
            }
        }
    }
    return best?.second ?: "127.0.0.1"
}

val hostIpv4 = detectHostIpv4()

android {
    namespace = "com.avatar.app"
    compileSdk = 34

    defaultConfig {
        applicationId = "com.avatar.app"
        minSdk = 26
        targetSdk = 34
        versionCode = 1
        versionName = "1.0"

        buildConfigField("String", "DEFAULT_BASE_URL", "\"http://$hostIpv4:8080\"")
    }

    buildFeatures {
        compose = true
        buildConfig = true
    }

    composeOptions {
        kotlinCompilerExtensionVersion = "1.5.14"
    }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }

    testOptions { unitTests.isIncludeAndroidResources = true }

    kotlinOptions {
        jvmTarget = "17"
    }
}

dependencies {
    implementation("androidx.core:core-ktx:1.13.1")
    implementation("androidx.lifecycle:lifecycle-runtime-ktx:2.8.3")
    implementation("androidx.activity:activity-compose:1.9.0")

    implementation(platform("androidx.compose:compose-bom:2024.06.00"))
    implementation("androidx.compose.ui:ui")
    implementation("androidx.compose.ui:ui-tooling-preview")
    implementation("androidx.compose.material3:material3")
    debugImplementation("androidx.compose.ui:ui-tooling")

    implementation("androidx.room:room-runtime:2.6.1")
    implementation("androidx.room:room-ktx:2.6.1")
    ksp("androidx.room:room-compiler:2.6.1")
    testImplementation("junit:junit:4.13.2")
    testImplementation("org.robolectric:robolectric:4.14.1")
    testImplementation("androidx.compose.ui:ui-test-junit4")
    testImplementation("com.squareup.okhttp3:mockwebserver:4.12.0")
    debugImplementation("androidx.compose.ui:ui-test-manifest")

    implementation("com.squareup.okhttp3:okhttp:4.12.0")
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-android:1.8.1")
}

jacoco { toolVersion = "0.8.16" }
tasks.withType<Test>().configureEach {
    // Native test libraries resolve "Windows" with a locale-independent JVM locale.
    systemProperty("user.language", "en")
    systemProperty("user.country", "US")
    maxHeapSize = "2g"
    systemProperty("robolectric.dependency.repo.url", "https://repo.maven.apache.org/maven2")
    extensions.configure<JacocoTaskExtension> {
        isIncludeNoLocationClasses = true
        excludes = listOf("jdk.internal.*")
    }
}
tasks.register<JacocoReport>("jacocoTestReport") {
    dependsOn("testDebugUnitTest")
    reports { xml.required.set(true); html.required.set(true) }
    sourceDirectories.setFrom(files("src/main/java"))
    classDirectories.setFrom(fileTree(layout.buildDirectory.dir("tmp/kotlin-classes/debug")) {
        exclude("**/R.class", "**/R$*.class", "**/BuildConfig.*", "**/*_Impl*")
    })
    executionData.setFrom(fileTree(layout.buildDirectory) {
        include("jacoco/testDebugUnitTest.exec")
    })
}
