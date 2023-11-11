package redis_status

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"webhook/logging"
	redisClientCustomized "webhook/redis"
)

type WebhookDeliveryStatus struct {
	WebhookID     string `json:"webhook_id"`
	Status        string `json:"status"`
	DeliveryError string `json:"delivery_error"`
}

// Subscribe initializes a subscription to a Redis channel and continuously listens for messages.
// It decodes these messages into WebhookDeliveryStatus and sends them to a provided channel.
func Subscribe(ctx context.Context, client *redis.Client, config redisClientCustomized.Configuration, startedChan ...chan bool) error {
	channelName := getRedisChannelName(config)

	pubSub := client.Subscribe(ctx, channelName)
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
			err = PublishStatus(status.WebhookID, status.Status, status.DeliveryError, client, config)
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

// getRedisChannelName fetches the Redis channel name from an environment variable.
// It defaults to "hooks" if not set.
func getRedisChannelName(configuration redisClientCustomized.Configuration) string {
	channel := configuration.RedisStatusChannelName
	if channel == "" {
		channel = "webhook-status-updates"
	}
	return channel
}

// PublishStatus This function publishes webhook status updates to the Redis channel.
func PublishStatus(webhookID, status, deliveryError string, client *redis.Client, config redisClientCustomized.Configuration) error {
	message := WebhookDeliveryStatus{
		WebhookID:     webhookID,
		Status:        status,
		DeliveryError: deliveryError,
	}

	messageJSON, err := json.Marshal(message)
	if err != nil {
		return err
	}

	channelName := getRedisChannelName(config)
	return client.Publish(context.Background(), channelName, messageJSON).Err()
}
