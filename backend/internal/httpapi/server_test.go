package httpapi

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"avatar/backend/internal/model"
	"avatar/backend/internal/repo"
)

// fakeRepo is an in-memory Repo used to test handlers without MySQL.
type fakeRepo struct {
	materials []model.Material

	addStockErr         error
	createProductionErr error
	createButcheryErr   error
	createSaleErr       error
	createRecallErr     error
	getLotByCodeErr     error
}

func (f *fakeRepo) CreateMaterial(_ context.Context, m model.Material) (model.Material, error) {
	m.ID = int64(len(f.materials) + 1)
	f.materials = append(f.materials, m)
	return m, nil
}
func (f *fakeRepo) ListMaterials(_ context.Context) ([]model.Material, error) {
	return f.materials, nil
}
func (f *fakeRepo) CreateProduct(_ context.Context, p model.Product) (model.Product, error) {
	return p, nil
}
func (f *fakeRepo) ListProducts(_ context.Context) ([]model.Product, error)     { return nil, nil }
func (f *fakeRepo) CreateLot(_ context.Context, l model.Lot) (model.Lot, error) { return l, nil }
func (f *fakeRepo) ListLots(_ context.Context, _ string) ([]model.Lot, error)   { return nil, nil }
func (f *fakeRepo) GetLotByCode(_ context.Context, _ string) (model.Lot, error) {
	return model.Lot{}, f.getLotByCodeErr
}
func (f *fakeRepo) ListLotLinksByParent(_ context.Context, _ int64) ([]model.LotLink, error) {
	return nil, nil
}
func (f *fakeRepo) ListLotLinksByChild(_ context.Context, _ int64) ([]model.LotLink, error) {
	return nil, nil
}
func (f *fakeRepo) AddStockMovement(_ context.Context, m model.StockMovement) (model.StockMovement, error) {
	return m, f.addStockErr
}
func (f *fakeRepo) ListStockMovements(_ context.Context, _ int64) ([]model.StockMovement, error) {
	return nil, nil
}
func (f *fakeRepo) ListAllStockMovements(_ context.Context) ([]model.StockMovement, error) {
	return nil, nil
}
func (f *fakeRepo) CreateProductionOrder(_ context.Context, po model.ProductionOrder, _ []model.ProductionOrderInput, _ []model.ProductionOrderOutput) (model.ProductionOrder, error) {
	return po, f.createProductionErr
}
func (f *fakeRepo) GetProductionOrder(_ context.Context, _ int64) (model.ProductionOrder, error) {
	return model.ProductionOrder{}, nil
}
func (f *fakeRepo) ListProductionOrderInputs(_ context.Context, _ int64) ([]model.ProductionOrderInput, error) {
	return nil, nil
}
func (f *fakeRepo) ListProductionOrderCodesByInputLot(_ context.Context, _ int64) ([]string, error) {
	return nil, nil
}
func (f *fakeRepo) ListProductionOrderOutputs(_ context.Context, _ int64) ([]model.ProductionOrderOutput, error) {
	return nil, nil
}
func (f *fakeRepo) GetLot(_ context.Context, _ int64) (model.Lot, error) {
	return model.Lot{}, nil
}
func (f *fakeRepo) GetProduct(_ context.Context, _ int64) (model.Product, error) {
	return model.Product{}, nil
}
func (f *fakeRepo) GetMarket(_ context.Context, _ int64) (model.Market, error) {
	return model.Market{}, nil
}
func (f *fakeRepo) GetSupplier(_ context.Context, _ int64) (model.Supplier, error) {
	return model.Supplier{}, nil
}
func (f *fakeRepo) CreateButcheryBatch(_ context.Context, batch model.ButcheryBatch, _ []model.ButcheryItem) (model.ButcheryBatch, error) {
	return batch, f.createButcheryErr
}
func (f *fakeRepo) GetButcheryBatch(_ context.Context, _ int64) (model.ButcheryBatch, error) {
	return model.ButcheryBatch{}, nil
}
func (f *fakeRepo) ListButcheryItems(_ context.Context, _ int64) ([]model.ButcheryItem, error) {
	return nil, nil
}
func (f *fakeRepo) ListButcheryBatches(_ context.Context) ([]model.ButcheryBatch, error) {
	return nil, nil
}
func (f *fakeRepo) ListProductionOrders(_ context.Context) ([]model.ProductionOrder, error) {
	return nil, nil
}
func (f *fakeRepo) CreateRecall(_ context.Context, e model.RecallEvent, _ []model.RecallBlock) (model.RecallEvent, error) {
	return e, f.createRecallErr
}
func (f *fakeRepo) CreateRecallEvent(_ context.Context, e model.RecallEvent) (model.RecallEvent, error) {
	return e, nil
}
func (f *fakeRepo) GetRecallEvent(_ context.Context, _ int64) (model.RecallEvent, error) {
	return model.RecallEvent{}, nil
}
func (f *fakeRepo) CreateRecallBlock(_ context.Context, b model.RecallBlock) (model.RecallBlock, error) {
	return b, nil
}
func (f *fakeRepo) ListRecallBlocks(_ context.Context, _ int64) ([]model.RecallBlock, error) {
	return nil, nil
}
func (f *fakeRepo) IsLotBlocked(_ context.Context, _ int64) (bool, error) {
	return false, nil
}
func (f *fakeRepo) ListBlockedLotCodes(_ context.Context) ([]string, error) {
	return nil, nil
}
func (f *fakeRepo) CreateSale(_ context.Context, sale model.Sale, _ []model.SaleLine) (model.Sale, error) {
	return sale, f.createSaleErr
}
func (f *fakeRepo) ListSales(_ context.Context) ([]model.Sale, error) {
	return nil, nil
}
func (f *fakeRepo) ListSaleLines(_ context.Context, _ int64) ([]model.SaleLine, error) {
	return nil, nil
}
func (f *fakeRepo) ListRecallEvents(_ context.Context) ([]model.RecallEvent, error) {
	return nil, nil
}
func (f *fakeRepo) ListMarkets(_ context.Context) ([]model.Market, error) {
	return nil, nil
}
func (f *fakeRepo) CreateRedistribution(_ context.Context, red model.Redistribution) (model.Redistribution, error) {
	return red, nil
}
func (f *fakeRepo) ListRedistributions(_ context.Context, _ int64) ([]model.Redistribution, error) {
	return nil, nil
}

