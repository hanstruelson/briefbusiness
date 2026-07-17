package utilities

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// DecodeJSON reads the request body and unmarshals it into the target struct.
// It returns the parsed value and a boolean indicating success.
// If it fails, it writes the error message directly to the response writer.
func DecodeJSON[T any](w http.ResponseWriter, r *http.Request, target *T) (bool, T) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error reading request body")
		return false, *target // Return zero value of T
	}

	err = json.Unmarshal(body, target)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid JSON: %v", err)
		return false, *target
	}

	return true, *target
}
