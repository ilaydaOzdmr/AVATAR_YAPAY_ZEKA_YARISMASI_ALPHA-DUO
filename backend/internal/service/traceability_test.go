package service

import (
	"context"
	"testing"

	"avatar/backend/internal/model"
)

// fakeTraceRepository is an in-memory TraceRepository for traceability tests.
type fakeTraceRepository struct {
	lotsByCode           map[string]model.Lot
	lotsByID             map[int64]model.Lot
	linksByParent        map[int64][]model.LotLink
	linksByChild         map[int64][]model.LotLink
	movementsByLot       map[int64][]model.StockMovement
	redistributionsByLot map[int64][]model.Redistribution
	marketsByID          map[int64]model.Market
	suppliersByID        map[int64]model.Supplier
	productsByID         map[int64]model.Product
}

func (f *fakeTraceRepository) GetLotByCode(_ context.Context, code string) (model.Lot, error) {
	return f.lotsByCode[code], nil
}

func (f *fakeTraceRepository) GetLot(_ context.Context, id int64) (model.Lot, error) {
	return f.lotsByID[id], nil
}

func (f *fakeTraceRepository) ListLotLinksByParent(_ context.Context, parentID int64) ([]model.LotLink, error) {
	return f.linksByParent[parentID], nil
}

func (f *fakeTraceRepository) ListLotLinksByChild(_ context.Context, childID int64) ([]model.LotLink, error) {
	return f.linksByChild[childID], nil
}

func (f *fakeTraceRepository) ListStockMovements(_ context.Context, lotID int64) ([]model.StockMovement, error) {
	return f.movementsByLot[lotID], nil
}

func (f *fakeTraceRepository) ListRedistributions(_ context.Context, lotID int64) ([]model.Redistribution, error) {
	return f.redistributionsByLot[lotID], nil
}

func (f *fakeTraceRepository) GetMarket(_ context.Context, id int64) (model.Market, error) {
	return f.marketsByID[id], nil
}

func (f *fakeTraceRepository) GetSupplier(_ context.Context, id int64) (model.Supplier, error) {
	return f.suppliersByID[id], nil
}

func (f *fakeTraceRepository) GetProduct(_ context.Context, id int64) (model.Product, error) {
	return f.productsByID[id], nil
}

func (f *fakeTraceRepository) ListProductionOrderCodesByInputLot(_ context.Context, lotID int64) ([]string, error) {
	return nil, nil
}

func seededTraceRepository() *fakeTraceRepository {
	supplierID1 := int64(1)
	supplierID2 := int64(2)
	supplierID3 := int64(3)
	refTypeMarket := "market"
	marketID1 := int64(1)
	marketID2 := int64(2)
	marketID3 := int64(3)
	subContact := "0532 000 00 01"

	lots := []model.Lot{
		{ID: 1, Code: "ET-2026-001", ItemType: "material", Qty: 100, Unit: "kg", SupplierID: &supplierID1},
		{ID: 2, Code: "BAH-2026-001", ItemType: "material", Qty: 2, Unit: "kg", SupplierID: &supplierID2},
		{ID: 3, Code: "BAR-2026-001", ItemType: "material", Qty: 5, Unit: "kg", SupplierID: &supplierID3},
		{ID: 4, Code: "SUC-2026-045", ItemType: "product", ItemID: 1, Qty: 85, Unit: "paket"},
	}
	lotsByCode := make(map[string]model.Lot, len(lots))
	lotsByID := make(map[int64]model.Lot, len(lots))
	for _, lot := range lots {
		lotsByCode[lot.Code] = lot
		lotsByID[lot.ID] = lot
	}

	links := []model.LotLink{
		{ID: 1, ParentLotID: 1, ChildLotID: 4, Qty: 100, LinkType: "production"},
		{ID: 2, ParentLotID: 2, ChildLotID: 4, Qty: 2, LinkType: "production"},
		{ID: 3, ParentLotID: 3, ChildLotID: 4, Qty: 5, LinkType: "production"},
	}
	linksByParent := make(map[int64][]model.LotLink)
	linksByChild := make(map[int64][]model.LotLink)
	for _, link := range links {
		linksByParent[link.ParentLotID] = append(linksByParent[link.ParentLotID], link)
		linksByChild[link.ChildLotID] = append(linksByChild[link.ChildLotID], link)
	}

	return &fakeTraceRepository{
		lotsByCode:    lotsByCode,
		lotsByID:      lotsByID,
		linksByParent: linksByParent,
		linksByChild:  linksByChild,
		movementsByLot: map[int64][]model.StockMovement{
			4: {
				{ID: 1, LotID: 4, MovementType: "shipment", Qty: -30, RefType: &refTypeMarket, RefID: &marketID1},
				{ID: 2, LotID: 4, MovementType: "shipment", Qty: -25, RefType: &refTypeMarket, RefID: &marketID2},
				{ID: 3, LotID: 4, MovementType: "shipment", Qty: -30, RefType: &refTypeMarket, RefID: &marketID3},
				{ID: 4, LotID: 4, MovementType: "return", Qty: 2, RefType: &refTypeMarket, RefID: &marketID1},
			},
		},
		redistributionsByLot: map[int64][]model.Redistribution{
			4: {{ID: 1, MarketID: 1, LotID: 4, SubBuyer: "Kebapçı Ali", Contact: &subContact, Qty: 10}},
		},
		marketsByID: map[int64]model.Market{
			1: {ID: 1, Code: "MKT-A", Name: "Market A", Contact: "0212 111 11 01"},
			2: {ID: 2, Code: "MKT-B", Name: "Market B", Contact: "0212 111 11 02"},
			3: {ID: 3, Code: "MKT-C", Name: "Market C", Contact: "0212 111 11 03"},
		},
		suppliersByID: map[int64]model.Supplier{
			1: {ID: 1, Code: "SUP-001", Name: "Et Tedarikçisi", Contact: "0212 000 00 01"},
			2: {ID: 2, Code: "SUP-002", Name: "Baharat Tedarikçisi", Contact: "0212 000 00 02"},
			3: {ID: 3, Code: "SUP-003", Name: "Bağırsak Tedarikçisi", Contact: "0212 000 00 03"},
		},
		productsByID: map[int64]model.Product{
			1: {ID: 1, Code: "PRD-001", Name: "Narin Sucuk (1 kg)"},
		},
	}
}

