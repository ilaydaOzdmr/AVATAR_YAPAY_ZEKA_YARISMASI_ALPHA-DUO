package service

import (
	"avatar/backend/internal/model"
	"avatar/backend/internal/repo"
	"context"
	"errors"
	"testing"
)

var errQualityData = errors.New("quality fixture read failed")

type qualityRepo struct {
	repo.Repo
	fail string
}

func intRef(value int64) *int64 { return &value }
func (r qualityRepo) problem(method string) error {
	if r.fail == method {
		return errQualityData
	}
	return nil
}

func (r qualityRepo) ListMaterials(ctx context.Context) ([]model.Material, error) {
	return []model.Material{{ID: 1, Code: "MAT", Name: "Meat"}}, r.problem("ListMaterials")
}

func (r qualityRepo) ListLots(ctx context.Context, itemType string) ([]model.Lot, error) {
	return []model.Lot{{ID: 1, Code: "RAW", ItemType: "material", ItemID: 1, SupplierID: intRef(1), Qty: 100, UnitCost: 1}, {ID: 3, Code: "FIN", ItemType: "product", Qty: 50}}, r.problem("ListLots")
}

func (r qualityRepo) GetSupplier(ctx context.Context, id int64) (model.Supplier, error) {
	return model.Supplier{Code: "SUP"}, r.problem("GetSupplier")
}

func (r qualityRepo) ListProductionOrders(ctx context.Context) ([]model.ProductionOrder, error) {
	return []model.ProductionOrder{{ID: 1, Code: "ORDER-1"}, {ID: 2, Code: "ORDER-2"}}, r.problem("ListProductionOrders")
}

func (r qualityRepo) GetProductionOrder(ctx context.Context, id int64) (model.ProductionOrder, error) {
	return model.ProductionOrder{ID: id, Code: "ORDER", ExpectedQty: 55}, r.problem("GetProductionOrder")
}

func (r qualityRepo) ListProductionOrderInputs(ctx context.Context, id int64) ([]model.ProductionOrderInput, error) {
	return []model.ProductionOrderInput{{MaterialLotID: id, Qty: 100}}, r.problem("ListProductionOrderInputs")
}

func (r qualityRepo) ListProductionOrderOutputs(ctx context.Context, id int64) ([]model.ProductionOrderOutput, error) {
	return []model.ProductionOrderOutput{{ProductLotID: 3, Qty: 50}}, r.problem("ListProductionOrderOutputs")
}

func (r qualityRepo) GetProduct(ctx context.Context, id int64) (model.Product, error) {
	return model.Product{ListPrice: 25}, r.problem("GetProduct")
}

func (r qualityRepo) ListButcheryBatches(ctx context.Context) ([]model.ButcheryBatch, error) {
	return []model.ButcheryBatch{{ID: 1, Code: "BATCH", InputWeight: 100, TotalCost: 1200}}, r.problem("ListButcheryBatches")
}

func (r qualityRepo) ListButcheryItems(ctx context.Context, id int64) ([]model.ButcheryItem, error) {
	return []model.ButcheryItem{{Weight: 60}, {Weight: 10, IsWaste: true}}, r.problem("ListButcheryItems")
}

func (r qualityRepo) ListAllStockMovements(ctx context.Context) ([]model.StockMovement, error) {
	return []model.StockMovement{{LotID: 3, MovementType: "sale", Qty: -1}}, r.problem("ListAllStockMovements")
}

func (r qualityRepo) ListStockMovements(ctx context.Context, id int64) ([]model.StockMovement, error) {
	return []model.StockMovement{{LotID: id, Qty: -1}}, r.problem("ListStockMovements")
}

func (r qualityRepo) ListReadAnomalyKeys(ctx context.Context) ([]string, error) {
	return nil, r.problem("ListReadAnomalyKeys")
}

func (r qualityRepo) MarkAnomalyRead(ctx context.Context, key string) error {
	return r.problem("MarkAnomalyRead")
}

func (r qualityRepo) ListMarkets(ctx context.Context) ([]model.Market, error) {
	return []model.Market{{ID: 1, Code: "MKT"}}, r.problem("ListMarkets")
}

func (r qualityRepo) ListSales(ctx context.Context) ([]model.Sale, error) {
	return []model.Sale{{ID: 1, Code: "SALE", MarketID: 1}}, r.problem("ListSales")
}

