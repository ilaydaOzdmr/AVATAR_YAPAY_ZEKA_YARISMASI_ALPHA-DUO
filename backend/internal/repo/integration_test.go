//go:build integration

package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"avatar/backend/internal/model"
)

const defaultTestDSN = "avatar:avatar@tcp(localhost:3307)/avatar_test?parseTime=true&charset=utf8mb4"

func testRepo(t *testing.T) (Repo, *sql.DB) {
	t.Helper()
	dsn := os.Getenv("TEST_MYSQL_DSN")
	if dsn == "" {
		dsn = defaultTestDSN
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("test mysql not available: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return NewMySQLRepo(db), db
}

func unique(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

func stringPtr(v string) *string { return &v }

func insertMaterial(t *testing.T, db *sql.DB, code string, price float64) int64 {
	t.Helper()
	res, err := db.Exec(
		`INSERT INTO materials (code, name, unit, is_raw, default_purchase_price) VALUES (?, ?, 'kg', TRUE, ?)`,
		code, code, price)
	if err != nil {
		t.Fatalf("insert material: %v", err)
	}
	id, _ := res.LastInsertId()
	return id
}

func insertProduct(t *testing.T, db *sql.DB, code string, listPrice float64) int64 {
	t.Helper()
	res, err := db.Exec(
		`INSERT INTO products (code, name, unit, list_price, category, is_finished) VALUES (?, ?, 'kg', ?, 'part', TRUE)`,
		code, code, listPrice)
	if err != nil {
		t.Fatalf("insert product: %v", err)
	}
	id, _ := res.LastInsertId()
	return id
}

func insertLot(t *testing.T, db *sql.DB, code, itemType string, itemID int64, qty, unitCost float64) int64 {
	t.Helper()
	res, err := db.Exec(
		`INSERT INTO lots (code, item_type, item_id, qty, unit, unit_cost, status) VALUES (?, ?, ?, ?, 'kg', ?, 'active')`,
		code, itemType, itemID, qty, unitCost)
	if err != nil {
		t.Fatalf("insert lot: %v", err)
	}
	id, _ := res.LastInsertId()
	return id
}

func insertMarket(t *testing.T, db *sql.DB, code string) int64 {
	t.Helper()
	res, err := db.Exec(`INSERT INTO markets (code, name, contact) VALUES (?, ?, '0212 000 00 00')`, code, code)
	if err != nil {
		t.Fatalf("insert market: %v", err)
	}
	id, _ := res.LastInsertId()
	return id
}

func TestAddStockMovementUpdatesLotQty(t *testing.T) {
	r, db := testRepo(t)
	ctx := context.Background()

	materialID := insertMaterial(t, db, unique("MAT"), 9.70)
	lotID := insertLot(t, db, unique("LOT"), "material", materialID, 10, 9.70)

	if _, err := r.AddStockMovement(ctx, model.StockMovement{
		LotID:        lotID,
		MovementType: "transfer",
		Qty:          -4,
	}); err != nil {
		t.Fatalf("AddStockMovement: %v", err)
	}

	lot, err := r.GetLot(ctx, lotID)
	if err != nil {
		t.Fatalf("GetLot: %v", err)
	}
	if lot.Qty != 6 {
		t.Fatalf("lot qty = %v, want 6", lot.Qty)
	}

	movements, err := r.ListStockMovements(ctx, lotID)
	if err != nil {
		t.Fatalf("ListStockMovements: %v", err)
	}
	if len(movements) != 1 {
		t.Fatalf("movements = %d, want 1", len(movements))
	}
	if movements[0].UnitCost != 9.70 || movements[0].TotalCost != 38.80 {
		t.Fatalf("movement cost = %+v, want unit 9.70 total 38.80", movements[0])
	}
}

func TestAddStockMovementRejectsInsufficientStock(t *testing.T) {
	r, db := testRepo(t)
	ctx := context.Background()

	materialID := insertMaterial(t, db, unique("MAT"), 9.70)
	lotID := insertLot(t, db, unique("LOT"), "material", materialID, 6, 9.70)

	_, err := r.AddStockMovement(ctx, model.StockMovement{
		LotID:        lotID,
		MovementType: "sale",
		Qty:          -7,
	})
	if !errors.Is(err, ErrInsufficientStock) {
		t.Fatalf("error = %v, want ErrInsufficientStock", err)
	}

	lot, err := r.GetLot(ctx, lotID)
	if err != nil {
		t.Fatalf("GetLot: %v", err)
	}
	if lot.Qty != 6 {
		t.Fatalf("lot qty = %v, want 6", lot.Qty)
	}

	movements, err := r.ListStockMovements(ctx, lotID)
	if err != nil {
		t.Fatalf("ListStockMovements: %v", err)
	}
	if len(movements) != 0 {
		t.Fatalf("movements = %d, want 0", len(movements))
	}
}

func TestCreateProductionOrderPersistsLinksAndStock(t *testing.T) {
	r, db := testRepo(t)
	ctx := context.Background()

	materialID := insertMaterial(t, db, unique("MAT"), 9.70)
	inputLotID := insertLot(t, db, unique("ET"), "material", materialID, 100, 9.70)
	productID := insertProduct(t, db, unique("PRD"), 24.90)
	outputLotID := insertLot(t, db, unique("SUC"), "product", productID, 0, 0)

	po, err := r.CreateProductionOrder(ctx, model.ProductionOrder{
		Code:            unique("URT"),
		Status:          "completed",
		ServiceName:     stringPtr("Dış üretici"),
		ServiceUnitCost: 5.50,
		ServiceQty:      85,
		ExpectedQty:     85,
	}, []model.ProductionOrderInput{{
		MaterialLotID: inputLotID,
		Qty:           100,
	}}, []model.ProductionOrderOutput{{
		ProductLotID: outputLotID,
		Qty:          85,
	}})
	if err != nil {
		t.Fatalf("CreateProductionOrder: %v", err)
	}

	inputLot, err := r.GetLot(ctx, inputLotID)
	if err != nil {
		t.Fatalf("GetLot input: %v", err)
	}
	if inputLot.Qty != 0 {
		t.Fatalf("input lot qty = %v, want 0", inputLot.Qty)
	}

	outputLot, err := r.GetLot(ctx, outputLotID)
	if err != nil {
		t.Fatalf("GetLot output: %v", err)
	}
	if outputLot.Qty != 85 {
		t.Fatalf("output lot qty = %v, want 85", outputLot.Qty)
	}
	if outputLot.UnitCost != 16.91 {
		t.Fatalf("output lot unit cost = %v, want 16.91", outputLot.UnitCost)
	}

	links, err := r.ListLotLinksByParent(ctx, inputLotID)
	if err != nil {
		t.Fatalf("ListLotLinksByParent: %v", err)
	}
	if len(links) != 1 || links[0].ChildLotID != outputLotID || links[0].Qty != 100 || links[0].LinkType != "production" {
		t.Fatalf("links = %+v, want one production link qty 100", links)
	}

	movements, err := r.ListStockMovements(ctx, inputLotID)
	if err != nil {
		t.Fatalf("input movements: %v", err)
	}
	if len(movements) != 1 || movements[0].MovementType != "production_consume" || movements[0].Qty != -100 {
		t.Fatalf("input movements = %+v, want production_consume -100", movements)
	}
	if movements[0].RefID == nil || *movements[0].RefID != po.ID {
		t.Fatalf("input movement ref = %+v, want order %d", movements[0], po.ID)
	}

	outputMovements, err := r.ListStockMovements(ctx, outputLotID)
	if err != nil {
		t.Fatalf("output movements: %v", err)
	}
	if len(outputMovements) != 1 || outputMovements[0].MovementType != "production_output" || outputMovements[0].Qty != 85 {
		t.Fatalf("output movements = %+v, want production_output +85", outputMovements)
	}
}

func TestCreateProductionOrderSplitsLinksAcrossOutputs(t *testing.T) {
	r, db := testRepo(t)
	ctx := context.Background()

	materialID := insertMaterial(t, db, unique("MAT"), 10)
	inputLotID := insertLot(t, db, unique("ET"), "material", materialID, 100, 10)
	productA := insertProduct(t, db, unique("PRD-A"), 20)
	productB := insertProduct(t, db, unique("PRD-B"), 30)
	outputA := insertLot(t, db, unique("OUT-A"), "product", productA, 0, 0)
	outputB := insertLot(t, db, unique("OUT-B"), "product", productB, 0, 0)

	if _, err := r.CreateProductionOrder(ctx, model.ProductionOrder{
		Code:   unique("URT"),
		Status: "completed",
	}, []model.ProductionOrderInput{{
		MaterialLotID: inputLotID,
		Qty:           100,
	}}, []model.ProductionOrderOutput{
		{ProductLotID: outputA, Qty: 60},
		{ProductLotID: outputB, Qty: 40},
	}); err != nil {
		t.Fatalf("CreateProductionOrder: %v", err)
	}

	links, err := r.ListLotLinksByParent(ctx, inputLotID)
	if err != nil {
		t.Fatalf("ListLotLinksByParent: %v", err)
	}
	if len(links) != 2 {
		t.Fatalf("links = %d, want 2", len(links))
	}
	if links[0].Qty != 60 || links[1].Qty != 40 {
		t.Fatalf("link quantities = %v/%v, want 60/40", links[0].Qty, links[1].Qty)
	}
}

func TestCreateProductionOrderRejectsInsufficientStock(t *testing.T) {
	r, db := testRepo(t)
	ctx := context.Background()

	materialID := insertMaterial(t, db, unique("MAT"), 9.70)
	inputLotID := insertLot(t, db, unique("ET"), "material", materialID, 100, 9.70)
	productID := insertProduct(t, db, unique("PRD"), 24.90)
	outputLotID := insertLot(t, db, unique("SUC"), "product", productID, 0, 0)
	orderCode := unique("URT")

	_, err := r.CreateProductionOrder(ctx, model.ProductionOrder{
		Code:   orderCode,
		Status: "completed",
	}, []model.ProductionOrderInput{{
		MaterialLotID: inputLotID,
		Qty:           120,
	}}, []model.ProductionOrderOutput{{
		ProductLotID: outputLotID,
		Qty:          85,
	}})
	if !errors.Is(err, ErrInsufficientStock) {
		t.Fatalf("error = %v, want ErrInsufficientStock", err)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM production_orders WHERE code = ?`, orderCode).Scan(&count); err != nil {
		t.Fatalf("count orders: %v", err)
	}
	if count != 0 {
		t.Fatalf("order rows = %d, want 0 (transaction rollback)", count)
	}

	lot, err := r.GetLot(ctx, inputLotID)
	if err != nil {
		t.Fatalf("GetLot: %v", err)
	}
	if lot.Qty != 100 {
		t.Fatalf("input lot qty = %v, want 100", lot.Qty)
	}
}

func TestCreateCompletedProductionOrderRequiresInputsAndOutputs(t *testing.T) {
	r, _ := testRepo(t)
	ctx := context.Background()

	_, err := r.CreateProductionOrder(ctx, model.ProductionOrder{
		Code:   unique("URT"),
		Status: "completed",
	}, nil, nil)
	if !errors.Is(err, ErrInvalidOrder) {
		t.Fatalf("error = %v, want ErrInvalidOrder", err)
	}
}

func TestCreateSaleUsesLotCostAndUpdatesStock(t *testing.T) {
	r, db := testRepo(t)
	ctx := context.Background()

	productID := insertProduct(t, db, unique("PRD"), 24.90)
	lotID := insertLot(t, db, unique("SUC"), "product", productID, 10, 18.16)
	marketID := insertMarket(t, db, unique("MKT"))

	sale, err := r.CreateSale(ctx, model.Sale{
		Code:     unique("SALE"),
		MarketID: marketID,
		Status:   "completed",
	}, []model.SaleLine{{
		LotID:      lotID,
		Qty:        2,
		UnitPrice:  24.90,
		TotalPrice: 49.80,
	}})
	if err != nil {
		t.Fatalf("CreateSale: %v", err)
	}

	lot, err := r.GetLot(ctx, lotID)
	if err != nil {
		t.Fatalf("GetLot: %v", err)
	}
	if lot.Qty != 8 {
		t.Fatalf("lot qty = %v, want 8", lot.Qty)
	}

	movements, err := r.ListStockMovements(ctx, lotID)
	if err != nil {
		t.Fatalf("ListStockMovements: %v", err)
	}
	if len(movements) != 1 {
		t.Fatalf("movements = %d, want 1", len(movements))
	}
	if movements[0].UnitCost != 18.16 || movements[0].TotalCost != 36.32 {
		t.Fatalf("movement = %+v, want unit cost 18.16 total 36.32", movements[0])
	}
	if movements[0].RefID == nil || *movements[0].RefID != sale.ID {
		t.Fatalf("movement ref = %+v, want sale %d", movements[0], sale.ID)
	}

	var stockMovementID sql.NullInt64
	if err := db.QueryRow(`SELECT stock_movement_id FROM sale_lines WHERE sale_id = ?`, sale.ID).Scan(&stockMovementID); err != nil {
		t.Fatalf("sale line: %v", err)
	}
	if !stockMovementID.Valid || stockMovementID.Int64 != movements[0].ID {
		t.Fatalf("sale line stock_movement_id = %+v, want %d", stockMovementID, movements[0].ID)
	}
}
