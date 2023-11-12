package redis

/*
This redis package provides utilities for subscribing and processing messages from a Redis stream.

It defines the structure for webhook payloads, offers mechanisms for subscribing to a Redis stream,
and handles message processing.

- Graceful shutdowns and context propagation: Ensures the infinite message listening loop can be halted
  gracefully when required.

- Error handling and logging: Ensures that transient errors don't halt the entire message processing
  pipeline and provides insights into potential issues via logging.

- Configurability: Allows specifying the Redis channel name via environment variables such as REDIS_STREAM_NAME.
*/

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"
	"webhook/logging"

	"github.com/go-redis/redis/v8"
)

// WebhookPayload represents the structure of the data from Redis.
type WebhookPayload struct {
	URL        string                 `json:"url"`
	WebhookID  string                 `json:"webhookId"`
	MessageID  string                 `json:"messageId"`
	Data       map[string]interface{} `json:"data"`
	SecretHash string                 `json:"secretHash"`
}

// WebhookDeliveryStatus represents the delivery status of a webhook.
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
func SubscribeToStream(ctx context.Context, client *redis.Client, webhookQueue chan<- WebhookPayload, startedChan ...chan bool) error {
	streamName := getRedisSubStreamName()

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
			logging.WebhookLogger(logging.WarningType, fmt.Errorf("dropped webhook due to channel overflow. Webhook ID: %s", payload.WebhookID))
		}
	}

	return nil
}

// readMessagesFromStream reads messages from a Redis stream and returns them.
func readMessagesFromStream(ctx context.Context, client *redis.Client, streamName string) ([]WebhookPayload, error) {

	entries, err := client.XRead(ctx, &redis.XReadArgs{
		Streams: []string{streamName, lastID},
		Count:   5,
		Block:   0,
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

// getRedisSubStreamName fetches the Redis stream name from an environment variable.
func getRedisSubStreamName() string {

	channel := os.Getenv("REDIS_STREAM_NAME")
	if channel == "" {
		channel = "hooks"
	}
	return channel
}

// getRedisPubStreamName fetches the Redis stream name from an environment variable.
func getRedisPubStreamName() string {

	channel := os.Getenv("REDIS_STATUS_CHANNEL_NAME")
	if channel == "" {
		channel = "webhook-status-updates"
	}
	return channel
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

// PublishStatus publishes webhook status updates to the Redis stream.
func PublishStatus(ctx context.Context, webhookID, url string, created string, delivered string, status, deliveryError string, client *redis.Client) error {
	message := WebhookDeliveryStatus{
		WebhookID:     webhookID,
		Status:        status,
		DeliveryError: deliveryError,
		URL:           url,
		Created:       created,
		Delivered:     delivered,
	}

	streamName := getRedisPubStreamName()
	return addMessageToStream(ctx, client, streamName, message)
}
