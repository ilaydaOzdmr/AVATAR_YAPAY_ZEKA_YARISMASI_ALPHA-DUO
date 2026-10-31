package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"avatar/backend/internal/model"
)

// Repo is the data access layer. All queries are parameterized.
type Repo interface {
	CreateMaterial(ctx context.Context, m model.Material) (model.Material, error)
	ListMaterials(ctx context.Context) ([]model.Material, error)
	GetMaterial(ctx context.Context, id int64) (model.Material, error)
	CreateProduct(ctx context.Context, p model.Product) (model.Product, error)
	ListProducts(ctx context.Context) ([]model.Product, error)
	UpdateProductShelfLife(ctx context.Context, id int64, days int) error
	CreateLot(ctx context.Context, l model.Lot) (model.Lot, error)
	ListLots(ctx context.Context, itemType string) ([]model.Lot, error)
	GetLotByCode(ctx context.Context, code string) (model.Lot, error)
	ListLotLinksByParent(ctx context.Context, parentID int64) ([]model.LotLink, error)
	ListLotLinksByChild(ctx context.Context, childID int64) ([]model.LotLink, error)
	AddStockMovement(ctx context.Context, m model.StockMovement) (model.StockMovement, error)
	ListStockMovements(ctx context.Context, lotID int64) ([]model.StockMovement, error)
	ListAllStockMovements(ctx context.Context) ([]model.StockMovement, error)
	CreateProductionOrder(ctx context.Context, po model.ProductionOrder, inputs []model.ProductionOrderInput, outputs []model.ProductionOrderOutput) (model.ProductionOrder, error)
	CompleteProductionOrder(ctx context.Context, id int64, finishedAt *string) (model.ProductionOrder, error)
	RejectProductionOrder(ctx context.Context, id int64, reason *string) (model.ProductionOrder, error)
	GetProductionOrder(ctx context.Context, id int64) (model.ProductionOrder, error)
	ListProductionOrderInputs(ctx context.Context, poID int64) ([]model.ProductionOrderInput, error)
	ListProductionOrderCodesByInputLot(ctx context.Context, lotID int64) ([]string, error)
	ListProductionOrderOutputs(ctx context.Context, poID int64) ([]model.ProductionOrderOutput, error)
	GetLot(ctx context.Context, id int64) (model.Lot, error)
	GetProduct(ctx context.Context, id int64) (model.Product, error)
	GetMarket(ctx context.Context, id int64) (model.Market, error)
	ListMarkets(ctx context.Context) ([]model.Market, error)
	GetSupplier(ctx context.Context, id int64) (model.Supplier, error)
	CreateButcheryBatch(ctx context.Context, batch model.ButcheryBatch, items []model.ButcheryItem) (model.ButcheryBatch, error)
	GetButcheryBatch(ctx context.Context, id int64) (model.ButcheryBatch, error)
	ListButcheryItems(ctx context.Context, batchID int64) ([]model.ButcheryItem, error)
	ListButcheryBatches(ctx context.Context) ([]model.ButcheryBatch, error)
	ListProductionOrders(ctx context.Context) ([]model.ProductionOrder, error)
	CreateRecall(ctx context.Context, event model.RecallEvent, blocks []model.RecallBlock) (model.RecallEvent, error)
	CreateRecallEvent(ctx context.Context, event model.RecallEvent) (model.RecallEvent, error)
	GetRecallEvent(ctx context.Context, id int64) (model.RecallEvent, error)
	CreateRecallBlock(ctx context.Context, block model.RecallBlock) (model.RecallBlock, error)
	ListRecallBlocks(ctx context.Context, recallEventID int64) ([]model.RecallBlock, error)
	IsLotBlocked(ctx context.Context, lotID int64) (bool, error)
	ListBlockedLotCodes(ctx context.Context) ([]string, error)
	CreateSale(ctx context.Context, sale model.Sale, lines []model.SaleLine) (model.Sale, error)
	ListSales(ctx context.Context) ([]model.Sale, error)
	ListSaleLines(ctx context.Context, saleID int64) ([]model.SaleLine, error)
	ListRecallEvents(ctx context.Context) ([]model.RecallEvent, error)
	CreateRedistribution(ctx context.Context, r model.Redistribution) (model.Redistribution, error)
	ListRedistributions(ctx context.Context, lotID int64) ([]model.Redistribution, error)
}

// NewMySQLRepo returns a MySQL-backed implementation of Repo.
func NewMySQLRepo(db *sql.DB) Repo { return &mysqlRepo{db: db} }

type mysqlRepo struct{ db *sql.DB }

func (r *mysqlRepo) CreateMaterial(ctx context.Context, m model.Material) (model.Material, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO materials (code, name, unit, is_raw, default_purchase_price, shelf_life_days)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		m.Code, m.Name, m.Unit, m.IsRaw, m.DefaultPurchasePrice, m.ShelfLifeDays)
	if err != nil {
		return model.Material{}, err
	}
	m.ID, err = res.LastInsertId()
	if err != nil {
		return model.Material{}, err
	}
	if err := r.db.QueryRowContext(ctx,
		`SELECT created_at FROM materials WHERE id = ?`, m.ID).Scan(&m.CreatedAt); err != nil {
		return model.Material{}, err
	}
	return m, nil
}

