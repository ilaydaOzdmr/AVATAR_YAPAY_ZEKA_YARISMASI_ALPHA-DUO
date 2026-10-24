# Tasarım Dokümanı — AVATAR Yapay Zeka Yarışması (ALPHA-DUO)

- **Tarih:** 2026-10-24
- **Takım:** ALPHA-DUO (2 kişi)
- **Durum:** Taslak (onay bekliyor)

---

## 1. Niyet ve Neden

Gerçek bir gıda işletmesinde, hammaddeden nihai satışa kadar maliyet ve stok
hareketlerini yöneten, lot (parti) bazında tam izlenebilirlik sağlayan ve iki
zorunlu yapay zekâ tahminiyle desteklenen **uçtan uca çalışan bir prototip**
geliştirmek.

Neden: yarışma değerlendirmesi şu beş şeye bakar — doğru maliyet hesabı,
eksiksiz lot ilişkileri, güvenilir stok hareketleri, kullanılabilir Kotlin
uygulaması ve ölçülebilir KI katkısı. Demo özellikle "ET-2026-001 lotundan
etkilenen marketlerin bulunması" üzerine kurulacak.

Sonuç, PDF'deki tüm detayları içeren, canlı çalışan bir sistem ve bunun video
kaydı olmalı. 1 Kasım'a kadar teslim.

---

## 2. Hedefler, Hedef-dışı, Karşıt-hedefler

### Hedefler
1. Çalışır, uçtan uca entegre sistem: Go REST API + MySQL + React web + Kotlin
   Android + ML servisi.
2. Görev 1 — Sözleşmeli üretim ve gerçek maliyet (EKP): hammadde kartları,
   alış fiyatları, ölçü birimleri, lotlar, stok; üretim emri, hammadde sevki,
   hizmet maliyeti, mamul kabul, miktar farkları; EKP/VKP/brüt kâr/marj; geriye
   dönük inceleme. Örnek: 100 kg et + 2 kg baharat + 5 kg bağırsak + 85 × 5,50 ₺
   hizmet = 1.543,50 ₺ / 85 = **18,16 ₺ EKP**.
3. Görev 2 — Bütün kuzu parçalama + satış: parçalama reçetesi ve dağıtım
   anahtarı; fiili ağırlık kaydı; fire/yan ürünler; ağırlık dengesi; maliyetin
   miktar veya değer ağırlıklı anahtarla ürünlere dağıtılması (dağıtılan toplam
   = toplam maliyet); oluşan ürünlerin lot bazında stoklanıp Kotlin ile satışı.
   Örnek görseldeki tutarsızlıkların (25 kg × 11,50 ₺ = 287,50 ₺ ≠ 351,00 ₺;
   parça miktarları 22 kg topluyor) tespit edilip açıklanması.
4. Görev 3 — Lot izlenebilirliği ve geri çağırma: çoktan-çoğa lot ilişkileri
   (bir üretim emrinde çok hammadde lotu, bir lot birden çok mamul lotuna);
   ileri/geri izleme; geri çağırma ekranı (etkilenen ürünleri bloke etme,
   market bazlı miktar/sevkiyat/iletişim listesi); ikinci kademe dağıtım yalnızca
   kaydedildiyse kesin gösterim; Kotlin'de geri çağrılan lotların satışının
   engellenmesi. Zorunlu senaryo: ET-2026-001 (100 kg) → URT-2026-045 →
   SUC-2026-045 (85 paket) → Market A/B/C (30/25/30).
5. İki zorunlu KI modeli, referans yöntemle karşılaştırılmış ve hata metrikli:
   (A) üretim verimi + EKP tahmini, (B) parçalama verimi tahmini.
6. SQLite↔MySQL senkronizasyon yaklaşımı hem açıklanmış hem çalışır durumda.
7. Temel otomatik testler; SonarQube'den geçen temiz, İngilizce isimlendirilmiş,
   LLM-okunur kod; ayrık sıralı commit'ler ve branch'ler; Obsidian vault'a
   şeffaf karar/ilerleme notları.

