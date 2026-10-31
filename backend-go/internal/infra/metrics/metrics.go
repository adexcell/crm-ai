package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
}

func New() *Metrics {
	return &Metrics{
		requestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "crm",
				Subsystem: "http",
				Name:      "requests_total",
				Help:      "Total number of HTTP requests processed by method, path and status code.",
			},
			[]string{"method", "path", "status"},
		),
		requestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "crm",
				Subsystem: "http",
				Name:      "request_duration_seconds",
				Help:      "HTTP request latency distributions in seconds.",
				Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0},
			},
			[]string{"method", "path", "status"},
		),
	}
}

func (m *Metrics) Observe(method, path, status string, duration time.Duration) {
	m.requestsTotal.WithLabelValues(method, path, status).Inc()

	m.requestDuration.WithLabelValues(method, path, status).Observe(duration.Seconds())
}
