package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func cleanBody(body string) string {
	bodyWords := strings.Split(body, " ")
	profraneWords := []string{"kerfuffle", "sharbert", "fornax"}

	for _, profaneWord := range profraneWords {
		for index, bodyWord := range bodyWords {
			if profaneWord == strings.ToLower(bodyWord) {
				bodyWords[index] = "****"
			}
		}
	}

	return strings.Join(bodyWords, " ")
}

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
