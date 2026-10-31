# SonarQube düzeltme ve test sonucu

- Son tarama (UTC): 2026-10-31T20:41:57.711494+00:00
- İncelenen kaynak revizyonu: `2ad4f658e5265b0d70eb8f6a700f97235bb1dbaf`.
- SonarQube: 9.9.8.100196; SonarScanner CLI: 8.1.0.6389.
- Analiz kimliği: `AaGTsndw_F98FOpq193G`.
- [SonarQube paneli](http://localhost:9000/dashboard?id=avatar-alpha-duo).

| Ölçüm | İlk durum | Son durum |
| --- | --- | --- |
| Quality Gate | Eksik coverage ile OK | **OK** |
| Açık bug | 0 | 0 |
| Açık güvenlik açığı bulgusu | 0 | 0 |
| Açık kod kalitesi bulgusu | 75 (29 Critical) | 0 |
| İncelenmemiş güvenlik noktası | 11 | 0 |
| Güvenlik inceleme oranı | %0 | %100.0 |
| Genel test kapsamı | Rapor aktarılmamıştı | %62.0 |
| Yeni kod test kapsamı | Rapor aktarılmamıştı | %83.6 |
| Tekrarlanan satır | %0,4 | %0.4 |

Kalite kapısının %80 yeni kod kapsamı, %100 güvenlik incelemesi ve %3 tekrar
eşikleri değiştirilmedi. Üç yeni kod kalite notu A'dır. Genel kapsam ile yeni kod
kapsamı farklıdır; tablodaki yüzde projenin tamamının %80 test edildiğini söylemez.

## Değişiklikler

- Karmaşık Go stok/üretim/parçalama ve raporlama işlevleri aynı transaction
  sınırları içinde daha küçük işlemlere ayrıldı. Depolama hatalarında rollback
  ve eksik veri akışında hata yayılımı test edildi.
- ML eğitim/değerlendirme işlevleri ayrıldı; grup bazlı ayırma, eğitimden
  hesaplanan baseline ve dağıtımda kullanılan fiziksel denklikler korunur.
  Uçtan uca eğitim testi özellik adları ile veri tablosunun karışmasını yakalar.
- React listeleri kalıcı satır kimlikleri kullanır. Android ekran/iş kuralı
  kodu ayrıldı; Room sütunları `ColumnInfo` ile, JSON anahtarları aynen korunur.
- Container süreçleri root yerine yetkisiz kullanıcılarla çalışır. Geniş
  dosya kopyaları açık dosya/dizin listeleriyle değiştirildi.
- Android yedekleme kapalıdır. Açık HTTP yalnız yerel geliştirme/debug için
  izinlidir. İki geliştirme hotspot'u gerekçeleriyle SAFE olarak incelendi;
  diğer dokuz nokta kodla giderildi. [İnceleme kaydı](SECURITY_REVIEW.md).

## Doğrulama

- 51 Python testi; 8 Node birim testi ve gerçek Chrome UI testleri geçti.
- Go birim, SQL hata/rollback, MySQL entegrasyonu ve canlı ML proxy testleri geçti.
  Entegrasyon testleri ayrı geçici veritabanında çalışır ve bu veritabanını siler.
- 21 Android JVM/Robolectric testi, JaCoCo raporu, debug APK ve release manifest
  birleştirmesi geçti. Sunucu/model hataları, ekran akışları, yerel ağ politikası,
  satış engelleme ve rapor metin yerleşimi kontrol edildi.
- Backend, ML ve iki web image'ı derlendi. Çalışan süreçlerin UID'leri sırasıyla
  100, 10001, 101, 101; MySQL/veri volume'ları yeniden oluşturulmadı.
- Gerçek localhost:5174 sayfasında sürüm 1.1.0 üretim tahmini 84.523 kg,
  kuzu karkas/parça toplamı 21/21 kg olarak doğrulandı. İstenen görünüm korundu;
  390 px ekranda sayfa taşması ve JavaScript hatası yok.

## Kapsam ve sınırlar

- Kapsam raporları Go coverprofile, Python Cobertura, web LCOV ve Android
  JaCoCo çıktılarıdır. Windows/modül yolları gerçek ölçüm sayaçları değiştirilmeden
  repo köküne eşlenir; bu eşlemenin testleri vardır.
- Tarayıcı testleri canlı yerel ML/backend ile çalışır. Rapor tablosunun bazı
  durumlarında yalnız test için QA fixture yanıtı kullanılır; işletme kaydı
  oluşturulmaz. Android ağ testleri yerel MockWebServer kullanır; LAN cihazlarına
  test trafiği gönderilmez.
- Robolectric native PdfDocument çıktısını desteklemediğinden gerçek PDF dosyası
  üretimi JVM'de doğrulanmadı; metin/sütun yerleşimi ve başarısız dışa aktarmada
  arayüzün toparlanması test edildi.
- Eski Python analizörü 3.12 hedefini 3.10 kurallarıyla analiz eder. Web analizörü
  Node 22 için uyarı verir; tüm 30 TypeScript dosyası Sonar'a özel yapılandırmayla
  analiz edilmiştir. Uygulamanın Vite/TypeScript derleme ayarları değiştirilmedi.
- Statik analiz, bağımlılık CVE taraması veya ML saha doğrulaması yerine geçmez.
  Modeller `synthetic_only` durumundadır; yeni gerçek veri/eğitim iddiası yoktur.
- Anahtarlar, yedekler, ham test/kapsam çıktıları ve büyük dosyalar Git dışındadır.
  Testlerin yeniden çalıştırılması için [talimatlar](sonarqube.md).
