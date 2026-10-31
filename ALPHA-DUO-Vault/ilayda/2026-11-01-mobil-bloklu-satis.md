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

## Test
- `SaleStatusTest`: reddedilen/senkron/bekleyen etiketleri ve son satış filtresi.
- Mobil Gradle birim testleri: **24/24 PASS**.
- `assembleDebug`: başarılı.
