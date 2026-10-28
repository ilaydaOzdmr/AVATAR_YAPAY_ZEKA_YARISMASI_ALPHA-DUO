package service

import (
	"context"

	"avatar/backend/internal/model"
)

// ButcheryItemCost is a butchery item together with its allocated cost.
type ButcheryItemCost struct {
	ProductID     int64   `json:"product_id"`
	Weight        float64 `json:"weight"`
	Quantity      int     `json:"quantity"`
	IsWaste       bool    `json:"is_waste"`
	AllocatedCost float64 `json:"allocated_cost"`
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
	return s.buildResult(created, items), nil
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
	return s.buildResult(batch, items), nil
}

func (s *ButcheryService) buildResult(batch model.ButcheryBatch, items []model.ButcheryItem) *ButcheryResult {
	var partWeights []float64
	var distItems []DistItem
	for _, item := range items {
		if item.IsWaste {
			continue
		}
		partWeights = append(partWeights, item.Weight)
		distItems = append(distItems, DistItem{ProductID: item.ProductID, Weight: item.Weight})
	}

	balanceOK, fire := WeightBalanceOK(batch.InputWeight, partWeights)
	expectedCost := round2(batch.InputWeight * batch.UnitCost)
	costMismatch := round2(expectedCost) != round2(batch.TotalCost)

	allocated := DistributeByWeight(batch.TotalCost, distItems)

	costs := make([]ButcheryItemCost, 0, len(items))
	next := 0
	for _, item := range items {
		allocatedCost := 0.0
		if !item.IsWaste {
			allocatedCost = allocated[next].AllocatedCost
			next++
		}
		costs = append(costs, ButcheryItemCost{
			ProductID:     item.ProductID,
			Weight:        item.Weight,
			Quantity:      item.Quantity,
			IsWaste:       item.IsWaste,
			AllocatedCost: allocatedCost,
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
