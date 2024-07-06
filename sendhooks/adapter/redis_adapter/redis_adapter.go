package redisadapter

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"sendhooks/adapter"
	"sendhooks/logging"
	worker "sendhooks/queue"
	"sendhooks/utils"

	"github.com/go-redis/redis/v8"
)

// RedisAdapter implements the Adapter interface for Redis.
type RedisAdapter struct {
	client      *redis.Client
	config      adapter.Configuration
	queueName   string
	statusQueue string
	lastID      string
}

// NewRedisAdapter creates a new RedisAdapter instance.
func NewRedisAdapter(config adapter.Configuration) *RedisAdapter {
	return &RedisAdapter{
		config:      config,
		queueName:   config.Redis.RedisStreamName,
		statusQueue: config.Redis.RedisStreamStatusName,
		lastID:      "0",
	}
}

// Connect initializes the Redis client and establishes a connection.
func (r *RedisAdapter) Connect() error {
	redisAddress := r.config.Redis.RedisAddress
	if redisAddress == "" {
		redisAddress = "localhost:6379" // Default address
	}

	redisDB := r.config.Redis.RedisDb
	if redisDB == "" {
		redisDB = "0" // Default database
	}

	redisDBInt, _ := strconv.Atoi(redisDB)
	redisPassword := r.config.Redis.RedisPassword

	useSSL := strings.ToLower(r.config.Redis.RedisSsl) == "true"
	var tlsConfig *tls.Config

	if useSSL {
		caCertPath := r.config.Redis.RedisCaCert
		clientCertPath := r.config.Redis.RedisClientCert
		clientKeyPath := r.config.Redis.RedisClientKey

		var err error
		tlsConfig, err = utils.CreateTLSConfig(caCertPath, clientCertPath, clientKeyPath)
		if err != nil {
			return err
		}
	}

	r.client = redis.NewClient(&redis.Options{
		Addr:      redisAddress,
		Password:  redisPassword,
		DB:        redisDBInt,
		TLSConfig: tlsConfig,
		PoolSize:  50,
	})

	return nil
}

// SubscribeToQueue subscribes to the specified Redis queue and processes messages.
func (r *RedisAdapter) SubscribeToQueue(ctx context.Context, queue chan<- adapter.WebhookPayload) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := r.processQueueMessages(ctx, queue); err != nil {
				return err
			}
		}
	}
}

// processQueueMessages retrieves, decodes, and dispatches messages from the Redis queue.
func (r *RedisAdapter) processQueueMessages(ctx context.Context, queue chan<- adapter.WebhookPayload) error {
	messages, err := r.readMessagesFromQueue(ctx)
	if err != nil {
		return err
	}

	for _, payload := range messages {
		select {
		case queue <- payload:
			_, delErr := r.client.XDel(ctx, r.queueName, payload.MessageID).Result()
			if delErr != nil {
				logging.WebhookLogger(logging.ErrorType, fmt.Errorf("failed to delete message %s: %v", payload.MessageID, delErr))
			}
		case <-ctx.Done():
			return ctx.Err()
		default:
			logging.WebhookLogger(logging.WarningType, fmt.Errorf("dropped webhook due to channel overflow. Webhook ID: %s", payload.WebhookID))
		}
	}

	return nil
}

// readMessagesFromQueue reads messages from the Redis queue.
func (r *RedisAdapter) readMessagesFromQueue(ctx context.Context) ([]adapter.WebhookPayload, error) {
	entries, err := r.client.XRead(ctx, &redis.XReadArgs{
		Streams: []string{r.queueName, r.lastID},
		Count:   5,
	}).Result()

	if err != nil {
		if err == redis.Nil {
			time.Sleep(time.Second)
			return r.readMessagesFromQueue(ctx)
		}
		return nil, err
	}

	var messages []adapter.WebhookPayload
	for _, entry := range entries[0].Messages {
		var payload adapter.WebhookPayload

		if data, ok := entry.Values["data"].(string); ok {
			if err := json.Unmarshal([]byte(data), &payload); err != nil {
				logging.WebhookLogger(logging.ErrorType, fmt.Errorf("error unmarshalling message data: %w", err))
				r.lastID = entry.ID
				return nil, err
			}
		} else {
			logging.WebhookLogger(logging.ErrorType, fmt.Errorf("expected string for 'data' field but got %T", entry.Values["data"]))
			r.lastID = entry.ID
			return nil, err
		}

		payload.MessageID = entry.ID
		messages = append(messages, payload)
		r.lastID = entry.ID
	}

	return messages, nil
}

// ProcessWebhooks processes webhooks from the specified queue.
func (r *RedisAdapter) ProcessWebhooks(ctx context.Context, queue chan adapter.WebhookPayload, queueAdapter adapter.Adapter) {

	worker.ProcessWebhooks(ctx, queue, r.config, queueAdapter)
}

// PublishStatus publishes the status of a webhook delivery.
func (r *RedisAdapter) PublishStatus(ctx context.Context, webhookID, url, created, delivered, status, deliveryError string) error {
	message := adapter.WebhookDeliveryStatus{
		WebhookID:     webhookID,
		Status:        status,
		DeliveryError: deliveryError,
		URL:           url,
		Created:       created,
		Delivered:     delivered,
	}

	jsonString, err := json.Marshal(message)
	if err != nil {
		return err
	}

	_, err = r.client.XAdd(ctx, &redis.XAddArgs{
		Stream: r.statusQueue,
		Values: map[string]interface{}{"data": jsonString},
	}).Result()

	return err
}
