package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/delroscol98/chirpy/internal/auth"
)

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
