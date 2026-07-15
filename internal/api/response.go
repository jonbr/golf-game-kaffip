package api

import (
	"encoding/json"
	"net/http"
)

// JSON writes any response body as JSON with the given status code.
func JSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if body != nil {
		_ = json.NewEncoder(w).Encode(body)
	}
}

// WriteError writes an APIError using the correct HTTP status code.
func WriteError(w http.ResponseWriter, err error) {

	status := StatusCode(err)
	JSON(w, status, err)
}

func WriteBadRequest(w http.ResponseWriter, code string, msg string, details any) {
	JSON(w, http.StatusBadRequest, APIError{
		Code:    code,
		Message: msg,
		Details: details,
	})
}

func WriteNotFound(w http.ResponseWriter, code string, msg string, details any) {
	JSON(w, http.StatusNotFound, APIError{
		Code:    code,
		Message: msg,
		Details: details,
	})
}

func WriteConflict(w http.ResponseWriter, code string, msg string, details any) {
	JSON(w, http.StatusConflict, APIError{
		Code:    code,
		Message: msg,
		Details: details,
	})
}

func WriteInternal(w http.ResponseWriter, msg string, details any) {
	JSON(w, http.StatusInternalServerError, APIError{
		Code:    "internal_error",
		Message: msg,
		Details: details,
	})
}

func WriteExternal(w http.ResponseWriter, err error, details any) {
	JSON(w, http.StatusBadGateway, APIError{
		Code:    "external_api_error",
		Message: err.Error(),
		Details: details,
	})
}
