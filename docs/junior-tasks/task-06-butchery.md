# JUNIOR (Meryem) GELİŞTİRİCİ TALİMATI + GÖREV 6

## ROL (önce bunu oku)
Sen **Meryem takımının JUNIOR** yazılım geliştiricisisin. Meryem takımının senior'ı
sana görev verdi; ilayda takımının senior'ı da işini inceler, Docker'da test eder ve
merge eder. Sen kodu YAZARSIN; **commit/push'u sen yapmazsın** — ama değişiklikleri
bir branch'e koyup **pull request** açarsın (senior PR'ı inceler, test eder, merge eder).

Kuralların (kesin uy):
1. **SADECE görevde isteneni yap.** Fazladan özellik, "iyileştirme", ek dosya/endpoint EKLEME.
2. Verilen isim, imza, tablo/sütun adı ve değerleri **BİREBİR** kullan.
3. Tüm SQL **parametreli** (`?`). Asla string birleştirme.
4. İsimlendirme İngilizce, açıklayıcı, mevcut kodla tutarlı. Clean code, LLM-okunur.
5. Verilen testleri + iddiaları tam yaz.
6. Derleme çalıştıramasan bile kod derlenecek şekilde yaz.
7. **Günlük notu:** yaptıklarını **kendi klasörün olan** `ALPHA-DUO-Vault/Meryem/YYYY-MM-DD.md`
   dosyasına yazabilirsin (`ilayda/` klasörü ilayda takımına aittir, oraya dokunma).
   Günlük **gerçeği** yaz; kendin çalıştırmadığın testi "geçti" diye yazma.
8. Teslimat: `feat/task-06-butchery` branch'i + PR. Değiştirdiğin/eklediğin dosyaların TAM içeriği.

## BAĞLAM (mevcut kod)
- Go modülü `avatar/backend` (go 1.23, sadece `github.com/go-sql-driver/mysql`).
- `internal/model/model.go` — Material, Product, Lot, StockMovement, ProductionOrder*, LotLink, Market, Supplier.
- `internal/repo/repo.go` — `Repo` arayüzü + `mysqlRepo`.
- `internal/service/cost.go` — `CostEngine`; `internal/service/traceability.go` — `Traceability` (örnek yapı).
- `internal/httpapi/server.go` — `Server` + `NewServer(r repo.Repo)`.
- `internal/httpapi/server_test.go` — `fakeRepo` (bellek içi Repo).
- MySQL şeması + seed hazır: `butchery_batches(id, code, animal_type, input_weight,
  unit_cost, total_cost, batch_date, status)`, `butchery_items(id, batch_id, product_id,
  product_lot_id, weight, is_waste)`.

## GÖREV: Parçalama + maliyet dağıtımı (Görev 2)

### 1) `internal/model/model.go` — 2 struct ekle
```go
type ButcheryBatch struct {
    ID          int64   `json:"id"`
    Code        string  `json:"code"`
    AnimalType  string  `json:"animal_type"`
    InputWeight float64 `json:"input_weight"`
    UnitCost    float64 `json:"unit_cost"`
    TotalCost   float64 `json:"total_cost"`
    BatchDate   string  `json:"batch_date"`
    Status      string  `json:"status"`
}
type ButcheryItem struct {
    ID           int64   `json:"id"`
    BatchID      int64   `json:"batch_id"`
    ProductID    int64   `json:"product_id"`
    ProductLotID *int64  `json:"product_lot_id"`
    Weight       float64 `json:"weight"`
    IsWaste      bool    `json:"is_waste"`
}
```

### 2) `internal/repo/repo.go` — `Repo` arayüzüne 3 metot ekle + `mysqlRepo`'da implement et
```go
CreateButcheryBatch(ctx context.Context, batch model.ButcheryBatch, items []model.ButcheryItem) (model.ButcheryBatch, error)
GetButcheryBatch(ctx context.Context, id int64) (model.ButcheryBatch, error)
ListButcheryItems(ctx context.Context, batchID int64) ([]model.ButcheryItem, error)
```
- `CreateButcheryBatch`: transaction (`BeginTx` + `defer Rollback` + `Commit`);
  `INSERT INTO butchery_batches (code, animal_type, input_weight, unit_cost, total_cost,
  batch_date, status) VALUES (?,?,?,?,?,?,?)`; sonra her item için
  `INSERT INTO butchery_items (batch_id, product_id, product_lot_id, weight, is_waste)
  VALUES (?,?,?,?,?)`.
- `GetButcheryBatch`: `SELECT id, code, animal_type, input_weight, unit_cost, total_cost,
  batch_date, status FROM butchery_batches WHERE id = ?`
- `ListButcheryItems`: `SELECT id, batch_id, product_id, product_lot_id, weight, is_waste
  FROM butchery_items WHERE batch_id = ? ORDER BY id` (nullable `product_lot_id` → `sql.NullInt64`).

