package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

type validResponse struct {
	Valid bool `json:"valid"`
}

type cleanedResponse struct {
	CleanedBody string `json:"cleaned_body"`
}

func handleValidation(w http.ResponseWriter, r *http.Request) {
	var body requestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		RespondWithError(w, r, "Something went wrong", http.StatusInternalServerError)
		return
	}

	if len(body.Body) > 140 {
		RespondWithError(w, r, "Chirp is too long", http.StatusBadRequest)
		return
	}

	response := cleanedResponse{
		CleanedBody: censorProfane(body.Body),
	}

	RespondWithJSON(w, r, response, http.StatusOK)
}

func censorProfane(text string) string {
	const (
		KERFUFFLE = "kerfuffle"
		SHARBERT  = "sharbert"
		FORNAX    = "fornax"
	)

	var result []string
	for s := range strings.SplitSeq(text, " ") {
		switch strings.ToLower(s) {
		case KERFUFFLE, SHARBERT, FORNAX:
			result = append(result, "****")
		default:
			result = append(result, s)
		}
	}

	return strings.Join(result, " ")
}
