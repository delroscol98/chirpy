package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/delroscol98/chirpy/internal/auth"
	"github.com/delroscol98/chirpy/internal/database"
)

func (cfg *apiConfig) handlerUpdatedUserEmailPassword(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	data, err := io.ReadAll(r.Body)
	if err != nil {
		errMsg := fmt.Sprintf("Error reading request body: %v", err)
		respondWithError(w, http.StatusInternalServerError, errMsg)
		return
	}

	var req UserRequestBody
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

	respondWithJSON(w, http.StatusOK, UserResponseBody{
		ID:        user.ID,
		UpdatedAt: time.Now(),
		Email:     user.Email,
	})
}

func (cfg *apiConfig) handlerCreateUsers(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	data, err := io.ReadAll(r.Body)
	if err != nil {
		errMsg := fmt.Sprintf("Error reading request body: %v", err)
		respondWithError(w, http.StatusInternalServerError, errMsg)
		return
	}

	params := UserRequestBody{}
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

	respondWithJSON(w, 201, UserResponseBody{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	})
}
