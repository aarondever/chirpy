package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

func main() {
	cfg := apiConfig{
		fileserverHits: atomic.Int32{},
	}

	serverMux := http.NewServeMux()

	// app enpoints
	serverMux.Handle("/app/",
		cfg.middlewareMetricsInt(
			http.StripPrefix("/app",
				http.FileServer(http.Dir(".")))))

	// api enpoints
	serverMux.HandleFunc("GET /api/healthz", handleReadiness)
	serverMux.HandleFunc("GET /api/metrics", cfg.handleMetrics)
	serverMux.HandleFunc("POST /api/reset", cfg.resetMetrics)

	server := http.Server{
		Handler: serverMux,
		Addr:    ":8080",
	}

	server.ListenAndServe()
}

func handleReadiness(responseWriter http.ResponseWriter, request *http.Request) {
	responseWriter.Header().Set("Content-Type", "text/plain; charset=utf-8")
	responseWriter.WriteHeader(http.StatusOK)

	body := []byte("OK")
	responseWriter.Write(body)
}

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInt(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)

		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handleMetrics(responseWriter http.ResponseWriter, request *http.Request) {
	body := fmt.Appendf(nil, "Hits: %d", cfg.fileserverHits.Load())
	responseWriter.Write(body)
}

func (cfg *apiConfig) resetMetrics(responseWriter http.ResponseWriter, request *http.Request) {
	cfg.fileserverHits.Store(0)
}
