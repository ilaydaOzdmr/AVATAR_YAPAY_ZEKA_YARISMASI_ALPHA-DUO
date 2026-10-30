# Veri kalite raporu — Aşama 2

Tarih: 2026-10-30
Üretici: `ml/src/synthetic.py` (`synthetic-v1`)
Komutlar: `python ml/scripts/download_usda.py`, `python ml/scripts/build_datasets.py`

## 1. Özet

- **Gerçek kamu verisi:** USDA ERS fiyat/istatistik serileri indirildi ve doğrulandı.
- **Hayvan bazlı gerçek karkas/parça verisi:** Açık ve doğrudan indirilebilir kaynak
  bulunamadı. North Wyke kayıt istiyor; Ngo 792, Irish ~200k ve Hanwoo 3.467/50.227
  ham verileri açık değil.
- **Gerçek üretim emri verisi:** Kamuya açık kaynak yok. Mevcut DB’deki 9 emir
  demo/seed verisidir; gerçek işletme kaydı değildir.
- **Karar:** Eksik hedefler için `synthetic_calibrated` veri üretildi. Her satır
  `data_origin` ve `generator_version` taşır. Bu veriyle elde edilecek metrikler
  gerçek saha başarısı olarak sunulmayacaktır.

## 2. Gerçek veri kaynakları (doğrulandı)

| Kaynak | Erişim | Gerçek indirilebilir N | Kullanılabilir alanlar |
|---|---|---|---|
| USDA Meat Price Spreads | Doğrudan CSV/XLSX | 13.080 + 800 + 547 + 516 satır | Aylık perakende/toptan fiyat, marj |
| USDA Livestock & Meat Domestic | Doğrudan CSV/XLSX/ZIP | 5.156 + 7.093 + 73.846 satır | Canlı/toptan fiyat, üretim, kesim |
| North Wyke Farm Platform | Kayıt/giriş gerekli | Anonim indirme yok | Canlı ağırlık, otlatma, çevre; parça ağırlığı yok |
| NZ Lamb (Ngo 2016) | Makale | 0 | 792 karkas çalışması; ham veri açık değil |
| Irish Cattle (ICBF/Teagasc) | PDF/sunum | 0 | ~200.000 hayvan; 191.847 doğrulanamadı |
| Hanwoo (Naserkheil 2021) | Makale | 0 | 3.467 hayvan, 10 primal cut |
| Hanwoo OpenAPI (2026) | data.go.kr | 0 (anahtar gerekli) | 50.227 karkas kaydı |

> Makale N’leri indirilebilir dosya satır sayısı olarak kullanılmadı.

## 3. Üretilen veri setleri

| Dosya | Satır | Sütun | Kaynak |
|---|---:|---:|---|
| `production_orders_synthetic.csv` | 480 | 21 | Sentetik + USDA fiyat endeksi |
| `butchery_lamb_synthetic.csv` | 800 | 22 | Sentetik |
| `butchery_cattle_synthetic.csv` | 800 | 22 | Sentetik |

Eksik hücre: **0**.

## 4. Fiziksel ve muhasebe kontrolleri

### Üretim

- 60 üretim partisi; grup bölmesi için `production_batch`.
- `yield_pct`: 0.7936 – 0.9342 (hedef aralık 0.72–0.94).
- `finished_qty`: 49.935 – 145.555.
- **EKP kimliği:** `actual_total_cost / finished_qty` ile `actual_ekp` arasındaki
  maksimum mutlak fark **0.005** (2 ondalık yuvarlama).
- USDA fiyat endeksi: 0.5309 – 1.9824 (ort. ~1.0).

### Parçalama — kuzu

- 12 çiftlik; `farm_id` ile GroupKFold.
- Parça toplamı − karkas: maksimum **≈0** (float toleransı; 1e-13).
- Satılabilir toplam kimliği: maksimum **≈0** (1e-13).
- Karkas üstü satılabilir ağırlık: **0**.
- `cutting_yield_pct`: 64.872 – 77.130.
- Ortalama karkas: 21.31 kg.

### Parçalama — sığır

- 12 çiftlik.
- Parça toplamı − karkas: maksimum **≈0** (1e-13).
- Satılabilir toplam kimliği: maksimum **≈0** (1e-13).
- Karkas üstü satılabilir ağırlık: **0**.
- `cutting_yield_pct`: 63.972 – 77.364.
- Ortalama karkas: 326.86 kg.

## 5. Sınırlılıklar

- Bu veri **gerçek işletme kaydı değildir**; sentetik üretim mekanizması
  `ml/data/README.md` içinde açıklanmıştır.
- USDA fiyatları ABD piyasasıdır; işletme fiyat kataloğu ile ölçeklenmiştir.
- Gerçek veriyle dış doğrulama yapılmamıştır; model raporlarında
  `model_status = synthetic_only` ve açık uyarı yer alacaktır.
- Hanwoo OpenAPI erişimi sağlanırsa gerçek primal cut verisiyle yeniden eğitim
  için `ml/scripts/` altına içe aktarma betiği eklenebilir.

## 6. Tekrarlanabilirlik

- Tohumlar: üretim 42, kuzu 43, sığır 44.
- `ml/data/usda_manifest.json` her dosyanın SHA-256’sını içerir.
- Kalite JSON’u: `ml/reports/data_quality.json`.
