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
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		writeJSON(w, appErr.HTTPStatus, map[string]interface{}{"error": appErr})
		return
	}

	log.Printf("internal error: %+v", err)
	writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
		"error": map[string]interface{}{
			"code":    "internal_error",
			"message": "internal server error",
		},
	})
}

func decodeJSON(r *http.Request, dst interface{}) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return apperror.BadRequest("invalid_body", "invalid request body", nil)
	}
	return nil
}
