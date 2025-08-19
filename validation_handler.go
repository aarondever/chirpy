package main

import (
	"encoding/json"
	"net/http"
)

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

	response := validResponse{
		Valid: true,
	}
	RespondWithJSON(responseWriter, request, response, http.StatusOK)
}
