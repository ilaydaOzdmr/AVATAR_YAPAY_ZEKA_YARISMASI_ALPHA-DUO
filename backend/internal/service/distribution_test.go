package service

import (
	"context"
	"testing"

	"avatar/backend/internal/model"
)

func TestDistributeSumEqualsTotal(t *testing.T) {
	items := []DistItem{
		{ProductID: 1, Weight: 10},
		{ProductID: 2, Weight: 8},
		{ProductID: 3, Weight: 4},
	}

	got := DistributeByWeight(287.50, items)

	var sum float64
	for _, item := range got {
		sum += item.AllocatedCost
	}
	if sum != 287.50 {
		t.Fatalf("sum = %v, want 287.50", sum)
	}
}

func TestWeightBalanceFlag(t *testing.T) {
	ok, fire := WeightBalanceOK(25, []float64{10, 8, 4})
	if ok {
		t.Fatal("ok = true, want false")
	}
	if fire != 3.0 {
		t.Fatalf("fire = %v, want 3.0", fire)
	}
}

type fakeButcheryRepository struct {
	batch model.ButcheryBatch
	items []model.ButcheryItem
}

func (f *fakeButcheryRepository) CreateButcheryBatch(_ context.Context, batch model.ButcheryBatch, items []model.ButcheryItem) (model.ButcheryBatch, error) {
	return batch, nil
}

func (f *fakeButcheryRepository) GetButcheryBatch(_ context.Context, _ int64) (model.ButcheryBatch, error) {
	return f.batch, nil
}

func (f *fakeButcheryRepository) ListButcheryItems(_ context.Context, _ int64) ([]model.ButcheryItem, error) {
	return f.items, nil
}

func TestCostMismatchFlag(t *testing.T) {
	repo := &fakeButcheryRepository{
		batch: model.ButcheryBatch{
			ID:          1,
			InputWeight: 25,
			UnitCost:    11.50,
			TotalCost:   351.00,
		},
		items: []model.ButcheryItem{
			{ProductID: 1, Weight: 25},
		},
	}

	result, err := NewButcheryService(repo).GetBatch(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetBatch: %v", err)
	}
	if !result.CostMismatch {
		t.Fatalf("CostMismatch = false, want true (expected %v, total %v)", result.ExpectedCost, result.Batch.TotalCost)
	}
	if result.ExpectedCost != 287.50 {
		t.Fatalf("ExpectedCost = %v, want 287.50", result.ExpectedCost)
	}
}
