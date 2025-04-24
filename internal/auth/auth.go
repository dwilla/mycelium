package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	datastar "github.com/starfederation/datastar/sdk/go"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	newHash, err := bcrypt.GenerateFromPassword([]byte(password), 0)
	if err != nil {
		return "", err
	}
	return string(newHash), nil
}

func CheckPasswordHash(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func GetBearerToken(r *http.Request) (string, error) {
	signal := struct {
		Bearer string `json:"token"`
	}{}
	if err := datastar.ReadSignals(r, &signal); err != nil {
		return "", fmt.Errorf("no bearer token present in signals")
	}
	if signal.Bearer == "" {
		return "", fmt.Errorf("bearer token is blank")
	}
	return signal.Bearer, nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "mycelium-chat",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   userID.String(),
	})
	signedString, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}

	return signedString, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.Nil, err
	}

	id, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}
	parsedId, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, err
	}
	return parsedId, nil
}
