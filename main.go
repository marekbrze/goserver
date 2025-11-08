package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/marekbrze/chirpy/internal/database"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	dbURL := os.Getenv("CHIRPY_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Couldn't connect to database")
	}
	dbQueries := database.New(db)
	apiCfg := apiConfig{
		fileserverhits: atomic.Int32{},
		dbQueries:      dbQueries,
		platform:       os.Getenv("PLATFORM"),
		jwtSecret:      os.Getenv("JWT_SECRET"),
	}
	serverMux := http.NewServeMux()
	serverMux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	serverMux.HandleFunc("GET /api/healthz", healthCheck)
	serverMux.HandleFunc("GET /admin/metrics", apiCfg.getNumberOfHits)
	serverMux.HandleFunc("POST /admin/reset", apiCfg.reset)
	serverMux.HandleFunc("POST /api/users", apiCfg.addUser)
	serverMux.HandleFunc("POST /api/login", apiCfg.loginUser)
	serverMux.HandleFunc("POST /api/chirps", apiCfg.addChirp)
	serverMux.HandleFunc("GET /api/chirps", apiCfg.getAllChirps)
	serverMux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.getChirpByID)
	server := &http.Server{
		Handler: serverMux,
		Addr:    ":8080",
	}
	fmt.Println("Server ready")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("There was a problem %v", err)
	}
}
