# Veritabanını İnceleme

Bağlantı: `localhost:3306` · DB `avatar` · kullanıcı `avatar` / `avatar` (root: `root`/`root`).

## 1) Aç (Docker üzerinden MySQL CLI)
```bash
docker exec -it avatar-mysql mysql -uavatar -pavatar avatar
```
> PowerShell'de `docker exec` yerine `docker compose -f infra/docker-compose.yml exec mysql mysql -uavatar -pavatar avatar`.

## 2) Şema ve satır sayıları
```sql
-- tüm tablolar
SHOW TABLES;

-- satır sayıları
SELECT 'materials' t, COUNT(*) n FROM materials
UNION ALL SELECT 'products', COUNT(*) FROM products
UNION ALL SELECT 'suppliers', COUNT(*) FROM suppliers
UNION ALL SELECT 'markets', COUNT(*) FROM markets
UNION ALL SELECT 'lots', COUNT(*) FROM lots
UNION ALL SELECT 'production_orders', COUNT(*) FROM production_orders
UNION ALL SELECT 'production_order_inputs', COUNT(*) FROM production_order_inputs
UNION ALL SELECT 'production_order_outputs', COUNT(*) FROM production_order_outputs
UNION ALL SELECT 'butchery_batches', COUNT(*) FROM butchery_batches
UNION ALL SELECT 'butchery_items', COUNT(*) FROM butchery_items
UNION ALL SELECT 'lot_links', COUNT(*) FROM lot_links
UNION ALL SELECT 'stock_movements', COUNT(*) FROM stock_movements
UNION ALL SELECT 'sales', COUNT(*) FROM sales
UNION ALL SELECT 'sale_lines', COUNT(*) FROM sale_lines
UNION ALL SELECT 'recall_events', COUNT(*) FROM recall_events
UNION ALL SELECT 'recall_blocks', COUNT(*) FROM recall_blocks
UNION ALL SELECT 'sync_outbox', COUNT(*) FROM sync_outbox;

-- bir tablonun sütunları
DESCRIBE lots;
```

## 3) İzlenebilirlik zincirini gör (ET-2026-001 senaryosu)
```sql
-- hammadde -> mamul bağlantıları
SELECT p.code parent, c.code child, l.qty
FROM lot_links l
JOIN lots p ON p.id = l.parent_lot_id
JOIN lots c ON c.id = l.child_lot_id;

-- sevkiyatlar (hangi markete ne gitti)
SELECT l.code lot, m.name market, ABS(sm.qty) qty
FROM stock_movements sm
JOIN lots l ON l.id = sm.lot_id
JOIN markets m ON m.id = sm.ref_id
WHERE sm.movement_type = 'shipment'
ORDER BY l.code, m.code;

-- üretim emri maliyeti (EKP)
SELECT po.code, SUM(i.qty * l.unit_cost) input_cost, po.service_unit_cost * po.service_qty service_cost,
       ROUND((SUM(i.qty * l.unit_cost) + po.service_unit_cost * po.service_qty) / MAX(o.qty), 2) ekp
FROM production_orders po
JOIN production_order_inputs i ON i.production_order_id = po.id
JOIN lots l ON l.id = i.material_lot_id
JOIN production_order_outputs o ON o.production_order_id = po.id
GROUP BY po.id;
```

## 4) GUI (isteğe bağlı)
[DBeaver](https://dbeaver.io) veya [MySQL Workbench](https://dev.mysql.com/downloads/workbench/) kur,
`localhost:3306 / avatar / avatar / avatar` ile bağlan — tabloları ve veriyi görsel incele.
