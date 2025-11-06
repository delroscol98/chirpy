package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/delroscol98/chirpy/internal/auth"
	"github.com/delroscol98/chirpy/internal/database"
)

func (cfg *apiConfig) handlerGetUserByEmail(w http.ResponseWriter, r *http.Request) {
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
	respondWithJSON(w, http.StatusOK, UserResponseBody{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refreshToken.Token,
	})
}
