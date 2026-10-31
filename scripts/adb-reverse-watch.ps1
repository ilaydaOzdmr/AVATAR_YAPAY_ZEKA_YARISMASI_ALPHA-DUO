# USB kopup geri takildiginda adb reverse'i otomatik geri kurar.
# Emulator + fiziksel cihaz ayni anda desteklenir.
# Kurulum: scripts/install-adb-reverse-autostart.ps1

$mutex = New-Object System.Threading.Mutex($false, "Global\AvatarAdbReverseWatcher")
if (-not $mutex.WaitOne(0)) {
    Write-Output "AvatarAdbReverseWatcher zaten calisiyor."
    exit 0
}

$log = Join-Path $env:TEMP "avatar-adb-reverse.log"
function Write-Log([string]$message) {
    $line = "$(Get-Date -Format 'yyyy-MM-dd HH:mm:ss') $message"
    Add-Content -Path $log -Value $line -Encoding UTF8
}

function Find-Adb {
    $command = Get-Command adb -ErrorAction SilentlyContinue
    if ($command) { return $command.Source }
    $candidates = @(
        (Join-Path $env:ANDROID_HOME "platform-tools\adb.exe"),
        (Join-Path $env:ANDROID_SDK_ROOT "platform-tools\adb.exe"),
        (Join-Path $env:LOCALAPPDATA "Android\Sdk\platform-tools\adb.exe"),
        "C:\Users\MONSTER\dev-tools\android-sdk\platform-tools\adb.exe"
    )
    foreach ($candidate in $candidates) {
        if ($candidate -and (Test-Path $candidate)) { return $candidate }
    }
    return $null
}

$adb = Find-Adb
if (-not $adb) {
    Write-Log "adb bulunamadi; izleyici durdu."
    exit 1
}
Write-Log "Izleyici basladi. adb=$adb"

while ($true) {
    try {
        $devices = & $adb devices 2>$null | Select-String "`tdevice$"
        foreach ($line in $devices) {
            $serial = ($line.Line -split "`t")[0].Trim()
            if ([string]::IsNullOrWhiteSpace($serial)) { continue }
            $reverse = & $adb -s $serial reverse --list 2>$null
            if (-not ($reverse -match "tcp:8080")) {
                & $adb -s $serial reverse tcp:8080 tcp:8080 2>$null | Out-Null
                Write-Log "$serial -> tcp:8080 reverse kuruldu"
            }
        }
    } catch {
        Write-Log "Hata: $($_.Exception.Message)"
    }
    Start-Sleep -Seconds 2
}
