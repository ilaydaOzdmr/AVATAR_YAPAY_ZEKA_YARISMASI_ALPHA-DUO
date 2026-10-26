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
	ListLotLinksByParent(ctx context.Context, parentID int64) ([]model.LotLink, error)
	ListLotLinksByChild(ctx context.Context, childID int64) ([]model.LotLink, error)
	AddStockMovement(ctx context.Context, m model.StockMovement) (model.StockMovement, error)
	ListStockMovements(ctx context.Context, lotID int64) ([]model.StockMovement, error)
	CreateProductionOrder(ctx context.Context, po model.ProductionOrder, inputs []model.ProductionOrderInput, outputs []model.ProductionOrderOutput) (model.ProductionOrder, error)
	GetProductionOrder(ctx context.Context, id int64) (model.ProductionOrder, error)
	ListProductionOrderInputs(ctx context.Context, poID int64) ([]model.ProductionOrderInput, error)
	ListProductionOrderOutputs(ctx context.Context, poID int64) ([]model.ProductionOrderOutput, error)
	GetLot(ctx context.Context, id int64) (model.Lot, error)
	GetProduct(ctx context.Context, id int64) (model.Product, error)
	GetMarket(ctx context.Context, id int64) (model.Market, error)
	GetSupplier(ctx context.Context, id int64) (model.Supplier, error)
	CreateButcheryBatch(ctx context.Context, batch model.ButcheryBatch, items []model.ButcheryItem) (model.ButcheryBatch, error)
	GetButcheryBatch(ctx context.Context, id int64) (model.ButcheryBatch, error)
	ListButcheryItems(ctx context.Context, batchID int64) ([]model.ButcheryItem, error)
	CreateRecallEvent(ctx context.Context, event model.RecallEvent) (model.RecallEvent, error)
	GetRecallEvent(ctx context.Context, id int64) (model.RecallEvent, error)
	CreateRecallBlock(ctx context.Context, block model.RecallBlock) (model.RecallBlock, error)
	ListRecallBlocks(ctx context.Context, recallEventID int64) ([]model.RecallBlock, error)
	IsLotBlocked(ctx context.Context, lotID int64) (bool, error)
	ListBlockedLotCodes(ctx context.Context) ([]string, error)
	CreateSale(ctx context.Context, sale model.Sale, lines []model.SaleLine) (model.Sale, error)
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

func (r *mysqlRepo) CreateProductionOrder(ctx context.Context, po model.ProductionOrder, inputs []model.ProductionOrderInput, outputs []model.ProductionOrderOutput) (model.ProductionOrder, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return model.ProductionOrder{}, err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`INSERT INTO production_orders (code, status, started_at, finished_at, notes, service_name, service_unit_cost, service_qty)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		po.Code, po.Status, nullString(po.StartedAt), nullString(po.FinishedAt),
		nullString(po.Notes), nullString(po.ServiceName), po.ServiceUnitCost, po.ServiceQty)
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

	if err := tx.Commit(); err != nil {
		return model.ProductionOrder{}, err
	}
	return po, nil
}

func (r *mysqlRepo) GetProductionOrder(ctx context.Context, id int64) (model.ProductionOrder, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, code, status, started_at, finished_at, notes, service_name, service_unit_cost, service_qty
		 FROM production_orders WHERE id = ?`, id)

	var po model.ProductionOrder
	var startedAt, finishedAt, notes, serviceName sql.NullString
	if err := row.Scan(&po.ID, &po.Code, &po.Status, &startedAt, &finishedAt, &notes, &serviceName, &po.ServiceUnitCost, &po.ServiceQty); err != nil {
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
		`SELECT id, code, name, unit, list_price, category, is_finished FROM products WHERE id = ?`, id)

	var p model.Product
	if err := row.Scan(&p.ID, &p.Code, &p.Name, &p.Unit, &p.ListPrice, &p.Category, &p.IsFinished); err != nil {
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
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return model.ButcheryBatch{}, err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`INSERT INTO butchery_batches (code, animal_type, input_weight, unit_cost, total_cost, batch_date, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		batch.Code, batch.AnimalType, batch.InputWeight, batch.UnitCost,
		batch.TotalCost, batch.BatchDate, batch.Status)
	if err != nil {
		return model.ButcheryBatch{}, err
	}
	batch.ID, err = res.LastInsertId()
	if err != nil {
		return model.ButcheryBatch{}, err
	}

	for _, item := range items {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO butchery_items (batch_id, product_id, product_lot_id, weight, is_waste) VALUES (?, ?, ?, ?, ?)`,
			batch.ID, item.ProductID, nullInt64(item.ProductLotID), item.Weight, item.IsWaste); err != nil {
			return model.ButcheryBatch{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return model.ButcheryBatch{}, err
	}
	return batch, nil
}

func (r *mysqlRepo) GetButcheryBatch(ctx context.Context, id int64) (model.ButcheryBatch, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, code, animal_type, input_weight, unit_cost, total_cost, batch_date, status
		 FROM butchery_batches WHERE id = ?`, id)

	var batch model.ButcheryBatch
	if err := row.Scan(&batch.ID, &batch.Code, &batch.AnimalType, &batch.InputWeight,
		&batch.UnitCost, &batch.TotalCost, &batch.BatchDate, &batch.Status); err != nil {
		return model.ButcheryBatch{}, err
	}
	return batch, nil
}

func (r *mysqlRepo) ListButcheryItems(ctx context.Context, batchID int64) ([]model.ButcheryItem, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, batch_id, product_id, product_lot_id, weight, is_waste
		 FROM butchery_items WHERE batch_id = ? ORDER BY id`, batchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.ButcheryItem
	for rows.Next() {
		var item model.ButcheryItem
		var productLotID sql.NullInt64
		if err := rows.Scan(&item.ID, &item.BatchID, &item.ProductID, &productLotID, &item.Weight, &item.IsWaste); err != nil {
			return nil, err
		}
		if productLotID.Valid {
			item.ProductLotID = &productLotID.Int64
		}
		out = append(out, item)
	}
	return out, rows.Err()
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
		`SELECT EXISTS(SELECT 1 FROM recall_blocks WHERE lot_id = ?)`, lotID).Scan(&exists)
	return exists, err
}

func (r *mysqlRepo) ListBlockedLotCodes(ctx context.Context) ([]string, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT DISTINCT l.code FROM recall_blocks rb JOIN lots l ON l.id = rb.lot_id ORDER BY l.code`)
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
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO sale_lines (sale_id, lot_id, qty, unit_price, total_price, stock_movement_id) VALUES (?, ?, ?, ?, ?, ?)`,
			sale.ID, line.LotID, line.Qty, line.UnitPrice, line.TotalPrice, nullInt64(line.StockMovementID)); err != nil {
			return model.Sale{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return model.Sale{}, err
	}
	return sale, nil
}
