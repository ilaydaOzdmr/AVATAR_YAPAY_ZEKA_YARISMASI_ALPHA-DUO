# API Referansı — AVATAR ALPHA-DUO backend

Go REST API. Base URL: `http://localhost:8080`. Tüm yanıtlar JSON.
Hata yanıtı: `{"error":"..."}` ve uygun HTTP kodu.

## Endpoint'ler

### Katalog ve lotlar
| Metot | Yol | Başarı | Açıklama |
|-------|-----|--------|----------|
| GET | `/health` | 200 | `{"status":"ok"}` |
| GET | `/materials` | 200 | Hammadde listesi |
| POST | `/materials` | 201 | Hammadde ekle |
| GET | `/products` | 200 | Ürün listesi |
| POST | `/products` | 201 | Ürün ekle |
| GET | `/lots` | 200 | Lot listesi (`?item_type=material\|product`) |
| POST | `/lots` | 201 | Lot ekle |
| GET | `/lots/{code}` | 200 | Lot detayı |
| GET | `/lots/{code}/movements` | 200 | Lot stok hareketleri |
| POST | `/stock-movements` | 201 | Stok hareketi ekle (`inbound/transfer/return/rework/shipment`) |
| GET | `/markets` | 200 | Market listesi |

### Üretim ve parçalama
| Metot | Yol | Başarı | Açıklama |
|-------|-----|--------|----------|
| GET | `/production-orders` | 200 | Üretim emri listesi |
| POST | `/production-orders` | 201 | Üretim emri ekle (inputs[]/outputs[]) |
| GET | `/production-orders/{id}` | 200 | Maliyet özeti (EKP/VKP/brüt kâr/marj) |
| GET | `/production-orders/{id}/inputs` | 200 | Emrin hammadde girdileri |
| GET | `/butchery-batches` | 200 | Parçalama kaydı listesi |
| POST | `/butchery-batches` | 201 | Parçalama kaydı ekle (batch + items[]) |
| GET | `/butchery-batches/{id}` | 200 | Parça + maliyet dağıtımı (ağırlık/değer) |

### İzlenebilirlik, geri çağırma, satış
| Metot | Yol | Başarı | Açıklama |
|-------|-----|--------|----------|
| GET | `/trace/forward?lot=` | 200 | İleriye izleme |
| GET | `/trace/backward?lot=` | 200 | Geriye izleme |
| POST | `/recalls` | 201 | Geri çağırma başlat (bloke) |
| GET | `/recalls/{id}` | 200 | Geri çağırma sonucu |
| GET | `/sales` | 200 | Satış listesi |
| POST | `/sales` | 201 | Satış ekle (bloklu lot → 409) |
| GET | `/redistributions?lot_id=` | 200 | İkinci kademe dağıtım |
| POST | `/redistributions` | 201 | Market → alt alıcı dağıtımı ekle |

### Senkronizasyon, anomali, OCR
| Metot | Yol | Başarı | Açıklama |
|-------|-----|--------|----------|
| GET | `/sync/pull` | 200 | Bloklu lot kodları + ürünler |
| POST | `/sync/push` | 200 | Mobil satışları toplu gönder (bloklu → 409) |
| GET | `/anomalies` | 200 | Kural tabanlı anomali uyarıları |
| POST | `/ocr/invoice` | 200 | Fatura/irsaliye metninden kalem çıkar |

### ML servisi (ayrı servis, `http://localhost:8000`)
| Metot | Yol | Açıklama |
|-------|-----|----------|
| POST | `/predict/yield` | Üretim verimi + EKP tahmini (`raw_material_kg`) |
| POST | `/predict/butchery` | Parçalama parça ağırlıkları (`animal_type`,`input_weight`) |
| GET | `/report` | Model vs baseline MAE/RMSE |

## Örnek istekler

### Üretim emri ekle
```json
POST /production-orders
{"code":"URT-2026-050","status":"draft","service_name":"Sucuk üretim",
 "service_unit_cost":5.50,"service_qty":85,
 "inputs":[{"material_lot_id":1,"qty":100}],
 "outputs":[{"product_lot_id":4,"qty":85}]}
```

### Parçalama kaydı ekle
```json
POST /butchery-batches
{"batch":{"code":"KUZ-2026-032","animal_type":"kuzu","input_weight":25,
  "unit_cost":11.50,"total_cost":287.50,"batch_date":"2026-10-27","status":"draft"},
 "items":[{"product_id":2,"weight":10,"quantity":1,"is_waste":false}]}
```

