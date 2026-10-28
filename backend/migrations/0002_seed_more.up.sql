-- 0002_seed_more.up.sql — ek senaryo verisi (tüm tablolarda 9+ kayıt)
SET NAMES utf8mb4;

-- Ek hammaddeler (toplam 10)
INSERT INTO materials (id, code, name, unit, is_raw, default_purchase_price) VALUES
  (6,  'MAT-005', 'Sarımsak', 'kg', 1, 18.00),
  (7,  'MAT-006', 'Kimyon', 'kg', 1, 45.00),
  (8,  'MAT-007', 'Pul biber', 'kg', 1, 32.00),
  (9,  'MAT-008', 'Kaya tuzu', 'kg', 1, 2.80),
  (10, 'ANM-002', 'Bütün dana', 'kg', 1, 13.20);

-- Ek ürünler (toplam 10)
INSERT INTO products (id, code, name, unit, list_price, category, is_finished) VALUES
  (9,  'PRD-002', 'Sucuk (500 g)', 'paket', 13.90, 'finished', 1),
  (10, 'PRD-108', 'Kuzu incik', 'kg', 30.00, 'butchery', 1);

-- Ek tedarikçiler (toplam 9)
INSERT INTO suppliers (id, code, name, contact) VALUES
  (4, 'SUP-004', 'Sarımsak Tedarikçisi', '0212 000 00 04'),
  (5, 'SUP-005', 'Baharat Evren', '0212 000 00 05'),
  (6, 'SUP-006', 'Kırmızı Et A.Ş.', '0212 000 00 06'),
  (7, 'SUP-007', 'Kuzu Yetiştiricisi', '0212 000 00 07'),
  (8, 'SUP-008', 'Dana Çiftliği', '0212 000 00 08'),
  (9, 'SUP-009', 'Ambalaj Tedarikçisi', '0212 000 00 09');

-- Ek marketler (toplam 9)
INSERT INTO markets (id, code, name, contact) VALUES
  (4, 'MKT-D', 'Market D', '0212 111 11 04'),
  (5, 'MKT-E', 'Market E', '0212 111 11 05'),
  (6, 'MKT-F', 'Market F', '0212 111 11 06'),
  (7, 'MKT-G', 'Market G', '0212 111 11 07'),
  (8, 'MKT-H', 'Market H', '0212 111 11 08'),
  (9, 'MKT-I', 'Market I', '0212 111 11 09');

-- Ek lotlar: hammadde (id 5-12) + mamul/parça (id 13-20)
INSERT INTO lots (id, code, item_type, item_id, qty, unit, supplier_id, purchase_price, unit_cost, received_at, expiry_date, status) VALUES
  (5,  'ET-2026-002',  'material', 1,  120.000, 'kg', 6, 9.70, 9.70, '2026-10-21 08:00:00', '2026-11-01', 'active'),
  (6,  'ET-2026-003',  'material', 1,  90.000,  'kg', 6, 9.80, 9.80, '2026-10-22 08:00:00', '2026-11-02', 'active'),
  (7,  'BAH-2026-002', 'material', 2,  3.000,   'kg', 5, 24.00, 24.00, '2026-10-21 08:00:00', NULL, 'active'),
  (8,  'BAR-2026-002', 'material', 3,  6.000,   'kg', 3, 11.60, 11.60, '2026-10-21 08:00:00', NULL, 'active'),
  (9,  'SAR-2026-001', 'material', 6,  1.500,   'kg', 4, 18.00, 18.00, '2026-10-21 08:00:00', NULL, 'active'),
  (10, 'KUZ-2026-001', 'material', 5,  25.000,  'kg', 7, 11.50, 11.50, '2026-10-22 09:00:00', NULL, 'active'),
  (11, 'KUZ-2026-002', 'material', 5,  22.000,  'kg', 7, 11.50, 11.50, '2026-10-23 09:00:00', NULL, 'active'),
  (12, 'DAN-2026-001', 'material', 10, 45.000,  'kg', 8, 13.20, 13.20, '2026-10-23 09:00:00', NULL, 'active'),
  (13, 'SUC-2026-046', 'product', 1,  90.000,  'paket', NULL, 0, 18.10, '2026-10-22 17:00:00', '2026-11-22', 'active'),
  (14, 'SUC-2026-047', 'product', 1,  80.000,  'paket', NULL, 0, 18.20, '2026-10-23 17:00:00', '2026-11-23', 'active'),
  (15, 'BUT-2026-001', 'product', 2,  4.500,   'kg', NULL, 0, 0, '2026-10-23 18:00:00', NULL, 'active'),
  (16, 'SIR-2026-001', 'product', 3,  2.500,   'kg', NULL, 0, 0, '2026-10-23 18:00:00', NULL, 'active'),
  (17, 'OMU-2026-001', 'product', 4,  3.700,   'kg', NULL, 0, 0, '2026-10-23 18:00:00', NULL, 'active'),
  (18, 'KAB-2026-001', 'product', 5,  2.000,   'kg', NULL, 0, 0, '2026-10-23 18:00:00', NULL, 'active'),
  (19, 'KIY-2026-001', 'product', 6,  3.700,   'kg', NULL, 0, 0, '2026-10-23 18:00:00', NULL, 'active'),
  (20, 'KUS-2026-001', 'product', 7,  1.500,   'kg', NULL, 0, 0, '2026-10-23 18:00:00', NULL, 'active');

