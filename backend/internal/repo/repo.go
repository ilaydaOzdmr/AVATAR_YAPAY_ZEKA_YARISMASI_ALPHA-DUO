package repo

import (
	"context"
	"database/sql"
	"time"

	"avatar/backend/internal/model"
)

// Repo is the data access layer. All queries are parameterized.
type Repo interface {
	CreateMaterial(ctx context.Context, m model.Material) (model.Material, error)
	ListMaterials(ctx context.Context) ([]model.Material, error)
	CreateProduct(ctx context.Context, p model.Product) (model.Product, error)
	ListProducts(ctx context.Context) ([]model.Product, error)
	CreateLot(ctx context.Context, l model.Lot) (model.Lot, error)
	ListLots(ctx context.Context, itemType string) ([]model.Lot, error)
	GetLotByCode(ctx context.Context, code string) (model.Lot, error)
	AddStockMovement(ctx context.Context, m model.StockMovement) (model.StockMovement, error)
	ListStockMovements(ctx context.Context, lotID int64) ([]model.StockMovement, error)
}

// NewMySQLRepo returns a MySQL-backed implementation of Repo.
func NewMySQLRepo(db *sql.DB) Repo { return &mysqlRepo{db: db} }

type mysqlRepo struct{ db *sql.DB }

func (r *mysqlRepo) CreateMaterial(ctx context.Context, m model.Material) (model.Material, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO materials (code, name, unit, is_raw, default_purchase_price) VALUES (?, ?, ?, ?, ?)`,
		m.Code, m.Name, m.Unit, m.IsRaw, m.DefaultPurchasePrice)
	if err != nil {
		return model.Material{}, err
	}
	m.ID, err = res.LastInsertId()
	return m, err
}

func (r *mysqlRepo) ListMaterials(ctx context.Context) ([]model.Material, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, code, name, unit, is_raw, default_purchase_price FROM materials ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Material
	for rows.Next() {
		var m model.Material
		if err := rows.Scan(&m.ID, &m.Code, &m.Name, &m.Unit, &m.IsRaw, &m.DefaultPurchasePrice); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *mysqlRepo) CreateProduct(ctx context.Context, p model.Product) (model.Product, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO products (code, name, unit, list_price, category, is_finished) VALUES (?, ?, ?, ?, ?, ?)`,
		p.Code, p.Name, p.Unit, p.ListPrice, p.Category, p.IsFinished)
	if err != nil {
		return model.Product{}, err
	}
	p.ID, err = res.LastInsertId()
	return p, err
}

func (r *mysqlRepo) ListProducts(ctx context.Context) ([]model.Product, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, code, name, unit, list_price, category, is_finished FROM products ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Product
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.ID, &p.Code, &p.Name, &p.Unit, &p.ListPrice, &p.Category, &p.IsFinished); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *mysqlRepo) CreateLot(ctx context.Context, l model.Lot) (model.Lot, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO lots (code, item_type, item_id, qty, unit, supplier_id, purchase_price, unit_cost, received_at, expiry_date, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		l.Code, l.ItemType, l.ItemID, l.Qty, l.Unit, nullInt64(l.SupplierID),
		l.PurchasePrice, l.UnitCost, nullString(l.ReceivedAt), nullString(l.ExpiryDate), l.Status)
	if err != nil {
		return model.Lot{}, err
	}
	l.ID, err = res.LastInsertId()
	return l, err
}

func (r *mysqlRepo) ListLots(ctx context.Context, itemType string) ([]model.Lot, error) {
	query := `SELECT id, code, item_type, item_id, qty, unit, supplier_id, purchase_price, unit_cost, received_at, expiry_date, status, created_at, updated_at FROM lots`
	var args []any
	if itemType != "" {
		query += ` WHERE item_type = ?`
		args = append(args, itemType)
	}
	query += ` ORDER BY id`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Lot
	for rows.Next() {
		l, err := scanLot(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (r *mysqlRepo) GetLotByCode(ctx context.Context, code string) (model.Lot, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, code, item_type, item_id, qty, unit, supplier_id, purchase_price, unit_cost, received_at, expiry_date, status, created_at, updated_at FROM lots WHERE code = ?`,
		code)
	return scanLot(row)
}

func (r *mysqlRepo) AddStockMovement(ctx context.Context, m model.StockMovement) (model.StockMovement, error) {
	if m.OccurredAt == "" {
		m.OccurredAt = time.Now().Format("2006-01-02 15:04:05")
	}
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO stock_movements (lot_id, movement_type, qty, unit_cost, total_cost, ref_type, ref_id, occurred_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		m.LotID, m.MovementType, m.Qty, m.UnitCost, m.TotalCost, nullString(m.RefType), nullInt64(m.RefID), m.OccurredAt)
	if err != nil {
		return model.StockMovement{}, err
	}
	m.ID, err = res.LastInsertId()
	return m, err
}

func (r *mysqlRepo) ListStockMovements(ctx context.Context, lotID int64) ([]model.StockMovement, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, lot_id, movement_type, qty, unit_cost, total_cost, ref_type, ref_id, occurred_at, created_at FROM stock_movements WHERE lot_id = ? ORDER BY occurred_at, id`,
		lotID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.StockMovement
	for rows.Next() {
		var m model.StockMovement
		var refType sql.NullString
		var refID sql.NullInt64
		if err := rows.Scan(&m.ID, &m.LotID, &m.MovementType, &m.Qty, &m.UnitCost, &m.TotalCost, &refType, &refID, &m.OccurredAt, &m.CreatedAt); err != nil {
			return nil, err
		}
		if refType.Valid {
			m.RefType = &refType.String
		}
		if refID.Valid {
			m.RefID = &refID.Int64
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

type scanner interface{ Scan(dest ...any) error }

func scanLot(s scanner) (model.Lot, error) {
	var l model.Lot
	var supplierID sql.NullInt64
	var receivedAt, expiryDate sql.NullString
	err := s.Scan(&l.ID, &l.Code, &l.ItemType, &l.ItemID, &l.Qty, &l.Unit, &supplierID,
		&l.PurchasePrice, &l.UnitCost, &receivedAt, &expiryDate, &l.Status, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		return model.Lot{}, err
	}
	if supplierID.Valid {
		l.SupplierID = &supplierID.Int64
	}
	if receivedAt.Valid {
		l.ReceivedAt = &receivedAt.String
	}
	if expiryDate.Valid {
		l.ExpiryDate = &expiryDate.String
	}
	return l, nil
}

func nullInt64(v *int64) any {
	if v == nil {
		return nil
	}
	return *v
}

func nullString(v *string) any {
	if v == nil {
		return nil
	}
	return *v
}
