package auth

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func TestTokens(t *testing.T) {
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("warning: assuming default configuration. .env unreadable: %v", err)
	}
	secret := os.Getenv("TOKEN_SECRET")
	testUuid := uuid.New()
	token, err := MakeJWT(testUuid, secret, 2*time.Second)
	if err != nil {
		t.Error("error making token: ", err)
	}

	validUser, err := ValidateJWT(token, secret)
	if err != nil {
		t.Error("error validating token: ", err)
	}
	if validUser != testUuid {
		t.Errorf("uuid's don't match, %v is not %v", validUser, testUuid)
	}
}
