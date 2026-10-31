package repo

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"avatar/backend/internal/model"
	"github.com/DATA-DOG/go-sqlmock"
)

var errSQLFault = errors.New("simulated storage failure")

// Real MySQL tests cover valid stock transitions. These driver fault cases
// verify that storage failures propagate and the transaction is rolled back.
func TestTransactionHelpersPropagateStorageFailures(t *testing.T) {
	ctx := context.Background()
	id := int64(1)
	batch := model.ButcheryBatch{ID: 1, Code: "BATCH", InputLotID: &id, InputWeight: 20, BatchDate: "2026-10-31"}
	item := model.ButcheryItem{ProductID: 2, ProductLotID: &id, Weight: 10}
	cases := []struct {
		name  string
		setup func(sqlmock.Sqlmock)
		call  func(*sql.Tx) error
	}{
		{"insert input", func(m sqlmock.Sqlmock) {
			m.ExpectExec("INSERT INTO production_order_inputs").WillReturnError(errSQLFault)
		},
			func(tx *sql.Tx) error {
				return insertProductionItemsTx(ctx, tx, model.ProductionOrder{}, []model.ProductionOrderInput{{Qty: 1}}, nil)
			}},
		{"insert output", func(m sqlmock.Sqlmock) {
			m.ExpectExec("INSERT INTO production_order_outputs").WillReturnError(errSQLFault)
		},
			func(tx *sql.Tx) error {
				return insertProductionItemsTx(ctx, tx, model.ProductionOrder{}, nil, []model.ProductionOrderOutput{{Qty: 1}})
			}},
		{"locked order", func(m sqlmock.Sqlmock) {
			m.ExpectQuery("(?s)SELECT.*FROM production_orders").WillReturnError(errSQLFault)
		},
			func(tx *sql.Tx) error { _, err := productionOrderForUpdate(ctx, tx, 1); return err }},
		{"input query", func(m sqlmock.Sqlmock) {
			m.ExpectQuery("(?s)SELECT.*FROM production_order_inputs").WillReturnError(errSQLFault)
		},
			func(tx *sql.Tx) error { _, err := productionInputsTx(ctx, tx, 1); return err }},
		{"output query", func(m sqlmock.Sqlmock) {
			m.ExpectQuery("(?s)SELECT.*FROM production_order_outputs").WillReturnError(errSQLFault)
		},
			func(tx *sql.Tx) error { _, err := productionOutputsTx(ctx, tx, 1); return err }},
		{"input stream", func(m sqlmock.Sqlmock) {
			m.ExpectQuery("(?s)SELECT.*FROM production_order_inputs").WillReturnRows(sqlmock.NewRows([]string{"id", "order", "lot", "qty"}).AddRow(1, 1, 1, 1).RowError(0, errSQLFault))
		},
			func(tx *sql.Tx) error { _, err := productionInputsTx(ctx, tx, 1); return err }},
		{"output stream", func(m sqlmock.Sqlmock) {
			m.ExpectQuery("(?s)SELECT.*FROM production_order_outputs").WillReturnRows(sqlmock.NewRows([]string{"id", "order", "lot", "qty"}).AddRow(1, 1, 1, 1).RowError(0, errSQLFault))
		},
			func(tx *sql.Tx) error { _, err := productionOutputsTx(ctx, tx, 1); return err }},
		{"output shelf life read", func(m sqlmock.Sqlmock) { m.ExpectQuery("(?s)SELECT p.shelf_life_days").WillReturnError(errSQLFault) },
			func(tx *sql.Tx) error {
				return stockProductionOutputTx(ctx, tx, model.ProductionOrder{}, model.ProductionOrderOutput{}, 10, "2026-10-31")
			}},
		{"output expiry write", func(m sqlmock.Sqlmock) {
			m.ExpectQuery("(?s)SELECT p.shelf_life_days").WillReturnRows(sqlmock.NewRows([]string{"days"}).AddRow(10))
			m.ExpectExec("UPDATE lots").WillReturnError(errSQLFault)
		},
			func(tx *sql.Tx) error {
				return stockProductionOutputTx(ctx, tx, model.ProductionOrder{}, model.ProductionOrderOutput{}, 10, "2026-10-31")
			}},
		{"output cost write", func(m sqlmock.Sqlmock) {
			m.ExpectQuery("(?s)SELECT p.shelf_life_days").WillReturnRows(sqlmock.NewRows([]string{"days"}).AddRow(0))
			m.ExpectExec("UPDATE lots").WillReturnError(errSQLFault)
		},
			func(tx *sql.Tx) error {
				return stockProductionOutputTx(ctx, tx, model.ProductionOrder{}, model.ProductionOrderOutput{}, 10, "2026-10-31")
			}},
		{"production input movement", func(m sqlmock.Sqlmock) { m.ExpectQuery("(?s)SELECT.*FROM lots").WillReturnError(errSQLFault) },
			func(tx *sql.Tx) error {
				return consumeProductionInputTx(ctx, tx, model.ProductionOrder{}, model.ProductionOrderInput{MaterialLotID: 1, Qty: 1}, nil, 1, "2026-10-31")
			}},
		{"butchery input read", func(m sqlmock.Sqlmock) { m.ExpectQuery("SELECT item_type, qty FROM lots").WillReturnError(errSQLFault) },
			func(tx *sql.Tx) error { return lockButcheryInputTx(ctx, tx, batch) }},
		{"draft item write", func(m sqlmock.Sqlmock) { m.ExpectExec("INSERT INTO butchery_items").WillReturnError(errSQLFault) },
			func(tx *sql.Tx) error {
				return insertDraftButcheryItemsTx(ctx, tx, batch, []model.ButcheryItem{item}, []float64{100})
			}},
		{"waste write", func(m sqlmock.Sqlmock) { m.ExpectExec("INSERT INTO butchery_items").WillReturnError(errSQLFault) },
			func(tx *sql.Tx) error {
				return stockButcheryItemTx(ctx, tx, batch, model.ButcheryItem{IsWaste: true}, 0)
			}},
		{"part card read", func(m sqlmock.Sqlmock) {
			m.ExpectQuery("SELECT code, shelf_life_days FROM products").WillReturnError(errSQLFault)
		},
			func(tx *sql.Tx) error { return stockButcheryItemTx(ctx, tx, batch, item, 100) }},
		{"existing part lot read", func(m sqlmock.Sqlmock) { m.ExpectQuery("SELECT item_type FROM lots").WillReturnError(errSQLFault) },
			func(tx *sql.Tx) error { _, err := existingButcheryLotTx(ctx, tx, item); return err }},
		{"part lot existence", func(m sqlmock.Sqlmock) { m.ExpectQuery("SELECT EXISTS").WillReturnError(errSQLFault) },
			func(tx *sql.Tx) error {
				_, err := newButcheryLotTx(ctx, tx, batch, item, "PRD", "2026-10-31", nil)
				return err
			}},
		{"new part lot insert", func(m sqlmock.Sqlmock) {
			m.ExpectQuery("SELECT EXISTS").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			m.ExpectExec("INSERT INTO lots").WillReturnError(errSQLFault)
		},
			func(tx *sql.Tx) error {
				_, err := newButcheryLotTx(ctx, tx, batch, item, "PRD", "2026-10-31", nil)
				return err
			}},
		{"part lot generated id", func(m sqlmock.Sqlmock) {
			m.ExpectQuery("SELECT EXISTS").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			m.ExpectExec("INSERT INTO lots").WillReturnResult(sqlmock.NewErrorResult(errSQLFault))
		},
			func(tx *sql.Tx) error {
				_, err := newButcheryLotTx(ctx, tx, batch, item, "PRD", "2026-10-31", nil)
				return err
			}},
		{"sale block check", func(m sqlmock.Sqlmock) { m.ExpectQuery("(?s)SELECT EXISTS").WillReturnError(errSQLFault) },
			func(tx *sql.Tx) error { return createSaleLineTx(ctx, tx, model.Sale{}, model.SaleLine{}) }},
		{"sale movement read", func(m sqlmock.Sqlmock) {
			m.ExpectQuery("(?s)SELECT EXISTS").WillReturnRows(sqlmock.NewRows([]string{"blocked"}).AddRow(false))
			m.ExpectQuery("(?s)SELECT.*FROM lots").WillReturnError(errSQLFault)
		},
			func(tx *sql.Tx) error {
				return createSaleLineTx(ctx, tx, model.Sale{}, model.SaleLine{LotID: 1, Qty: 1})
			}},
		{"butchery consumption", func(m sqlmock.Sqlmock) { m.ExpectQuery("(?s)SELECT.*FROM lots").WillReturnError(errSQLFault) },
			func(tx *sql.Tx) error { return consumeButcheryInputTx(ctx, tx, batch, 0) }},
		{"butchery waste movement", func(m sqlmock.Sqlmock) { m.ExpectQuery("(?s)SELECT.*FROM lots").WillReturnError(errSQLFault) },
			func(tx *sql.Tx) error { return consumeButcheryInputTx(ctx, tx, batch, 20) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			mock.ExpectBegin()
			tx, err := db.Begin()
			if err != nil {
				t.Fatal(err)
			}
			tc.setup(mock)
			mock.ExpectRollback()
			if err := tc.call(tx); !errors.Is(err, errSQLFault) {
				t.Errorf("expected storage error, got %v", err)
			}
			if err := tx.Rollback(); err != nil {
				t.Fatal(err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
