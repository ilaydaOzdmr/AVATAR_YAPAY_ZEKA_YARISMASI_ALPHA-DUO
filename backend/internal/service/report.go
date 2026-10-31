package service

import (
	"context"
	"time"

	"avatar/backend/internal/model"
	"avatar/backend/internal/repo"
)

// PurchaseReportRow is a material lot purchase line.
type PurchaseReportRow struct {
	LotCode      string  `json:"lot_code"`
	MaterialCode string  `json:"material_code"`
	MaterialName string  `json:"material_name"`
	SupplierCode string  `json:"supplier_code"`
	Qty          float64 `json:"qty"`
	UnitCost     float64 `json:"unit_cost"`
	TotalCost    float64 `json:"total_cost"`
	ReceivedAt   *string `json:"received_at"`
}

// ProductionReportRow is a production order summary with real EKP.
type ProductionReportRow struct {
	Code        string  `json:"code"`
	Status      string  `json:"status"`
	InputKg     float64 `json:"input_kg"`
	FinishedQty float64 `json:"finished_qty"`
	TotalCost   float64 `json:"total_cost"`
	EKP         float64 `json:"ekp"`
}

// ButcheryReportRow is a butchery batch summary.
type ButcheryReportRow struct {
	Code        string  `json:"code"`
	AnimalType  string  `json:"animal_type"`
	InputWeight float64 `json:"input_weight"`
	PartsWeight float64 `json:"parts_weight"`
	FireWeight  float64 `json:"fire_weight"`
	TotalCost   float64 `json:"total_cost"`
	PartCount   int     `json:"part_count"`
}

// MovementReportRow is a lot-based stock movement.
type MovementReportRow struct {
	LotCode      string  `json:"lot_code"`
	MovementType string  `json:"movement_type"`
	Qty          float64 `json:"qty"`
	UnitCost     float64 `json:"unit_cost"`
	TotalCost    float64 `json:"total_cost"`
	OccurredAt   string  `json:"occurred_at"`
}

// SalesReportRow is a sale with aggregated line quantities.
type SalesReportRow struct {
	Code       string  `json:"code"`
	MarketCode string  `json:"market_code"`
	SoldAt     string  `json:"sold_at"`
	LineCount  int     `json:"line_count"`
	Qty        float64 `json:"qty"`
	TotalPrice float64 `json:"total_price"`
}

// RecallReportRow is a recall event with its blocked lots.
type RecallReportRow struct {
	Code          string   `json:"code"`
	Reason        string   `json:"reason"`
	CreatedAt     string   `json:"created_at"`
	BlockedLots   []string `json:"blocked_lots"`
	BlockedLotCnt int      `json:"blocked_lot_count"`
}

// ReportSummary is the server-authoritative report snapshot for mobile/web.
type ReportSummary struct {
	GeneratedAt  string                `json:"generated_at"`
	Purchases    []PurchaseReportRow   `json:"purchases"`
	Production   []ProductionReportRow `json:"production"`
	Butchery     []ButcheryReportRow   `json:"butchery"`
	LotMovements []MovementReportRow   `json:"lot_movements"`
	Sales        []SalesReportRow      `json:"sales"`
	Recalls      []RecallReportRow     `json:"recalls"`
}

// ReportService builds reports from MySQL, which is the source of truth.
type ReportService struct {
	repo repo.Repo
	cost *CostEngine
}

// NewReportService returns a ReportService backed by the given repository.
func NewReportService(r repo.Repo) *ReportService {
	return &ReportService{repo: r, cost: NewCostEngine(r)}
}

// Summary returns all six required report sections in one response.
func (s *ReportService) Summary(ctx context.Context) (ReportSummary, error) {
	summary := ReportSummary{
		GeneratedAt:  time.Now().UTC().Format(time.RFC3339),
		Purchases:    make([]PurchaseReportRow, 0),
		Production:   make([]ProductionReportRow, 0),
		Butchery:     make([]ButcheryReportRow, 0),
		LotMovements: make([]MovementReportRow, 0),
		Sales:        make([]SalesReportRow, 0),
		Recalls:      make([]RecallReportRow, 0),
	}

	materials, err := s.repo.ListMaterials(ctx)
	if err != nil {
		return summary, err
	}
	materialByID := make(map[int64]model.Material, len(materials))
	for _, material := range materials {
		materialByID[material.ID] = material
	}

	lots, err := s.repo.ListLots(ctx, "")
	if err != nil {
		return summary, err
	}
	lotByID := make(map[int64]model.Lot, len(lots))
	for _, lot := range lots {
		lotByID[lot.ID] = lot
	}
	if err := s.appendPurchases(ctx, &summary, lots, materialByID); err != nil {
		return summary, err
	}

	if err := s.appendProduction(ctx, &summary); err != nil {
		return summary, err
	}

	if err := s.appendButchery(ctx, &summary); err != nil {
		return summary, err
	}

	if err := s.appendMovements(ctx, &summary, lotByID); err != nil {
		return summary, err
	}

	if err := s.appendSales(ctx, &summary); err != nil {
		return summary, err
	}

	if err := s.appendRecalls(ctx, &summary, lotByID); err != nil {
		return summary, err
	}

	return summary, nil
}

