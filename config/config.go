package config

import (
"fmt"
"os"
)

// Config holds all application configuration, loaded from environment variables.
type Config struct {
// Server
Port string

// PostgreSQL
DBDSN string

// Redis
RedisAddr     string
RedisPassword string
RedisDB       int

// Auth
JWTSecret string

// App
Env string
}

// Load reads config from environment variables with sensible defaults.
func Load() (*Config, error) {
cfg := &Config{
Port:          getEnv("PORT", "8080"),
DBDSN:         getEnv("DATABASE_URL", "host=localhost user=postgres password=postgres dbname=usersdb port=5432 sslmode=disable"),
RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
RedisPassword: getEnv("REDIS_PASSWORD", ""),
JWTSecret:     getEnv("JWT_SECRET", ""),
Env:           getEnv("ENV", "development"),
}

if cfg.JWTSecret == "" && cfg.Env == "production" {
return nil, fmt.Errorf("JWT_SECRET must be set in production")
}

return cfg, nil
}

func getEnv(key, fallback string) string {
if v := os.Getenv(key); v != "" {
return v
}
return fallback
}
