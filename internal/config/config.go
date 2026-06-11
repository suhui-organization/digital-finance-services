package config

import "os"

// Config holds all configuration for the application.
type Config struct {
	DatabaseURL   string
	RedisURL      string
	JWTSecret     string
	AIServiceURL  string
	PublicBaseURL string
	Port          string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://localhost:5432/digital_finance?sslmode=disable"),
		RedisURL:      getEnv("REDIS_URL", "localhost:6379"),
		JWTSecret:     getEnv("JWT_SECRET", ""),
		AIServiceURL:  getEnv("AI_SERVICE_URL", "http://localhost:16081"),
		PublicBaseURL: getEnv("PUBLIC_BASE_URL", "http://localhost:16080"),
		Port:          getEnv("PORT", "16080"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
