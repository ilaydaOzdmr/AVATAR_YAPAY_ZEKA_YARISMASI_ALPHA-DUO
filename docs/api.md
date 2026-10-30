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

---

## ML proxy (Go backend)

Web ve Android, Python ML servisine doğrudan gitmez. Go backend sunucu tarafında
proxy yapar. Tarayıcı/mobil yalnız `http://localhost:8080` bilir.

| Metot | Yol | Başarı | Açıklama |
|---|---|---|---|
| GET | `/ml/health` | 200 | ML sağlık + yüklü model sayısı |
| GET | `/ml/models` | 200 | Model sürümü/durumu/hedefleri |
| GET | `/ml/report` | 200 | Model vs baseline metrikleri |
| POST | `/ml/predict/v2/yield` | 200 | Üretim verimi, mamul, tahmini EKP |
| POST | `/ml/predict/v2/butchery` | 200 | Parçalama parça ağırlıkları |
| POST | `/ml/predict/yield` | 200 | Eski sözleşme (uyumluluk) |
| POST | `/ml/predict/butchery` | 200 | Eski sözleşme (uyumluluk) |

Hata kodları:
- `502` — ML servisine bağlanılamadı.
- `503` — model artefaktı yüklenmemiş (`model_status=unavailable`).
- `504` — ML servisi zaman aşımı (10 sn).
- `422` — Pydantic doğrulama hatası (v2 gövdesi).

### Üretim v2

```json
POST /ml/predict/v2/yield
{
  "raw_material_kg": 100,
  "spice_kg": 2,
  "casing_kg": 3,
  "service_cost": 400,
  "labor_cost": 80,
  "actual_finished_qty": 80
}
```

Cevap `prediction` (`finished_qty`, `yield_pct`, `estimated_ekp`),
`model_version`, `model_status`, `applicable_input_range`, `warnings`,
`uncertainty`, `cost_breakdown` ve `ekp_kind` içerir. `actual_finished_qty`
verilirse `accounting_ekp` ayrıca döner.

### Parçalama v2

```json
POST /ml/predict/v2/butchery
{"animal_type":"lamb","live_weight_kg":45,"carcass_weight_kg":21}
```

`animal_type`: `lamb` veya `cattle`. `carcass_weight_kg` verilmezse canlı
ağırlıktan tahmin edilir ve uyarı üretilir. Cevapta `physical_check` ve
`warnings` bulunur.

Model kartı ve veri sınırlılıkları: `docs/ml.md`.
