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

	// admin enpoints
	serverMux.HandleFunc("GET /admin/metrics", cfg.handleMetrics)
	serverMux.HandleFunc("POST /admin/reset", cfg.resetMetrics)

	// api enpoints
	serverMux.HandleFunc("GET /api/healthz", handleReadiness)

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
	html := fmt.Sprintf(`<html>
							<body>
								<h1>Welcome, Chirpy Admin</h1>
								<p>Chirpy has been visited %d times!</p>
							</body>
						</html>`,
		cfg.fileserverHits.Load())
	body := []byte(html)

	responseWriter.Header().Set("Content-Type", "text/html")
	responseWriter.Write(body)
}

func (cfg *apiConfig) resetMetrics(responseWriter http.ResponseWriter, request *http.Request) {
	cfg.fileserverHits.Store(0)
}
