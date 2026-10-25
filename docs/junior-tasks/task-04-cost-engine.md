# JUNIOR GELİŞTİRİCİ TALİMATI + GÖREV 4

## ROL (önce bunu oku)
Sen bu projede çalışan **JUNIOR** bir yazılım geliştiricisin. Bir senior geliştirici
sana görev verdi. Sen kodu YAZARSIN; senior inceler, düzeltir, test eder ve commit eder.

Senin kuralların (bunlara kesin uy):
1. **SADECE görevde isteneni yap.** Fazladan özellik, "iyileştirme", yeniden yapılandırma,
   ek dosya, ek endpoint, ek doğrulama EKLEME.
2. Verilen isimleri, imzaları, tablo/sütun adlarını ve değerleri **BİREBİR** kullan.
   Kendi başına isim veya arayüz değiştirme.
3. Tüm SQL **parametreli** (`?`). Asla string birleştirme ile SQL kurma.
4. İsimlendirme **İngilizce**, açıklayıcı, mevcut kodla tutarlı.
5. Kod **clean code**, LLM-okunur: kısa fonksiyonlar, anlaşılır isimler.
6. Testleri **tam** yaz (aşağıda verilen testler + iddialar).
7. Derleme/test çalıştıramasan bile kod **derlenecek şekilde** yaz.
8. Teslimat: dosyaların **TAM içeriği**, dosya yollarıyla birlikte. Açıklama değil, kod.

