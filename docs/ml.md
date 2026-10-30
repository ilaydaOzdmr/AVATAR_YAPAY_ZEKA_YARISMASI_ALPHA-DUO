# ML servisi — model kartı ve çalıştırma

## 1. Kısa özet

| Alan | Değer |
|---|---|
| Servis | Python 3.12 + FastAPI |
| Modeller | Üretim verimi/EKP; kuzu parçalama; sığır parçalama |
| Model sürümü | `1.0.0` |
| Model durumu | `synthetic_only` |
| Eğitim verisi | `ml/data/processed/*_synthetic.csv` |
| Gerçek kamu verisi | USDA ERS fiyat serileri (yalnız fiyat girdisi) |
| API erişimi | Go backend proxy: `http://localhost:8080/ml/*` |
| Doğrudan ML portu | `http://localhost:8000` (iç servis) |

> **Önemli:** Modeller gerçek işletme kayıtlarıyla eğitilmedi. Bu nedenle
> metrikler gerçek saha başarısı olarak sunulmaz. Gerçek veri sağlanana kadar
> durum `synthetic_only` kalır.

## 2. Veri kaynakları

Ayrıntılı kaynak doğrulaması: `ml/data/README.md`, `ml/reports/data_quality_report.md`.

- **Gerçek ve indirilen:** USDA ERS Meat Price Spreads ve Livestock & Meat
  Domestic Data (aylık fiyat/istatistik serileri).
- **İndirilemeyen/kısıtlı:** North Wyke (kayıt gerekli), NZ Lamb 792 (ham veri
  açık değil), Irish Cattle ~200k (ham veri açık değil), Hanwoo (OpenAPI anahtarı
  gerekli).
- **Sentetik ve kalibre:** üretim emirleri (480), kuzu (800), sığır (800).
  Her satırda `data_origin=synthetic_calibrated`, `generator_version=synthetic-v1`.

## 3. Modeller ve doğrulama

- **Görev A:** `finished_qty`, `yield_pct`. EKP model hedefi değildir.
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

Seçilen modeller ve tüm metrikler: `ml/reports/model_comparison.md`,
`ml/reports/model_report.md`, `ml/reports/model_performance_report.md`.
Grafikler: `ml/reports/plots/`. Notebook: `ml/notebooks/model_analysis.ipynb`.

## 4. Artefaktlar

```
ml/artifacts/
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
  "prediction": {"finished_qty": 80.3, "yield_pct": 0.796, "estimated_ekp": 13.11},
  "model_version": "1.0.0",
  "model_status": "synthetic_only",
  "applicable_input_range": {"raw_material_kg": {"min": 60.1, "max": 159.9}},
  "warnings": ["Model is trained on synthetic_calibrated data..."],
  "uncertainty": {"finished_qty": {"test_rmse": 1.99}},
  "cost_breakdown": {"total_cost": 1052.6},
  "ekp_kind": "estimated_pre_production"
}
```

`actual_finished_qty` verilirse `accounting_ekp` ve `ekp_kind=actual_accounting`
döner. Bu, tahmini EKP ile gerçek muhasebe EKP’sini ayırır.

### Parçalama v2 örneği

```json
POST /ml/predict/v2/butchery
{"animal_type":"lamb","live_weight_kg":45,"carcass_weight_kg":21}
```

Cevap `prediction` içinde parça ağırlıkları, `physical_check` içinde
`parts_sum_kg`/`parts_over_carcass` ve `warnings` döner. Fiziksel olarak tutarsız
değerler sessizce düzeltilmez; uyarı üretilir.

## 6. Çalıştırma

```cmd
python ml/scripts/download_usda.py
python ml/scripts/build_datasets.py
python ml/train.py
python ml/evaluate.py
python -m pytest ml/tests -q
docker compose -f infra/docker-compose.yml up -d --build
```

## 7. Sınırlılıklar

- Sentetik veri gerçek saha doğrulaması değildir.
- `yield_pct` R² ≈ 0.69; belirsizlik diğer hedeflerden yüksektir.
- Lineer modellerin kazanması üreticinin lineer olmasındandır; gerçek veride
  sıralama değişebilir.
- USDA fiyatları ABD piyasasıdır; işletme fiyat kataloğuyla ölçeklenmiştir.
- Gerçek veri geldiğinde aynı hat yeniden çalıştırılmalı ve model durumu
  `validated_real` olarak güncellenmelidir.
