package main

import "strings"

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
