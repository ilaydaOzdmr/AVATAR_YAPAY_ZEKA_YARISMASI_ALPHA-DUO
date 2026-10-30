package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"avatar/backend/internal/model"
)

// fakeRepo is an in-memory Repo used to test handlers without MySQL.
type fakeRepo struct {
	materials []model.Material
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
	return model.Lot{}, nil
}
func (f *fakeRepo) ListLotLinksByParent(_ context.Context, _ int64) ([]model.LotLink, error) {
	return nil, nil
}
func (f *fakeRepo) ListLotLinksByChild(_ context.Context, _ int64) ([]model.LotLink, error) {
	return nil, nil
}
func (f *fakeRepo) AddStockMovement(_ context.Context, m model.StockMovement) (model.StockMovement, error) {
	return m, nil
}
func (f *fakeRepo) ListStockMovements(_ context.Context, _ int64) ([]model.StockMovement, error) {
	return nil, nil
}
func (f *fakeRepo) CreateProductionOrder(_ context.Context, po model.ProductionOrder, _ []model.ProductionOrderInput, _ []model.ProductionOrderOutput) (model.ProductionOrder, error) {
	return po, nil
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
	return batch, nil
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
	return e, nil
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
	return sale, nil
}
func (f *fakeRepo) ListSales(_ context.Context) ([]model.Sale, error) {
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
