package main

import (
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

	serveMux.HandleFunc("GET /api/healthz", handlerReadiness)
	serveMux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)
	serveMux.HandleFunc("GET /admin/metrics", cfg.handlerWriteRequestsNumber)
	serveMux.HandleFunc("POST /admin/reset", cfg.handlerResetRequestsNumber)

	server := &http.Server{
		Handler: serveMux,
		Addr:    ":" + port,
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("Error:", err)
	}
}
