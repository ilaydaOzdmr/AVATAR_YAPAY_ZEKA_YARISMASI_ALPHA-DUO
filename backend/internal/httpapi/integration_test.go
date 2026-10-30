//go:build integration

package httpapi

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"avatar/backend/internal/repo"
)

const defaultHTTPTestDSN = "avatar:avatar@tcp(localhost:3307)/avatar_test?parseTime=true&charset=utf8mb4"

func openTestRepo(t *testing.T) (repo.Repo, *sql.DB) {
	t.Helper()
	dsn := os.Getenv("TEST_MYSQL_DSN")
	if dsn == "" {
		dsn = defaultHTTPTestDSN
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("test mysql not available: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return repo.NewMySQLRepo(db), db
}

func uniqueCode(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
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

func lotCode(t *testing.T, db *sql.DB, id int64) string {
	t.Helper()
	var code string
	if err := db.QueryRow(`SELECT code FROM lots WHERE id = ?`, id).Scan(&code); err != nil {
		t.Fatalf("lot code: %v", err)
	}
	return code
}

func doJSON(t *testing.T, client *http.Client, method, url string, payload any, want int) []byte {
	t.Helper()
	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal payload: %v", err)
		}
		body = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != want {
		t.Fatalf("%s %s status = %d, want %d; body=%s", method, url, resp.StatusCode, want, raw)
	}
	return raw
}

func TestHTTPWorkflowIntegration(t *testing.T) {
	r, db := openTestRepo(t)

	srv := httptest.NewServer(NewServer(r))
	defer srv.Close()
	client := srv.Client()

	materialID := insertMaterial(t, db, uniqueCode("MAT"), 10)
	inputLotID := insertLot(t, db, uniqueCode("ET"), "material", materialID, 10, 10)
	inputLotCode := lotCode(t, db, inputLotID)

	productID := insertProduct(t, db, uniqueCode("PRD"), 25)
	outputLotID := insertLot(t, db, uniqueCode("SUC"), "product", productID, 0, 0)
	outputLotCode := lotCode(t, db, outputLotID)
	marketID := insertMarket(t, db, uniqueCode("MKT"))

	// 1. Production consumes 10 kg and produces 8 kg with a real EKP.
	doJSON(t, client, http.MethodPost, srv.URL+"/production-orders", map[string]any{
		"code":              uniqueCode("URT"),
		"status":            "completed",
		"service_name":      "test",
		"service_unit_cost": 0,
		"service_qty":       0,
		"expected_qty":      8,
		"inputs":            []map[string]any{{"material_lot_id": inputLotID, "qty": 10}},
		"outputs":           []map[string]any{{"product_lot_id": outputLotID, "qty": 8}},
	}, http.StatusCreated)

	var outputQty float64
	if err := db.QueryRow(`SELECT qty FROM lots WHERE id = ?`, outputLotID).Scan(&outputQty); err != nil {
		t.Fatalf("output lot qty: %v", err)
	}
	if outputQty != 8 {
		t.Fatalf("output lot qty = %v, want 8", outputQty)
	}

	// 2. Butchery creates product lots for every part and stocks them.
	material2ID := insertMaterial(t, db, uniqueCode("MAT2"), 11.5)
	butcheryInputID := insertLot(t, db, uniqueCode("KUZU"), "material", material2ID, 20, 11.5)
	part1 := insertProduct(t, db, uniqueCode("BUT"), 30)
	part2 := insertProduct(t, db, uniqueCode("SIRT"), 28)
	raw := doJSON(t, client, http.MethodPost, srv.URL+"/butchery-batches", map[string]any{
		"batch": map[string]any{
			"code":         uniqueCode("PRC"),
			"animal_type":  "kuzu",
			"input_lot_id": butcheryInputID,
			"input_weight": 20,
			"unit_cost":    11.5,
			"total_cost":   230,
			"batch_date":   "2026-10-30",
			"status":       "completed",
		},
		"items": []map[string]any{
			{"product_id": part1, "weight": 8, "quantity": 1},
			{"product_id": part2, "weight": 10, "quantity": 1},
		},
	}, http.StatusCreated)

	var butcheryResult struct {
		Items []struct {
			ProductLotID *int64  `json:"product_lot_id"`
			LotCode      string  `json:"lot_code"`
			Weight       float64 `json:"weight"`
		} `json:"items"`
		WeightDiff    float64 `json:"weight_diff"`
		CostBalanceOK bool    `json:"cost_balance_ok"`
	}
	if err := json.Unmarshal(raw, &butcheryResult); err != nil {
		t.Fatalf("unmarshal butchery: %v", err)
	}
	if len(butcheryResult.Items) != 2 {
		t.Fatalf("butchery items = %d, want 2", len(butcheryResult.Items))
	}
	for _, item := range butcheryResult.Items {
		if item.ProductLotID == nil || item.LotCode == "" {
			t.Fatalf("butchery item missing lot: %+v", item)
		}
	}
	if butcheryResult.WeightDiff != 2 {
		t.Fatalf("weight diff = %v, want 2 (20 in - 18 parts)", butcheryResult.WeightDiff)
	}
	if !butcheryResult.CostBalanceOK {
		t.Fatal("cost balance not ok, want allocated total to equal 230")
	}

	// 3. Sale moves stock and uses the lot cost.
	doJSON(t, client, http.MethodPost, srv.URL+"/sales", map[string]any{
		"code":      uniqueCode("SALE"),
		"market_id": marketID,
		"status":    "completed",
		"lines": []map[string]any{{
			"lot_id":      outputLotID,
			"qty":         2,
			"unit_price":  25,
			"total_price": 50,
		}},
	}, http.StatusCreated)

	var soldQty float64
	if err := db.QueryRow(`SELECT qty FROM lots WHERE id = ?`, outputLotID).Scan(&soldQty); err != nil {
		t.Fatalf("sold lot qty: %v", err)
	}
	if soldQty != 6 {
		t.Fatalf("sold lot qty = %v, want 6", soldQty)
	}

	// 4. Traceability follows the runtime link back to the raw material.
	raw = doJSON(t, client, http.MethodGet, srv.URL+"/trace/backward?lot="+outputLotCode, nil, http.StatusOK)
	var sources []struct {
		LotCode string `json:"lot_code"`
	}
	if err := json.Unmarshal(raw, &sources); err != nil {
		t.Fatalf("unmarshal trace: %v", err)
	}
	if len(sources) != 1 || sources[0].LotCode != inputLotCode {
		t.Fatalf("trace sources = %+v, want %q", sources, inputLotCode)
	}

	// 5. Recall blocks the source and the affected product lot; sales then conflict.
	doJSON(t, client, http.MethodPost, srv.URL+"/recalls", map[string]any{
		"source_lot_code": inputLotCode,
		"reason":          "integration test",
	}, http.StatusCreated)

	doJSON(t, client, http.MethodPost, srv.URL+"/sales", map[string]any{
		"code":      uniqueCode("SALE"),
		"market_id": marketID,
		"status":    "completed",
		"lines": []map[string]any{{
			"lot_id":      outputLotID,
			"qty":         1,
			"unit_price":  25,
			"total_price": 25,
		}},
	}, http.StatusConflict)

	var status string
	if err := db.QueryRow(`SELECT status FROM lots WHERE id = ?`, outputLotID).Scan(&status); err != nil {
		t.Fatalf("output lot status: %v", err)
	}
	if status != "blocked" {
		t.Fatalf("output lot status = %q, want blocked", status)
	}

	// 6. The affected lot is visible in the sync pull blocklist.
	raw = doJSON(t, client, http.MethodGet, srv.URL+"/sync/pull", nil, http.StatusOK)
	var pull struct {
		Blocked []string `json:"blocked_lot_codes"`
	}
	if err := json.Unmarshal(raw, &pull); err != nil {
		t.Fatalf("unmarshal pull: %v", err)
	}
	found := false
	for _, code := range pull.Blocked {
		if code == outputLotCode {
			found = true
		}
	}
	if !found {
		t.Fatalf("blocked codes = %v, want %q", pull.Blocked, outputLotCode)
	}
}
