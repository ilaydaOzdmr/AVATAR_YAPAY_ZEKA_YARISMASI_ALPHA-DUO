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
}
