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
| `sync_outbox` | Mobil senkronizasyon kuyruğu |

## Anahtarlar ve ilişkiler
- `lots.item_type` (`material`/`product`) çok-biçimli olduğu için `item_id` FK taşımaz.
- `lots.supplier_id → suppliers.id`.
- `production_order_inputs` ve `production_order_outputs` çoktan-çoğa lot ilişkisini taşır.
- `lot_links(parent_lot_id, child_lot_id, qty)` izlenebilirliğin omurgasıdır.
- `stock_movements.qty` **işaretlidir**: giriş pozitif, çıkış (sevk/satış) negatif.
- `sale_lines.lot_id` satışın hangi lottan yapıldığını, `stock_movement_id` ilgili stok hareketini bağlar.

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
