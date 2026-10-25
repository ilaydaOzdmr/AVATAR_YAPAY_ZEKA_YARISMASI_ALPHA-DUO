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
