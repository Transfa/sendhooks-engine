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
	"encoding/json"
	"fmt"
	"os"
	"webhook/logging"

	"github.com/go-redis/redis/v8"
)

// This type will be useful in the context of mocking for requests in tests.
type RedisClient interface {
	Subscribe(ctx context.Context, channel string) *redis.PubSub
}

// This type represents the expected structure of the data from Redis. It contains
// the webhook URL, its ID, and the relevant data to be sent. There is no strict type on the data
// as it the structure can definitely vary.
type WebhookPayload struct {
	Url        string                 `json:"url"`
	WebhookId  string                 `json:"webhookId"`
	Data       map[string]interface{} `json:"data"`
	SecretHash string                 `json:"secretHash"`
}

// Subscribe initializes a subscription to a Redis channel and continuously listens for messages.
// It decodes these messages into WebhookPayload and sends them to a provided channel.
func Subscribe(ctx context.Context, client *redis.Client, webhookQueue chan<- WebhookPayload, startedChan ...chan bool) error {
	channelName := getRedisChannelName()

	pubSub := client.Subscribe(ctx, channelName)
	defer closePubSub(pubSub)

	for {
		if len(startedChan) > 0 {

			startedChan[0] <- true
			// Clear the channel slice so we don't send more signals. Needed and used for tests.
			startedChan = nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:

			if err := processMessage(ctx, pubSub, webhookQueue); err != nil {
				return err
			}
		}
	}
}

// getRedisChannelName fetches the Redis channel name from an environment variable.
// It defaults to "hooks" if not set.
func getRedisChannelName() string {
	channel := os.Getenv("REDIS_CHANNEL_NAME")
	if channel == "" {
		channel = "hooks"
	}
	return channel
}

// closePubSub is a utility function to close the PubSub connection.
// Separating this as a function to handle any errors during closure
// in a centralized manner.
func closePubSub(pubSub *redis.PubSub) {
	if err := pubSub.Close(); err != nil {
		logging.WebhookLogger(logging.ErrorType, fmt.Errorf("Error closing PubSub: %w", err))
	}
}

// processMessage retrieves, decodes, and dispatches a single message from the Redis channel.
// Separating message processing into its own function to enhance testability and maintainability.
func processMessage(ctx context.Context, pubSub *redis.PubSub, webhookQueue chan<- WebhookPayload) error {

	msg, err := pubSub.ReceiveMessage(ctx)

	if err != nil {
		return err
	}

	var payload WebhookPayload
	if err = json.Unmarshal([]byte(msg.Payload), &payload); err != nil {

		logging.WebhookLogger(logging.ErrorType, fmt.Errorf("error unmarshalling payload: %s", err))

		return nil
	}

	// Non-blocking write to the webhookQueue. This ensures that if the queue is full, we
	// don't get stuck. Instead, we log the overflow and continue the execution.
	select {
	case webhookQueue <- payload:

	case <-ctx.Done():
		return ctx.Err()
	default:
		logging.WebhookLogger(logging.WarningType, fmt.Errorf("dropped webhook due to channel overflow. Webhook ID: %s", payload.WebhookId))

	}

	return nil
}