### Hedef-dışı (bu projede yapılmayacak)
- Tam teşekküllü auth/RBAC sistemi (demo için basit/sıfır kimlik doğrulama).
- Üretim seviyesi ölçeklenebilirlik, yük testi, yüksek erişilebilirlik.
- Karmaşık CRDT/çakışma çözümü (basit ve açıklanabilir senkronizasyon yeterli).
- Fatura/irsaliye OCR ve anomali uyarıları **zorunlu değil** — "stretch goal",
  zaman kalırsa eklenir, aksi hâlde dokümanda açıklanır.

### Karşıt-hedefler (teknik olarak çalışsa bile başarısız sayılır)
- Maliyet hesabının (EKP/VKP/marj) yanlış olması.
- Lot ilişkilerinin eksik/yanlış olması (izlenebilirlik zincirinin kopması).
- Derlenmeyen veya açılmayan prototip.
- KI katkısının ölçülememesi (metrik/baseline olmaması).

---

## 3. Kısıtlar

- **Süre:** kod için 5 gün; ardından 2 gün anlama + video. Teslim 1 Kasım.
- **İnsan:** 2 kişi; kodun büyük bölümü AI ile üretilecek, insanlar inceleyip
  sahiplenecek.
- **Teknik (PDF zorunlu):** Kotlin (Android) + SQLite; MySQL; Go + REST API;
  React (veya uygun modern web); uygun ML yaklaşımı + API entegrasyonu.
- **Süreç (takım zorunlu):** ayrık ve sıralı commit'ler; branch kullanımı;
  Obsidian vault'a şeffaf not; clean code + İngilizce isimlendirme; Docker;
  SonarQube taraması.

---

## 4. Yaklaşımlar

Değerlendirilen üç yaklaşım:

| # | Yaklaşım | Takas |
|---|----------|-------|
| 1 | **Omurga önce** — tüm katmanları tek ince dikey dilimle bağla (ilk dilim geri çağırma senaryosu), sonra genişlet | Demo erken çalışır, kayarsak bile bütün var; her katman başta sığ |
| 2 | **Katman katman** — backend→web→Android→ML sırayla eksiksiz | Her parça tam; entegrasyon ve demo en sonda, kayma riski yüksek |
| 3 | **Minimal prototip** — Go+MySQL+React ağırlıklı, Kotlin ince, ML baseline | En hızlı; "kullanılabilir Kotlin" değerlendirme maddesi zayıf |

**Seçilen: Yaklaşım 1.** Gerekçe: "bitmiş ve eksiksiz" önceliğiyle en uyumlu;
demo omurgası en erken çalışır duruma gelir, kapsam genişlemesi kontrollü kalır,
kayma olursa Yaklaşım 3'e doğru kırpılır.

---

## 5. Kararlaştırılan "nasıl"

- **Monorepo**, aşağıdaki düzen (takım kararı):
  ```
  backend/    Go — REST API
  mobile/     Kotlin — Android (Jetpack Compose, SQLite, offline-first)
  web/        React — yönetim + raporlama + geri çağırma ekranı
  ml/         Python — KI servisi (FastAPI)
  infra/      docker-compose, migrations, SonarQube
  docs/       veri modeli, API, kurulum
  ALPHA-DUO-Vault/  Obsidian notları
  ```
- **İş bölümü (esnek):** Kişi A → backend + MySQL + ML + Docker/SonarQube +
  vault; Kişi B → React + Kotlin + senkronizasyon + vault.

---

## 6. Benim kararlarım (tasarımcının çağrıları)

Bu seçimleri ben yaptım; sakıncası varsa düzelt:

1. **ML servisi:** Python 3.12 + scikit-learn + FastAPI. Model A (üretim verimi +
   EKP) ve Model B (parçalama verimi) **lineer regresyon** tabanlı; her ikisi
   **basit baseline** (tarihsel ortalama) ile karşılaştırılır, MAE/RMSE raporlanır.
   Veri azlığı açıkça belirtilir. Tahminler lot ilişkilerine/geri çağırmaya
   **asla** dayanak yapılmaz.
2. **Senkronizasyon (SQLite↔MySQL):** Android offline-first. Yerel değişiklikler
   `outbox` tablosunda birikir; satır bazında `updated_at` + `client_id`;
   `last-write-wins` + açık dokümantasyon (tam çakışma motoru değil). Pull:
   referans verileri + bloke lotlar. Push: satış/satırlar. Basit, güvenilir,
   açıklanabilir.
