package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	const filePathRoot = "."
	const port = "8080"

	var cfg apiConfig

	handler := http.FileServer(http.Dir(filePathRoot))

	serveMux := http.NewServeMux()
	serveMux.Handle("/app/", http.StripPrefix("/app/", cfg.middlewareMetricsInc(handler)))

	serveMux.HandleFunc("GET /healthz", handlerReadiness)
	serveMux.HandleFunc("GET /metrics", cfg.handlerWriteRequestsNumber)
	serveMux.HandleFunc("POST /reset", cfg.handlerResetRequestsNumber)

	server := &http.Server{
		Handler: serveMux,
		Addr:    ":" + port,
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("Error:", err)
	}
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	fmt.Printf("Received request body from handlerReadiness: %s\n", string(body))

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (cfg *apiConfig) handlerWriteRequestsNumber(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	fmt.Printf("Received request body from handlerWriteRequestsNumber: %s\n", string(body))

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	message := fmt.Sprintf("Hits: %v\n", cfg.fileserverHits.Load())

	w.Write([]byte(message))
}

func (cfg *apiConfig) handlerResetRequestsNumber(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	fmt.Printf("Recieved from body handlerResetRequestsNumber: %s\n", string(body))

	cfg.fileserverHits.Store(0)

	w.Header().Set("Content-Type", "text/plain; charset=utf=8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("FileserverHits set back to 0"))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}
