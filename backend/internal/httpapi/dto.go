package httpapi

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"avatar/backend/internal/repo"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

// writeRepoError maps domain errors to HTTP status codes so clients can tell a
// bad payload (400) from a stock/recall conflict (409) and a missing row (404).
func writeRepoError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		writeError(w, http.StatusNotFound, err)
	case errors.Is(err, repo.ErrInsufficientStock),
		errors.Is(err, repo.ErrLotBlocked):
		writeError(w, http.StatusConflict, err)
	case errors.Is(err, repo.ErrInvalidOrder),
		errors.Is(err, repo.ErrInvalidStockMovement):
		writeError(w, http.StatusBadRequest, err)
	default:
		writeError(w, http.StatusInternalServerError, err)
	}
}
