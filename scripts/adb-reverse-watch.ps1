# USB koptuğunda/yeniden bağlandığında adb reverse'i otomatik geri kurar.
# Kullanım: PowerShell'de ayrı bir pencerede çalıştırın.
while ($true) {
    $devices = adb devices | Select-String "`tdevice$"
    if ($devices) {
        $reverse = adb reverse --list 2>$null
        if (-not ($reverse -match "tcp:8080")) {
            adb reverse tcp:8080 tcp:8080 | Out-Null
            Write-Output "$(Get-Date -Format HH:mm:ss) adb reverse kuruldu"
        }
    }
    Start-Sleep -Seconds 3
}