func (r *mysqlRepo) ListMaterials(ctx context.Context) ([]model.Material, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, code, name, unit, is_raw, default_purchase_price, shelf_life_days, created_at
		 FROM materials ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Material
	for rows.Next() {
		var m model.Material
		if err := rows.Scan(&m.ID, &m.Code, &m.Name, &m.Unit, &m.IsRaw,
			&m.DefaultPurchasePrice, &m.ShelfLifeDays, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *mysqlRepo) GetMaterial(ctx context.Context, id int64) (model.Material, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, code, name, unit, is_raw, default_purchase_price, shelf_life_days, created_at
		 FROM materials WHERE id = ?`, id)
	var m model.Material
	if err := row.Scan(&m.ID, &m.Code, &m.Name, &m.Unit, &m.IsRaw,
		&m.DefaultPurchasePrice, &m.ShelfLifeDays, &m.CreatedAt); err != nil {
		return model.Material{}, err
	}
	return m, nil
}

func (r *mysqlRepo) CreateProduct(ctx context.Context, p model.Product) (model.Product, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO products (code, name, unit, list_price, category, is_finished, shelf_life_days)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		p.Code, p.Name, p.Unit, p.ListPrice, p.Category, p.IsFinished, p.ShelfLifeDays)
	if err != nil {
		return model.Product{}, err
	}
	p.ID, err = res.LastInsertId()
	if err != nil {
		return model.Product{}, err
	}
	if err := r.db.QueryRowContext(ctx,
		`SELECT created_at FROM products WHERE id = ?`, p.ID).Scan(&p.CreatedAt); err != nil {
		return model.Product{}, err
	}
	return p, nil
}

func (r *mysqlRepo) ListProducts(ctx context.Context) ([]model.Product, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, code, name, unit, list_price, category, is_finished, shelf_life_days, created_at
		 FROM products ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Product
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.ID, &p.Code, &p.Name, &p.Unit, &p.ListPrice,
			&p.Category, &p.IsFinished, &p.ShelfLifeDays, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *mysqlRepo) UpdateProductShelfLife(ctx context.Context, id int64, days int) error {
	if days < 0 {
		return ErrInvalidLot
	}
	res, err := r.db.ExecContext(ctx,
		`UPDATE products SET shelf_life_days = ? WHERE id = ?`, days, id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *mysqlRepo) CreateLot(ctx context.Context, l model.Lot) (model.Lot, error) {
	if l.Code == "" || l.ItemType == "" || l.Qty < 0 {
		return model.Lot{}, ErrInvalidLot
	}
	if l.Status == "" {
		l.Status = "active"
	}
	if l.ReceivedAt == nil || *l.ReceivedAt == "" {
		now := time.Now().Format("2006-01-02 15:04:05")
		l.ReceivedAt = &now
	}

	var shelfLifeDays int
	switch l.ItemType {
	case "product":
		product, err := r.GetProduct(ctx, l.ItemID)
		if err != nil {
			return model.Lot{}, err
		}
		shelfLifeDays = product.ShelfLifeDays
		if l.Unit == "" {
			l.Unit = product.Unit
		}
		if l.UnitCost == 0 {
			l.UnitCost = product.ListPrice
		}
	case "material":
		material, err := r.GetMaterial(ctx, l.ItemID)
		if err != nil {
			return model.Lot{}, err
		}
		shelfLifeDays = material.ShelfLifeDays
		if l.Unit == "" {
			l.Unit = material.Unit
		}
		if l.UnitCost == 0 {
			l.UnitCost = material.DefaultPurchasePrice
		}
		if l.PurchasePrice == 0 {
			l.PurchasePrice = material.DefaultPurchasePrice
		}
	default:
		return model.Lot{}, ErrInvalidLot
	}
	if shelfLifeDays > 0 && (l.ExpiryDate == nil || *l.ExpiryDate == "") {
		expiry := addDays(*l.ReceivedAt, shelfLifeDays)
		l.ExpiryDate = &expiry
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return model.Lot{}, err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`INSERT INTO lots (code, item_type, item_id, qty, unit, supplier_id, purchase_price, unit_cost, received_at, expiry_date, status)
		 VALUES (?, ?, ?, 0, ?, ?, ?, ?, ?, ?, ?)`,
		l.Code, l.ItemType, l.ItemID, l.Unit, nullInt64(l.SupplierID),
		l.PurchasePrice, l.UnitCost, nullString(l.ReceivedAt), nullString(l.ExpiryDate), l.Status)
	if err != nil {
		return model.Lot{}, err
	}
	l.ID, err = res.LastInsertId()
	if err != nil {
		return model.Lot{}, err
	}

	if l.Qty > 0 {
		refType := "lot"
		if _, err := applyMovementTx(ctx, tx, model.StockMovement{
			LotID:        l.ID,
			MovementType: "inbound",
			Qty:          l.Qty,
			UnitCost:     l.UnitCost,
			RefType:      &refType,
			RefID:        &l.ID,
			OccurredAt:   *l.ReceivedAt,
		}); err != nil {
			return model.Lot{}, err
		}
	}
	if err := tx.QueryRowContext(ctx,
		`SELECT created_at FROM lots WHERE id = ?`, l.ID).Scan(&l.CreatedAt); err != nil {
		return model.Lot{}, err
	}
	if err := tx.Commit(); err != nil {
		return model.Lot{}, err
	}
	return l, nil
}

func addDays(dateTime string, days int) string {
	layouts := []string{"2006-01-02 15:04:05", "2006-01-02T15:04:05", "2006-01-02"}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, dateTime); err == nil {
			return parsed.AddDate(0, 0, days).Format("2006-01-02")
		}
	}
	return time.Now().AddDate(0, 0, days).Format("2006-01-02")
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

func (r *mysqlRepo) ListLotLinksByParent(ctx context.Context, parentID int64) ([]model.LotLink, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, parent_lot_id, child_lot_id, qty, link_type FROM lot_links WHERE parent_lot_id = ? ORDER BY id`,
		parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.LotLink
	for rows.Next() {
		var link model.LotLink
		if err := rows.Scan(&link.ID, &link.ParentLotID, &link.ChildLotID, &link.Qty, &link.LinkType); err != nil {
			return nil, err
		}
		out = append(out, link)
	}
	return out, rows.Err()
}

func (r *mysqlRepo) ListLotLinksByChild(ctx context.Context, childID int64) ([]model.LotLink, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, parent_lot_id, child_lot_id, qty, link_type FROM lot_links WHERE child_lot_id = ? ORDER BY id`,
		childID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.LotLink
	for rows.Next() {
		var link model.LotLink
		if err := rows.Scan(&link.ID, &link.ParentLotID, &link.ChildLotID, &link.Qty, &link.LinkType); err != nil {
			return nil, err
		}
		out = append(out, link)
	}
	return out, rows.Err()
}

func (r *mysqlRepo) AddStockMovement(ctx context.Context, m model.StockMovement) (model.StockMovement, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return model.StockMovement{}, err
	}
	defer tx.Rollback()

	moved, err := applyMovementTx(ctx, tx, m)
	if err != nil {
		return model.StockMovement{}, err
	}
	if err := tx.Commit(); err != nil {
		return model.StockMovement{}, err
	}
	return moved, nil
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

func (r *mysqlRepo) ListAllStockMovements(ctx context.Context) ([]model.StockMovement, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, lot_id, movement_type, qty, unit_cost, total_cost, ref_type, ref_id, occurred_at, created_at
		 FROM stock_movements ORDER BY occurred_at, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.StockMovement
	for rows.Next() {
		var m model.StockMovement
		var refType sql.NullString
		var refID sql.NullInt64
		if err := rows.Scan(&m.ID, &m.LotID, &m.MovementType, &m.Qty, &m.UnitCost, &m.TotalCost,
			&refType, &refID, &m.OccurredAt, &m.CreatedAt); err != nil {
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

func (r *mysqlRepo) CreateProductionOrder(ctx context.Context, po model.ProductionOrder, inputs []model.ProductionOrderInput, outputs []model.ProductionOrderOutput) (model.ProductionOrder, error) {
	if po.Status == "" {
		po.Status = "draft"
	}
	if po.Status == "completed" && (len(inputs) == 0 || len(outputs) == 0) {
		return model.ProductionOrder{}, ErrInvalidOrder
	}
	if (len(inputs) == 0) != (len(outputs) == 0) {
		return model.ProductionOrder{}, ErrInvalidOrder
	}
	for _, input := range inputs {
		if input.Qty <= 0 {
			return model.ProductionOrder{}, ErrInvalidOrder
		}
	}
	for _, output := range outputs {
		if output.Qty <= 0 {
			return model.ProductionOrder{}, ErrInvalidOrder
		}
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return model.ProductionOrder{}, err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`INSERT INTO production_orders (code, status, started_at, finished_at, notes, service_name, service_unit_cost, service_qty, expected_qty)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		po.Code, po.Status, nullString(po.StartedAt), nullString(po.FinishedAt),
		nullString(po.Notes), nullString(po.ServiceName), po.ServiceUnitCost, po.ServiceQty, po.ExpectedQty)
	if err != nil {
		return model.ProductionOrder{}, err
	}
	po.ID, err = res.LastInsertId()
	if err != nil {
		return model.ProductionOrder{}, err
	}

	for _, input := range inputs {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO production_order_inputs (production_order_id, material_lot_id, qty) VALUES (?, ?, ?)`,
			po.ID, input.MaterialLotID, input.Qty); err != nil {
			return model.ProductionOrder{}, err
		}
	}
	for _, output := range outputs {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO production_order_outputs (production_order_id, product_lot_id, qty) VALUES (?, ?, ?)`,
			po.ID, output.ProductLotID, output.Qty); err != nil {
			return model.ProductionOrder{}, err
		}
	}

	// A draft is only a plan: stock and lot links are applied when the order is
	// completed, so an in-progress order never consumes raw material twice.
	if po.Status == "completed" {
		occurredAt := time.Now().Format("2006-01-02 15:04:05")
		if po.FinishedAt != nil && *po.FinishedAt != "" {
			occurredAt = *po.FinishedAt
		}
		if err := applyProductionOrderTx(ctx, tx, po, inputs, outputs, occurredAt); err != nil {
			return model.ProductionOrder{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return model.ProductionOrder{}, err
	}
	return po, nil
}

// applyProductionOrderTx consumes inputs, stocks outputs and writes runtime lot
// links. It is shared by create-completed and the explicit complete transition.
func applyProductionOrderTx(ctx context.Context, tx *sql.Tx, po model.ProductionOrder, inputs []model.ProductionOrderInput, outputs []model.ProductionOrderOutput, occurredAt string) error {
	refType := "production_order"

	var inputCost float64
	for _, input := range inputs {
		var lotUnitCost float64
		if err := tx.QueryRowContext(ctx,
			`SELECT unit_cost FROM lots WHERE id = ? FOR UPDATE`, input.MaterialLotID).
			Scan(&lotUnitCost); err != nil {
			return err
		}
		inputCost += input.Qty * lotUnitCost
	}

	var totalOutputQty float64
	for _, output := range outputs {
		totalOutputQty += output.Qty
	}
	if totalOutputQty <= 0 {
		return ErrInvalidOrder
	}
	outputUnitCost := roundMoney((inputCost + po.ServiceUnitCost*po.ServiceQty) / totalOutputQty)

	for _, output := range outputs {
		var shelfLifeDays int
		if err := tx.QueryRowContext(ctx,
			`SELECT p.shelf_life_days
			 FROM lots l JOIN products p ON p.id = l.item_id
			 WHERE l.id = ?`, output.ProductLotID).Scan(&shelfLifeDays); err != nil {
			return err
		}
		if shelfLifeDays > 0 {
			expiryDate := addDays(occurredAt, shelfLifeDays)
			if _, err := tx.ExecContext(ctx,
				`UPDATE lots SET unit_cost = ?, received_at = COALESCE(received_at, ?),
				 expiry_date = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
				outputUnitCost, occurredAt, expiryDate, output.ProductLotID); err != nil {
				return err
			}
		} else if _, err := tx.ExecContext(ctx,
			`UPDATE lots SET unit_cost = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
			outputUnitCost, output.ProductLotID); err != nil {
			return err
		}
		if _, err := applyMovementTx(ctx, tx, model.StockMovement{
			LotID:        output.ProductLotID,
			MovementType: "production_output",
			Qty:          output.Qty,
			UnitCost:     outputUnitCost,
			RefType:      &refType,
			RefID:        &po.ID,
			OccurredAt:   occurredAt,
		}); err != nil {
			return err
		}
	}

	for _, input := range inputs {
		if _, err := applyMovementTx(ctx, tx, model.StockMovement{
			LotID:        input.MaterialLotID,
			MovementType: "production_consume",
			Qty:          -input.Qty,
			RefType:      &refType,
			RefID:        &po.ID,
			OccurredAt:   occurredAt,
		}); err != nil {
			return err
		}

		remaining := input.Qty
		for i, output := range outputs {
			linkQty := roundQty(input.Qty * output.Qty / totalOutputQty)
			if i == len(outputs)-1 {
				linkQty = roundQty(remaining)
			}
			if linkQty < 0 {
				linkQty = 0
			}
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO lot_links (parent_lot_id, child_lot_id, qty, link_type) VALUES (?, ?, ?, 'production')`,
				input.MaterialLotID, output.ProductLotID, linkQty); err != nil {
				return err
			}
			remaining -= linkQty
		}
	}
	return nil
}

// CompleteProductionOrder applies stock and links once, then marks the order
// completed. Repeated calls are idempotent.
func (r *mysqlRepo) CompleteProductionOrder(ctx context.Context, id int64, finishedAt *string) (model.ProductionOrder, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return model.ProductionOrder{}, err
	}
	defer tx.Rollback()

	var po model.ProductionOrder
	var startedAt, existingFinished, notes, serviceName sql.NullString
	if err := tx.QueryRowContext(ctx,
		`SELECT id, code, status, started_at, finished_at, notes, service_name,
		        service_unit_cost, service_qty, expected_qty
		 FROM production_orders WHERE id = ? FOR UPDATE`, id).
		Scan(&po.ID, &po.Code, &po.Status, &startedAt, &existingFinished,
			&notes, &serviceName, &po.ServiceUnitCost, &po.ServiceQty, &po.ExpectedQty); err != nil {
		return model.ProductionOrder{}, err
	}
	if startedAt.Valid {
		po.StartedAt = &startedAt.String
	}
	if existingFinished.Valid {
		po.FinishedAt = &existingFinished.String
	}
	if notes.Valid {
		po.Notes = &notes.String
	}
	if serviceName.Valid {
		po.ServiceName = &serviceName.String
	}
	if po.Status == "completed" {
		return po, nil
	}

	inputRows, err := tx.QueryContext(ctx,
		`SELECT id, production_order_id, material_lot_id, qty
		 FROM production_order_inputs WHERE production_order_id = ? ORDER BY id`, id)
	if err != nil {
		return model.ProductionOrder{}, err
	}
	var inputs []model.ProductionOrderInput
	for inputRows.Next() {
		var input model.ProductionOrderInput
		if err := inputRows.Scan(&input.ID, &input.ProductionOrderID, &input.MaterialLotID, &input.Qty); err != nil {
			inputRows.Close()
			return model.ProductionOrder{}, err
		}
		inputs = append(inputs, input)
	}
	inputRows.Close()

	outputRows, err := tx.QueryContext(ctx,
		`SELECT id, production_order_id, product_lot_id, qty
		 FROM production_order_outputs WHERE production_order_id = ? ORDER BY id`, id)
	if err != nil {
		return model.ProductionOrder{}, err
	}
	var outputs []model.ProductionOrderOutput
	for outputRows.Next() {
		var output model.ProductionOrderOutput
		if err := outputRows.Scan(&output.ID, &output.ProductionOrderID, &output.ProductLotID, &output.Qty); err != nil {
			outputRows.Close()
			return model.ProductionOrder{}, err
		}
		outputs = append(outputs, output)
	}
	outputRows.Close()
	if len(inputs) == 0 || len(outputs) == 0 {
		return model.ProductionOrder{}, ErrInvalidOrder
	}

	occurredAt := time.Now().Format("2006-01-02 15:04:05")
	if finishedAt != nil && *finishedAt != "" {
		occurredAt = *finishedAt
	} else if po.FinishedAt != nil && *po.FinishedAt != "" {
		occurredAt = *po.FinishedAt
	}

	// Old drafts may have been created by the previous buggy path; do not
	// double-consume stock if movements already exist for this order.
	var movementCount int
	if err := tx.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM stock_movements
		 WHERE ref_type = 'production_order' AND ref_id = ?`, id).Scan(&movementCount); err != nil {
		return model.ProductionOrder{}, err
	}
	if movementCount == 0 {
		if err := applyProductionOrderTx(ctx, tx, po, inputs, outputs, occurredAt); err != nil {
			return model.ProductionOrder{}, err
		}
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE production_orders SET status = 'completed', finished_at = ? WHERE id = ?`,
		occurredAt, id); err != nil {
		return model.ProductionOrder{}, err
	}
	if err := tx.Commit(); err != nil {
		return model.ProductionOrder{}, err
	}
	po.Status = "completed"
	po.FinishedAt = &occurredAt
	return po, nil
}

func (r *mysqlRepo) RejectProductionOrder(ctx context.Context, id int64, reason *string) (model.ProductionOrder, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return model.ProductionOrder{}, err
	}
	defer tx.Rollback()

	var po model.ProductionOrder
	var startedAt, finishedAt, notes, serviceName sql.NullString
	if err := tx.QueryRowContext(ctx,
		`SELECT id, code, status, started_at, finished_at, notes, service_name,
		        service_unit_cost, service_qty, expected_qty
		 FROM production_orders WHERE id = ? FOR UPDATE`, id).
		Scan(&po.ID, &po.Code, &po.Status, &startedAt, &finishedAt,
			&notes, &serviceName, &po.ServiceUnitCost, &po.ServiceQty, &po.ExpectedQty); err != nil {
		return model.ProductionOrder{}, err
	}
	if startedAt.Valid {
		po.StartedAt = &startedAt.String
	}
	if finishedAt.Valid {
		po.FinishedAt = &finishedAt.String
	}
	if notes.Valid {
		po.Notes = &notes.String
	}
	if serviceName.Valid {
		po.ServiceName = &serviceName.String
	}
	if po.Status == "completed" {
		return model.ProductionOrder{}, ErrInvalidOrder
	}
	if po.Status == "rejected" {
		return po, nil
	}

	// A rejected order must not carry applied stock movements; otherwise the
	// stock ledger and the order status would contradict each other.
	var movementCount int
	if err := tx.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM stock_movements
		 WHERE ref_type = 'production_order' AND ref_id = ?`, id).Scan(&movementCount); err != nil {
		return model.ProductionOrder{}, err
	}
	if movementCount > 0 {
		return model.ProductionOrder{}, ErrInvalidOrder
	}

	newNotes := ""
	if notes.Valid {
		newNotes = notes.String
	}
	if reason != nil && strings.TrimSpace(*reason) != "" {
		if newNotes != "" {
			newNotes += "\n"
		}
		newNotes += "Ret: " + strings.TrimSpace(*reason)
	}
	var notesArg any
	if newNotes != "" {
		notesArg = newNotes
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE production_orders SET status = 'rejected', notes = ? WHERE id = ?`,
		notesArg, id); err != nil {
		return model.ProductionOrder{}, err
	}
	if err := tx.Commit(); err != nil {
		return model.ProductionOrder{}, err
	}
	po.Status = "rejected"
	if newNotes != "" {
		po.Notes = &newNotes
	}
	return po, nil
}

func (r *mysqlRepo) GetProductionOrder(ctx context.Context, id int64) (model.ProductionOrder, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, code, status, started_at, finished_at, notes, service_name, service_unit_cost, service_qty, expected_qty
		 FROM production_orders WHERE id = ?`, id)

	var po model.ProductionOrder
	var startedAt, finishedAt, notes, serviceName sql.NullString
	if err := row.Scan(&po.ID, &po.Code, &po.Status, &startedAt, &finishedAt, &notes, &serviceName, &po.ServiceUnitCost, &po.ServiceQty, &po.ExpectedQty); err != nil {
		return model.ProductionOrder{}, err
	}
	if startedAt.Valid {
		po.StartedAt = &startedAt.String
	}
	if finishedAt.Valid {
		po.FinishedAt = &finishedAt.String
	}
	if notes.Valid {
		po.Notes = &notes.String
	}
	if serviceName.Valid {
		po.ServiceName = &serviceName.String
	}
	return po, nil
}

func (r *mysqlRepo) ListProductionOrderInputs(ctx context.Context, poID int64) ([]model.ProductionOrderInput, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, production_order_id, material_lot_id, qty FROM production_order_inputs WHERE production_order_id = ? ORDER BY id`, poID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.ProductionOrderInput
	for rows.Next() {
		var input model.ProductionOrderInput
		if err := rows.Scan(&input.ID, &input.ProductionOrderID, &input.MaterialLotID, &input.Qty); err != nil {
			return nil, err
		}
		out = append(out, input)
	}
	return out, rows.Err()
}

func (r *mysqlRepo) ListProductionOrderCodesByInputLot(ctx context.Context, lotID int64) ([]string, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT DISTINCT po.code FROM production_order_inputs poi
		 JOIN production_orders po ON po.id = poi.production_order_id
		 WHERE poi.material_lot_id = ? ORDER BY po.code`, lotID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}
		out = append(out, code)
	}
	return out, rows.Err()
}

func (r *mysqlRepo) ListProductionOrderOutputs(ctx context.Context, poID int64) ([]model.ProductionOrderOutput, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, production_order_id, product_lot_id, qty FROM production_order_outputs WHERE production_order_id = ? ORDER BY id`, poID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.ProductionOrderOutput
	for rows.Next() {
		var output model.ProductionOrderOutput
		if err := rows.Scan(&output.ID, &output.ProductionOrderID, &output.ProductLotID, &output.Qty); err != nil {
			return nil, err
		}
		out = append(out, output)
	}
	return out, rows.Err()
}

func (r *mysqlRepo) GetLot(ctx context.Context, id int64) (model.Lot, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, code, item_type, item_id, qty, unit, supplier_id, purchase_price, unit_cost, received_at, expiry_date, status, created_at, updated_at FROM lots WHERE id = ?`,
		id)
	return scanLot(row)
}

func (r *mysqlRepo) GetProduct(ctx context.Context, id int64) (model.Product, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, code, name, unit, list_price, category, is_finished, shelf_life_days, created_at
		 FROM products WHERE id = ?`, id)

	var p model.Product
	if err := row.Scan(&p.ID, &p.Code, &p.Name, &p.Unit, &p.ListPrice,
		&p.Category, &p.IsFinished, &p.ShelfLifeDays, &p.CreatedAt); err != nil {
		return model.Product{}, err
	}
	return p, nil
}

func (r *mysqlRepo) GetMarket(ctx context.Context, id int64) (model.Market, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, code, name, contact FROM markets WHERE id = ?`, id)

	var market model.Market
	if err := row.Scan(&market.ID, &market.Code, &market.Name, &market.Contact); err != nil {
		return model.Market{}, err
	}
	return market, nil
}

