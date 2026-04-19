package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setup() {
	store.Clear()
}

func TestHealthHandler(t *testing.T) {
	setup()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	healthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["status"] != "ok" {
		t.Fatalf("expected status ok, got %s", resp["status"])
	}
	if resp["service"] != "metrics-engine" {
		t.Fatalf("expected service metrics-engine, got %s", resp["service"])
	}
}

func TestIngestHandler_Success(t *testing.T) {
	setup()
	body, _ := json.Marshal(map[string]interface{}{
		"service": "auth-service",
		"name":    "request_count",
		"value":   42.0,
	})
	req := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewReader(body))
	w := httptest.NewRecorder()
	ingestHandler(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	metrics := store.List("")
	if len(metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(metrics))
	}
}

func TestIngestHandler_MissingFields(t *testing.T) {
	setup()
	body, _ := json.Marshal(map[string]interface{}{
		"value": 42.0,
	})
	req := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewReader(body))
	w := httptest.NewRecorder()
	ingestHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestIngestHandler_InvalidJSON(t *testing.T) {
	setup()
	req := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewReader([]byte("invalid")))
	w := httptest.NewRecorder()
	ingestHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestIngestHandler_WrongMethod(t *testing.T) {
	setup()
	req := httptest.NewRequest(http.MethodGet, "/ingest", nil)
	w := httptest.NewRecorder()
	ingestHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestQueryHandler_All(t *testing.T) {
	setup()
	body1, _ := json.Marshal(map[string]interface{}{"service": "svc-a", "name": "cpu", "value": 10.0})
	body2, _ := json.Marshal(map[string]interface{}{"service": "svc-b", "name": "mem", "value": 20.0})

	req1 := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewReader(body1))
	w1 := httptest.NewRecorder()
	ingestHandler(w1, req1)

	req2 := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewReader(body2))
	w2 := httptest.NewRecorder()
	ingestHandler(w2, req2)

	req := httptest.NewRequest(http.MethodGet, "/query", nil)
	w := httptest.NewRecorder()
	queryHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	count := int(resp["count"].(float64))
	if count != 2 {
		t.Fatalf("expected 2 metrics, got %d", count)
	}
}

func TestQueryHandler_FilterByService(t *testing.T) {
	setup()
	body1, _ := json.Marshal(map[string]interface{}{"service": "svc-a", "name": "cpu", "value": 10.0})
	body2, _ := json.Marshal(map[string]interface{}{"service": "svc-b", "name": "mem", "value": 20.0})

	req1 := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewReader(body1))
	ingestHandler(httptest.NewRecorder(), req1)
	req2 := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewReader(body2))
	ingestHandler(httptest.NewRecorder(), req2)

	req := httptest.NewRequest(http.MethodGet, "/query?service=svc-a", nil)
	w := httptest.NewRecorder()
	queryHandler(w, req)

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	count := int(resp["count"].(float64))
	if count != 1 {
		t.Fatalf("expected 1 metric, got %d", count)
	}
}

func TestQueryHandler_WrongMethod(t *testing.T) {
	setup()
	req := httptest.NewRequest(http.MethodPost, "/query", nil)
	w := httptest.NewRecorder()
	queryHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}
