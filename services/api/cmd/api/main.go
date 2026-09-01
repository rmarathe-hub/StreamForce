package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/rmarathe-hub/StreamForce/services/api/internal/config"
	"github.com/rmarathe-hub/StreamForce/services/api/internal/handlers"
	"github.com/rmarathe-hub/StreamForce/services/api/internal/storage"
	"github.com/rmarathe-hub/StreamForce/shared/database"
	"github.com/rmarathe-hub/StreamForce/shared/kafka"
	"github.com/rmarathe-hub/StreamForce/shared/redis"
	"github.com/rmarathe-hub/StreamForce/shared/repository"
)

func main() {
	cfg := config.Load()

	ctx := context.Background()
	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer pool.Close()

	if err := database.RunMigrations(ctx, pool, cfg.MigrationsPath); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	brokers := kafka.ParseBrokers(cfg.KafkaBrokers)
	if err := kafka.EnsureTopic(ctx, brokers, cfg.KafkaTopic); err != nil {
		log.Fatalf("kafka topic setup failed: %v", err)
	}

	producer := kafka.NewProducer(brokers, cfg.KafkaTopic)
	defer producer.Close()

	publisher := kafka.NewJobPublisher(producer)

	redisClient, err := redis.NewClient(cfg.RedisURL)
	if err != nil {
		log.Fatalf("redis connection failed: %v", err)
	}
	defer redisClient.Close()

	progressStore := redis.NewProgressStore(redisClient)

	store, err := storage.NewLocalStorage(cfg.StoragePath)
	if err != nil {
		log.Fatalf("storage init failed: %v", err)
	}

	repo := repository.NewVideoRepository(pool)
	h := handlers.New(repo, store, publisher, progressStore, cfg)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(120 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", h.Health)
	r.Handle("/media/*", mediaHandler(cfg.StoragePath))
	r.Route("/api", func(r chi.Router) {
		r.Get("/videos", h.ListVideos)
		r.Post("/videos", h.CreateVideo)
		r.Get("/videos/{id}", h.GetVideo)
	})
	r.NotFound(h.NotFound)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("StreamForge API listening on http://localhost:%s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server shutdown failed: %v", err)
	}
}

func mediaHandler(storagePath string) http.Handler {
	absPath, err := filepath.Abs(storagePath)
	if err != nil {
		log.Fatalf("resolve storage path: %v", err)
	}

	fileServer := http.FileServer(http.Dir(absPath))
	return http.StripPrefix("/media/", fileServer)
}
