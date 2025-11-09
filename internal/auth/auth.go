// Package auth is used for storing and checking passwords
package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func HashPassword(password string) (string, error) {
	hashedPassword, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}
	return hashedPassword, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	isCorrect, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}
	return isCorrect, nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject:   userID.String(),
	}
	newJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtString, err := newJWT.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}
	return jwtString, nil
}

func ValidateJWT(tokenString string, tokenSecret string) (uuid.UUID, error) {
	parsedClaims := &jwt.RegisteredClaims{}
	keyFunc := func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		fmt.Println("keyFunc call")
		return []byte(tokenSecret), nil // Use the loaded secret
	}
	fmt.Println(tokenString)
	token, err := jwt.ParseWithClaims(tokenString, parsedClaims, keyFunc)
	if err != nil {
		return uuid.Nil, fmt.Errorf("error parsing token: %w", err) // Use %w for wrapping errors
	}

	if !token.Valid {
		return uuid.Nil, errors.New("invalid JWT token")
	}

	subject, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, fmt.Errorf("error extracting token subject (user ID): %w", err)
	}

	userID, err := uuid.Parse(subject)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid UUID format in token subject: %w", err)
	}

	return userID, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	tokenFromHeader := headers.Get("Authorization")
	if strings.HasPrefix(tokenFromHeader, "Bearer ") {
		modifiedString := tokenFromHeader[len("Bearer "):]
		return modifiedString, nil
	} else {
		return "", fmt.Errorf("header is not bearer type")
	}
}
