package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/adexcell/crm-ai/internal/infra/http/response"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("panic recovered",
					"panic", rec,
					"method", r.Method,
					"path", r.URL.Path,
					"stack", string(debug.Stack()),
				)
				response.Error(
					w,
					http.StatusInternalServerError,
					response.CodeInternalError,
					"internal sever error",
					nil,
				)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
