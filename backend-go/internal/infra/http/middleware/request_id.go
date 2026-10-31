package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type ctxKey string

const requestIDKey ctxKey = "request_id"
const requestHeader = "X-Request-ID"
const correlationHeader = "X-Correlation-ID"

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(requestHeader)
		if id == "" {
			id = r.Header.Get(correlationHeader)
		}
		if id == "" {
			id = uuid.NewString()
		}

		ctx := context.WithValue(r.Context(), requestIDKey, id)
		w.Header().Set(requestHeader, id)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func IDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}
