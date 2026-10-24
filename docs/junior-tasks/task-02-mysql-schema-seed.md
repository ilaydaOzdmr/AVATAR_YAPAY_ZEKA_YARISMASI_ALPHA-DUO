# Görev 2 — MySQL şeması + seed verisi (junior)

## Bağlam
Gıda izlenebilirlik sistemi (monorepo). Veritabanı: MySQL 8, InnoDB, utf8mb4.
Tüm tablo/sütun adları İngilizce ve aşağıda birebir verilmiştir.

## Teslimat (3 dosya)
1. `backend/migrations/0001_init.up.sql` — CREATE TABLE'lar + FK + index.
2. `backend/migrations/0001_seed.up.sql` — fiyat kataloğu + senaryo verisi.
3. `docs/data-model.md` — tablolar + ilişkilerin kısa açıklaması.

## Kural
- Para: `DECIMAL(12,2)`. Miktar/ağırlık: `DECIMAL(12,3)`.
- PK: `BIGINT AUTO_INCREMENT`. FK'lar: `BIGINT`.
- InnoDB, `utf8mb4`. Migration sırası: önce init, sonra seed.

## Şema (tablo → sütunlar)

```
materials(id PK, code VARCHAR UNIQUE, name VARCHAR, unit VARCHAR,
          is_raw BOOL, default_purchase_price DECIMAL(12,2))
products(id PK, code VARCHAR UNIQUE, name VARCHAR, unit VARCHAR,
         list_price DECIMAL(12,2), category VARCHAR, is_finished BOOL)
suppliers(id PK, code VARCHAR UNIQUE, name VARCHAR, contact VARCHAR)
markets(id PK, code VARCHAR UNIQUE, name VARCHAR, contact VARCHAR)
lots(id PK, code VARCHAR UNIQUE, item_type ENUM('material','product'),
     item_id BIGINT, qty DECIMAL(12,3), unit VARCHAR,
     supplier_id BIGINT NULL, purchase_price DECIMAL(12,2),
     unit_cost DECIMAL(12,2), received_at DATETIME, expiry_date DATE NULL,
     status VARCHAR, created_at DATETIME, updated_at DATETIME)
production_orders(id PK, code VARCHAR UNIQUE, status VARCHAR,
     started_at DATETIME, finished_at DATETIME NULL, notes TEXT,
     service_name VARCHAR, service_unit_cost DECIMAL(12,2), service_qty DECIMAL(12,3))
production_order_inputs(id PK, production_order_id BIGINT,
     material_lot_id BIGINT, qty DECIMAL(12,3))
production_order_outputs(id PK, production_order_id BIGINT,
     product_lot_id BIGINT, qty DECIMAL(12,3))
butchery_batches(id PK, code VARCHAR UNIQUE, animal_type VARCHAR,
     input_weight DECIMAL(12,3), unit_cost DECIMAL(12,2),
     total_cost DECIMAL(12,2), batch_date DATE, status VARCHAR)
butchery_items(id PK, batch_id BIGINT, product_id BIGINT,
     product_lot_id BIGINT, weight DECIMAL(12,3), is_waste BOOL)
lot_links(id PK, parent_lot_id BIGINT, child_lot_id BIGINT,
     qty DECIMAL(12,3), link_type VARCHAR)
stock_movements(id PK, lot_id BIGINT, movement_type VARCHAR,
     qty DECIMAL(12,3), unit_cost DECIMAL(12,2), total_cost DECIMAL(12,2),
     ref_type VARCHAR, ref_id BIGINT, occurred_at DATETIME, created_at DATETIME)
sales(id PK, code VARCHAR UNIQUE, market_id BIGINT, sold_at DATETIME,
     status VARCHAR, client_id VARCHAR NULL)
sale_lines(id PK, sale_id BIGINT, lot_id BIGINT, qty DECIMAL(12,3),
     unit_price DECIMAL(12,2), total_price DECIMAL(12,2), stock_movement_id BIGINT)
recall_events(id PK, code VARCHAR UNIQUE, reason TEXT,
     source_lot_id BIGINT, created_at DATETIME)
recall_blocks(id PK, recall_event_id BIGINT, lot_id BIGINT, blocked_at DATETIME)
sync_outbox(id PK, client_id VARCHAR, entity VARCHAR, entity_id BIGINT,
     operation VARCHAR, payload JSON, created_at DATETIME, synced_at DATETIME NULL)
```

