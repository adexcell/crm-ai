package router

import (
	"net/http"

	"github.com/adexcell/crm-ai/internal/infra/http/middleware"
	"github.com/go-chi/chi/v5"
)

func New(mw *middleware.Middleware) *chi.Mux {
	r := chi.NewRouter()

	r.Use(mw.RequestID)
	r.Use(mw.Recovery)
	r.Use(mw.OTel)
	r.Use(mw.Metrics)
	r.Use(mw.Logger)
	r.Use(mw.RateLimit)

	// Liveness & Readiness пробы Kubernetes
	r.Get("/live", probe)
	r.Get("/ready", probe)

	return r
}

func probe(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}
