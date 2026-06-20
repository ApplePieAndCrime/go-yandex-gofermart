package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/model"
	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/repository"
)

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	}

	if req.Login == "" || req.Password == "" {
		http.Error(w, "login and password are required", http.StatusBadRequest)
		return
	}

	userID, token, err := h.services.RegisterUser(r.Context(), req.Login, req.Password)
	if err != nil {
		if errors.Is(err, repository.ErrLoginAlreadyExists) {
			http.Error(w, "login already exists", http.StatusConflict)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id": userID,
		"token":   token,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	}

	if req.Login == "" || req.Password == "" {
		http.Error(w, "login and password are required", http.StatusBadRequest)
		return
	}

	userID, token, err := h.services.LoginUser(r.Context(), req.Login, req.Password)
	if err != nil {
		if err.Error() == "invalid credentials" {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id": userID,
		"token":   token,
	})
}
