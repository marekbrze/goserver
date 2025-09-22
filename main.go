package main

import (
	"fmt"
	"encoding/json"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverhits atomic.Int32
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
	type chirp struct {
		Body string `json:"body"`
	}

	type errorResponse struct {
		Error string `json:"error"`
	}

	type validResponse struct {
		Valid bool `json:"valid"`
	}
	genericError := errorResponse{Error: "something went wrong"}

	decoder := json.NewDecoder(r.Body)
	receivedChirp := chirp{}
	err := decoder.Decode(&receivedChirp)
	if err != nil {
		returnError, err := json.Marshal(genericError)

		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(500)
		w.Header().Set("Content-Type", "application/json")
		w.Write(returnError)
		return
	}
	if len(receivedChirp.Body) > 140 {
		returnError, err := json.Marshal(errorResponse{Error: "Chirp is too long"})
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(400)
		w.Header().Set("Content-Type", "application/json")
		w.Write(returnError)
		return
	}


	response, err := json.Marshal(validResponse{Valid: true})
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
	return
}

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
