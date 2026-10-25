package service

import (
	"context"
	"math"

	"avatar/backend/internal/repo"
)

// CostSummary is the cost breakdown for a production order.
type CostSummary struct {
	InputCost   float64 `json:"input_cost"`
	ServiceCost float64 `json:"service_cost"`
	TotalCost   float64 `json:"total_cost"`
	FinishedQty float64 `json:"finished_qty"`
	EKP         float64 `json:"ekp"`
	VKP         float64 `json:"vkp"`
	GrossProfit float64 `json:"gross_profit"`
	Margin      float64 `json:"margin"`
}

// CostEngine computes production order costs (EKP, VKP, gross profit, margin).
type CostEngine struct {
	repo repo.Repo
}

// NewCostEngine returns a CostEngine backed by the given repo.
func NewCostEngine(r repo.Repo) *CostEngine {
	return &CostEngine{repo: r}
}

func round2(x float64) float64 {
	return math.Round(x*100) / 100
}

// EKP is the real unit cost per accepted finished unit.
func EKP(inputCost, serviceCost, finishedQty float64) float64 {
	if finishedQty <= 0 {
		return 0
	}
	return round2((inputCost + serviceCost) / finishedQty)
}

// GrossProfit is the per-unit sale price minus the real unit cost.
func GrossProfit(vkp, ekp float64) float64 {
	return round2(vkp - ekp)
}

// Margin is the gross profit as a percentage of the sale price.
func Margin(vkp, ekp float64) float64 {
	if vkp <= 0 {
		return 0
	}
	return round2((vkp-ekp)/vkp*100)
}

// ProductionOrderCost builds the cost summary for a production order.
func (e *CostEngine) ProductionOrderCost(ctx context.Context, poID int64) (CostSummary, error) {
	po, err := e.repo.GetProductionOrder(ctx, poID)
	if err != nil {
		return CostSummary{}, err
	}

	inputs, err := e.repo.ListProductionOrderInputs(ctx, poID)
	if err != nil {
		return CostSummary{}, err
	}

	var inputCost float64
	for _, input := range inputs {
		lot, err := e.repo.GetLot(ctx, input.MaterialLotID)
		if err != nil {
			return CostSummary{}, err
		}
		inputCost += input.Qty * lot.UnitCost
	}

	serviceCost := po.ServiceUnitCost * po.ServiceQty
	totalCost := inputCost + serviceCost

	outputs, err := e.repo.ListProductionOrderOutputs(ctx, poID)
	if err != nil {
		return CostSummary{}, err
	}

	var finishedQty float64
	for _, output := range outputs {
		finishedQty += output.Qty
	}

	ekp := EKP(inputCost, serviceCost, finishedQty)

	var vkp float64
	if len(outputs) > 0 {
		lot, err := e.repo.GetLot(ctx, outputs[0].ProductLotID)
		if err != nil {
			return CostSummary{}, err
		}
		prod, err := e.repo.GetProduct(ctx, lot.ItemID)
		if err != nil {
			return CostSummary{}, err
		}
		vkp = prod.ListPrice
	}

	return CostSummary{
		InputCost:   round2(inputCost),
		ServiceCost: round2(serviceCost),
		TotalCost:   round2(totalCost),
		FinishedQty: finishedQty,
		EKP:         ekp,
		VKP:         vkp,
		GrossProfit: GrossProfit(vkp, ekp),
		Margin:      Margin(vkp, ekp),
	}, nil
}
