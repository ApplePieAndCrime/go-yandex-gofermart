package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/middleware"
	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/model"
	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/repository"
)

func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeJSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	balance, err := h.services.GetBalance(r.Context(), userID)
	if err != nil {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(balance)
}

func (h *Handler) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeJSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req model.WithdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid request format", http.StatusBadRequest)
		return
	}

	if req.Order == "" || req.Sum <= 0 {
		writeJSONError(w, "invalid request", http.StatusBadRequest)
		return
	}

	err := h.services.Withdraw(r.Context(), userID, req.Order, req.Sum)
	if err != nil {
		switch {
		case err.Error() == "invalid order number":
			writeJSONError(w, "invalid order number", http.StatusUnprocessableEntity)
		case errors.Is(err, repository.ErrInsufficientFunds):
			writeJSONError(w, "insufficient funds", http.StatusPaymentRequired)
		default:
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeJSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	withdrawals, err := h.services.GetWithdrawals(r.Context(), userID)
	if err != nil {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(withdrawals)
}
