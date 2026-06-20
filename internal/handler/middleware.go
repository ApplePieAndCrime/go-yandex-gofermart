package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/auth"
	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/response"
)

func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			response.JSONError(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.JSONError(w, "Invalid authorization header", http.StatusUnauthorized)
			return
		}
		token := parts[1]

		userID, err := auth.ValidateToken(token)
		if err != nil {
			response.JSONError(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "userID", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
