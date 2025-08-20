package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/aarondever/chirpy/internal/auth"
	"github.com/aarondever/chirpy/internal/utils"
	"github.com/google/uuid"
)

func (cfg *apiConfig) polkaWebhook(w http.ResponseWriter, r *http.Request) {
	key, err := auth.GetAPIKey(r.Header)
	if err != nil {
		utils.RespondWithError(w, r, err.Error(), http.StatusUnauthorized)
		return
	}

	if key != cfg.polkaKey {
		utils.RespondWithError(w, r, "Unauthorized", http.StatusUnauthorized)
		return
	}

	type requestBody struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	var body requestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.RespondWithError(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	if body.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	_, err = cfg.dbQueries.UpgradeUser(r.Context(), body.Data.UserID)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
