package auth

import "testing"

func TestHashPassword(t *testing.T) {
	type Case struct {
		password string
		email    string
	}

	cases := []Case{
		{
			password: "04234",
			email:    "lane@example.com",
		},
		{
			password: "password",
			email:    "collin@example.com",
		},
		{
			password: "password1234",
			email:    "tony@example.com",
		},
	}

	for _, c := range cases {
		pw := c.password
		hashedPw, err := HashPassword(pw)
		if err != nil {
			t.Error(err)
		}

		bool, err := CheckPasswordHash(pw, hashedPw)
		if err != nil {
			t.Error(err)
		}

		if !bool {
			t.Error("Failed test: Password and Hashed Password do not match")
		}
	}
}
