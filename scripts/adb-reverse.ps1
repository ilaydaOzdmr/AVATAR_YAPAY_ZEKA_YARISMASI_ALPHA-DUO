# Tek cihaz için hızlı yardımcı; tüm cihazlar için connect-all.ps1 kullanın.
adb wait-for-device
adb reverse tcp:8080 tcp:8080
adb reverse --list
