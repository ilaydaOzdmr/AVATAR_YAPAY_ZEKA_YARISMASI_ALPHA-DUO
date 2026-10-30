package service

import (
	"context"
	"testing"

	"avatar/backend/internal/model"
)

// storedButcheryRepository returns pre-built stored items so the service can be
// tested without MySQL. CreateButcheryBatch assigns an ID and the subsequent
// ListButcheryItems call returns the rows the repo would have persisted.
type storedButcheryRepository struct {
	batch model.ButcheryBatch
	items []model.ButcheryItem
}

func (f *storedButcheryRepository) CreateButcheryBatch(_ context.Context, batch model.ButcheryBatch, _ []model.ButcheryItem) (model.ButcheryBatch, error) {
	batch.ID = 42
	f.batch = batch
	return batch, nil
}

func (f *storedButcheryRepository) GetButcheryBatch(_ context.Context, _ int64) (model.ButcheryBatch, error) {
	return f.batch, nil
}

func (f *storedButcheryRepository) ListButcheryItems(_ context.Context, _ int64) ([]model.ButcheryItem, error) {
	return f.items, nil
}

func (f *storedButcheryRepository) GetProduct(_ context.Context, id int64) (model.Product, error) {
	return model.Product{ID: id, ListPrice: 10}, nil
}

func TestCreateBatchReloadsStoredItemsAndReportsBalance(t *testing.T) {
	lotIDs := []int64{101, 102, 103, 104, 105, 106}
	weights := []float64{4.5, 3.0, 3.5, 2.5, 4.0, 4.5}
	costs := []float64{71.80, 47.86, 55.84, 39.89, 63.82, 71.79}
	items := make([]model.ButcheryItem, 0, len(weights))
	for i, weight := range weights {
		lotID := lotIDs[i]
		items = append(items, model.ButcheryItem{
			ProductID:     int64(i + 1),
			ProductLotID:  &lotID,
			LotCode:       "PRC-1-P" + string(rune('A'+i)),
			Weight:        weight,
			Quantity:      1,
			AllocatedCost: costs[i],
		})
	}

	repo := &storedButcheryRepository{items: items}
	result, err := NewButcheryService(repo).CreateBatch(context.Background(), model.ButcheryBatch{
		Code:        "PRC-1",
		AnimalType:  "kuzu",
		InputWeight: 25,
		UnitCost:    11.50,
		TotalCost:   351,
		Status:      "completed",
	}, nil)
	if err != nil {
		t.Fatalf("CreateBatch: %v", err)
	}

	if result.Batch.ID != 42 {
		t.Fatalf("batch ID = %d, want 42", result.Batch.ID)
	}
	if len(result.Items) != len(items) {
		t.Fatalf("items = %d, want %d", len(result.Items), len(items))
	}

	var allocated float64
	for i, item := range result.Items {
		if item.LotCode != items[i].LotCode {
			t.Fatalf("item %d lot code = %q, want %q", i, item.LotCode, items[i].LotCode)
		}
		if item.ProductLotID == nil || *item.ProductLotID != *items[i].ProductLotID {
			t.Fatalf("item %d product lot = %v, want %v", i, item.ProductLotID, *items[i].ProductLotID)
		}
		allocated += item.AllocatedCost
	}
	if !approxFloat(allocated, 351) {
		t.Fatalf("allocated total = %v, want 351", allocated)
	}

	if !approxFloat(result.WeightDiff, 3) {
		t.Fatalf("WeightDiff = %v, want 3", result.WeightDiff)
	}
	if !approxFloat(result.FireWeight, 3) {
		t.Fatalf("FireWeight = %v, want 3", result.FireWeight)
	}
	if !result.CostBalanceOK {
		t.Fatalf("CostBalanceOK = false, want true (diff %v)", result.CostDiff)
	}
	if !approxFloat(result.CostDiff, 0) {
		t.Fatalf("CostDiff = %v, want 0", result.CostDiff)
	}
}

func TestBuildResultCountsWasteInBalance(t *testing.T) {
	batch := model.ButcheryBatch{InputWeight: 25, UnitCost: 10, TotalCost: 200}
	items := []model.ButcheryItem{
		{ProductID: 1, Weight: 20, AllocatedCost: 200},
		{ProductID: 2, Weight: 2, IsWaste: true},
	}

	result := NewButcheryService(&storedButcheryRepository{}).buildResult(context.Background(), batch, items)

	// 20 kg parts + 2 kg explicit waste leave 3 kg fire.
	if !approxFloat(result.WeightDiff, 3) {
		t.Fatalf("WeightDiff = %v, want 3", result.WeightDiff)
	}
	if !approxFloat(result.FireWeight, 5) {
		t.Fatalf("FireWeight = %v, want 5 (fire + explicit waste)", result.FireWeight)
	}
}

func approxFloat(a, b float64) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d < 0.01
}
