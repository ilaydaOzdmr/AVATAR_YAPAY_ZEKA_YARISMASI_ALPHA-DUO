# AVATAR AI Competition — ALPHA-DUO Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Uçtan uca çalışan bir gıda izlenebilirlik sistemi: hammaddeden satışa maliyet (EKP), parçalama, lot izlenebilirliği + geri çağırma, iki KI tahmini, Go API + MySQL + React web + Kotlin Android.

**Architecture:** Monorepo; Go REST API tek doğruluk kaynağı (MySQL), React web yönetim/raporlama, Kotlin Android offline-first satış + senkronizasyon, Python ML servisi (FastAPI). Dikey dilimler: önce veri + maliyet, sonra izlenebilirlik/geri çağırma (demo odağı), parçalama, ML, web, Android.

**Tech Stack:** Go (stdlib `net/http` + `database/sql`), MySQL 8, React 18 + Vite + TypeScript, Kotlin + Jetpack Compose + Room (SQLite), Python 3.12 + FastAPI + scikit-learn, Docker Compose, SonarQube.

**Spec:** `docs/superpowers/specs/2026-10-24-avatar-ai-competition-design.md`

## Global Constraints

- Tüm fiyatlar önceden tanımlı (seed); PDF rakamları birebir sabit (EKP 18,16 ₺; kuzu 11,50 ₺/kg).
- Kod İngilizce isimlendirme, LLM-okunur, clean code.
- Ayrık + sıralı commit'ler (Conventional Commits); branch `main` korumalı.
- Secret'lar `.env` (git-ignore); `.env.example` örnek.
- Tüm SQL sorguları parametreli (injection yok).
- Para 2 ondalık; maliyet yöntemi ağırlıklı ortalama; fire EKP'ye yedirilir.
- Geri çağırma sunucu-otoriter; lot izlenebilirliği/geri çağırma asla KI tahminine dayanmaz.
- Her görev sonunda çalışır durumda, test edilebilir yazılım.

## Review Focus

1. **Maliyet doğruluğu** — EKP/VKP/brüt kâr/marj yanlış olursa proje başarısız. Test: Task 4 (18,16 ₺).
2. **Dağıtılan toplam = toplam maliyet** — parçalama maliyeti eksik dağıtılırsa izlenemez. Test: Task 6.
3. **İzlenebilirlik zinciri kopukluğu** — çoktan-çoğa lot ilişkisi eksik olursa marketler bulunamaz. Test: Task 5 (forward/backward).
4. **Bloke lot satışı** — geri çağrılan lot çevrimdışı satılırsa sunucu reddetmeli. Test: Task 9 + Task 5.
5. **Çevrimdışı senkronizasyon tutarlılığı** — aynı satır iki cihazdan çakışırsa LWW tutarlı olmalı. Test: Task 9.

---

## File Structure

```
backend/
  cmd/server/main.go                 # sunucu başlatma, router bağlama
  internal/config/config.go          # env yükleme
  internal/db/db.go                  # MySQL bağlantı + migration runner
  internal/model/model.go            # domain struct'ları (Material, Product, Lot, ...)
  internal/repo/repo.go              # MySQL repository'leri (SQL)
  internal/service/cost.go           # EKP/VKP/brüt kâr/marj (saf fonksiyonlar)
  internal/service/distribution.go   # maliyet dağıtımı (saf fonksiyonlar)
  internal/service/traceability.go   # ileri/geri izleme (repo arayüzü üzerinde)
  internal/service/recall.go         # geri çağırma + bloke doğrulama
  internal/httpapi/server.go         # router, handler'lar
  internal/httpapi/dto.go            # request/response tipleri
  migrations/0001_init.up.sql        # şema
  migrations/0001_seed.up.sql        # fiyat kataloğu + senaryo verisi
  go.mod
mobile/                              # Kotlin Android (Compose + Room)
web/                                 # React + Vite + TS
ml/                                  # Python FastAPI + sklearn
infra/
  docker-compose.yml                 # mysql, backend, ml, web, sonarqube
  sonar-project.properties
docs/
  data-model.md                      # şema + ilişkiler (Task 2'de üretilir)
  api.md                             # endpoint referansı (Task 3'te üretilir)
ALPHA-DUO-Vault/                     # Obsidian karar/ilerleme notları
```

