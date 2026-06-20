package auth

import (
	"os"
	"testing"
)

func TestGenerateAndValidate(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	if err := InitJWT(); err != nil {
		t.Fatal(err)
	}

	userID := 123
	token, err := GenerateToken(userID)
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}
	if token == "" {
		t.Error("token is empty")
	}

	gotUserID, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken error: %v", err)
	}
	if gotUserID != userID {
		t.Errorf("userID mismatch: got %d, want %d", gotUserID, userID)
	}

	_, err = ValidateToken("invalid.token.here")
	if err == nil {
		t.Error("expected error for invalid token, got nil")
	}
}
