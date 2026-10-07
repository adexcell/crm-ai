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
