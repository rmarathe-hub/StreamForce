package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	DatabaseURL  string
	StoragePath  string
	FFmpegPath   string
	FFprobePath  string
	WorkerID     string
	PollInterval time.Duration
}

func Load() Config {
	pollSeconds := 2
	if v := os.Getenv("WORKER_POLL_SECONDS"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			pollSeconds = parsed
		}
	}

	workerID := os.Getenv("WORKER_ID")
	if workerID == "" {
		if host, err := os.Hostname(); err == nil && host != "" {
			workerID = host
		} else {
			workerID = "worker-1"
		}
	}

	return Config{
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://streamforge:streamforge@localhost:15433/streamforge?sslmode=disable"),
		StoragePath:  getEnv("STORAGE_PATH", "../../storage"),
		FFmpegPath:   getEnv("FFMPEG_PATH", "ffmpeg"),
		FFprobePath:  getEnv("FFPROBE_PATH", "ffprobe"),
		WorkerID:     workerID,
		PollInterval: time.Duration(pollSeconds) * time.Second,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
