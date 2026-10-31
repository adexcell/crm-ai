package middleware

import (
	"net"
	"net/http"

	"github.com/adexcell/crm-ai/internal/infra/http/response"
	"golang.org/x/time/rate"
)

type IPLimiter struct {
	limiters typedSyncMap[string, *rate.Limiter]
	rps      rate.Limit
	burst    int
}

func NewIPLimiter(rps rate.Limit, burst int) *IPLimiter {
	return &IPLimiter{
		rps:   rps,
		burst: burst,
	}
}

func (il *IPLimiter) get(ip string) *rate.Limiter {
	if l, ok := il.limiters.Load(ip); ok {
		return l
	}

	l := rate.NewLimiter(il.rps, il.burst)
	actual, _ := il.limiters.LoadOrStore(ip, l)

	return actual
}

func RateLimit(il *IPLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = r.RemoteAddr
			}

			if !il.get(ip).Allow() {
				response.Error(
					w,
					http.StatusTooManyRequests,
					response.CodeTooManyRequests,
					"too many requests",
					nil,
				)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
