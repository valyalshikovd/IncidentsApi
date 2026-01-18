package main

import (
	"RedColarTest/internal/incident/handlers"
	"RedColarTest/internal/incident/repository"
	"RedColarTest/internal/incident/services"
	locationHandlers "RedColarTest/internal/locations/handlers"
	locationRepo "RedColarTest/internal/locations/repository"
	locationServices "RedColarTest/internal/locations/services"
	"RedColarTest/internal/routes"
	systemHandlers "RedColarTest/internal/system/handlers"
	"RedColarTest/internal/webhook"
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found (ok if using real env vars):", err)
	}

	dsn := getEnv("DATABASE_URL", os.Getenv("database_url"))
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}

	operatorKey := getEnv("OPERATOR_API_KEY", os.Getenv("operator_api_key"))
	if operatorKey == "" {
		log.Fatal("OPERATOR_API_KEY is required")
	}

	statsWindowMinutes := getEnvInt("STATS_TIME_WINDOW_MINUTES", 60)
	cacheTTLSeconds := getEnvInt("CACHE_INCIDENTS_TTL_SECONDS", 60)
	webhookURL := getEnv("WEBHOOK_URL", "")
	webhookMaxRetries := getEnvInt("WEBHOOK_MAX_RETRIES", 5)
	webhookRetryBaseSeconds := getEnvInt("WEBHOOK_RETRY_BASE_SECONDS", 10)

	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	redisPassword := getEnv("REDIS_PASSWORD", "")
	redisDB := getEnvInt("REDIS_DB", 0)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	incRepo := repository.NewIncidentRepo(pool)
	incSvc := services.NewIncidentService(incRepo, redisClient)
	incHandler := handlers.NewIncidentHandler(incSvc)

	webhookQueue := webhook.NewQueue(redisClient, webhookURL, webhookMaxRetries, time.Duration(webhookRetryBaseSeconds)*time.Second)

	localRepo := locationRepo.NewLocationRepo(pool)
	localSvc := locationServices.NewLocationService(
		localRepo,
		incRepo,
		redisClient,
		time.Duration(cacheTTLSeconds)*time.Second,
		webhookQueue,
	)
	localHandler := locationHandlers.NewLocationHandler(localSvc, statsWindowMinutes)
	healthHandler := systemHandlers.NewHandler(pool, redisClient)

	r := routes.NewRouter(routes.RouterDeps{
		IncidentHandler: incHandler,
		LocationHandler: localHandler,
		HealthHandler:   healthHandler,
		OperatorKey:     operatorKey,
	})

	go webhookQueue.Run(context.Background())

	addr := ":8080"
	log.Println("listening on", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return def
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	return v
}
