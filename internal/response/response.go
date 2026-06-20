package response

import (
	"encoding/json"
	"net/http"
)

// JSONError отправляет ответ с ошибкой в формате JSON.
// Используется во всех хендлерах и middleware для единообразия.
func JSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
