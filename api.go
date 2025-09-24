package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type chirp struct {
	Body string `json:"body"`
}

func healthCheck(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	responseBody := []byte(http.StatusText(http.StatusOK))

	_, err := writer.Write(responseBody)
	if err != nil {
		// Log any error that occurs during writing the response.
		log.Println("Failed to write response:", err)
	}
}

func validateChirp(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	receivedChirp := chirp{}
	err := decoder.Decode(&receivedChirp)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}
	if len(receivedChirp.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}
	cleaned := eraseProfane(receivedChirp)
	respondWithJSON(w, 200, cleaned)
}