-- Ek üretim emirleri (toplam 9: 1 mevcut + 8 yeni)
INSERT INTO production_orders (id, code, status, started_at, finished_at, notes, service_name, service_unit_cost, service_qty) VALUES
  (2, 'URT-2026-046', 'completed', '2026-10-21 09:00:00', '2026-10-21 17:00:00', 'Sözleşmeli üretim 2', 'Sucuk üretim hizmeti', 5.50, 90.000),
  (3, 'URT-2026-047', 'completed', '2026-10-22 09:00:00', '2026-10-22 17:00:00', 'Sözleşmeli üretim 3', 'Sucuk üretim hizmeti', 5.50, 80.000),
  (4, 'URT-2026-048', 'completed', '2026-10-23 09:00:00', '2026-10-23 17:00:00', 'Sözleşmeli üretim 4', 'Sucuk üretim hizmeti', 5.50, 70.000),
  (5, 'URT-2026-049', 'completed', '2026-10-24 09:00:00', '2026-10-24 17:00:00', 'Sözleşmeli üretim 5', 'Sucuk üretim hizmeti', 5.50, 95.000),
  (6, 'URT-2026-050', 'completed', '2026-10-24 09:00:00', '2026-10-24 17:00:00', 'Sözleşmeli üretim 6', 'Sucuk üretim hizmeti', 5.50, 60.000),
  (7, 'URT-2026-051', 'completed', '2026-10-25 09:00:00', '2026-10-25 17:00:00', 'Sözleşmeli üretim 7', 'Sucuk üretim hizmeti', 5.50, 100.000),
  (8, 'URT-2026-052', 'completed', '2026-10-25 09:00:00', '2026-10-25 17:00:00', 'Sözleşmeli üretim 8', 'Sucuk üretim hizmeti', 5.50, 55.000),
  (9, 'URT-2026-053', 'completed', '2026-10-26 09:00:00', '2026-10-26 17:00:00', 'Sözleşmeli üretim 9', 'Sucuk üretim hizmeti', 5.50, 88.000);

-- Ek üretim girdileri/çıktıları (her emir için)
INSERT INTO production_order_inputs (production_order_id, material_lot_id, qty) VALUES
  (2, 5, 100.000), (2, 7, 2.000), (2, 8, 5.000), (2, 9, 1.000),
  (3, 6, 90.000),  (3, 7, 2.000), (3, 8, 5.000),
  (4, 5, 80.000),  (4, 9, 1.500),
  (5, 6, 100.000), (5, 8, 6.000),
  (6, 5, 70.000),  (6, 7, 2.000),
  (7, 6, 110.000), (7, 9, 1.000),
  (8, 5, 60.000),
  (9, 6, 95.000),  (9, 8, 5.000);

INSERT INTO production_order_outputs (production_order_id, product_lot_id, qty) VALUES
  (2, 13, 90.000),
  (3, 14, 80.000),
  (4, 13, 70.000),
  (5, 13, 95.000),
  (6, 14, 60.000),
  (7, 13, 100.000),
  (8, 14, 55.000),
  (9, 13, 88.000);

-- Ek lot bağlantıları (hammadde -> mamul, parçalama)
INSERT INTO lot_links (parent_lot_id, child_lot_id, qty, link_type) VALUES
  (5, 13, 100.000, 'production'), (7, 13, 2.000, 'production'), (8, 13, 5.000, 'production'), (9, 13, 1.000, 'production'),
  (6, 14, 90.000, 'production'),  (7, 14, 2.000, 'production'), (8, 14, 5.000, 'production'),
  (10, 15, 4.500, 'butchery'), (10, 16, 2.500, 'butchery'), (10, 17, 3.700, 'butchery'),
  (10, 18, 2.000, 'butchery'), (10, 19, 3.700, 'butchery'), (10, 20, 1.500, 'butchery'),
  (11, 15, 4.000, 'butchery'), (11, 16, 2.200, 'butchery'), (11, 17, 3.300, 'butchery'),
  (12, 19, 5.000, 'butchery'), (12, 20, 2.000, 'butchery');

