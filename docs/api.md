# API Referansı — AVATAR ALPHA-DUO backend

Go REST API. Base URL: `http://localhost:8080`. Tüm yanıtlar JSON.
Hata yanıtı: `{"error":"..."}` ve uygun HTTP kodu.

## Endpoint'ler

| Metot | Yol | Başarı | Açıklama |
|-------|-----|--------|----------|
| GET | `/health` | 200 | `{"status":"ok"}` |
| GET | `/materials` | 200 | Hammadde listesi |
| POST | `/materials` | 201 | Hammadde ekle |
| GET | `/products` | 200 | Ürün listesi |
| POST | `/products` | 201 | Ürün ekle |
| GET | `/lots` | 200 | Lot listesi (`?item_type=material\|product` opsiyonel) |
| POST | `/lots` | 201 | Lot ekle |
| GET | `/lots/{code}` | 200 | Lot detayı |
| GET | `/lots/{code}/movements` | 200 | Lot stok hareketleri |
| POST | `/stock-movements` | 201 | Stok hareketi ekle |

## Örnek istekler

### Hammadde ekle
```json
POST /materials
{"code":"MAT-009","name":"Tuz","unit":"kg","is_raw":true,"default_purchase_price":3.50}
```

### Lot ekle
```json
POST /lots
{"code":"ET-2026-002","item_type":"material","item_id":1,"qty":100,"unit":"kg",
 "supplier_id":1,"purchase_price":9.70,"unit_cost":9.70,
 "received_at":"2026-10-20 08:00:00","status":"active"}
```

### Stok hareketi ekle
```json
POST /stock-movements
{"lot_id":4,"movement_type":"shipment","qty":-30,"unit_cost":18.16,"total_cost":544.80,
 "ref_type":"market","ref_id":1,"occurred_at":"2026-10-20 10:00:00"}
```

## Notlar
- `stock_movements.qty` işaretli: giriş `+`, çıkış `-`.
- Para `DECIMAL(12,2)`, JSON'da sayı olarak döner.
- Tüm SQL parametreli.