### Parçalama fire ve maliyet kontrolü
`GET /butchery-batches/{id}` (ve `POST` yanıtı) kayıttan otomatik hesaplanan
kontrol alanlarını döner. Bu alanlar istemciden alınmaz; backend hesaplar ve
`butchery_batches` tablosunda saklar.

| Alan | Anlam |
|---|---|
| `part_weight` | Fire olmayan parçaların toplam ağırlığı |
| `fire_weight` | `input_weight - part_weight` |
| `weight_diff` | `input_weight - (parça + is_waste satırları)` |
| `weight_check_status` | `ok` / `mismatch` — 3 ondalık |
| `expected_cost` | `input_weight × unit_cost` — 2 ondalık |
| `recorded_cost` | `total_cost` |
| `cost_difference` | `recorded_cost - expected_cost` |
| `cost_check_status` | `ok` / `mismatch` — 2 ondalık |

- Fire tek başına veri hatası değildir. `weight_check_status` yalnızca satılabilir
  parçalar giriş ağırlığını aştığında (negatif fire) `mismatch` olur.
- Maliyet ve ağırlık kontrolleri birbirinden bağımsızdır.
- `KUZ-2026-DEMO`: 25 kg, 11,50 €/kg, 351,00 € → fire 3 kg, beklenen 287,50 €,
  kayıtlı 351,00 €, fark 63,50 €, `cost_check_status = mismatch`.

### Geri çağırma
```json
POST /recalls
{"source_lot_code":"ET-2026-001","reason":"Kalite sorunu"}
```

### Geri çağırma alıcı teyidi
```json
POST /recalls/{id}/acknowledge
{"market_code":"MKT-A"}
```

- Teyit, kaynak lot bazında kalıcı saklanır; aynı lotun tüm geri çağırma
  kayıtlarında ortak görünür.
- Aynı lot için tekrar geri çağırma başlatılırsa yeni kayıt açılmaz; mevcut
  kayıt ve teyitleri döner.
- Teyit satış blokajını kaldırmaz; lot güvenlik kapsamında `blocked` kalır.
- `GET /recalls/{id}` ve `POST /recalls` yanıtlarında `confirmed_markets`
  alanı döner.

### Uyarılar
```json
GET /anomalies
POST /anomalies/read
{"keys":["stock:SUC-2026-045","yield:URT-2026-045"]}
```

- Her uyarı `key`, `type`, `severity`, `message`, `ref` ve `read` alanlarını
  döner. `key` = `<type>:<ref>` biçimindedir.
- Okundu bilgisi `anomaly_reads` tablosunda kalıcıdır; okunan uyarı listede
  soluk görünür ve bildirim rozetinde sayılmaz.
- Stok kuralı artık negatif bakiye yerine **"Stok az olduğu için satış
  engellendi"** mesajını üretir.
- Stoktan fazla satış `409` ve `insufficient stock` hatasıyla reddedilir;
  lot miktarı değişmez.

## Notlar
- `stock_movements.qty` işaretli: giriş `+`, çıkış `-`.
- Hareket tipleri: `inbound`, `shipment`, `transfer`, `return`, `rework`.
- Para `DECIMAL(12,2)`, JSON'da sayı olarak döner.
- Tüm SQL parametreli.

---

## ML proxy (Go backend)

### Web model bilgileri

`GET /ml/health` ve `/ml/models`, yüklü modellerin sürümü, veri sürümü,
`data_origin`, eğitim tarihi, özellikler, eğitim/test sayıları, doğrulama yöntemi
ve `input_ranges` alanlarını döner. Üç modelden biri eksikse sağlık gövdesi
`status=degraded` olur; HTTP 200 gövdesi eksik model bilgisini taşır.

`GET /ml/report` eski alanları korur; ek `model_reports` nesnesi `yield`,
`butchery_lamb`, `butchery_cattle` anahtarlarıyla **fiziksel kısıt sonrası**
eğitim/test/oransal baseline metriklerini ve model/veri sürümlerini sunar.
`production_quantity` kg cinsinden mamul özetidir; eski `yield` özeti ise
0–1 verim oranının hatasını taşır. Web bunları aynı birimmiş gibi göstermez.