---

## Task 1: Repo iskeleti + Docker + SonarQube

**Files:**
- Create: `infra/docker-compose.yml`, `infra/sonar-project.properties`, `.gitignore`, `.env.example`, `README.md`
- Create: `ALPHA-DUO-Vault/00-Karar-Gunlugu.md`

**Interfaces:**
- Produces: `docker compose up` ile MySQL (port 3306, db `avatar`), SonarQube (9000) ayağa kalkar. Servis adları: `mysql`, `sonarqube`.

- [ ] **Step 1:** `.gitignore` yaz (`.env`, `node_modules/`, `target/`, `build/`, `dist/`, `.superpowers/`, `coverage/`).
- [ ] **Step 2:** `.env.example` yaz: `MYSQL_ROOT_PASSWORD`, `MYSQL_DATABASE=avatar`, `MYSQL_USER`, `MYSQL_PASSWORD`, `SONARQUBE_VERSION`, `SONAR_TOKEN`.
- [ ] **Step 3:** `docker-compose.yml` yaz: `mysql:8` (volume, healthcheck), `sonarqube:lts-community`.
- [ ] **Step 4:** `infra/sonar-project.properties` yaz (`sonar.projectKey=avatar-alpha-duo`, dil ayarları).
- [ ] **Step 5:** `README.md` iskeleti (başlık, hedef, hızlı başlangıç taslağı).
- [ ] **Step 6:** `ALPHA-DUO-Vault/00-Karar-Gunlugu.md` yaz (bugüne kadarki kararlar: yaklaşım 1, süre, fiyat kataloğu kararı, branch/commit kuralları, spec linki).
- [ ] **Step 7:** Commit: `chore: scaffold repo with docker compose and sonarqube`.

---

## Task 2: MySQL şeması + seed verisi

**Files:**
- Create: `migrations/0001_init.up.sql`, `migrations/0001_seed.up.sql`
- Create: `docs/data-model.md`

**Interfaces:**
- Produces: tablo ve sütun adları (aşağıda); tüm sonraki görevler bunları kullanır.