3. **Veri modeli (özet, ayrıntı `docs/data-model.md`):**
   `materials` (hammadde), `products` (mamul/parça), `lots`, `suppliers`,
   `markets` (müşteri/market), `production_orders`, `production_order_inputs`
   (çoktan-çoğa lot↔emir), `production_order_outputs`, `butchery_batches`,
   `butchery_items`, `lot_links` (ebeveyn→çocuk lot + miktar), `stock_movements`
   (giriş, sevk, transfer, satış, iade, yeniden işleme, fire), `sales`,
   `sale_lines`, `recall_events`, `recall_blocks`, `sync_outbox`.
4. **Git/branch:** `main` korumalı (her zaman çalışır durumda). Her dilim/bileşen
   için kısa ömürlü branch (`feat/…`, `fix/…`). Conventional Commits
   (`feat:`, `fix:`, `docs:`, `chore:`, `test:`). Her mantıksal değişiklik ayrı
   commit; toplu tek commit yasak.
5. **Güvenlik:** tüm sorgular parametreli (SQL injection yok); `.env` ile secret
   yönetimi (git-ignore, örnek `.env.example`); API'de temel input doğrulama.
   Demo için kimlik doğrulama opsiyonel.
6. **Docker:** `docker-compose` ile MySQL + backend + ml + web + SonarQube ayağa
   kalkar; `docker compose up` tek komutla kurulum.
