package service

import (
	"context"
	"math"

	"avatar/backend/internal/model"
)

// MarketImpact is a shipped quantity of an affected product lot for a market.
type MarketImpact struct {
	MarketCode string  `json:"market_code"`
	MarketName string  `json:"market_name"`
	Contact    string  `json:"contact"`
	Qty        float64 `json:"qty"`
}

// AffectedLot is a product lot reachable from a source lot, with its shipments.
type AffectedLot struct {
	LotCode     string           `json:"lot_code"`
	ProductCode string           `json:"product_code"`
	ProductName string           `json:"product_name"`
	ItemType    string           `json:"item_type"`
	Qty         float64          `json:"qty"`
	Unit        string           `json:"unit"`
	Markets     []MarketImpact   `json:"markets"`
	SubBuyers   []SubBuyerImpact `json:"sub_buyers"`
}

// SubBuyerImpact is a second-level distribution from a market to a sub-buyer.
type SubBuyerImpact struct {
	MarketCode string  `json:"market_code"`
	MarketName string  `json:"market_name"`
	SubBuyer   string  `json:"sub_buyer"`
	Contact    string  `json:"contact"`
	Qty        float64 `json:"qty"`
}

// RawMaterialSource is a material lot, its supplier and production orders behind a product lot.
type RawMaterialSource struct {
	LotCode          string   `json:"lot_code"`
	SupplierCode     string   `json:"supplier_code"`
	SupplierName     string   `json:"supplier_name"`
	ProductionOrders []string `json:"production_orders"`
	Qty              float64  `json:"qty"`
	ReceivedAt       string   `json:"received_at"`
	ExpiryDate       string   `json:"expiry_date"`
}

// TraceRepository is the data access contract required for traceability.
type TraceRepository interface {
	GetLotByCode(ctx context.Context, code string) (model.Lot, error)
	GetLot(ctx context.Context, id int64) (model.Lot, error)
	ListLotLinksByParent(ctx context.Context, parentID int64) ([]model.LotLink, error)
	ListLotLinksByChild(ctx context.Context, childID int64) ([]model.LotLink, error)
	ListStockMovements(ctx context.Context, lotID int64) ([]model.StockMovement, error)
	ListRedistributions(ctx context.Context, lotID int64) ([]model.Redistribution, error)
	GetMarket(ctx context.Context, id int64) (model.Market, error)
	GetSupplier(ctx context.Context, id int64) (model.Supplier, error)
	GetProduct(ctx context.Context, id int64) (model.Product, error)
	ListProductionOrderCodesByInputLot(ctx context.Context, lotID int64) ([]string, error)
}

// Traceability traces lots forward to markets and backward to raw materials.
type Traceability struct{ repo TraceRepository }

// NewTraceability returns a Traceability backed by the given repository.
func NewTraceability(r TraceRepository) *Traceability { return &Traceability{repo: r} }

// Forward returns product lots reachable from sourceLotCode and their shipments.
func (t *Traceability) Forward(ctx context.Context, sourceLotCode string) ([]AffectedLot, error) {
	start, err := t.repo.GetLotByCode(ctx, sourceLotCode)
	if err != nil {
		return nil, err
	}

	visited := map[int64]bool{start.ID: true}
	queue := []model.Lot{start}
	var affected []AffectedLot

	for len(queue) > 0 {
		lot := queue[0]
		queue = queue[1:]

		if lot.ItemType == "product" {
			markets, err := t.marketImpacts(ctx, lot.ID)
			if err != nil {
				return nil, err
			}
			subBuyers, err := t.subBuyerImpacts(ctx, lot.ID)
			if err != nil {
				return nil, err
			}
			affectedLot := AffectedLot{
				LotCode:   lot.Code,
				ItemType:  lot.ItemType,
				Qty:       lot.Qty,
				Unit:      lot.Unit,
				Markets:   markets,
				SubBuyers: subBuyers,
			}
			if product, err := t.repo.GetProduct(ctx, lot.ItemID); err == nil {
				affectedLot.ProductCode = product.Code
				affectedLot.ProductName = product.Name
			}
			affected = append(affected, affectedLot)
		}

		links, err := t.repo.ListLotLinksByParent(ctx, lot.ID)
		if err != nil {
			return nil, err
		}
		for _, link := range links {
			if visited[link.ChildLotID] {
				continue
			}
			child, err := t.repo.GetLot(ctx, link.ChildLotID)
			if err != nil {
				return nil, err
			}
			visited[link.ChildLotID] = true
			queue = append(queue, child)
		}
	}

	return affected, nil
}

