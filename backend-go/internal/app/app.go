package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"

	"github.com/adexcell/crm-ai-pet/internal/config"
	"github.com/adexcell/crm-ai-pet/internal/database"
	"github.com/adexcell/crm-ai-pet/internal/middleware"
	"github.com/adexcell/crm-ai-pet/internal/modules/health"
)

func Run(ctx context.Context, cfg *config.Config) error {
	dbPool, err := database.New(ctx, cfg.Postgres)
	if err != nil {
		return fmt.Errorf("database.New(): %w", err)
	}
	defer dbPool.Close()
	slog.Info("Connected to PostgreSQL successfully")

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password,
	})
	defer redisClient.Close()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		slog.Warn("Redis ping failed on startup (check container)", "error", err)
	} else {
		slog.Info("Connected to Redis successfully")
	}

	r := chi.NewRouter()
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.CorrelationID)

	healthHandler := health.NewHandler(dbPool, redisClient)
	r.Route("/api/v1", func(r chi.Router) {
		healthHandler.RegisterRoutes(r)
	})

	serverAddr := fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.Port)
	server := &http.Server{
		Addr:         serverAddr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("HTTP server listening", "addr", serverAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- fmt.Errorf("server.ListenAndServe: %w", err)
		}
	}()

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		slog.Info("Shutdown signal recieved, draining connections...")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server.Shutdown: %w", err)
	}

	slog.Info("Server stopped cleanly. Goodbye")
	return nil
}