func (r qualityRepo) ListSaleLines(ctx context.Context, id int64) ([]model.SaleLine, error) {
	return []model.SaleLine{{Qty: 2, TotalPrice: 10}, {Qty: 3, TotalPrice: 20}}, r.problem("ListSaleLines")
}

func (r qualityRepo) ListRecallEvents(ctx context.Context) ([]model.RecallEvent, error) {
	return []model.RecallEvent{{ID: 1, Code: "RECALL"}}, r.problem("ListRecallEvents")
}

func (r qualityRepo) ListRecallBlocks(ctx context.Context, id int64) ([]model.RecallBlock, error) {
	return []model.RecallBlock{{LotID: 1}, {LotID: 999}}, r.problem("ListRecallBlocks")
}

func (r qualityRepo) GetLot(ctx context.Context, id int64) (model.Lot, error) {
	cost := 1.0
	if id == 2 {
		cost = 10
	}
	return model.Lot{ID: id, UnitCost: cost, ItemID: 1}, r.problem("GetLot")
}

func TestQualitySummaryPreservesBusinessTotals(t *testing.T) {
	summary, err := NewReportService(qualityRepo{}).Summary(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(summary.Purchases) != 1 || summary.Purchases[0].TotalCost != 100 {
		t.Fatalf("purchases: %+v", summary.Purchases)
	}
	if len(summary.Production) != 2 || summary.Production[0].EKP != 2 || summary.Production[1].EKP != 20 {
		t.Fatalf("production: %+v", summary.Production)
	}
	if summary.Butchery[0].PartsWeight != 60 || summary.Butchery[0].FireWeight != 30 || summary.Butchery[0].PartCount != 1 {
		t.Fatalf("butchery: %+v", summary.Butchery)
	}
	if summary.Sales[0].Qty != 5 || summary.Sales[0].TotalPrice != 30 || summary.Sales[0].MarketCode != "MKT" {
		t.Fatalf("sales: %+v", summary.Sales)
	}
	if summary.Recalls[0].BlockedLotCnt != 1 || summary.Recalls[0].BlockedLots[0] != "RAW" {
		t.Fatalf("recalls: %+v", summary.Recalls)
	}
	if summary.LotMovements[0].LotCode != "FIN" {
		t.Fatal(summary.LotMovements)
	}
}

func TestQualitySummaryStopsAtDataErrors(t *testing.T) {
	methods := []string{"ListMaterials", "ListLots", "ListProductionOrders", "ListProductionOrderInputs", "GetProductionOrder", "GetLot", "GetProduct", "ListProductionOrderOutputs", "ListButcheryBatches", "ListButcheryItems", "ListAllStockMovements", "ListMarkets", "ListSales", "ListSaleLines", "ListRecallEvents", "ListRecallBlocks"}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			_, err := NewReportService(qualityRepo{fail: method}).Summary(context.Background())
			if !errors.Is(err, errQualityData) {
				t.Fatalf("expected data error, got %v", err)
			}
		})
	}
	summary, err := NewReportService(qualityRepo{fail: "GetSupplier"}).Summary(context.Background())
	if err != nil || summary.Purchases[0].SupplierCode != "" {
		t.Fatal("optional supplier enrichment must not lose purchase rows")
	}
}

func TestQualityAnomalyRulesRetainCostsYieldAndStockChecks(t *testing.T) {
	anomalies, err := NewAnomalyService(qualityRepo{}).Detect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	counts := map[string]int{}
	for _, anomaly := range anomalies {
		counts[anomaly.Type]++
	}
	if counts["fire"] != 1 || counts["yield"] != 2 || counts["cost"] != 1 || counts["stock"] != 2 {
		t.Fatalf("unexpected alerts: %+v", anomalies)
	}
	for _, method := range []string{"ListButcheryBatches", "ListButcheryItems", "ListProductionOrders", "ListLots"} {
		_, err := NewAnomalyService(qualityRepo{fail: method}).Detect(context.Background())
		if !errors.Is(err, errQualityData) {
			t.Errorf("%s: %v", method, err)
		}
	}
	anomalies, err = NewAnomalyService(qualityRepo{fail: "ListStockMovements"}).Detect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, anomaly := range anomalies {
		if anomaly.Type == "stock" {
			t.Fatal("failed balance reads must not become false stock alerts")
		}
	}
	if len(costAnomalies(nil, 0, 0)) != 0 {
		t.Fatal("empty production must not generate a cost alert")
	}
}
