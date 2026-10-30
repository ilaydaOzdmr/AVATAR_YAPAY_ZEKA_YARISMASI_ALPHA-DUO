-- 0001_seed.up.sql — AVATAR ALPHA-DUO seed data
-- Price catalog + ET-2026-001 traceability scenario (matches PDF)
SET NAMES utf8mb4;

-- Raw materials
INSERT INTO materials (id, code, name, unit, is_raw, default_purchase_price) VALUES
  (1, 'MAT-001', 'Et (dana)', 'kg', 1, 9.70),
  (2, 'MAT-002', 'Baharat karışımı', 'kg', 1, 24.00),
  (3, 'MAT-003', 'Bağırsak (kuzu)', 'kg', 1, 11.60),
  (4, 'MAT-004', 'Tuz', 'kg', 1, 3.50),
  (5, 'ANM-001', 'Bütün kuzu', 'kg', 1, 11.50);

-- Products (finished + butchery parts)
INSERT INTO products (id, code, name, unit, list_price, category, is_finished) VALUES
  (1, 'PRD-001', 'Narin Sucuk (1 kg)', 'paket', 24.90, 'finished', 1),
  (2, 'PRD-101', 'But', 'kg', 28.00, 'butchery', 1),
  (3, 'PRD-102', 'Sırt', 'kg', 32.00, 'butchery', 1),
  (4, 'PRD-103', 'Omuz', 'kg', 22.00, 'butchery', 1),
  (5, 'PRD-104', 'Kaburga', 'kg', 18.00, 'butchery', 1),
  (6, 'PRD-105', 'Kıyma', 'kg', 24.00, 'butchery', 1),
  (7, 'PRD-106', 'Kuşbaşı', 'kg', 26.00, 'butchery', 1),
  (8, 'PRD-107', 'Kemik / yan ürün', 'kg', 2.00, 'butchery', 1);

-- Suppliers
INSERT INTO suppliers (id, code, name, contact) VALUES
  (1, 'SUP-001', 'Et Tedarikçisi', '0212 000 00 01'),
  (2, 'SUP-002', 'Baharat Tedarikçisi', '0212 000 00 02'),
  (3, 'SUP-003', 'Bağırsak Tedarikçisi', '0212 000 00 03');

-- Markets
INSERT INTO markets (id, code, name, contact) VALUES
  (1, 'MKT-A', 'Market A', '0212 111 11 01'),
  (2, 'MKT-B', 'Market B', '0212 111 11 02'),
  (3, 'MKT-C', 'Market C', '0212 111 11 03');

-- Lots (traceability scenario)
INSERT INTO lots (id, code, item_type, item_id, qty, unit, supplier_id, purchase_price, unit_cost, received_at, expiry_date, status) VALUES
  (1, 'ET-2026-001', 'material', 1, 100.000, 'kg', 1, 9.70, 9.70, '2026-10-19 08:00:00', '2026-10-29', 'active'),
  (2, 'BAH-2026-001', 'material', 2, 2.000, 'kg', 2, 24.00, 24.00, '2026-10-19 08:00:00', NULL, 'active'),
  (3, 'BAR-2026-001', 'material', 3, 5.000, 'kg', 3, 11.60, 11.60, '2026-10-19 08:00:00', NULL, 'active'),
  (4, 'SUC-2026-045', 'product', 1, 85.000, 'paket', NULL, 0.00, 18.16, '2026-10-19 17:00:00', '2026-11-18', 'active');

-- Production order (contract manufacturing)
INSERT INTO production_orders (id, code, status, started_at, finished_at, notes, service_name, service_unit_cost, service_qty, expected_qty) VALUES
  (1, 'URT-2026-045', 'completed', '2026-10-19 09:00:00', '2026-10-19 17:00:00', 'Sözleşmeli üretim: 100 kg et, 2 kg baharat, 5 kg bağırsak', 'Sucuk üretim hizmeti', 5.50, 85.000, 90.000);

-- Production order inputs
INSERT INTO production_order_inputs (production_order_id, material_lot_id, qty) VALUES
  (1, 1, 100.000),
  (1, 2, 2.000),
  (1, 3, 5.000);

-- Production order outputs
INSERT INTO production_order_outputs (production_order_id, product_lot_id, qty) VALUES
  (1, 4, 85.000);

-- Lot links (parent material lot -> child product lot)
INSERT INTO lot_links (parent_lot_id, child_lot_id, qty, link_type) VALUES
  (1, 4, 100.000, 'production'),
  (2, 4, 2.000, 'production'),
  (3, 4, 5.000, 'production');

-- Stock movements (signed qty: positive inflow, negative outflow)
INSERT INTO stock_movements (lot_id, movement_type, qty, unit_cost, total_cost, ref_type, ref_id, occurred_at) VALUES
  (1, 'inbound', 100.000, 9.70, 970.00, 'supplier', 1, '2026-10-19 08:00:00'),
  (2, 'inbound', 2.000, 24.00, 48.00, 'supplier', 2, '2026-10-19 08:00:00'),
  (3, 'inbound', 5.000, 11.60, 58.00, 'supplier', 3, '2026-10-19 08:00:00'),
  (4, 'shipment', -30.000, 18.16, 544.80, 'market', 1, '2026-10-20 10:00:00'),
  (4, 'shipment', -25.000, 18.16, 454.00, 'market', 2, '2026-10-20 11:00:00'),
  (4, 'shipment', -30.000, 18.16, 544.80, 'market', 3, '2026-10-20 12:00:00'),
  -- Diğer hareket tipleri (depo transferi / iade / yeniden işleme)
  (4, 'transfer', -5.000, 18.16, 90.80, 'warehouse', NULL, '2026-10-21 09:00:00'),
  (4, 'return', 2.000, 18.16, 36.32, 'market', 1, '2026-10-21 14:00:00'),
  (1, 'rework', -3.000, 9.70, 29.10, NULL, NULL, '2026-10-22 10:00:00');

-- Second-level distribution (market -> sub-buyer), shown only when recorded
INSERT INTO market_redistributions (market_id, lot_id, sub_buyer, contact, qty) VALUES
  (1, 4, 'Restoran X', '0212 222 22 01', 10.000),
  (1, 4, 'Kasap Y', '0212 222 22 02', 5.000);
