package main

import (
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
}

type userData struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
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
	receivedUserData := userData{}
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
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
	respondWithJSON(w, 201, responseUser)
}

func (cfg *apiConfig) loginUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	receivedUserData := userData{}
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
	responseUser := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
	respondWithJSON(w, 200, responseUser)
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
