# AVATAR Food Intelligence · ALPHA-DUO

**Hammaddeden üretime, üretimden satışa kadar maliyet ve lot takibi.**

AVATAR; gıda işletmelerinin stok, dış üretim, parçalama, satış ve geri çağırma işlemlerini aynı kayıt zincirinde buluşturan bir yarışma prototipidir. React yönetim paneli, Kotlin Android uygulaması ve iki zorunlu ML görevi, Go API üzerinden birlikte çalışır.

[Kurulum](#kurulum) · [Ekranlar](#ekranlar) · [Demo senaryosu](#demo-senaryosu) · [Yapay zekâ ve veri](#yapay-zekâ-ve-veri) · [Dokümantasyon](#dokümantasyon)

![AVATAR web yönetim paneli: stok değeri, üretim emirleri, ortalama EKP ve dikkat gerektiren kayıtlar](web-arayuz-gorsel/Ekran%20g%C3%B6r%C3%BCnt%C3%BCs%C3%BC%202026-11-01%20012337.png)

## Proje neyi çözüyor?

Çalışma, [yarışma değerlendirme belgesindeki](ALPHA-DUO-Vault/AVATAR_Assessment.pdf) üç ana görevi ve iki yapay zekâ görevini temel alır.

| İhtiyaç | Uygulamadaki karşılığı |
|---|---|
| **Dış üretim ve gerçek maliyet** | Gönderilen hammadde lotları, hizmet bedeli, mamul kabulü ve miktar farkları kaydedilir. Fiili mamule göre EKP; satış fiyatı (VKP), birim brüt kâr ve marj hesaplanır. |
| **Kuzu parçalama ve satış** | Giriş ağırlığı, parça ürünleri, fire ve maliyet dağılımı izlenir. Miktar veya değer ağırlıklı dağıtım yapılır; oluşan ürünler lotlarıyla stoklanır ve satılır. |
| **İleri/geri izlenebilirlik** | Hammaddeden etkilenen mamullere ve kayıtlı alıcılara; mamulden kaynak lotlara, tedarikçilere ve üretim emirlerine ulaşılır. |
| **Geri çağırma** | Etkilenen lotlar bloke edilir, alıcı/miktar/iletişim listesi hazırlanır. Bloklu lotların satışı backend tarafından reddedilir. |
| **Üretim ve parçalama tahmini** | Beklenen mamul, verim, tahmini EKP ve parça ağırlıkları gösterilir; kayıtlı gerçekleşen miktarlarla karşılaştırılır. |
| **Mobil operasyon ve raporlama** | Yerel satış kaydı, senkronizasyon, stok sorgulama, tahmin geçmişi ve altı başlıkta raporlama sunulur. |

Lot bazında depo transferi, sevkiyat, iade ve yeniden işleme hareketleri desteklenir. Marketten alt alıcıya dağıtım, yalnızca kaydı varsa kesin olarak gösterilir.

## Ekranlar

Görseller, projeye eklenen **1 Kasım 2026 tarihli ekran görüntüleridir**. İçlerindeki sayılar demo oturumuna aittir; yeni kurulumdaki kayıt sayıları ve sonuçlar farklı olabilir.

### Web yönetim paneli

| Ekran | Kullanıcı ne yapar? |
|---|---|
| Genel Bakış | Stok değeri, üretim durumu ve kural tabanlı uyarıları takip eder. |
| Stok & Hammaddeler | Hammadde/ürün kartı, raf ömrü, lot, son tüketim tarihi ve stok hareketi yönetir. |
| Üretim Emirleri | Girdileri ve mamul lotlarını ilişkilendirir; üretimi tamamlar veya reddeder, maliyet dökümünü inceler. |
| Kuzu Parçalama | Giriş partisini ve fiili parça ağırlıklarını kaydeder; ağırlık dengesi, fire ve maliyet dağılımını görür. |
| Satış & Sevkiyat | Mamul lotu, market ve miktar seçerek satış kaydeder. |
| Lot İzlenebilirliği / Geri Çağırma | Kayıt zincirini izler; etkilenen lot ve alıcıları bulur, geri çağırma başlatır. |
| Yapay Zekâ Merkezi | Üretim ve kuzu/sığır tahmini alır; model–gerçek karşılaştırmasını ve hata metriklerini inceler. |
| Raporlar | Alış, üretim maliyeti, parçalama, lot hareketi, satış ve geri çağırma raporlarını açar; CSV indirir veya yazdırma penceresinden PDF kaydeder. |
| Fatura / İrsaliye OCR | Yapıştırılan belge metninden ürün, miktar, lot ve maliyet önerileri çıkarır; onaylanan kalemleri stoğa işler. |

<details>
<summary><strong>Stok, üretim, parçalama ve satış ekranlarını görüntüle</strong></summary>

**Stok yönetimi:** kartlar, raf ömrü ve lotlar aynı ekrandadır.

![Web stok yönetimi ve ürün kartları](web-arayuz-gorsel/Ekran%20g%C3%B6r%C3%BCnt%C3%BCs%C3%BC%202026-11-01%20012423.png)

**Üretim emri:** hammadde ve mamul lotları işlemle ilişkilendirilir.

![Web yeni üretim emri ve kayıtlı üretimler](web-arayuz-gorsel/Ekran%20g%C3%B6r%C3%BCnt%C3%BCs%C3%BC%202026-11-01%20012528.png)

**Parçalama:** giriş, parçalar ve maliyet dağılımı birlikte takip edilir.

![Web parçalama kaydı ve geçmiş partiler](web-arayuz-gorsel/Ekran%20g%C3%B6r%C3%BCnt%C3%BCs%C3%BC%202026-11-01%20012545.png)

**Satış:** lot, market ve miktar üzerinden sevkiyat kaydı oluşturulur.

![Web hızlı satış formu ve son satışlar](web-arayuz-gorsel/Ekran%20g%C3%B6r%C3%BCnt%C3%BCs%C3%BC%202026-11-01%20012610.png)

</details>

<details>
<summary><strong>İzlenebilirlik ve geri çağırma ekranlarını görüntüle</strong></summary>

**İleri izleme:** kaynak lotun mamule ve marketlere uzanan ilişkisi gösterilir.

![Web lot izlenebilirliği ve etkilenen marketler](web-arayuz-gorsel/Ekran%20g%C3%B6r%C3%BCnt%C3%BCs%C3%BC%202026-11-01%20012625.png)

**Geri çağırma:** ürün lotları, alıcı miktarları ve teyit durumu listelenir.

![Web geri çağırma sonucu ve market iletişim listesi](web-arayuz-gorsel/Ekran%20g%C3%B6r%C3%BCnt%C3%BCs%C3%BC%202026-11-01%20012703.png)

Alıcı teyit düğmesi ekran üzerindeki takip durumunu değiştirir; dışarıya otomatik e-posta/SMS gönderildiği anlamına gelmez.

</details>

<details>
<summary><strong>Yapay zekâ, raporlar ve belge okuma ekranlarını görüntüle</strong></summary>

**Tahminler:** beklenen mamul, verim, tahmini EKP ve parça ağırlıkları.

![Yapay Zekâ Merkezi üretim ve parçalama tahminleri](web-arayuz-gorsel/Ekran%20g%C3%B6r%C3%BCnt%C3%BCs%C3%BC%202026-11-01%20012720.png)

**Karşılaştırma:** model tahmini ve gerçekleşen miktar aynı grafikte gösterilir.

![Üretim emirleri için model tahmini ve gerçekleşen miktar grafiği](web-arayuz-gorsel/Ekran%20g%C3%B6r%C3%BCnt%C3%BCs%C3%BC%202026-11-01%20012730.png)

**Raporlar:** kayıtlar konuya göre ayrılır; CSV ve PDF çıktısı alınabilir.

![Web alış raporu, rapor sekmeleri ve dışa aktarma seçenekleri](web-arayuz-gorsel/Ekran%20g%C3%B6r%C3%BCnt%C3%BCs%C3%BC%202026-11-01%20012759.png)

**Belge metni:** önerilen kalemler kullanıcı onayından sonra kaydedilir. Mevcut akış metin ayrıştırır; görsel yükleyip otomatik OCR yapan ayrı bir motor içermez.

![Fatura metninden çıkarılan ürün, miktar, lot ve maliyet önerisi](web-arayuz-gorsel/Ekran%20g%C3%B6r%C3%BCnt%C3%BCs%C3%BC%202026-11-01%20013147.png)

</details>

### Android operasyon uygulaması

Ana sayfadan stoklara, satışa, tahminlere ve raporlara geçilir. Satışta ürünün **lotu, marketi ve miktarı** seçilir; bağlantı/senkronizasyon sonucu ekranda gösterilir.

<table>
  <tr>
    <th>Ana sayfa ve bağlantı</th>
    <th>Başarılı satış</th>
    <th>Bloklu lotun reddi</th>
  </tr>
  <tr>
    <td><img src="mobil-arayuz-gorsel/WhatsApp%20Image%202026-11-01%20at%2001.04.07.jpeg" alt="Android ana sayfası, sunucu bağlantısı ve operasyon kısayolları" width="220"></td>
    <td><img src="mobil-arayuz-gorsel/WhatsApp%20Image%202026-11-01%20at%2001.06.08.jpeg" alt="Android satış formunda başarılı senkronizasyon mesajı" width="220"></td>
    <td><img src="mobil-arayuz-gorsel/WhatsApp%20Image%202026-11-01%20at%2001.05.17.jpeg" alt="Android satış formunda bloklu lotun reddedilmesi" width="220"></td>
  </tr>
</table>

<details>
<summary><strong>Mobil tahmin, PDF raporu ve satış geçmişini görüntüle</strong></summary>

<table>
  <tr>
    <th>Parçalama tahmini</th>
    <th>Rapor ve PDF indirme</th>
    <th>Satış geçmişi</th>
  </tr>
  <tr>
    <td><img src="mobil-arayuz-gorsel/WhatsApp%20Image%202026-11-01%20at%2001.04.08%20%282%29.jpeg" alt="Android kuzu parçalama tahmini, parça ağırlıkları ve verim" width="220"></td>
    <td><img src="mobil-arayuz-gorsel/WhatsApp%20Image%202026-11-01%20at%2001.04.11.jpeg" alt="Android rapor tablosu ve PDF dosyasının indirilme bildirimi" width="220"></td>
    <td><img src="mobil-arayuz-gorsel/WhatsApp%20Image%202026-11-01%20at%2001.04.12%20%281%29.jpeg" alt="Android satış geçmişi ve bekleyen satışları senkronize etme" width="220"></td>
  </tr>
</table>

Mobilde Room/SQLite, çevrimdışı önbellek ve gönderim kuyruğudur. Bağlantı yokken satış yerelde bekler; sunucu kabul etmeden kesinleşmiş merkezi satış sayılmaz. Sunucuya yeniden gönderimde blokaj kontrolü yapılır. Aynı satışın tekrar gönderilmesi çift stok düşümüne yol açmamalıdır; satış kodu tekrar gönderim kontrolünde kullanılır.

Tahmin için sunucu gerekir; çevrimdışıyken yeni tahmin üretilmez. Ana ekrandaki barkod/QR kısayolu mevcut prototipte simülasyondur.

</details>

Tüm görseller: [web galerisi](web-arayuz-gorsel/) · [mobil galeri](mobil-arayuz-gorsel/).

## Kurulum

### 1. Web, API ve ML servislerini başlatın

Docker Desktop ve Docker Compose gereklidir. Aşağıdaki komutları **proje kökünde**, PowerShell ile çalıştırın. `.env` yoksa örnek dosyadan oluşturun:

```powershell
Copy-Item .env.example .env
```

Servisleri derleyip başlatın:

```powershell
docker compose --env-file .env -f infra/docker-compose.yml up -d --build mysql ml backend web-m3
```

İlk derleme Python bağımlılıklarını kurar ve demo modellerini eğitir; biraz zaman alabilir. MySQL şeması ve örnek kayıtlar **ilk veritabanı oluşturulurken** yüklenir. Uygulama sonraki açılışlarda mevcut kayıtları kullanır.

| Adres | Kullanım |
|---|---|
| **[localhost:5174](http://localhost:5174)** | **Ana web arayüzü: Food Intelligence** |
| [localhost:8080/health](http://localhost:8080/health) | Go API sağlık kontrolü |
| [localhost:8080/ml/health](http://localhost:8080/ml/health) | ML bağlantısı ve yüklenen modeller |
| [localhost:8000/docs](http://localhost:8000/docs) | FastAPI servis dokümantasyonu |

Adresler varsayılan portlara göredir.

Kod güncel olduğu hâlde eski ekran görünüyorsa aynı `up -d --build` komutunu yeniden çalıştırın ve tarayıcıyı `Ctrl + F5` ile yenileyin.

<details>
<summary>İsteğe bağlı araçlar ve yerel geliştirme</summary>

Ham/API test paneli, veritabanı görüntüleyicisi ve SonarQube ayrıca başlatılabilir:

```powershell
docker compose --env-file .env -f infra/docker-compose.yml up -d --build web adminer sonarqube
```

- Ham yönetim paneli: `http://localhost:5173` — ana Food Intelligence panelinden ayrı bir giriş.
- Adminer: `http://localhost:8081` — varsayılan demo ayarlarıyla sunucu `mysql`, kullanıcı/veritabanı `avatar`, şifre `avatar`.
- SonarQube: `http://localhost:9000` — servis başlatmak taramayı otomatik çalıştırmaz; [tarama kılavuzunu](docs/sonarqube.md) izleyin.

Web'i Vite ile geliştirmek için Node.js 20+ kullanın:

```powershell
cd web
npm.cmd ci
npm.cmd run dev
```

Vite'ın terminalde gösterdiği adresin **`/m3.html`** yolunu açın. API ve ML servislerinin ayrıca çalışıyor olması gerekir. Docker ile başlatma bu ek ayarları gerektirmez.

ML'yi Python ortamında başlatacaksanız önce [model hazırlama ve çalıştırma adımlarını](docs/ml.md) uygulayın; model dosyası yoksa servis tahmin için `503` döner.

</details>

### 2. Android uygulamasını çalıştırın

Android Studio, JDK 17 ve Android SDK 34 gereklidir. SDK yolu `mobile/local.properties` içinde yerel makineye göre ayarlanır. API ve ML servisleri açıkken proje kökünden debug APK oluşturun:

```powershell
cd mobile
.\gradlew.bat :app:assembleDebug
adb install -r app/build/outputs/apk/debug/app-debug.apk
```

Ana ekrandaki **Sunucu → Değiştir** alanından bağlantıyı kontrol edin:

| Ortam | Backend adresi |
|---|---|
| Telefon ve bilgisayar aynı Wi-Fi ağında | `http://<PC-IP>:8080` |
| Android emülatörü | `http://10.0.2.2:8080` |
| USB bağlantısı | `adb reverse tcp:8080 tcp:8080` sonrası `http://127.0.0.1:8080` |

**Dene** ile bağlantıyı yeniden kontrol edebilirsiniz. Wi-Fi erişiminde bilgisayarın güvenlik duvarı 8080 portuna izin vermelidir. Ayrıntılar: [Android çalıştırma](docs/android.md).

> **Mevcut kurulum debug/demo içindir.** Debug sürümü yerel HTTP bağlantısına izin verir. Release sürümü HTTP'yi engeller; release için çalışan bir HTTPS backend adresi ve geçerli sertifika yapılandırması henüz tamamlanmamıştır.

## Demo senaryosu

Yarışma PDF'sindeki ana akış:

```text
ET-2026-001 · 100 kg et
        ↓
URT-2026-045 · dış üretim emri
        ↓
SUC-2026-045 · 85 adet 1 kg sucuk
        ↓
Market A: 30 · Market B: 25 · Market C: 30 paket
```

1. **Üretim Emirleri:** `URT-2026-045` kaydının girdilerini, mamul kabulünü ve gerçek maliyetini inceleyin.
2. **Lot İzlenebilirliği:** `ET-2026-001` için ileri izleme yapın; mamul lotunu ve market bağlantılarını bulun. `SUC-2026-045` için geri izlemeyi deneyin.
3. **Geri Çağırma:** kaynak lotu ve sebebi girin. Etkilenen lotlar bloke edilir, marketlerin miktar/iletişim listesi gösterilir.
4. **Android satış:** bloklu lot için satışın reddedildiğini, uygun bir lot için normal satış ve senkronizasyonu gösterin.
5. **Yapay Zekâ Merkezi:** üretim/parçalama tahmini alın ve kayıtlı gerçekleşen sonuçlarla karşılaştırın.

Geri çağırma demo verisinin durumunu değiştirir. Aynı kurulumda önceki satış ve geri çağırmalar varsa sonuçlar başlangıç örneğinden farklı olabilir.

<details>
<summary>PDF'deki maliyet hesabı ve parçalama tutarsızlıkları</summary>

| Maliyet unsuru | Hesap | Tutar |
|---|---|---:|
| Et | 100 kg × 9,70 € | 970,00 € |
| Baharat | 2 kg × 24,00 € | 48,00 € |
| Bağırsak | 5 kg × 11,60 € | 58,00 € |
| Hizmet | 85 paket × 5,50 € | 467,50 € |
| **Toplam** | | **1.543,50 €** |

**Gerçek EKP = 1.543,50 / 85 = 18,16 €/paket.** Hammadde bedelleri işletmeye ait olduğu için yalnızca hizmet faturasına bakılarak maliyet hesaplanmaz. Bu örnekte paketler 1 kg'dır; genel kullanımda paket ile kg birbirinin yerine geçirilmez.

Parçalama örneğinde **25 kg × 11,50 € = 287,50 €** eder; belgede anılan **351,00 €** ile aynı değildir. Parçaların **22 kg** toplamı da **3 kg** fark bırakır. Uygulama maliyet farkını ve ağırlık dengesini kontrol eder; bu değerleri sessizce birbirine eşitlemez.

</details>

## Yapay zekâ ve veri

| Görev | Nasıl çalışır? |
|---|---|
| **A · Üretim verimi ve EKP** | Et ve diğer üretim girdilerinden verim tahmin edilir. Beklenen mamul, et miktarı × tahmini verim ile; tahmini EKP ise maliyet hesabıyla türetilir. EKP ayrı bir ML hedefi değildir. |
| **B · Parçalama verimi** | Kuzu ve sığır için ayrı modeller parça ağırlıkları üretir. Parçalar ölçülmüş karkas ağırlığına dengelenir; satılabilir toplam ve verim bu parçalardan hesaplanır. |

**Model sürümü `1.1.0`; eğitim durumu `synthetic_only`.** Eğitimde 480 sentetik üretim kaydı, 800 kuzu ve 800 sığır kaydı kullanılır. Gerçek kamu ölçümleri bulunup ayrı referans tablolarına aktarılmıştır; ancak yarışmanın üretim/maliyet ve parça hedeflerini birlikte içermedikleri için eğitime karıştırılmamıştır.

Eğitim/test ayrımı üretim partisi veya çiftlik gruplarına göre yapılır. Model seçimi eğitim içi çapraz doğrulama ile yürütülür; ayrılan test verisinde MAE, RMSE, R² ve baseline karşılaştırması raporlanır. Bu metrikler **sentetik test başarısını** anlatır, gerçek işletme başarısını kanıtlamaz. Model kısıtları ve veri lisansları: [model kartı](docs/ml.md) · [veri kaynakları](docs/DATA_SOURCES.md).

Web ve mobil tahminleri Go API üzerinden alır. Model dosyaları API açılışında yüklenir; her tahmin isteğinde yeniden eğitim yapılmaz. ML servisine veya modele erişilemezse hata gösterilir.

**Lot izlenebilirliği, geri çağırma ve satış blokajı tahminlere değil, kesin kayıt ilişkilerine dayanır.**

## Mimari ve teknoloji

```mermaid
flowchart LR
    Web[React web paneli] --> API[Go REST API]
    Android[Kotlin Android] --> API
    Android <--> Local[(Room / SQLite)]
    API <--> DB[(MySQL 8)]
    API --> ML[Python ML / FastAPI]
```

| Katman | Teknoloji ve sorumluluk |
|---|---|
| Web | React, TypeScript, Vite — yönetim, karşılaştırma ve raporlama |
| Mobil | Kotlin, Jetpack Compose, Room/SQLite — satış, önbellek ve gönderim kuyruğu |
| Backend | Go `net/http`, REST API — iş kuralları, stok, maliyet ve senkronizasyon |
| Merkezi veri | MySQL 8 — kalıcı ve merkezi kayıtlar |
| ML | Python 3.12, FastAPI, scikit-learn — eğitim, değerlendirme ve tahmin |
| Altyapı / kalite | Docker Compose, Nginx, SonarQube — servis çalıştırma ve statik analiz |

Merkezi kayıt kaynağı MySQL'dir; mobil SQLite bağımsız bir merkezi stok sistemi değildir. Senkronizasyonun veri modeli ve tekrar gönderim yaklaşımı [mobil senkronizasyon belgesinde](docs/mobile-sync.md) açıklanır.

<details>
<summary>Geliştiriciler için klasörler, API ve testler</summary>

```text
backend/              Go API, iş kuralları, MySQL şeması ve örnek kayıtlar
web/                  Ana Food Intelligence ve ham yönetim paneli
mobile/               Kotlin Android uygulaması
ml/                   Veri hazırlama, eğitim, değerlendirme ve FastAPI
infra/                Docker Compose ve SonarQube yapılandırması
docs/                 API, veri modeli, kurulum ve doğrulama belgeleri
web-arayuz-gorsel/    Web ekran görüntüleri
mobil-arayuz-gorsel/  Android ekran görüntüleri
ALPHA-DUO-Vault/      Yarışma PDF'si ve proje notları
```

Öne çıkan API yolları: `/production-orders`, `/butchery-batches`, `/lots`, `/trace/forward`, `/trace/backward`, `/recalls`, `/sales`, `/sync/pull`, `/sync/push` ve `/reports/summary`.

Tahminler: `POST /ml/predict/v2/yield`, `POST /ml/predict/v2/butchery`. Model raporu: `GET /ml/report`. Tam sözleşmeler: [API referansı](docs/api.md) ve [ML model kartı](docs/ml.md).

Temel kontroller proje kökünden çalıştırılabilir. Python bağımlılıkları ve Android SDK/JDK önceden kurulmuş olmalıdır:

```powershell
go -C backend test ./...
python -m pytest ml/tests ml/test_model.py -q
npm.cmd --prefix web ci
npm.cmd --prefix web run build
```

Android JVM testleri için `mobile/` içinde `.\gradlew.bat :app:testDebugUnitTest` çalıştırılır. MySQL entegrasyonu, Chrome testleri ve kapsam raporlarının üretimi için [ayrıntılı test/tarama adımlarını](docs/sonarqube.md) izleyin.

31 Ekim 2026 tarihli son kayıtlı SonarQube analizinde kalite kapısı geçti; genel kapsam **%62,0**, yeni kod kapsamı **%83,6** olarak ölçüldü. Bu iki oran farklıdır. Sonuç ve test sınırlılıkları [SonarQube raporunda](docs/SONARQUBE_REPORT.md) kayıtlıdır; statik analiz sonucu gerçek cihaz/release doğrulaması yerine geçmez.

</details>

## Dokümantasyon

| Belge | İçerik |
|---|---|
| [Yarışma PDF'si](ALPHA-DUO-Vault/AVATAR_Assessment.pdf) | Görevler, örnek kayıtlar ve değerlendirme ölçütleri |
| [Veri modeli](docs/data-model.md) · [ER diyagramı](docs/er-diagram.md) | Tablolar, lot ilişkileri ve veri yapısı |
| [API referansı](docs/api.md) | Servis yolları ve istek/yanıt sözleşmeleri |
| [Android kurulumu](docs/android.md) · [Senkronizasyon](docs/mobile-sync.md) | Cihaz bağlantısı, yerel kayıtlar ve veri aktarımı |
| [ML model kartı](docs/ml.md) · [Veri kaynakları](docs/DATA_SOURCES.md) | Modeller, sentetik/gerçek veri ayrımı, lisanslar ve sınırlılıklar |
| [ML inceleme ve düzeltmeler](docs/ML_AUDIT_AND_IMPROVEMENTS.md) | Model doğrulaması ve fiziksel tutarlılık kontrolleri |
| [SonarQube kılavuzu](docs/sonarqube.md) · [Sonuç raporu](docs/SONARQUBE_REPORT.md) | Test kapsamı, analiz adımları ve kayıtlı sonuçlar |
| [Güvenlik incelemesi](docs/SECURITY_REVIEW.md) | Container kullanıcıları ve Android geliştirme ayarları |