**Şema (sütunlar — karar):**
```sql
materials(id BIGINT PK AI, code VARCHAR UNIQUE, name VARCHAR, unit VARCHAR,
          is_raw BOOL, default_purchase_price DECIMAL(12,2))
products(id BIGINT PK AI, code VARCHAR UNIQUE, name VARCHAR, unit VARCHAR,
         list_price DECIMAL(12,2), category VARCHAR, is_finished BOOL)
suppliers(id BIGINT PK AI, code VARCHAR UNIQUE, name VARCHAR, contact VARCHAR)
markets(id BIGINT PK AI, code VARCHAR UNIQUE, name VARCHAR, contact VARCHAR)
lots(id BIGINT PK AI, code VARCHAR UNIQUE, item_type ENUM('material','product'),
     item_id BIGINT, qty DECIMAL(12,3), unit VARCHAR, supplier_id BIGINT NULL,
     purchase_price DECIMAL(12,2), unit_cost DECIMAL(12,2),
     received_at DATETIME, expiry_date DATE NULL, status VARCHAR,
     created_at DATETIME, updated_at DATETIME)
production_orders(id BIGINT PK AI, code VARCHAR UNIQUE, status VARCHAR,
     started_at DATETIME, finished_at DATETIME NULL, notes TEXT,
     service_name VARCHAR, service_unit_cost DECIMAL(12,2), service_qty DECIMAL(12,3))
production_order_inputs(id BIGINT PK AI, production_order_id BIGINT,
     material_lot_id BIGINT, qty DECIMAL(12,3))
production_order_outputs(id BIGINT PK AI, production_order_id BIGINT,
     product_lot_id BIGINT, qty DECIMAL(12,3))
butchery_batches(id BIGINT PK AI, code VARCHAR UNIQUE, animal_type VARCHAR,
     input_weight DECIMAL(12,3), unit_cost DECIMAL(12,2),
     total_cost DECIMAL(12,2), batch_date DATE, status VARCHAR)
butchery_items(id BIGINT PK AI, batch_id BIGINT, product_id BIGINT,
     product_lot_id BIGINT, weight DECIMAL(12,3), is_waste BOOL)
lot_links(id BIGINT PK AI, parent_lot_id BIGINT, child_lot_id BIGINT,
     qty DECIMAL(12,3), link_type VARCHAR)
stock_movements(id BIGINT PK AI, lot_id BIGINT, movement_type VARCHAR,
     qty DECIMAL(12,3), unit_cost DECIMAL(12,2), total_cost DECIMAL(12,2),
     ref_type VARCHAR, ref_id BIGINT, occurred_at DATETIME, created_at DATETIME)
sales(id BIGINT PK AI, code VARCHAR UNIQUE, market_id BIGINT, sold_at DATETIME,
     status VARCHAR, client_id VARCHAR NULL)
sale_lines(id BIGINT PK AI, sale_id BIGINT, lot_id BIGINT, qty DECIMAL(12,3),
     unit_price DECIMAL(12,2), total_price DECIMAL(12,2), stock_movement_id BIGINT)
recall_events(id BIGINT PK AI, code VARCHAR UNIQUE, reason TEXT,
     source_lot_id BIGINT, created_at DATETIME)
recall_blocks(id BIGINT PK AI, recall_event_id BIGINT, lot_id BIGINT,
     blocked_at DATETIME)
sync_outbox(id BIGINT PK AI, client_id VARCHAR, entity VARCHAR, entity_id BIGINT,
     operation VARCHAR, payload JSON, created_at DATETIME, synced_at DATETIME NULL)
```

**Seed (karar — PDF ile birebir):** fiyat kataloğu (spec §7) + senaryo: `MAT-001` et 9,70 ₺, `MAT-002` baharat 24,00 ₺, `MAT-003` bağırsak 11,60 ₺, `MAT-004` tuz 3,50 ₺, `SVC-001` hizmet 5,50 ₺, `PRD-001` sucuk VKP 24,90 ₺, `ANM-001` kuzu 11,50 ₺, parça ürünler PRD-101..107; tedarikçiler, marketler A/B/C (iletişim bilgisiyle); `ET-2026-001` (100 kg), `URT-2026-045`, `SUC-2026-045` (85 paket) ve market sevkiyatları 30/25/30.

- [ ] **Step 1:** `0001_init.up.sql` yaz (yukarıdaki şema, FK + index: `lot_links(parent_lot_id)`, `lot_links(child_lot_id)`, `stock_movements(lot_id)`, `sale_lines(lot_id)`).
- [ ] **Step 2:** `0001_seed.up.sql` yaz (katalog + senaryo + sevkiyat `stock_movements` kayıtları).
- [ ] **Step 3:** `docs/data-model.md` yaz (tablolar + ilişkiler + örnek zincir şeması).
- [ ] **Step 4:** Migration'ları `docker compose exec mysql mysql ... < migrations/*.sql` ile uygula; `SHOW TABLES` ile doğrula.
- [ ] **Step 5:** Commit: `feat: add mysql schema and scenario seed data`.

---

## Task 3: Go backend iskeleti + temel CRUD + stok

**Files:**
- Create: `backend/go.mod`, `backend/cmd/server/main.go`, `backend/internal/config/config.go`, `backend/internal/db/db.go`, `backend/internal/model/model.go`, `backend/internal/repo/repo.go`, `backend/internal/httpapi/server.go`, `backend/internal/httpapi/dto.go`
- Create: `backend/Dockerfile`
- Modify: `infra/docker-compose.yml` (mysql init-script mount + backend servisi)
- Test: `backend/internal/httpapi/server_test.go`

