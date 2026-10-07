package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/adexcell/crm-ai-pet/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Pool struct {
	*pgxpool.Pool
}

func New(ctx context.Context, cfg config.PostgresConfig) (*Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("pgxpool.ParseConfig: %w", err)
	}

	poolCfg.MaxConns = cfg.PoolSize + cfg.MaxOverflow
	poolCfg.MinConns = cfg.PoolSize
	poolCfg.MaxConnIdleTime = cfg.MaxConnIdle
	poolCfg.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.NewWithConfig: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pool.Ping(): %w", err)
	}

	_, err = pool.Exec(ctx, "CREATE EXTENSION IF NOT EXISTS vector;")
	if err != nil {
		slog.Warn("Failed to create extension vector (optional on init)", "error", err)
	}

	return &Pool{Pool: pool}, nil
}

func (p *Pool) Close() {
	if p.Pool != nil {
		p.Pool.Close()
	}
}
