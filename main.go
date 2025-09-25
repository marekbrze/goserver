package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	_ "github.com/lib/pq"
	"github.com/marekbrze/chirpy/internal/database"
)

func main() {
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Couldn't connect to database")
	}
	dbQueries := database.New(db)
	apiCfg := apiConfig{
		fileserverhits: atomic.Int32{},
		dbQueries:      dbQueries,
	}
	serverMux := http.NewServeMux()
	serverMux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	serverMux.HandleFunc("GET /api/healthz", healthCheck)
	serverMux.HandleFunc("GET /admin/metrics", apiCfg.getNumberOfHits)
	serverMux.HandleFunc("POST /admin/reset", apiCfg.resetMetrics)
	serverMux.HandleFunc("POST /api/validate_chirp", validateChirp)
	server := &http.Server{
		Handler: serverMux,
		Addr:    ":8080",
	}
	server.ListenAndServe()
}
