package redis_status

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"os"
	"webhook/logging"
)

type WebhookDeliveryStatus struct {
	WebhookID     string `json:"webhook_id"`
	Status        string `json:"status"`
	DeliveryError string `json:"delivery_error"`
	Url           string `json:"url"`
	Created       string `json:"created"`
	Delivered     string `json:"delivered"`
}

func addMessageToStream(ctx context.Context, client *redis.Client, streamName string, message WebhookDeliveryStatus) error {
	// Convert your message struct to a map[string]interface{} for XAdd.
	msgMap := map[string]interface{}{
		"url":       message.Url,
		"id":        message.WebhookID,
		"status":    message.Status,
		"created":   message.Created,
		"delivered": message.Delivered,
		"error":     message.DeliveryError,
	}

	// The "*" ID tells Redis to auto-generate a unique ID for the message.
	_, err := client.XAdd(ctx, &redis.XAddArgs{
		Stream: streamName,
		Values: msgMap,
	}).Result()

	return err
}

// Subscribe initializes a subscription to a Redis channel and continuously listens for messages.
// It decodes these messages into WebhookDeliveryStatus and sends them to a provided channel.
func Subscribe(ctx context.Context, client *redis.Client, startedChan ...chan bool) error {
	streamName := getRedisStreamName()

	pubSub := client.Subscribe(ctx, streamName)
	defer closePubSub(pubSub)

	for {
		if len(startedChan) > 0 {
			startedChan[0] <- true
			// Clear the channel slice, so we don't send more signals. Needed and used for tests.
			startedChan = nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg := <-pubSub.Channel():
			var status WebhookDeliveryStatus
			err := json.Unmarshal([]byte(msg.Payload), &status)
			if err != nil {
				logging.WebhookLogger(logging.ErrorType, fmt.Errorf("error decoding message: %s", err))
				continue
			}
			err = PublishStatus(ctx, status.WebhookID, status.Status, status.DeliveryError, client)
			if err != nil {
				logging.WebhookLogger(logging.ErrorType, fmt.Errorf("error publishing status: %s", err))
			}
		}
	}
}

// closePubSub is a utility function to close the PubSub connection.
// Separating this as a function to handle any errors during closure
// in a centralized manner.
func closePubSub(pubSub *redis.PubSub) {
	if err := pubSub.Close(); err != nil {
		logging.WebhookLogger(logging.ErrorType, fmt.Errorf("error closing PubSub: %w", err))
	}
}

// getRedisStreamName fetches the Redis channel name from an environment variable.
// It defaults to "hooks" if not set.
func getRedisStreamName() string {
	channel := os.Getenv("REDIS_STATUS_CHANNEL_NAME")
	if channel == "" {
		channel = "webhook-status-updates"
	}
	return channel
}

// PublishStatus This function publishes webhook status updates to the Redis channel.
func PublishStatus(ctx context.Context, webhookID, status, deliveryError string, client *redis.Client) error {
	message := WebhookDeliveryStatus{
		WebhookID:     webhookID,
		Status:        status,
		DeliveryError: deliveryError,
	}

	channelName := getRedisStreamName()
	return addMessageToStream(ctx, client, channelName, message)
}
