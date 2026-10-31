package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/adexcell/crm-ai/internal/infra/metrics"
	"github.com/go-chi/chi/v5"
)

func HTTP(m *metrics.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			wrapped := WriterWrapper(w)
			next.ServeHTTP(wrapped, r)

			pattern := ""
			if rctx := chi.RouteContext(r.Context()); rctx != nil {
				pattern = rctx.RoutePattern()
			}
			if pattern == "" {
				pattern = r.URL.Path
			}

			statusStr := strconv.Itoa(wrapped.Code())
			m.Observe(r.Method, pattern, statusStr, time.Since(start))
		})
	}
}