func (r *mysqlRepo) GetSupplier(ctx context.Context, id int64) (model.Supplier, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, code, name, contact FROM suppliers WHERE id = ?`, id)

	var supplier model.Supplier
	if err := row.Scan(&supplier.ID, &supplier.Code, &supplier.Name, &supplier.Contact); err != nil {
		return model.Supplier{}, err
	}
	return supplier, nil
}

func (r *mysqlRepo) CreateButcheryBatch(ctx context.Context, batch model.ButcheryBatch, items []model.ButcheryItem) (model.ButcheryBatch, error) {
	if batch.Code == "" || batch.InputWeight <= 0 || batch.UnitCost < 0 {
		return model.ButcheryBatch{}, ErrInvalidOrder
	}
	if batch.Status == "" {
		batch.Status = "draft"
	}
	if batch.BatchDate == "" {
		batch.BatchDate = time.Now().Format("2006-01-02")
	}

	hasPart := false
	var partWeight float64
	for _, item := range items {
		if item.Weight < 0 {
			return model.ButcheryBatch{}, ErrInvalidOrder
		}
		if !item.IsWaste && item.Weight > 0 {
			hasPart = true
			partWeight += item.Weight
		}
	}
	if batch.Status == "completed" && (batch.InputLotID == nil || !hasPart) {
		return model.ButcheryBatch{}, ErrInvalidOrder
	}
	fire := roundQty(batch.InputWeight - partWeight)
	if fire < -0.0005 {
		return model.ButcheryBatch{}, ErrInvalidOrder
	}
	if fire < 0 {
		fire = 0
	}
	if batch.TotalCost <= 0 {
		batch.TotalCost = roundMoney(batch.InputWeight * batch.UnitCost)
	}
	allocated := allocateByWeight(batch.TotalCost, items)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return model.ButcheryBatch{}, err
	}
	defer tx.Rollback()

	if batch.InputLotID != nil {
		var itemType string
		var lotQty float64
		if err := tx.QueryRowContext(ctx,
			`SELECT item_type, qty FROM lots WHERE id = ? FOR UPDATE`, *batch.InputLotID).
			Scan(&itemType, &lotQty); err != nil {
			return model.ButcheryBatch{}, err
		}
		if itemType != "material" {
			return model.ButcheryBatch{}, ErrInvalidOrder
		}
		if lotQty+0.0005 < batch.InputWeight {
			return model.ButcheryBatch{}, ErrInsufficientStock
		}
	}

	res, err := tx.ExecContext(ctx,
		`INSERT INTO butchery_batches (code, animal_type, input_lot_id, input_weight, unit_cost, total_cost, batch_date, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		batch.Code, batch.AnimalType, nullInt64(batch.InputLotID), batch.InputWeight,
		batch.UnitCost, batch.TotalCost, batch.BatchDate, batch.Status)
	if err != nil {
		return model.ButcheryBatch{}, err
	}
	batch.ID, err = res.LastInsertId()
	if err != nil {
		return model.ButcheryBatch{}, err
	}

	// A draft without an input lot is a planning record: keep the parts but do
	// not create stock or lot links until the batch is completed.
	if batch.InputLotID == nil {
		for i, item := range items {
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO butchery_items (batch_id, product_id, product_lot_id, weight, quantity, is_waste, allocated_cost)
				 VALUES (?, ?, ?, ?, ?, ?, ?)`,
				batch.ID, item.ProductID, nullInt64(item.ProductLotID), item.Weight,
				item.Quantity, item.IsWaste, allocated[i]); err != nil {
				return model.ButcheryBatch{}, err
			}
		}
		if err := tx.Commit(); err != nil {
			return model.ButcheryBatch{}, err
		}
		return batch, nil
	}

	refType := "butchery_batch"
	occurredAt := batch.BatchDate + " 00:00:00"
	for i, item := range items {
		if item.IsWaste {
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO butchery_items (batch_id, product_id, product_lot_id, weight, quantity, is_waste, allocated_cost)
				 VALUES (?, ?, NULL, ?, ?, TRUE, 0)`,
				batch.ID, item.ProductID, item.Weight, item.Quantity); err != nil {
				return model.ButcheryBatch{}, err
			}
			continue
		}

		lotID := int64(0)
		var productCode string
		var shelfLifeDays int
		if err := tx.QueryRowContext(ctx,
			`SELECT code, shelf_life_days FROM products WHERE id = ?`, item.ProductID).
			Scan(&productCode, &shelfLifeDays); err != nil {
			return model.ButcheryBatch{}, err
		}
		receivedAt := batch.BatchDate + " 00:00:00"
		expiryDate := ""
		if shelfLifeDays > 0 {
			expiryDate = addDays(receivedAt, shelfLifeDays)
		}
		var expiryArg any
		if expiryDate != "" {
			expiryArg = expiryDate
		}
		if item.ProductLotID != nil {
			lotID = *item.ProductLotID
			var lotType string
			if err := tx.QueryRowContext(ctx,
				`SELECT item_type FROM lots WHERE id = ? FOR UPDATE`, lotID).Scan(&lotType); err != nil {
				return model.ButcheryBatch{}, err
			}
			if lotType != "product" {
				return model.ButcheryBatch{}, ErrInvalidOrder
			}
		} else {
			lotCode, err := uniqueLotCodeTx(ctx, tx, batch.Code+"-"+productCode)
			if err != nil {
				return model.ButcheryBatch{}, err
			}
			res, err := tx.ExecContext(ctx,
				`INSERT INTO lots (code, item_type, item_id, qty, unit, unit_cost, received_at, expiry_date, status)
				 VALUES (?, 'product', ?, 0, 'kg', 0, ?, ?, 'active')`,
				lotCode, item.ProductID, receivedAt, expiryArg)
			if err != nil {
				return model.ButcheryBatch{}, err
			}
			lotID, err = res.LastInsertId()
			if err != nil {
				return model.ButcheryBatch{}, err
			}
		}
		if shelfLifeDays > 0 {
			if _, err := tx.ExecContext(ctx,
				`UPDATE lots SET received_at = COALESCE(received_at, ?),
				 expiry_date = COALESCE(expiry_date, ?), updated_at = CURRENT_TIMESTAMP
				 WHERE id = ?`, receivedAt, expiryDate, lotID); err != nil {
				return model.ButcheryBatch{}, err
			}
		}

		unitCost := 0.0
		if item.Weight > 0 {
			unitCost = roundMoney(allocated[i] / item.Weight)
		}
		if _, err := tx.ExecContext(ctx,
			`UPDATE lots SET unit_cost = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
			unitCost, lotID); err != nil {
			return model.ButcheryBatch{}, err
		}
		if _, err := applyMovementTx(ctx, tx, model.StockMovement{
			LotID:        lotID,
			MovementType: "butchery_output",
			Qty:          item.Weight,
			UnitCost:     unitCost,
			RefType:      &refType,
			RefID:        &batch.ID,
			OccurredAt:   occurredAt,
		}); err != nil {
			return model.ButcheryBatch{}, err
		}
		if batch.InputLotID != nil {
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO lot_links (parent_lot_id, child_lot_id, qty, link_type) VALUES (?, ?, ?, 'butchery')`,
				*batch.InputLotID, lotID, item.Weight); err != nil {
				return model.ButcheryBatch{}, err
			}
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO butchery_items (batch_id, product_id, product_lot_id, weight, quantity, is_waste, allocated_cost)
			 VALUES (?, ?, ?, ?, ?, FALSE, ?)`,
			batch.ID, item.ProductID, lotID, item.Weight, item.Quantity, allocated[i]); err != nil {
			return model.ButcheryBatch{}, err
		}
	}

	if batch.InputLotID != nil {
		consumeQty := roundQty(-(batch.InputWeight - fire))
		if consumeQty != 0 {
			if _, err := applyMovementTx(ctx, tx, model.StockMovement{
				LotID:        *batch.InputLotID,
				MovementType: "butchery_consume",
				Qty:          consumeQty,
				RefType:      &refType,
				RefID:        &batch.ID,
				OccurredAt:   occurredAt,
			}); err != nil {
				return model.ButcheryBatch{}, err
			}
		}
		if fire > 0 {
			if _, err := applyMovementTx(ctx, tx, model.StockMovement{
				LotID:        *batch.InputLotID,
				MovementType: "waste",
				Qty:          -fire,
				RefType:      &refType,
				RefID:        &batch.ID,
				OccurredAt:   occurredAt,
			}); err != nil {
				return model.ButcheryBatch{}, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return model.ButcheryBatch{}, err
	}
	return batch, nil
}

func allocateByWeight(total float64, items []model.ButcheryItem) []float64 {
	allocated := make([]float64, len(items))
	var sum float64
	lastPart := -1
	for i, item := range items {
		if item.IsWaste {
			continue
		}
		sum += item.Weight
		lastPart = i
	}
	if lastPart == -1 {
		return allocated
	}
	if sum <= 0 {
		allocated[lastPart] = roundMoney(total)
		return allocated
	}

	var used float64
	for i, item := range items {
		if item.IsWaste {
			continue
		}
		if i == lastPart {
			allocated[i] = roundMoney(total - used)
			continue
		}
		share := roundMoney(total * item.Weight / sum)
		allocated[i] = share
		used += share
	}
	return allocated
}

func uniqueLotCodeTx(ctx context.Context, tx *sql.Tx, base string) (string, error) {
	base = strings.ToUpper(strings.TrimSpace(base))
	if base == "" {
		return "", ErrInvalidOrder
	}
	if len(base) > 64 {
		base = base[:64]
	}
	for i := 1; i <= 100; i++ {
		candidate := base
		if i > 1 {
			suffix := fmt.Sprintf("-%d", i)
			trimmed := base
			if len(trimmed) > 64-len(suffix) {
				trimmed = trimmed[:64-len(suffix)]
			}
			candidate = trimmed + suffix
		}
		var exists bool
		if err := tx.QueryRowContext(ctx,
			`SELECT EXISTS(SELECT 1 FROM lots WHERE code = ?)`, candidate).Scan(&exists); err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("could not allocate a unique lot code for %q", base)
}

func (r *mysqlRepo) GetButcheryBatch(ctx context.Context, id int64) (model.ButcheryBatch, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, code, animal_type, input_lot_id, input_weight, unit_cost, total_cost, batch_date, status
		 FROM butchery_batches WHERE id = ?`, id)

	var batch model.ButcheryBatch
	var inputLotID sql.NullInt64
	if err := row.Scan(&batch.ID, &batch.Code, &batch.AnimalType, &inputLotID, &batch.InputWeight,
		&batch.UnitCost, &batch.TotalCost, &batch.BatchDate, &batch.Status); err != nil {
		return model.ButcheryBatch{}, err
	}
	if inputLotID.Valid {
		batch.InputLotID = &inputLotID.Int64
	}
	return batch, nil
}

