# USB ile bağlı fiziksel cihazda backend'e erişim.
# Telefonun localhost:8080 adresini PC'ye yönlendirir.
adb wait-for-device
adb reverse tcp:8080 tcp:8080
adb reverse --list
