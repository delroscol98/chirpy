package main

import (
	"fmt"
	"io"
	"net/http"
)

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	fmt.Printf("Received request body from handlerReadiness: %s\n", string(body))

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (cfg *apiConfig) handlerWriteRequestsNumber(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	fmt.Printf("Received request body from handlerWriteRequestsNumber: %s\n", string(body))

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
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	fmt.Printf("Recieved request body from handlerResetRequestsNumber: %s\n", string(body))

	cfg.fileserverHits.Store(0)

	w.Header().Set("Content-Type", "text/plain; charset=utf=8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("FileserverHits set back to 0"))
}
