package middleware

import (
	"encoding/json"
	"net/http"

	"esports-backend/internal/apperror"
)

func writeError(w http.ResponseWriter, err *apperror.AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.HTTPStatus)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err,
	})
}