7. **Maliyet yöntemi (hammadde):** ağırlıklı ortalama (weighted average). Fire/
   atık/miktar farkı EKP'ye yedirilir: `EKP = (girdi maliyeti + hizmet maliyeti +
   diğer giderler) ÷ kabul edilen fiili mamul miktarı` (fire birim maliyeti
   yükseltir). Her `stock_movement` satırı `unit_cost` + `total_cost` taşır;
   maliyet geçmişi satır bazında geriye dönük izlenir.
8. **VKP ve formüller:** VKP = ürünün liste satış fiyatı (`products.list_price`,
   birim). Brüt kâr = `VKP − EKP` (birim). Marj = `(VKP − EKP) / VKP × 100`.
9. **Geri çağırma yetkisi:** sunucu otoriterdir. Go API gelen her satışı (push)
   güncel bloke-lot kümesine karşı doğrular ve geri çağrılan lotlu satışı açık
   hata ile reddeder; Kotlin ayrıca çevrimdışı kullanıcı deneyimi için önbellekteki
   bloke listesini kontrol eder ve sync sırasında reddi gösterir. Böylece "geri
   çağrılan lotların satışı engellenir" garantisi her durumda sağlanır.
10. **Görev 2 tutarsızlığı:** hem sistem içi tespit hem yazılı açıklama. Parçalama
    ekranı ağırlık dengesi bozukluğunu ve maliyet/toplam uyuşmazlığını programatik
    olarak işaretler (uyarı + doğrulama testleri); örnek görseldeki tutarsızlık
    ayrıca dokümanda açıklanır.
11. **İkinci kademe dağıtım ve iletişim listesi:** ikinci kademe = marketten alıcıya
    kaydedilmiş ileri sevkiyat (downstream). İletişim listesi = market bazlı tablo
    (market adı, etkilenen miktar, sevkiyat referansları, iletişim bilgisi) + üretilen
    geri çağırma mesajı metni.
12. **Sabit sayılar (fixture):** PDF'deki rakamlar birebir seed olarak sabitlenir
    (EKP 18,16 ₺ ve maliyet kırılımı; kuzu 25 kg × 11,50 ₺ = 287,50 ₺ ≠ 351,00 ₺
    örneği; parçalar toplamı 22 kg). Para 2 ondalık yuvarlanır.

---

## 7. Fiyat Kataloğu (önceden tanımlı, seed)

Tüm fiyatlar önceden belirlenir ve seed verisi olarak sabitlenir; PDF'deki
rakamlarla birebir tutarlı, gerçekçi bir katalog. Kullanıcı arayüzden
düzenleyebilir ama başlangıç kataloğu budur.

### Hammadde (alış fiyatı)
| Kod | Hammadde | Birim | Alış fiyatı |
|-----|----------|-------|-------------|
| MAT-001 | Et (dana) | kg | 9,70 ₺ |
| MAT-002 | Baharat karışımı | kg | 24,00 ₺ |
| MAT-003 | Bağırsak (kuzu) | kg | 11,60 ₺ |
| MAT-004 | Tuz | kg | 3,50 ₺ |

### Hizmet
| Kod | Hizmet | Birim | Bedel |
|-----|--------|-------|-------|
| SVC-001 | Sucuk üretim/parçalama hizmeti | paket | 5,50 ₺ |

### Mamul (VKP / liste satış fiyatı)
| Kod | Ürün | Birim | VKP |
|-----|------|-------|-----|
| PRD-001 | Narin Sucuk (1 kg) | paket | 24,90 ₺ |

### Parçalama (kuzu) — girdi ve ürünler
| Kod | Kalem | Birim | Fiyat |
|-----|-------|-------|-------|
| ANM-001 | Bütün kuzu (girdi, alış) | kg | 11,50 ₺ |
| PRD-101 | But | kg | 28,00 ₺ |
| PRD-102 | Sırt | kg | 32,00 ₺ |
| PRD-103 | Omuz | kg | 22,00 ₺ |
| PRD-104 | Kaburga | kg | 18,00 ₺ |
| PRD-105 | Kıyma | kg | 24,00 ₺ |
| PRD-106 | Kuşbaşı | kg | 26,00 ₺ |
| PRD-107 | Kemik / yan ürün (fire hariç) | kg | 2,00 ₺ |

> **Sabit senaryo kontrolü:** 100 kg et × 9,70 + 2 kg baharat × 24,00 + 5 kg
> bağırsak × 11,60 + 85 × 5,50 hizmet = 970,00 + 48,00 + 58,00 + 467,50 =
> **1.543,50 ₺ → 18,16 ₺ EKP**. VKP 24,90 ₺ → brüt kâr 6,74 ₺, marj ≈ %27,07.

---

## 8. Yapıcıya kalan (builder'ın kendi kararı)

- Somut endpoint imzaları ve HTTP yöntemleri.
- React sayfa/ekran tasarımı ve Kotlin Compose ekran detayları.
- ML feature-engineering detayı (özellik seçimi, ölçekleme).
- Hata mesajları, validasyon kuralları, seed veri miktarı.

---

## 9. Uygulama dilimleri (sıra)

> Detaylı adımlar ve commit'ler `writing-plans` aşamasında çıkarılacak; bu sıra
> kaba yol haritasıdır.

1. **Slice 0 — Temel:** monorepo iskeleti, docker-compose, SonarQube, vault notu.
2. **Slice 1 — Veri:** veri modeli, migrations, örnek senaryo seed verisi
   (ET-2026-001 zinciri).
3. **Slice 2 — Görev 1:** EKP maliyet hesabı + web ekranı (uçtan uca).
4. **Slice 3 — Görev 3 (demo odağı):** izlenebilirlik + ileri/geri izleme +
   geri çağırma ekranı (uçtan uca, marketleri bulma).
5. **Slice 4 — Görev 2:** parçalama + maliyet dağıtımı + tutarsızlık açıklaması.
6. **Slice 5 — KI:** Model A + B, baseline karşılaştırma, hata metrikleri.
7. **Slice 6 — Kotlin:** Android satış + offline SQLite + senkronizasyon + bloke
   lotların satışının engellenmesi.
8. **Slice 7 — Stretch:** OCR + anomali uyarıları (zaman kalırsa).
9. **Slice 8 — Kapanış:** testler, SonarQube temizliği, dokümantasyon, kurulum
   talimatı, demo scripti, video.

---

## 10. Testler

- Maliyet: EKP hesabı (18,16 ₺ örneği), VKP/brüt kâr/marj.
- Parçalama: dağıtılan toplam = toplam maliyet; ağırlık dengesi.
- İzlenebilirlik: ileri/geri izleme sorguları (ET-2026-001 → marketler).
- Geri çağırma: bloke mantığı; satış engeli.
- API: temel uçtan uca istekler.

---

## 11. Teslimatlar

- Çalışır prototip + kaynak kod + kurulum (`docker compose up`).
- Veri modeli + API dokümanı + örnek kayıtlar + temel otomatik testler.
- Üç görevin uçtan uca demosu (özellikle ET-2026-001 → marketler).
- KI yöntemleri açıklaması + örnek tahminler + doğruluk değerlendirmesi.
