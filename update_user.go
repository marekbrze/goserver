package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/marekbrze/chirpy/internal/auth"
	"github.com/marekbrze/chirpy/internal/database"
)

type userUpdateData struct {
	NewEmail    string `json:"email"`
	NewPassword string `json:"password"`
}

func (cfg *apiConfig) updateUser(w http.ResponseWriter, r *http.Request) {
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
	receivedUserData := userUpdateData{}
	err = decoder.Decode(&receivedUserData)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}
	hashedPassword, err := auth.HashPassword(receivedUserData.NewPassword)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}
	updateUserParams := database.UpdateUserParams{
		ID:             userID,
		Email:          receivedUserData.NewEmail,
		HashedPassword: hashedPassword,
		UpdatedAt:      time.Now().UTC(),
	}
	user, err := cfg.dbQueries.UpdateUser(r.Context(), updateUserParams)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}
	responseUser := User{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	respondWithJSON(w, 200, responseUser)
}
