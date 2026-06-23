package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/model"
	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/repository"
	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/response"
)

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSONError(w, "invalid request format", http.StatusBadRequest)
		return
	}

	if req.Login == "" || req.Password == "" {
		response.JSONError(w, "login and password are required", http.StatusBadRequest)
		return
	}

	userID, token, err := h.services.RegisterUser(r.Context(), req.Login, req.Password)
	if err != nil {
		if errors.Is(err, repository.ErrLoginAlreadyExists) {
			response.JSONError(w, "login already exists", http.StatusConflict)
			return
		}
		response.JSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)

	h.JsonEncode(w, map[string]interface{}{
		"user_id": userID,
		"token":   token,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSONError(w, "invalid request format", http.StatusBadRequest)
		return
	}

	if req.Login == "" || req.Password == "" {
		response.JSONError(w, "login and password are required", http.StatusBadRequest)
		return
	}

	userID, token, err := h.services.LoginUser(r.Context(), req.Login, req.Password)
	if err != nil {
		if err.Error() == "invalid credentials" {
			response.JSONError(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		response.JSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	h.JsonEncode(w, map[string]interface{}{
		"user_id": userID,
		"token":   token,
	})
}
