# JUNIOR GELİŞTİRİCİ TALİMATI + GÖREV 5

## ROL (önce bunu oku)
Sen bu projede çalışan **JUNIOR** bir yazılım geliştiricisin. Bir senior geliştirici
sana görev verdi. Sen kodu YAZARSIN; senior inceler, düzeltir, Docker'da test eder ve
**commit'i senior atar** (sen commit/push YAPMAZ).

Kuralların (kesin uy):
1. **SADECE görevde isteneni yap.** Fazladan özellik, "iyileştirme", ek dosya/endpoint EKLEME.
2. Verilen isim, imza, tablo/sütun adı ve değerleri **BİREBİR** kullan.
3. Tüm SQL **parametreli** (`?`). Asla string birleştirme.
4. İsimlendirme İngilizce, açıklayıcı, mevcut kodla tutarlı. Clean code, LLM-okunur.
5. Verilen testleri + iddiaları tam yaz.
6. Derleme çalıştıramasan bile kod derlenecek şekilde yaz.
7. **Günlük notu:** her gün yaptıklarını **kendi klasörün olan** `ALPHA-DUO-Vault/Meryem/YYYY-MM-DD.md`
   dosyasına yazabilirsin (o klasörde değişiklik yapma yetkin var; `ilayda/` klasörü
   senior'a aittir, oraya dokunma). Günlük **gerçeği** yaz; kendin çalıştırmadığın
   bir testi "geçti" diye yazma.
8. Teslimat: değiştirdiğin/eklediğin dosyaların **TAM içeriği**, dosya yollarıyla.

## BAĞLAM (mevcut kod)
- Go modülü `avatar/backend` (go 1.23, sadece `github.com/go-sql-driver/mysql`).
- `internal/model/model.go` — Material, Product, Lot, StockMovement, ProductionOrder*.
- `internal/repo/repo.go` — `Repo` arayüzü + `mysqlRepo`.
- `internal/service/cost.go` — `CostEngine` (örnek: `NewCostEngine(r repo.Repo)`).
- `internal/httpapi/server.go` — `Server` + `NewServer(r repo.Repo)`.
- `internal/httpapi/server_test.go` — `fakeRepo` (bellek içi Repo, tüm metotları stub).
- MySQL şeması + seed hazır: `lot_links(parent_lot_id, child_lot_id, qty, link_type)`,
  `stock_movements(lot_id, movement_type, qty, ref_type, ref_id)`,
  `markets(id, code, name, contact)`, `suppliers(id, code, name, contact)`.
  Seed: ET-2026-001(100kg) → BAH/BAR → SUC-2026-045(85 paket); sevk MKT-A 30, MKT-B 25, MKT-C 30.

## GÖREV: Lot izlenebilirliği (ileri/geri izleme)

### 1) `internal/model/model.go` — 3 struct ekle
```go
type LotLink struct {
    ID          int64   `json:"id"`
    ParentLotID int64   `json:"parent_lot_id"`
    ChildLotID  int64   `json:"child_lot_id"`
    Qty         float64 `json:"qty"`
    LinkType    string  `json:"link_type"`
}
type Market struct {
    ID      int64  `json:"id"`
    Code    string `json:"code"`
    Name    string `json:"name"`
    Contact string `json:"contact"`
}
type Supplier struct {
    ID      int64  `json:"id"`
    Code    string `json:"code"`
    Name    string `json:"name"`
    Contact string `json:"contact"`
}
```

### 2) `internal/repo/repo.go` — `Repo` arayüzüne 4 metot ekle + `mysqlRepo`'da implement et
```go
ListLotLinksByParent(ctx context.Context, parentID int64) ([]model.LotLink, error)
ListLotLinksByChild(ctx context.Context, childID int64) ([]model.LotLink, error)
GetMarket(ctx context.Context, id int64) (model.Market, error)
GetSupplier(ctx context.Context, id int64) (model.Supplier, error)
```
SQL:
- `ListLotLinksByParent`: `SELECT id, parent_lot_id, child_lot_id, qty, link_type FROM lot_links WHERE parent_lot_id = ? ORDER BY id`
- `ListLotLinksByChild`: aynı, `WHERE child_lot_id = ?`
- `GetMarket`: `SELECT id, code, name, contact FROM markets WHERE id = ?`
- `GetSupplier`: `SELECT id, code, name, contact FROM suppliers WHERE id = ?`

### 3) `internal/service/traceability.go` — yeni dosya
```go
package service

type MarketImpact struct {
    MarketCode string  `json:"market_code"`
    MarketName string  `json:"market_name"`
    Contact    string  `json:"contact"`
    Qty        float64 `json:"qty"`
}
type AffectedLot struct {
    LotCode  string         `json:"lot_code"`
    ItemType string         `json:"item_type"`
    Qty      float64        `json:"qty"`
    Unit     string         `json:"unit"`
    Markets  []MarketImpact `json:"markets"`
}
type RawMaterialSource struct {
    LotCode      string  `json:"lot_code"`
    SupplierCode string  `json:"supplier_code"`
    SupplierName string  `json:"supplier_name"`
    Qty          float64 `json:"qty"`
}

type TraceRepository interface {
    GetLotByCode(ctx context.Context, code string) (model.Lot, error)
    GetLot(ctx context.Context, id int64) (model.Lot, error)
    ListLotLinksByParent(ctx context.Context, parentID int64) ([]model.LotLink, error)
    ListLotLinksByChild(ctx context.Context, childID int64) ([]model.LotLink, error)
    ListStockMovements(ctx context.Context, lotID int64) ([]model.StockMovement, error)
    GetMarket(ctx context.Context, id int64) (model.Market, error)
    GetSupplier(ctx context.Context, id int64) (model.Supplier, error)
}

type Traceability struct { repo TraceRepository }
func NewTraceability(r TraceRepository) *Traceability { return &Traceability{repo: r} }

func (t *Traceability) Forward(ctx context.Context, sourceLotCode string) ([]AffectedLot, error)
func (t *Traceability) Backward(ctx context.Context, productLotCode string) ([]RawMaterialSource, error)
```
Algoritmalar:
- **Forward:** `start := GetLotByCode(sourceLotCode)`; BFS ile `ListLotLinksByParent` üzerinden
  ulaşılabilir tüm lot ID'lerini topla (visited set ile). Ulaşılan **product** lotlar için
  `ListStockMovements(lotID)`; `movement_type == "shipment" && ref_type == "market"` olanları
  `GetMarket(ref_id)` ile eşleştir, `MarketImpact{Qty: abs(movement.Qty)}` ekle. Material lotlar
  ara düğümdür, sonuca girmez.
- **Backward:** `start := GetLotByCode(productLotCode)`; BFS ile `ListLotLinksByChild` üzerinden
  geriye git; ulaşılan **material** lotlar için `GetSupplier(lot.SupplierID)` (nil değilse)
  ile `RawMaterialSource` üret.

### 4) `internal/httpapi/server.go` — 2 endpoint + trace alanı
- `Server` struct'a `trace *service.Traceability` ekle; `NewServer` içinde
  `trace: service.NewTraceability(r)` ata.
- Route ekle: `GET /trace/forward` (`?lot=CODE`), `GET /trace/backward` (`?lot=CODE`).
- Handler: `r.URL.Query().Get("lot")` boşsa 400; değilse `trace.Forward/Backward` → 200.

### 5) `internal/httpapi/server_test.go` — `fakeRepo`'ya 4 stub ekle
Yeni `Repo` metotları için sıfır-değer döndüren stub'lar (mevcut testler bozulmasın).

### 6) `internal/service/traceability_test.go` — yeni test (küçük fake TraceRepository)
```go
func TestForwardET2026001(t *testing.T) {
    // fake: lots ET-2026-001(material), SUC-2026-045(product, qty 85, unit "paket");
    // lot_link parent=ET child=SUC; SUC sevk: market 1 -> -30, 2 -> -25, 3 -> -30;
    // markets 1=MKT-A, 2=MKT-B, 3=MKT-C.
    res := NewTraceability(fake).Forward(ctx, "ET-2026-001")
    // 1 AffectedLot (SUC-2026-045), 3 MarketImpact: A 30, B 25, C 30
}
func TestBackwardSUC2026045(t *testing.T) {
    // aynı fake + lot_links BAH/BAR -> SUC; suppliers.
    res := NewTraceability(fake).Backward(ctx, "SUC-2026-045")
    // 3 RawMaterialSource (ET, BAH, BAR)
}
```

## KABUL (senior kontrol edecek)
- `go build ./...` + `go vet ./...` + `go test ./...` geçer.
- `GET /trace/forward?lot=ET-2026-001` → SUC-2026-045 + Market A(30) B(25) C(30).
- `GET /trace/backward?lot=SUC-2026-045` → ET-2026-001, BAH-2026-001, BAR-2026-001 + tedarikçiler.
- Parametreli SQL, İngilizce isimlendirme, clean code.
