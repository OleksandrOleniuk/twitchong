package utils

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse represents an error API response
type ErrorResponse struct {
	Error string `json:"error"`
}

// RespondJSON sends a JSON response with the provided status code and payload
func RespondJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Error marshalling JSON"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(response)
}

// RespondError sends a JSON error response with the provided status code and message
func RespondError(w http.ResponseWriter, status int, message string) {
	RespondJSON(w, status, ErrorResponse{Error: message})
}
