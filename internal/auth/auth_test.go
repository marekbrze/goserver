package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestValidateJwt(t *testing.T) {
	testuserid := uuid.New()

	// define a secret for signing and validating
	validsecret := "supersecret"
	invalidsecret := "wrongsecret"

	testcases := []struct {
		name           string
		tokenstring    string
		tokensecret    string
		expectederror  bool
		expecteduserid uuid.UUID
	}{
		{
			name: "valid token",
			tokenstring: func() string {
				token, _ := MakeJWT(testuserid, validsecret, time.Hour)
				return token
			}(),
			tokensecret:    validsecret,
			expectederror:  false,
			expecteduserid: testuserid,
		},
		{
			name: "expired token",
			tokenstring: func() string {
				token, _ := MakeJWT(testuserid, validsecret, -time.Hour) // token expired an hour ago
				return token
			}(),
			tokensecret:    validsecret,
			expectederror:  true,
			expecteduserid: uuid.Nil, // expecting nil uuid on error
		},
		{
			name: "invalid secret",
			tokenstring: func() string {
				token, _ := MakeJWT(testuserid, validsecret, time.Hour)
				return token
			}(),
			tokensecret:    invalidsecret, // using wrong secret
			expectederror:  true,
			expecteduserid: uuid.Nil,
		},
		{
			name:           "malformed token",
			tokenstring:    "this.is.not.a.jwt.token",
			tokensecret:    validsecret,
			expectederror:  true,
			expecteduserid: uuid.Nil,
		},
		{
			name:           "empty token string",
			tokenstring:    "",
			tokensecret:    validsecret,
			expectederror:  true,
			expecteduserid: uuid.Nil,
		},
		{
			name: "token without user_id claim",
			tokenstring: func() string {
				claims := jwt.MapClaims{
					"exp": time.Now().Add(time.Hour).Unix(),
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tkn, _ := token.SignedString([]byte(validsecret))
				return tkn
			}(),
			tokensecret:    validsecret,
			expectederror:  true,
			expecteduserid: uuid.Nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			userid, err := ValidateJWT(tc.tokenstring, tc.tokensecret)

			if tc.expectederror {
				if err == nil {
					t.Errorf("expected an error for test case '%s', but got none", tc.name)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for test case '%s': %v", tc.name, err)
				}
				if userid != tc.expecteduserid {
					t.Errorf("expected user id %s, but got %s for test case '%s'", tc.expecteduserid, userid, tc.name)
				}
			}
		})
	}
}
