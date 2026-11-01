package service

import (
	"context"

	"avatar/backend/internal/repo"
)

// Anomaly is a rule-based alert over production, butchery and stock data.
type Anomaly struct {
	Key      string `json:"key"`
	Type     string `json:"type"`     // fire | cost | yield | stock
	Severity string `json:"severity"` // warning | critical
	Message  string `json:"message"`
	Ref      string `json:"ref"`
	Read     bool   `json:"read"`
}

// AnomalyService detects unexpected fire, cost increase, low yield and
// inconsistent lot movements.
type AnomalyService struct {
	repo repo.Repo
	cost *CostEngine
}

// NewAnomalyService returns an AnomalyService backed by the given repository.
func NewAnomalyService(r repo.Repo) *AnomalyService {
	return &AnomalyService{repo: r, cost: NewCostEngine(r)}
}

// Detect scans the current data and returns all anomalies found.
func (s *AnomalyService) Detect(ctx context.Context) ([]Anomaly, error) {
	anomalies := []Anomaly{}
	checks := []func(context.Context) ([]Anomaly, error){s.butcheryAnomalies, s.productionAnomalies, s.stockAnomalies}
	for _, check := range checks {
		found, err := check(ctx)
		if err != nil {
			return nil, err
		}
		anomalies = append(anomalies, found...)
	}
	readKeys, err := s.repo.ListReadAnomalyKeys(ctx)
	if err != nil {
		return nil, err
	}
	read := make(map[string]bool, len(readKeys))
	for _, key := range readKeys {
		read[key] = true
	}
	for i := range anomalies {
		anomalies[i].Key = anomalies[i].Type + ":" + anomalies[i].Ref
		anomalies[i].Read = read[anomalies[i].Key]
	}
	return anomalies, nil
}

func (s *AnomalyService) butcheryAnomalies(ctx context.Context) ([]Anomaly, error) {
	anomalies := []Anomaly{}
	// 1) Unexpected fire (butchery)
	batches, err := s.repo.ListButcheryBatches(ctx)
	if err != nil {
		return nil, err
	}
	for _, batch := range batches {
		items, err := s.repo.ListButcheryItems(ctx, batch.ID)
		if err != nil {
			return nil, err
		}
		var parts float64
		for _, item := range items {
			if !item.IsWaste {
				parts += item.Weight
			}
		}
		if batch.InputWeight > 0 && (batch.InputWeight-parts)/batch.InputWeight*100 > 25 {
			anomalies = append(anomalies, Anomaly{
				Type: "fire", Severity: "warning",
				Message: "Yüksek fire oranı", Ref: batch.Code,
			})
		}
	}

	return anomalies, nil
}

func (s *AnomalyService) productionAnomalies(ctx context.Context) ([]Anomaly, error) {
	anomalies := []Anomaly{}
	// 2) Cost increase (EKP above average) + 3) low yield
	orders, err := s.repo.ListProductionOrders(ctx)
	if err != nil {
		return nil, err
	}
	var stats []orderStat
	var ekpTotal, ekpCount float64
	for _, order := range orders {
		summary, err := s.cost.ProductionOrderCost(ctx, order.ID)
		if err != nil {
			continue
		}
		if summary.EKP > 0 {
			stats = append(stats, orderStat{code: order.Code, ekp: summary.EKP})
			ekpTotal += summary.EKP
			ekpCount++
		}

		inputQty, outputQty := s.productionQuantities(ctx, order.ID)
		if inputQty > 0 && outputQty/inputQty < 0.80 {
			anomalies = append(anomalies, Anomaly{
				Type: "yield", Severity: "warning",
				Message: "Düşük üretim verimi", Ref: order.Code,
			})
		}
	}
	anomalies = append(anomalies, costAnomalies(stats, ekpTotal, ekpCount)...)
	return anomalies, nil
}

func (s *AnomalyService) stockAnomalies(ctx context.Context) ([]Anomaly, error) {
	anomalies := []Anomaly{}
	// 4) Inconsistent lot movements (negative balance)
	lots, err := s.repo.ListLots(ctx, "")
	if err != nil {
		return nil, err
	}
	for _, lot := range lots {
		movements, err := s.repo.ListStockMovements(ctx, lot.ID)
		if err != nil {
			continue
		}
		var balance float64
		for _, m := range movements {
			balance += m.Qty
		}
		if balance < 0 {
			anomalies = append(anomalies, Anomaly{
				Type: "stock", Severity: "warning",
				Message: "Stok az olduğu için satış engellendi", Ref: lot.Code,
			})
		}
	}

	return anomalies, nil
}

func (s *AnomalyService) productionQuantities(ctx context.Context, orderID int64) (float64, float64) {
	var inputQty, outputQty float64
	if inputs, err := s.repo.ListProductionOrderInputs(ctx, orderID); err == nil {
		for _, in := range inputs {
			inputQty += in.Qty
		}
	}
	if outputs, err := s.repo.ListProductionOrderOutputs(ctx, orderID); err == nil {
		for _, out := range outputs {
			outputQty += out.Qty
		}
	}
	return inputQty, outputQty
}

type orderStat struct {
	code string
	ekp  float64
}

func costAnomalies(stats []orderStat, ekpTotal, ekpCount float64) []Anomaly {
	anomalies := []Anomaly{}
	if ekpCount > 0 {
		avg := ekpTotal / ekpCount
		for _, stat := range stats {
			if stat.ekp > avg*1.10 {
				anomalies = append(anomalies, Anomaly{
					Type: "cost", Severity: "warning",
					Message: "Birim maliyet ortalamanın üzerinde", Ref: stat.code,
				})
			}
		}
	}

	return anomalies
}
