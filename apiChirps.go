package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"

	"github.com/delroscol98/chirpy/internal/auth"
	"github.com/delroscol98/chirpy/internal/database"
	"github.com/google/uuid"
)

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

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	var chirps []database.Chirp
	authorID, err := uuid.Parse(r.URL.Query().Get("author_id"))
	if err != nil {
		chirps, err = cfg.database.GetAllChirpsAsc(r.Context())
		if err != nil {
			errMsg := fmt.Sprintf("Error getting chirps: %v", err)
			respondWithError(w, http.StatusInternalServerError, errMsg)
			return
		}
	} else {
		chirps, err = cfg.database.GetAllChirpsByIdAsc(r.Context(), authorID)
		if err != nil {
			errMsg := fmt.Sprintf("Error getting chirps: %v", err)
			respondWithError(w, http.StatusInternalServerError, errMsg)
			return
		}
	}

	sortParam := r.URL.Query().Get("sort")
	if sortParam == "desc" {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].CreatedAt.After(chirps[j].CreatedAt)
		})
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

func (cfg *apiConfig) handlerDeleteChirpByID(w http.ResponseWriter, r *http.Request) {
	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		errMsg := fmt.Sprintf("Error getting bearer token: %v", err)
		respondWithError(w, http.StatusUnauthorized, errMsg)
		return
	}

	userID, err := auth.ValidateJWT(accessToken, cfg.secret)
	if err != nil {
		errMsg := fmt.Sprintf("Error validating access token: %v", err)
		respondWithError(w, http.StatusForbidden, errMsg)
		return
	}

	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		errMsg := fmt.Sprintf("Error parsing string uuid: %v", err)
		respondWithError(w, http.StatusInternalServerError, errMsg)
		return
	}

	chirp, err := cfg.database.GetChirpById(r.Context(), chirpID)
	if err != nil {
		errMsg := fmt.Sprintf("Error getting chirp by ID: %v", err)
		respondWithError(w, http.StatusNotFound, errMsg)
		return
	}

	if userID != chirp.UserID {
		errMsg := "User forbidden for this action"
		respondWithError(w, http.StatusForbidden, errMsg)
		return
	}

	err = cfg.database.DeleteChirpById(r.Context(), database.DeleteChirpByIdParams{
		ID:     chirpID,
		UserID: userID,
	})
	if err != nil {
		errMsg := "Error deleting Chirp: Chirp not found"
		respondWithError(w, http.StatusNotFound, errMsg)
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}