**Interfaces:**
- Produces: `repo.Repo` arayüzü (sonraki görevler tüketir): `CreateMaterial`, `ListMaterials`, `CreateProduct`, `ListProducts`, `CreateLot`, `GetLotByCode`, `AddStockMovement`, `ListStockMovements(lotID)`.
- Produces: `service.NewCostEngine(repo)` (Task 4), `service.NewTraceability(repo)` (Task 5).

- [ ] **Step 1:** `go.mod` (`module github.com/alpha-duo/avatar/backend`), bağımlılık yok (stdlib).
- [ ] **Step 2:** `internal/config/config.go`: `Config{DSN string, Addr string}`; `Load() (*Config, error)` env'den okur.
- [ ] **Step 3:** `internal/db/db.go`: `Open(dsn string) (*sql.DB, error)` (ping + pool ayarları). Migration'lar MySQL init script olarak uygulanır (`infra/docker-compose.yml` içinde `backend/migrations` mount); backend migration çalıştırmaz.
- [ ] **Step 4:** `internal/model/model.go`: şema struct'ları (JSON tag'leriyle).
- [ ] **Step 5:** `internal/repo/repo.go`: yukarıdaki `Repo` arayüzü + MySQL implementasyonu (parametreli sorgular).
- [ ] **Step 6:** `internal/httpapi/server.go`: router (`GET /health` → `{"status":"ok"}`; materials/products/lots CRUD; `POST /stock-movements`).
- [ ] **Step 7:** Test yaz: `TestHealth` (200 + body), `TestCreateMaterialList` (oluştur→listele→ad eşleşir).
- [ ] **Step 8:** Testleri çalıştır (`go test ./...`), geçtiğini doğrula.
- [ ] **Step 9:** `docs/api.md` başlat (endpoint listesi).
- [ ] **Step 10:** Commit: `feat: add go backend skeleton with materials, products, lots, stock movements`.

---

## Task 4: Maliyet motoru (Görev 1 — EKP/VKP/brüt kâr/marj)

**Files:**
- Create: `internal/service/cost.go`
- Test: `backend/internal/service/cost_test.go`
- Modify: `internal/httpapi/server.go`, `internal/repo/repo.go` (production order oluşturma)

**Interfaces:**
- Produces: `func (e *CostEngine) ProductionOrderCost(poID int64) (CostSummary, error)`
  `type CostSummary struct { InputCost, ServiceCost, TotalCost, FinishedQty, EKP, VKP, GrossProfit, Margin float64 }`

- [ ] **Step 1:** Başarısız test yaz `cost_test.go::TestEKPSucukScenario`:
  `TotalCost == 1543.50`, `EKP == 18.16` (2 ondalık), `VKP == 24.90`, `GrossProfit == 6.74`, `Margin ≈ 27.07`.
- [ ] **Step 2:** Testi çalıştır, FAIL olduğunu doğrula.
- [ ] **Step 3:** `cost.go` içinde saf hesaplama fonksiyonlarını uygula: `ekp(inputCost, serviceCost, finishedQty)`, `grossProfit(vkp, ekp)`, `margin(vkp, ekp)`; `CostEngine.ProductionOrderCost` repo'dan girdi lot maliyeti + hizmet + çıktı miktarını toplayıp doldurur (ağırlıklı ortalama).
- [ ] **Step 4:** Testi çalıştır, PASS olduğunu doğrula.
- [ ] **Step 5:** `POST /production-orders` (girdi lotları + hizmet + çıktı) ve `GET /production-orders/{id}` (CostSummary döner) handler'larını ekle.
- [ ] **Step 6:** Uçtan uca test: `TestProductionOrderCostEndpoint` (seed senaryosu → 18,16 ₺).
- [ ] **Step 7:** Commit: `feat: add production order cost engine with EKP/VKP/margin`.

