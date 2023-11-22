package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
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

// Declare a variable 'config' of type 'redisClient.Configuration'.
var Config redisClient.Configuration

// LoadConfiguration is a function that takes a filename as a string and returns an error.
// It's used to load and parse a configuration file for the Redis client.
func LoadConfiguration(filename string) error {

	// Open the file specified by 'filename'. If an error occurs (e.g., file not found),
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	// Defer the closing of the file until the end of the function's execution.
	// This ensures the file is closed once the function finishes, even if an error occurs.
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			// Currently, this error is ignored.
		}
	}(file)

	// Create a JSON decoder that reads from the opened file.
	decoder := json.NewDecoder(file)
	// Use the decoder to decode the JSON content into the 'config' variable.
	// If an error occurs during decoding (e.g., JSON format issue), return the error.
	err = decoder.Decode(&Config)
	if err != nil {
		return err
	}

	// If everything is successful, return nil indicating no error occurred.
	return nil
}

func main() {
	// Load configuration file
	err := LoadConfiguration("config.json")
	if err != nil {
		log.Printf("%v. anyway, using default configuration to the project\n", err)
	}
	// Create a context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = logging.WebhookLogger(logging.EventType, "starting sendhooks engine")
	if err != nil {
		log.Fatalf("Failed to log webhook event: %v", err)
	}

	client, err := createRedisClient()
	if err != nil {
		log.Fatalf("Failed to create Redis client: %v", err)
	}

	// Create a channel to act as the queue
	webhookQueue := make(chan redisClient.WebhookPayload, 100) // Buffer size 100

	go queue.ProcessWebhooks(ctx, webhookQueue, client, Config)

	// Subscribe to the "hooks" Redis stream
	err = redisClient.SubscribeToStream(ctx, client, webhookQueue, Config)

	if err != nil {
		logging.WebhookLogger(logging.ErrorType, fmt.Errorf("error initializing connection: %s", err))
		log.Fatalf("error initializing connection: %v", err)
		return
	}

	select {}
}

func createRedisClient() (*redis.Client, error) {
	redisAddress := Config.RedisAddress
	if redisAddress == "" {
		redisAddress = "localhost:6379" // Default address
	}

	redisDB := Config.RedisDb
	if redisDB == "" {
		redisDB = "0" // Default database
	}

	redisDBInt, _ := strconv.Atoi(redisDB)

	redisPassword := Config.RedisPassword

	// SSL/TLS configuration
	useSSL := strings.ToLower(Config.RedisSsl) == "true"
	var tlsConfig *tls.Config

	if useSSL {
		caCertPath := Config.RedisCaCert
		clientCertPath := Config.RedisClientCert
		clientKeyPath := Config.RedisClientKey

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
