package redis

/*
This redis package provides utilities for subscribing and processing messages from a Redis Pub/Sub channel.

It defines the structure for webhook payloads, offers mechanisms for subscribing to a Redis channel,
and handles message processing.

- Graceful shutdowns and context propagation: Ensures the infinite message listening loop can be halted
  gracefully when required.

- Error handling and logging: Ensures that transient errors don't halt the entire message processing
  pipeline and provides insights into potential issues via logging.

- Configurability: Allows specifying the Redis channel name via environment variables such as REDIS_CHANNEL_NAME.
*/

import (
	"context"
	"fmt"
	"os"
	"webhook/logging"

	"github.com/go-redis/redis/v8"
)

// WebhookPayload This type represents the expected structure of the data from Redis. It contains
// the webhook URL, its ID, and the relevant data to be sent. There is no strict type on the data
// as it the structure can definitely vary.
type WebhookPayload struct {
	Url        string                 `json:"url"`
	WebhookId  string                 `json:"webhookId"`
	Data       map[string]interface{} `json:"data"`
	SecretHash string                 `json:"secretHash"`
}

type WebhookPayloadResponse struct {
	Url       string `json:"url"`
	Id        string `json:"id"`
	status    string `json:"status"`
	created   string `json:"created"`
	delivered string `json:"delivered"`
	error     string `json:"error"`
}

// Subscribe initializes a subscription to a Redis channel and continuously listens for messages.
// It decodes these messages into WebhookPayload and sends them to a provided channel.
func Subscribe(ctx context.Context, client *redis.Client, webhookQueue chan<- WebhookPayload, startedChan ...chan bool) error {
	streamName := getRedisStreamName()

	for {
		if len(startedChan) > 0 {
			startedChan[0] <- true
			// Clear the channel slice, so we don't send more signals. Needed and used for tests.
			startedChan = nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := processMessage(ctx, client, streamName, webhookQueue); err != nil {
				return err
			}
		}
	}
}

// getRedisStreamName fetches the Redis channel name from an environment variable.
// It defaults to "hooks" if not set.
func getRedisStreamName() string {
	channel := os.Getenv("REDIS_CHANNEL_NAME")
	if channel == "" {
		channel = "hooks"
	}
	return channel
}

func addMessageToStream(ctx context.Context, client *redis.Client, streamName string, message WebhookPayloadResponse) error {
	// Convert your message struct to a map[string]interface{} for XAdd.
	msgMap := map[string]interface{}{
		"url":       message.Url,
		"id":        message.Id,
		"status":    message.status,
		"created":   message.created,
		"delivered": message.delivered,
		"error":     message.error,
	}

	// The "*" ID tells Redis to auto-generate a unique ID for the message.
	_, err := client.XAdd(ctx, &redis.XAddArgs{
		Stream: streamName,
		Values: msgMap,
	}).Result()

	return err
}

func readMessagesFromStream(ctx context.Context, client *redis.Client, streamName string) ([]WebhookPayload, error) {
	entries, err := client.XRead(ctx, &redis.XReadArgs{
		Streams: []string{streamName, "0"}, // Read from the last ID.
		Count:   100,
		Block:   0,
	}).Result()

	if err != nil {
		return nil, err
	}

	var messages []WebhookPayload
	for _, entry := range entries[0].Messages {
		message := WebhookPayload{
			Url:        entry.Values["url"].(string),
			SecretHash: entry.Values["url"].(string),
			WebhookId:  entry.Values["webhookId"].(string),
			Data:       entry.Values["data"].(map[string]interface{}),
		}
		messages = append(messages, message)
	}

	return messages, nil
}

// processMessage retrieves, decodes, and dispatches a single message from the Redis channel.
// Separating message processing into its own function to enhance testability and maintainability.
func processMessage(ctx context.Context, client *redis.Client, streamName string, webhookQueue chan<- WebhookPayload) error {
	messages, err := readMessagesFromStream(ctx, client, streamName)

	if err != nil {
		return err
	}

	for index := 0; index < len(messages); index++ {
		payload := messages[index]

		// Non-blocking write to the webhookQueue. This ensures that if the queue is full, we
		// don't get stuck. Instead, we log the overflow and continue the execution.
		select {
		case webhookQueue <- payload:

		case <-ctx.Done():
			return ctx.Err()
		default:
			logging.WebhookLogger(logging.WarningType, fmt.Errorf("dropped webhook due to channel overflow. Webhook ID: %s", payload.WebhookId))

		}
	}

	return nil
}
