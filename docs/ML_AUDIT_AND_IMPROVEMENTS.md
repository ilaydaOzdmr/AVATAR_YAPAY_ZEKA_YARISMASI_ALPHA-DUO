# ML denetimi ve iyileştirmeler

Tarih: **31 Ekim 2026**. Dal: `fix/ml-reliability`. Kapsam: ekli görev metni ve
gerçek veri araştırması. Başlangıçta izlenen dosyalarda kullanıcı değişikliği
yoktu; iki bağımsız `Başlıksız*.base` dosyasına dokunulmadı ve commit'e alınmadı.
GitHub'a push yapılmadı.

## Sonuç ve PDF uyumu

`ALPHA-DUO-Vault/AVATAR_Assessment.pdf` s. 1–2 incelendi. A: geçmiş emirlerden
mamul/EKP; B: hayvan/parti ve giriş ağırlığından parçalar. Lot izlenebilirliği
ve geri çağırma kesin kayıt ilişkileriyle çalışmaya devam eder. Muhasebe EKP'si
tahmine değil fiili mamule bölünür. 100 kg et + 2 kg baharat + 5 kg bağırsak,
85 × 1 kg paket ve 5,50 €/paket hizmet örneği otomatik testte **1.543,50 €** ve
**18,16 €/paket** üretir. 25 kg × 11,50 € = 287,50 € ve parça toplamı 22 kg
tutarsızlıkları mevcut backend testleriyle korunur; model bu değerleri “düzeltmek”
için gerçek kayıtlara müdahale etmez.

Gerçek veriler bulundu ve indirildi: **173 USDA pişirme özeti**, **2.978 North
Wyke kuzusu**, **61 Hararghe sığırı**. Kaynak/lisans, indirme, hedef karşılaştırması,
temizleme ve eksikler [DATA_SOURCES.md](DATA_SOURCES.md) içindedir. Bunlar gerçek
referans tablolarıdır; eksik mamul/EKP/parça hedefleri uydurulmadı. Yarışmanın
modelleri **synthetic_only** durumunda kalır. Gerçek saha testi veya gerçek
veriyle model başarısı iddia edilmez.

## Bulunan hatalar ve uygulanan düzeltmeler

