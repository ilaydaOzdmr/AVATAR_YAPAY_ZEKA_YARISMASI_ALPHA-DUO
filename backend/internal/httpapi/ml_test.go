package httpapi

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"avatar/backend/internal/ml"
)

func TestMLProxyHealthAndModels(t *testing.T) {
	mlServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/health":
			_, _ = w.Write([]byte(`{"status":"ok","models_loaded":3}`))
		case "/models":
			_, _ = w.Write([]byte(`[{"name":"yield"}]`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer mlServer.Close()

	handler := NewServerWithML(&fakeRepo{}, ml.NewClient(mlServer.URL))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ml/health", nil))
	if rec.Code != http.StatusOK || rec.Body.String() != `{"status":"ok","models_loaded":3}` {
		t.Fatalf("health = %d %s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ml/models", nil))
	if rec.Code != http.StatusOK || rec.Body.String() != `[{"name":"yield"}]` {
		t.Fatalf("models = %d %s", rec.Code, rec.Body.String())
	}
}

func TestMLProxyForwardsPredictionBody(t *testing.T) {
	var received string
	mlServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/predict/v2/yield" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		raw, _ := io.ReadAll(r.Body)
		received = string(raw)
		_, _ = w.Write([]byte(`{"prediction":{"finished_qty":80.0},"model_status":"synthetic_only"}`))
	}))
	defer mlServer.Close()

	handler := NewServerWithML(&fakeRepo{}, ml.NewClient(mlServer.URL))
	body := bytes.NewBufferString(`{"raw_material_kg":100}`)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/ml/predict/v2/yield", body))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	if received != `{"raw_material_kg":100}` {
		t.Fatalf("forwarded body = %s", received)
	}
}

func TestMLProxyPassesThrough503(t *testing.T) {
	mlServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"error":"model missing","model_status":"unavailable"}`))
	}))
	defer mlServer.Close()

	handler := NewServerWithML(&fakeRepo{}, ml.NewClient(mlServer.URL))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ml/report", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}
	if rec.Body.String() != `{"error":"model missing","model_status":"unavailable"}` {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestMLProxyTimeoutReturns504(t *testing.T) {
	mlServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(100 * time.Millisecond)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer mlServer.Close()

	client := ml.NewClientWithTimeout(mlServer.URL, 10*time.Millisecond)
	handler := NewServerWithML(&fakeRepo{}, client)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ml/health", nil))

	if rec.Code != http.StatusGatewayTimeout {
		t.Fatalf("status = %d, want 504", rec.Code)
	}
}

func TestMLProxyWithoutClientIs503(t *testing.T) {
	handler := NewServerWithML(&fakeRepo{}, nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ml/health", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}
}