## FK'lar
- `production_order_inputs.production_order_id → production_orders.id`; `.material_lot_id → lots.id`
- `production_order_outputs.production_order_id → production_orders.id`; `.product_lot_id → lots.id`
- `butchery_items.batch_id → butchery_batches.id`; `.product_id → products.id`; `.product_lot_id → lots.id`
- `lot_links.parent_lot_id → lots.id`; `.child_lot_id → lots.id`
- `stock_movements.lot_id → lots.id`
- `sales.market_id → markets.id`
- `sale_lines.sale_id → sales.id`; `.lot_id → lots.id`; `.stock_movement_id → stock_movements.id`
- `recall_events.source_lot_id → lots.id`
- `recall_blocks.recall_event_id → recall_events.id`; `.lot_id → lots.id`
- `lots.supplier_id → suppliers.id` (NULL olabilir)
- Not: `lots.item_id` çok-biçimli (material|product) olduğu için FK yok.

## Index'ler
`lot_links(parent_lot_id)`, `lot_links(child_lot_id)`, `stock_movements(lot_id)`,
`sale_lines(lot_id)`, `production_order_inputs(production_order_id)`,
`production_order_outputs(production_order_id)`.

## Seed verisi (birebir, PDF ile tutarlı)

### materials (is_raw=1)
| code | name | unit | default_purchase_price |
|------|------|------|------------------------|
| MAT-001 | Et (dana) | kg | 9.70 |
| MAT-002 | Baharat karışımı | kg | 24.00 |
| MAT-003 | Bağırsak (kuzu) | kg | 11.60 |
| MAT-004 | Tuz | kg | 3.50 |
| ANM-001 | Bütün kuzu | kg | 11.50 |

### products (is_finished=1)
| code | name | unit | list_price | category |
|------|------|------|-----------|----------|
| PRD-001 | Narin Sucuk (1 kg) | paket | 24.90 | finished |
| PRD-101 | But | kg | 28.00 | butchery |
| PRD-102 | Sırt | kg | 32.00 | butchery |
| PRD-103 | Omuz | kg | 22.00 | butchery |
| PRD-104 | Kaburga | kg | 18.00 | butchery |
| PRD-105 | Kıyma | kg | 24.00 | butchery |
| PRD-106 | Kuşbaşı | kg | 26.00 | butchery |
| PRD-107 | Kemik / yan ürün | kg | 2.00 | butchery |

### suppliers
`SUP-001` "Et Tedarikçisi", `SUP-002` "Baharat Tedarikçisi", `SUP-003` "Bağırsak Tedarikçisi" (contact: telefon örnek).

### markets (geri çağırma iletişim listesi için contact şart)
`MKT-A` "Market A" + contact, `MKT-B` "Market B" + contact, `MKT-C` "Market C" + contact.

### lots (senaryo)
- `ET-2026-001` material, MAT-001, qty=100, unit='kg', supplier SUP-001, purchase_price=9.70, unit_cost=9.70
- `BAH-2026-001` material, MAT-002, qty=2, supplier SUP-002, unit_cost=24.00
- `BAR-2026-001` material, MAT-003, qty=5, supplier SUP-003, unit_cost=11.60
- `SUC-2026-045` product, PRD-001, qty=85, unit='paket', unit_cost=18.16

### production_orders
`URT-2026-045` status='completed', service_name='Sucuk üretim hizmeti',
service_unit_cost=5.50, service_qty=85.

### production_order_inputs
(URT-2026-045 → ET-2026-001, 100), (→ BAH-2026-001, 2), (→ BAR-2026-001, 5)

### production_order_outputs
(URT-2026-045 → SUC-2026-045, 85)

### lot_links (parent material lot → child product lot)
(ET-2026-001 → SUC-2026-045, 100), (BAH-2026-001 → SUC-2026-045, 2),
(BAR-2026-001 → SUC-2026-045, 5)

### stock_movements
- Giriş: ET-2026-001 (+100), BAH-2026-001 (+2), BAR-2026-001 (+5) — movement_type='inbound'
- Sevk: SUC-2026-045 → MKT-A 30, MKT-B 25, MKT-C 30 — movement_type='shipment', ref_type='market'

## Kabul kriteri (senior kontrol edecek)
- `0001_init.up.sql` ve `0001_seed.up.sql` sırayla hatasız çalışır.
- 17 tablo oluşur.
- `SELECT * FROM lots` senaryo lotlarını (ET-2026-001, SUC-2026-045 vs.) gösterir.
- Fiyat kataloğu ve senaryo yukarıdaki değerlerle birebir.
