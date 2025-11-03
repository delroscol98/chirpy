package main

import (
	"fmt"
	"io"
	"net/http"
)

// NOTE: GET requests
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

// NOTE: POST requests
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
