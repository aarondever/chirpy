package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/aarondever/chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Email string `json:"email"`
	}

	var body requestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		RespondWithError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	user, err := cfg.dbQueries.CreateUser(r.Context(), body.Email)
	if err != nil {
		RespondWithError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	type responseBody struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	response := responseBody{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	RespondWithJSON(w, r, response, http.StatusCreated)
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	var body requestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		RespondWithError(w, r, "Something went wrong", http.StatusInternalServerError)
		return
	}

	if len(body.Body) > 140 {
		RespondWithError(w, r, "Chirp is too long", http.StatusBadRequest)
		return
	}

	chirp, err := cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   censorProfane(body.Body),
		UserID: body.UserID,
	})
	if err != nil {
		RespondWithError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	type responseBody struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}

	response := responseBody{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}

	RespondWithJSON(w, r, response, http.StatusCreated)
}
