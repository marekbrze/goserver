package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/marekbrze/chirpy/internal/auth"
	"github.com/marekbrze/chirpy/internal/database"
)

type apiConfig struct {
	fileserverhits atomic.Int32
	dbQueries      *database.Queries
	platform       string
	jwtSecret      string
	apiKey         string
}

type UserData struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type User struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}

type TokenResponse struct {
	User
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

type SimpleTokenResponse struct {
	Token string `json:"token"`
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverhits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) resetMetrics(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverhits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	responseBody := []byte("Hits reseted")

	_, err := w.Write(responseBody)
	if err != nil {
		// Log any error that occurs during writing the response.
		log.Println("Failed to write response:", err)
	}
}

func (cfg *apiConfig) getNumberOfHits(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	responseString := fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, cfg.fileserverhits.Load())
	responseBody := []byte(responseString)

	_, err := w.Write(responseBody)
	if err != nil {
		// Log any error that occurs during writing the response.
		log.Println("Failed to write response:", err)
	}
}

func (cfg *apiConfig) addUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	receivedUserData := UserData{}
	err := decoder.Decode(&receivedUserData)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}
	hashedPassword, err := auth.HashPassword(receivedUserData.Password)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}
	userParams := database.CreateUserParams{
		ID:             uuid.New(),
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
		Email:          receivedUserData.Email,
		HashedPassword: hashedPassword,
	}
	user, err := cfg.dbQueries.CreateUser(r.Context(), userParams)
	if err != nil {
		log.Println("Failed to write response:", err)
	}
	responseUser := User{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	}
	respondWithJSON(w, 201, responseUser)
}

func (cfg *apiConfig) loginUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	receivedUserData := UserData{}
	err := decoder.Decode(&receivedUserData)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}
	user, err := cfg.dbQueries.GetUser(r.Context(), receivedUserData.Email)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}
	loginCorrect, err := auth.CheckPasswordHash(receivedUserData.Password, user.HashedPassword)
	if err != nil || !loginCorrect {
		respondWithError(w, 401, "Incorrect email or password")
		return
	}
	expiresIn := time.Hour
	token, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Duration(expiresIn))
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
	}
	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
	}
	refreshTokenParams := database.CreateRefreshTokenParams{
		Token:     refreshToken,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().AddDate(0, 0, 60).UTC(),
		UserID:    user.ID,
	}
	savedToken, err := cfg.dbQueries.CreateRefreshToken(r.Context(), refreshTokenParams)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
	}
	responseWithToken := TokenResponse{
		User: User{
			ID:          user.ID,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
			Email:       user.Email,
			IsChirpyRed: user.IsChirpyRed,
		},
		Token:        token,
		RefreshToken: savedToken.Token,
	}
	respondWithJSON(w, 200, responseWithToken)
}

func (cfg *apiConfig) reset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		respondWithJSON(w, 403, nil)
	}
	err := cfg.dbQueries.DeleteUsers(r.Context())
	if err != nil {
		log.Println("Failed to write response:", err)
	}
	respondWithJSON(w, 200, nil)
}

func (cfg *apiConfig) refreshToken(w http.ResponseWriter, r *http.Request) {
	headerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "Unauthorized")
		return
	}
	tokenInfo, err := cfg.dbQueries.GetTokenInfo(r.Context(), headerToken)
	if err != nil {
		respondWithError(w, 401, "Unauthorized")
		return
	}
	if tokenInfo.ExpiresAt.Before(time.Now()) || tokenInfo.RevokedAt.Valid {
		respondWithError(w, 401, "Unauthorized")
		return
	}
	newToken, err := auth.MakeJWT(tokenInfo.UserID, cfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, 401, "Unauthorized")
		return
	}
	responseWithToken := SimpleTokenResponse{
		Token: newToken,
	}
	respondWithJSON(w, 200, responseWithToken)
}

func (cfg *apiConfig) revokeToken(w http.ResponseWriter, r *http.Request) {
	headerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "Unauthorized")
		return
	}
	tokenInfo, err := cfg.dbQueries.GetTokenInfo(r.Context(), headerToken)
	if err != nil {
		respondWithError(w, 401, "Unauthorized")
		return
	}
	if tokenInfo.ExpiresAt.Before(time.Now()) || tokenInfo.RevokedAt.Valid {
		respondWithError(w, 401, "Unauthorized")
		return
	}
	revokeTokenParams := database.RevokeTokenParams{
		UpdatedAt: time.Now().UTC(),
		RevokedAt: sql.NullTime{
			Time:  time.Now().UTC(),
			Valid: true,
		},
		Token: tokenInfo.Token,
	}
	err = cfg.dbQueries.RevokeToken(r.Context(), revokeTokenParams)
	if err != nil {
		respondWithError(w, 401, "Unauthorized")
		return
	}
	respondWithError(w, 204, "Token revoked")
}