| Hata / neden | Düzeltme |
|---|---|
| Sentetik casing, hizmet ve işçilik miktarları `finished_qty` sonucundan türetiliyordu; fatura tutarları tahmin öncesi bilinmeyebilir | Bağırsak miktarı et girdisinden; hizmet bütçesi planlanan mamulden; işçilik et girdisinden üretilir. Fiyat, hizmet, işçilik ve sonuç/plan alanları verim modeline girmez |
| Plan `expected_qty` hedefe eklenen küçük gürültüydü | Ürün tipi ve et girdisine dayalı üretim öncesi plan kullanılır. Sonuç gürültüsünü değiştirmenin plan/girdileri değiştirmediği test edilir |
| Bağımsız mamul ve verim regresyonları aynı fiziksel kimliği garanti etmiyordu | API mamulü `et kg × tahmin edilen verim` olarak üretir. Ham mamul regresyonu tanısal raporda ayrıca tutulur |
| M3 emir karşılaştırması et+baharat+bağırsak miktarlarını tek et ağırlığı yapıyordu | Lot→malzeme kartı eşlemesiyle girdi miktarları/fiyatları ayrı V2 alanlarına gider. Bilinmeyen kart veya kg dışı birim reddedilir |
| Web boş/geçersiz girdileri `parseFloat(...) || 0` yapıyordu, API hatası sıfır tahmin gibi görünüyordu | Sayılar tam ve sonlu olarak doğrulanır. Eksik tahmin `null` ve hata mesajı olur; grafiğe sıfır gözlem eklenmez |
| Legacy parçalama ağırlığı karkas olmasına rağmen canlı gibi 0,47/0,58 ile küçültülüyordu | Legacy `input_weight` = ölçülmüş karkas kg. V2 karkas zorunlu, canlı opsiyonel. Kaynaksız dressing dönüşümü kaldırıldı |
| Bağımsız parça/toplam/verim modelleri negatif miktar veya karkası aşan toplam verebiliyordu | Sekiz parça negatif olmayan oranlarla ölçülmüş karkasa dengelenir; binde kg yuvarlama kalanı dağıtılır. Satılabilir toplam ve verim aynı parçalardan türetilir; işlem yanıtta açıklanır |
| Parçalama ekranı farklı kapsamlı toplamları karşılaştırıyordu | Yalnız fiilen kaydı bulunan aynı ürün kodları karşılaştırılır. Kemik/fire satılabilir toplamdan çıkarılır; parça bazlı mutlak ve yüzde hata gösterilir |
| Yüzde sapmalar ve eksik tahminler belirsizdi | Mutlak hata = `abs(tahmin−fiili)`, yüzde hata = mutlak hata/fiili×100. Fiili sıfırsa yüzde tanımsız gösterilir |
| İllüstratif RMSE bandı %95 güven izlenimi veriyordu | Fiziksel kısıt sonrası test RMSE kullanılır; kapsamanın kalibre edilmediği açıkça belirtilir, `coverage_validated=false` |
| Gerçek hayvan verisinin bulunmadığına dair eski genel iddia | Açık Mendeley veri yayınları ve USDA dosyaları indirildi. Eski kaynak raporları tarihsel olarak işaretlendi |
| Veride tek bir negatif mamul satırı denetimden geçebiliyordu (`all`) | `any` ile reddedilir; sonlu değer, karkas/canlı, parça/toplam/verim kimlikleri denetlenir |
| Artefakt/özellik sözleşmesi uyumsuz yüklenebilirdi | v1.1 özelliği ve sürümü metadata ile doğrulanır; yeni artefakt yolu v1'den ayrıdır. Ortam scikit-learn 1.7.2 / numpy 2.1.3 / pandas 2.3.3 ile sabitlendi |

Üretim model girdileri: `raw_material_kg`, `spice_kg`, `casing_kg`,
`other_material_kg`, `product_type`. Et bazlı verim oranı 0–1 ile sınırlanır;
bu demo model su eklenen ürünler için genel bir kütle modeli değildir.
Maliyetler ayrı hesaplanır: sabit hizmet bütçesi veya `service_unit_cost` tarifesi
seçilir. Tarife tahminde beklenen mamulle, muhasebede fiili mamulle çarpılır.
Eksik opsiyonel API alanlarının demo varsayımları uyarıda listelenir.

Kuzu/sığır ayrı modellerinin girdileri: `breed`, `sex`, `age_months`,
`carcass_weight_kg`, `conformation_score`, `fat_score`. Canlı ağırlık, sunulursa
karkasın daha büyük olmaması için denetlenir; parça tahminine girdi yapılmaz.
Çıktılar altı satılabilir parça, kemik, fire, toplam satılabilir ve verimdir.
Son iki çıktı bağımsız regresyon sonucundan değil dengelenmiş parçalardan alınır.

## Veri sızıntısı ve değerlendirme koşulları

Ön işleme her CV turunda yalnız eğitim bölümüne fit edilen pipeline içindedir.
Üretim `production_batch`, parçalama `farm_id` ile ayrılır. Aynı kimlik/grup
train/test'e düşmez; tekrarlı/eksik gözlem kimlikleri ve eksik grup reddedilir.
Test hedeflerini değiştirmek model seçimini veya scaler ortalamasını değiştirmez;
otomatik test bunun kontrolünü yapar. Gerçek referans satırları mevcut sentetik
eğitici tarafından reddedilir. North Wyke'da satış sonrası tartım seçilmediği
ayrıca test edilir.

