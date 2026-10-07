package health

import (
	"context"
	"net/http"
	"time"

	"github.com/adexcell/crm-ai-pet/internal/database"
	"github.com/adexcell/crm-ai-pet/pkg/response"
	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
)

type Handler struct {
	db  *database.Pool
	rdb *redis.Client
}

func NewHandler(db *database.Pool, rdb *redis.Client) *Handler {
	return &Handler{db: db, rdb: rdb}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/health", func(r chi.Router) {
		r.Get("/live", h.Liveness)
		r.Get("/ready", h.Readiness)
	})
}

func (h *Handler) Liveness(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]any{
		"status": "alive",
	})
}

func (h *Handler) Readiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	checks := make(map[string]any)
	isHealthy := true

	if err := h.db.Ping(ctx); err != nil {
		checks["postgres"] = "error" + err.Error()
		isHealthy = false
	} else {
		checks["postgres"] = "ok"
	}

	if err := h.rdb.Ping(ctx).Err(); err != nil {
		checks["redis"] = "error" + err.Error()
		isHealthy = false
	} else {
		checks["redis"] = "ok"
	}

	if isHealthy {
		response.JSON(w, http.StatusOK, checks)
	} else {
		response.Error(
			w,
			http.StatusServiceUnavailable,
			"SERVICE_UNAVAILABLE",
			"One or more dependencies are unhealthy",
			checks,
		)
	}
}
