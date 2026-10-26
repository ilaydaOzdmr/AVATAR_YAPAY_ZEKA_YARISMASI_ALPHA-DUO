package model

// Material is a raw material card (meat, spices, casing, ...).
type Material struct {
	ID                   int64   `json:"id"`
	Code                 string  `json:"code"`
	Name                 string  `json:"name"`
	Unit                 string  `json:"unit"`
	IsRaw                bool    `json:"is_raw"`
	DefaultPurchasePrice float64 `json:"default_purchase_price"`
}

// Product is a finished good or butchery part.
type Product struct {
	ID         int64   `json:"id"`
	Code       string  `json:"code"`
	Name       string  `json:"name"`
	Unit       string  `json:"unit"`
	ListPrice  float64 `json:"list_price"`
	Category   string  `json:"category"`
	IsFinished bool    `json:"is_finished"`
}

// Lot is a batch of a material or product.
type Lot struct {
	ID            int64   `json:"id"`
	Code          string  `json:"code"`
	ItemType      string  `json:"item_type"`
	ItemID        int64   `json:"item_id"`
	Qty           float64 `json:"qty"`
	Unit          string  `json:"unit"`
	SupplierID    *int64  `json:"supplier_id"`
	PurchasePrice float64 `json:"purchase_price"`
	UnitCost      float64 `json:"unit_cost"`
	ReceivedAt    *string `json:"received_at"`
	ExpiryDate    *string `json:"expiry_date"`
	Status        string  `json:"status"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

// LotLink is a parent-to-child relationship between two lots.
type LotLink struct {
	ID          int64   `json:"id"`
	ParentLotID int64   `json:"parent_lot_id"`
	ChildLotID  int64   `json:"child_lot_id"`
	Qty         float64 `json:"qty"`
	LinkType    string  `json:"link_type"`
}

// Market is a sales market that receives shipped product lots.
type Market struct {
	ID      int64  `json:"id"`
	Code    string `json:"code"`
	Name    string `json:"name"`
	Contact string `json:"contact"`
}

// Supplier is a vendor that supplies raw material lots.
type Supplier struct {
	ID      int64  `json:"id"`
	Code    string `json:"code"`
	Name    string `json:"name"`
	Contact string `json:"contact"`
}

// StockMovement is a signed lot-based stock movement. Positive qty is inflow,
// negative qty is outflow.
type StockMovement struct {
	ID           int64   `json:"id"`
	LotID        int64   `json:"lot_id"`
	MovementType string  `json:"movement_type"`
	Qty          float64 `json:"qty"`
	UnitCost     float64 `json:"unit_cost"`
	TotalCost    float64 `json:"total_cost"`
	RefType      *string `json:"ref_type"`
	RefID        *int64  `json:"ref_id"`
	OccurredAt   string  `json:"occurred_at"`
	CreatedAt    string  `json:"created_at"`
}

// ProductionOrder is a contract manufacturing order and its service cost.
type ProductionOrder struct {
	ID              int64   `json:"id"`
	Code            string  `json:"code"`
	Status          string  `json:"status"`
	StartedAt       *string `json:"started_at"`
	FinishedAt      *string `json:"finished_at"`
	Notes           *string `json:"notes"`
	ServiceName     *string `json:"service_name"`
	ServiceUnitCost float64 `json:"service_unit_cost"`
	ServiceQty      float64 `json:"service_qty"`
}

// ProductionOrderInput is a raw material lot consumed by a production order.
type ProductionOrderInput struct {
	ID                int64   `json:"id"`
	ProductionOrderID int64   `json:"production_order_id"`
	MaterialLotID     int64   `json:"material_lot_id"`
	Qty               float64 `json:"qty"`
}

// ProductionOrderOutput is a finished product lot produced by a production order.
type ProductionOrderOutput struct {
	ID                int64   `json:"id"`
	ProductionOrderID int64   `json:"production_order_id"`
	ProductLotID      int64   `json:"product_lot_id"`
	Qty               float64 `json:"qty"`
}

// ButcheryBatch is a whole-animal butchery batch and its purchase cost.
type ButcheryBatch struct {
	ID          int64   `json:"id"`
	Code        string  `json:"code"`
	AnimalType  string  `json:"animal_type"`
	InputWeight float64 `json:"input_weight"`
	UnitCost    float64 `json:"unit_cost"`
	TotalCost   float64 `json:"total_cost"`
	BatchDate   string  `json:"batch_date"`
	Status      string  `json:"status"`
}

// ButcheryItem is a part (or waste) produced from a butchery batch.
type ButcheryItem struct {
	ID           int64   `json:"id"`
	BatchID      int64   `json:"batch_id"`
	ProductID    int64   `json:"product_id"`
	ProductLotID *int64  `json:"product_lot_id"`
	Weight       float64 `json:"weight"`
	IsWaste      bool    `json:"is_waste"`
}

// RecallEvent is a recall triggered from a source (raw material) lot.
type RecallEvent struct {
	ID          int64  `json:"id"`
	Code        string `json:"code"`
	Reason      string `json:"reason"`
	SourceLotID int64  `json:"source_lot_id"`
	CreatedAt   string `json:"created_at"`
}

// RecallBlock marks a lot as blocked by a recall event.
type RecallBlock struct {
	ID            int64  `json:"id"`
	RecallEventID int64  `json:"recall_event_id"`
	LotID         int64  `json:"lot_id"`
	BlockedAt     string `json:"blocked_at"`
}

// Sale is a sale to a market.
type Sale struct {
	ID       int64   `json:"id"`
	Code     string  `json:"code"`
	MarketID int64   `json:"market_id"`
	SoldAt   string  `json:"sold_at"`
	Status   string  `json:"status"`
	ClientID *string `json:"client_id"`
}

// SaleLine is a lot-based line of a sale.
type SaleLine struct {
	ID              int64   `json:"id"`
	SaleID          int64   `json:"sale_id"`
	LotID           int64   `json:"lot_id"`
	Qty             float64 `json:"qty"`
	UnitPrice       float64 `json:"unit_price"`
	TotalPrice      float64 `json:"total_price"`
	StockMovementID *int64  `json:"stock_movement_id"`
}
