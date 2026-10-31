# USB her takildiginda adb reverse'i otomatik kuran izleyiciyi Windows
# oturum acilisina ekler ve hemen baslatir.
#   powershell -ExecutionPolicy Bypass -File scripts/install-adb-reverse-autostart.ps1

$ErrorActionPreference = "Stop"
$watcher = (Resolve-Path (Join-Path $PSScriptRoot "adb-reverse-watch.ps1")).Path
$taskName = "AvatarAdbReverse"

$action = New-ScheduledTaskAction `
    -Execute "powershell.exe" `
    -Argument "-NoProfile -WindowStyle Hidden -ExecutionPolicy Bypass -File `"$watcher`""
$trigger = New-ScheduledTaskTrigger -AtLogOn
$settings = New-ScheduledTaskSettingsSet `
    -AllowStartIfOnBatteries `
    -DontStopIfGoingOnBatteries `
    -RestartCount 3 `
    -RestartInterval (New-TimeSpan -Minutes 1) `
    -ExecutionTimeLimit (New-TimeSpan -Days 3650)

try {
    Register-ScheduledTask `
        -TaskName $taskName `
        -Action $action `
        -Trigger $trigger `
        -Settings $settings `
        -Description "USB yeniden baglandiginda AVATAR adb reverse tcp:8080 yonlendirmesini geri kurar." `
        -Force | Out-Null
    Start-ScheduledTask -TaskName $taskName
    Write-Output "Gorev kaydedildi ve baslatildi: $taskName"
} catch {
    Write-Warning "Zamanlanmis gorev kaydedilemedi: $($_.Exception.Message)"
    $startup = Join-Path $env:APPDATA "Microsoft\Windows\Start Menu\Programs\Startup"
    $shortcut = Join-Path $startup "AvatarAdbReverse.vbs"
    $vbs = @"
Set shell = CreateObject("WScript.Shell")
shell.Run "powershell.exe -NoProfile -WindowStyle Hidden -ExecutionPolicy Bypass -File ""$watcher""", 0, False
"@
    Set-Content -Path $shortcut -Value $vbs -Encoding ASCII
    Write-Output "Baslangic klasorune eklendi: $shortcut"
    Start-Process "powershell.exe" -ArgumentList "-NoProfile -WindowStyle Hidden -ExecutionPolicy Bypass -File `"$watcher`"" -WindowStyle Hidden
}

Write-Output "Durum:"
Get-ScheduledTask -TaskName $taskName -ErrorAction SilentlyContinue |
    Select-Object TaskName, State
