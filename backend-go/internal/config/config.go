// Package config
package config

import (
	"fmt"

	httpserver "github.com/adexcell/crm-ai/internal/infra/http/server"
	"github.com/adexcell/crm-ai/internal/infra/jwt"
	"github.com/adexcell/crm-ai/internal/infra/postgres"
	"github.com/adexcell/crm-ai/internal/infra/redis"
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	HTTP     httpserver.Config
	Postgres postgres.Config
	Redis    redis.Config
	JWT      jwt.Config
	CORS     CORSConfig
}

type AppConfig struct {
	Name         string `env:"APP_NAME"      envDefault:"CRM AI Go Backend"`
	Env          string `env:"APP_ENV"       envDefault:"development"`
	Debug        bool   `env:"DEBUG"         envDefault:"true"`
	CookieSecure bool   `env:"COOKIE_SECURE" envDefault:"false"`
}

type CORSConfig struct {
	Origins []string `env:"CORS_ORIGINS" envSeparator:"," envDefault:"http://localhost:5173,http://localhost:3000,http://localhost:6006"`
}

func Load() (*Config, error) {
	_ = godotenv.Load("../.env", ".env")
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("env.Parse: %w", err)
	}
	return &cfg, nil
}
