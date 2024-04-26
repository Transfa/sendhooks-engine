package redis

/*
This redis package provides utilities for subscribing and processing messages from a Redis stream.

It defines the structure for sendhooks payloads, offers mechanisms for subscribing to a Redis stream,
and handles message processing.

- Graceful shutdowns and context propagation: Ensures the infinite message listening loop can be halted
  gracefully when required.

- Error handling and logging: Ensures that transient errors don't halt the entire message processing
  pipeline and provides insights into potential issues via logging.

- Configurability: Allows specifying the Redis stream name via the config-example.json file with the `redisStreamName` key.
*/

import (
	"context"
	"encoding/json"
	"fmt"
	"sendhooks/logging"
	"time"

	"github.com/go-redis/redis/v8"
)

// WebhookPayload represents the structure of the data from Redis.
type WebhookPayload struct {
	URL        string                 `json:"url"`
	WebhookID  string                 `json:"webhookId"`
	MessageID  string                 `json:"messageId"`
	Data       map[string]interface{} `json:"data"`
	SecretHash string                 `json:"secretHash"`
	MetaData   map[string]interface{} `json:"metaData"`
}

// Configuration is a struct that holds various settings for a Redis client connection.
type Configuration struct {
	RedisAddress          string `json:"redisAddress"`
	RedisPassword         string `json:"redisPassword"`
	RedisDb               string `json:"redisDb"`
	RedisSsl              string `json:"redisSsl"`
	RedisCaCert           string `json:"redisCaCert"`
	RedisClientCert       string `json:"redisClientCert"`
	RedisClientKey        string `json:"redisClientKey"`
	RedisStreamName       string `json:"redisStreamName"`
	RedisStreamStatusName string `json:"redisStreamStatusName"`
	SecretHashHeaderName  string `json:"secretHashHeaderName"`
}

// WebhookDeliveryStatus represents the delivery status of a sendhooks.
type WebhookDeliveryStatus struct {
	WebhookID     string `json:"webhook_id"`
	Status        string `json:"status"`
	DeliveryError string `json:"delivery_error"`
	URL           string `json:"url"`
	Created       string `json:"created"`
	Delivered     string `json:"delivered"`
}

var lastID = "0" // Start reading from the beginning of the stream

// SubscribeToStream initializes a subscription to a Redis stream and continuously listens for messages.
func SubscribeToStream(ctx context.Context, client *redis.Client, webhookQueue chan<- WebhookPayload, config Configuration, startedChan ...chan bool) error {
	streamName := getRedisStreamName(config)

	for {
		if len(startedChan) > 0 {
			startedChan[0] <- true
			startedChan = nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := processStreamMessages(ctx, client, streamName, webhookQueue); err != nil {
				return err
			}
		}
	}
}

// processStreamMessages retrieves, decodes, and dispatches messages from the Redis stream.
func processStreamMessages(ctx context.Context, client *redis.Client, streamName string, webhookQueue chan<- WebhookPayload) error {
	messages, err := readMessagesFromStream(ctx, client, streamName)

	if err != nil {
		return err
	}

	for _, payload := range messages {
		select {
		case webhookQueue <- payload:
			_, delErr := client.XDel(ctx, streamName, payload.MessageID).Result()
			if delErr != nil {
				logging.WebhookLogger(logging.ErrorType, fmt.Errorf("failed to delete message %s: %v", payload.MessageID, delErr))
			}
		case <-ctx.Done():
			return ctx.Err()
		default:
			logging.WebhookLogger(logging.WarningType, fmt.Errorf("dropped sendhooks due to channel Golang overflow. Webhook ID: %s", payload.WebhookID))
		}
	}

	return nil
}

// readMessagesFromStream reads messages from a Redis stream and returns them.
func readMessagesFromStream(ctx context.Context, client *redis.Client, streamName string) ([]WebhookPayload, error) {

	entries, err := client.XRead(ctx, &redis.XReadArgs{
		Streams: []string{streamName, lastID},
		Count:   5,
	}).Result()

	if err != nil {
		if err == redis.Nil {
			// No new messages, sleep for a bit and try again
			time.Sleep(time.Second)

			return readMessagesFromStream(ctx, client, streamName)
		}
		return nil, err
	}

	var messages []WebhookPayload
	for _, entry := range entries[0].Messages {
		var payload WebhookPayload

		// Safely assert types and handle potential errors
		if data, ok := entry.Values["data"].(string); ok {
			if err := json.Unmarshal([]byte(data), &payload); err != nil {
				logging.WebhookLogger(logging.ErrorType, fmt.Errorf("error unmarshalling message data: %w", err))
				lastID = entry.ID

				return nil, err

			}
		} else {
			logging.WebhookLogger(logging.ErrorType, fmt.Errorf("error: expected string for 'data' field but got %T", entry.Values["data"]))
			lastID = entry.ID

			return nil, err

		}

		payload.MessageID = entry.ID

		messages = append(messages, payload)

		// Update lastID to the ID of the last message read
		lastID = entry.ID

		continue
	}

	return messages, nil
}

// getRedisStreamName fetches the Redis stream name from an environment variable.
func getRedisStreamName(configuration Configuration) string {

	streamName := configuration.RedisStreamName
	if streamName == "" {
		streamName = "hooks"
	}
	return streamName
}

// getRedisStreamStatusName fetches the Redis stream name from an environment variable.
func getRedisStreamStatusName(configuration Configuration) string {

	streamStatusName := configuration.RedisStreamStatusName
	if streamStatusName == "" {
		streamStatusName = "sendhooks-status-updates"
	}
	return streamStatusName
}

// addMessageToStream adds a message to a Redis stream.
func addMessageToStream(ctx context.Context, client *redis.Client, streamName string, message WebhookDeliveryStatus) error {
	msgMap, err := message.toMap()
	if err != nil {
		return fmt.Errorf("converting message to map: %w", err)
	}

	_, err = client.XAdd(ctx, &redis.XAddArgs{
		Stream: streamName,
		Values: msgMap,
	}).Result()

	return err
}

// toMap converts WebhookDeliveryStatus to a map for Redis XAdd.
func (wds WebhookDeliveryStatus) toMap() (map[string]interface{}, error) {
	data, err := json.Marshal(wds)
	if err != nil {
		return nil, err
	}

	var msgMap map[string]interface{}
	err = json.Unmarshal(data, &msgMap)
	if err != nil {
		return nil, err
	}

	return msgMap, nil
}

// PublishStatus publishes sendhooks status updates to the Redis stream.
func PublishStatus(ctx context.Context, webhookID, url string, created string, delivered string, status, deliveryError string, client *redis.Client, config Configuration) error {
	message := WebhookDeliveryStatus{
		WebhookID:     webhookID,
		Status:        status,
		DeliveryError: deliveryError,
		URL:           url,
		Created:       created,
		Delivered:     delivered,
	}

	streamName := getRedisStreamStatusName(config)
	return addMessageToStream(ctx, client, streamName, message)
}