-- Ek parçalama partileri (toplam 9)
INSERT INTO butchery_batches (id, code, animal_type, input_weight, unit_cost, total_cost, batch_date, status) VALUES
  (1, 'PRC-2026-001', 'kuzu', 25.000, 11.50, 287.50, '2026-10-22', 'completed'),
  (2, 'PRC-2026-002', 'kuzu', 22.000, 11.50, 253.00, '2026-10-23', 'completed'),
  (3, 'PRC-2026-003', 'dana', 45.000, 13.20, 594.00, '2026-10-23', 'completed'),
  (4, 'PRC-2026-004', 'kuzu', 20.000, 11.50, 230.00, '2026-10-24', 'completed'),
  (5, 'PRC-2026-005', 'kuzu', 26.000, 11.50, 299.00, '2026-10-24', 'completed'),
  (6, 'PRC-2026-006', 'dana', 50.000, 13.20, 660.00, '2026-10-25', 'completed'),
  (7, 'PRC-2026-007', 'kuzu', 24.000, 11.50, 276.00, '2026-10-25', 'completed'),
  (8, 'PRC-2026-008', 'dana', 40.000, 13.20, 528.00, '2026-10-26', 'completed'),
  (9, 'PRC-2026-009', 'kuzu', 23.000, 11.50, 264.50, '2026-10-26', 'completed');

-- Parçalama ürünleri: her hayvan -> 2 but, 1 sırt, 2 omuz/kol, 1 kaburga, kıyma, kuşbaşı, kemik
-- (toplam ~%86; fire ~%14). product_id: 2=but 3=sırt 4=omuz 5=kaburga 6=kıyma 7=kuşbaşı 8=kemik
INSERT INTO butchery_items (batch_id, product_id, product_lot_id, weight, quantity, is_waste) VALUES
  -- PRC-2026-001 kuzu 25kg
  (1, 2, 15, 4.500, 2, 0), (1, 3, 16, 3.000, 1, 0), (1, 4, 17, 3.500, 2, 0),
  (1, 5, 18, 2.000, 1, 0), (1, 6, 19, 3.500, 1, 0), (1, 7, 20, 1.750, 1, 0), (1, 8, NULL, 3.250, 1, 0),
  -- PRC-2026-002 kuzu 22kg
  (2, 2, NULL, 3.960, 2, 0), (2, 3, NULL, 2.640, 1, 0), (2, 4, NULL, 3.080, 2, 0),
  (2, 5, NULL, 1.760, 1, 0), (2, 6, NULL, 3.080, 1, 0), (2, 7, NULL, 1.540, 1, 0), (2, 8, NULL, 2.860, 1, 0),
  -- PRC-2026-003 dana 45kg
  (3, 2, NULL, 7.200, 2, 0), (3, 3, NULL, 5.850, 1, 0), (3, 4, NULL, 5.850, 2, 0),
  (3, 5, NULL, 4.050, 1, 0), (3, 6, NULL, 6.750, 1, 0), (3, 7, NULL, 3.600, 1, 0), (3, 8, NULL, 5.400, 1, 0),
  -- PRC-2026-004 kuzu 20kg
  (4, 2, NULL, 3.600, 2, 0), (4, 3, NULL, 2.400, 1, 0), (4, 4, NULL, 2.800, 2, 0),
  (4, 5, NULL, 1.600, 1, 0), (4, 6, NULL, 2.800, 1, 0), (4, 7, NULL, 1.400, 1, 0), (4, 8, NULL, 2.600, 1, 0),
  -- PRC-2026-005 kuzu 26kg
  (5, 2, NULL, 4.680, 2, 0), (5, 3, NULL, 3.120, 1, 0), (5, 4, NULL, 3.640, 2, 0),
  (5, 5, NULL, 2.080, 1, 0), (5, 6, NULL, 3.640, 1, 0), (5, 7, NULL, 1.820, 1, 0), (5, 8, NULL, 3.380, 1, 0),
  -- PRC-2026-006 dana 50kg
  (6, 2, NULL, 8.000, 2, 0), (6, 3, NULL, 6.500, 1, 0), (6, 4, NULL, 6.500, 2, 0),
  (6, 5, NULL, 4.500, 1, 0), (6, 6, NULL, 7.500, 1, 0), (6, 7, NULL, 4.000, 1, 0), (6, 8, NULL, 6.000, 1, 0),
  -- PRC-2026-007 kuzu 24kg
  (7, 2, NULL, 4.320, 2, 0), (7, 3, NULL, 2.880, 1, 0), (7, 4, NULL, 3.360, 2, 0),
  (7, 5, NULL, 1.920, 1, 0), (7, 6, NULL, 3.360, 1, 0), (7, 7, NULL, 1.680, 1, 0), (7, 8, NULL, 3.120, 1, 0),
  -- PRC-2026-008 dana 40kg
  (8, 2, NULL, 6.400, 2, 0), (8, 3, NULL, 5.200, 1, 0), (8, 4, NULL, 5.200, 2, 0),
  (8, 5, NULL, 3.600, 1, 0), (8, 6, NULL, 6.000, 1, 0), (8, 7, NULL, 3.200, 1, 0), (8, 8, NULL, 4.800, 1, 0),
  -- PRC-2026-009 kuzu 23kg
  (9, 2, NULL, 4.140, 2, 0), (9, 3, NULL, 2.760, 1, 0), (9, 4, NULL, 3.220, 2, 0),
  (9, 5, NULL, 1.840, 1, 0), (9, 6, NULL, 3.220, 1, 0), (9, 7, NULL, 1.610, 1, 0), (9, 8, NULL, 2.990, 1, 0);

