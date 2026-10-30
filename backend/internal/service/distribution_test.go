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

func (f *fakeButcheryRepository) GetProduct(_ context.Context, id int64) (model.Product, error) {
	return model.Product{ID: id, ListPrice: 10}, nil
}

func TestDistributeValueSumEqualsTotal(t *testing.T) {
	items := []DistItem{
		{ProductID: 1, Weight: 10, Value: 100},
		{ProductID: 2, Weight: 8, Value: 80},
		{ProductID: 3, Weight: 4, Value: 40},
	}

	got := DistributeByValue(287.50, items)

	var sum float64
	for _, item := range got {
		sum += item.AllocatedCost
	}
	if sum != 287.50 {
		t.Fatalf("value sum = %v, want 287.50", sum)
	}
	// Highest-value item (ProductID 1) should get the largest share.
	if got[0].AllocatedCost <= got[2].AllocatedCost {
		t.Fatalf("value-weighted: highest value item should get the largest share, got %v vs %v", got[0].AllocatedCost, got[2].AllocatedCost)
	}
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
