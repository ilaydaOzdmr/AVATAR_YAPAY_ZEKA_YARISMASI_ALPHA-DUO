package db

import (
	"context"
	"database/sql"
	"fmt"
)

// EnsureSchema applies idempotent upgrades for databases whose volume was
// created before a column or constraint was added to the init scripts. Fresh
// databases already have everything, so each step is a cheap existence check.
func EnsureSchema(ctx context.Context, database *sql.DB) error {
	if err := ensureColumn(ctx, database, "butchery_batches", "input_lot_id",
		`ALTER TABLE butchery_batches ADD COLUMN input_lot_id BIGINT NULL`); err != nil {
		return err
	}
	if err := ensureColumn(ctx, database, "butchery_items", "allocated_cost",
		`ALTER TABLE butchery_items ADD COLUMN allocated_cost DECIMAL(12,2) NOT NULL DEFAULT 0.00`); err != nil {
		return err
	}
	for _, column := range []struct{ name, ddl string }{
		{"expected_cost", `ALTER TABLE butchery_batches ADD COLUMN expected_cost DECIMAL(12,2) NOT NULL DEFAULT 0.00 AFTER status`},
		{"cost_difference", `ALTER TABLE butchery_batches ADD COLUMN cost_difference DECIMAL(12,2) NOT NULL DEFAULT 0.00 AFTER expected_cost`},
		{"cost_check_status", `ALTER TABLE butchery_batches ADD COLUMN cost_check_status VARCHAR(16) NOT NULL DEFAULT 'ok' AFTER cost_difference`},
		{"weight_check_status", `ALTER TABLE butchery_batches ADD COLUMN weight_check_status VARCHAR(16) NOT NULL DEFAULT 'ok' AFTER cost_check_status`},
	} {
		if err := ensureColumn(ctx, database, "butchery_batches", column.name, column.ddl); err != nil {
			return err
		}
	}
	if err := backfillButcheryChecks(ctx, database); err != nil {
		return err
	}
	if err := ensureColumn(ctx, database, "production_orders", "expected_qty",
		`ALTER TABLE production_orders ADD COLUMN expected_qty DECIMAL(12,3) NOT NULL DEFAULT 0.000 AFTER service_qty`); err != nil {
		return err
	}
	if err := ensureColumn(ctx, database, "materials", "shelf_life_days",
		`ALTER TABLE materials ADD COLUMN shelf_life_days INT NOT NULL DEFAULT 0`); err != nil {
		return err
	}
	if err := ensureColumn(ctx, database, "materials", "created_at",
		`ALTER TABLE materials ADD COLUMN created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP`); err != nil {
		return err
	}
	if err := ensureColumn(ctx, database, "products", "shelf_life_days",
		`ALTER TABLE products ADD COLUMN shelf_life_days INT NOT NULL DEFAULT 0`); err != nil {
		return err
	}
	if err := ensureColumn(ctx, database, "products", "created_at",
		`ALTER TABLE products ADD COLUMN created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP`); err != nil {
		return err
	}
	if err := ensureTable(ctx, database, "market_redistributions", `CREATE TABLE IF NOT EXISTS market_redistributions (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		market_id BIGINT NOT NULL,
		lot_id BIGINT NOT NULL,
		sub_buyer VARCHAR(255) NOT NULL,
		contact VARCHAR(255) NULL,
		qty DECIMAL(12,3) NOT NULL,
		CONSTRAINT fk_mr_market FOREIGN KEY (market_id) REFERENCES markets(id),
		CONSTRAINT fk_mr_lot FOREIGN KEY (lot_id) REFERENCES lots(id),
		INDEX idx_mr_market (market_id),
		INDEX idx_mr_lot (lot_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`); err != nil {
		return err
	}
	if err := ensureTable(ctx, database, "recall_acknowledgements", `CREATE TABLE IF NOT EXISTS recall_acknowledgements (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		source_lot_id BIGINT NOT NULL,
		market_code VARCHAR(64) NOT NULL,
		confirmed_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		CONSTRAINT fk_ra_source_lot FOREIGN KEY (source_lot_id) REFERENCES lots(id),
		UNIQUE KEY uq_ra_source_market (source_lot_id, market_code)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`); err != nil {
		return err
	}
	if err := ensureTable(ctx, database, "anomaly_reads", `CREATE TABLE IF NOT EXISTS anomaly_reads (
		alert_key VARCHAR(191) PRIMARY KEY,
		read_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`); err != nil {
		return err
	}
	return ensureConstraint(ctx, database, "butchery_batches", "fk_bb_input_lot",
		`ALTER TABLE butchery_batches
		 ADD CONSTRAINT fk_bb_input_lot FOREIGN KEY (input_lot_id) REFERENCES lots(id)`)
}

