// Package main provides a mock API server for testing jprobe.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"
)

var (
	healthy     atomic.Bool
	ready       atomic.Bool
	requestCount atomic.Int64
)

func init() {
	healthy.Store(true)
	ready.Store(true)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", healthHandler)

	// Readiness check endpoint
	mux.HandleFunc("/ready", readyHandler)

	// Echo endpoint - returns request info
	mux.HandleFunc("/echo", echoHandler)

	// Slow endpoint - simulates slow response
	mux.HandleFunc("/slow", slowHandler)

	// Error endpoint - returns error
	mux.HandleFunc("/error", errorHandler)

	// Control endpoints for testing
	mux.HandleFunc("/control/health", controlHealthHandler)
	mux.HandleFunc("/control/ready", controlReadyHandler)

	// Stats endpoint
	mux.HandleFunc("/stats", statsHandler)

	// Auth test endpoint
	mux.HandleFunc("/api/v1/auth/validate", authValidateHandler)

	// Additional endpoints for comprehensive testing
	mux.HandleFunc("/live", liveHandler)
	mux.HandleFunc("/metrics", metricsHandler)
	mux.HandleFunc("/version", versionHandler)
	mux.HandleFunc("/api/v1/status", apiStatusHandler)
	mux.HandleFunc("/api/v1/users/health", serviceHealthHandler("users"))
	mux.HandleFunc("/api/v1/orders/health", serviceHealthHandler("orders"))
	mux.HandleFunc("/api/v1/payments/health", serviceHealthHandler("payments"))
	mux.HandleFunc("/api/v1/inventory/health", serviceHealthHandler("inventory"))
	mux.HandleFunc("/api/v1/notifications/health", serviceHealthHandler("notifications"))
	mux.HandleFunc("/api/v1/search/health", serviceHealthHandler("search"))

	log.Printf("Mock API server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, loggingMiddleware(mux)); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestCount.Add(1)
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if !healthy.Load() {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "unhealthy",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
	})
}

func readyHandler(w http.ResponseWriter, r *http.Request) {
	if !ready.Load() {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ready":    false,
			"database": "disconnected",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ready":    true,
		"database": "connected",
	})
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"method":  r.Method,
		"path":    r.URL.Path,
		"query":   r.URL.Query(),
		"headers": r.Header,
	}

	if r.Method == http.MethodPost || r.Method == http.MethodPut {
		var body interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
			response["body"] = body
		}
	}

	json.NewEncoder(w).Encode(response)
}

func slowHandler(w http.ResponseWriter, r *http.Request) {
	delay := r.URL.Query().Get("delay")
	if delay == "" {
		delay = "2s"
	}

	d, err := time.ParseDuration(delay)
	if err != nil {
		d = 2 * time.Second
	}

	time.Sleep(d)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"delayed": d.String(),
	})
}

func errorHandler(w http.ResponseWriter, r *http.Request) {
	code := http.StatusInternalServerError
	if codeParam := r.URL.Query().Get("code"); codeParam != "" {
		if c, err := parseInt(codeParam); err == nil && c >= 400 && c < 600 {
			code = c
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   true,
		"code":    code,
		"message": http.StatusText(code),
	})
}

func controlHealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Healthy bool `json:"healthy"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	healthy.Store(body.Healthy)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{
		"healthy": healthy.Load(),
	})
}

func controlReadyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Ready bool `json:"ready"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ready.Store(body.Ready)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{
		"ready": ready.Load(),
	})
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"requests": requestCount.Load(),
		"healthy":  healthy.Load(),
		"ready":    ready.Load(),
	})
}

func authValidateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Simple token validation - any non-empty token is valid
	if body.Token != "" {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"valid": true,
			"user":  "test-user",
		})
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"valid": false,
			"error": "invalid token",
		})
	}
}

func parseInt(s string) (int, error) {
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, nil
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}

func liveHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"alive": true,
		"pid":   os.Getpid(),
	})
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(`# HELP http_requests_total Total HTTP requests
# TYPE http_requests_total counter
http_requests_total{method="GET",status="200"} ` + fmt.Sprintf("%d", requestCount.Load()) + `
# HELP up Service is up
# TYPE up gauge
up 1
`))
}

func versionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"version":    "1.0.0",
		"build":      "abc123",
		"build_date": "2024-01-19",
		"go_version": "go1.23",
	})
}

func apiStatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "operational",
		"services": map[string]string{
			"database": "healthy",
			"cache":    "healthy",
			"queue":    "healthy",
		},
		"uptime_seconds": 86400,
	})
}

func serviceHealthHandler(serviceName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"service": serviceName,
			"status":  "healthy",
			"latency_ms": 5,
			"connections": map[string]interface{}{
				"active": 10,
				"idle":   5,
			},
		})
	}
}
