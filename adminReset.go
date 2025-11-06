package main

import (
	"fmt"
	"net/http"
)

func (cfg *apiConfig) handlerResetRequestsNumber(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		errMsg := "This extrememly dangerous endpoint can only be accessed in a local environment"
		respondWithError(w, 403, errMsg)
		return
	}

	err := cfg.database.DeleteAllUsers(r.Context())
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
