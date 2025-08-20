package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/aarondever/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	cfg := apiConfig{
		fileserverHits: atomic.Int32{},
		platform:       os.Getenv("PLATFORM"),
		dbQueries:      database.New(db),
		jwtSecret:      os.Getenv("JWT_SECRET"),
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
	serverMux.HandleFunc("GET /api/chirps", cfg.handleGetChirps)
	serverMux.HandleFunc("GET /api/chirps/{chirpID}", cfg.handleGetChirpByID)
	serverMux.HandleFunc("POST /api/chirps", cfg.handleCreateChirp)
	serverMux.HandleFunc("DELETE /api/chirps/{chirpID}", cfg.handleDeleteChirp)
	serverMux.HandleFunc("POST /api/users", cfg.handleCreateUser)
	serverMux.HandleFunc("PUT /api/users", cfg.handleUpdateUser)
	serverMux.HandleFunc("POST /api/login", cfg.handleLogin)
	serverMux.HandleFunc("POST /api/refresh", cfg.handleRefreshToken)
	serverMux.HandleFunc("POST /api/revoke", cfg.handleRevokeToken)
	serverMux.HandleFunc("POST /api/polka/webhooks", cfg.polkaWebhook)

	server := http.Server{
		Handler: serverMux,
		Addr:    ":8080",
	}

	server.ListenAndServe()
}

func handleReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	body := []byte("OK")
	w.Write(body)
}
