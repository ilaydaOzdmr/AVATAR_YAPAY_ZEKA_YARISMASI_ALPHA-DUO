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

func TestBuildResultDemoBatchReportsFireAndCostDifference(t *testing.T) {
	batch := model.ButcheryBatch{
		Code:        "KUZ-2026-DEMO",
		InputWeight: 25,
		UnitCost:    11.50,
		TotalCost:   351,
	}
	items := []model.ButcheryItem{
		{ProductID: 2, Weight: 10},
		{ProductID: 3, Weight: 8},
		{ProductID: 6, Weight: 4},
	}

	result := NewButcheryService(&storedButcheryRepository{}).buildResult(context.Background(), batch, items)

	if !approxFloat(result.PartWeight, 22) {
		t.Fatalf("PartWeight = %v, want 22", result.PartWeight)
	}
	if !approxFloat(result.FireWeight, 3) {
		t.Fatalf("FireWeight = %v, want 3", result.FireWeight)
	}
	if !approxFloat(result.WeightDiff, 3) {
		t.Fatalf("WeightDiff = %v, want 3", result.WeightDiff)
	}
	if !approxFloat(result.ExpectedCost, 287.50) {
		t.Fatalf("ExpectedCost = %v, want 287.50", result.ExpectedCost)
	}
	if !approxFloat(result.RecordedCost, 351) {
		t.Fatalf("RecordedCost = %v, want 351", result.RecordedCost)
	}
	if !approxFloat(result.CostDifference, 63.50) {
		t.Fatalf("CostDifference = %v, want 63.50", result.CostDifference)
	}
	if result.CostCheckStatus != CheckMismatch {
		t.Fatalf("CostCheckStatus = %q, want %q", result.CostCheckStatus, CheckMismatch)
	}
	if result.WeightCheckStatus != CheckOK {
		t.Fatalf("WeightCheckStatus = %q, want %q (3 kg fire is normal)", result.WeightCheckStatus, CheckOK)
	}
}

func TestBuildResultCostAndWeightChecksAreIndependent(t *testing.T) {
	repo := &storedButcheryRepository{}

	// Balanced weights, wrong recorded cost: only the cost check warns.
	costOnly := NewButcheryService(repo).buildResult(context.Background(), model.ButcheryBatch{
		InputWeight: 20, UnitCost: 10, TotalCost: 250,
	}, []model.ButcheryItem{{ProductID: 1, Weight: 20}})
	if costOnly.WeightCheckStatus != CheckOK {
		t.Fatalf("cost-only weight status = %q, want %q", costOnly.WeightCheckStatus, CheckOK)
	}
	if costOnly.CostCheckStatus != CheckMismatch {
		t.Fatalf("cost-only cost status = %q, want %q", costOnly.CostCheckStatus, CheckMismatch)
	}

	// Parts exceeding the input weight, correct recorded cost: only the
	// weight check warns because fire cannot be negative.
	weightOnly := NewButcheryService(repo).buildResult(context.Background(), model.ButcheryBatch{
		InputWeight: 20, UnitCost: 10, TotalCost: 200,
	}, []model.ButcheryItem{{ProductID: 1, Weight: 22}})
	if weightOnly.CostCheckStatus != CheckOK {
		t.Fatalf("weight-only cost status = %q, want %q", weightOnly.CostCheckStatus, CheckOK)
	}
	if weightOnly.WeightCheckStatus != CheckMismatch {
		t.Fatalf("weight-only weight status = %q, want %q", weightOnly.WeightCheckStatus, CheckMismatch)
	}
	if !approxFloat(weightOnly.CostDifference, 0) {
		t.Fatalf("weight-only cost difference = %v, want 0", weightOnly.CostDifference)
	}
}

func TestBuildResultRecordedWasteAccountsForWeight(t *testing.T) {
	// 5 kg is recorded as waste, so the weight balance is consistent even
	// though fire (input minus saleable parts) is 5 kg.
	result := NewButcheryService(&storedButcheryRepository{}).buildResult(context.Background(), model.ButcheryBatch{
		InputWeight: 25, UnitCost: 10, TotalCost: 250,
	}, []model.ButcheryItem{
		{ProductID: 1, Weight: 20},
		{ProductID: 2, Weight: 5, IsWaste: true},
	})

	if !approxFloat(result.FireWeight, 5) {
		t.Fatalf("FireWeight = %v, want 5", result.FireWeight)
	}
	if result.WeightCheckStatus != CheckOK {
		t.Fatalf("WeightCheckStatus = %q, want %q (waste row is recorded)", result.WeightCheckStatus, CheckOK)
	}
	if result.CostCheckStatus != CheckOK {
		t.Fatalf("CostCheckStatus = %q, want %q", result.CostCheckStatus, CheckOK)
	}
}

