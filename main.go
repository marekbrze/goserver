package main

import (
	"net/http"
	"sync/atomic"
)

func main() {
	apiCfg := apiConfig{
		fileserverhits: atomic.Int32{},
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
