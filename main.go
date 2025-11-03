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
}

func main() {
	godotenv.Load()

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Error opening database")
	}
	dbQueries := database.New(db)

	const filePathRoot = "."
	const port = "8080"

	platform := os.Getenv("PLATFORM")
	cfg := apiConfig{
		database: dbQueries,
		platform: platform,
	}

	handler := http.FileServer(http.Dir(filePathRoot))

	serveMux := http.NewServeMux()
	serveMux.Handle("/app/", http.StripPrefix("/app/", cfg.middlewareMetricsInc(handler)))

	serveMux.HandleFunc("GET /api/healthz", handlerReadiness)
	serveMux.HandleFunc("POST /api/chirps", cfg.handlerValidateChirp)
	serveMux.HandleFunc("POST /api/users", cfg.handlerCreateUsers)
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