V2 tahmin yanıtları `model_version` yanında `dataset_version` da taşır. Web,
metadata ve tahminin model/veri sürümlerinin eşleşmesini kontrol eder.

`GET /production-orders/{id}/outputs`, kaydedilmiş mamul kabul satırlarını
(`product_lot_id`, `qty` dahil) döner. Bu salt okunur uç nokta, webin ürün
tipini ve paket→kg dönüşümünü ürün kartıyla doğrulamasını sağlar. Geçersiz
kimlik 400, bulunmayan emir 404, boş liste `[]` döner. Paket ağırlığı ürün
adında açıkça `(1 kg)` veya `(500 g)` gibi belirtilmiyorsa web karşılaştırmayı
reddeder; farklı paket ağırlıklarını tek plan biriminde birleştirmez.

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

---

## Kök, senkronizasyon ve raporlar

### `GET /`
Tarayıcıda 404 yerine servis bilgisi döner:
```json
{"service":"avatar-backend","status":"ok","version":"1.0.0","endpoints":["/health","/ml/health", "..."]}
```

### `GET /sync/pull`
Mobil için tek snapshot:
```json
{
  "server_time": "2026-10-31T...",
  "blocked_lot_codes": ["SUC-2026-045"],
  "products": [...],
  "lots": [...],
  "markets": [...]
}
```
Mobil bu yanıtı tek Room transaction’ında uygular; yarım pull SQLite’ı boşaltmaz.
Ayrıntı: `docs/mobile-sync.md`.

### `GET /reports/summary`
MySQL’den altı rapor bölümü:
```json
{
  "generated_at": "2026-10-31T...",
  "purchases": [...],
  "production": [...],
  "butchery": [...],
  "lot_movements": [...],
  "sales": [...],
  "recalls": [...]
}
```
Alanlar: alış lotu/miktar/maliyet, üretim emri/EKP, parçalama fire/parça,
lot hareketi, satış/market/tutar, recall/bloklu lotlar.

### Satış idempotency
`POST /sales` aynı `code` ile tekrar çağrılırsa mevcut satışı döndürür; stok
ikinci kez düşmez. Mobil outbox retry’leri bu sayede güvenlidir.

## Ürün ve lot oluşturma (raf ömrü / SKT)

`POST /products`:
```json
{"code":"PRD-200","name":"Test Sucuk","unit":"kg","list_price":25,
 "category":"finished","is_finished":true,"shelf_life_days":15}
```
Cevapta `created_at` ve `shelf_life_days` döner.

`POST /materials` aynı şekilde `shelf_life_days` kabul eder.

`POST /lots`:
```json
{"code":"LOT-2026-100","item_type":"product","item_id":11,"qty":3,
 "unit":"kg","unit_cost":12,"received_at":"2026-10-01 00:00:00","status":"active"}
```
- `received_at` ürün lotunda **üretim tarihi**dir.
- `expiry_date` boşsa backend `received_at + shelf_life_days` ile hesaplar.
- `qty > 0` ise aynı transaction’da `inbound` stok hareketi yazılır.
- Ürün/hammadde bulunamazsa `400` (`ErrInvalidLot`).

### Üretim emri tamamlama
`POST /production-orders/{id}/complete`

```json
{}
```
veya bitiş zamanı ile:
```json
{"finished_at":"2026-10-31 17:00:00"}
```

- `draft` emir yalnızca plandır; stok hareketi/lot bağı yazılmaz.
- Tamamlamada girdi lotları tüketilir, çıktı lotları stoğa girer, lot bağları
  ve SKT işlenir; emir `completed` olur.
- Aynı emir ikinci kez tamamlanırsa stok tekrar düşmez (idempotent).
- Girdi/çıktı satırı yoksa `400`.

### Üretim emri reddetme
`POST /production-orders/{id}/reject`

```json
{"reason":"kalite reddi"}
```

- Yalnızca `completed` olmayan ve stok hareketi uygulanmamış emirler reddedilebilir.
- Durum `rejected` olur; ret gerekçesi `notes` alanına yazılır.
- Stok, lot bağı ve SKT değişmez.
- Tamamlanmış emir reddedilemez (`400`).

### Beklenen mamul miktarı
`POST /production-orders` gövdesinde `expected_qty` gönderilmelidir. Web formu
artık bu alanı gönderiyor; liste ve maliyet özetinde görünür.
