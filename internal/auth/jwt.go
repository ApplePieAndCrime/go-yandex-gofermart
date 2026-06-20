package auth

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	secretKey []byte
)

type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

func InitJWT() error {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default-secret-key"
		log.Println("WARNING: JWT_SECRET not set, using default key (not secure)")
	}
	secretKey = []byte(secret)
	return nil
}

// GenerateToken создаёт JWT для указанного userID.
func GenerateToken(userID int) (string, error) {
	if secretKey == nil {
		return "", errors.New("JWT not initialized")
	}
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

// ValidateToken проверяет и извлекает userID из токена.
func ValidateToken(tokenString string) (int, error) {
	if secretKey == nil {
		return 0, errors.New("JWT not initialized")
	}
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		return 0, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return 0, errors.New("invalid token")
	}
	return claims.UserID, nil
}