`python ml/train.py --quick` ile **22 hedef** için Dummy, Linear, Ridge,
RandomForest karşılaştırıldı; modeller CV MAE ile seçildi. Adına rağmen quick
mod yine **5 katlı** GroupKFold kullanır; daha geniş aday kümesi bu oturumda
çalıştırılmadı. Ayrı doğrulama holdout'u yoktur; aşağıdaki sayılar eğitim içindeki
CV doğrulama katlarının büyüklükleridir. Test bölümü seçimden bağımsızdır.

| Veri grubu | Eğitim | CV doğrulama katları | Test |
|---|---:|---|---:|
| Üretim | 376 | 75 / 74 / 77 / 75 / 75 | 104 |
| Kuzu | 592 | 78 / 122 / 131 / 131 / 130 | 208 |
| Sığır | 600 | 76 / 134 / 126 / 132 / 132 | 200 |

### Eski ve yeni sonuçlar

Eski sayılar izlenen v1 metadata'sından; yeni sayılar API ile aynı fiziksel
kimliklerden sonra hesaplanan v1.1 test sonuçlarıdır. **Hepsi sentetiktir.**
Üretici ve aday kümesi değişti; üretimde test satırları da 86'dan 104'e değişti.
Bu tablo kontrollü aynı-test karşılaştırması veya gerçek dünya iyileşme kanıtı
değildir. Kuzu satılabilir MAE'sinin az da olsa yükselmesi gizlenmemiştir.

| Hedef | Eski MAE | Yeni MAE | Eski RMSE | Yeni RMSE | Eski R² | Yeni R² | Yeni MAPE % |
|---|---:|---:|---:|---:|---:|---:|---:|
| Mamul kg | 1.4535 | 1.4015 | 1.9975 | 1.7698 | 0.9945 | 0.9945 | 1.5069 |
| Üretim verim oranı | 0.0131 | 0.0129 | 0.0166 | 0.0156 | 0.6882 | 0.7227 | 1.5072 |
| Kuzu satılabilir kg | 0.1433 | 0.1435 | 0.1790 | 0.1798 | 0.9936 | 0.9936 | 0.9437 |
| Sığır satılabilir kg | 2.1227 | 2.0970 | 2.6721 | 2.6468 | 0.9891 | 0.9893 | 0.9031 |

Tüm hedefler, ham CV MAE/std, aday metrikleri, train/test metrikleri, fiziksel
kısıt sonrası ölçümler, hata quantile'ları ve kullanılan ortam:
`ml/reports/model_comparison.json`, `model_comparison.md`, `model_report.md`.
`ml/evaluate.py` kaydedilmiş test indeksleriyle, verinin hash'ini doğrulayarak,
yeniden eğitim yapmadan aynı fiziksel kurallarla `evaluation_report.json` üretir.

| Hedef | Eğitim MAE | Test−eğitim MAE | Oransal baseline test MAE | Yeni model test MAE |
|---|---:|---:|---:|---:|
| Mamul kg | 1.2654 | 0.1361 | 1.4765 | 1.4015 |
| Üretim verim oranı | 0.0116 | 0.0013 | 0.0135 | 0.0129 |
| Kuzu satılabilir kg | 0.1353 | 0.0082 | 0.4276 | 0.1435 |
| Sığır satılabilir kg | 2.0153 | 0.0817 | 6.8790 | 2.0970 |

Oransal baseline üretimde ürün başına eğitim verimi ortalaması × et kg;
parçalamada eğitimdeki ortalama parça/karkas oranları × karkastır. Baseline
aynı fiziksel kimliklerle, aynı test satırlarında değerlendirilir; testten oran
öğrenmez. Dummy mean karşılaştırması ayrıca korunur. Seçili modeller mevcut
testte bu oransal baseline'ı geçer; buna bakarak model/hiperparametre değişikliği
yapılmadı. Gerçek saha üstünlüğü ve fallback kararı için yeterli kanıt yoktur.

