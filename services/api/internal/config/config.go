package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port         string
	DatabaseURL  string
	StoragePath  string
	MaxUploadMB  int64
	MigrationsPath string
}

func Load() Config {
	maxUploadMB := int64(500)
	if v := os.Getenv("MAX_UPLOAD_MB"); v != "" {
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			maxUploadMB = parsed
		}
	}

	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = "../../migrations"
	}

	return Config{
		Port:         getEnv("PORT", "8081"),
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://streamforge:streamforge@localhost:15433/streamforge?sslmode=disable"),
		StoragePath:  getEnv("STORAGE_PATH", "../../storage"),
		MaxUploadMB:  maxUploadMB,
		MigrationsPath: migrationsPath,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
