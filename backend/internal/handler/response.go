package handler

import (
	"encoding/json"
	"log"
	"net/http"
)

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("failed to encode JSON response: %v", err)
	}
}

func respondError(w http.ResponseWriter, status int, code string, message string) {
	response := ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	}

	respondJSON(w, status, response)
}
