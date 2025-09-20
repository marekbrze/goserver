package main

import (
	"log"
	"net/http"
)

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

func main() {
	serverMux := http.NewServeMux()
	serverMux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	serverMux.HandleFunc("/healthz", healthCheck)
	server := &http.Server{
		Handler: serverMux,
		Addr:    ":8080",
	}
	server.ListenAndServe()
}