// Backward returns material lots and suppliers reachable from productLotCode.
func (t *Traceability) Backward(ctx context.Context, productLotCode string) ([]RawMaterialSource, error) {
	start, err := t.repo.GetLotByCode(ctx, productLotCode)
	if err != nil {
		return nil, err
	}

	visited := map[int64]bool{start.ID: true}
	queue := []model.Lot{start}
	var sources []RawMaterialSource

	for len(queue) > 0 {
		lot := queue[0]
		queue = queue[1:]

		if lot.ItemType == "material" {
			source := RawMaterialSource{
				LotCode: lot.Code,
				Qty:     lot.Qty,
			}
			if lot.ReceivedAt != nil {
				source.ReceivedAt = *lot.ReceivedAt
			}
			if lot.ExpiryDate != nil {
				source.ExpiryDate = *lot.ExpiryDate
			}
			if lot.SupplierID != nil {
				supplier, err := t.repo.GetSupplier(ctx, *lot.SupplierID)
				if err != nil {
					return nil, err
				}
				source.SupplierCode = supplier.Code
				source.SupplierName = supplier.Name
			}
			if orders, err := t.repo.ListProductionOrderCodesByInputLot(ctx, lot.ID); err == nil {
				source.ProductionOrders = orders
			}
			sources = append(sources, source)
		}

		links, err := t.repo.ListLotLinksByChild(ctx, lot.ID)
		if err != nil {
			return nil, err
		}
		for _, link := range links {
			if visited[link.ParentLotID] {
				continue
			}
			parent, err := t.repo.GetLot(ctx, link.ParentLotID)
			if err != nil {
				return nil, err
			}
			visited[link.ParentLotID] = true
			queue = append(queue, parent)
		}
	}

	return sources, nil
}

func (t *Traceability) marketImpacts(ctx context.Context, lotID int64) ([]MarketImpact, error) {
	movements, err := t.repo.ListStockMovements(ctx, lotID)
	if err != nil {
		return nil, err
	}

	shipped := make(map[int64]float64)
	returned := make(map[int64]float64)
	seen := make(map[int64]bool)
	var order []int64
	for _, movement := range movements {
		if movement.RefType == nil || *movement.RefType != "market" || movement.RefID == nil {
			continue
		}
		if !seen[*movement.RefID] {
			seen[*movement.RefID] = true
			order = append(order, *movement.RefID)
		}
		switch movement.MovementType {
		case "shipment":
			shipped[*movement.RefID] += math.Abs(movement.Qty)
		case "return":
			returned[*movement.RefID] += math.Abs(movement.Qty)
		}
	}

	markets := make([]MarketImpact, 0, len(order))
	for _, marketID := range order {
		net := shipped[marketID] - returned[marketID]
		if net <= 0 {
			continue
		}
		market, err := t.repo.GetMarket(ctx, marketID)
		if err != nil {
			return nil, err
		}
		markets = append(markets, MarketImpact{
			MarketCode: market.Code,
			MarketName: market.Name,
			Contact:    market.Contact,
			Qty:        net,
		})
	}

	return markets, nil
}

func (t *Traceability) subBuyerImpacts(ctx context.Context, lotID int64) ([]SubBuyerImpact, error) {
	redistributions, err := t.repo.ListRedistributions(ctx, lotID)
	if err != nil {
		return nil, err
	}

	subBuyers := make([]SubBuyerImpact, 0, len(redistributions))
	for _, red := range redistributions {
		market, err := t.repo.GetMarket(ctx, red.MarketID)
		if err != nil {
			return nil, err
		}
		contact := ""
		if red.Contact != nil {
			contact = *red.Contact
		}
		subBuyers = append(subBuyers, SubBuyerImpact{
			MarketCode: market.Code,
			MarketName: market.Name,
			SubBuyer:   red.SubBuyer,
			Contact:    contact,
			Qty:        red.Qty,
		})
	}
	return subBuyers, nil
}
