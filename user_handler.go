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
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
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
	// parse request
	var body userRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.RespondWithError(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	// get user from db
	user, err := cfg.dbQueries.GetUserByEmail(r.Context(), body.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.RespondWithJSON(w, r, err.Error(), http.StatusNotFound)
			return
		}

		utils.RespondWithError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	// password check
	if err := auth.CheckPasswordHash(body.Password, user.HashedPassword); err != nil {
		utils.RespondWithError(w, r, "Incorrect email or password", http.StatusUnauthorized)
		return
	}

	// generate jwt
	token, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Hour)
	if err != nil {
		utils.RespondWithJSON(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	// generate refresh token
	refreshToken := auth.MakeRefreshToken()
	cfg.dbQueries.CreateRfreshToken(r.Context(), database.CreateRfreshTokenParams{
		Token:  refreshToken,
		UserID: user.ID,
	})

	// return response
	response := userResponse{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refreshToken,
		IsChirpyRed:  user.IsChirpyRed,
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
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	}

	utils.RespondWithJSON(w, r, response, http.StatusCreated)
}

func (cfg *apiConfig) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	userID, err := cfg.getUserFromToken(r)
	if err != nil {
		utils.RespondWithError(w, r, err.Error(), http.StatusUnauthorized)
		return
	}

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

	user, err := cfg.dbQueries.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:             userID,
		Email:          body.Email,
		HashedPassword: hash,
	})
	if err != nil {
		utils.RespondWithError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	response := userResponse{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	}

	utils.RespondWithJSON(w, r, response, http.StatusOK)
}

func (cfg *apiConfig) handleCreateChirp(w http.ResponseWriter, r *http.Request) {
	userID, err := cfg.getUserFromToken(r)
	if err != nil {
		utils.RespondWithError(w, r, err.Error(), http.StatusUnauthorized)
		return
	}

	type requestBody struct {
		Body string `json:"body"`
	}

	var body requestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.RespondWithError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(body.Body) > 140 {
		utils.RespondWithError(w, r, "Chirp is too long", http.StatusBadRequest)
		return
	}

	chirp, err := cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   censorProfane(body.Body),
		UserID: userID,
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
			utils.RespondWithJSON(w, r, err.Error(), http.StatusNotFound)
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

func (cfg *apiConfig) handleDeleteChirp(w http.ResponseWriter, r *http.Request) {
	userID, err := cfg.getUserFromToken(r)
	if err != nil {
		utils.RespondWithError(w, r, err.Error(), http.StatusUnauthorized)
		return
	}

	v := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(v)
	if err != nil {
		utils.RespondWithError(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	chirp, err := cfg.dbQueries.GetChirpById(r.Context(), chirpID)
	if err != nil {
		utils.RespondWithError(w, r, err.Error(), http.StatusNotFound)
		return
	}
	if chirp.UserID != userID {
		utils.RespondWithError(w, r, "Forbidden", http.StatusForbidden)
		return
	}

	if err := cfg.dbQueries.DeleteChirp(r.Context(), database.DeleteChirpParams{
		ID:     chirpID,
		UserID: userID,
	}); err != nil {
		utils.RespondWithError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) handleRefreshToken(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		utils.RespondWithError(w, r, err.Error(), http.StatusUnauthorized)
		return
	}

	userID, err := cfg.dbQueries.GetUserFromRefreshToken(r.Context(), refreshToken)
	if err != nil {
		utils.RespondWithError(w, r, err.Error(), http.StatusUnauthorized)
		return
	}

	token, err := auth.MakeJWT(userID, cfg.jwtSecret, time.Hour)
	if err != nil {
		utils.RespondWithError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct {
		Token string `json:"token"`
	}{
		Token: token,
	}

	utils.RespondWithJSON(w, r, response, http.StatusOK)
}

func (cfg *apiConfig) handleRevokeToken(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		utils.RespondWithError(w, r, err.Error(), http.StatusUnauthorized)
		return
	}

	if err := cfg.dbQueries.RevokeRefreshToken(r.Context(), token); err != nil {
		utils.RespondWithError(w, r, err.Error(), http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
