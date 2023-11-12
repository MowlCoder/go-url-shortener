// Package httputil
// contains useful utils function to send response to the client.
package httputil

import (
	"encoding/json"
	"io"
	"net/http"
)

// HTTPError is representing error that sends to user.
type HTTPError struct {
	Error string `json:"error"`
}

// SendTextResponse send response with content-type text/plain to the client
// and given status code.
func SendTextResponse(w http.ResponseWriter, code int, text string) error {
	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(code)

	if _, err := io.WriteString(w, text); err != nil {
		return err
	}

	return nil
}

// SendJSONResponse send response with content-type application/json to the client
// and given status code.
func SendJSONResponse(w http.ResponseWriter, code int, data interface{}) error {
	w.Header().Set("content-type", "application/json")

	jsonData, err := json.Marshal(data)

	if err != nil {
		return err
	}

	w.WriteHeader(code)

	if _, err := w.Write(jsonData); err != nil {
		return err
	}

	return nil
}

// SendJSONErrorResponse send error response with content-type application/json to the client
// and given status code. Error structure is HTTPError.
func SendJSONErrorResponse(w http.ResponseWriter, code int, error string) error {
	return SendJSONResponse(w, code, HTTPError{Error: error})
}

// SendRedirectResponse send redirect response with status code 307
// and given location in header Location.
func SendRedirectResponse(w http.ResponseWriter, location string) {
	w.Header().Set("Location", location)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// SendStatusCode send given status code to client.
func SendStatusCode(w http.ResponseWriter, code int) {
	w.WriteHeader(code)
}
