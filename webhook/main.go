package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"webhook/logging"
	"webhook/queue"
	redisClient "webhook/redis"
	redis_tls_config "webhook/utils"

	"github.com/go-redis/redis/v8"
)

func main() {
	// Create a context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logging.WebhookLogger(logging.EventType, "starting sendhooks engine")

	client, err := createRedisClient()
	if err != nil {
		log.Fatalf("Failed to create Redis client: %v", err)
	}

	// Create a channel to act as the queue
	webhookQueue := make(chan redisClient.WebhookPayload, 100) // Buffer size 100

	go queue.ProcessWebhooks(ctx, webhookQueue, client)

	// Subscribe to the "transactions" channel
	err = redisClient.SubscribeToStream(ctx, client, webhookQueue)

	if err != nil {
		logging.WebhookLogger(logging.ErrorType, fmt.Errorf("error initializing connection: %s", err))
		log.Fatalf("error initializing connection: %v", err)
		return
	}

	select {}
}

func createRedisClient() (*redis.Client, error) {
	redisAddress := os.Getenv("REDIS_ADDRESS")
	if redisAddress == "" {
		redisAddress = "localhost:6379" // Default address
	}

	redisDB := os.Getenv("REDIS_DB")
	if redisDB == "" {
		redisDB = "0" // Default database
	}

	redisDBInt, _ := strconv.Atoi(redisDB)

	redisPassword := os.Getenv("REDIS_PASSWORD")

	// SSL/TLS configuration
	useSSL := strings.ToLower(os.Getenv("REDIS_SSL")) == "true"
	var tlsConfig *tls.Config

	if useSSL {
		caCertPath := os.Getenv("REDIS_CA_CERT")
		clientCertPath := os.Getenv("REDIS_CLIENT_CERT")
		clientKeyPath := os.Getenv("REDIS_CLIENT_KEY")

		var err error
		tlsConfig, err = redis_tls_config.CreateTLSConfig(caCertPath, clientCertPath, clientKeyPath)
		if err != nil {
			return nil, err
		}
	}

	client := redis.NewClient(&redis.Options{
		Addr:      redisAddress,
		Password:  redisPassword,
		DB:        redisDBInt,
		TLSConfig: tlsConfig,
	})

	return client, nil
}
