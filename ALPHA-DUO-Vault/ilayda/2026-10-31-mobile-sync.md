# 2026-10-31 — 8080/404 ve mobil senkronizasyon düzeltmesi

## Kök neden

1. `http://localhost:8080/` için Go router’da rota yoktu; API uçları
   (`/health`, `/sync/pull`, `/lots`, …) 200 dönüyordu, yalnızca kök 404’tü.
2. Mobil `DEFAULT_BASE_URL = http://localhost:8080` kullanıyordu. Bu adres
   emülatörde emülatörün kendisidir; host için `10.0.2.2` veya fiziksel cihazda
   `adb reverse tcp:8080 tcp:8080` gerekir. Uygulama adresi otomatik denemiyordu.
3. Pull üç ayrı istekte (`/sync/pull`, `/lots`, `/markets`) yapılıyor ve Room
   tabloları önce `clear()` ediliyordu; ağ hatasında yarım/boş SQLite kalabilirdi.
4. Satış push’u idempotent değildi; aynı satış tekrar gönderilirse stok ikinci
   kez düşebilirdi.
5. 409 alan satış outbox’ta kalıp sonsuz retry ediliyordu.
6. Mobilde alış/üretim/parçalama/lot hareketi/satış/recall rapor ekranı yoktu.

## Yapılanlar

### Backend
- `GET /{$}` eklendi: kök URL artık 404 değil, servis bilgisi döndürüyor.
- `GET /sync/pull` tek snapshot: `server_time`, `products`, `lots`, `markets`,
  `blocked_lot_codes`.
- `GET /reports/summary`: altı rapor bölümü (alış, üretim, parçalama, lot
  hareketleri, satış, recall).
- `CreateSale` idempotent: aynı `code` tekrar gelirse mevcut satış döner; stok
  ikinci kez düşmez.
- Yeni repo metodları: `ListAllStockMovements`, `ListSaleLines`,
  `ListRecallEvents`.
- Testler: `TestRootAndReportSummary`, `TestCreateSaleIsIdempotent`.

### Mobil
- Sunucu adresi otomatik çözülüyor: yapılandırılmış adres → `10.0.2.2:8080`
  (emülatör) → `localhost:8080` (adb reverse). `/health` ile doğrulanıyor.
- Pull tek istek + **tek Room transaction**; hata durumunda eski önbellek korunur.
- Satış push’u satış satış yapılıyor; `code` idempotency anahtarı.
- 409’da satış `rejected` işaretleniyor, outbox’tan siliniyor; retry döngüsü yok.
- “Raporlar” sekmesi: `/reports/summary` altı bölümü gösteriyor; çevrimdışıysa
  açık hata veriyor.
- Ana ekranda bağlantı durumu gösteriliyor.

### Doküman
- `docs/mobile-sync.md`: MySQL = source of truth, SQLite = cache + outbox,
  atomik pull, idempotent push, 409 ve stok uzlaşması.
- `docs/api.md`, `docs/android.md` güncellendi.

## Doğrulama
- Canlı: `/` 200; `/sync/pull` 10 ürün, 20 lot, 9 market; `/reports/summary`
  11 alış, 9 üretim, 9 parçalama, 18 lot hareketi, 9 satış.
- Go unit + integration + vet → temiz.
- Android `assembleDebug` → başarılı.

## Fiziksel cihaz doğrulaması
- Hata bir **fiziksel telefon** üzerindeydi; `localhost` telefonun kendisi,
  `10.0.2.2` yalnızca emülatör içindir.
- `adb reverse tcp:8080 tcp:8080` uygulandı; uygulama yeniden başlatıldı.
- Ana ekran: **“Sunucu: http://localhost:8080 · ürün, lot ve market senkronize
  edildi.”**
- Raporlar ekranı cihazda açıldı: **11 alış**, lot hareketleri ve **9 satış**
  MySQL’den geldi.
- Kablosuz kullanım için telefon + PC aynı Wi-Fi’da olmalı; PC LAN IP’si
  (`http://<PC-IP>:8080`) Windows güvenlik duvarında açık olmalı.
- Uygulamaya **Sunucu → Değiştir** ayarı eklendi; adres kalıcı saklanıyor ve
  **Dene** ile yeniden bağlanılıyor.
- `scripts/adb-reverse.ps1` ve `scripts/adb-reverse-watch.ps1` eklendi; USB
  yeniden bağlandığında reverse’i geri kurar.
- Hata mesajları artık kırmızı, başarı mesajları yeşil gösteriliyor.

