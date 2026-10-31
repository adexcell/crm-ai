package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/adexcell/crm-ai/internal/infra/transaction"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Host        string        `env:"POSTGRES_HOST" envDefault:"localhost"`
	Port        int           `env:"POSTGRES_PORT" envDefault:"5432"`
	User        string        `env:"POSTGRES_USER" envDefault:"admin"`
	Password    string        `env:"POSTGRES_PASSWORD" envDefault:"password"`
	DB          string        `env:"POSTGRES_DB" envDefault:"crm_storage"`
	PoolSize    int32         `env:"DATABASE_POOL_SIZE" envDefault:"10"`
	MaxOverflow int32         `env:"DATABASE_MAX_OVERFLOW" envDefault:"20"`
	MaxConnIdle time.Duration `env:"DATABASE_MAX_IDLE_TIME" envDefault:"5m"`
}

func (p Config) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		p.User, p.Password, p.Host, p.Port, p.DB,
	)
}

// DBTX - единый контракт выполнения SQL (транзакции и пул)
type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Pool struct {
	*pgxpool.Pool
}

func (p *Pool) Executor(ctx context.Context) DBTX {
	if tx := transaction.GetTx(ctx); tx != nil {
		return tx
	}
	return p.Pool
}

func (p *Pool) Close() {
	if p.Pool != nil {
		p.Pool.Close()
		slog.Info("Postgres closed")
	}
}

func New(ctx context.Context, cfg Config) (*Pool, error) {
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
