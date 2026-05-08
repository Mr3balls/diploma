package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"esports-backend/internal/apperror"
)

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, err error) {
	log.Printf("HTTP error: %+v", err)

	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		writeJSON(w, appErr.HTTPStatus, map[string]interface{}{"error": appErr})
		return
	}

	writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
		"error": map[string]interface{}{
			"code":    "internal_error",
			"message": err.Error(),
		},
	})
}

func decodeJSON(r *http.Request, dst interface{}) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}
