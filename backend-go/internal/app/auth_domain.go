package app

import (
	"fmt"

	"github.com/adexcell/crm-ai/internal/config"
	"github.com/adexcell/crm-ai/internal/modules/auth/adapter/postgres"
	"github.com/adexcell/crm-ai/internal/modules/auth/adapter/redis"
	"github.com/adexcell/crm-ai/internal/modules/auth/controller/httprouter"

	"github.com/adexcell/crm-ai/internal/modules/auth/service"
	"github.com/go-chi/chi/v5"
)

// AuthDomain представляет собой агрегат домена аутентификации.
type AuthDomain struct {
	Service *service.Service
}

func NewAuthDomain(deps *Dependencies, cfg *config.Config) (*AuthDomain, error) {
	if deps == nil || deps.Postgres == nil || deps.Redis == nil ||
		deps.RouterHTTP == nil || deps.JWTManager == nil || deps.TXManager == nil {
		return nil, fmt.Errorf("auth domain: missing required dependencies in DI container")
	}

	pgAdapter := postgres.New(deps.Postgres)
	redisAdapter := redis.New(deps.Redis)

	authService := service.New(
		pgAdapter,
		redisAdapter,
		deps.TXManager,
		deps.JWTManager,
		cfg,
	)

	// Монтируем роуты авторизации под глобальный префикс /api (/api/v1/auth/...)
	deps.RouterHTTP.Route("/api", func(r chi.Router) {
		httprouter.AuthRouter(r, authService, cfg, deps.JWTManager)
	})

	return &AuthDomain{
		Service: authService,
	}, nil
}
