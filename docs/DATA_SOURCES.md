# Gerçek veri kaynakları ve uygunluk kararı

Araştırma ve dosya doğrulama tarihi: **31 Ekim 2026**. Bu belge önceki
`ml/data/README.md` araştırmasının kapsamını genişletir. “Kamuya açık hayvan
verisi yok” şeklinde genel bir sonuç doğru değildir; indirilebilir gerçek
ölçümler bulundu. Ancak aşağıdaki dosyalar yarışmanın bütün eğitim hedeflerini
karşılamaz. **Gerçek veriyle yarışma modelleri yeniden eğitilmedi.**

## Yarışma PDF’sine göre veri gereksinimleri

Kaynak: `ALPHA-DUO-Vault/AVATAR_Assessment.pdf`, s. 1–2, özellikle bölüm 4.

| Görev | Gerekli gözlem ve girdiler | Gerekli sonuçlar | Bulunan gerçek verinin açığı |
|---|---|---|---|
| A: üretim / EKP | Aynı emre bağlı et, baharat, bağırsak ve diğer girdiler; ürün tipi; üretim öncesi fiyat / hizmet tarifi | Kabul edilen mamul ve birimi, fire, fiili maliyet; EKP = toplam maliyet / mamul | USDA pişirme özetlerinde emir, reçete miktarları, maliyet ve mamul kabul kaydı yok |
| B: parçalama | Hayvan/parti kimliği, tür, ölçülmüş karkas giriş ağırlığı; mevcutsa canlı ağırlık ve hayvan bilgileri | Aynı karkasın but, sırt, omuz, kaburga, kıyma, kuşbaşı, kemik, fire miktarları | Canlı/karkas ölçümleri bulundu; gerekli parça hedefleri birlikte yok |

Canlı→karkas randımanı, karkas→parça verimi ve pişirme randımanı farklı
hedeflerdir. Birbirinin yerine kullanılmadılar. PDF’nin 25 kg parçalama girdisi
API’de karkas girdisidir; belgede belirtilmeyen sabit bir canlı→karkas dönüşümü
uygulanmaz. PDF üretim örneği 100 kg **et** + 2 kg baharat + 5 kg bağırsaktır;
modelin et girdisi 107 kg değildir. Örneğin 85 paketi, yalnız örnekte açıkça
belirtilen 1 kg/paket nedeniyle 85 kg kabul edebiliriz.

## İndirilen ve temizlenen kaynaklar

