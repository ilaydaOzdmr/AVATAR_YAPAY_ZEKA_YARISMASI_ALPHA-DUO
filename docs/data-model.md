# Veri Modeli — AVATAR ALPHA-DUO

MySQL 8, InnoDB, utf8mb4. Para `DECIMAL(12,2)`, miktar/ağırlık `DECIMAL(12,3)`.
Şema: `backend/migrations/0001_init.up.sql`. Seed: `backend/migrations/0001_seed.up.sql`.

## Tablolar

| Tablo | Amaç |
|-------|------|
| `materials` | Hammadde kartları (et, baharat, bağırsak, tuz, bütün kuzu) |
| `products` | Mamul ve parça ürünler (sucuk, but, sırt, ...) |
| `suppliers` | Tedarikçiler |
| `markets` | Müşteri/marketler (geri çağırma iletişim listesi) |
| `lots` | Hammadde ve mamul partileri (lot) |
| `production_orders` | Üretim emirleri + hizmet maliyeti |
| `production_order_inputs` | Emirde kullanılan hammadde lotları (çoktan-çoğa) |
| `production_order_outputs` | Emirden çıkan mamul lotları |
| `butchery_batches` | Bütün kuzu parçalama partileri |
| `butchery_items` | Parçalama sonucu ürünler + fire |
| `lot_links` | Ebeveyn → çocuk lot bağlantıları (izlenebilirlik) |
| `stock_movements` | Lot bazlı stok hareketleri (giriş, sevk, satış, fire) |
| `sales` / `sale_lines` | Satışlar + lot bazlı satırlar |
| `recall_events` / `recall_blocks` | Geri çağırma + bloke lotlar |
| `market_redistributions` | İkinci kademe dağıtım (market → alt alıcı) |
| `sync_outbox` | Mobil senkronizasyon kuyruğu |

## Anahtarlar ve ilişkiler
- `lots.item_type` (`material`/`product`) çok-biçimli olduğu için `item_id` FK taşımaz.
- `lots.supplier_id → suppliers.id`.
- `production_order_inputs` ve `production_order_outputs` çoktan-çoğa lot ilişkisini taşır.
- `lot_links(parent_lot_id, child_lot_id, qty)` izlenebilirliğin omurgasıdır.
- `stock_movements.qty` **işaretlidir**: giriş pozitif, çıkış (sevk/satış) negatif.
- `sale_lines.lot_id` satışın hangi lottan yapıldığını, `stock_movement_id` ilgili stok hareketini bağlar.
- `market_redistributions(market_id, lot_id, sub_buyer, contact, qty)` ikinci kademe
  dağıtımı taşır; izlenebilirlik/geri çağırmada **yalnızca kayıt varsa** gösterilir.

## Hareket tipleri
`stock_movements.movement_type`: `inbound` (giriş), `shipment` (sevk/satış),
`transfer` (depo transferi), `return` (iade), `rework` (yeniden işleme).

## Örnek zincir (ET-2026-001 senaryosu)
```
ET-2026-001 (100 kg et) ─┐
BAH-2026-001 (2 kg)     ─┼─► URT-2026-045 ─► SUC-2026-045 (85 paket) ─┬─► Market A (30)
BAR-2026-001 (5 kg)     ─┘                                             ├─► Market B (25)
                                                                       └─► Market C (30)
```

- Üretim emri: 100 kg et (9,70) + 2 kg baharat (24,00) + 5 kg bağırsak (11,60)
  + 85 × 5,50 hizmet = **1.543,50 ₺ → 18,16 ₺ EKP**.
- Sevk miktarları `stock_movements` (`movement_type='shipment'`, `ref_type='market'`,
  negatif `qty`) üzerinden market bazında okunur.

## İşaretli miktar kuralı
`stock_movements.qty`: giriş `+`, çıkış `-`. Etkilenen market miktarı `ABS(qty)`.

## Ürün raf ömrü, eklenme tarihi ve SKT
- `products` ve `materials` tablolarında:
  - `shelf_life_days` — ürün/hammadde bazlı kullanım süresi (gün).
  - `created_at` — kartın eklenme tarihi (DB otomatik).
- Lot oluştururken:
  - `lots.received_at` = üretim/alış tarihi (ürün lotu için üretim tarihi).
  - `lots.expiry_date` = `received_at + shelf_life_days`.
  - Örnek: üretim 2026-10-01, raf ömrü 15 gün → SKT 2026-10-16.
- `POST /lots` ile lot oluşturulurken `qty > 0` ise aynı transaction içinde
  `stock_movements` (`inbound`) yazılır; lot bakiyesi ve stok defteri tutarlı kalır.
- Üretim emri çıktı lotlarında ve butchery’de backend’in oluşturduğu ürün
  lotlarında SKT, üretim/parti tarihi + ürün raf ömrü olarak hesaplanır.

## Parçalama fire ve maliyet kontrolleri

`butchery_batches` tablosuna migration `0003_butchery_consistency` ile eklenen
alanlar:

- `expected_cost` `DECIMAL(12,2)` — giriş ağırlığı × birim maliyet.
- `cost_difference` `DECIMAL(12,2)` — kayıtlı toplam maliyet − beklenen maliyet.
- `cost_check_status` `VARCHAR(16)` — `ok` / `mismatch` (2 ondalık).
- `weight_check_status` `VARCHAR(16)` — `ok` / `mismatch` (3 ondalık).

`EnsureSchema` bu kolonları eski hacimlere de ekler ve mevcut kayıtları yeniden
hesaplayarak doldurur. Geri alma dosyası:
`backend/migrations/down/0003_butchery_consistency.down.sql`.

Hesaplar:
- `part_weight` = fire olmayan (`is_waste = FALSE`) parçaların toplamı.
- `fire_weight` = `input_weight - part_weight`.
- `expected_cost` = `input_weight × unit_cost`.
- `cost_difference` = `total_cost - expected_cost`.
- Fire tek başına hata değildir; yalnızca negatif fire ağırlık kontrolünü
  `mismatch` yapar. Maliyet ve ağırlık kontrolleri bağımsızdır.

## Geri çağırma alıcı teyitleri

`recall_acknowledgements` tablosu (migration `0004_recall_acknowledgements`):

- `source_lot_id` — geri çağırmanın kaynak lotu.
- `market_code` — teyit edilen market kodu.
- `confirmed_at` — teyit zamanı.
- `(source_lot_id, market_code)` benzersizdir; aynı teyit tekrarı idempotenttir.

Teyitler aynı kaynak lotun tüm geri çağırma kayıtlarında ortaktır. Teyit
satış blokajını kaldırmaz; lot `blocked` durumunda kalır. Aynı lot için yeniden
geri çağırma başlatılırsa yeni kayıt açılmaz, mevcut kayıt döner.


