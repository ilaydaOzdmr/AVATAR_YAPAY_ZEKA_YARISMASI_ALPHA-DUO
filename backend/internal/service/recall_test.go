package service

import (
	"context"
	"strings"
	"testing"

	"avatar/backend/internal/model"
	"avatar/backend/internal/repo"
)

// recallFakeRepo embeds repo.Repo so only the methods used by RecallService
// need real implementations.
type recallFakeRepo struct {
	repo.Repo
	lotsByCode      map[string]model.Lot
	lotsByID        map[int64]model.Lot
	linksByParent   map[int64][]model.LotLink
	movementsByLot  map[int64][]model.StockMovement
	redistributions map[int64][]model.Redistribution
	marketsByID     map[int64]model.Market
	productsByID    map[int64]model.Product
	createdEvent    model.RecallEvent
	createdBlocks   []model.RecallBlock
}

func (f *recallFakeRepo) GetLotByCode(_ context.Context, code string) (model.Lot, error) {
	return f.lotsByCode[code], nil
}

func (f *recallFakeRepo) GetLot(_ context.Context, id int64) (model.Lot, error) {
	return f.lotsByID[id], nil
}

func (f *recallFakeRepo) ListLotLinksByParent(_ context.Context, parentID int64) ([]model.LotLink, error) {
	return f.linksByParent[parentID], nil
}

func (f *recallFakeRepo) ListStockMovements(_ context.Context, lotID int64) ([]model.StockMovement, error) {
	return f.movementsByLot[lotID], nil
}

func (f *recallFakeRepo) ListRedistributions(_ context.Context, lotID int64) ([]model.Redistribution, error) {
	return f.redistributions[lotID], nil
}

func (f *recallFakeRepo) GetMarket(_ context.Context, id int64) (model.Market, error) {
	return f.marketsByID[id], nil
}

func (f *recallFakeRepo) GetProduct(_ context.Context, id int64) (model.Product, error) {
	return f.productsByID[id], nil
}

func (f *recallFakeRepo) CreateRecall(_ context.Context, event model.RecallEvent, blocks []model.RecallBlock) (model.RecallEvent, error) {
	event.ID = 7
	f.createdEvent = event
	f.createdBlocks = blocks
	return event, nil
}

func TestRecallCreateBlocksSourceAndAffectedLots(t *testing.T) {
	refTypeMarket := "market"
	marketID := int64(1)
	subContact := "0532 000 00 01"

	fake := &recallFakeRepo{
		lotsByCode: map[string]model.Lot{
			"MAT-1": {ID: 1, Code: "MAT-1", ItemType: "material", Qty: 20, Unit: "kg"},
			"PRD-1": {ID: 2, Code: "PRD-1", ItemType: "product", ItemID: 1, Qty: 8, Unit: "kg"},
		},
		lotsByID: map[int64]model.Lot{
			1: {ID: 1, Code: "MAT-1", ItemType: "material", Qty: 20, Unit: "kg"},
			2: {ID: 2, Code: "PRD-1", ItemType: "product", ItemID: 1, Qty: 8, Unit: "kg"},
		},
		linksByParent: map[int64][]model.LotLink{
			1: {{ID: 1, ParentLotID: 1, ChildLotID: 2, Qty: 20, LinkType: "production"}},
		},
		movementsByLot: map[int64][]model.StockMovement{
			2: {
				{ID: 1, LotID: 2, MovementType: "shipment", Qty: -10, RefType: &refTypeMarket, RefID: &marketID},
				{ID: 2, LotID: 2, MovementType: "return", Qty: 2, RefType: &refTypeMarket, RefID: &marketID},
			},
		},
		redistributions: map[int64][]model.Redistribution{
			2: {{ID: 1, MarketID: 1, LotID: 2, SubBuyer: "Kebapçı Ali", Contact: &subContact, Qty: 4}},
		},
		marketsByID: map[int64]model.Market{
			1: {ID: 1, Code: "MKT-A", Name: "Market A", Contact: "0212 111 11 01"},
		},
		productsByID: map[int64]model.Product{
			1: {ID: 1, Code: "PRD-001", Name: "Narin Sucuk"},
		},
	}

	result, err := NewRecallService(fake).Create(context.Background(), "MAT-1", "test")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if fake.createdEvent.SourceLotID != 1 {
		t.Fatalf("event source = %d, want 1", fake.createdEvent.SourceLotID)
	}
	blocked := map[int64]bool{}
	for _, block := range fake.createdBlocks {
		blocked[block.LotID] = true
	}
	if len(blocked) != 2 || !blocked[1] || !blocked[2] {
		t.Fatalf("blocked lots = %+v, want source 1 and product 2", fake.createdBlocks)
	}

	if len(result.AffectedLots) != 1 {
		t.Fatalf("affected lots = %d, want 1", len(result.AffectedLots))
	}
	lot := result.AffectedLots[0]
	if lot.Markets[0].Qty != 8 {
		t.Fatalf("net market qty = %v, want 8 (10 shipped - 2 returned)", lot.Markets[0].Qty)
	}
	if len(lot.SubBuyers) != 1 || lot.SubBuyers[0].Qty != 4 {
		t.Fatalf("sub buyers = %+v, want one 4 kg sub-buyer", lot.SubBuyers)
	}
	if !strings.Contains(result.Message, "ikinci kademe alıcı sayısı: 1") {
		t.Fatalf("message = %q, want second-level count", result.Message)
	}
}