func TestBuildResultConsistentBatchHasNoWarnings(t *testing.T) {
	result := NewButcheryService(&storedButcheryRepository{}).buildResult(context.Background(), model.ButcheryBatch{
		InputWeight: 50, UnitCost: 12.40, TotalCost: 620,
	}, []model.ButcheryItem{
		{ProductID: 1, Weight: 30},
		{ProductID: 2, Weight: 20},
	})

	if result.WeightCheckStatus != CheckOK {
		t.Fatalf("WeightCheckStatus = %q, want %q", result.WeightCheckStatus, CheckOK)
	}
	if result.CostCheckStatus != CheckOK {
		t.Fatalf("CostCheckStatus = %q, want %q", result.CostCheckStatus, CheckOK)
	}
	if !approxFloat(result.FireWeight, 0) {
		t.Fatalf("FireWeight = %v, want 0", result.FireWeight)
	}
	if !approxFloat(result.CostDifference, 0) {
		t.Fatalf("CostDifference = %v, want 0", result.CostDifference)
	}
}

func TestBuildResultUsesTwoAndThreeDecimalTolerances(t *testing.T) {
	// A cost difference below half a cent is not a mismatch.
	costOK := NewButcheryService(&storedButcheryRepository{}).buildResult(context.Background(), model.ButcheryBatch{
		InputWeight: 10, UnitCost: 10, TotalCost: 100.004,
	}, []model.ButcheryItem{{ProductID: 1, Weight: 10}})
	if costOK.CostCheckStatus != CheckOK {
		t.Fatalf("cost status = %q, want %q", costOK.CostCheckStatus, CheckOK)
	}

	// A one-gram excess is a mismatch at three decimals.
	weightDiff := NewButcheryService(&storedButcheryRepository{}).buildResult(context.Background(), model.ButcheryBatch{
		InputWeight: 10, UnitCost: 10, TotalCost: 100,
	}, []model.ButcheryItem{{ProductID: 1, Weight: 10.001}})
	if weightDiff.WeightCheckStatus != CheckMismatch {
		t.Fatalf("weight status = %q, want %q", weightDiff.WeightCheckStatus, CheckMismatch)
	}

	// A smaller excess rounds away at three decimals.
	weightOK := NewButcheryService(&storedButcheryRepository{}).buildResult(context.Background(), model.ButcheryBatch{
		InputWeight: 10, UnitCost: 10, TotalCost: 100,
	}, []model.ButcheryItem{{ProductID: 1, Weight: 10.0004}})
	if weightOK.WeightCheckStatus != CheckOK {
		t.Fatalf("rounded weight status = %q, want %q", weightOK.WeightCheckStatus, CheckOK)
	}
}

func TestBuildResultRejectsExcessiveFire(t *testing.T) {
	repo := &storedButcheryRepository{}

	// 50 kg input, 27 kg parts: 23 kg fire (46%) is not a consistent record.
	excessive := NewButcheryService(repo).buildResult(context.Background(), model.ButcheryBatch{
		InputWeight: 50, UnitCost: 10, TotalCost: 500,
	}, []model.ButcheryItem{{ProductID: 1, Weight: 27}})
	if excessive.WeightCheckStatus != CheckMismatch {
		t.Fatalf("excessive fire status = %q, want %q", excessive.WeightCheckStatus, CheckMismatch)
	}
	if !approxFloat(excessive.FireWeight, 23) {
		t.Fatalf("excessive fire weight = %v, want 23", excessive.FireWeight)
	}
	if excessive.CostCheckStatus != CheckOK {
		t.Fatalf("excessive fire cost status = %q, want %q", excessive.CostCheckStatus, CheckOK)
	}

	// 50 kg input, 40 kg parts: exactly 10 kg (20%) fire is the allowed limit.
	boundary := NewButcheryService(repo).buildResult(context.Background(), model.ButcheryBatch{
		InputWeight: 50, UnitCost: 10, TotalCost: 500,
	}, []model.ButcheryItem{{ProductID: 1, Weight: 40}})
	if boundary.WeightCheckStatus != CheckOK {
		t.Fatalf("boundary fire status = %q, want %q (10 kg at 50 kg)", boundary.WeightCheckStatus, CheckOK)
	}
}

func approxFloat(a, b float64) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d < 0.01
}
