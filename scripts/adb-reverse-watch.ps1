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

    $candidates = @()
    if ($env:ANDROID_HOME) { $candidates += (Join-Path $env:ANDROID_HOME "platform-tools\adb.exe") }
    if ($env:ANDROID_SDK_ROOT) { $candidates += (Join-Path $env:ANDROID_SDK_ROOT "platform-tools\adb.exe") }
    if ($env:LOCALAPPDATA) { $candidates += (Join-Path $env:LOCALAPPDATA "Android\Sdk\platform-tools\adb.exe") }
    if ($env:USERPROFILE) { $candidates += (Join-Path $env:USERPROFILE "AppData\Local\Android\Sdk\platform-tools\adb.exe") }
    # Android Studio'nun yaygin kurulum yerleri
    $candidates += "$env:ProgramFiles\Android\Android Studio\platform-tools\adb.exe"
    $candidates += "${env:ProgramFiles(x86)}\Android\Android Studio\platform-tools\adb.exe"
    $candidates += "C:\Android\platform-tools\adb.exe"

    foreach ($candidate in $candidates) {
        if ($candidate -and (Test-Path $candidate)) { return $candidate }
    }
    return $null
}

Write-Log "Izleyici basladi."

while ($true) {
    try {
        $adb = Find-Adb
        if (-not $adb) {
            Write-Log "adb bulunamadi; 10 sn sonra tekrar denenecek."
            Start-Sleep -Seconds 10
            continue
        }
        $devices = & $adb devices 2>$null | Select-String "`tdevice$"
        foreach ($line in $devices) {
            $serial = ($line.Line -split "`t")[0].Trim()
            if ([string]::IsNullOrWhiteSpace($serial)) { continue }
            $reverse = & $adb -s $serial reverse --list 2>$null
            if (-not ($reverse -match "tcp:8080")) {
                & $adb -s $serial reverse tcp:8080 tcp:8080 2>$null | Out-Null
                Write-Log "$serial -> tcp:8080 reverse kuruldu (adb=$adb)"
            }
        }
    } catch {
        Write-Log "Hata: $($_.Exception.Message)"
    }
    Start-Sleep -Seconds 2
}
