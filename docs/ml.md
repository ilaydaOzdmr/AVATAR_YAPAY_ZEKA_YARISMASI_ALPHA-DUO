# ML servisi — model kartı ve çalıştırma

## 1. Kısa özet

| Alan | Değer |
|---|---|
| Servis | Python 3.12 + FastAPI |
| Modeller | Üretim verimi/EKP; kuzu parçalama; sığır parçalama |
| Model sürümü | `1.1.0` |
| Model durumu | `synthetic_only` |
| Eğitim verisi | `ml/data/processed/synthetic_v2/*_synthetic.csv` |
| Gerçek kamu verisi | USDA verim özetleri, North Wyke kuzu, Hararghe sığır; ayrı referanslar |
| API erişimi | Go backend proxy: `http://localhost:8080/ml/*` |
| Doğrudan ML portu | `http://localhost:8000` (iç servis) |

> **Önemli:** Modeller gerçek işletme kayıtlarıyla eğitilmedi. Bu nedenle
> metrikler gerçek saha başarısı olarak sunulmaz. Gerçek veri sağlanana kadar
> durum `synthetic_only` kalır.

## 2. Veri kaynakları

Ayrıntılı güncel doğrulama: [DATA_SOURCES.md](DATA_SOURCES.md).

- **Gerçek ve indirilen:** USDA pişirme tablosu (173 özet), North Wyke
  (2.978 kuzu), Hararghe (61 sığır). Mamul/EKP/parça hedefleri eksik
  olduğundan eğitimde kullanılmazlar.
- **Sentetik ve kalibre:** üretim emirleri (480), kuzu (800), sığır (800).
  Her satırda `data_origin=synthetic_calibrated`, `generator_version=synthetic-v2`.

## 3. Modeller ve doğrulama

- **Görev A:** `finished_qty`, `yield_pct`. EKP model hedefi değildir.
  Fatura/maliyet alanları model girdisi değildir. API mamulü `et × tahmini verim`,
  EKP'yi ayrı maliyet hesabıyla üretir. Mamul regresyonu ayrıca tanısal amaçlı tutulur.
- **Görev B:** kuzu ve sığır için ayrı modeller; 10 hedef (parça ağırlıkları,
  toplam satılabilir, parçalama verimi).
- Adaylar: Dummy, Linear, Ridge, Lasso, RandomForest, ExtraTrees,
  HistGradientBoosting, XGBoost, CatBoost.
- Ön işleme: `ColumnTransformer` + `SimpleImputer` + `OneHotEncoder` +
  `StandardScaler`.
- Ayırma: `GroupShuffleSplit(test_size=0.20, seed=42)`; gruplar üretimde
  `production_batch`, parçalamada `farm_id`.
- CV: `GroupKFold(5)` yalnız eğitim verisi; hiperparametre yalnız CV’de.
- Test seti seçim/ayarlamada kullanılmaz.
- `--quick`: Dummy, Linear, Ridge, RandomForest; yine 5 katlı grup CV.
  Bu oturumda bu aday kümesi yeniden eğitildi. Ham regresyon ve API'nin
  fiziksel kurallarından sonraki metrikler ayrı raporlanır.

Seçilen modeller ve tüm metrikler: `ml/reports/model_comparison.md`,
`ml/reports/model_report.md`, `ml/reports/evaluation_report.json`.
Eski `model_performance_report.md` ve izlenen `plots/` tarihsel v1'dir.
Yeni grafikler ve kayıt bazlı hatalar `ml/reports/local/` (Git dışı) içindedir.

## 4. Artefaktlar

```
ml/artifacts/reliability_v2/
  yield_model/{model.joblib,metadata.json,metrics.json,split.json}
  butchery_lamb_model/{...}
  butchery_cattle_model/{...}
```

API açılışında **yalnız artefakt yüklenir**; yeniden eğitim yapılmaz. Artefakt
yoksa tahmin endpoint’leri **503** döner (`model_status=unavailable`).

