package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/middleware"
	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/response"
	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/service"
)

func (h *Handler) UploadOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.JSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		response.JSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	orderNumber := string(body)

	status, err := h.services.UploadOrder(r.Context(), userID, orderNumber)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidOrderNumber):
			response.JSONError(w, "invalid order number", http.StatusUnprocessableEntity)
		case errors.Is(err, service.ErrOrderAlreadyExists):
			response.JSONError(w, "order already uploaded by another user", http.StatusConflict)
		default:
			h.logger.Errorw("upload order failed", "error", err)
			response.JSONError(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if status == service.OrderNewAccepted {
		w.WriteHeader(http.StatusAccepted)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func (h *Handler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.JSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	orders, err := h.services.GetUserOrders(r.Context(), userID)
	if err != nil {
		response.JSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(orders)
}