// backfillButcheryChecks recalculates the persisted consistency fields for
// batches created before the columns existed. It is idempotent and also keeps
// the values correct if items were edited outside the normal flow.
func backfillButcheryChecks(ctx context.Context, database *sql.DB) error {
	_, err := database.ExecContext(ctx, `
		UPDATE butchery_batches b
		SET b.expected_cost = ROUND(b.input_weight * b.unit_cost, 2),
		    b.cost_difference = ROUND(b.total_cost - ROUND(b.input_weight * b.unit_cost, 2), 2),
		    b.cost_check_status = CASE
		        WHEN ROUND(b.total_cost, 2) = ROUND(b.input_weight * b.unit_cost, 2) THEN 'ok'
		        ELSE 'mismatch'
		    END,
		    b.weight_check_status = CASE
		        WHEN ROUND(b.input_weight - COALESCE((
		            SELECT SUM(i.weight) FROM butchery_items i
		            WHERE i.batch_id = b.id AND i.is_waste = FALSE
		        ), 0), 3) < 0
		        OR ROUND(b.input_weight - COALESCE((
		            SELECT SUM(i.weight) FROM butchery_items i
		            WHERE i.batch_id = b.id AND i.is_waste = FALSE
		        ), 0), 3) > ROUND(b.input_weight * 0.20, 3) THEN 'mismatch'
		        ELSE 'ok'
		    END`)
	if err != nil {
		return fmt.Errorf("backfill butchery checks: %w", err)
	}
	return nil
}

func ensureTable(ctx context.Context, database *sql.DB, table, ddl string) error {
	var exists bool
	if err := database.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM information_schema.tables
		 WHERE table_schema = DATABASE() AND table_name = ?)`,
		table).Scan(&exists); err != nil {
		return fmt.Errorf("check table %s: %w", table, err)
	}
	if exists {
		return nil
	}
	if _, err := database.ExecContext(ctx, ddl); err != nil {
		return fmt.Errorf("create table %s: %w", table, err)
	}
	return nil
}

func ensureColumn(ctx context.Context, database *sql.DB, table, column, ddl string) error {
	var exists bool
	if err := database.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM information_schema.columns
		 WHERE table_schema = DATABASE() AND table_name = ? AND column_name = ?)`,
		table, column).Scan(&exists); err != nil {
		return fmt.Errorf("check column %s.%s: %w", table, column, err)
	}
	if exists {
		return nil
	}
	if _, err := database.ExecContext(ctx, ddl); err != nil {
		return fmt.Errorf("add column %s.%s: %w", table, column, err)
	}
	return nil
}

func ensureConstraint(ctx context.Context, database *sql.DB, table, constraint, ddl string) error {
	var exists bool
	if err := database.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM information_schema.table_constraints
		 WHERE constraint_schema = DATABASE() AND table_name = ? AND constraint_name = ?)`,
		table, constraint).Scan(&exists); err != nil {
		return fmt.Errorf("check constraint %s: %w", constraint, err)
	}
	if exists {
		return nil
	}
	if _, err := database.ExecContext(ctx, ddl); err != nil {
		return fmt.Errorf("add constraint %s: %w", constraint, err)
	}
	return nil
}
