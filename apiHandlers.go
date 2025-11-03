package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/delroscol98/chirpy/internal/database"
	"github.com/google/uuid"
)

type ChirpRequestBody struct {
	Body   string    `json:"body"`
	UserID uuid.UUID `json:"user_id"`
}

type ChirpResponseBody struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

// NOTE: GET requests
func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	_, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

// NOTE: POST requests
func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	data, err := io.ReadAll(r.Body)
	if err != nil {
		errMsg := fmt.Sprintf("Error reading request body: %v\n", err)
		respondWithError(w, http.StatusInternalServerError, errMsg)
		return
	}

	params := ChirpRequestBody{}
	err = json.Unmarshal(data, &params)
	if err != nil {
		errMsg := fmt.Sprintf("Error unmarshalling data: %v\n", err)
		respondWithError(w, http.StatusInternalServerError, errMsg)
		return
	}

	if len(params.Body) > 140 {
		errMsg := "chirps must be at most 140 characters long"
		respondWithError(w, http.StatusBadRequest, errMsg)
		return
	}

	chirp, err := cfg.database.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleanBody(params.Body),
		UserID: params.UserID,
	})
	if err != nil {
		errMsg := fmt.Sprintf("Error creating chirp: %v", err)
		respondWithError(w, http.StatusInternalServerError, errMsg)
		return
	}

	respondWithJSON(w, http.StatusCreated, ChirpResponseBody{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.database.GetAllChirpsAsc(r.Context())
	if err != nil {
		errMsg := fmt.Sprintf("Error getting chirps: %v", err)
		respondWithError(w, http.StatusInternalServerError, errMsg)
		return
	}

	out := make([]ChirpResponseBody, 0, len(chirps))
	for _, chirp := range chirps {
		out = append(out, ChirpResponseBody{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		})
	}

	respondWithJSON(w, http.StatusOK, out)
}

func (cfg *apiConfig) handlerCreateUsers(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	type requestBody struct {
		Email string `json:"email"`
	}
	type User struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		errMsg := fmt.Sprintf("Error reading request body: %v", err)
		respondWithError(w, http.StatusInternalServerError, errMsg)
		return
	}

	params := requestBody{}
	err = json.Unmarshal(data, &params)
	if err != nil {
		errMsg := fmt.Sprintf("Error unmarshalling data: %v", err)
		respondWithError(w, http.StatusInternalServerError, errMsg)
		return
	}

	user, err := cfg.database.CreateUser(r.Context(), params.Email)
	if err != nil {
		errMsg := fmt.Sprintf("Error creating new user: %v", err)
		respondWithError(w, http.StatusInternalServerError, errMsg)
		return
	}

	respondWithJSON(w, 201, User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	})
}