-- Ek stok hareketleri (girişler + sevkiyatlar)
INSERT INTO stock_movements (lot_id, movement_type, qty, unit_cost, total_cost, ref_type, ref_id, occurred_at) VALUES
  (5, 'inbound', 120.000, 9.70, 1164.00, 'supplier', 6, '2026-10-21 08:00:00'),
  (6, 'inbound', 90.000, 9.80, 882.00, 'supplier', 6, '2026-10-22 08:00:00'),
  (10, 'inbound', 25.000, 11.50, 287.50, 'supplier', 7, '2026-10-22 09:00:00'),
  (11, 'inbound', 22.000, 11.50, 253.00, 'supplier', 7, '2026-10-23 09:00:00'),
  (12, 'inbound', 45.000, 13.20, 594.00, 'supplier', 8, '2026-10-23 09:00:00'),
  (13, 'shipment', -30.000, 18.10, 543.00, 'market', 1, '2026-10-22 10:00:00'),
  (13, 'shipment', -25.000, 18.10, 452.50, 'market', 2, '2026-10-22 11:00:00'),
  (13, 'shipment', -20.000, 18.10, 362.00, 'market', 4, '2026-10-22 12:00:00'),
  (13, 'shipment', -15.000, 18.10, 271.50, 'market', 5, '2026-10-23 10:00:00'),
  (14, 'shipment', -28.000, 18.20, 509.60, 'market', 3, '2026-10-23 11:00:00'),
  (14, 'shipment', -22.000, 18.20, 400.40, 'market', 6, '2026-10-24 10:00:00'),
  (14, 'shipment', -30.000, 18.20, 546.00, 'market', 7, '2026-10-24 11:00:00');

-- Ek satışlar (toplam 9)
INSERT INTO sales (id, code, market_id, sold_at, status, client_id) VALUES
  (1, 'SAT-2026-001', 1, '2026-10-22 14:00:00', 'completed', 'android'),
  (2, 'SAT-2026-002', 2, '2026-10-22 15:00:00', 'completed', 'android'),
  (3, 'SAT-2026-003', 3, '2026-10-23 14:00:00', 'completed', 'android'),
  (4, 'SAT-2026-004', 4, '2026-10-23 15:00:00', 'completed', 'android'),
  (5, 'SAT-2026-005', 5, '2026-10-24 14:00:00', 'completed', 'android'),
  (6, 'SAT-2026-006', 6, '2026-10-24 15:00:00', 'completed', 'android'),
  (7, 'SAT-2026-007', 7, '2026-10-25 14:00:00', 'completed', 'android'),
  (8, 'SAT-2026-008', 8, '2026-10-25 15:00:00', 'completed', 'android'),
  (9, 'SAT-2026-009', 9, '2026-10-26 14:00:00', 'completed', 'android');

INSERT INTO sale_lines (sale_id, lot_id, qty, unit_price, total_price, stock_movement_id) VALUES
  (1, 4, 5, 24.90, 124.50, NULL),
  (2, 4, 3, 24.90, 74.70, NULL),
  (3, 13, 6, 24.90, 149.40, NULL),
  (4, 13, 4, 24.90, 99.60, NULL),
  (5, 13, 7, 24.90, 174.30, NULL),
  (6, 14, 5, 24.90, 124.50, NULL),
  (7, 14, 6, 24.90, 149.40, NULL),
  (8, 13, 8, 24.90, 199.20, NULL),
  (9, 14, 4, 24.90, 99.60, NULL);
