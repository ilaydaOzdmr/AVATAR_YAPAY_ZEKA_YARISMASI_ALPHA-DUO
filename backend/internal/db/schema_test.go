//go:build integration

package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const defaultRootDSN = "root:root@tcp(localhost:3307)/"

func rootDSN() string {
	if dsn := os.Getenv("TEST_MYSQL_ROOT_DSN"); dsn != "" {
		return dsn
	}
	return defaultRootDSN
}

func TestEnsureSchemaUpgradesOldButcheryTables(t *testing.T) {
	admin, err := sql.Open("mysql", rootDSN())
	if err != nil {
		t.Fatalf("open admin: %v", err)
	}
	defer admin.Close()
	if err := admin.Ping(); err != nil {
		t.Skipf("test mysql not available: %v", err)
	}

	name := fmt.Sprintf("avatar_schema_%d", time.Now().UnixNano())
	if _, err := admin.Exec("CREATE DATABASE " + name); err != nil {
		t.Fatalf("create database: %v", err)
	}
	t.Cleanup(func() { _, _ = admin.Exec("DROP DATABASE " + name) })

	database, err := sql.Open("mysql", rootDSN()+name+"?parseTime=true")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	defer database.Close()

	// Old schema: no input_lot_id and no allocated_cost.
	stmts := []string{
		`CREATE TABLE lots (id BIGINT AUTO_INCREMENT PRIMARY KEY) ENGINE=InnoDB`,
		`CREATE TABLE markets (id BIGINT AUTO_INCREMENT PRIMARY KEY) ENGINE=InnoDB`,
		`CREATE TABLE materials (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			code VARCHAR(64) NOT NULL UNIQUE,
			name VARCHAR(255) NOT NULL,
			unit VARCHAR(32) NOT NULL,
			is_raw BOOL NOT NULL DEFAULT TRUE,
			default_purchase_price DECIMAL(12,2) NOT NULL DEFAULT 0.00
		) ENGINE=InnoDB`,
		`CREATE TABLE products (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			code VARCHAR(64) NOT NULL UNIQUE,
			name VARCHAR(255) NOT NULL,
			unit VARCHAR(32) NOT NULL,
			list_price DECIMAL(12,2) NOT NULL DEFAULT 0.00,
			category VARCHAR(64) NOT NULL DEFAULT 'general',
			is_finished BOOL NOT NULL DEFAULT TRUE
		) ENGINE=InnoDB`,
		`CREATE TABLE production_orders (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			code VARCHAR(64) NOT NULL UNIQUE,
			status VARCHAR(32) NOT NULL DEFAULT 'draft',
			started_at DATETIME NULL,
			finished_at DATETIME NULL,
			notes TEXT NULL,
			service_name VARCHAR(255) NULL,
			service_unit_cost DECIMAL(12,2) NOT NULL DEFAULT 0.00,
			service_qty DECIMAL(12,3) NOT NULL DEFAULT 0.000
		) ENGINE=InnoDB`,
		`CREATE TABLE butchery_batches (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			code VARCHAR(64) NOT NULL UNIQUE,
			animal_type VARCHAR(64) NOT NULL,
			input_weight DECIMAL(12,3) NOT NULL,
			unit_cost DECIMAL(12,2) NOT NULL,
			total_cost DECIMAL(12,2) NOT NULL,
			batch_date DATE NOT NULL,
			status VARCHAR(32) NOT NULL DEFAULT 'draft'
		) ENGINE=InnoDB`,
		`CREATE TABLE butchery_items (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			batch_id BIGINT NOT NULL,
			product_id BIGINT NOT NULL,
			product_lot_id BIGINT NULL,
			weight DECIMAL(12,3) NOT NULL,
			quantity INT NOT NULL DEFAULT 1,
			is_waste BOOL NOT NULL DEFAULT FALSE
		) ENGINE=InnoDB`,
	}
	for _, stmt := range stmts {
		if _, err := database.Exec(stmt); err != nil {
			t.Fatalf("create old schema: %v", err)
		}
	}

	// One legacy batch with a 3 kg unrecorded fire and a 63.50 cost gap.
	if _, err := database.Exec(
		`INSERT INTO butchery_batches (id, code, animal_type, input_weight, unit_cost, total_cost, batch_date, status)
		 VALUES (10, 'KUZ-2026-DEMO', 'kuzu', 25, 11.50, 351, '2026-10-27', 'completed')`); err != nil {
		t.Fatalf("insert legacy batch: %v", err)
	}
	if _, err := database.Exec(
		`INSERT INTO butchery_items (batch_id, product_id, weight, quantity, is_waste)
		 VALUES (10, 1, 10, 1, FALSE), (10, 2, 8, 1, FALSE), (10, 3, 4, 1, FALSE)`); err != nil {
		t.Fatalf("insert legacy items: %v", err)
	}

	ctx := context.Background()
	if err := EnsureSchema(ctx, database); err != nil {
		t.Fatalf("EnsureSchema: %v", err)
	}
	// Must be safe to run again on every boot.
	if err := EnsureSchema(ctx, database); err != nil {
		t.Fatalf("EnsureSchema second run: %v", err)
	}

	for _, col := range []struct{ table, column string }{
		{"butchery_batches", "input_lot_id"},
		{"butchery_items", "allocated_cost"},
		{"butchery_batches", "expected_cost"},
		{"butchery_batches", "cost_difference"},
		{"butchery_batches", "cost_check_status"},
		{"butchery_batches", "weight_check_status"},
		{"materials", "shelf_life_days"},
		{"materials", "created_at"},
		{"products", "shelf_life_days"},
		{"products", "created_at"},
		{"production_orders", "expected_qty"},
	} {
		var exists bool
		if err := database.QueryRow(
			`SELECT EXISTS(SELECT 1 FROM information_schema.columns
			 WHERE table_schema = DATABASE() AND table_name = ? AND column_name = ?)`,
			col.table, col.column).Scan(&exists); err != nil {
			t.Fatalf("check column: %v", err)
		}
		if !exists {
			t.Fatalf("column %s.%s missing after EnsureSchema", col.table, col.column)
		}
	}

	var expectedCost, costDifference float64
	var costStatus, weightStatus string
	if err := database.QueryRow(
		`SELECT expected_cost, cost_difference, cost_check_status, weight_check_status
		 FROM butchery_batches WHERE id = 10`).
		Scan(&expectedCost, &costDifference, &costStatus, &weightStatus); err != nil {
		t.Fatalf("read backfilled checks: %v", err)
	}
	if expectedCost != 287.50 || costDifference != 63.50 {
		t.Fatalf("backfilled cost = %v/%v, want 287.50/63.50", expectedCost, costDifference)
	}
	if costStatus != "mismatch" || weightStatus != "ok" {
		t.Fatalf("backfilled statuses = %q/%q, want mismatch/ok", costStatus, weightStatus)
	}

	var constraint bool
	if err := database.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM information_schema.table_constraints
		 WHERE constraint_schema = DATABASE() AND table_name = 'butchery_batches'
		   AND constraint_name = 'fk_bb_input_lot')`).Scan(&constraint); err != nil {
		t.Fatalf("check constraint: %v", err)
	}
	if !constraint {
		t.Fatal("constraint fk_bb_input_lot missing after EnsureSchema")
	}

	var redistributionTable bool
	if err := database.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM information_schema.tables
		 WHERE table_schema = DATABASE() AND table_name = 'market_redistributions')`).Scan(&redistributionTable); err != nil {
		t.Fatalf("check table: %v", err)
	}
	if !redistributionTable {
		t.Fatal("market_redistributions missing after EnsureSchema")
	}

	var acknowledgementTable bool
	if err := database.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM information_schema.tables
		 WHERE table_schema = DATABASE() AND table_name = 'recall_acknowledgements')`).Scan(&acknowledgementTable); err != nil {
		t.Fatalf("check acknowledgement table: %v", err)
	}
	if !acknowledgementTable {
		t.Fatal("recall_acknowledgements missing after EnsureSchema")
	}

	var anomalyReadsTable bool
	if err := database.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM information_schema.tables
		 WHERE table_schema = DATABASE() AND table_name = 'anomaly_reads')`).Scan(&anomalyReadsTable); err != nil {
		t.Fatalf("check anomaly_reads table: %v", err)
	}
	if !anomalyReadsTable {
		t.Fatal("anomaly_reads missing after EnsureSchema")
	}
}
