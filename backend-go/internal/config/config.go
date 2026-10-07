package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	Postgres PostgresConfig
	Redis    RedisConfig
	Kafka    KafkaConfig
	CORS     CORSConfig
}

type AppConfig struct {
	Name  string `env:"APP_NAME" envDefault:"CRM AI Go Backend"`
	Env   string `env:"APP_ENV" envDefault:"development"`
	Debug bool   `env:"DEBUG" envDefault:"true"`
	Port  int    `env:"PORT" envDefault:"8081"`
	Host  string `env:"HOST" envDefault:"0.0.0.0"`
}

type PostgresConfig struct {
	Host        string        `env:"POSTGRES_HOST" envDefault:"localhost"`
	Port        int           `env:"POSTGRES_PORT" envDefault:"5432"`
	User        string        `env:"POSTGRES_USER" envDefault:"crm_admin"`
	Password    string        `env:"POSTGRES_PASSWORD" envDefault:"crm_password"`
	DB          string        `env:"POSTGRES_DB" envDefault:"crm_db"`
	PoolSize    int32         `env:"DATABASE_POOL_SIZE" envDefault:"10"`
	MaxOverflow int32         `env:"DATABASE_MAX_OVERFLOW" envDefault:"20"`
	MaxConnIdle time.Duration `env:"DATABASE_MAX_IDLE_TIME" envDefault:"5m"`
}

func (p PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		p.User, p.Password, p.Host, p.Port, p.DB,
	)
}

type RedisConfig struct {
	Host     string `env:"REDIS_HOST" envDefault:"localhost"`
	Port     int    `env:"REDIS_PORT" envDefault:"6379"`
	Password string `env:"REDIS_PASSWORD" envDefault:"redis_password"`
}

func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type KafkaConfig struct {
	BootstrapServers string `env:"KAFKA_BOOTSTRAP_SERVERS" envDefault:"localhost:9094"`
}

type CORSConfig struct {
	Origins []string `env:"CORS_ORIGINS" envSeparator:"," envDefault:"http://localhost:5173,http://localhost:3000,http://localhost:6006"`
}

func Load() (*Config, error) {
	// Игнорируем ошибку, если файла .env нет (например, в Docker)
	_ = godotenv.Load(".env", "../infra/.env")

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("env.Parse: %w", err)
	}

	return &cfg, nil
}
