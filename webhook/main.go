package main

import (
	"context"
	"fmt"
	"os"

	"webhook/logging"
	redisClient "webhook/redis"

	"webhook/queue"

	"github.com/go-redis/redis/v8" // Make sure to use the correct version
)

func main() {
	// Create a context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	redisAddress := os.Getenv("REDIS_ADDRESS")
	if redisAddress == "" {
		redisAddress = "localhost:6379" // Default address
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")
	if redisPassword == "" {
		redisPassword = "" // Default password (empty in this case)
	}

	client := redis.NewClient(&redis.Options{
		Addr:     redisAddress,
		Password: redisPassword,
		DB:       0, // Default DB
	})
	// Create a channel to act as the queue
	webhookQueue := make(chan redisClient.WebhookPayload, 100) // Buffer size 100

	go queue.ProcessWebhooks(ctx, webhookQueue)

	// Subscribe to the "transactions" channel
	err := redisClient.Subscribe(ctx, client, webhookQueue)

	if err != nil {
		logging.WebhookLogger(logging.ErrorType, fmt.Errorf("error initializing connection: %s", err))
	}

	select {}

}