| Kaynak / atıf | Gerçeklik kanıtı ve lisans | Doğrulanan dosya / gözlem | Entegrasyon ve karar |
|---|---|---|---|
| USDA ARS, *Table of Cooking Yields for Meat and Poultry*, Release 2 (2014), [DOI](https://doi.org/10.15482/USDA.ADC/1409031), [resmî katalog ve lisans](https://catalog.data.gov/dataset/usda-table-of-cooking-yields-for-meat-and-poultry) | Araştırmalardaki pişirme ağırlık değişimlerinin yayımlanmış özetleri. Katalog veri dosyaları için [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/) bildiriyor | [CSV](https://ndownloader.figshare.com/files/43365402), [veri sözlüğü](https://ndownloader.figshare.com/files/43365405), [PDF](https://ndownloader.figshare.com/files/43365399). **173 özet satır**; bunlar 173 üretim emri değildir | `ml/data/real/usda_cooking.csv`: tanım/yöntem ve pişirme verimi yüzdesi; mevcutsa n, SD, aralık ve nem/yağ değişimleri. Referans; A/B eğitimine alınmadı |
| Jones, Fleming, Griffith, Takahashi, Lee, Harris (2021), *North Wyke Farm Platform sheep production data*, [Mendeley V1](https://data.mendeley.com/datasets/xy3ndcy8jd/1), DOI 10.17632/xy3ndcy8jd.1 | [Veri makalesi](https://pmc.ncbi.nlm.nih.gov/articles/PMC8042253/) çiftlikte canlı ağırlık ve mezbahada soğuk karkas ölçümünü açıklıyor. Veri sürümü [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/); lisans ayrıca [snapshot](https://data.mendeley.com/public-api/datasets/xy3ndcy8jd/snapshot/1) içinde doğrulandı | [ZIP](https://data.mendeley.com/public-files/datasets/xy3ndcy8jd/files/20ced925-37ec-4991-80af-82b8ea2cc681/file_downloaded), 155.671 bayt. CSV 2.979 satır; biri açıklama, **2.978 kuzu**, 343 sütun | `ml/data/real/nwfp_lamb.csv`: 2.978 hayvan korunur; **2.607** geçerli canlı/karkas çifti, **371** eksik çift/tarih. Parça ağırlıkları yok. Sadece gerçek referans |
| Takele Wolkaro (2024), *Prediction of slaughter weight … Hararge Cattle at Haramaya University Abattoir, Ethiopia*, [Mendeley V1](https://data.mendeley.com/datasets/w9wbn9mpr8/1), DOI 10.17632/w9wbn9mpr8.1 | Kaynak, mezbahada alınan vücut ve sıcak karkas ölçümlerini tanımlıyor. [Snapshot](https://data.mendeley.com/public-api/datasets/w9wbn9mpr8/snapshot/1) lisansı [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/) | [XLSX](https://data.mendeley.com/public-files/datasets/w9wbn9mpr8/files/ef186adb-558d-472d-b95f-5e851eaddd43/file_downloaded), 17.395 bayt. Sheet1: **61 kayıt**, 21 sütun; diğer iki sayfa boş. Açıklamadaki 62 hayvanla uyuşmaz | `ml/data/real/hararghe_cattle.csv`: canlı, sıcak karkas, boş vücut ağırlığı; kaynak HDP ve ayrı hesaplanan karkas/canlı %. Kesim parça hedefleri yok; eğitime alınmadı |

Bu atıflar veri sahiplerinin projeyi onayladığı anlamına gelmez. CC BY verilerde
yapılan değişiklikler aşağıda belirtilmiştir. Her temiz satır `data_origin`,
`source_doi`, `license`, `source_row` taşır; yeni hayvan veya üretim kaydı üretilmez.

### Temizleme ve kalite bulguları

- USDA: başlıklardaki boşluklar ve tanım/yöntem kenar boşlukları temizlendi;
  yüzde birimleri korundu. Eksik n/SD/değerler tamamlanmadı. Bir özet satırı n
  adet sahte bağımsız gözleme çoğaltılmadı. Pişirme kaybından mamul kg üretilmedi.
- North Wyke: insan tarafından okunacak açıklama satırı çıkarıldı; kimlikler
  korundu (tekrarlı kuzu kimliği yok). Tarihler gün/ay/yıl olarak çözüldü.
  `W.yy.mm.dd` sütunlarından doğum ile satış arasındaki **son doğrudan tartım**
  seçildi; tarihi ve satışa uzaklığı saklandı. Sonradan yapılan ölçümler
  kullanılmadı. Dört/sekiz haftalık interpolasyonlar “ölçülmüş kesim canlı
  ağırlığı” diye yeniden adlandırılmadı. Son tartımın kesim günü tartımı olduğu
  varsayılmaz. Eksik kayıtlar referans tablosunda bırakılır; uygun çiftler ayrı
  bayrakla işaretlidir. Soğuk karkas ile sıcak karkas birleştirilmez.
- Hararghe: ilk sayfanın gerçek satır sayısı 61’dir; eksik olduğu düşünülen 62.
  hayvan eklenmedi. `100 × HCW / Wt` ile kaynak `HDP` arasında **56/61** satırda
  0,1 yüzde puandan fazla fark var. İki sütun ayrı korunur; kaynak HDP bir eğitim
  hedefi olarak kabul edilmez. Anlamı/bazı kesinleşmeyen `Bone`, projenin karkas
  kemiği hedefine eşlenmez. `source_row` hayvan kimliği değildir.

Null sayıları, sayısal dağılımlar, IQR dışı gözlem sayıları, tekrarlar ve
SHA-256’lar: `ml/reports/real_data_inventory.json`. IQR işaretleri tanısaldır;
gerçek uç gözlemler keyfî biçimde silinmez. SHA-256 dosya bütünlüğünü doğrular;
ölçümlerin bağımsız bir saha denetiminden geçtiği iddiası değildir.

## Araştırılan fakat eğitim için seçilmeyen diğer adaylar

| Kaynak | Ölçüm / hedef ve lisans durumu | Eleme / erişim sonucu |
|---|---|---|
| [Mendeley: Bovine carcasses Uruguay V2](https://data.mendeley.com/datasets/3jnn876my4/2), Corbellini ve ark., 2019 | Mezbahada gerçek 30.151 sığır gözlemi kaynak açıklamasında; CC BY 4.0. Karkas ağırlığı, yaş/diş, konformasyon, yağ ve fasciolosis | Canlı ağırlık ve gerekli parçalar sunulmuyor; bu çalışmada dosya indirilmedi, 30.151 indirilen satır sayısı olarak raporlanmaz |
| [Zenodo 4611489](https://zenodo.org/records/4611489), Zeng ve ark., 2021 | Gerçek mezbaha gözlem çalışması; hedef Sarcocystis türlerinin görülmesi. CSV ve açıklama listeleniyor | Parça verimi hedefi değil. Tarayıcı metninde lisans adı görünmedi; API HTTP 403 verdi. Yeniden kullanım lisansı doğrulanamadı, dosya entegre edilmedi |
| [Kaggle: Cattle Weight Detection, sürüm 3](https://www.kaggle.com/dsv/8858637), Acme AI/bhalo/mPower | Saha fotoğraflarıyla canlı ağırlık görevi; veri kartı CC BY 4.0 bildiriyor. Yaklaşık 47,85 GB görüntü arşivi | Karkas/parça etiketleri yok; hacimli arşiv indirilmedi. Veri kartı volumetrik tahmin ile canlı tartımın farklı olduğunu belirtiyor |
| [Kaggle: Livestock Growth and Weight Prediction](https://www.kaggle.com/muhammedfakl/livestock-growth-and-weight-prediction-dataset) | Kartta açıkça **sentetik**, CC0; 12 aylık büyüme hedefi | Gerçek veri olarak reddedildi |
| [CowDatabase3](https://github.com/ruchaya/CowDatabase3), [ilgili yayın](https://www.sciencedirect.com/science/article/pii/S0168169925003096) | Çalışma canlı ve kesim özelliklerini tanımlıyor; depo ölçüm/görüntü dosyaları içeriyor | İncelenen depo görünümünde açık veri lisansı bulunamadı; yayının erişimi depo verisine otomatik izin sayılmaz. Eğitim için indirilmedi |
| [Chalmers CPM: Hot-Dog production](https://cpmdatabase.cpm.chalmers.se/Scripts/sheet.asp?ActId=KI-2010-09-24-249) | 100 kg ürün için girdiler verilmiş, fakat sayfa veriyi türetilmiş ve reçeteyi modellenmiş olarak tanımlıyor; telif bildirimi var, açık veri lisansı doğrulanmadı | Gerçek emir bazlı girdi/çıktı ölçümü olarak kabul edilmedi |
| [Palumbo, Smith, Zaika 1976: pepperoni drying yield](https://ift.onlinelibrary.wiley.com/doi/abs/10.1111/j.1365-2621.1976.tb01149.x) | Akademik deney, kurutma sonrası ağırlık yüzdesi. Makale görünümü | Lisanslı indirilebilir emir bazlı ham dosya doğrulanamadı; makale sonuçlarından kayıt üretilmedi |
| [USDA ERS Meat Price Spreads](https://www.ers.usda.gov/data-products/meat-price-spreads), [Livestock and Meat Domestic Data](https://www.ers.usda.gov/data-products/livestock-and-meat-domestic-data) | Önceki projede kullanılan fiyat/istatistik kaynakları; bu oturumda yeniden indirilmedi | Fiyat ve ulusal istatistikler emir/hayvan eşleşmeli mamul/parça hedeflerini sağlamaz. Bu eğitimde yerel fiyat dosyası yoktu; manifest açıkça sabit 1,0 fiyat endeksi fallback’ini bildirir |

## Tekrarlama, erişim ve Git sınırı

Depo kökünden:

```powershell
.\ml\.venv\Scripts\python.exe -m ml.scripts.import_real_data
# İndirilen dosyalarla çevrimdışı tekrar:
.\ml\.venv\Scripts\python.exe -m ml.scripts.import_real_data --offline
```

İndirme yalnız yukarıdaki üç doğrulanmış veri dosyasını alır. Beklenen SHA-256
değişirse durur; yeni sürüm otomatik kabul edilmez. Mendeley dosya kimlikleri,
sürüm ve indirme URL’leri betikte sabittir; anonim erişim doğrulandı. Güvenli
ZIP okuması yalnız adı bilinen CSV’yi bellekte açar. Ham dosyalar
`ml/data/raw/verified_real/`, temiz gerçek tablolar `ml/data/real/` içindedir.
Metadata/araştırma kanıtları yerel `ml/data/raw/research/` altında tutulmuştur.
Veri dosyaları Git dışındadır; küçük envanter ve tekrar üretme kodu commit edilir.

Sentetik eğitim `ml/data/processed/synthetic_v2/` ve modeller
`ml/artifacts/reliability_v2/` altında, gerçek referanslardan ayrı kalır.
`train_group` gerçek referans satırlarını reddeder. Önceki izlenen veri/model
dosyaları tarihsel v1 olarak korunur; v1.1 API bunları yüklemez.

Gerçek eğitim için A’da emir bazlı miktar/birim/maliyet kayıtları; B’de aynı
hayvanın tam parça/fire ağırlıkları gerekir. Gelecekte North Wyke ile yardımcı
bir karkas modeli çalışılırsa dam/kardeş grupları ve yıllar arasında ayrım,
tartım–kesim gecikmesi ve sıcak/soğuk karkas tanımları ayrıca ele alınmalıdır.
Hararghe’de çiftlik/hayvan kimliği yoktur; bağımsız çiftlik testi iddia edilemez.