### 3) `internal/service/distribution.go` — yeni dosya (saf fonksiyonlar)
```go
package service

type DistItem struct {
    ProductID     int64
    Weight        float64
    AllocatedCost float64
}

// DistributeByWeight: totalCost'u Weight oranında dağıtır; her kalem round2,
// SON kaleme kalan eklenir → sum(AllocatedCost) == round2(totalCost) garantili.
func DistributeByWeight(totalCost float64, items []DistItem) []DistItem

// WeightBalanceOK: parça ağırlıklarının toplamı inputWeight'e eşit mi?
// ok=false ise fire = inputWeight - sum(weights).
func WeightBalanceOK(inputWeight float64, weights []float64) (ok bool, fire float64)

func round2(x float64) float64  // math.Round(x*100)/100 (cost.go'da zaten var, yeniden tanımlama)
```
> Not: `round2` zaten `cost.go` içinde tanımlı; aynı pakette olduğu için tekrar
> tanımlama, doğrudan kullan.

### 4) `internal/service/butchery.go` — yeni dosya (orchestration)
```go
type ButcheryItemCost struct {
    ProductID     int64   `json:"product_id"`
    Weight        float64 `json:"weight"`
    IsWaste       bool    `json:"is_waste"`
    AllocatedCost float64 `json:"allocated_cost"`
}
type ButcheryResult struct {
    Batch           model.ButcheryBatch `json:"batch"`
    Items           []ButcheryItemCost  `json:"items"`
    WeightBalanceOK bool                `json:"weight_balance_ok"`
    FireWeight      float64             `json:"fire_weight"`
    CostMismatch    bool                `json:"cost_mismatch"`
    ExpectedCost    float64             `json:"expected_cost"`
}
type ButcheryRepository interface {
    CreateButcheryBatch(ctx, model.ButcheryBatch, []model.ButcheryItem) (model.ButcheryBatch, error)
    GetButcheryBatch(ctx, id) (model.ButcheryBatch, error)
    ListButcheryItems(ctx, batchID) ([]model.ButcheryItem, error)
}
type ButcheryService struct { repo ButcheryRepository }
func NewButcheryService(r ButcheryRepository) *ButcheryService

func (s *ButcheryService) CreateBatch(ctx, batch model.ButcheryBatch, items []model.ButcheryItem) (*ButcheryResult, error)
func (s *ButcheryService) GetBatch(ctx, id int64) (*ButcheryResult, error)
```
`buildResult(batch, items)` (CreateBatch ve GetBatch ortak) şunu yapar:
1. parça ağırlıkları = `is_waste == false` item'ların `Weight` listesi.
2. `weightBalanceOK, fire := WeightBalanceOK(batch.InputWeight, partWeights)`.
3. `expectedCost := round2(batch.InputWeight * batch.UnitCost)`; `costMismatch := round2(expectedCost) != round2(batch.TotalCost)`.
4. `distItems := DistributeByWeight(batch.TotalCost, [...])` (is_waste false item'lar; ProductID + Weight ile);
   her item'e `AllocatedCost` ata → `[]ButcheryItemCost`.
5. `ButcheryResult` doldur.

`CreateBatch`: `repo.CreateButcheryBatch` çağır → `buildResult`.

### 5) `internal/httpapi/server.go` — 2 endpoint + butchery alanı
- `Server` struct'a `butchery *service.ButcheryService` ekle; `NewServer` içinde
  `butchery: service.NewButcheryService(r)` ata.
- Route ekle: `POST /butchery-batches`, `GET /butchery-batches/{id}`.
- `POST`: gövde `{batch: {...}, items: [...]}` → `CreateBatch` → 201 `ButcheryResult`.
- `GET`: `strconv.ParseInt(PathValue("id"))` → `GetBatch` → 200; `sql.ErrNoRows` → 404.

### 6) `internal/httpapi/server_test.go` — `fakeRepo`'ya 3 stub ekle
`CreateButcheryBatch`, `GetButcheryBatch`, `ListButcheryItems` için sıfır-değer stub.

### 7) `internal/service/distribution_test.go` — yeni test (saf fonksiyonlar)
```go
func TestDistributeSumEqualsTotal(t *testing.T) {
    // total 287.50, weights [10, 8, 4] → sum(AllocatedCost) == 287.50 (tam eşitlik)
}
func TestWeightBalanceFlag(t *testing.T) {
    // input 25, weights [10, 8, 4] (toplam 22) → ok=false, fire=3.0
}
func TestCostMismatchFlag(t *testing.T) {
    // input 25, unit 11.50 → expected 287.50; total 351.00 → mismatch=true
}
```
`TestCostMismatchFlag` saf `CostMismatch` yerine, `buildResult` mantığını doğrulamak
için bir fake `ButcheryRepository` ile `ButcheryService.GetBatch` üzerinden de test edebilirsin
(küçük fake `ButcheryRepository` yaz).

## KABUL (senior kontrol edecek)
- `go build ./...` + `go vet ./...` + `go test ./...` geçer.
- `DistributeByWeight` toplamı tam eşit; ağırlık dengesi ve maliyet uyuşmazlığı bayrakları doğru.
- Parametreli SQL, İngilizce isimlendirme, clean code.
