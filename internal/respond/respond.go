// Package respond provides shared JSON response helpers used by all HTTP handlers.
// Centralising JSON writes here means handlers never call json.NewEncoder directly,
// keeping error handling consistent and Content-Type always correct.
//
// Dependency graph: respond ← handlers ← api (router)
// No package in the graph imports respond's dependents, so there are no cycles.
package respond

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// envelope wraps every JSON response so clients always get a consistent shape:
//
//	{"data": ...}  or  {"error": "...", "details": {...}}
type envelope map[string]any

// JSON serialises v as JSON, sets Content-Type, and writes the given HTTP status.
// Logs and returns an error if serialisation fails.
func JSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ") // pretty-print; remove SetIndent for bandwidth-sensitive production
	if err := enc.Encode(envelope{"data": v}); err != nil {
		slog.Error("json encode failed", "err", err)
		return err
	}
	return nil
}

// Error writes a JSON error body with a human-readable message.
//
//	{"error": "not found"}
func Error(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(envelope{"error": message})
}

// ValidationError writes a 422 JSON body containing field-level validation messages.
//
//	{"error": "validation failed", "details": {"title": "required"}}
func ValidationError(w http.ResponseWriter, errs map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity)

	_ = json.NewEncoder(w).Encode(envelope{
		"error":   "validation failed",
		"details": errs,
	})
}
