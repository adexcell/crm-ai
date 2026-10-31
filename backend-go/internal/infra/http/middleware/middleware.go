package middleware

import (
	"net/http"

	"github.com/adexcell/crm-ai/internal/infra/jwt"
	"github.com/adexcell/crm-ai/internal/infra/metrics"
	"golang.org/x/time/rate"
)

type Middleware struct {
	RequestID func(http.Handler) http.Handler
	Recovery  func(http.Handler) http.Handler
	OTel      func(http.Handler) http.Handler
	Metrics   func(http.Handler) http.Handler
	Logger    func(http.Handler) http.Handler
	RateLimit func(http.Handler) http.Handler
	Auth      func(http.Handler) http.Handler
}

func New(jwtManager *jwt.Manager, metricsServer *metrics.Metrics) *Middleware {
	// Лимит: 20 запросов в секунду с одного IP, всплеск (burst) до 50
	limiter := NewIPLimiter(rate.Limit(20), 50)
	return &Middleware{
		RequestID: RequestID,
		Recovery:  Recovery,
		OTel:      OTel("crm-ai"),
		Metrics:   HTTP(metricsServer),
		Logger:    Logger,
		RateLimit: RateLimit(limiter),
		Auth:      Auth(jwtManager),
	}
}
