package api

import (
	"errors"
	"net/http"
)

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func (e *APIError) Error() string {
	return e.Message
}

// serviceCoder is satisfied by application.ServiceError without api
// needing to import the application package.
type serviceCoder interface {
	error
	ServiceCode() string
}

var serviceErrorStatus = map[string]int{
	"player_not_found":              http.StatusNotFound,
	"player_in_active_game":         http.StatusConflict,
	"player_already_in_active_game": http.StatusConflict,
	"player_not_in_game":            http.StatusConflict,
	"invalid_gross_score":           http.StatusBadRequest,
	"hole_out_of_range":             http.StatusBadRequest,
	"wrong_hole":                    http.StatusConflict,
	"game_finished":                 http.StatusConflict,
	"validation_error":              http.StatusBadRequest,
	"invalid_team_size":             http.StatusBadRequest,
	"invalid_game_type":             http.StatusBadRequest,
	"invalid_game_params":           http.StatusBadRequest,
}

func StatusCode(err error) int {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		switch apiErr.Code {
		case "not_found":
			return http.StatusNotFound
		case "bad_request":
			return http.StatusBadRequest
		case "conflict":
			return http.StatusConflict
		case "external_api_error":
			return http.StatusBadGateway
		default:
			return http.StatusInternalServerError
		}
	}

	var svcErr serviceCoder
	if errors.As(err, &svcErr) {
		if status, ok := serviceErrorStatus[svcErr.ServiceCode()]; ok {
			return status
		}
		return http.StatusUnprocessableEntity
	}

	return http.StatusInternalServerError
}
