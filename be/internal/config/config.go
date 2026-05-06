package config

import (
	"context"
	"fmt"
	"time"

	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
)

type Config struct {
	Server   ServerConfig
	DB       DBConfig
	JWT      JWTConfig
	CORS     CORSConfig
}

type ServerConfig struct {
	Port         int           `env:"PORT,default=8080"`
	ReadTimeout  time.Duration `env:"SERVER_READ_TIMEOUT,default=10s"`
	WriteTimeout time.Duration `env:"SERVER_WRITE_TIMEOUT,default=10s"`
	IdleTimeout  time.Duration `env:"SERVER_IDLE_TIMEOUT,default=120s"`
}

type DBConfig struct {
	Host     string `env:"DB_HOST,default=localhost"`
	Port     int    `env:"DB_PORT,default=5432"`
	User     string `env:"DB_USER,required"`
	Password string `env:"DB_PASSWORD,required"`
	Name     string `env:"DB_NAME,required"`
	SSLMode  string `env:"DB_SSLMODE,default=disable"`
	MaxConns int    `env:"DB_MAX_CONNS,default=25"`
}

func (d DBConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode,
	)
}

type JWTConfig struct {
	Secret            string        `env:"JWT_SECRET,required"`
	AccessExpiration  time.Duration `env:"JWT_ACCESS_EXPIRATION,default=15m"`
	RefreshExpiration time.Duration `env:"JWT_REFRESH_EXPIRATION,default=168h"`
}

type CORSConfig struct {
	AllowedOrigin string `env:"CORS_ORIGIN,default=http://localhost:3000"`
}

func Load(ctx context.Context) (*Config, error) {
	_ = godotenv.Load()

	var cfg Config
	if err := envconfig.Process(ctx, &cfg); err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}
	return &cfg, nil
}
