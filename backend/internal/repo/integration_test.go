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

// approx reports whether a and b are equal within a cent/milligram tolerance.
func approx(a, b float64) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d < 0.001
}

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

func TestCreateButcheryBatchCreatesLotsLinksAndMovements(t *testing.T) {
	r, db := testRepo(t)
	ctx := context.Background()

	materialID := insertMaterial(t, db, unique("MAT"), 11.50)
	inputLotID := insertLot(t, db, unique("KUZU"), "material", materialID, 25, 11.50)

	productIDs := []int64{
		insertProduct(t, db, unique("BUT"), 30),
		insertProduct(t, db, unique("SIRT"), 28),
		insertProduct(t, db, unique("OMUZ"), 24),
		insertProduct(t, db, unique("KABURGA"), 20),
		insertProduct(t, db, unique("KIYMA"), 18),
		insertProduct(t, db, unique("KUSBASI"), 26),
		insertProduct(t, db, unique("KEMIK"), 5),
	}
	weights := []float64{4.5, 3.0, 3.5, 2.0, 3.5, 3.5, 2.0}
	items := make([]model.ButcheryItem, 0, len(productIDs))
	for i, productID := range productIDs {
		items = append(items, model.ButcheryItem{
			ProductID: productID,
			Weight:    weights[i],
			Quantity:  1,
		})
	}

	created, err := r.CreateButcheryBatch(ctx, model.ButcheryBatch{
		Code:        unique("PRC"),
		AnimalType:  "kuzu",
		InputLotID:  &inputLotID,
		InputWeight: 25,
		UnitCost:    11.50,
		TotalCost:   351,
		BatchDate:   "2026-10-30",
		Status:      "completed",
	}, items)
	if err != nil {
		t.Fatalf("CreateButcheryBatch: %v", err)
	}

	storedItems, err := r.ListButcheryItems(ctx, created.ID)
	if err != nil {
		t.Fatalf("ListButcheryItems: %v", err)
	}
	if len(storedItems) != len(items) {
		t.Fatalf("stored items = %d, want %d", len(storedItems), len(items))
	}
	var allocatedTotal float64
	for _, item := range storedItems {
		if item.ProductLotID == nil || *item.ProductLotID == 0 {
			t.Fatalf("item %d has no product lot: %+v", item.ProductID, item)
		}
		if item.LotCode == "" {
			t.Fatalf("item %d has no lot code: %+v", item.ProductID, item)
		}
		allocatedTotal += item.AllocatedCost
	}
	if !approx(allocatedTotal, 351) {
		t.Fatalf("allocated total = %v, want 351", allocatedTotal)
	}

	links, err := r.ListLotLinksByParent(ctx, inputLotID)
	if err != nil {
		t.Fatalf("ListLotLinksByParent: %v", err)
	}
	if len(links) != len(items) {
		t.Fatalf("links = %d, want %d", len(links), len(items))
	}
	var linkedWeight float64
	for _, link := range links {
		if link.LinkType != "butchery" {
			t.Fatalf("link type = %q, want butchery", link.LinkType)
		}
		linkedWeight += link.Qty
	}
	if !approx(linkedWeight, 22) {
		t.Fatalf("linked weight = %v, want 22", linkedWeight)
	}

	inputLot, err := r.GetLot(ctx, inputLotID)
	if err != nil {
		t.Fatalf("GetLot input: %v", err)
	}
	if inputLot.Qty != 0 {
		t.Fatalf("input lot qty = %v, want 0", inputLot.Qty)
	}

	movements, err := r.ListStockMovements(ctx, inputLotID)
	if err != nil {
		t.Fatalf("input movements: %v", err)
	}
	var consume, waste float64
	for _, movement := range movements {
		switch movement.MovementType {
		case "butchery_consume":
			consume += movement.Qty
		case "waste":
			waste += movement.Qty
		}
	}
	if !approx(consume, -22) || !approx(waste, -3) {
		t.Fatalf("consume/waste = %v/%v, want -22/-3", consume, waste)
	}

	for i, item := range storedItems {
		lot, err := r.GetLot(ctx, *item.ProductLotID)
		if err != nil {
			t.Fatalf("GetLot output %d: %v", i, err)
		}
		if !approx(lot.Qty, weights[i]) {
			t.Fatalf("output lot %d qty = %v, want %v", i, lot.Qty, weights[i])
		}
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

func TestCreateRecallBlocksLotsAndHandlesCodeCollision(t *testing.T) {
	r, db := testRepo(t)
	ctx := context.Background()

	materialID := insertMaterial(t, db, unique("MAT"), 10)
	sourceLotID := insertLot(t, db, unique("SRC"), "material", materialID, 100, 10)
	productID := insertProduct(t, db, unique("PRD"), 25)
	productLotID := insertLot(t, db, unique("PRC"), "product", productID, 50, 18)
	marketID := insertMarket(t, db, unique("MKT"))

	if _, err := db.Exec(
		`INSERT INTO lot_links (parent_lot_id, child_lot_id, qty, link_type) VALUES (?, ?, 50, 'production')`,
		sourceLotID, productLotID); err != nil {
		t.Fatalf("insert lot link: %v", err)
	}

	code := unique("REC")
	event, err := r.CreateRecall(ctx, model.RecallEvent{
		Code:        code,
		Reason:      "test recall",
		SourceLotID: sourceLotID,
	}, []model.RecallBlock{{LotID: sourceLotID}, {LotID: productLotID}})
	if err != nil {
		t.Fatalf("CreateRecall: %v", err)
	}
	if event.ID == 0 || event.Code != code {
		t.Fatalf("event = %+v, want id>0 code %q", event, code)
	}

	for _, lotID := range []int64{sourceLotID, productLotID} {
		var status string
		if err := db.QueryRow(`SELECT status FROM lots WHERE id = ?`, lotID).Scan(&status); err != nil {
			t.Fatalf("lot %d status: %v", lotID, err)
		}
		if status != "blocked" {
			t.Fatalf("lot %d status = %q, want blocked", lotID, status)
		}
	}

	blocked, err := r.IsLotBlocked(ctx, productLotID)
	if err != nil {
		t.Fatalf("IsLotBlocked: %v", err)
	}
	if !blocked {
		t.Fatal("IsLotBlocked = false, want true")
	}

	// Same requested code must get a unique suffix instead of failing.
	event2, err := r.CreateRecall(ctx, model.RecallEvent{
		Code:        code,
		Reason:      "second recall",
		SourceLotID: sourceLotID,
	}, []model.RecallBlock{{LotID: productLotID}})
	if err != nil {
		t.Fatalf("CreateRecall collision: %v", err)
	}
	if event2.Code != code+"-2" {
		t.Fatalf("collision code = %q, want %q", event2.Code, code+"-2")
	}

	if _, err := r.CreateSale(ctx, model.Sale{
		Code:     unique("SALE"),
		MarketID: marketID,
		Status:   "completed",
	}, []model.SaleLine{{
		LotID:      productLotID,
		Qty:        1,
		UnitPrice:  25,
		TotalPrice: 25,
	}}); !errors.Is(err, ErrLotBlocked) {
		t.Fatalf("sale error = %v, want ErrLotBlocked", err)
	}

	codes, err := r.ListBlockedLotCodes(ctx)
	if err != nil {
		t.Fatalf("ListBlockedLotCodes: %v", err)
	}
	var productLotCode string
	if err := db.QueryRow(`SELECT code FROM lots WHERE id = ?`, productLotID).Scan(&productLotCode); err != nil {
		t.Fatalf("product lot code: %v", err)
	}
	found := false
	for _, c := range codes {
		if c == productLotCode {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("blocked codes = %v, want to include %q", codes, productLotCode)
	}

	// An invalid block must roll the whole recall back.
	badCode := unique("RECBAD")
	if _, err := r.CreateRecall(ctx, model.RecallEvent{
		Code:        badCode,
		Reason:      "bad",
		SourceLotID: sourceLotID,
	}, []model.RecallBlock{{LotID: 999999999}}); err == nil {
		t.Fatal("CreateRecall with invalid lot: want error")
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM recall_events WHERE code = ?`, badCode).Scan(&count); err != nil {
		t.Fatalf("count events: %v", err)
	}
	if count != 0 {
		t.Fatalf("recall events = %d, want 0 after rollback", count)
	}
}