func TestForwardET2026001(t *testing.T) {
	ctx := context.Background()
	res, err := NewTraceability(seededTraceRepository()).Forward(ctx, "ET-2026-001")
	if err != nil {
		t.Fatalf("Forward: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("len(result) = %d, want 1: %+v", len(res), res)
	}

	got := res[0]
	if got.LotCode != "SUC-2026-045" || got.ItemType != "product" || got.Qty != 85 || got.Unit != "paket" {
		t.Fatalf("affected lot = %+v, want SUC-2026-045/product/85/paket", got)
	}
	if got.ProductCode != "PRD-001" || got.ProductName != "Narin Sucuk (1 kg)" {
		t.Fatalf("product = %+v, want PRD-001/Narin Sucuk (1 kg)", got)
	}
	if len(got.Markets) != 3 {
		t.Fatalf("len(markets) = %d, want 3: %+v", len(got.Markets), got.Markets)
	}

	wantMarkets := []MarketImpact{
		{MarketCode: "MKT-A", MarketName: "Market A", Contact: "0212 111 11 01", Qty: 28},
		{MarketCode: "MKT-B", MarketName: "Market B", Contact: "0212 111 11 02", Qty: 25},
		{MarketCode: "MKT-C", MarketName: "Market C", Contact: "0212 111 11 03", Qty: 30},
	}
	for i, want := range wantMarkets {
		if got.Markets[i] != want {
			t.Fatalf("markets[%d] = %+v, want %+v", i, got.Markets[i], want)
		}
	}

	if len(got.SubBuyers) != 1 {
		t.Fatalf("len(sub_buyers) = %d, want 1: %+v", len(got.SubBuyers), got.SubBuyers)
	}
	wantSub := SubBuyerImpact{
		MarketCode: "MKT-A",
		MarketName: "Market A",
		SubBuyer:   "Kebapçı Ali",
		Contact:    "0532 000 00 01",
		Qty:        10,
	}
	if got.SubBuyers[0] != wantSub {
		t.Fatalf("sub_buyers[0] = %+v, want %+v", got.SubBuyers[0], wantSub)
	}
}

func TestBackwardSUC2026045(t *testing.T) {
	ctx := context.Background()
	res, err := NewTraceability(seededTraceRepository()).Backward(ctx, "SUC-2026-045")
	if err != nil {
		t.Fatalf("Backward: %v", err)
	}
	if len(res) != 3 {
		t.Fatalf("len(result) = %d, want 3: %+v", len(res), res)
	}

	byLotCode := make(map[string]RawMaterialSource, len(res))
	for _, source := range res {
		byLotCode[source.LotCode] = source
	}
	want := []RawMaterialSource{
		{LotCode: "ET-2026-001", SupplierCode: "SUP-001", SupplierName: "Et Tedarikçisi", Qty: 100},
		{LotCode: "BAH-2026-001", SupplierCode: "SUP-002", SupplierName: "Baharat Tedarikçisi", Qty: 2},
		{LotCode: "BAR-2026-001", SupplierCode: "SUP-003", SupplierName: "Bağırsak Tedarikçisi", Qty: 5},
	}
	for _, wantSource := range want {
		got := byLotCode[wantSource.LotCode]
		if got.SupplierCode != wantSource.SupplierCode || got.SupplierName != wantSource.SupplierName || got.Qty != wantSource.Qty {
			t.Fatalf("source %q = %+v, want %+v", wantSource.LotCode, got, wantSource)
		}
	}
}

func TestBackwardIncludesSupplierlessMaterial(t *testing.T) {
	ctx := context.Background()
	productLot := model.Lot{ID: 10, Code: "PRD-X", ItemType: "product", Qty: 1, Unit: "pcs"}
	materialLot := model.Lot{ID: 11, Code: "MAT-X", ItemType: "material", Qty: 7, Unit: "kg"} // no supplier
	fake := &fakeTraceRepository{
		lotsByCode: map[string]model.Lot{"PRD-X": productLot, "MAT-X": materialLot},
		lotsByID:   map[int64]model.Lot{10: productLot, 11: materialLot},
		linksByChild: map[int64][]model.LotLink{
			10: {{ID: 1, ParentLotID: 11, ChildLotID: 10, LinkType: "production"}},
		},
		linksByParent:  map[int64][]model.LotLink{},
		movementsByLot: map[int64][]model.StockMovement{},
		marketsByID:    map[int64]model.Market{},
		suppliersByID:  map[int64]model.Supplier{},
	}

	res, err := NewTraceability(fake).Backward(ctx, "PRD-X")
	if err != nil {
		t.Fatalf("Backward: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("len(result) = %d, want 1", len(res))
	}
	got := res[0]
	if got.LotCode != "MAT-X" || got.SupplierCode != "" || got.SupplierName != "" || got.Qty != 7 {
		t.Fatalf("source = %+v, want MAT-X with empty supplier", got)
	}
}
