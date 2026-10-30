# Model performans raporu — Aşama 3

Model sürümü: `1.0.0` · Durum: `synthetic_only`
Eğitim: `python ml/train.py` · Değerlendirme: `python ml/evaluate.py`

> **Kritik:** Eğitim verisi `synthetic_calibrated`. Bu metrikler gerçek saha
> başarısı değildir; gerçek emir/karkas verisiyle dış doğrulama yapılmamıştır.

## 1. Deney tasarımı

- **Gruplu ayırma:** Üretimde `production_batch`, parçalamada `farm_id`.
  `GroupShuffleSplit(test_size=0.20, seed=42)`.
- **Çapraz doğrulama:** `GroupKFold(5)` yalnız eğitim verisi üzerinde.
- **Hiperparametre:** `RandomizedSearchCV`, 4 örnek, yalnız eğitim/CV.
- **Test:** Seçim ve ayarlamada kullanılmadı; yalnız raporlama.
- **Adaylar:** DummyRegressor, Linear, Ridge, Lasso, RandomForest, ExtraTrees,
  HistGradientBoosting, XGBoost, CatBoost.
- **Ön işleme:** ColumnTransformer + SimpleImputer + OneHotEncoder + StandardScaler.
- **Seçim ölçütü:** CV MAE (düşük daha iyi).

## 2. Sonuç özeti

22 hedefin **tamamı** ortalama baseline’ı geçti. Öne çıkanlar:

| Grup | Hedef | Seçilen | CV MAE | Test MAE | Test R² |
|---|---|---|---:|---:|---:|
| yield | finished_qty | lasso | 1.4014 | 1.4535 | 0.9945 |
| yield | yield_pct | linear | 0.0118 | 0.0131 | 0.6882 |
| butchery_lamb | total_saleable_weight | lasso | 0.1378 | 0.1433 | 0.9936 |
| butchery_lamb | cutting_yield_pct | lasso | 0.5905 | 0.5985 | 0.9088 |
| butchery_cattle | total_saleable_weight | linear | 2.0682 | 2.1227 | 0.9891 |
| butchery_cattle | cutting_yield_pct | ridge | 0.6034 | 0.6244 | 0.9129 |

Tam tablo: `ml/reports/model_comparison.md` (9 aday × 22 hedef).

## 3. Bilimsel yorum

- **Lineer/ridge/lasso modelleri kazandı.** Bunun nedeni sentetik üreticinin
  ağırlıklı olarak lineer olmasıdır; ağaç tabanlı modeller (RF, ET, HGB, XGBoost,
  CatBoost) bu veride ek kazanç sağlamadı. Bu, gerçek veride de aynı sıralama
  olacağı anlamına gelmez.
- **`finished_qty` R² ≈ 0.99**, çünkü hedef hammadde miktarıyla güçlü ilişkili.
- **`yield_pct` R² ≈ 0.69**, çünkü verimde parti etkisi ve gürültü var; bu hedef
  için belirsizlik daha yüksek. API bu nedenle `yield_pct` için daha geniş
  belirsizlik/warning üretmeli.
- **EKP bir model hedefi değildir.** Gerçek EKP = toplam gerçek maliyet / kabul
  edilen fiili mamul. Tahmini EKP, tahmini mamul ve bilinen maliyetlerden
  türetilecek; muhasebe EKP’si ile karıştırılmayacak.
- **Parça toplamı fiziksel olarak tutarlı:** parça toplamı = karkas; satılabilir
  toplam karkası aşmıyor (Aşama 2 kalite raporu).

## 4. Tanılama çıktıları

`ml/reports/plots/` altında:

- `*_actual_vs_predicted.png` — gerçek vs tahmin
- `*_residuals.png` — residual
- `*_model_comparison.png` — aday karşılaştırma
- `*_feature_importance.png` — permütasyon önemi
- `*_shap.png` — SHAP özeti (desteklenen modellerde)

İnteraktif inceleme: `ml/notebooks/model_analysis.ipynb`.

## 5. Artefaktlar

| Klasör | İçerik |
|---|---|
| `ml/artifacts/yield_model/` | `model.joblib`, `metadata.json`, `metrics.json`, `split.json` |
| `ml/artifacts/butchery_lamb_model/` | aynı yapı |
| `ml/artifacts/butchery_cattle_model/` | aynı yapı |

Metadata alanları: model sürümü, eğitim tarihi, veri sürümü/kaynağı, özellik ve
hedef adları, örnek sayıları, doğrulama yöntemi, aday metrikleri, uyarılar.

## 6. Sınırlılıklar ve sonraki adım

- Gerçek veri yok; model durumu `synthetic_only`.
- Gerçek veri sağlanırsa aynı hat yeniden çalıştırılmalı; sıralama değişebilir.
- Hanwoo OpenAPI anahtarı sağlanırsa gerçek primal cut verisiyle içe aktarma
  yapılabilir; bu depoda anahtar yok.
- EKP tahmini Aşama 4’te ayrı muhasebe kimliği olarak uygulanacak.