## 5. API

Go proxy (önerilen):

| Metot | Yol | Açıklama |
|---|---|---|
| GET | `/ml/health` | ML sağlık + yüklü modeller |
| GET | `/ml/models` | Model metadata |
| GET | `/ml/report` | Model metrikleri |
| POST | `/ml/predict/v2/yield` | Üretim verimi/EKP |
| POST | `/ml/predict/v2/butchery` | Parçalama |
| POST | `/ml/predict/yield` | Eski sözleşme |
| POST | `/ml/predict/butchery` | Eski sözleşme |

ML servisi iç endpoint’leri: `/health`, `/models`, `/report`,
`/predict/v2/yield`, `/predict/v2/butchery`, `/predict/yield`, `/predict/butchery`.

### Üretim v2 örneği

```json
POST /ml/predict/v2/yield
{
  "raw_material_kg": 100,
  "spice_kg": 2,
  "casing_kg": 3,
  "service_cost": 400,
  "labor_cost": 80
}
```

Cevap:

```json
{
  "prediction": {"finished_qty": 84, "yield_pct": 0.84, "estimated_ekp": 18.2476},
  "model_version": "1.1.0",
  "model_status": "synthetic_only",
  "applicable_input_range": {"raw_material_kg": {"min": 60.1, "max": 159.9}},
  "warnings": ["Model is trained on synthetic_calibrated data..."],
  "uncertainty": {"finished_qty": {"coverage_validated": false}},
  "cost_breakdown": {"total_cost": 1532.8},
  "ekp_kind": "estimated_pre_production"
}
```

Yukarıdaki cevap yalnız şema örneğidir; çalıştırılmış tahmin/test metriği değildir.
`actual_finished_qty` verilirse `accounting_ekp` ve `ekp_kind=actual_accounting`
döner. Bu, tahmini EKP ile gerçek muhasebe EKP’sini ayırır.
`service_unit_cost` tarifesi tahminde tahmini, muhasebede fiili mamulle çarpılır;
sabit toplam `service_cost` ile birlikte verilmez. Eksik alanların demo
varsayımları uyarıda bildirilir.

### Parçalama v2 örneği

```json
POST /ml/predict/v2/butchery
{"animal_type":"lamb","live_weight_kg":45,"carcass_weight_kg":21}
```

Cevap `prediction` içinde parça ağırlıkları, `physical_check` içinde
`parts_sum_kg`/`parts_over_carcass` ve `warnings` döner. Negatif olmayan parça
oranları karkasa dengelenir; 0,001 kg yuvarlama kalanı dağıtılır. Satılabilir
toplam ve verim bu parçalardan türetilir; düzeltme uyarıda açıklanır.
Karkas kg zorunlu, canlı kg opsiyoneldir; sabit dressing oranı yoktur.
Legacy `input_weight` de karkas kg anlamındadır.

## 6. Çalıştırma

```cmd
python -m ml.scripts.import_real_data
python ml/scripts/build_datasets.py
python ml/train.py --quick
python ml/evaluate.py
python -m pytest ml/tests -q
docker compose -f infra/docker-compose.yml up -d --build
```

## 7. Sınırlılıklar

- Sentetik veri gerçek saha doğrulaması değildir.
- `yield_pct` test R² ≈ 0.72; yalnız sentetik veri üzerinde ölçülmüştür.
- Lineer modellerin kazanması üreticinin lineer olmasındandır; gerçek veride
  sıralama değişebilir.
- Bu eğitimde yerel USDA fiyat dosyası yoktu; fiyat endeksi 1,0 fallback'i kullanıldı.
- Gerçek verinin gelmesi tek başına `validated_real` statüsü vermez; lisans,
  kapsam, gözlem kimliği ve bağımsız grup/zaman testi ayrıca doğrulanmalıdır.