Kayıt bazlı fiili/tahmin/mutlak hata/yüzde hata dosyaları
`ml/reports/local/*_test_errors.csv` içindedir. Test hata p50/p90/p95/maksimum
özetleri metadata'da yer alır. Yeni grafikler yerel `local/plots/` altındadır;
SHAP opsiyoneldir, bu ortamda yoktu ve üretilmiş gibi raporlanmaz.

## Testler ve doğrulama sınırları

- Başlangıç: mevcut Python testlerinin 22'si geçti; eksik fiziksel/sızıntı
  kontrollerini yakalamıyorlardı. Eski artefaktlar yüklü sklearn sürümüyle
  uyumsuzluk uyarısı verdi; v1.1 artefaktları mevcut sürümle yeniden eğitildi.
- Son Python turu: **44 başarılı, 0 başarısız, 0 atlanan**. Kapsam: gerçek
  dosyalarla çevrimdışı import, checksum reddi, zaman seçimi, girdi/kimlik/grup
  denetimi, testten bağımsız seçim, deterministik tahmin, üretim maliyet
  bağımsızlığı, PDF muhasebesi, kuzu/sığır fiziksel kısıtları, legacy/V2 uyumu.
- Web: **4 başarılı, 0 başarısız** Node testi. PDF reçetesi, boş/geçersiz
  girdiler, hata paydası, aynı parçaların kapsam eşlemesi. TypeScript + Vite
  üretim derlemesi başarılı.
- Backend: `go test ./...` başarılı; servis/maliyet/parçalama ve ML proxy
  sözleşmesi testleri dahil. Backend kodunda düzeltme gerekmiyordu; proxy
  V2 alanlarını aynen iletiyor.
- İlk web denemesi PowerShell `npm.ps1` yürütme ilkesi/dizin kısıtına takıldı;
  `npm.cmd` ile ve Vite için gerekli sandbox izniyle başarıyla tamamlandı.
  Go önbellek erişimi de gerekli izinle yeniden çalıştırıldı.
- MySQL gerektiren `-tags integration` testleri, canlı tarayıcı/MySQL/Python/Go
  zinciri, Android, Docker image build ve gerçek işletme saha testi çalıştırılmadı.
  Tam sistem uçtan uca doğrulaması yapıldı iddiası yoktur. FastAPI TestClient
  kullanımında httpx deprecation uyarısı kalır; test hatası değildir.

## Çalıştırma ve artefaktlar

Depo kökünden PowerShell:

```powershell
python -m venv ml/.venv
.\ml\.venv\Scripts\python.exe -m pip install -r ml/requirements.txt
.\ml\.venv\Scripts\python.exe -m ml.scripts.import_real_data
.\ml\.venv\Scripts\python.exe ml/scripts/build_datasets.py
.\ml\.venv\Scripts\python.exe ml/train.py --quick
.\ml\.venv\Scripts\python.exe ml/evaluate.py
.\ml\.venv\Scripts\python.exe -m pytest ml/tests ml/test_model.py -q
.\ml\.venv\Scripts\python.exe -m uvicorn ml.app:app --reload
```

Bu oturumda ortam `--system-site-packages` ile kuruldu; yalnız eksik FastAPI ve
httpx bağımlılıkları proje venv'ine kuruldu, global ortam değiştirilmedi.
Backend için `cd backend; go run ./cmd/server`, web için
`cd web; npm.cmd run dev` kullanılır; MySQL/.env kurulumu README'dedir.
Web testleri web dizininde `node --test tests/predictionInputs.test.mjs`
(bu test komutu Node 24 ile doğrulandı); derleme `npm.cmd run build`.

