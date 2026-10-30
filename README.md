# AVATAR Yapay Zeka Yarışması — ALPHA-DUO

Uçtan uca bir **gıda izlenebilirlik sistemi**: hammaddeden nihai satışa kadar
maliyet (EKP), stok, parçalama, lot (parti) izlenebilirliği, geri çağırma ve iki
zorunlu yapay zekâ tahmini. Gerçek bir gıda işletmesinde kullanılabilecek, çalışır
bir prototip.

---

## Neler yapıyor?

| # | Görev | Durum |
|---|-------|-------|
| 1 | Sözleşmeli üretim + gerçek maliyet (**EKP**), VKP, brüt kâr, marj | ✅ |
| 2 | Bütün kuzu **parçalama** + maliyet dağıtımı (**miktar/değer ağırlıklı**) + tutarsızlık tespiti | ✅ |
| 3 | **Lot izlenebilirliği** (ileri/geri) + **geri çağırma** + satış engeli | ✅ |
| — | İki zorunlu **KI modeli** (üretim verimi/EKP, parçalama verimi) | ✅ |
| — | **Android** satış + offline SQLite + senkronizasyon | ✅ |
| — | **React web** yönetim paneli (Food Intelligence, 10+ ekran, canlı veri, rapor indirme) | ✅ |
| — | Opsiyonel: fatura **OCR** + **anomali uyarıları** | ✅ |

Lot hareket tipleri (giriş/sevkiyat + **depo transferi, iade, yeniden işleme**) ve
**ikinci kademe dağıtım** (market → alt alıcı; yalnızca kaydedilmişse gösterilir)
de destekleniyor.

### Demo senaryosu (değerlendirme odağı)
`ET-2026-001` (100 kg et) → üretim emri `URT-2026-045` → `SUC-2026-045` (85 paket sucuk)
→ Market A/B/C (30/25/30). Kalite sorunu çıkınca **geri çağırma**, etkilenen marketleri
bulur, lotları bloke eder ve satışı engeller.

---

## Mimari

```
                    ┌─────────────────────┐
                    │   React Web (yönetim)│
                    │   rapor + geri çağırma│
                    └──────────┬──────────┘
                               │ REST
┌──────────────┐     ┌──────────▼──────────┐     ┌──────────────────┐
│ Kotlin Android│────▶│    Go REST API      │────▶│      MySQL 8     │
│  satış + sync │     │  (doğruluk kaynağı) │     │  (merkezi veri)  │
└──────────────┘     └──────────▲──────────┘     └──────────────────┘
                               │
                    ┌──────────┴──────────┐
                    │  Python ML servisi  │
                    │  FastAPI + sklearn  │
                    └─────────────────────┘
```

- **Go backend** — tek doğruluk kaynağı (REST, MySQL, parametreli SQL).
- **React web** — Food Intelligence yönetim paneli (`:5174`): genel bakış, stok, üretim,
  parçalama, satış, lot izlenebilirliği, geri çağırma, yapay zekâ, raporlar (CSV indirme),
  fatura OCR ve anomali uyarıları. Tüm ekranlar canlı backend verisi kullanır.
- **Kotlin Android** — satış, offline SQLite, senkronizasyon (outbox + pull/push).
- **Python ML** — iki lineer regresyon modeli + baseline karşılaştırma + MAE/RMSE.

---

## Teknoloji yığını

| Katman | Teknoloji |
|--------|-----------|
| Android satış/operasyon | Kotlin + Jetpack Compose + Room (SQLite) |
| Yerel veri | SQLite (Android) |
| Merkezi veri | MySQL 8 |
| Backend + API | Go (`net/http`, `go-sql-driver/mysql`) + REST |
| Web | React + Vite + TypeScript |
| KI servisi | Python 3.12 + FastAPI + scikit-learn |

---

## Klasör yapısı

```
backend/            Go REST API
  cmd/server/       sunucu giriş noktası
  internal/         config, db, model, repo, service, httpapi
  migrations/       MySQL şema + seed
mobile/             Kotlin Android (Compose + Room)
web/                React + Vite + TS
ml/                 Python ML servisi (FastAPI)
infra/              docker-compose (mysql, backend, sonarqube) + sonar-project.properties
docs/               veri modeli, API, tasarım, plan, junior görevleri
skills/             ajan becerileri (material-3, impeccable)
ALPHA-DUO-Vault/    Obsidian notları (karar günlüğü, görev dağılımı, günlükler)
```

---

## Hızlı başlangıç

### Tek komutla her şey (Docker)
```bash
cp .env.example .env
docker compose -f infra/docker-compose.yml up -d --build
```
- **Web (ham / API testi): `http://localhost:5173`** · **Web (Food Intelligence): `http://localhost:5174`** · API: `http://localhost:8080` · ML: `http://localhost:8000` · SonarQube: `http://localhost:9000`
- **Veritabanı (web arayüzü — Adminer): `http://localhost:8081`** (sunucu `mysql`, kullanıcı `avatar`, şifre `avatar`, db `avatar`)
- MySQL şeması ve seed verisi ilk açılışta otomatik uygulanır.

