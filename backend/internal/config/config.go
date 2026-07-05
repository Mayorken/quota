package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all runtime configuration, loaded from environment variables.
type Config struct {
	Port        string
	DatabaseURL string // Postgres DSN. If empty, falls back to a local SQLite file.
	SQLitePath  string
	JWTSecret      string
	CORSOrigin     string
	GoogleClientID string // Google OAuth client ID; empty disables Google sign-in.
}

// Load reads configuration from a .env file (if present) and the environment.
func Load() *Config {
	// .env is optional; ignore the error if it doesn't exist.
	_ = godotenv.Load()

	cfg := &Config{
		Port:        getenv("PORT", "8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		SQLitePath:  getenv("SQLITE_PATH", "quota.db"),
		JWTSecret:      getenv("JWT_SECRET", "dev-insecure-secret-change-me"),
		CORSOrigin:     getenv("CORS_ORIGIN", "http://localhost:5173"),
		GoogleClientID: os.Getenv("GOOGLE_CLIENT_ID"),
	}

	if cfg.JWTSecret == "dev-insecure-secret-change-me" {
		log.Println("WARNING: using the default insecure JWT secret. Set JWT_SECRET in production.")
	}

	return cfg
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