func (r *mysqlRepo) ListButcheryItems(ctx context.Context, batchID int64) ([]model.ButcheryItem, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT bi.id, bi.batch_id, bi.product_id, bi.product_lot_id, bi.weight, bi.quantity,
		        bi.is_waste, bi.allocated_cost, l.code
		 FROM butchery_items bi
		 LEFT JOIN lots l ON l.id = bi.product_lot_id
		 WHERE bi.batch_id = ? ORDER BY bi.id`, batchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.ButcheryItem
	for rows.Next() {
		var item model.ButcheryItem
		var productLotID sql.NullInt64
		var lotCode sql.NullString
		if err := rows.Scan(&item.ID, &item.BatchID, &item.ProductID, &productLotID,
			&item.Weight, &item.Quantity, &item.IsWaste, &item.AllocatedCost, &lotCode); err != nil {
			return nil, err
		}
		if productLotID.Valid {
			item.ProductLotID = &productLotID.Int64
		}
		if lotCode.Valid {
			item.LotCode = lotCode.String
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *mysqlRepo) ListButcheryBatches(ctx context.Context) ([]model.ButcheryBatch, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, code, animal_type, input_lot_id, input_weight, unit_cost, total_cost, batch_date, status
		 FROM butchery_batches ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.ButcheryBatch
	for rows.Next() {
		var batch model.ButcheryBatch
		var inputLotID sql.NullInt64
		if err := rows.Scan(&batch.ID, &batch.Code, &batch.AnimalType, &inputLotID, &batch.InputWeight,
			&batch.UnitCost, &batch.TotalCost, &batch.BatchDate, &batch.Status); err != nil {
			return nil, err
		}
		if inputLotID.Valid {
			batch.InputLotID = &inputLotID.Int64
		}
		out = append(out, batch)
	}
	return out, rows.Err()
}

func (r *mysqlRepo) ListProductionOrders(ctx context.Context) ([]model.ProductionOrder, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, code, status, started_at, finished_at, notes, service_name, service_unit_cost, service_qty, expected_qty
		 FROM production_orders ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.ProductionOrder
	for rows.Next() {
		var po model.ProductionOrder
		var startedAt, finishedAt, notes, serviceName sql.NullString
		if err := rows.Scan(&po.ID, &po.Code, &po.Status, &startedAt, &finishedAt, &notes, &serviceName, &po.ServiceUnitCost, &po.ServiceQty, &po.ExpectedQty); err != nil {
			return nil, err
		}
		if startedAt.Valid {
			po.StartedAt = &startedAt.String
		}
		if finishedAt.Valid {
			po.FinishedAt = &finishedAt.String
		}
		if notes.Valid {
			po.Notes = &notes.String
		}
		if serviceName.Valid {
			po.ServiceName = &serviceName.String
		}
		out = append(out, po)
	}
	return out, rows.Err()
}

func (r *mysqlRepo) CreateRecall(ctx context.Context, event model.RecallEvent, blocks []model.RecallBlock) (model.RecallEvent, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return model.RecallEvent{}, err
	}
	defer tx.Rollback()

	code, err := uniqueRecallCodeTx(ctx, tx, event.Code)
	if err != nil {
		return model.RecallEvent{}, err
	}

	res, err := tx.ExecContext(ctx,
		`INSERT INTO recall_events (code, reason, source_lot_id) VALUES (?, ?, ?)`,
		code, event.Reason, event.SourceLotID)
	if err != nil {
		return model.RecallEvent{}, err
	}
	event.ID, err = res.LastInsertId()
	if err != nil {
		return model.RecallEvent{}, err
	}
	event.Code = code

	blockedAt := time.Now().Format("2006-01-02 15:04:05")
	for _, block := range blocks {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO recall_blocks (recall_event_id, lot_id, blocked_at) VALUES (?, ?, ?)`,
			event.ID, block.LotID, blockedAt); err != nil {
			return model.RecallEvent{}, err
		}
		if _, err := tx.ExecContext(ctx,
			`UPDATE lots SET status = 'blocked', updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
			block.LotID); err != nil {
			return model.RecallEvent{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return model.RecallEvent{}, err
	}
	return event, nil
}

func uniqueRecallCodeTx(ctx context.Context, tx *sql.Tx, base string) (string, error) {
	base = strings.ToUpper(strings.TrimSpace(base))
	if base == "" {
		base = "REC-" + time.Now().Format("20060102-150405")
	}
	if len(base) > 64 {
		base = base[:64]
	}
	for i := 1; i <= 100; i++ {
		candidate := base
		if i > 1 {
			suffix := fmt.Sprintf("-%d", i)
			trimmed := base
			if len(trimmed) > 64-len(suffix) {
				trimmed = trimmed[:64-len(suffix)]
			}
			candidate = trimmed + suffix
		}
		var exists bool
		if err := tx.QueryRowContext(ctx,
			`SELECT EXISTS(SELECT 1 FROM recall_events WHERE code = ?)`, candidate).Scan(&exists); err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("could not allocate a unique recall code for %q", base)
}

func (r *mysqlRepo) CreateRecallEvent(ctx context.Context, event model.RecallEvent) (model.RecallEvent, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO recall_events (code, reason, source_lot_id) VALUES (?, ?, ?)`,
		event.Code, event.Reason, event.SourceLotID)
	if err != nil {
		return model.RecallEvent{}, err
	}
	event.ID, err = res.LastInsertId()
	return event, err
}