---

## Task 5: İzlenebilirlik + geri çağırma (Görev 3 — demo odağı)

**Files:**
- Create: `internal/service/traceability.go`, `internal/service/recall.go`
- Test: `backend/internal/service/traceability_test.go`, `backend/internal/service/recall_test.go`
- Modify: `internal/httpapi/server.go`, `internal/repo/repo.go`

**Interfaces:**
- Produces:
  `func (t *Traceability) Forward(sourceLotCode string) ([]AffectedLot, error)` (etkilenen mamul lotları + marketler + sevkiyatlar)
  `func (t *Traceability) Backward(productLotCode string) ([]RawMaterialSource, error)`
  `func (r *Recall) Create(sourceLotCode, reason string) (*RecallResult, error)`
  `func (r *Recall) IsLotBlocked(lotCode string) (bool, error)`
  `func (s *SaleService) CreateSale(sale Sale, lines []SaleLine) (*Sale, error)` — her satırda `IsLotBlocked` kontrolü, bloke ise `ErrLotBlocked` (HTTP 409).
  Endpoint'ler: `POST /sales`, `POST /sync/push` (outbox ingest), `GET /sync/pull` (referans ürünler + bloke lotlar).

- [ ] **Step 1:** Başarısız test `traceability_test.go::TestForwardET2026001`: `ET-2026-001` için sonuç `SUC-2026-045` lotunu ve Market A/B/C'yi (30/25/30) içerir.
- [ ] **Step 2:** Başarısız test `traceability_test.go::TestBackwardSUC2026045`: `SUC-2026-045` → `ET-2026-001` kaynağına ulaşır.
- [ ] **Step 3:** Testleri çalıştır, FAIL.
- [ ] **Step 4:** `traceability.go` uygula: `lot_links` + `production_order_inputs/outputs` + `stock_movements` üzerinden BFS; repo arayüzü üzerinde (test için in-memory fake).
- [ ] **Step 5:** Testleri çalıştır, PASS.
- [ ] **Step 6:** Başarısız test `recall_test.go::TestRecallBlocksSale`: geri çağırma oluştur → `IsLotBlocked(SUC-2026-045) == true`.
- [ ] **Step 7:** `recall.go` uygula (bloke lotlar + `RecallResult{Markets []MarketImpact}`).
- [ ] **Step 8:** Handler'lar: `GET /trace/forward?lot=`, `GET /trace/backward?lot=`, `POST /recalls`, `GET /recalls/{id}`.
- [ ] **Step 9:** Başarısız test `recall_test.go::TestSaleOfBlockedLotRejected`: `SUC-2026-045` bloklanmışken `POST /sales` (o lotlu satır) → HTTP 409 + `ErrLotBlocked`.
- [ ] **Step 10:** `SaleService.CreateSale` + `POST /sales` handler'ı (sunucu-otoriter bloke doğrulaması) uygula; `POST /sync/push` (outbox'tan satış ingest, aynı doğrulama) + `GET /sync/pull` (ürün + bloke lotlar) handler'larını ekle.
- [ ] **Step 11:** Testleri çalıştır, PASS.
- [ ] **Step 12:** Commit: `feat: add lot traceability and recall with forward/backward queries`.

---

## Task 6: Parçalama + maliyet dağıtımı (Görev 2)

**Files:**
- Create: `internal/service/distribution.go`
- Test: `backend/internal/service/distribution_test.go`
- Modify: `internal/httpapi/server.go`, `internal/repo/repo.go`

**Interfaces:**
- Produces:
  `func DistributeByWeight(totalCost float64, items []DistItem) ([]DistItem, error)` — `DistItem{ProductID int64, Weight float64, AllocatedCost float64}`; garantisi: `sum(AllocatedCost) == totalCost` (2 ondalık, son kaleme kalan).
  `func (s *ButcheryService) CreateBatch(batch ButcheryBatch, items []ButcheryItemInput) (*ButcheryBatch, error)` — ağırlık dengesi kontrolü, fire kaydı.

