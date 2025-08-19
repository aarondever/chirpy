package utils

import (
	"encoding/json"
	"net/http"
)

func RespondWithJSON(w http.ResponseWriter, r *http.Request, payload any, statusCode int) {
	w.Header().Set("Content-Type", "application/json")

	data, err := json.Marshal(payload)
	if err != nil {
		RespondWithError(w, r, "Something went wrong", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(statusCode)
	w.Write(data)
}

func RespondWithError(w http.ResponseWriter, r *http.Request, err string, statusCode int) {
	type errorResponse struct {
		Error string `json:"error"`
	}
	response := errorResponse{Error: err}
	RespondWithJSON(w, r, response, statusCode)
}
