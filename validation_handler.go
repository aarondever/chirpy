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

func handleValidation(responseWriter http.ResponseWriter, request *http.Request) {
	var body requestBody
	if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
		RespondWithError(
			responseWriter,
			request,
			"Something went wrong",
			http.StatusInternalServerError)
		return
	}

	if len(body.Body) > 140 {
		RespondWithError(
			responseWriter,
			request,
			"Chirp is too long",
			http.StatusBadRequest)
		return
	}

	response := cleanedResponse{
		CleanedBody: censorProfane(body.Body),
	}

	RespondWithJSON(responseWriter, request, response, http.StatusOK)
}

func censorProfane(text string) string {
	const (
		KERFUFFLE = "kerfuffle"
		SHARBERT  = "sharbert"
		FORNAX    = "fornax"
	)

	var result []string
	for _, s := range strings.Split(text, " ") {
		switch strings.ToLower(s) {
		case KERFUFFLE, SHARBERT, FORNAX:
			result = append(result, "****")
		default:
			result = append(result, s)
		}
	}

	return strings.Join(result, " ")
}
