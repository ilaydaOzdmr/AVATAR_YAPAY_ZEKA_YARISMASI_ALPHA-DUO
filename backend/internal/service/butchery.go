package service

import (
	"context"
	"math"

	"avatar/backend/internal/butchery"
	"avatar/backend/internal/model"
)

// CheckOK and CheckMismatch are the status values of the weight and cost
// checks. Fire on its own is not a data error.
const (
	CheckOK       = butchery.CheckOK
	CheckMismatch = butchery.CheckMismatch
)

// ButcheryItemCost is a butchery item together with its allocated cost under
// both allocation keys (weight-based and value-based).
type ButcheryItemCost struct {
	ProductID          int64   `json:"product_id"`
	ProductLotID       *int64  `json:"product_lot_id,omitempty"`
	LotCode            string  `json:"lot_code,omitempty"`
	Weight             float64 `json:"weight"`
	Quantity           int     `json:"quantity"`
	IsWaste            bool    `json:"is_waste"`
	AllocatedCost      float64 `json:"allocated_cost"`       // weight-based
	AllocatedCostValue float64 `json:"allocated_cost_value"` // value-based
}

// ButcheryResult is the stored butchery batch plus its balance and cost checks.
type ButcheryResult struct {
	Batch             model.ButcheryBatch `json:"batch"`
	Items             []ButcheryItemCost  `json:"items"`
	WeightBalanceOK   bool                `json:"weight_balance_ok"`
	WeightCheckStatus string              `json:"weight_check_status"`
	PartWeight        float64             `json:"part_weight"`
	FireWeight        float64             `json:"fire_weight"`
	WeightDiff        float64             `json:"weight_diff"`
	CostMismatch      bool                `json:"cost_mismatch"`
	CostCheckStatus   string              `json:"cost_check_status"`
	ExpectedCost      float64             `json:"expected_cost"`
	RecordedCost      float64             `json:"recorded_cost"`
	CostDifference    float64             `json:"cost_difference"`
	CostBalanceOK     bool                `json:"cost_balance_ok"`
	CostDiff          float64             `json:"cost_diff"`
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
// The result is built from the rows the repository persisted, so lot codes and
// allocated costs reflect what was actually written.
func (s *ButcheryService) CreateBatch(ctx context.Context, batch model.ButcheryBatch, items []model.ButcheryItem) (*ButcheryResult, error) {
	created, err := s.repo.CreateButcheryBatch(ctx, batch, items)
	if err != nil {
		return nil, err
	}
	stored, err := s.repo.ListButcheryItems(ctx, created.ID)
	if err != nil {
		return nil, err
	}
	return s.buildResult(ctx, created, stored), nil
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
	var distItems []DistItem
	for _, item := range items {
		if item.IsWaste {
			continue
		}
		distItems = append(distItems, DistItem{ProductID: item.ProductID, Weight: item.Weight})
	}

	// Resolve each part's value (weight x list price) for value-weighted allocation.
	for i := range distItems {
		if product, err := s.repo.GetProduct(ctx, distItems[i].ProductID); err == nil {
			distItems[i].Value = distItems[i].Weight * product.ListPrice
		}
	}

	checks := butchery.Compute(batch, items)
	batch.ExpectedCost = checks.ExpectedCost
	batch.CostDifference = checks.CostDifference
	batch.CostCheckStatus = checks.CostCheckStatus
	batch.WeightCheckStatus = checks.WeightCheckStatus

	weightItems := make([]DistItem, len(distItems))
	copy(weightItems, distItems)
	byWeight := DistributeByWeight(batch.TotalCost, weightItems)

	valueItems := make([]DistItem, len(distItems))
	copy(valueItems, distItems)
	byValue := DistributeByValue(batch.TotalCost, valueItems)

	costs := make([]ButcheryItemCost, 0, len(items))
	next := 0
	var allocatedTotal float64
	for _, item := range items {
		allocatedWeight := 0.0
		allocatedValue := 0.0
		if !item.IsWaste {
			allocatedWeight = byWeight[next].AllocatedCost
			allocatedValue = byValue[next].AllocatedCost
			// Prefer the persisted weight-based cost; it is the source of truth
			// for stock valuation and may differ from a re-computation.
			if item.AllocatedCost != 0 {
				allocatedWeight = item.AllocatedCost
			}
			next++
		}
		allocatedTotal += allocatedWeight
		costs = append(costs, ButcheryItemCost{
			ProductID:          item.ProductID,
			ProductLotID:       item.ProductLotID,
			LotCode:            item.LotCode,
			Weight:             item.Weight,
			Quantity:           item.Quantity,
			IsWaste:            item.IsWaste,
			AllocatedCost:      allocatedWeight,
			AllocatedCostValue: allocatedValue,
		})
	}

	costDiff := round2(allocatedTotal - batch.TotalCost)
	return &ButcheryResult{
		Batch:             batch,
		Items:             costs,
		WeightBalanceOK:   checks.WeightCheckStatus == CheckOK,
		WeightCheckStatus: checks.WeightCheckStatus,
		PartWeight:        checks.PartWeight,
		FireWeight:        checks.FireWeight,
		WeightDiff:        checks.WeightDiff,
		CostMismatch:      checks.CostCheckStatus == CheckMismatch,
		CostCheckStatus:   checks.CostCheckStatus,
		ExpectedCost:      checks.ExpectedCost,
		RecordedCost:      checks.RecordedCost,
		CostDifference:    checks.CostDifference,
		CostBalanceOK:     round2(costDiff) == 0,
		CostDiff:          costDiff,
	}
}

func round3(v float64) float64 {
	return math.Round(v*1000) / 1000
}
