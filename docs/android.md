# Android Uygulaması — Çalıştırma (demo)

## Ön koşullar
- JDK 17 + Android SDK 34 kurulu (`C:\Users\MONSTER\dev-tools\android-sdk`).
- Emülatör (`emulator`) + sistem imajı kurulu (aşağıda oluşturulur).

## 1) Emülatörü kur ve başlat (ilk kez)
```bash
# Sistem imajı + emülatör (bir kez)
sdkmanager emulator "system-images/android-34/google_apis/x86_64"

# AVD oluştur (bir kez)
avdmanager create avd -n avatar -k "system-images/android-34/google_apis/x86_64" --device "pixel_5"

# Emülatörü başlat
emulator -avd avatar
```

## 2) Backend'e erişim (emülatörden)
Emülatör içinde `localhost` emülatörün kendisidir; host backend'e ulaşmak için port yönlendir:
```bash
adb reverse tcp:8080 tcp:8080
```
(Bu, `SyncRepository`'nin `http://localhost:8080` taban adresiyle çalışmasını sağlar.)

## 3) APK'yı derle ve kur
```bash
cd mobile && gradlew assembleDebug
adb install app/build/outputs/apk/debug/app-debug.apk
```

## 4) Uygulamada gösterilecekler
- **Products** sekmesi → ürünler (bloklu lotlar devre dışı).
- **Sale** sekmesi → ürün + lot + miktar seç → satış oluştur (senkronize).
- **History** sekmesi → yerel satışlar + senkron durumu.
- Geri çağrılan (bloklu) lotu satmaya çalışınca → **"Blocked lot: sale rejected"** (409).

## Not
- Backend `docker compose` ile veya lokal `go run` ile çalışıyor olmalı (port 8080).
- Web için `http://localhost:5173` (docker) veya `npm run dev`.
