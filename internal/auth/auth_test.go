package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestHashPassword tests the hashing and verification of a password
func TestHashPassword(t *testing.T) {
	password := "fh2ojtg4j39gj910"
	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Error(err)
	}

	if hashedPassword == "" {
		t.Error("Hashed password is empty")
	}

	isValid, err := VerifyPassword(password, hashedPassword)
	if err != nil {
		t.Error(err)
	}

	if !isValid {
		t.Error("Password and hashed password do not match")
	}
}

func TestValidateJWTValidToken(t *testing.T) {
	testUUID, err := uuid.NewRandom()
	if err != nil {
		t.Error(err)
	}

	token, err := MakeJWT(testUUID, "secret", "chirpy", 1*time.Hour)
	if err != nil {
		t.Error(err)
	}

	userID, err := ValidateJWT(token, "secret", "chirpy")
	if err != nil {
		t.Error(err)
	}

	if userID != testUUID {
		t.Errorf("Expected user ID %s, got %s", testUUID, userID)
	}
}

func TestValidateJWTInvalidSecret(t *testing.T) {
	testUUID, err := uuid.NewRandom()
	if err != nil {
		t.Error(err)
	}

	token, err := MakeJWT(testUUID, "secret", "chirpy", 1*time.Hour)
	if err != nil {
		t.Error(err)
	}

	_, err = ValidateJWT(token, "invalid", "chirpy")
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestValidateJWTInvalidIssuer(t *testing.T) {
	testUUID, err := uuid.NewRandom()
	if err != nil {
		t.Error(err)
	}

	token, err := MakeJWT(testUUID, "secret", "chirpy", 1*time.Hour)
	if err != nil {
		t.Error(err)
	}

	_, err = ValidateJWT(token, "secret", "invalid")
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestValidateJWTExpiredToken(t *testing.T) {
	testUUID, err := uuid.NewRandom()
	if err != nil {
		t.Error(err)
	}

	token, err := MakeJWT(testUUID, "secret", "chirpy", -1*time.Hour)
	if err != nil {
		t.Error(err)
	}

	_, err = ValidateJWT(token, "secret", "chirpy")
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestGetBearerToken(t *testing.T) {
	headers := http.Header{}
	headers.Add("Authorization", "Bearer TEST.TOKEN.LOCAL")
	token, err := GetBearerToken(headers)
	if err != nil {
		t.Error(err)
	}
	if token != "TEST.TOKEN.LOCAL" {
		t.Errorf("Expected token TEST.TOKEN.LOCAL, got %s", token)
	}
}

func TestGetBearerTokenEmpty(t *testing.T) {
	headers := http.Header{}
	_, err := GetBearerToken(headers)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}
