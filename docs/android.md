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
- **Tahmin** sekmesi → üretim girdileri, verim/EKP; kuzu/sığır parçalama
  tahmini; model sürümü, uyarılar ve tahmin geçmişi.
- **History** sekmesi → yerel satışlar + senkron durumu.
- Geri çağrılan (bloklu) lotu satmaya çalışınca → **"Blocked lot: sale rejected"** (409).

## 5) Tahmin ekranı
- Tahminler Go backend üzerinden alınır: `http://localhost:8080/ml/predict/v2/...`.
- ML servisine doğrudan bağlanılmaz.
- Sunucuya ulaşılamazsa ekranda açık **“Çevrimdışı: tahmin için sunucuya
  ulaşılamadı”** mesajı gösterilir; sessizce eski/rastgele tahmin üretilmez.
- Model hazır değilse (503) **“Model hazır değil”** gösterilir.
- Başarılı tahminler Room `prediction_history` tablosuna yazılır; model sürümü
  ve `synthetic_only` durumu geçmişte de görünür.

## 6) Sunucu adresi ve raporlar
- Uygulama `/health` ile erişilebilir backend’i otomatik bulur:
  emülatörde `http://10.0.2.2:8080`, fiziksel cihazda `adb reverse` sonrası
  `http://localhost:8080`.
- **Fiziksel cihaz:** `10.0.2.2` yalnızca emülatör içindir. USB ile bağlı cihazda
  `adb reverse tcp:8080 tcp:8080` çalıştırın. Kablosuz kullanacaksanız telefonla
  PC aynı Wi-Fi’da olmalı ve PC’nin LAN IP’si (`http://<PC-IP>:8080`) Windows
  güvenlik duvarında açık olmalıdır.
- Ana ekrandaki **Sunucu → Değiştir** ile adres uygulamadan girilebilir ve
  kalıcı olarak saklanır. **Dene** bağlantıyı yeniden dener.
- USB her yeniden bağlandığında `adb reverse` kaybolur; tüm cihazlar için
  `scripts/connect-all.ps1` veya `scripts/adb-reverse-watch.ps1` kullanın.
  Emülatör de ADB cihazı olduğu için aynı script ikisinde de çalışır.
- **Raporlar** sekmesi MySQL’den `GET /reports/summary` ile alış, üretim,
  parçalama, lot hareketleri, satış ve geri çağırma bölümlerini gösterir.
- Senkronizasyon modeli: `docs/mobile-sync.md`.

## Not
- Backend `docker compose` ile veya lokal `go run` ile çalışıyor olmalı (port 8080).
- Web için `http://localhost:5173` (docker) veya `npm run dev`.
