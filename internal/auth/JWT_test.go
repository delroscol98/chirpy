package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWT(t *testing.T) {
	userID := uuid.New()
	validToken, err := MakeJWT(userID, "secret", time.Hour)
	if err != nil {
		t.Error(err)
	}

	type Case struct {
		name        string
		tokenString string
		tokenSecret string
		wantUserID  uuid.UUID
		wantErr     bool
	}

	cases := []Case{
		{
			name:        "Valid token",
			tokenString: validToken,
			tokenSecret: "secret",
			wantUserID:  userID,
			wantErr:     false,
		},
		{
			name:        "Invalid token",
			tokenString: "invalid.token.string",
			tokenSecret: "secret",
			wantUserID:  uuid.Nil,
			wantErr:     true,
		},
		{
			name:        "Wrong secret",
			tokenString: validToken,
			tokenSecret: "wrong_secret",
			wantUserID:  uuid.Nil,
			wantErr:     true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gotUserID, err := ValidateJWT(c.tokenString, c.tokenSecret)
			if (err != nil) != c.wantErr {
				t.Errorf("ValidatedJWT() error = %v, wanterr %v", err, c.wantErr)
				return
			}

			if gotUserID != c.wantUserID {
				t.Errorf("ValidateJWT() gotUserID = %v, want %v", gotUserID, c.wantUserID)
			}
		})
	}
}

func TestGetBearerToken(t *testing.T) {
	type Case struct {
		name               string
		authorizationSlice map[string][]string
		bearer             string
		wantErr            bool
	}

	cases := []Case{
		{
			name: "Bearer Found",
			authorizationSlice: map[string][]string{
				"Authorization": {"Bearer 1234567890"},
			},
			bearer:  "1234567890",
			wantErr: false,
		},
		{
			name: "Bearer Not Found",
			authorizationSlice: map[string][]string{
				"test": {"Bearer not found"},
			},
			bearer:  "",
			wantErr: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			bearer, err := GetBearerToken(c.authorizationSlice)
			if (err != nil) && !c.wantErr {
				t.Errorf("GetBearerToken error = %v, wantErr = %v", err, c.wantErr)
			}

			if bearer != c.bearer {
				t.Errorf("Bearers don't match. Expected: %v, Actual: %v", c.bearer, bearer)
			}
		})
	}
}
