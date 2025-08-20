package main

import (
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/aarondever/chirpy/internal/auth"
	"github.com/aarondever/chirpy/internal/database"
	"github.com/aarondever/chirpy/internal/utils"
	"github.com/google/uuid"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	platform       string
	dbQueries      *database.Queries
	jwtSecret      string
	polkaKey       string
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
		utils.RespondWithError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (cfg *apiConfig) getUserFromToken(r *http.Request) (uuid.UUID, error) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		return uuid.UUID{}, err
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		return uuid.UUID{}, err
	}

	return userID, nil
}
