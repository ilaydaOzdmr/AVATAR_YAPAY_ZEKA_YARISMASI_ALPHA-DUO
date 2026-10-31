package repo

import (
	"context"
	"errors"
	"testing"

	"avatar/backend/internal/model"
	"github.com/DATA-DOG/go-sqlmock"
)

func expectLockedDraft(mock sqlmock.Sqlmock) {
	rows := sqlmock.NewRows([]string{"id", "code", "status", "started", "finished", "notes", "service", "cost", "service_qty", "expected"}).
		AddRow(1, "ORDER", "draft", nil, nil, nil, nil, 0, 1, 1)
	mock.ExpectQuery("(?s)SELECT.*FROM production_orders").WillReturnRows(rows)
}

func expectOrderInputs(mock sqlmock.Sqlmock) {
	mock.ExpectQuery("(?s)SELECT.*FROM production_order_inputs").WillReturnRows(
		sqlmock.NewRows([]string{"id", "order", "lot", "qty"}).AddRow(1, 1, 1, 1))
}

func expectOrderOutputs(mock sqlmock.Sqlmock) {
	mock.ExpectQuery("(?s)SELECT.*FROM production_order_outputs").WillReturnRows(
		sqlmock.NewRows([]string{"id", "order", "lot", "qty"}).AddRow(1, 1, 2, 1))
}

func TestCompleteOrderRollsBackAtEveryStorageBoundary(t *testing.T) {
	for _, failure := range []string{"begin", "order", "inputs", "outputs", "ledger", "apply", "update", "commit"} {
		t.Run(failure, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			if failure == "begin" {
				mock.ExpectBegin().WillReturnError(errSQLFault)
			} else {
				mock.ExpectBegin()
				configureCompletionFault(mock, failure)
				if failure != "commit" {
					mock.ExpectRollback()
				}
			}
			result, err := NewMySQLRepo(db).CompleteProductionOrder(context.Background(), 1, nil)
			if !errors.Is(err, errSQLFault) {
				t.Fatalf("expected storage failure, got %v", err)
			}
			if result.ID != 0 {
				t.Fatalf("failed transaction returned accepted order: %+v", result)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func configureCompletionFault(mock sqlmock.Sqlmock, failure string) {
	if failure == "order" {
		mock.ExpectQuery("(?s)SELECT.*FROM production_orders").WillReturnError(errSQLFault)
		return
	}
	expectLockedDraft(mock)
	if failure == "inputs" {
		mock.ExpectQuery("(?s)SELECT.*FROM production_order_inputs").WillReturnError(errSQLFault)
		return
	}
	expectOrderInputs(mock)
	if failure == "outputs" {
		mock.ExpectQuery("(?s)SELECT.*FROM production_order_outputs").WillReturnError(errSQLFault)
		return
	}
	expectOrderOutputs(mock)
	if failure == "ledger" {
		mock.ExpectQuery("SELECT COUNT").WillReturnError(errSQLFault)
		return
	}
	if failure == "apply" {
		mock.ExpectQuery("SELECT COUNT").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		mock.ExpectQuery("SELECT unit_cost").WillReturnError(errSQLFault)
		return
	}
	mock.ExpectQuery("SELECT COUNT").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	if failure == "update" {
		mock.ExpectExec("UPDATE production_orders").WillReturnError(errSQLFault)
		return
	}
	mock.ExpectExec("UPDATE production_orders").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit().WillReturnError(errSQLFault)
}

func TestCreateBatchRollbackPreservesSourceStock(t *testing.T) {
	id := int64(1)
	for _, failure := range []string{"begin", "input", "insert", "id", "draft", "parts", "commit"} {
		t.Run(failure, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			batch := model.ButcheryBatch{Code: "BATCH", InputWeight: 20, UnitCost: 1, InputLotID: &id}
			items := []model.ButcheryItem{{ProductID: 2, Weight: 10}}
			if failure == "draft" || failure == "commit" {
				batch.InputLotID = nil
			}
			configureBatchFault(mock, failure)
			result, err := NewMySQLRepo(db).CreateButcheryBatch(context.Background(), batch, items)
			if !errors.Is(err, errSQLFault) {
				t.Fatalf("expected storage failure, got %v", err)
			}
			if result.ID != 0 {
				t.Fatal("failed batch returned as accepted")
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func configureBatchFault(mock sqlmock.Sqlmock, failure string) {
	if failure == "begin" {
		mock.ExpectBegin().WillReturnError(errSQLFault)
		return
	}
	mock.ExpectBegin()
	if failure == "input" {
		mock.ExpectQuery("SELECT item_type, qty").WillReturnError(errSQLFault)
	} else {
		if failure != "draft" && failure != "commit" {
			mock.ExpectQuery("SELECT item_type, qty").WillReturnRows(sqlmock.NewRows([]string{"type", "qty"}).AddRow("material", 20))
		}
		configureBatchWriteFault(mock, failure)
	}
	if failure != "commit" {
		mock.ExpectRollback()
	}
}

func configureBatchWriteFault(mock sqlmock.Sqlmock, failure string) {
	if failure == "insert" {
		mock.ExpectExec("INSERT INTO butchery_batches").WillReturnError(errSQLFault)
		return
	}
	if failure == "id" {
		mock.ExpectExec("INSERT INTO butchery_batches").WillReturnResult(sqlmock.NewErrorResult(errSQLFault))
		return
	}
	mock.ExpectExec("INSERT INTO butchery_batches").WillReturnResult(sqlmock.NewResult(1, 1))
	switch failure {
	case "draft":
		mock.ExpectExec("INSERT INTO butchery_items").WillReturnError(errSQLFault)
	case "parts":
		mock.ExpectQuery("SELECT code, shelf_life_days").WillReturnError(errSQLFault)
	case "commit":
		mock.ExpectExec("INSERT INTO butchery_items").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit().WillReturnError(errSQLFault)
	}
}
