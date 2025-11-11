package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/marekbrze/chirpy/internal/auth"
	"github.com/marekbrze/chirpy/internal/database"
)

type upgradeParams struct {
	Event string `json:"event"`
	Data  struct {
		UserID uuid.UUID `json:"user_id"`
	}
}

func (cfg *apiConfig) upgradeUser(w http.ResponseWriter, r *http.Request) {
	headerToken, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, 401, "Unauthorized")
		return
	}
	if headerToken != cfg.apiKey {
		respondWithError(w, 401, "Unauthorized")
		return
	}
	decoder := json.NewDecoder(r.Body)
	receivedParams := upgradeParams{}
	err = decoder.Decode(&receivedParams)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}
	if receivedParams.Event != "user.upgraded" {
		respondWithJSON(w, 204, nil)
	}
	upgradeUserParams := database.UpgradeUserParams{
		ID:          receivedParams.Data.UserID,
		UpdatedAt:   time.Now().UTC(),
		IsChirpyRed: true,
	}
	user, err := cfg.dbQueries.UpgradeUser(r.Context(), upgradeUserParams)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, 404, "User doesn't exist")
			return
		}
		respondWithError(w, 500, "Something went wrong")
		return
	}
	responseUser := User{
		ID:          user.ID,
		Email:       user.Email,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		IsChirpyRed: user.IsChirpyRed,
	}
	respondWithJSON(w, 204, responseUser)
}
