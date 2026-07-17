package utilities

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// DecodeJSON reads the request body and unmarshals it into the target struct.
// It returns the parsed value and a boolean indicating success.
// If it fails, it writes the error message directly to the response writer.
func SendJSON[T any](w http.ResponseWriter, data *T) {
	responseBytes, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error marshaling JSON")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseBytes)
}
