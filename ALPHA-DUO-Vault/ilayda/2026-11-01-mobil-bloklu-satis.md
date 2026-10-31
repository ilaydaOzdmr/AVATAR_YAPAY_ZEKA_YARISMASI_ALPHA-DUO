# 2026-11-01 — Mobilde bloklu lot satış reddi

## Sorun
Bloklu lot satışı backend tarafından reddediliyordu; `submitSale` yerel kaydı
`status = "rejected"` yapıyordu. Ancak History ekranı yalnızca `synced` alanına
baktığı için reddedilen satış yeşil **"Senkron"** olarak görünüyordu. Ana
sayfadaki "Son satış" kartı da reddedilen denemeyi son satış sanabiliyordu.

## Çözüm
- `SaleStatus.kt` eklendi:
  - `rejected` → **"Bloklandı"** (kırmızı)
  - `synced` → "Senkron" (yeşil)
  - aksi halde → "Bekliyor" (amber)
- `History` artık `status` alanını dikkate alıyor.
- `Home` son satışı seçerken reddedilen kayıtları atlıyor
  (`latestAcceptedSale`).
- Reddedilen satış yerelde geçmiş kaydı olarak kalıyor; outbox'a geri
  eklenmiyor, tekrar gönderilmiyor.

## Ek: senkron sonrası durum ve web uyumu
- `push()` ile senkron sırasında 409 alan satış `rejected` kalıyor; `pull()`
  bu durumu silmiyor veya `completed` yapmıyor.
- Reddedilen satışta `saveSale` sırasında düşülen yerel lot miktarı geri
  ekleniyor; stok satış olmamış gibi korunuyor.
- `submitSale` aynı reddedilmiş satış için tekrar çağrılırsa stoğu ikinci kez
  geri eklemiyor (idempotent).
- Web navigasyonundaki sabit **"Geri Çağırma 1"** rozeti kaldırıldı. Rozet ve
  Inventory "Bloke lot" sayısı artık `/sync/pull.blocked_lot_codes` listesinden
  geliyor; mobil ile aynı kaynak.
- Recall oluşturulunca rozet anında yenileniyor; bloke lot yoksa rozet gizli.
- Bildirim toast'ı da sabit metin yerine gerçek bloke lot sayısını gösteriyor.

## Test
- `SaleStatusTest`: reddedilen/senkron/bekleyen etiketleri ve son satış filtresi.
- `BlockedSaleSyncTest`: 409 → `rejected`, yerel stok iadesi ve `pull()`
  sonrası durumun korunması.
- Mobil Gradle birim testleri: **25 test** (mevcut PDF testi Robolectric'te
  zaman zaman 10 sn timeout veriyor; ilgisiz flake).
- `assembleDebug`: başarılı.
- Web: `npm run build` ve 8 Node testi başarılı; canlıda geçici recall ile
  rozet/Inventory 1, temizlik sonrası 0 doğrulandı.
