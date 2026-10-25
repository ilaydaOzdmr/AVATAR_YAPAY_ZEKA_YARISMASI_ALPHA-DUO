# Görev 3 — Go backend iskeleti + temel CRUD + stok (junior)

## Bağlam
Monorepo. MySQL 8 şeması ve seed hazır (`backend/migrations/*.sql`). Go ile REST API.
Sadece bir dış bağımlılık: `github.com/go-sql-driver/mysql`.

## Teslimat (dosyalar)
- `backend/go.mod` — `module avatar/backend`, go 1.23, mysql driver v1.8.1
- `backend/cmd/server/main.go`
- `backend/internal/config/config.go`
- `backend/internal/db/db.go`
- `backend/internal/model/model.go`
- `backend/internal/repo/repo.go`
- `backend/internal/httpapi/server.go`
- `backend/internal/httpapi/dto.go`
- `backend/internal/httpapi/server_test.go`
- `docs/api.md`

## Config (`config.Load()`)
- `DATABASE_DSN` (default `avatar:avatar@tcp(localhost:3306)/avatar?parseTime=true&charset=utf8mb4`)
- `HTTP_ADDR` (default `:8080`)

## Model (JSON tag'leriyle)
```go
Material{ID int64, Code, Name, Unit string, IsRaw bool, DefaultPurchasePrice float64}
Product{ID int64, Code, Name, Unit string, ListPrice float64, Category string, IsFinished bool}
Lot{ID int64, Code, ItemType string, ItemID int64, Qty float64, Unit string,
    SupplierID *int64, PurchasePrice, UnitCost float64, ReceivedAt *string,
    ExpiryDate *string, Status string, CreatedAt, UpdatedAt string}
StockMovement{ID int64, LotID int64, MovementType string, Qty, UnitCost, TotalCost float64,
    RefType *string, RefID *int64, OccurredAt, CreatedAt string}
```

## Repo arayüzü (birebir)
```go
type Repo interface {
    CreateMaterial(ctx context.Context, m model.Material) (model.Material, error)
    ListMaterials(ctx context.Context) ([]model.Material, error)
    CreateProduct(ctx context.Context, p model.Product) (model.Product, error)
    ListProducts(ctx context.Context) ([]model.Product, error)
    CreateLot(ctx context.Context, l model.Lot) (model.Lot, error)
    GetLotByCode(ctx context.Context, code string) (model.Lot, error)
    AddStockMovement(ctx context.Context, m model.StockMovement) (model.StockMovement, error)
    ListStockMovements(ctx context.Context, lotID int64) ([]model.StockMovement, error)
}
func NewMySQLRepo(db *sql.DB) Repo
```

## Endpoint'ler
| Metot | Yol | Dönüş |
|-------|-----|-------|
| GET | `/health` | `{"status":"ok"}` |
| GET | `/materials` | `[]Material` |
| POST | `/materials` | 201 `Material` |
| GET | `/products` | `[]Product` |
| POST | `/products` | 201 `Product` |
| GET | `/lots` (`?item_type=material\|product` opsiyonel) | `[]Lot` |
| POST | `/lots` | 201 `Lot` |
| GET | `/lots/{code}` | `Lot` |
| GET | `/lots/{code}/movements` | `[]StockMovement` |
| POST | `/stock-movements` | 201 `StockMovement` |

- HTTP standart kütüphane (`net/http`), `http.ServeMux`; Go 1.22+ path parametreleri (`GET /lots/{code}`).
- Tüm SQL **parametreli** (`?`). String birleştirme yasak.
- JSON sayı olarak döner; DECIMAL alanlar `float64` taranır.
- Hata durumunda uygun HTTP kodu + `{"error":"..."}`.

## Test (`server_test.go`, fake repo)
- `TestHealth`: `GET /health` → 200, gövde `{"status":"ok"}`.
- `TestCreateMaterialThenList`: `POST /materials` → 201; `GET /materials` → oluşturulan kayıt listede.
- Fake repo: `Repo` arayüzünü bellek içi uygular (MySQL gerekmez).

## Kabul kriteri (senior kontrol edecek)
- `go build ./...` hatasız.
- `go test ./...` geçer.
- Health + CRUD + stok uçları çalışır.
- Kod İngilizce isimlendirme, okunur, parametreli SQL.
