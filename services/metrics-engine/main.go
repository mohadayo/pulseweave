package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type Metric struct {
	Service   string    `json:"service"`
	Name      string    `json:"name"`
	Value     float64   `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}

type MetricsStore struct {
	mu      sync.RWMutex
	metrics []Metric
}

var store = &MetricsStore{}

func (s *MetricsStore) Add(m Metric) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m.Timestamp = time.Now()
	s.metrics = append(s.metrics, m)
}

func (s *MetricsStore) List(service string) []Metric {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if service == "" {
		result := make([]Metric, len(s.metrics))
		copy(result, s.metrics)
		return result
	}
	var result []Metric
	for _, m := range s.metrics {
		if m.Service == service {
			result = append(result, m)
		}
	}
	return result
}

func (s *MetricsStore) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.metrics = nil
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "metrics-engine",
	})
}

func ingestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var m Metric
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		log.Printf("[WARN] Invalid metric payload: %v", err)
		http.Error(w, `{"error":"invalid payload"}`, http.StatusBadRequest)
		return
	}

	if m.Service == "" || m.Name == "" {
		log.Printf("[WARN] Missing required fields in metric")
		http.Error(w, `{"error":"service and name are required"}`, http.StatusBadRequest)
		return
	}

	store.Add(m)
	log.Printf("[INFO] Metric ingested: service=%s name=%s value=%.2f", m.Service, m.Name, m.Value)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "metric ingested"})
}

func queryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	service := r.URL.Query().Get("service")
	metrics := store.List(service)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"count":   len(metrics),
		"metrics": metrics,
	})
}

func main() {
	port := os.Getenv("METRICS_ENGINE_PORT")
	if port == "" {
		port = "5002"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/ingest", ingestHandler)
	mux.HandleFunc("/query", queryHandler)

	log.Printf("[INFO] Starting metrics-engine on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("[FATAL] Server failed: %v", err)
	}
}
