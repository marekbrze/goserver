package main

import (
	"database/sql"
	"net/http"

	"github.com/google/uuid"
	"github.com/marekbrze/chirpy/internal/auth"
)

func (cfg *apiConfig) deleteChirp(w http.ResponseWriter, r *http.Request) {
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
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}
	dbChirp, err := cfg.dbQueries.GetChirp(r.Context(), chirpID)
	if err != nil {
		if err == sql.ErrNoRows && userID == dbChirp.UserID {
			respondWithError(w, 404, "Chirp doesn't exist")
			return
		}
		respondWithError(w, 500, "Something went wrong")
		return
	}
	if userID != dbChirp.UserID {
		respondWithError(w, 403, "Unauthorized")
		return
	}
	err = cfg.dbQueries.DeleteChirp(r.Context(), dbChirp.ID)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}
	respondWithJSON(w, 204, nil)
}
