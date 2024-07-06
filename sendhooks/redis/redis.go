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
	"sendhooks/adapter"
	"sendhooks/logging"

	"time"

	"github.com/go-redis/redis/v8"
)

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
func SubscribeToStream(ctx context.Context, client *redis.Client, webhookQueue chan<- adapter.WebhookPayload, config adapter.Configuration, startedChan ...chan bool) error {
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
func processStreamMessages(ctx context.Context, client *redis.Client, streamName string, webhookQueue chan<- adapter.WebhookPayload) error {
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
func readMessagesFromStream(ctx context.Context, client *redis.Client, streamName string) ([]adapter.WebhookPayload, error) {

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

	var messages []adapter.WebhookPayload
	for _, entry := range entries[0].Messages {
		var payload adapter.WebhookPayload

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
func getRedisStreamName(configuration adapter.Configuration) string {

	streamName := configuration.Redis.RedisStreamName
	if streamName == "" {
		streamName = "hooks"
	}
	return streamName
}

// getRedisStreamStatusName fetches the Redis stream name from an environment variable.
func getRedisStreamStatusName(configuration adapter.Configuration) string {

	streamStatusName := configuration.Redis.RedisStreamStatusName
	if streamStatusName == "" {
		streamStatusName = "sendhooks-status-updates"
	}
	return streamStatusName
}

// addMessageToStream adds a message to a Redis stream.
func addMessageToStream(ctx context.Context, client *redis.Client, streamName string, jsonString string) error {
	_, err := client.XAdd(ctx, &redis.XAddArgs{
		Stream: streamName,
		Values: map[string]interface{}{"data": jsonString},
	}).Result()

	return err
}

// toMap converts WebhookDeliveryStatus to a map for Redis XAdd.
// toJSONString converts WebhookDeliveryStatus to a JSON string for Redis XAdd.
func (wds WebhookDeliveryStatus) toJSONString() (string, error) {
	data, err := json.Marshal(wds)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// PublishStatus publishes sendhooks status updates to the Redis stream.
func PublishStatus(ctx context.Context, webhookID, url string, created string, delivered string, status, deliveryError string, client *redis.Client, config adapter.Configuration) error {
	message := WebhookDeliveryStatus{
		WebhookID:     webhookID,
		Status:        status,
		DeliveryError: deliveryError,
		URL:           url,
		Created:       created,
		Delivered:     delivered,
	}

	jsonString, err := message.toJSONString()
	if err != nil {
		return err
	}

	streamName := getRedisStreamStatusName(config)
	return addMessageToStream(ctx, client, streamName, jsonString)
}
