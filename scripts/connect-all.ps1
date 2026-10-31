# Emülatör + fiziksel cihaz için tek komut:
#   powershell -ExecutionPolicy Bypass -File scripts/connect-all.ps1
# APK'yı da kurmak için:
#   powershell -ExecutionPolicy Bypass -File scripts/connect-all.ps1 -Install
param([switch]$Install)

$ErrorActionPreference = "Continue"

function Find-Adb {
    $command = Get-Command adb -ErrorAction SilentlyContinue
    if ($command) { return $command.Source }
    $candidates = @(
        (Join-Path $env:ANDROID_HOME "platform-tools\adb.exe"),
        (Join-Path $env:ANDROID_SDK_ROOT "platform-tools\adb.exe"),
        (Join-Path $env:LOCALAPPDATA "Android\Sdk\platform-tools\adb.exe")
    )
    foreach ($candidate in $candidates) {
        if ($candidate -and (Test-Path $candidate)) { return $candidate }
    }
    return $null
}

try {
    $health = Invoke-RestMethod -Uri "http://localhost:8080/health" -TimeoutSec 3
    Write-Output "Backend: $($health.status)"
} catch {
    Write-Warning "Backend 8080 yanit vermiyor. Once: docker compose -f infra/docker-compose.yml up -d"
}

$lanIps = Get-NetIPAddress -AddressFamily IPv4 |
    Where-Object { $_.IPAddress -notlike "127.*" -and $_.IPAddress -notlike "169.254.*" } |
    Select-Object -ExpandProperty IPAddress
Write-Output "PC LAN IP adaylari (Wi-Fi icin): $($lanIps -join ', ')"
Write-Output "Windows guvenlik duvari kapaliysa: scripts/allow-backend-firewall.ps1 (yonetici)"

$adb = Find-Adb
if (-not $adb) {
    throw "adb bulunamadi. Android SDK platform-tools kurulu mu?"
}

& $adb wait-for-device
$devices = & $adb devices | Select-String "`tdevice$"
if (-not $devices) {
    throw "ADB cihazi bulunamadi. Emulatoru baslatin veya USB hata ayiklamayi acin."
}

foreach ($line in $devices) {
    $serial = ($line.Line -split "`t")[0].Trim()
    if ([string]::IsNullOrWhiteSpace($serial)) { continue }
    & $adb -s $serial reverse tcp:8080 tcp:8080 | Out-Null
    Write-Output "$serial -> adb reverse tcp:8080 kuruldu (http://127.0.0.1:8080)"
    if ($Install) {
        $apk = Join-Path $PSScriptRoot "..\mobile\app\build\outputs\apk\debug\app-debug.apk"
        if (Test-Path $apk) {
            & $adb -s $serial install -r $apk
        } else {
            Write-Warning "APK bulunamadi: $apk"
        }
    }
}

Write-Output "Hazir. Uygulama sunucuyu kendisi bulur; gerekirse 'Dene' ile yeniden arar."
