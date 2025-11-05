package auth

import "testing"

func TestHashPassword(t *testing.T) {
	password1 := "asdfghjkl1234567890!"
	password2 := "1234567890!asdfghjkl"
	hash1, err := HashPassword(password1)
	if err != nil {
		t.Errorf("Error hashing password 1: %v", err)
	}
	hash2, err := HashPassword(password2)
	if err != nil {
		t.Errorf("Error hashing password 2: %v", err)
	}

	type Case struct {
		name          string
		password      string
		hash          string
		wantErr       bool
		matchPassword bool
	}

	cases := []Case{
		{
			name:          "Correct Password",
			password:      password1,
			hash:          hash1,
			wantErr:       false,
			matchPassword: true,
		},
		{
			name:          "Incorrect Password",
			password:      "wrongPassword",
			hash:          hash1,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "Passwords don't match",
			password:      password1,
			hash:          hash2,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "Empty Password",
			password:      "",
			hash:          "invalid hash",
			wantErr:       true,
			matchPassword: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			match, err := CheckPasswordHash(c.password, c.hash)
			if (err != nil) != c.wantErr {
				t.Errorf("CheckPasswordHash() error = %v, want %v", err, c.wantErr)
			}
			if !c.wantErr && match != c.matchPassword {
				t.Errorf("CheckPassword() expected %v, got %v", c.matchPassword, match)
			}
		})
	}
}
