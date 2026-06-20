package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/auth"
	"go.uber.org/zap"
)

type contextKey string

const UserIDKey contextKey = "userID"

func Auth(logger *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeJSONError(w, "missing authorization header", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				writeJSONError(w, "invalid authorization header format", http.StatusUnauthorized)
				return
			}
			tokenString := parts[1]

			userID, err := auth.ValidateToken(tokenString)
			if err != nil {
				logger.Debugw("invalid token", "error", err)
				writeJSONError(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserID(ctx context.Context) (int, bool) {
	val := ctx.Value(UserIDKey)
	if val == nil {
		return 0, false
	}
	userID, ok := val.(int)
	return userID, ok
}