func TestHealth(t *testing.T) {
	h := NewServer(&fakeRepo{})
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("status field = %q, want %q", body["status"], "ok")
	}
}

func TestCreateMaterialThenList(t *testing.T) {
	h := NewServer(&fakeRepo{})
	payload := `{"code":"MAT-009","name":"Test","unit":"kg","is_raw":true,"default_purchase_price":1.5}`

	req := httptest.NewRequest(http.MethodPost, "/materials", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d", rec.Code, http.StatusCreated)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/materials", nil)
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d", rec2.Code, http.StatusOK)
	}
	var items []model.Material
	if err := json.Unmarshal(rec2.Body.Bytes(), &items); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(items) != 1 || items[0].Code != "MAT-009" {
		t.Fatalf("unexpected materials: %+v", items)
	}
}

func TestErrorStatusMapping(t *testing.T) {
	cases := []struct {
		name   string
		method string
		path   string
		body   string
		repo   *fakeRepo
		want   int
	}{
		{
			name:   "stock insufficient is conflict",
			method: http.MethodPost,
			path:   "/stock-movements",
			body:   `{"lot_id":1,"movement_type":"sale","qty":-1}`,
			repo:   &fakeRepo{addStockErr: repo.ErrInsufficientStock},
			want:   http.StatusConflict,
		},
		{
			name:   "stock invalid movement is bad request",
			method: http.MethodPost,
			path:   "/stock-movements",
			body:   `{"lot_id":1,"movement_type":"sale","qty":0}`,
			repo:   &fakeRepo{addStockErr: repo.ErrInvalidStockMovement},
			want:   http.StatusBadRequest,
		},
		{
			name:   "production invalid order is bad request",
			method: http.MethodPost,
			path:   "/production-orders",
			body:   `{}`,
			repo:   &fakeRepo{createProductionErr: repo.ErrInvalidOrder},
			want:   http.StatusBadRequest,
		},
		{
			name:   "production insufficient stock is conflict",
			method: http.MethodPost,
			path:   "/production-orders",
			body:   `{}`,
			repo:   &fakeRepo{createProductionErr: repo.ErrInsufficientStock},
			want:   http.StatusConflict,
		},
		{
			name:   "butchery invalid order is bad request",
			method: http.MethodPost,
			path:   "/butchery-batches",
			body:   `{"batch":{},"items":[]}`,
			repo:   &fakeRepo{createButcheryErr: repo.ErrInvalidOrder},
			want:   http.StatusBadRequest,
		},
		{
			name:   "butchery insufficient stock is conflict",
			method: http.MethodPost,
			path:   "/butchery-batches",
			body:   `{"batch":{},"items":[]}`,
			repo:   &fakeRepo{createButcheryErr: repo.ErrInsufficientStock},
			want:   http.StatusConflict,
		},
		{
			name:   "blocked sale is conflict",
			method: http.MethodPost,
			path:   "/sales",
			body:   `{"code":"S-1","market_id":1,"status":"completed","lines":[]}`,
			repo:   &fakeRepo{createSaleErr: repo.ErrLotBlocked},
			want:   http.StatusConflict,
		},
		{
			name:   "recall invalid order is bad request",
			method: http.MethodPost,
			path:   "/recalls",
			body:   `{"source_lot_code":"X"}`,
			repo:   &fakeRepo{createRecallErr: repo.ErrInvalidOrder},
			want:   http.StatusBadRequest,
		},
		{
			name:   "recall unknown lot is not found",
			method: http.MethodPost,
			path:   "/recalls",
			body:   `{"source_lot_code":"UNKNOWN"}`,
			repo:   &fakeRepo{getLotByCodeErr: sql.ErrNoRows},
			want:   http.StatusNotFound,
		},
		{
			name:   "unknown lot is not found",
			method: http.MethodGet,
			path:   "/lots/UNKNOWN",
			repo:   &fakeRepo{getLotByCodeErr: sql.ErrNoRows},
			want:   http.StatusNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := NewServer(tc.repo)
			var body *bytes.Buffer
			if tc.body != "" {
				body = bytes.NewBufferString(tc.body)
			} else {
				body = bytes.NewBuffer(nil)
			}
			req := httptest.NewRequest(tc.method, tc.path, body)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != tc.want {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tc.want, rec.Body.String())
			}
		})
	}
}

