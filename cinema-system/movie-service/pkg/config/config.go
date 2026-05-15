package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	PostgresURL string
	RedisAddr   string
	NATSURL     string
	GRPCPort    string
}

func Load() (*Config, error) {
	_ = godotenv.Load()
	return &Config{
		PostgresURL: getenv("POSTGRES_URL", "postgres://user:pass@localhost:5433/movie_db?sslmode=disable"),
		RedisAddr:   getenv("REDIS_ADDR", "localhost:6379"),
		NATSURL:     getenv("NATS_URL", "nats://localhost:4222"),
		GRPCPort:    getenv("GRPC_PORT", "50051"),
	}, nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
