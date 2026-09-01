package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/rmarathe-hub/StreamForce/services/worker/internal/config"
	"github.com/rmarathe-hub/StreamForce/shared/database"
	"github.com/rmarathe-hub/StreamForce/shared/kafka"
	"github.com/rmarathe-hub/StreamForce/shared/processor"
	"github.com/rmarathe-hub/StreamForce/shared/repository"
)

func main() {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer pool.Close()

	repo := repository.NewVideoRepository(pool)

	proc := processor.New(repo, processor.Config{
		StoragePath: cfg.StoragePath,
		FFmpegPath:  cfg.FFmpegPath,
		FFprobePath: cfg.FFprobePath,
	})

	brokers := kafka.ParseBrokers(cfg.KafkaBrokers)
	consumer := kafka.NewConsumer(kafka.Config{
		Brokers: brokers,
		Topic:   cfg.KafkaTopic,
		GroupID: cfg.KafkaConsumerGroup,
	})
	defer consumer.Close()

	runner := processor.NewKafkaRunner(consumer, repo, proc, cfg.WorkerID, cfg.KafkaConsumerGroup)
	if err := runner.Run(ctx); err != nil && err != context.Canceled {
		log.Fatalf("worker failed: %v", err)
	}

	os.Exit(0)
}
