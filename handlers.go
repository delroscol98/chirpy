package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func respondWithJSON(w http.ResponseWriter, code int, payload any) error {
	response, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	w.Write(response)
	return nil
}

func respondWithError(w http.ResponseWriter, code int, msg string) error {
	return respondWithJSON(w, code, map[string]string{"error": msg})
}

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

func (cfg *apiConfig) handlerWriteRequestsNumber(w http.ResponseWriter, r *http.Request) {
	_, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	message := fmt.Sprintf(`<html>
	<body>
		<h1>Welcome, Chirpy Admin</h1>
		<p>Chirpy has been visited %d times!</p>
	</body>
</html>`, cfg.fileserverHits.Load())

	w.Write([]byte(message))
}

func (cfg *apiConfig) handlerResetRequestsNumber(w http.ResponseWriter, r *http.Request) {
	fmt.Println(cfg.platform)
	if cfg.platform != "dev" {
		errMsg := "This extrememly dangerous endpoint can only be accessed in a local environment"
		respondWithError(w, 403, errMsg)
		return
	}

	defer r.Body.Close()
	_, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	err = cfg.database.DeleteAllUsers(r.Context())
	if err != nil {
		errMsg := fmt.Sprintf("Error deleting all users: %v", err)
		respondWithError(w, 500, errMsg)
		return
	}

	cfg.fileserverHits.Store(0)

	w.Header().Set("Content-Type", "text/plain; charset=utf=8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("FileserverHits set back to 0"))
}

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	type requestBody struct {
		Body string `json:"body"`
	}
	type responseBody struct {
		CleanedBody string `json:"cleaned_body"`
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		errMsg := fmt.Sprintf("Error reading request body: %v\n", err)
		respondWithError(w, 500, errMsg)
		return
	}

	params := requestBody{}
	err = json.Unmarshal(data, &params)
	if err != nil {
		errMsg := fmt.Sprintf("Error unmarshalling data: %v\n", err)
		respondWithError(w, 500, errMsg)
		return
	}

	if len(params.Body) > 140 {
		errMsg := "chirps must be at most 140 characters long"
		respondWithError(w, 400, errMsg)
		return
	}

	cleanedBody := cleanBody(params.Body)
	respondWithJSON(w, 200, responseBody{
		CleanedBody: cleanedBody,
	})
}

func (cfg *apiConfig) handlerCreateUsers(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	type requestBody struct {
		Email string `json:"email"`
	}
	type User struct {
		Id         uuid.UUID `json:"id"`
		Created_at time.Time `json:"created_at"`
		Updated_at time.Time `json:"updated_at"`
		Email      string    `json:"email"`
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
		Id:         uuid.New(),
		Created_at: time.Now(),
		Updated_at: time.Now(),
		Email:      user.Email,
	})
}
