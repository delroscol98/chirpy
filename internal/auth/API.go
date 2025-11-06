package auth

import (
	"errors"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
	apiKey := strings.TrimSpace(strings.TrimLeft(headers.Get("Authorization"), "ApiKey"))
	if apiKey == "" {
		return "", errors.New("No ApiKey found")
	}

	return apiKey, nil
}
