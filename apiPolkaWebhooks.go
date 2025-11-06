package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUpgradeUserChirpyRed(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	data, err := io.ReadAll(r.Body)
	if err != nil {
		errMsg := fmt.Sprintf("Error reading request body: %v", err)
		respondWithError(w, http.StatusInternalServerError, errMsg)
		return
	}

	var req WebhookRequestBody
	err = json.Unmarshal(data, &req)
	if err != nil {
		errMsg := fmt.Sprintf("Error unmarshalling data: %v", err)
		respondWithError(w, http.StatusInternalServerError, errMsg)
		return
	}

	if req.Event != "user.upgraded" {
		respondWithJSON(w, http.StatusNoContent, nil)
		return
	}

	userID, err := uuid.Parse(req.Data.UserID)
	if err != nil {
		errMsg := fmt.Sprintf("Error parsing UserID to uuid: %v", err)
		respondWithError(w, http.StatusInternalServerError, errMsg)
		return
	}

	user, err := cfg.database.UpgradeUserChirpyRed(r.Context(), userID)
	if err != nil {
		errMsg := fmt.Sprintf("Error upgrading user to chirpy red: %v", err)
		respondWithError(w, http.StatusNotFound, errMsg)
		return
	}

	if user.IsChirpyRed == false {
		errMsg := "User not upgraded"
		respondWithError(w, http.StatusInternalServerError, errMsg)
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}
