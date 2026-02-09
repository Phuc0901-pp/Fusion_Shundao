package errors

import (
	"encoding/json"
	"net/http"
)

// AppError represents a structured API error
type AppError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	return e.Message
}

// Common error codes
const (
	ErrCodeUnauthorized   = "UNAUTHORIZED"
	ErrCodeNotFound       = "NOT_FOUND"
	ErrCodeBadRequest     = "BAD_REQUEST"
	ErrCodeInternal       = "INTERNAL_ERROR"
	ErrCodeRateLimited    = "RATE_LIMITED"
	ErrCodeSessionExpired = "SESSION_EXPIRED"
	ErrCodeFetchFailed    = "FETCH_FAILED"
)

// Predefined errors
var (
	ErrUnauthorized   = &AppError{Code: ErrCodeUnauthorized, Message: "Authentication required"}
	ErrNotFound       = &AppError{Code: ErrCodeNotFound, Message: "Resource not found"}
	ErrBadRequest     = &AppError{Code: ErrCodeBadRequest, Message: "Invalid request"}
	ErrInternal       = &AppError{Code: ErrCodeInternal, Message: "Internal server error"}
	ErrSessionExpired = &AppError{Code: ErrCodeSessionExpired, Message: "Session has expired, please re-login"}
	ErrFetchFailed    = &AppError{Code: ErrCodeFetchFailed, Message: "Failed to fetch data from source"}
)

// NewAppError creates a new AppError
func NewAppError(code, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// WithDetails adds details to an error
func (e *AppError) WithDetails(details interface{}) *AppError {
	return &AppError{
		Code:    e.Code,
		Message: e.Message,
		Details: details,
	}
}

// ErrorResponse wraps an error for JSON response
type ErrorResponse struct {
	Error *AppError `json:"error"`
}

// WriteError writes an error response to the http.ResponseWriter
func WriteError(w http.ResponseWriter, err *AppError, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Error: err})
}

// HTTP Status helpers
func WriteUnauthorized(w http.ResponseWriter, err *AppError) {
	WriteError(w, err, http.StatusUnauthorized)
}

func WriteNotFound(w http.ResponseWriter, err *AppError) {
	WriteError(w, err, http.StatusNotFound)
}

func WriteBadRequest(w http.ResponseWriter, err *AppError) {
	WriteError(w, err, http.StatusBadRequest)
}

func WriteInternalError(w http.ResponseWriter, err *AppError) {
	WriteError(w, err, http.StatusInternalServerError)
}

func WriteTooManyRequests(w http.ResponseWriter, err *AppError) {
	WriteError(w, err, http.StatusTooManyRequests)
}