func TestRootAndReportSummary(t *testing.T) {
	handler := NewServerWithML(&fakeRepo{}, nil)

	root := httptest.NewRecorder()
	handler.ServeHTTP(root, httptest.NewRequest(http.MethodGet, "/", nil))
	if root.Code != http.StatusOK {
		t.Fatalf("root status = %d, want 200", root.Code)
	}
	var rootBody map[string]any
	if err := json.Unmarshal(root.Body.Bytes(), &rootBody); err != nil {
		t.Fatalf("root json: %v", err)
	}
	if rootBody["service"] != "avatar-backend" {
		t.Fatalf("root service = %v", rootBody["service"])
	}

	report := httptest.NewRecorder()
	handler.ServeHTTP(report, httptest.NewRequest(http.MethodGet, "/reports/summary", nil))
	if report.Code != http.StatusOK {
		t.Fatalf("report status = %d, body=%s", report.Code, report.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(report.Body.Bytes(), &body); err != nil {
		t.Fatalf("report json: %v", err)
	}
	for _, key := range []string{"purchases", "production", "butchery", "lot_movements", "sales", "recalls"} {
		if _, ok := body[key]; !ok {
			t.Fatalf("report missing %q: %s", key, report.Body.String())
		}
	}
}

func TestSyncPullIncludesFullReferenceSnapshot(t *testing.T) {
	handler := NewServerWithML(&fakeRepo{}, nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/sync/pull", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json: %v", err)
	}
	for _, key := range []string{"server_time", "blocked_lot_codes", "products", "lots", "markets"} {
		if _, ok := body[key]; !ok {
			t.Fatalf("sync pull missing %q: %s", key, rec.Body.String())
		}
	}
}
