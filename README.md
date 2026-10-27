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
| 2 | Bütün kuzu **parçalama** + maliyet dağıtımı + tutarsızlık tespiti | ✅ |
| 3 | **Lot izlenebilirliği** (ileri/geri) + **geri çağırma** + satış engeli | ✅ |
| — | İki zorunlu **KI modeli** (üretim verimi/EKP, parçalama verimi) | ✅ |
| — | **Android** satış + offline SQLite + senkronizasyon | ✅ |
| — | Opsiyonel: fatura OCR + anomali uyarıları | kapsam dışı |

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
- **React web** — yönetim, raporlama, izlenebilirlik ve geri çağırma ekranı.
- **Kotlin Android** — satış, offline SQLite, senkronizasyon (outbox + LWW).
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
ALPHA-DUO-Vault/    Obsidian notları (karar günlüğü, görev dağılımı, günlükler)
```

---

## Hızlı başlangıç

### Tek komutla her şey (Docker)
```bash
cp .env.example .env
docker compose -f infra/docker-compose.yml up -d --build
```
- Web: `http://localhost:5173` · API: `http://localhost:8080` · ML: `http://localhost:8000` · SonarQube: `http://localhost:9000`
- MySQL şeması ve seed verisi ilk açılışta otomatik uygulanır.

### Lokal geliştirme (Docker'sız)
```bash
cd backend && go run ./cmd/server                    # Go 1.23+ (API :8080)
cd web && npm install && npm run dev                 # Node 18+ (web :5173)
cd ml && pip install -r requirements.txt && uvicorn ml.app:app --reload   # Python 3.11+ (ML :8000)
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
| POST | `/production-orders`, GET `/production-orders/{id}` | üretim emri + maliyet |
| POST | `/butchery-batches`, GET `/butchery-batches/{id}` | parçalama + dağıtım |
| GET | `/trace/forward?lot=`, `/trace/backward?lot=` | ileri/geri izleme |
| POST | `/recalls`, GET `/recalls/{id}` | geri çağırma + bloke |
| POST | `/sales` | satış (bloklu lot → 409) |
| POST | `/sync/push`, GET `/sync/pull` | mobil senkronizasyon |

Tam referans: `docs/api.md`. Veri modeli: `docs/data-model.md`.

---

## KI (ML) servisi

| Endpoint | Açıklama |
|----------|----------|
| `POST /predict/yield` | üretim verimi + EKP tahmini |
| `POST /predict/butchery` | parçalama parça ağırlıkları tahmini |
| `GET /report` | model vs baseline MAE/RMSE |

Modeller **basit lineer regresyon**; her biri **ortalama (baseline)** ile karşılaştırılır.
Veri seti örnek/anonymize ve küçük olduğu için sonuçlar gösterim amaçlıdır (raporda belirtilir).
Lot izlenebilirliği ve geri çağırma **KI tahminine değil**, kesin kayıt ilişkilerine dayanır.

---

## Dokümantasyon

- Tasarım: `docs/superpowers/specs/2026-10-24-avatar-ai-competition-design.md`
- Uygulama planı: `docs/superpowers/plans/2026-10-24-avatar-ai-competition.md`
- Veri modeli: `docs/data-model.md` · API: `docs/api.md`
- Demo scripti: `docs/demo-script.md` · SonarQube: `docs/sonarqube.md` · Logo: `docs/logo.svg`
- Görev dağılımı: `ALPHA-DUO-Vault/00-Gorev-Dagilimi.md`
