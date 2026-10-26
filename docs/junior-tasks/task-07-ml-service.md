# JUNIOR (ilayda) GELİŞTİRİCİ TALİMATI + GÖREV 7

## ROL (önce bunu oku)
Sen **ilayda takımının JUNIOR** yazılım geliştiricisisin. ilayda takımının senior'ı
sana görev verdi. Sen kodu YAZARSIN; senior inceler, düzeltir, Docker'da test eder ve
**commit'i senior atar** (sen commit/push/PR YAPMAZ; sadece lokalde dosyaları değiştirirsin).

Kuralların (kesin uy):
1. **SADECE görevde isteneni yap.** Fazladan özellik, ek endpoint, ek model EKLEME.
2. Verilen isim, imza ve çıktı alanlarını **BİREBİR** kullan.
3. İsimlendirme İngilizce, açıklayıcı. Clean code, LLM-okunur.
4. Verilen testleri + iddiaları tam yaz.
5. Günlük notu: yaptıklarını **kendi klasörün olan** `ALPHA-DUO-Vault/ilayda/YYYY-MM-DD.md`
   dosyasına yaz (`Meryem/` klasörüne dokunma). Gerçeği yaz; çalıştırmadığın testi "geçti" diye yazma.
6. Teslimat: eklediğin/değiştirdiğin dosyaların TAM içeriği, dosya yollarıyla.

## BAĞLAM
Monorepo. Yeni `ml/` dizini açılacak. Python + FastAPI + scikit-learn. İki zorunlu KI
tahmini (PDF): (A) üretim verimi + EKP, (B) parçalama verimi. Her ikisi **basit referans
yöntemle (ortalama) karşılaştırılacak** ve **MAE/RMSE** raporlanacak. Veri az olduğu için
**sınırlılık açıkça belirtilecek**. Lot izlenebilirliği/geri çağırma KI tahminine dayanmaz
(bu görevde yok).

## GÖREV: ML servisi (iki model + metrik raporu)

### 1) `ml/requirements.txt`
```
fastapi
uvicorn
scikit-learn
numpy
pandas
pytest
```

### 2) `ml/model.py` — saf fonksiyonlar (test edilebilir)
```python
import numpy as np
from sklearn.linear_model import LinearRegression

def train_yield(X, y):
    """Doğrusal regresyon eğitir. X: (n, k) ndarray, y: (n,) ndarray."""
    model = LinearRegression()
    model.fit(X, y)
    return model

def predict(model, X):
    return model.predict(X)

def baseline_mean(y):
    """Referans: tüm örnekler için ortalama tahmin."""
    return float(np.mean(y))

def evaluate(y_true, y_pred):
    """MAE ve RMSE döndürür: {"mae": float, "rmse": float}."""
    ...
```

### 3) `ml/train.py` — örnek/anonymize veri + eğitim + `report.json`
- Küçük, **anonymize/örnek** bir veri seti oluştur (sabit liste/DataFrame yeter):
  - üretim emirleri: `raw_material_kg`, `service_cost` → hedef `finished_qty` ve `unit_cost`
    (örn. 5-8 kayıt).
  - parçalama: `input_weight` (ve hayvan türü kodlanmış) → hedef parça ağırlıkları (örn. 4-6 kayıt).
- Her model için: eğit, **baseline (ortalama)** ile karşılaştır, `evaluate` ile MAE/RMSE hesapla.
- `ml/report.json` yaz:
  ```json
  {
    "yield": {"model_mae": ..., "model_rmse": ..., "baseline_mae": ..., "baseline_rmse": ...},
    "butchery": {"model_mae": ..., "model_rmse": ..., "baseline_mae": ..., "baseline_rmse": ...},
    "note": "Veri seti örnek/anonymize ve küçüktür; sonuçlar gösterim amaçlıdır."
  }
  ```

### 4) `ml/app.py` — FastAPI
- `GET /health` → `{"status":"ok"}`
- `POST /predict/yield` — gövde `{"raw_material_kg": 100, "service_cost": 467.5}` →
  `{"predicted_qty": ..., "predicted_unit_cost": ...}` (model + basit EKP = toplam maliyet / qty).
- `POST /predict/butchery` — gövde `{"animal_type": "kuzu", "input_weight": 25}` →
  `{"parts": [{"product_code": "PRD-101", "weight": ...}, ...]}`.
- `GET /report` → `report.json` içeriğini döndür.
- Modelleri başlangıçta `train.py`'deki aynı örnek veriyle eğit (veya `train.py` çağır).

### 5) `ml/test_model.py` — test
```python
def test_baseline_mean():
    assert baseline_mean([10, 20, 30]) == 20.0

def test_evaluate():
    m = evaluate([1, 2, 3], [1, 2, 3])
    assert m["mae"] == 0.0 and m["rmse"] == 0.0

def test_train_predict_shape():
    X = [[1], [2], [3], [4]]; y = [2, 4, 6, 8]
    model = train_yield(X, y)
    assert predict(model, [[5]])[0] == pytest.approx(10.0, abs=1e-6)
```

## KABUL (senior kontrol edecek)
- `pip install -r requirements.txt` + `pytest ml/test_model.py` geçer.
- `python ml/train.py` → `report.json` üretir (model + baseline MAE/RMSE).
- `uvicorn ml.app:app` → `/health`, `/predict/yield`, `/predict/butchery`, `/report` çalışır.
- İngilizce isimlendirme, clean code.