- [ ] **Step 1:** Başarısız test `TestDistributeSumEqualsTotal`: 287,50 ₺ toplam → dağıtılan toplam == 287,50 (tam eşitlik).
- [ ] **Step 2:** Başarısız test `TestWeightBalanceFlag`: 25 kg girdi, parçalar 22 kg → `weightBalanceOK == false`, fire 3 kg.
- [ ] **Step 3:** Başarısız test `TestCostMismatchFlag`: 25 kg × 11,50 = 287,50 ≠ 351,00 → `costMismatch == true`.
- [ ] **Step 4:** Testleri çalıştır, FAIL.
- [ ] **Step 5:** `distribution.go` uygula (miktar/değer ağırlıklı anahtar; tutarsızlık tespiti).
- [ ] **Step 6:** Testleri çalıştır, PASS.
- [ ] **Step 7:** Handler'lar: `POST /butchery-batches`, `GET /butchery-batches/{id}` (dengesizlik + maliyet uyuşmazlığı uyarıları döner).
- [ ] **Step 8:** Commit: `feat: add butchery with cost distribution and inconsistency detection`.

---

## Task 7: ML servisi (iki zorunlu KI)

**Files:**
- Create: `ml/app.py`, `ml/model.py`, `ml/train.py`, `ml/requirements.txt`
- Test: `ml/test_model.py`

