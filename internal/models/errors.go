package models

import (
	"net/http"
)

// APIError represents an API error with HTTP status code and details
type APIError struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Details   string `json:"details,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

func (e *APIError) Error() string {
	return e.Message
}

func (e *APIError) HTTPStatus() int {
	return e.Code
}

func NewAPIError(code int, message, details string) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

var (
	ErrNotFound      = &APIError{Code: http.StatusNotFound, Message: "Resource not found"}
	ErrDuplicatedID  = &APIError{Code: http.StatusBadRequest, Message: "Duplicate ID provided"}
	ErrInvalidSeats  = &APIError{Code: http.StatusBadRequest, Message: "Invalid number of seats"}
	ErrInvalidInput  = &APIError{Code: http.StatusBadRequest, Message: "Invalid input provided"}
	ErrInternalError = &APIError{Code: http.StatusInternalServerError, Message: "Internal server error"}
)
