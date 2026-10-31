# Android Uygulaması — Çalıştırma (demo)

## Ön koşullar
- JDK 17+ ve Android SDK 34 kurulu (`mobile/local.properties` içindeki `sdk.dir`).
- Backend 8080 portunda çalışıyor olmalı (`docker compose -f infra/docker-compose.yml up -d`).

## 1) Fiziksel cihaz (varsayılan)
Uygulamanın varsayılan sunucu adresi, **APK derlenirken PC'nin LAN IP'sinden**
otomatik üretilir (`BuildConfig.DEFAULT_BASE_URL`). Telefon aynı Wi-Fi ağındaysa
ek ayar gerekmez.

```bash
cd mobile && gradlew assembleDebug
adb install -r app/build/outputs/apk/debug/app-debug.apk
```

Gerekli olabilecek tek seferlik ayarlar:
1. **Windows güvenlik duvarı** (yönetici olarak):
   ```powershell
   powershell -ExecutionPolicy Bypass -File scripts/allow-backend-firewall.ps1
   ```
2. **USB ile kullanacaksanız** `adb reverse` (Wi-Fi yoksa tek yol):
   ```bash
   adb reverse tcp:8080 tcp:8080
   ```
   USB her takıldığında otomatik kurmak için bir kez:
   `powershell -ExecutionPolicy Bypass -File scripts/install-adb-reverse-autostart.ps1`

Uygulama açılışta şu adresleri **paralel** dener, ilk cevap vereni kullanır:
1. Kaydedilmiş/elle girilen adres
2. Derleme sırasında gömülen PC LAN IP'si (`http://<PC-IP>:8080`)
3. Emülatör köprüsü (`http://10.0.2.2:8080`)
4. USB `adb reverse` (`http://127.0.0.1:8080`)

Hiçbiri yanıt vermezse telefonun bulunduğu özel ağda (`/24`) `/health`
taraması yapılır; böylece DHCP ile PC IP'si değişse bile backend bulunur.
LAN IP'sini elle sabitlemek için `mobile/local.properties` dosyasına
`avatar.hostIp=192.168.1.50` ekleyin veya `-PavatarHostIp=...` verin.

## 2) Emülatör
```bash
sdkmanager emulator "system-images/android-34/google_apis/x86_64"
avdmanager create avd -n avatar -k "system-images/android-34/google_apis/x86_64" --device "pixel_5"
emulator -avd avatar
```
Emülatörde `localhost` emülatörün kendisidir; uygulama `10.0.2.2:8080`
üzerinden host backend'e ulaşır. Gerekirse `adb reverse tcp:8080 tcp:8080`
da çalışır. Aynı APK hem fiziksel cihazda hem emülatörde kullanılabilir.

## 3) APK'yı derle ve kur
```bash
cd mobile && gradlew assembleDebug
adb install -r app/build/outputs/apk/debug/app-debug.apk
```

## 4) Uygulamada gösterilecekler
- **Products** sekmesi → ürünler (bloklu lotlar devre dışı).
- **Sale** sekmesi → ürün + lot + miktar seç → satış oluştur (senkronize).
- **Tahmin** sekmesi → üretim girdileri, verim/EKP; kuzu/sığır parçalama
  tahmini; model sürümü, uyarılar ve tahmin geçmişi.
- **History** sekmesi → yerel satışlar + senkron durumu.
- Geri çağrılan (bloklu) lotu satmaya çalışınca → **"Blocked lot: sale rejected"** (409).

## 5) Tahmin ekranı
- Tahminler Go backend üzerinden alınır: `http://<sunucu>:8080/ml/predict/v2/...`.
- ML servisine doğrudan bağlanılmaz.
- Sunucuya ulaşılamazsa ekranda açık **“Çevrimdışı: tahmin için sunucuya
  ulaşılamadı”** mesajı gösterilir; sessizce eski/rastgele tahmin üretilmez.
- Model hazır değilse (503) **“Model hazır değil”** gösterilir.
- Başarılı tahminler Room `prediction_history` tablosuna yazılır; model sürümü
  ve `synthetic_only` durumu geçmişte de görünür.

## 6) Sunucu adresi ve raporlar
- Ana ekranda **Sunucu: ...** satırı o an kullanılan adresi gösterir.
  **Dene** bağlantıyı sıfırlayıp yeniden arar, **Değiştir** ile adres elle
  girilebilir ve kalıcı olarak saklanır.
- Fiziksel cihazda Wi-Fi için PC LAN IP'si + güvenlik duvarı kuralı; USB için
  `adb reverse tcp:8080 tcp:8080` gerekir. `10.0.2.2` yalnızca emülatör içindir.
- Eski bir MySQL volume'u kullanılıyorsa backend açılışta `EnsureSchema` ile
  eksik kolonları (ör. `production_orders.expected_qty`) kendisi ekler.
- **Raporlar** sekmesi MySQL’den `GET /reports/summary` ile alış, üretim,
  parçalama, lot hareketleri, satış ve geri çağırma bölümlerini gösterir.
- Senkronizasyon modeli: `docs/mobile-sync.md`.

## Not
- Backend `docker compose` ile veya lokal `go run` ile çalışıyor olmalı (port 8080).
- Web için `http://localhost:5173` (docker) veya `npm run dev`.
