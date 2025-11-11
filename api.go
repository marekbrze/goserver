package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/marekbrze/chirpy/internal/auth"
	"github.com/marekbrze/chirpy/internal/database"
)

type receivedChirp struct {
	Body string `json:"body"`
}

type chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
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

func (cfg *apiConfig) addChirp(w http.ResponseWriter, r *http.Request) {
	headerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "Unauthorized")
		return
	}
	userID, err := auth.ValidateJWT(headerToken, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, 401, "Unauthorized")
		return
	}
	decoder := json.NewDecoder(r.Body)
	receivedChirp := receivedChirp{}
	err = decoder.Decode(&receivedChirp)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}
	if len(receivedChirp.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}
	cleanedChirp := eraseProfane(receivedChirp.Body)
	chirpParams := database.CreateChirpParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Body:      cleanedChirp,
		UserID:    userID,
	}
	savedChirp, err := cfg.dbQueries.CreateChirp(r.Context(), chirpParams)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}
	responseChirp := chirp{
		ID:        savedChirp.ID,
		CreatedAt: savedChirp.CreatedAt,
		UpdatedAt: savedChirp.UpdatedAt,
		Body:      savedChirp.Body,
		UserID:    savedChirp.UserID,
	}
	respondWithJSON(w, 201, responseChirp)
}

func (cfg *apiConfig) getAllChirps(w http.ResponseWriter, r *http.Request) {
	authorInfo := uuid.NullUUID{}
	s := r.URL.Query().Get("author_id")
	if s != "" {
		parserdUUID, err := uuid.Parse(s)
		authorInfo = uuid.NullUUID{
			UUID:  parserdUUID,
			Valid: true,
		}
		if err != nil {
			respondWithError(w, 500, "Something went wrong")
			return
		}
	}
	chirps, err := cfg.dbQueries.GetAllChirps(r.Context(), authorInfo)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}
	var responseChirps []chirp
	for _, dbChirp := range chirps {
		responseChirp := chirp{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			Body:      dbChirp.Body,
			UserID:    dbChirp.UserID,
		}

		responseChirps = append(responseChirps, responseChirp)
	}

	s = r.URL.Query().Get("sort")
	if s == "desc" {
		sort.Slice(responseChirps, func(i, j int) bool {
			return responseChirps[i].CreatedAt.After(responseChirps[j].CreatedAt)
		})
	}
	respondWithJSON(w, 200, responseChirps)
}

func (cfg *apiConfig) getChirpByID(w http.ResponseWriter, r *http.Request) {
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}
	dbChirp, err := cfg.dbQueries.GetChirp(r.Context(), chirpID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, 404, "Chirp doesn't exist")
			return
		}
		respondWithError(w, 500, "Something went wrong")
		return
	}
	responseChirp := chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	}
	respondWithJSON(w, 200, responseChirp)
}