### Lokal geliştirme (Docker'sız)
```bash
cd backend && go run ./cmd/server                    # Go 1.23+ (API :8080)
cd web && npm install && npm run dev                 # Node 18+ (web :5173)
pip install -r ml/requirements.txt && uvicorn ml.app:app --reload   # Python 3.11+ (ML :8000, repo kökünden)
cd mobile && gradlew assembleDebug                   # JDK 17 + Android SDK 34 (APK)
```

### Testler
```bash
cd backend && go test ./...      # backend
cd ml && pytest                  # ml
cd web && npm run build          # web
```

### SonarQube
`infra/sonar-project.properties` hazır; talimatlar `docs/sonarqube.md` içinde.

---

## API özeti

| Metot | Yol | Açıklama |
|-------|-----|----------|
| GET | `/health` | sağlık |
| GET/POST | `/materials`, `/products` | katalog |
| GET/POST | `/lots` | lotlar |
| GET | `/lots/{code}`, `/lots/{code}/movements` | lot + hareketler |
| POST | `/stock-movements` | stok hareketi |
| GET/POST | `/production-orders`, GET `/production-orders/{id}` | üretim emri listesi/oluşturma + maliyet |
| GET/POST | `/butchery-batches`, GET `/butchery-batches/{id}` | parçalama listesi/oluşturma + dağıtım |
| GET | `/trace/forward?lot=`, `/trace/backward?lot=` | ileri/geri izleme |
| POST | `/recalls`, GET `/recalls/{id}` | geri çağırma + bloke |
| GET/POST | `/sales` | satış listesi/oluşturma (bloklu lot → 409) |
| GET | `/markets` | market listesi |
| GET/POST | `/redistributions` | ikinci kademe dağıtım (market → alt alıcı) |
| POST | `/sync/push`, GET `/sync/pull` | mobil senkronizasyon |
| GET | `/anomalies` | kural tabanlı anomali uyarıları |
| POST | `/ocr/invoice` | fatura/irsaliye metninden kalem çıkarma |

Tam referans: `docs/api.md`. Veri modeli: `docs/data-model.md`.

---

## KI (ML) servisi

Ayrıntılı model kartı: **`docs/ml.md`**.

| Endpoint (Go proxy) | Açıklama |
|----------|----------|
| `GET /ml/health` | ML sağlık durumu + yüklü modeller |
| `GET /ml/models` | Model sürümü, durumu, hedefleri |
| `GET /ml/report` | Model vs baseline metrikleri |
| `POST /ml/predict/v2/yield` | Üretim verimi, mamul ve tahmini EKP |
| `POST /ml/predict/v2/butchery` | Kuzu/sığır parça ağırlıkları |
| `POST /ml/predict/yield` | Eski sözleşme (uyumluluk) |
| `POST /ml/predict/butchery` | Eski sözleşme (uyumluluk) |

Modeller **sürümlü artefaktlardan yüklenir**; API açılışında yeniden eğitim
yapılmaz. Eğitim verisi `synthetic_calibrated` olduğu için model durumu
`synthetic_only`’dır ve metrikler gerçek saha başarısı olarak sunulmaz.
Gerçek kamu verisi olarak yalnız USDA fiyat serileri kullanılmıştır.

Lot izlenebilirliği ve geri çağırma **KI tahminine değil**, kesin kayıt
ilişkilerine dayanır.

---

## Dokümantasyon

- Tasarım: `docs/superpowers/specs/2026-10-24-avatar-ai-competition-design.md`
- Uygulama planı: `docs/superpowers/plans/2026-10-24-avatar-ai-competition.md`
- Veri modeli: `docs/data-model.md` · API: `docs/api.md`
- ER diyagramı: `docs/er-diagram.md` · DB inceleme: `docs/db-inspection.md`
- Demo scripti: `docs/demo-script.md` · SonarQube: `docs/sonarqube.md` · Logo: `docs/logo.svg`
- Android çalıştırma: `docs/android.md`
- Görev dağılımı: `ALPHA-DUO-Vault/00-Gorev-Dagilimi.md`

---

## Ajan becerileri (Skills)

Arayüz geliştirmede kullanılan beceriler `skills/` klasöründe projeye dahil edildi:

- **`skills/material-3/`** — Material Design 3 rehberi (renk, bileşen, tipografi, layout).
  Kaynak: <https://github.com/hamen/material-3-skill>
- **`skills/impeccable/`** — tasarım sözlüğü (polish/bolder/critique/audit) + 62 kuralılık
  slop dedektörü. Kaynak: <https://github.com/pbakaus/impeccable>

Web arayüzünün görsel dili: koyu orman yeşili (`#173f35`), kırık beyaz (`#f7f8f5`),
yumuşak gri ve ölçülü amber uyarılar.
