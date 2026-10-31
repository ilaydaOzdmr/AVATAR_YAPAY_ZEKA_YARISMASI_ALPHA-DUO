package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"avatar/backend/internal/model"
)

type outputRepo struct {
	*fakeRepo
	outputs []model.ProductionOrderOutput
}

func (r *outputRepo) GetProductionOrder(_ context.Context, id int64) (model.ProductionOrder, error) {
	if id != 1 {
		return model.ProductionOrder{}, sql.ErrNoRows
	}
	return model.ProductionOrder{ID: id}, nil
}

func (r *outputRepo) ListProductionOrderOutputs(_ context.Context, _ int64) ([]model.ProductionOrderOutput, error) {
	return r.outputs, nil
}

func TestProductionOutputsExposeProductLotForMLUnitMapping(t *testing.T) {
	repository := &outputRepo{fakeRepo: &fakeRepo{}, outputs: []model.ProductionOrderOutput{{ID: 1, ProductionOrderID: 1, ProductLotID: 4, Qty: 85}}}
	handler := NewServer(repository)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/production-orders/1/outputs", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d", recorder.Code)
	}
	var outputs []model.ProductionOrderOutput
	if err := json.Unmarshal(recorder.Body.Bytes(), &outputs); err != nil || len(outputs) != 1 || outputs[0].ProductLotID != 4 || outputs[0].Qty != 85 {
		t.Fatalf("outputs = %s, error = %v", recorder.Body.String(), err)
	}
	for path, want := range map[string]int{"/production-orders/bad/outputs": 400, "/production-orders/0/outputs": 400, "/production-orders/2/outputs": 404} {
		recorder = httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		if recorder.Code != want {
			t.Errorf("%s: status %d, want %d", path, recorder.Code, want)
		}
	}
	repository.outputs = nil
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/production-orders/1/outputs", nil))
	if recorder.Body.String() != "[]\n" {
		t.Errorf("empty outputs = %s", recorder.Body.String())
	}
}
