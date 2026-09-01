package config

import (
	"os"
)

type Config struct {
	DatabaseURL       string
	StoragePath       string
	FFmpegPath        string
	FFprobePath       string
	WorkerID          string
	KafkaBrokers      string
	KafkaTopic        string
	KafkaConsumerGroup string
}

func Load() Config {
	workerID := os.Getenv("WORKER_ID")
	if workerID == "" {
		if host, err := os.Hostname(); err == nil && host != "" {
			workerID = host
		} else {
			workerID = "worker-1"
		}
	}

	return Config{
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://streamforge:streamforge@localhost:15433/streamforge?sslmode=disable"),
		StoragePath:        getEnv("STORAGE_PATH", "../../storage"),
		FFmpegPath:         getEnv("FFMPEG_PATH", "ffmpeg"),
		FFprobePath:        getEnv("FFPROBE_PATH", "ffprobe"),
		WorkerID:           workerID,
		KafkaBrokers:       getEnv("KAFKA_BROKERS", "localhost:29092"),
		KafkaTopic:         getEnv("KAFKA_TOPIC", "video.jobs"),
		KafkaConsumerGroup: getEnv("KAFKA_CONSUMER_GROUP", "streamforge-workers"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
