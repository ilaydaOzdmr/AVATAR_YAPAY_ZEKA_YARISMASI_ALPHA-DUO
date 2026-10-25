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
	LotCode  string         `json:"lot_code"`
	ItemType string         `json:"item_type"`
	Qty      float64        `json:"qty"`
	Unit     string         `json:"unit"`
	Markets  []MarketImpact `json:"markets"`
}

// RawMaterialSource is a material lot and its supplier behind a product lot.
type RawMaterialSource struct {
	LotCode      string  `json:"lot_code"`
	SupplierCode string  `json:"supplier_code"`
	SupplierName string  `json:"supplier_name"`
	Qty          float64 `json:"qty"`
}

// TraceRepository is the data access contract required for traceability.
type TraceRepository interface {
	GetLotByCode(ctx context.Context, code string) (model.Lot, error)
	GetLot(ctx context.Context, id int64) (model.Lot, error)
	ListLotLinksByParent(ctx context.Context, parentID int64) ([]model.LotLink, error)
	ListLotLinksByChild(ctx context.Context, childID int64) ([]model.LotLink, error)
	ListStockMovements(ctx context.Context, lotID int64) ([]model.StockMovement, error)
	GetMarket(ctx context.Context, id int64) (model.Market, error)
	GetSupplier(ctx context.Context, id int64) (model.Supplier, error)
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
			affected = append(affected, AffectedLot{
				LotCode:  lot.Code,
				ItemType: lot.ItemType,
				Qty:      lot.Qty,
				Unit:     lot.Unit,
				Markets:  markets,
			})
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
			if lot.SupplierID != nil {
				supplier, err := t.repo.GetSupplier(ctx, *lot.SupplierID)
				if err != nil {
					return nil, err
				}
				source.SupplierCode = supplier.Code
				source.SupplierName = supplier.Name
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

	markets := make([]MarketImpact, 0)
	for _, movement := range movements {
		if movement.MovementType != "shipment" ||
			movement.RefType == nil || *movement.RefType != "market" ||
			movement.RefID == nil {
			continue
		}

		market, err := t.repo.GetMarket(ctx, *movement.RefID)
		if err != nil {
			return nil, err
		}
		markets = append(markets, MarketImpact{
			MarketCode: market.Code,
			MarketName: market.Name,
			Contact:    market.Contact,
			Qty:        math.Abs(movement.Qty),
		})
	}

	return markets, nil
}
