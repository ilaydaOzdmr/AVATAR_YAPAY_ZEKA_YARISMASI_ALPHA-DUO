# USB koptuğunda/yeniden bağlandığında adb reverse'i tüm cihazlara geri kurar.
# Emülatör ve fiziksel cihaz aynı anda desteklenir.
while ($true) {
    $devices = adb devices | Select-String "`tdevice$"
    foreach ($line in $devices) {
        $serial = ($line.Line -split "`t")[0].Trim()
        if ([string]::IsNullOrWhiteSpace($serial)) { continue }
        $reverse = adb -s $serial reverse --list 2>$null
        if (-not ($reverse -match "tcp:8080")) {
            adb -s $serial reverse tcp:8080 tcp:8080 | Out-Null
            Write-Output "$(Get-Date -Format HH:mm:ss) $serial -> reverse kuruldu"
        }
    }
    Start-Sleep -Seconds 3
}
