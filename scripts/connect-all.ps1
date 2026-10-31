# Emülatör + fiziksel cihaz için tek komut:
#   powershell -ExecutionPolicy Bypass -File scripts/connect-all.ps1
# APK'yı da kurmak için:
#   powershell -ExecutionPolicy Bypass -File scripts/connect-all.ps1 -Install
param([switch]$Install)

$ErrorActionPreference = "Continue"

try {
    $health = Invoke-RestMethod -Uri "http://localhost:8080/health" -TimeoutSec 3
    Write-Output "Backend: $($health.status)"
} catch {
    Write-Warning "Backend 8080 yanit vermiyor. Once: docker compose -f infra/docker-compose.yml up -d"
}

adb wait-for-device
$devices = adb devices | Select-String "`tdevice$"
if (-not $devices) {
    throw "ADB cihazi bulunamadi. Emulatoru baslatin veya USB hata ayiklamayi acin."
}

foreach ($line in $devices) {
    $serial = ($line.Line -split "`t")[0].Trim()
    if ([string]::IsNullOrWhiteSpace($serial)) { continue }
    adb -s $serial reverse tcp:8080 tcp:8080 | Out-Null
    Write-Output "$serial -> localhost:8080"
    if ($Install) {
        $apk = Join-Path $PSScriptRoot "..\mobile\app\build\outputs\apk\debug\app-debug.apk"
        if (Test-Path $apk) {
            adb -s $serial install -r $apk
        } else {
            Write-Warning "APK bulunamadi: $apk"
        }
    }
}

Write-Output "Hazir. Uygulamada Sunucu: http://localhost:8080 kullanin."