**Interfaces:**
- Produces: `POST /predict/yield` (girdi: `{production_order_id}` veya feature'lar → `{predicted_qty, predicted_unit_cost}`); `POST /predict/butchery` (`{animal_type, input_weight}` → `{parts: [{product_code, weight}]}`); `GET /report` (baseline karşılaştırma + MAE/RMSE).

- [ ] **Step 1:** `requirements.txt` (`fastapi`, `uvicorn`, `scikit-learn`, `numpy`, `pandas`, `pytest`).
- [ ] **Step 2:** `model.py`: `train_yield(df)` ve `train_butchery(df)` → lineer regresyon; `baseline` = tarihsel ortalama; `evaluate(y_true, y_pred)` → MAE/RMSE.
- [ ] **Step 3:** `train.py`: örnek/anonymize veriyle eğitir, metrikleri `report.json`'a yazar (veri azlığı notu ekler).
- [ ] **Step 4:** `app.py`: FastAPI endpoint'leri (yukarıdaki arayüz) + `/report`.
- [ ] **Step 5:** Test `test_model.py::TestBaselineBeatsOrMatches` (MAE ≤ baseline MAE) ve `TestPredictShapes`.
- [ ] **Step 6:** `pytest` çalıştır, PASS.
- [ ] **Step 7:** Commit: `feat: add ml service with yield and butchery prediction models`.

---

## Task 8: React web (yönetim + raporlama + geri çağırma)

**Files:**
- Create: `web/` (Vite + React + TS): `src/api.ts`, `src/App.tsx`, `src/pages/{Dashboard,Materials,Products,ProductionOrders,Butchery,Traceability,Recall}.tsx`, `src/components/*`

**Interfaces:**
- Consumes: Go API endpoint'leri (Task 3-6) ve ML `/report` (Task 7).

- [ ] **Step 1:** Vite + React + TS iskeleti; `api.ts` (fetch wrapper, base URL env).
- [ ] **Step 2:** `ProductionOrders.tsx` — üretim emri oluştur + EKP/VKP/brüt kâr/marj tablosu.
- [ ] **Step 3:** `Traceability.tsx` — ileri/geri izleme (lot kodu gir → etkilenen lotlar/marketler).
- [ ] **Step 4:** `Recall.tsx` — geri çağırma oluştur + market bazlı miktar/sevkiyat/iletişim listesi + üretilen mesaj.
- [ ] **Step 5:** `Butchery.tsx` — parçalama + dağıtım + dengesizlik/uyuşmazlık uyarıları.
- [ ] **Step 6:** `Dashboard.tsx` + `Materials.tsx` + `Products.tsx` (katalog görüntüleme).
- [ ] **Step 7:** `npm run build` ile derle, hatasız olduğunu doğrula.
- [ ] **Step 8:** Commit: `feat: add react web admin, reports, and recall screens`.

---

## Task 9: Kotlin Android (satış + offline + senkronizasyon)

**Files:**
- Create: `mobile/` (Compose + Room): `data/AppDatabase.kt`, `data/entities/*.kt`, `data/SyncRepository.kt`, `ui/screens/{ProductList,Sale,History}.kt`

**Interfaces:**
- Consumes: Go API `POST /sales` (bloke doğrulaması), `POST /sync/push`, `GET /sync/pull`, `GET /lots` (bloke lot listesi).

- [ ] **Step 1:** Room şeması: `LocalProduct`, `LocalLot`, `LocalSale`, `LocalSaleLine`, `Outbox`.
- [ ] **Step 2:** `ProductList.kt` + `Sale.kt` — lot seçerek satış, bloke lot devre dışı.
- [ ] **Step 3:** `SyncRepository.kt` — outbox'a yaz, push (satışları sunucuya gönder), pull (ürün + bloke lotlar); LWW (`updated_at`).
- [ ] **Step 4:** Test: `SyncRepositoryTest::testLastWriteWins` — aynı satır iki cihaz `updated_at` farklı → son güncelleyen kazanır.
- [ ] **Step 5:** Bloke lot çevrimdışı satılırsa sunucu reddi (HTTP 409) UI'da göster.
- [ ] **Step 6:** Derle (`./gradlew assembleDebug`), APK üret.
- [ ] **Step 7:** Commit: `feat: add android sales app with offline sqlite and sync`.

---

## Task 10: Stretch — OCR + anomali uyarıları

**Files:**
- Create: `ml/ocr.py`, `backend/internal/service/anomaly.go` (zaman kalırsa)

- [ ] **Step 1:** Zaman kısıtı değerlendir; kalırsa `ocr.py` (fatura/irsaliye → ürün/miktar/lot/maliyet önerisi, insan onayıyla kayıt) ve `anomaly.go` (beklenmedik fire, maliyet artışı, düşük verim, tutarsız lot hareketi).
- [ ] **Step 2:** Yoksa dokümanda "opsiyonel, kapsam dışı" notu bırak.
- [ ] **Step 3:** Commit: `feat: add optional ocr and anomaly alerts` veya `docs: mark optional features as out of scope`.

---

## Task 11: Kapanış — testler, SonarQube, dokümantasyon, demo

**Files:**
- Modify: `docs/api.md`, `docs/data-model.md`, `README.md`, `ALPHA-DUO-Vault/*`
- Test: tam suite

- [ ] **Step 1:** `go test ./...` (backend) + `pytest` (ml) + `npm run build` (web) + `./gradlew assembleDebug` (mobile) hepsi yeşil.
- [ ] **Step 2:** `docker compose up` ile tam sistemi ayağa kaldır; uçtan uca demo akışını elle doğrula (ET-2026-001 → marketler).
- [ ] **Step 3:** SonarQube taraması çalıştır; critical/high issue'ları sıfırla (bugs, code smells, security).
- [ ] **Step 4:** `README.md` kurulum + çalıştırma talimatını tamamla.
- [ ] **Step 5:** `docs/api.md`, `docs/data-model.md` tamamla; KI açıklaması + doğruluk raporu ekle.
- [ ] **Step 6:** Vault'a ilerleme + demo scripti notu.
- [ ] **Step 7:** Demo scripti (`docs/demo-script.md`): adım adım senaryo (maliyet → parçalama → izlenebilirlik → geri çağırma → KI).
- [ ] **Step 8:** Commit: `docs: finalize setup, api, demo script`.
