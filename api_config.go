package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

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
