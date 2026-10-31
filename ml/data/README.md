# ML veri kaynakları ve veri sözlüğü

Bu klasör iki tür veriyi **ayırır**:

1. `raw/usda/` — gerçek, kamuya açık USDA ERS fiyat serileri.
2. `processed/synthetic_v2/*_synthetic.csv` — gerçek işletme kaydı olmayan, belgelenmiş
   sentetik veri. Her satırda `data_origin = synthetic_calibrated` vardır.

> Sentetik satırlar gerçek üretim emri veya gerçek karkas ölçümü değildir.
> Bu veriyle elde edilen metrikler gerçek saha başarısı olarak sunulmaz.

3. `raw/verified_real/` ve `real/` — USDA pişirme verimi, North Wyke kuzu ve
   Hararghe sığır ölçümleri. Güncel doğrulama:
   [DATA_SOURCES.md](../../docs/DATA_SOURCES.md). Eksik hedefler nedeniyle
   eğitimde kullanılmazlar. Önceki `processed/` dosyaları tarihsel v1'dir;
   v1.1 eğitimi onları okumaz.

## 1. Kaynak doğrulama tablosu (Aşama 2)

| Kaynak | İndirme | Format | Lisans | Makale N | İndirilebilir N | Hedefler | Durum |
|---|---|---|---|---|---|---|---|
| USDA ERS — Meat Price Spreads | [sayfa](https://www.ers.usda.gov/data-products/meat-price-spreads) | CSV/XLSX | ABD kamu malı | — | 13.080 + 800 + 547 + 516 satır | Aylık perakende/toptan fiyat ve marj | **Gerçek, indirildi** |
| USDA ERS — Livestock & Meat Domestic Data | [sayfa](https://www.ers.usda.gov/data-products/livestock-and-meat-domestic-data) | CSV/XLSX/ZIP | ABD kamu malı | — | 5.156 + 7.093 + 73.846 satır | Canlı/toptan fiyat, üretim, kesim istatistikleri | **Gerçek, indirildi** |
| North Wyke sheep V1 | [veri yayını](https://data.mendeley.com/datasets/xy3ndcy8jd/1) | ZIP/CSV | CC BY 4.0 | 2.978 | 2.978 (2.607 geçerli çift) | Canlı/soğuk karkas, parça yok | **Gerçek, indirildi** |
| NZ Lamb Primal Cuts (Ngo et al. 2016) | [DOI](https://doi.org/10.1016/j.meatsci.2015.10.012) | Makale | Elsevier telif | 792 karkas | 0 | Primal/retail cut ağırlıkları | **İndirilemez** |
| Irish Cattle Primal Cuts (ICBF/Teagasc) | [ICBF](https://www.icbf.com/) | PDF/sunum | ICBF/Teagasc | ~200.000 (191.847 doğrulanamadı) | 0 | Primal cut ağırlıkları, genetik değerlendirme | **İndirilemez** |
| Hanwoo carcass (Naserkheil et al. 2021) | [PMC](https://pmc.ncbi.nlm.nih.gov/articles/PMC8701981) | Makale | CC BY (makale) | 3.467 | 0 | 10 primal cut ağırlığı | **İndirilemez** |
| Hanwoo OpenAPI (2026 WssGWAS) | [data.go.kr](https://www.data.go.kr/) | OpenAPI | data.go.kr kullanım koşulları | 50.227 | 0 (servis anahtarı gerekli) | Karkas ve primal cut kayıtları | **Anahtar gerekli** |

### Doğrulama notları

- **USDA** dosyaları gerçekten indirildi; satır sayıları `scripts/download_usda.py`
  ve `reports/dataset_inventory.json` ile doğrulanır. Bunlar hayvan bazlı kesim
  verisi değil, aylık fiyat/istatistik serileridir. Bu yüzden yalnızca fiyat
  girdisi olarak kullanılır; parça ağırlığı hedefi olarak kullanılmaz.
- **North Wyke** için önceki portal incelemesi eksikti: Mendeley V1 anonim
  indirilebilir. Gerçek canlı/soğuk karkas eşleşmeleri vardır; parça hedefleri yoktur.
- **Ngo 2016** 792 karkaslık çalışmadır; ham veri makale ekinde açık değildir.
  Makalenin 792 sayısı indirilebilir dosya satır sayısı değildir.
- **Irish** kaynaklarında ~200.000 hayvandan söz edilir; 191.847 sayısı
  bağımsız olarak indirilebilir bir dosyada doğrulanamadı. Ham veri ICBF
  veritabanındadır; erişim talebi gerekir.
- **Hanwoo** için 3.467 hayvanlık makale verisi açık değildir. 50.227 kayıtlı
  güncel çalışma data.go.kr OpenAPI’sine dayanır ve servis anahtarı ister.

## 2. İşlenmiş veri dosyaları

### `processed/synthetic_v2/production_orders_synthetic.csv`

| Sütun | Tip | Açıklama |
|---|---|---|
| `order_id` | string | Sentetik emir kimliği |
| `production_batch` | string | Grup bölmesi için parti |
| `product_type` | kategori | sucuk / salam / sausage |
| `raw_material_kg` | float | Et girdisi |
| `spice_kg` | float | Baharat girdisi |
| `casing_kg` | float | Bağırsak girdisi |
| `other_material_kg` | float | Diğer girdi |
| `material_unit_price_*` | float | Girdi birim fiyatları |
| `service_cost` | float | Planlanan mamule bağlı demo bütçe; model girdisi değil |
| `labor_cost` | float | Et girdisine bağlı demo bütçe; model girdisi değil |
| `usda_price_index` | float | USDA boxed beef fiyat endeksi (ort. 1.0) |
| `expected_qty` | float | Planlanan mamul |
| `finished_qty` | float | **Hedef** gerçekleşen mamul |
| `yield_pct` | float | **Hedef** verim oranı |
| `actual_total_cost` | float | Gerçek maliyet (muhasebe) |
| `actual_ekp` | float | **Muhasebe** EKP = toplam maliyet / finished_qty |
| `data_origin` | string | `synthetic_calibrated` |
| `generator_version` | string | `synthetic-v2` |

`actual_ekp` bir model çıktısı değildir; gerçek maliyet kimliğidir. Model
yalnızca üretim öncesi `finished_qty` / `yield_pct` tahmini yapar.

### `processed/synthetic_v2/butchery_lamb_synthetic.csv` ve `butchery_cattle_synthetic.csv`

| Sütun | Tip | Açıklama |
|---|---|---|
| `animal_id` | string | Sentetik hayvan kimliği |
| `farm_id` | string | Grup bölmesi (GroupKFold) |
| `animal_type` | kategori | lamb / cattle |
| `breed`, `sex` | kategori | Irk, cinsiyet |
| `age_months` | float | Yaş |
| `live_weight_kg` | float | Sentetik canlı ağırlık; parçalama model girdisi değil |
| `carcass_weight_kg` | float | Karkas ağırlığı (girdi) |
| `conformation_score` | 1–5 | Konformasyon |
| `fat_score` | 1–5 | Yağ skoru |
| `leg_weight`, `loin_weight`, `shoulder_weight`, `rib_weight` | float | **Hedef** parça ağırlıkları |
| `mince_weight`, `diced_meat_weight` | float | **Hedef** parça ağırlıkları |
| `bone_weight`, `waste_weight` | float | **Hedef** kemik/fire |
| `total_saleable_weight` | float | **Hedef** satılabilir toplam |
| `cutting_yield_pct` | float | **Hedef** verim yüzdesi |
| `data_origin` | string | `synthetic_calibrated` |
| `generator_version` | string | `synthetic-v2` |

Fiziksel kısıt: `leg+loin+shoulder+rib+mince+diced+bone+waste = carcass_weight`
(1e-6 tolerans) ve `total_saleable ≤ carcass_weight`.

## 3. Sentetik üretim mekanizması

- Üretim: ürün tipine göre temel verim (sucuk 0.84, salam 0.87, sosis 0.90),
  parti rastgele etkisi, baharat/diğer girdi oranı etkisi ve ölçüm gürültüsü.
  Verim 0.72–0.94 aralığına kırpılır. Maliyetler gerçek fiyat kataloğu +
  USDA fiyat endeksiyle hesaplanır.
- Parçalama: canlı ağırlık, dressing %, konformasyon ve yağ skoru ile karkas
  ağırlığı; yayınlanmış parça oranı aralıklarına yakın temel oranlar + küçük
  rastgele sapma; toplam karkasa normalize edilir.
- Tohumlar sabittir; aynı komut aynı dosyaları üretir.

## 4. Komutlar

```cmd
python ml/scripts/download_usda.py
python ml/scripts/build_datasets.py
python -m ml.scripts.import_real_data
python ml/train.py --quick
python -m pytest ml/tests -q
```

## 5. Sınırlılıklar

- Gerçek üretim emri, gerçek karkas parça ölçümü veya gerçek EKP yok.
- Sentetik veriyle eğitilen modeller **gerçek saha doğrulaması** sayılmaz.
- North Wyke açık yayını indirildi; parça hedefleri yok. Diğer kaynakların
  güncel erişim/uygunluk durumu `docs/DATA_SOURCES.md` içindedir.
- USDA fiyatları ABD piyasasıdır; işletme fiyat kataloğu ile ölçeklenmiştir.
- `ml/data/usda_manifest.json` indirilen dosyaların URL, boyut ve SHA-256 kaydıdır.
