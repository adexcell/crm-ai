package redis

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Host     string `env:"REDIS_HOST" envDefault:"localhost"`
	Port     int    `env:"REDIS_PORT" envDefault:"6379"`
	Password string `env:"REDIS_PASSWORD" envDefault:"password"`
}

func (c Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

type Client struct {
	*redis.Client
}

func New(ctx context.Context, cfg Config) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       0,
	})

	// Fail-fast: проверяем связь таймаутом в 3 секунды
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := rdb.Ping(pingCtx).Err(); err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("redis.Ping: %w", err)
	}

	slog.Info("Connected to Redis successfully", "addr", cfg.Addr())
	return &Client{Client: rdb}, nil
}

func (c *Client) Close() error {
	if c.Client != nil {
		slog.Info("Closing Redis connection...")
		return c.Client.Close()
	}
	return nil
}
