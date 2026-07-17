package utilities

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

func SendError(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	errorResponse := ErrorResponse{Message: message}
	responseBytes, _ := json.Marshal(errorResponse)
	w.Write(responseBytes)
}
