package main

import (
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/aarondever/chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	platform       string
	dbQueries      *database.Queries
}

func (cfg *apiConfig) middlewareMetricsInt(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)

		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handleMetrics(w http.ResponseWriter, r *http.Request) {
	html := fmt.Sprintf(`<html>
							<body>
								<h1>Welcome, Chirpy Admin</h1>
								<p>Chirpy has been visited %d times!</p>
							</body>
						</html>`,
		cfg.fileserverHits.Load())
	body := []byte(html)

	w.Header().Set("Content-Type", "text/html")
	w.Write(body)
}

func (cfg *apiConfig) resetMetrics(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	cfg.fileserverHits.Store(0)

	if err := cfg.dbQueries.ResetUsers(r.Context()); err != nil {
		RespondWithError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
}
