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
	return ensureConstraint(ctx, database, "butchery_batches", "fk_bb_input_lot",
		`ALTER TABLE butchery_batches
		 ADD CONSTRAINT fk_bb_input_lot FOREIGN KEY (input_lot_id) REFERENCES lots(id)`)
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