func (s *ReportService) appendPurchases(ctx context.Context, summary *ReportSummary, lots []model.Lot, materialByID map[int64]model.Material) error {
	for _, lot := range lots {
		if lot.ItemType != "material" {
			continue
		}
		material := materialByID[lot.ItemID]
		supplierCode := ""
		if lot.SupplierID != nil {
			if supplier, err := s.repo.GetSupplier(ctx, *lot.SupplierID); err == nil {
				supplierCode = supplier.Code
			}
		}
		summary.Purchases = append(summary.Purchases, PurchaseReportRow{
			LotCode:      lot.Code,
			MaterialCode: material.Code,
			MaterialName: material.Name,
			SupplierCode: supplierCode,
			Qty:          round3(lot.Qty),
			UnitCost:     round2(lot.UnitCost),
			TotalCost:    round2(lot.Qty * lot.UnitCost),
			ReceivedAt:   lot.ReceivedAt,
		})
	}

	return nil
}

func (s *ReportService) appendProduction(ctx context.Context, summary *ReportSummary) error {
	orders, err := s.repo.ListProductionOrders(ctx)
	if err != nil {
		return err
	}
	for _, order := range orders {
		inputs, err := s.repo.ListProductionOrderInputs(ctx, order.ID)
		if err != nil {
			return err
		}
		var inputKg float64
		for _, input := range inputs {
			inputKg += input.Qty
		}
		cost, err := s.cost.ProductionOrderCost(ctx, order.ID)
		if err != nil {
			return err
		}
		summary.Production = append(summary.Production, ProductionReportRow{
			Code:        order.Code,
			Status:      order.Status,
			InputKg:     round3(inputKg),
			FinishedQty: round3(cost.FinishedQty),
			TotalCost:   round2(cost.TotalCost),
			EKP:         round2(cost.EKP),
		})
	}

	return nil
}

func (s *ReportService) appendButchery(ctx context.Context, summary *ReportSummary) error {
	batches, err := s.repo.ListButcheryBatches(ctx)
	if err != nil {
		return err
	}
	for _, batch := range batches {
		items, err := s.repo.ListButcheryItems(ctx, batch.ID)
		if err != nil {
			return err
		}
		var partsWeight, allWeight float64
		partCount := 0
		for _, item := range items {
			allWeight += item.Weight
			if !item.IsWaste {
				partsWeight += item.Weight
				partCount++
			}
		}
		fire := batch.InputWeight - allWeight
		if fire < 0 {
			fire = 0
		}
		summary.Butchery = append(summary.Butchery, ButcheryReportRow{
			Code:        batch.Code,
			AnimalType:  batch.AnimalType,
			InputWeight: round3(batch.InputWeight),
			PartsWeight: round3(partsWeight),
			FireWeight:  round3(fire),
			TotalCost:   round2(batch.TotalCost),
			PartCount:   partCount,
		})
	}

	return nil
}

func (s *ReportService) appendMovements(ctx context.Context, summary *ReportSummary, lotByID map[int64]model.Lot) error {
	movements, err := s.repo.ListAllStockMovements(ctx)
	if err != nil {
		return err
	}
	for _, movement := range movements {
		summary.LotMovements = append(summary.LotMovements, MovementReportRow{
			LotCode:      lotByID[movement.LotID].Code,
			MovementType: movement.MovementType,
			Qty:          round3(movement.Qty),
			UnitCost:     round2(movement.UnitCost),
			TotalCost:    round2(movement.TotalCost),
			OccurredAt:   movement.OccurredAt,
		})
	}

	return nil
}

func (s *ReportService) appendSales(ctx context.Context, summary *ReportSummary) error {
	markets, err := s.repo.ListMarkets(ctx)
	if err != nil {
		return err
	}
	marketByID := make(map[int64]model.Market, len(markets))
	for _, market := range markets {
		marketByID[market.ID] = market
	}
	sales, err := s.repo.ListSales(ctx)
	if err != nil {
		return err
	}
	for _, sale := range sales {
		lines, err := s.repo.ListSaleLines(ctx, sale.ID)
		if err != nil {
			return err
		}
		var qty, total float64
		for _, line := range lines {
			qty += line.Qty
			total += line.TotalPrice
		}
		summary.Sales = append(summary.Sales, SalesReportRow{
			Code:       sale.Code,
			MarketCode: marketByID[sale.MarketID].Code,
			SoldAt:     sale.SoldAt,
			LineCount:  len(lines),
			Qty:        round3(qty),
			TotalPrice: round2(total),
		})
	}

	return nil
}

func (s *ReportService) appendRecalls(ctx context.Context, summary *ReportSummary, lotByID map[int64]model.Lot) error {
	events, err := s.repo.ListRecallEvents(ctx)
	if err != nil {
		return err
	}
	for _, event := range events {
		blocks, err := s.repo.ListRecallBlocks(ctx, event.ID)
		if err != nil {
			return err
		}
		blocked := make([]string, 0, len(blocks))
		for _, block := range blocks {
			if lot, ok := lotByID[block.LotID]; ok {
				blocked = append(blocked, lot.Code)
			}
		}
		summary.Recalls = append(summary.Recalls, RecallReportRow{
			Code:          event.Code,
			Reason:        event.Reason,
			CreatedAt:     event.CreatedAt,
			BlockedLots:   blocked,
			BlockedLotCnt: len(blocked),
		})
	}

	return nil
}