func (r *mysqlRepo) GetRecallEvent(ctx context.Context, id int64) (model.RecallEvent, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, code, reason, source_lot_id, created_at FROM recall_events WHERE id = ?`, id)
	var event model.RecallEvent
	if err := row.Scan(&event.ID, &event.Code, &event.Reason, &event.SourceLotID, &event.CreatedAt); err != nil {
		return model.RecallEvent{}, err
	}
	return event, nil
}

func (r *mysqlRepo) ListRecallEvents(ctx context.Context) ([]model.RecallEvent, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, code, reason, source_lot_id, created_at FROM recall_events ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.RecallEvent
	for rows.Next() {
		var event model.RecallEvent
		var reason sql.NullString
		if err := rows.Scan(&event.ID, &event.Code, &reason, &event.SourceLotID, &event.CreatedAt); err != nil {
			return nil, err
		}
		if reason.Valid {
			event.Reason = reason.String
		}
		out = append(out, event)
	}
	return out, rows.Err()
}

func (r *mysqlRepo) CreateRecallBlock(ctx context.Context, block model.RecallBlock) (model.RecallBlock, error) {
	if block.BlockedAt == "" {
		block.BlockedAt = time.Now().Format("2006-01-02 15:04:05")
	}
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO recall_blocks (recall_event_id, lot_id, blocked_at) VALUES (?, ?, ?)`,
		block.RecallEventID, block.LotID, block.BlockedAt)
	if err != nil {
		return model.RecallBlock{}, err
	}
	block.ID, err = res.LastInsertId()
	return block, err
}

