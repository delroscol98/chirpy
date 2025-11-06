package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/delroscol98/chirpy/internal/auth"
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
type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
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
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		errMsg := fmt.Sprintf("Error getting bearer token: %v", err)
		respondWithError(w, http.StatusInternalServerError, errMsg)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		errMsg := "User unauthorized for this action"
		respondWithError(w, http.StatusUnauthorized, errMsg)
		return
	}

	chirp, err := cfg.database.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleanBody(params.Body),
		UserID: userID,
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

func (cfg *apiConfig) handlerGetChirpById(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		errMsg := fmt.Sprintf("Error parsing string uuid: %v", err)
		respondWithError(w, http.StatusInternalServerError, errMsg)
		return
	}
	chirp, err := cfg.database.GetChirpById(r.Context(), id)
	if err != nil {
		errMsg := fmt.Sprintf("Error fetching chirp by ID: %v", err)
		respondWithError(w, http.StatusNotFound, errMsg)
		return
	}

	out := ChirpResponseBody{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}

	respondWithJSON(w, http.StatusOK, out)
}

func (cfg *apiConfig) handlerCreateUsers(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	type requestBody struct {
		Password string `json:"password"`
		Email    string `json:"email"`
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

	hashedPw, err := auth.HashPassword(params.Password)

	user, err := cfg.database.CreateUser(r.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPw,
	})
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

func (cfg *apiConfig) handlerGetUserByEmail(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	type requestBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type User struct {
		ID             uuid.UUID `json:"id"`
		CreatedAt      time.Time `json:"created_at"`
		UpdatedAt      time.Time `json:"updated_at"`
		Email          string    `json:"email"`
		HashedPassword string    `json:"hashed_password"`
		Token          string    `json:"token"`
		RefreshToken   string    `json:"refresh_token"`
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		errMsg := fmt.Sprintf("Error reading request body: %v", err)
		respondWithError(w, http.StatusInternalServerError, errMsg)
		return
	}

	var req requestBody
	err = json.Unmarshal(data, &req)
	if err != nil {
		errMsg := fmt.Sprintf("Error unmarshalling data: %v", err)
		respondWithError(w, http.StatusUnauthorized, errMsg)
		return
	}

	user, err := cfg.database.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		errMsg := "Incorrect email or password"
		respondWithError(w, http.StatusUnauthorized, errMsg)
		return
	}

	bool, err := auth.CheckPasswordHash(req.Password, user.HashedPassword)
	if !bool {
		errMsg := "Incorrect email or password"
		respondWithError(w, http.StatusUnauthorized, errMsg)
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.secret, time.Hour)
	if err != nil {
		errMsg := fmt.Sprintf("Error making JWT: %v", err)
		respondWithError(w, http.StatusInternalServerError, errMsg)
		return
	}

	refToken, _ := auth.MakeRefreshToken()
	refreshToken, err := cfg.database.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 60),
		RevokedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: false,
		},
	})
	if err != nil {
		errMsg := fmt.Sprintf("Error creating refresh token: %v", err)
		respondWithError(w, http.StatusInternalServerError, errMsg)
		return
	}

	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", token))
	respondWithJSON(w, http.StatusOK, User{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refreshToken.Token,
	})
}

func (cfg *apiConfig) handlerGetRefreshToken(w http.ResponseWriter, r *http.Request) {
	refToken := strings.TrimSpace(strings.TrimLeft(r.Header.Get("Authorization"), "Bearer"))
	refreshToken, err := cfg.database.GetRefreshToken(r.Context(), refToken)
	if err != nil {
		errMsg := fmt.Sprintf("Error getting refresh token: %v", err)
		respondWithError(w, http.StatusUnauthorized, errMsg)
		return
	}

	if refreshToken.ExpiresAt.Before(time.Now()) {
		errMsg := "Refresh token is expired"
		respondWithError(w, http.StatusUnauthorized, errMsg)
		return
	}

	if refreshToken.RevokedAt.Valid {
		errMsg := "Refresh token is revoked"
		respondWithError(w, http.StatusUnauthorized, errMsg)
		return
	}

	token, err := auth.MakeJWT(refreshToken.UserID, cfg.secret, time.Hour)
	if err != nil {
		errMsg := fmt.Sprintf("Error making token: %v", err)
		respondWithError(w, http.StatusInternalServerError, errMsg)
		return
	}

	respondWithJSON(w, http.StatusOK, struct {
		Token string `json:"token"`
	}{
		Token: token,
	})
}

func (cfg *apiConfig) handlerRevokeRefreshToken(w http.ResponseWriter, r *http.Request) {
	refToken := strings.TrimSpace(strings.TrimLeft(r.Header.Get("Authorization"), "Bearer"))
	err := cfg.database.RevokeRefeshToken(r.Context(), refToken)
	if err != nil {
		errMsg := fmt.Sprintf("Error revoking refresh token: %v", err)
		respondWithError(w, http.StatusInternalServerError, errMsg)
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}

func (cfg *apiConfig) handlerUpdatedUserEmailPassword(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	defer r.Body.Close()
	data, err := io.ReadAll(r.Body)
	if err != nil {
		errMsg := fmt.Sprintf("Error reading request body: %v", err)
		respondWithError(w, http.StatusInternalServerError, errMsg)
		return
	}

	var req requestBody
	err = json.Unmarshal(data, &req)
	if err != nil {
		errMsg := fmt.Sprintf("Error unmarshalling data: %v", err)
		respondWithError(w, http.StatusInternalServerError, errMsg)
		return
	}

	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		errMsg := fmt.Sprintf("Error getting access token: %v", err)
		respondWithError(w, http.StatusUnauthorized, errMsg)
		return
	}

	userID, err := auth.ValidateJWT(accessToken, cfg.secret)
	if err != nil {
		errMsg := fmt.Sprintf("Error validating access token: %v", err)
		respondWithError(w, http.StatusUnauthorized, errMsg)
		return
	}

	hashedPw, err := auth.HashPassword(req.Password)
	if err != nil {
		errMsg := fmt.Sprintf("Error hashing password: %v", err)
		respondWithError(w, http.StatusInternalServerError, errMsg)
		return
	}

	user, err := cfg.database.UpdateUserEmailPassword(r.Context(), database.UpdateUserEmailPasswordParams{
		Email:          req.Email,
		HashedPassword: hashedPw,
		ID:             userID,
	})

	respondWithJSON(w, http.StatusOK, User{
		ID:        user.ID,
		UpdatedAt: time.Now(),
		Email:     user.Email,
	})
}
