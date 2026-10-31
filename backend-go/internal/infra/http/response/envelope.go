package response

import (
	"encoding/json"
	"net/http"
)

type ErrorDetail struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

type ApiResponse[T any] struct {
	OK    bool         `json:"ok"`
	Data  *T           `json:"data"`
	Error *ErrorDetail `json:"error"`
}

func JSON[T any](w http.ResponseWriter, statusCode int, data T) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	resp := ApiResponse[T]{
		OK:    true,
		Data:  &data,
		Error: nil,
	}

	_ = json.NewEncoder(w).Encode(resp)
}

func Error(w http.ResponseWriter, statusCode int, code string, message string, details map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	resp := ApiResponse[any]{
		OK:   false,
		Data: nil,
		Error: &ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
	}

	_ = json.NewEncoder(w).Encode(resp)
}

func BadRequest(w http.ResponseWriter, code string, message string) {
	Error(w, http.StatusBadRequest, code, message, nil)
}

// InvalidJSON — 400 Bad Request "Invalid JSON format in request body"
func InvalidJSON(w http.ResponseWriter) {
	Error(w, http.StatusBadRequest, CodeInvalidBody, "Invalid JSON format in request body", nil)
}

// ValidationError — 422 Unprocessable Entity
func ValidationError(w http.ResponseWriter, details map[string]any) {
	Error(w, http.StatusUnprocessableEntity, CodeValidationError, "Validation failed", details)
}

// InternalError — 500 Internal Server Error "An internal server error occurred"
func InternalError(w http.ResponseWriter, message string) {
	if message == "" {
		message = "An internal server error occurred"
	}
	Error(w, http.StatusInternalServerError, CodeInternalError, message, nil)
}

// Unauthorized — 401 Unauthorized "Authentication required"
func Unauthorized(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Authentication required"
	}
	Error(w, http.StatusUnauthorized, CodeUnauthorized, message, nil)
}

// Forbidden — 403 Forbidden "Access forbidden"
func Forbidden(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Access forbidden"
	}
	Error(w, http.StatusForbidden, CodeForbidden, message, nil)
}

// NotFound — 404 Not Found "Resource not found"
func NotFound(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Resource not found"
	}
	Error(w, http.StatusNotFound, CodeNotFound, message, nil)
}
