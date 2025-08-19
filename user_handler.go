package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/aarondever/chirpy/internal/auth"
	"github.com/aarondever/chirpy/internal/database"
	"github.com/aarondever/chirpy/internal/utils"
	"github.com/google/uuid"
)

type userResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type chirpResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

type userRequest struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	var body userRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.RespondWithError(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := cfg.dbQueries.GetUserByEmail(r.Context(), body.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.RespondWithJSON(w, r, "User not found", http.StatusNotFound)
			return
		}

		utils.RespondWithError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := auth.CheckPasswordHash(body.Password, user.HashedPassword); err != nil {
		utils.RespondWithError(w, r, "Incorrect email or password", http.StatusUnauthorized)
		return
	}

	response := userResponse{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	utils.RespondWithJSON(w, r, response, http.StatusOK)
}

func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var body userRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.RespondWithError(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	hash, err := auth.HashPassword(body.Password)
	if err != nil {
		utils.RespondWithError(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := cfg.dbQueries.CreateUser(r.Context(), database.CreateUserParams{
		Email:          body.Email,
		HashedPassword: hash,
	})
	if err != nil {
		utils.RespondWithError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	response := userResponse{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	utils.RespondWithJSON(w, r, response, http.StatusCreated)
}

func (cfg *apiConfig) handleCreateChirp(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	var body requestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.RespondWithError(w, r, "Something went wrong", http.StatusInternalServerError)
		return
	}

	if len(body.Body) > 140 {
		utils.RespondWithError(w, r, "Chirp is too long", http.StatusBadRequest)
		return
	}

	chirp, err := cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   censorProfane(body.Body),
		UserID: body.UserID,
	})
	if err != nil {
		utils.RespondWithError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	response := chirpResponse{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}

	utils.RespondWithJSON(w, r, response, http.StatusCreated)
}

func (cfg *apiConfig) handleGetChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.dbQueries.GetChirps(r.Context())
	if err != nil {
		utils.RespondWithError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	var response = make([]chirpResponse, 0, len(chirps))
	for _, chirp := range chirps {
		response = append(response, chirpResponse{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		})
	}

	utils.RespondWithJSON(w, r, response, http.StatusOK)
}

func (cfg *apiConfig) handleGetChirpByID(w http.ResponseWriter, r *http.Request) {
	v := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(v)
	if err != nil {
		utils.RespondWithError(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	chirp, err := cfg.dbQueries.GetChirpById(r.Context(), chirpID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.RespondWithJSON(w, r, "Chirp not found", http.StatusNotFound)
			return
		}

		utils.RespondWithError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	response := chirpResponse{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}

	utils.RespondWithJSON(w, r, response, http.StatusOK)
}
