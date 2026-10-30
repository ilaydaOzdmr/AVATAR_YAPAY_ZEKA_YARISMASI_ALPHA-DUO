package repo

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"time"

	"avatar/backend/internal/model"
)

// ErrInsufficientStock is returned when an outflow would make a lot negative.
var ErrInsufficientStock = errors.New("insufficient stock")

// ErrLotBlocked is returned when a sale targets a lot blocked by a recall.
var ErrLotBlocked = errors.New("lot is blocked by a recall")

// ErrInvalidStockMovement is returned for a zero-quantity movement.
var ErrInvalidStockMovement = errors.New("stock movement quantity cannot be zero")

// ErrInvalidOrder is returned when a production or butchery document is incomplete.
var ErrInvalidOrder = errors.New("invalid order: inputs and outputs are required")

func roundMoney(v float64) float64 {
	return math.Round(v*100) / 100
}

func roundQty(v float64) float64 {
	return math.Round(v*1000) / 1000
}

// applyMovementTx inserts a stock movement and updates lots.qty in the same
// transaction. Positive qty is inflow, negative qty is outflow. When unitCost
// is zero, the lot's current unit cost is used.
func applyMovementTx(ctx context.Context, tx *sql.Tx, m model.StockMovement) (model.StockMovement, error) {
	if m.Qty == 0 {
		return model.StockMovement{}, ErrInvalidStockMovement
	}
	if m.OccurredAt == "" {
		m.OccurredAt = time.Now().Format("2006-01-02 15:04:05")
	}

	var currentQty, lotUnitCost float64
	if err := tx.QueryRowContext(ctx,
		`SELECT qty, unit_cost FROM lots WHERE id = ? FOR UPDATE`, m.LotID).
		Scan(&currentQty, &lotUnitCost); err != nil {
		return model.StockMovement{}, err
	}
	if m.Qty < 0 && currentQty+m.Qty < -0.0005 {
		return model.StockMovement{}, ErrInsufficientStock
	}
	if m.UnitCost == 0 {
		m.UnitCost = lotUnitCost
	}
	m.TotalCost = math.Round(math.Abs(m.Qty)*m.UnitCost*100) / 100

	res, err := tx.ExecContext(ctx,
		`INSERT INTO stock_movements (lot_id, movement_type, qty, unit_cost, total_cost, ref_type, ref_id, occurred_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		m.LotID, m.MovementType, m.Qty, m.UnitCost, m.TotalCost,
		nullString(m.RefType), nullInt64(m.RefID), m.OccurredAt)
	if err != nil {
		return model.StockMovement{}, err
	}
	m.ID, err = res.LastInsertId()
	if err != nil {
		return model.StockMovement{}, err
	}

	updated, err := tx.ExecContext(ctx,
		`UPDATE lots SET qty = qty + ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ? AND qty + ? >= 0`, m.Qty, m.LotID, m.Qty)
	if err != nil {
		return model.StockMovement{}, err
	}
	rows, err := updated.RowsAffected()
	if err != nil {
		return model.StockMovement{}, err
	}
	if rows == 0 {
		return model.StockMovement{}, ErrInsufficientStock
	}
	return m, nil
}
