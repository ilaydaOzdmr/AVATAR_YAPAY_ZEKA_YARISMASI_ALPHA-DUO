@echo off
chcp 65001 >nul
echo ================================================
echo   AVATAR VERITABANI  (localhost:3306 / avatar)
echo ================================================
echo.

echo [1] TABLOLAR:
docker exec avatar-mysql mysql -uavatar -pavatar avatar -e "SHOW TABLES;"
echo.

echo [2] SATIR SAYILARI:
docker exec avatar-mysql mysql -uavatar -pavatar avatar -e "SELECT 'materials' t, COUNT(*) n FROM materials UNION ALL SELECT 'products', COUNT(*) FROM products UNION ALL SELECT 'suppliers', COUNT(*) FROM suppliers UNION ALL SELECT 'markets', COUNT(*) FROM markets UNION ALL SELECT 'lots', COUNT(*) FROM lots UNION ALL SELECT 'production_orders', COUNT(*) FROM production_orders UNION ALL SELECT 'butchery_batches', COUNT(*) FROM butchery_batches UNION ALL SELECT 'butchery_items', COUNT(*) FROM butchery_items UNION ALL SELECT 'lot_links', COUNT(*) FROM lot_links UNION ALL SELECT 'stock_movements', COUNT(*) FROM stock_movements UNION ALL SELECT 'sales', COUNT(*) FROM sales UNION ALL SELECT 'sale_lines', COUNT(*) FROM sale_lines UNION ALL SELECT 'recall_events', COUNT(*) FROM recall_events UNION ALL SELECT 'recall_blocks', COUNT(*) FROM recall_blocks;"
echo.

echo [3] ORNEK VERI (lots):
docker exec avatar-mysql mysql -uavatar -pavatar avatar -e "SELECT id, code, item_type, qty, unit FROM lots LIMIT 10;"
echo.

echo [4] Interaktif kabuk aciliyor (cikmak icin: exit)
docker exec -it avatar-mysql mysql -uavatar -pavatar avatar

pause
