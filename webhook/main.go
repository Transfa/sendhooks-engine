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
	redisWebhookStatusClient "webhook/redis_status"
	redis_tls_config "webhook/utils"

	"github.com/go-redis/redis/v8"
)

var config Configuration

type Configuration struct {
	RedisAddress           string `json:"redis_address"`
	RedisPassword          string `json:"redis_password"`
	RedisDb                string `json:"redis_db"`
	RedisSsl               string `json:"redis_ssl"`
	RedisCaCert            string `json:"redis_ca_cert"`
	RedisClientCert        string `json:"redis_client_cert"`
	RedisClientKey         string `json:"redis_client_key"`
	RedisChannelName       string `json:"redis_channel_name"`
	RedisStatusChannelName string `json:"redis_status_channel_name"`
}

func LoadConfiguration(filename string) error {

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	// Load configuration file
	err := LoadConfiguration("config.json")
	if err != nil {
		log.Fatal(err)
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

	go queue.ProcessWebhooks(ctx, webhookQueue, client)

	// Subscribe to the "transactions" channel
	err = redisClient.Subscribe(ctx, client, webhookQueue)

	// Subscribe to the "webhook status updates" channel
	err = redisWebhookStatusClient.Subscribe(ctx, client)

	if err != nil {
		logging.WebhookLogger(logging.ErrorType, fmt.Errorf("error initializing connection: %s", err))
		log.Fatalf("error initializing connection: %v", err)
		return
	}

	select {}
}

func createRedisClient() (*redis.Client, error) {
	redisAddress := config.RedisAddress
	if redisAddress == "" {
		redisAddress = "localhost:6379" // Default address
	}

	redisDB := config.RedisDb
	if redisDB == "" {
		redisDB = "0" // Default database
	}

	redisDBInt, _ := strconv.Atoi(redisDB)

	redisPassword := config.RedisPassword

	// SSL/TLS configuration
	useSSL := strings.ToLower(config.RedisSsl) == "true"
	var tlsConfig *tls.Config

	if useSSL {
		caCertPath := config.RedisCaCert
		clientCertPath := config.RedisClientCert
		clientKeyPath := config.RedisClientKey

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
