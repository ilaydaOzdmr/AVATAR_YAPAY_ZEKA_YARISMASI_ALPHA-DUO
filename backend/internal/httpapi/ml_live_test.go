package httpapi

import (
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"avatar/backend/internal/ml"
)

// Optional read-only smoke test against an actual loaded Python model service.
// Set AVATAR_ML_TEST_URL explicitly; normal unit tests never need a running service.
func TestMLLiveWebContract(t *testing.T) {
	baseURL := os.Getenv("AVATAR_ML_TEST_URL")
	if baseURL == "" {
		t.Skip("AVATAR_ML_TEST_URL is not set")
	}
	handler := NewServerWithML(&fakeRepo{}, ml.NewClient(baseURL))
	call := func(method, path, body string) map[string]any {
		t.Helper()
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(method, path, strings.NewReader(body)))
		if recorder.Code != http.StatusOK {
			t.Fatalf("%s returned %d: %s", path, recorder.Code, recorder.Body.String())
		}
		var result map[string]any
		if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		return result
	}
	health := call(http.MethodGet, "/ml/health", "")
	if health["models_loaded"] != float64(3) || health["status"] != "ok" {
		t.Fatalf("models unavailable: %v", health)
	}
	report := call(http.MethodGet, "/ml/report", "")
	if len(report["model_reports"].(map[string]any)) != 3 {
		t.Fatal("missing model reports")
	}
	production := call(http.MethodPost, "/ml/predict/v2/yield", `{"product_type":"sucuk","raw_material_kg":100,"spice_kg":2,"casing_kg":5,"service_unit_cost":5.5,"actual_finished_qty":85}`)
	if math.Abs(production["accounting_ekp"].(float64)-18.1588) > 0.0001 {
		t.Fatal("PDF accounting mismatch")
	}
	for _, animal := range []string{"lamb", "cattle"} {
		weight := 25.0
		if animal == "cattle" {
			weight = 300
		}
		body, err := json.Marshal(map[string]any{"animal_type": animal, "carcass_weight_kg": weight})
		if err != nil {
			t.Fatal(err)
		}
		result := call(http.MethodPost, "/ml/predict/v2/butchery", string(body))
		physical := result["physical_check"].(map[string]any)
		if physical["parts_sum_kg"] != weight || physical["parts_over_carcass"] != false {
			t.Fatalf("invalid carcass balance for %s: %v", animal, physical)
		}
	}
}
