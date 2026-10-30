package ml

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHealthSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Fatalf("path = %s, want /health", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	raw, err := NewClient(server.URL).Health(context.Background())
	if err != nil {
		t.Fatalf("Health: %v", err)
	}
	if string(raw) != `{"status":"ok"}` {
		t.Fatalf("body = %s", raw)
	}
}

func TestNon2xxBecomesStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"error":"model missing"}`))
	}))
	defer server.Close()

	_, err := NewClient(server.URL).Models(context.Background())
	var statusErr *StatusError
	if !errors.As(err, &statusErr) {
		t.Fatalf("error = %v, want StatusError", err)
	}
	if statusErr.Status != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", statusErr.Status)
	}
	if string(statusErr.Body) != `{"error":"model missing"}` {
		t.Fatalf("body = %s", statusErr.Body)
	}
}

func TestTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(100 * time.Millisecond)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	_, err := NewClientWithTimeout(server.URL, 10*time.Millisecond).Health(context.Background())
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestPredictForwardsBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/predict/v2/yield" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != `{"raw_material_kg":100}` {
			t.Fatalf("body = %s", body)
		}
		_, _ = w.Write([]byte(`{"prediction":{}}`))
	}))
	defer server.Close()

	raw, err := NewClient(server.URL).Predict(
		context.Background(), "/predict/v2/yield", []byte(`{"raw_material_kg":100}`),
	)
	if err != nil {
		t.Fatalf("Predict: %v", err)
	}
	if string(raw) != `{"prediction":{}}` {
		t.Fatalf("body = %s", raw)
	}
}