func (r *mysqlRepo) ListRecallBlocks(ctx context.Context, recallEventID int64) ([]model.RecallBlock, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, recall_event_id, lot_id, blocked_at FROM recall_blocks WHERE recall_event_id = ? ORDER BY id`, recallEventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.RecallBlock
	for rows.Next() {
		var block model.RecallBlock
		if err := rows.Scan(&block.ID, &block.RecallEventID, &block.LotID, &block.BlockedAt); err != nil {
			return nil, err
		}
		out = append(out, block)
	}
	return out, rows.Err()
}

func (r *mysqlRepo) IsLotBlocked(ctx context.Context, lotID int64) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM lots WHERE id = ? AND status = 'blocked')
		    OR EXISTS(SELECT 1 FROM recall_blocks WHERE lot_id = ?)`, lotID, lotID).Scan(&exists)
	return exists, err
}

func (r *mysqlRepo) ListBlockedLotCodes(ctx context.Context) ([]string, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT l.code FROM lots l
		 WHERE l.status = 'blocked'
		    OR EXISTS(SELECT 1 FROM recall_blocks rb WHERE rb.lot_id = l.id)
		 ORDER BY l.code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}
		out = append(out, code)
	}
	return out, rows.Err()
}

func (r *mysqlRepo) CreateSale(ctx context.Context, sale model.Sale, lines []model.SaleLine) (model.Sale, error) {
	// Client retries must not create a second sale or deduct stock twice.
	var existing model.Sale
	var existingClient sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT id, code, market_id, sold_at, status, client_id FROM sales WHERE code = ?`,
		sale.Code).
		Scan(&existing.ID, &existing.Code, &existing.MarketID, &existing.SoldAt, &existing.Status, &existingClient)
	if err == nil {
		if existingClient.Valid {
			existing.ClientID = &existingClient.String
		}
		return existing, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return model.Sale{}, err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return model.Sale{}, err
	}
	defer tx.Rollback()

	if sale.SoldAt == "" {
		sale.SoldAt = time.Now().Format("2006-01-02 15:04:05")
	}
	res, err := tx.ExecContext(ctx,
		`INSERT INTO sales (code, market_id, sold_at, status, client_id) VALUES (?, ?, ?, ?, ?)`,
		sale.Code, sale.MarketID, sale.SoldAt, sale.Status, nullString(sale.ClientID))
	if err != nil {
		return model.Sale{}, err
	}
	sale.ID, err = res.LastInsertId()
	if err != nil {
		return model.Sale{}, err
	}

	for _, line := range lines {
		var blocked bool
		if err := tx.QueryRowContext(ctx,
			`SELECT EXISTS(SELECT 1 FROM lots WHERE id = ? AND status = 'blocked')
			    OR EXISTS(SELECT 1 FROM recall_blocks WHERE lot_id = ?)`,
			line.LotID, line.LotID).Scan(&blocked); err != nil {
			return model.Sale{}, err
		}
		if blocked {
			return model.Sale{}, ErrLotBlocked
		}

		// Satış fiyatı line.unit_price'ta kalır; stok hareketi lot maliyetini kullanır.
		refType := "sale"
		movement, err := applyMovementTx(ctx, tx, model.StockMovement{
			LotID:        line.LotID,
			MovementType: "sale",
			Qty:          -line.Qty,
			RefType:      &refType,
			RefID:        &sale.ID,
			OccurredAt:   sale.SoldAt,
		})
		if err != nil {
			return model.Sale{}, err
		}

		if _, err := tx.ExecContext(ctx,
			`INSERT INTO sale_lines (sale_id, lot_id, qty, unit_price, total_price, stock_movement_id) VALUES (?, ?, ?, ?, ?, ?)`,
			sale.ID, line.LotID, line.Qty, line.UnitPrice, line.TotalPrice, movement.ID); err != nil {
			return model.Sale{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return model.Sale{}, err
	}
	return sale, nil
}

func (r *mysqlRepo) ListSales(ctx context.Context) ([]model.Sale, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, code, market_id, sold_at, status, client_id FROM sales ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Sale
	for rows.Next() {
		var sale model.Sale
		var clientID sql.NullString
		if err := rows.Scan(&sale.ID, &sale.Code, &sale.MarketID, &sale.SoldAt, &sale.Status, &clientID); err != nil {
			return nil, err
		}
		if clientID.Valid {
			sale.ClientID = &clientID.String
		}
		out = append(out, sale)
	}
	return out, rows.Err()
}

func (r *mysqlRepo) ListSaleLines(ctx context.Context, saleID int64) ([]model.SaleLine, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, sale_id, lot_id, qty, unit_price, total_price, stock_movement_id
		 FROM sale_lines WHERE sale_id = ? ORDER BY id`, saleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.SaleLine
	for rows.Next() {
		var line model.SaleLine
		var movementID sql.NullInt64
		if err := rows.Scan(&line.ID, &line.SaleID, &line.LotID, &line.Qty,
			&line.UnitPrice, &line.TotalPrice, &movementID); err != nil {
			return nil, err
		}
		if movementID.Valid {
			line.StockMovementID = &movementID.Int64
		}
		out = append(out, line)
	}
	return out, rows.Err()
}

