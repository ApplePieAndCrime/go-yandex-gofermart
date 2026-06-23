package auth

import (
	"testing"
)

func TestGenerateAndValidate(t *testing.T) {
	manager, err := NewJWTManager("test-secret")
	if err != nil {
		t.Fatalf("NewJWTManager error: %v", err)
	}

	userID := 123
	token, err := manager.GenerateToken(userID)
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}
	if token == "" {
		t.Error("token is empty")
	}

	gotUserID, err := manager.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken error: %v", err)
	}
	if gotUserID != userID {
		t.Errorf("ValidateToken returned %d, want %d", gotUserID, userID)
	}
}