## BAĞLAM (mevcut kod)
- Go modülü: `avatar/backend` (go 1.23, sadece `github.com/go-sql-driver/mysql`).
- `internal/model/model.go` — mevcut struct'lar: Material, Product, Lot, StockMovement.
- `internal/repo/repo.go` — `Repo` arayüzü + `mysqlRepo` implementasyonu.
- `internal/httpapi/server.go` — `Server` (router + handler'lar), `NewServer(r repo.Repo)`.
- `internal/httpapi/server_test.go` — `fakeRepo` (bellek içi Repo).
- MySQL şeması + seed hazır (`production_orders`, `production_order_inputs`,
  `production_order_outputs` tabloları var; URT-2026-045 id=1 ile seed'li).

## GÖREV: Maliyet motoru (EKP/VKP/brüt kâr/marj)

### 1) `internal/model/model.go` — 3 struct ekle
```go
type ProductionOrder struct {
    ID              int64   `json:"id"`
    Code            string  `json:"code"`
    Status          string  `json:"status"`
    StartedAt       *string `json:"started_at"`
    FinishedAt      *string `json:"finished_at"`
    Notes           *string `json:"notes"`
    ServiceName     *string `json:"service_name"`
    ServiceUnitCost float64 `json:"service_unit_cost"`
    ServiceQty      float64 `json:"service_qty"`
}
type ProductionOrderInput struct {
    ID                int64   `json:"id"`
    ProductionOrderID int64   `json:"production_order_id"`
    MaterialLotID     int64   `json:"material_lot_id"`
    Qty               float64 `json:"qty"`
}
type ProductionOrderOutput struct {
    ID                int64   `json:"id"`
    ProductionOrderID int64   `json:"production_order_id"`
    ProductLotID      int64   `json:"product_lot_id"`
    Qty               float64 `json:"qty"`
}
```

### 2) `internal/repo/repo.go` — Repo arayüzüne 6 metod ekle + implement et
```go
// Repo arayüzüne ekle:
CreateProductionOrder(ctx context.Context, po model.ProductionOrder, inputs []model.ProductionOrderInput, outputs []model.ProductionOrderOutput) (model.ProductionOrder, error)
GetProductionOrder(ctx context.Context, id int64) (model.ProductionOrder, error)
ListProductionOrderInputs(ctx context.Context, poID int64) ([]model.ProductionOrderInput, error)
ListProductionOrderOutputs(ctx context.Context, poID int64) ([]model.ProductionOrderOutput, error)
GetLot(ctx context.Context, id int64) (model.Lot, error)
GetProduct(ctx context.Context, id int64) (model.Product, error)
```

Implementasyon detayları (parametreli SQL):
- `CreateProductionOrder`: `db.BeginTx` ile tek transaction; `INSERT INTO production_orders
  (code, status, started_at, finished_at, notes, service_name, service_unit_cost, service_qty)
  VALUES (?,?,?,?,?,?,?,?)`; sonra `production_order_inputs` ve `production_order_outputs`
  satırlarını `po.ID` ile ekle; `tx.Commit`.
- `GetProductionOrder`: `SELECT id, code, status, started_at, finished_at, notes,
  service_name, service_unit_cost, service_qty FROM production_orders WHERE id = ?`
  (nullable alanlar `sql.NullString` ile taranır, `*string`'e çevrilir).
- `ListProductionOrderInputs`: `SELECT id, production_order_id, material_lot_id, qty
  FROM production_order_inputs WHERE production_order_id = ?`
- `ListProductionOrderOutputs`: aynı kalıp, `production_order_outputs`.
- `GetLot`: mevcut `scanLot` yardımcısını kullan, `WHERE id = ?`.
- `GetProduct`: `SELECT id, code, name, unit, list_price, category, is_finished
  FROM products WHERE id = ?`

### 3) `internal/service/cost.go` — yeni dosya
```go
package service

type CostSummary struct {
    InputCost   float64 `json:"input_cost"`
    ServiceCost float64 `json:"service_cost"`
    TotalCost   float64 `json:"total_cost"`
    FinishedQty float64 `json:"finished_qty"`
    EKP         float64 `json:"ekp"`
    VKP         float64 `json:"vkp"`
    GrossProfit float64 `json:"gross_profit"`
    Margin      float64 `json:"margin"`
}

type CostEngine struct { repo repo.Repo }
func NewCostEngine(r repo.Repo) *CostEngine { return &CostEngine{repo: r} }

func round2(x float64) float64 { return math.Round(x*100) / 100 }

func EKP(inputCost, serviceCost, finishedQty float64) float64 {
    if finishedQty <= 0 { return 0 }
    return round2((inputCost + serviceCost) / finishedQty)
}
func GrossProfit(vkp, ekp float64) float64 { return round2(vkp - ekp) }
func Margin(vkp, ekp float64) float64 {
    if vkp <= 0 { return 0 }
    return round2((vkp - ekp) / vkp * 100)
}
```
`ProductionOrderCost(ctx, poID) (CostSummary, error)` algoritması:
1. `po := GetProductionOrder(poID)`
2. `inputs := ListProductionOrderInputs(poID)`; her input için `lot := GetLot(MaterialLotID)`,
   `inputCost += input.Qty * lot.UnitCost`
3. `serviceCost := po.ServiceUnitCost * po.ServiceQty`; `totalCost := inputCost + serviceCost`
4. `outputs := ListProductionOrderOutputs(poID)`; `finishedQty := sum(outputs.Qty)`
5. `ekp := EKP(inputCost, serviceCost, finishedQty)`
6. `vkp`: `len(outputs) > 0` ise `lot := GetLot(outputs[0].ProductLotID)`,
   `prod := GetProduct(lot.ItemID)`, `vkp = prod.ListPrice`
7. `CostSummary` doldur: alanları `round2` ile, `EKP=ekp`, `VKP=vkp`,
   `GrossProfit=GrossProfit(vkp,ekp)`, `Margin=Margin(vkp,ekp)`.

### 4) `internal/httpapi/server.go` — 2 endpoint + cost alanı
- `Server` struct'a `cost *service.CostEngine` alanı ekle.
- `NewServer(r repo.Repo)` içinde `cost: service.NewCostEngine(r)` ata (imza değişmez).
- Route ekle: `POST /production-orders`, `GET /production-orders/{id}`.
- `POST /production-orders`: gövdeyi decode et →
  `{code, status, started_at, finished_at, notes, service_name, service_unit_cost,
  service_qty, inputs[], outputs[]}` → `CreateProductionOrder` → 201.
- `GET /production-orders/{id}`: `r.PathValue("id")` → `strconv.ParseInt` →
  `cost.ProductionOrderCost(ctx, id)` → 200 `CostSummary`; `sql.ErrNoRows` ise 404.

### 5) `internal/httpapi/server_test.go` — fakeRepo'ya 6 stub ekle
`fakeRepo` artık büyüyen `Repo` arayüzünü uygulamalı. Yeni 6 metod için sıfır-değer
döndüren stub ekle (mevcut testlerin bozulmaması için).

### 6) `internal/service/cost_test.go` — yeni test dosyası (saf fonksiyonlar)
```go
func TestEKPSucukScenario(t *testing.T) {
    if got := EKP(1076.00, 467.50, 85); got != 18.16 {
        t.Fatalf("EKP = %v, want 18.16", got)
    }
}
func TestGrossProfitMargin(t *testing.T) {
    if got := GrossProfit(24.90, 18.16); got != 6.74 { t.Fatalf(...) }
    if got := Margin(24.90, 18.16); got != 27.07 { t.Fatalf(...) }
}
func TestEKPZeroQty(t *testing.T) {
    if got := EKP(100, 50, 0); got != 0 { t.Fatalf(...) }
}
```

## KABUL (senior kontrol edecek)
- `go build ./...` + `go vet ./...` + `go test ./...` geçer.
- `GET /production-orders/1` → `EKP=18.16`, `VKP=24.90`, `GrossProfit=6.74`, `Margin=27.07`,
  `InputCost=1076.00`, `ServiceCost=467.50`, `TotalCost=1543.50`, `FinishedQty=85`.
- Parametreli SQL, İngilizce isimlendirme, clean code.
