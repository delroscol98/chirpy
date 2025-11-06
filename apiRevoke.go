package main

import (
	"fmt"
	"net/http"
	"strings"
)

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