func (r *mysqlRepo) ListMarkets(ctx context.Context) ([]model.Market, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, code, name, contact FROM markets ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Market
	for rows.Next() {
		var market model.Market
		if err := rows.Scan(&market.ID, &market.Code, &market.Name, &market.Contact); err != nil {
			return nil, err
		}
		out = append(out, market)
	}
	return out, rows.Err()
}

func (r *mysqlRepo) CreateRedistribution(ctx context.Context, red model.Redistribution) (model.Redistribution, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO market_redistributions (market_id, lot_id, sub_buyer, contact, qty) VALUES (?, ?, ?, ?, ?)`,
		red.MarketID, red.LotID, red.SubBuyer, nullString(red.Contact), red.Qty)
	if err != nil {
		return model.Redistribution{}, err
	}
	red.ID, err = res.LastInsertId()
	return red, err
}

func (r *mysqlRepo) ListRedistributions(ctx context.Context, lotID int64) ([]model.Redistribution, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, market_id, lot_id, sub_buyer, contact, qty FROM market_redistributions WHERE lot_id = ? ORDER BY id`, lotID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Redistribution
	for rows.Next() {
		var red model.Redistribution
		var contact sql.NullString
		if err := rows.Scan(&red.ID, &red.MarketID, &red.LotID, &red.SubBuyer, &contact, &red.Qty); err != nil {
			return nil, err
		}
		if contact.Valid {
			red.Contact = &contact.String
		}
		out = append(out, red)
	}
	return out, rows.Err()
}
