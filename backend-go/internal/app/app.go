// Package app
package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/adexcell/crm-ai/internal/config"
	"github.com/adexcell/crm-ai/internal/infra/http/middleware"
	"github.com/adexcell/crm-ai/internal/infra/http/router"
	httpserver "github.com/adexcell/crm-ai/internal/infra/http/server"
	"github.com/adexcell/crm-ai/internal/infra/jwt"
	"github.com/adexcell/crm-ai/internal/infra/metrics"
	"github.com/adexcell/crm-ai/internal/infra/postgres"
	"github.com/adexcell/crm-ai/internal/infra/redis"
	"github.com/adexcell/crm-ai/internal/infra/transaction"
	"github.com/go-chi/chi/v5"
)

// Dependencies содержит все внешние инфраструктурные компоненты и синглтоны приложения.
type Dependencies struct {
	// Adapters
	Postgres *postgres.Pool
	Redis    *redis.Client

	// Controllers & Cross-cutting
	RouterHTTP *chi.Mux
	Metrics    *metrics.Metrics
	MW         *middleware.Middleware
	TXManager  *transaction.Manager
	JWTManager *jwt.Manager
}

func Run(ctx context.Context, cfg *config.Config) error {
	var (
		err  error
		deps Dependencies
	)

	// --------- 1. Инициализация адаптеров баз данных ---------
	deps.Postgres, err = postgres.New(ctx, cfg.Postgres)
	if err != nil {
		return fmt.Errorf("postgres.New: %w", err)
	}
	defer deps.Postgres.Close()

	deps.Redis, err = redis.New(ctx, cfg.Redis)
	if err != nil {
		return fmt.Errorf("redis.New: %w", err)
	}
	defer func() {
		err := deps.Redis.Close()
		slog.DebugContext(ctx, "closing redis", "error", err)
	}()

	// --------- 2. Метрики, безопасность, транзакции ---------
	deps.Metrics = metrics.New()
	deps.JWTManager = jwt.New(cfg.JWT)
	deps.TXManager = transaction.New(deps.Postgres.Pool)
	deps.MW = middleware.New(deps.JWTManager, deps.Metrics)

	// --------- 3. Корневой HTTP-роутер ---------
	deps.RouterHTTP = router.New(deps.MW)

	// --------- 4. Инициализация и монтирование доменов ---------
	authDomain, err := NewAuthDomain(&deps, cfg)
	if err != nil {
		return fmt.Errorf("initializing auth domain: %w", err)
	}
	_ = authDomain

	// --------- 5. Запуск HTTP сервера ---------
	server := httpserver.New(deps.RouterHTTP, cfg.HTTP)
	defer server.Close()

	slog.Info("Application started successfully")

	// Ожидаем сигнал завершения (отмена переданного ctx из main.go)
	<-ctx.Done()

	slog.Info("Application received shutdown signal, draining connections...")

	// При выходе сработают все defer в порядке LIFO:
	// 1. server.Close()        — перестает принимать новые запросы и дожидается активных
	// 2. deps.Redis.Close()    — закрывает пул Redis
	// 3. deps.Postgres.Close() — закрывает пул соединений pgxpool
	return nil
}