Yeni modeller `ml/artifacts/reliability_v2/` içindedir, sürüm `1.1.0` ve durum
`synthetic_only`. JSON metadata tarih, sklearn sürümü, veri hash'i, özellik ve
hedef sırası, grup ayrımı, metrikler ve eğitim aralıklarını içerir. API başlangıcı
yeniden eğitmez. Model yoksa tahmin 503 döner; yerel ilk çalıştırmada önce
eğitim gerekir. Dockerfile aynı sentetik üretim/eğitim adımlarını image build
sırasında çalıştıracak şekilde güncellendi; Docker build bu oturumda denenmedi.

İndirilen veriler, yeni sentetik CSV'ler, model binary'leri ve kayıt bazlı hata
dosyaları Git dışıdır. Tarihsel v1 izlenen veriler/modeller silinmedi veya
yeniden commit edilmedi. Yerel v1.1 modelleri verilen komutlarla yeniden üretilebilir.

## Kalan sınırlamalar

Gerçek mamul/EKP/parça hedefleri ve işletme bazında bağımsız test hâlâ eksik.
Hararghe'de 62/61 sayım farkı ve HDP tutarsızlıkları; North Wyke'da tartımın
kesimden önceki gecikmesi ve eksik ölçümler korunur. Bu dosyalar doğrudan
mevcut eğitim hedeflerine zorla eşlenmedi. Kamu fiyat dosyası bu oturumda
yerelde olmadığı için sentetik fiyat endeksi 1,0'dır; gerçek fiyat kullanılmış
gibi gösterilmez.

Ekrandaki otomatik üretim eşlemesi mevcut MAT-001…004 kataloğu ve sucuk ürün
varsayımına dayanır. Yeni ürün/kart/birimler için açık eşleme gerekir;
bilinmeyen kartlar sessizce ete eklenmez. Manuel ekran ürün tipi sucuktur.
kg/paket dönüşümü yalnız PDF'deki 1 kg paket senaryosu için geçerlidir.
Tahmin aralıkları dışında uyarı üretilir; güvenilir genelleme iddia edilmez.
Eksik hayvan özelliklerinde demo varsayımları uyarılır. Fiziksel dengeleme,
yanlış modelin doğruluğunu garanti etmez. Belirsizlik bandı kalibre edilmiş
bir güven aralığı değildir.

## Değişen dosyalar

- Çalıştırma/bağımlılık: `.gitignore`, `README.md`, `ml/.dockerignore`,
  `ml/Dockerfile`, `ml/requirements.txt`, `ml/app.py`.
- Veri: `ml/src/config.py`, `ml/src/data.py`, `ml/src/synthetic.py`,
  `ml/scripts/build_datasets.py`, `ml/scripts/data_quality.py`,
  `ml/scripts/import_real_data.py`, `ml/data/README.md`.
- Eğitim/tahmin: `ml/src/models.py`, `ml/src/constraints.py`,
  `ml/src/registry.py`, `ml/src/schemas.py`, `ml/src/service.py`,
  `ml/train.py`, `ml/evaluate.py`.
- Testler: `ml/tests/test_api.py`, `ml/tests/test_reliability.py`,
  `ml/tests/test_real_data.py`, `web/tests/predictionInputs.test.mjs`.
- Web: `web/src/predictionInputs.ts`, `web/src/m3/pages/Predictions.tsx`,
  `web/src/pages/Predictions.tsx`.
- Belgeler: `docs/DATA_SOURCES.md`, `docs/ML_AUDIT_AND_IMPROVEMENTS.md`, `docs/ml.md`.
- Sayısal raporlar: `ml/reports/data_quality.json`, `dataset_inventory.json`,
  `real_data_inventory.json`, `evaluation_report.json`, `model_comparison.json`,
  `model_comparison.md`, `model_report.md`. Eski `data_quality_report.md` ve
  `model_performance_report.md` başına tarihsel rapor notları eklendi.

Commit mesajı:
`fix(ml): prevent leakage, correct prediction inputs and validate model outputs`.
Commit hash'i son kullanıcı yanıtında bildirilir; rapor kendi commit hash'ini içermez.
