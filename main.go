package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/delroscol98/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	database       *database.Queries
	platform       string
	secret         string
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	secret := os.Getenv("SECRET")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Error opening database")
	}
	dbQueries := database.New(db)

	const filePathRoot = "."
	const port = "8080"

	cfg := apiConfig{
		database: dbQueries,
		platform: platform,
		secret:   secret,
	}

	handler := http.FileServer(http.Dir(filePathRoot))

	serveMux := http.NewServeMux()
	serveMux.Handle("/app/", http.StripPrefix("/app/", cfg.middlewareMetricsInc(handler)))

	serveMux.HandleFunc("GET /api/healthz", handlerReadiness)
	serveMux.HandleFunc("POST /api/chirps", cfg.handlerCreateChirp)
	serveMux.HandleFunc("GET /api/chirps", cfg.handlerGetChirps)
	serveMux.HandleFunc("GET /api/chirps/{chirpID}", cfg.handlerGetChirpById)
	serveMux.HandleFunc("POST /api/users", cfg.handlerCreateUsers)
	serveMux.HandleFunc("PUT /api/users", cfg.handlerUpdatedUserEmailPassword)
	serveMux.HandleFunc("POST /api/login", cfg.handlerGetUserByEmail)
	serveMux.HandleFunc("POST /api/refresh", cfg.handlerGetRefreshToken)
	serveMux.HandleFunc("POST /api/revoke", cfg.handlerRevokeRefreshToken)
	serveMux.HandleFunc("GET /admin/metrics", cfg.handlerWriteRequestsNumber)
	serveMux.HandleFunc("POST /admin/reset", cfg.handlerResetRequestsNumber)

	server := &http.Server{
		Handler: serveMux,
		Addr:    ":" + port,
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Fatal("Error:", err)
	}
}
