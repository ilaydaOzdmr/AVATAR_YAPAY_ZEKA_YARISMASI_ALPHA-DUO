package service

import (
	"context"

	"avatar/backend/internal/model"
)

// ButcheryItemCost is a butchery item together with its allocated cost under
// both allocation keys (weight-based and value-based).
type ButcheryItemCost struct {
	ProductID          int64   `json:"product_id"`
	Weight             float64 `json:"weight"`
	Quantity           int     `json:"quantity"`
	IsWaste            bool    `json:"is_waste"`
	AllocatedCost      float64 `json:"allocated_cost"`       // weight-based
	AllocatedCostValue float64 `json:"allocated_cost_value"` // value-based
}

// ButcheryResult is the stored butchery batch plus its balance and cost checks.
type ButcheryResult struct {
	Batch           model.ButcheryBatch `json:"batch"`
	Items           []ButcheryItemCost  `json:"items"`
	WeightBalanceOK bool                `json:"weight_balance_ok"`
	FireWeight      float64             `json:"fire_weight"`
	CostMismatch    bool                `json:"cost_mismatch"`
	ExpectedCost    float64             `json:"expected_cost"`
}

// ButcheryRepository is the data access contract required for butchery.
type ButcheryRepository interface {
	CreateButcheryBatch(ctx context.Context, batch model.ButcheryBatch, items []model.ButcheryItem) (model.ButcheryBatch, error)
	GetButcheryBatch(ctx context.Context, id int64) (model.ButcheryBatch, error)
	ListButcheryItems(ctx context.Context, batchID int64) ([]model.ButcheryItem, error)
	GetProduct(ctx context.Context, id int64) (model.Product, error)
}

// ButcheryService creates butchery batches and computes their cost split.
type ButcheryService struct{ repo ButcheryRepository }

// NewButcheryService returns a ButcheryService backed by the given repository.
func NewButcheryService(r ButcheryRepository) *ButcheryService {
	return &ButcheryService{repo: r}
}

// CreateBatch stores a batch and returns its weight-balance and cost summary.
func (s *ButcheryService) CreateBatch(ctx context.Context, batch model.ButcheryBatch, items []model.ButcheryItem) (*ButcheryResult, error) {
	created, err := s.repo.CreateButcheryBatch(ctx, batch, items)
	if err != nil {
		return nil, err
	}
	return s.buildResult(ctx, created, items), nil
}

// GetBatch loads a stored batch and returns its weight-balance and cost summary.
func (s *ButcheryService) GetBatch(ctx context.Context, id int64) (*ButcheryResult, error) {
	batch, err := s.repo.GetButcheryBatch(ctx, id)
	if err != nil {
		return nil, err
	}
	items, err := s.repo.ListButcheryItems(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.buildResult(ctx, batch, items), nil
}

func (s *ButcheryService) buildResult(ctx context.Context, batch model.ButcheryBatch, items []model.ButcheryItem) *ButcheryResult {
	var partWeights []float64
	var distItems []DistItem
	for _, item := range items {
		if item.IsWaste {
			continue
		}
		partWeights = append(partWeights, item.Weight)
		distItems = append(distItems, DistItem{ProductID: item.ProductID, Weight: item.Weight})
	}

	// Resolve each part's value (weight x list price) for value-weighted allocation.
	for i := range distItems {
		if product, err := s.repo.GetProduct(ctx, distItems[i].ProductID); err == nil {
			distItems[i].Value = distItems[i].Weight * product.ListPrice
		}
	}

	balanceOK, fire := WeightBalanceOK(batch.InputWeight, partWeights)
	expectedCost := round2(batch.InputWeight * batch.UnitCost)
	costMismatch := round2(expectedCost) != round2(batch.TotalCost)

	weightItems := make([]DistItem, len(distItems))
	copy(weightItems, distItems)
	byWeight := DistributeByWeight(batch.TotalCost, weightItems)

	valueItems := make([]DistItem, len(distItems))
	copy(valueItems, distItems)
	byValue := DistributeByValue(batch.TotalCost, valueItems)

	costs := make([]ButcheryItemCost, 0, len(items))
	next := 0
	for _, item := range items {
		allocatedWeight := 0.0
		allocatedValue := 0.0
		if !item.IsWaste {
			allocatedWeight = byWeight[next].AllocatedCost
			allocatedValue = byValue[next].AllocatedCost
			next++
		}
		costs = append(costs, ButcheryItemCost{
			ProductID:          item.ProductID,
			Weight:             item.Weight,
			Quantity:           item.Quantity,
			IsWaste:            item.IsWaste,
			AllocatedCost:      allocatedWeight,
			AllocatedCostValue: allocatedValue,
		})
	}

	return &ButcheryResult{
		Batch:           batch,
		Items:           costs,
		WeightBalanceOK: balanceOK,
		FireWeight:      fire,
		CostMismatch:    costMismatch,
		ExpectedCost:    expectedCost,
	}
}
