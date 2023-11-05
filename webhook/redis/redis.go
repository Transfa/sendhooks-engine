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
	"log"
	"os"
	"time"
	"webhook/logging"

	"github.com/go-redis/redis/v8"
)

// WebhookPayload represents the structure of the data from Redis.
type WebhookPayload struct {
	URL        string                 `json:"url"`
	WebhookID  string                 `json:"webhookId"`
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

// SubscribeToStream initializes a subscription to a Redis stream and continuously listens for messages.
func SubscribeToStream(ctx context.Context, client *redis.Client, groupName string, consumerName string, webhookQueue chan<- WebhookPayload, startedChan ...chan bool) error {
	streamName := getRedisSubStreamName()

	if err := createConsumerGroup(ctx, client, streamName, groupName, "$"); err != nil {
		return err
	}

	for {
		if len(startedChan) > 0 {
			startedChan[0] <- true
			startedChan = nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := processStreamMessages(ctx, client, streamName, groupName, consumerName, webhookQueue); err != nil {
				return err
			}
		}
	}
}

// processStreamMessages retrieves, decodes, and dispatches messages from the Redis stream.
func processStreamMessages(ctx context.Context, client *redis.Client, streamName string, groupName string, consumerName string, webhookQueue chan<- WebhookPayload) error {
	messages, lastID, err := readMessagesFromStream(ctx, client, streamName, groupName, consumerName)

	log.Println(messages)

	if err != nil {
		return err
	}

	for _, payload := range messages {
		select {
		case webhookQueue <- payload:
			if ackErr := acknowledgeMessage(ctx, client, streamName, groupName, lastID); ackErr != nil {
				logging.WebhookLogger(logging.ErrorType, fmt.Errorf("error acknowledging message: %w", ackErr))
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
func readMessagesFromStream(ctx context.Context, client *redis.Client, streamName, groupName, consumerName string) ([]WebhookPayload, string, error) {
	lastReadID := "0" // Start reading from the beginning of the stream

	entries, err := client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    groupName,
		Consumer: consumerName,
		Streams:  []string{streamName, lastReadID},
		Count:    5,
		Block:    0,
		NoAck:    false, // Set to true if you don't want to use XACK later
	}).Result()

	if err != nil {
		if err == redis.Nil {
			// No new messages, sleep for a bit and try again
			time.Sleep(time.Second)
			// If you want to continue trying, you can loop or use recursion
			// For a single batch read, you should return here
			return nil, "", nil
		}
		return nil, "", err
	}

	var messages []WebhookPayload
	for _, entry := range entries[0].Messages {
		var payload WebhookPayload

		// Safely assert types and handle potential errors
		if data, ok := entry.Values["data"].(string); ok {
			if err := json.Unmarshal([]byte(data), &payload); err != nil {
				logging.WebhookLogger(logging.ErrorType, fmt.Errorf("error unmarshalling message data: %w", err))
				return nil, "", err

			}
		} else {
			logging.WebhookLogger(logging.ErrorType, fmt.Errorf("error: expected string for 'data' field but got %T", entry.Values["data"]))
			return nil, "", err

		}

		messages = append(messages, payload)

		// Update lastID to the ID of the last message read
		lastReadID = entry.ID

		continue
	}

	return messages, lastReadID, nil
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

// acknowledgeMessage marks a message as processed in the given stream and consumer group.
func acknowledgeMessage(ctx context.Context, client *redis.Client, streamName, groupName, messageID string) error {
	// XACK command removes the message from the PEL of the consumer group.
	acknowledged, err := client.XAck(ctx, streamName, groupName, messageID).Result()
	if err != nil {
		return fmt.Errorf("error acknowledging message: %w", err)
	}

	if acknowledged == 0 {
		return fmt.Errorf("no message was acknowledged, possibly already acknowledged or does not exist")
	}

	fmt.Printf("Message %s acknowledged successfully in stream %s, group %s\n", messageID, streamName, groupName)
	return nil
}

// PublishStatus publishes webhook status updates to the Redis stream.
func PublishStatus(ctx context.Context, webhookID, status, deliveryError string, client *redis.Client) error {
	message := WebhookDeliveryStatus{
		WebhookID:     webhookID,
		Status:        status,
		DeliveryError: deliveryError,
	}

	streamName := getRedisPubStreamName()
	return addMessageToStream(ctx, client, streamName, message)
}

// createConsumerGroup creates a consumer group for a given stream.
func createConsumerGroup(ctx context.Context, client *redis.Client, streamName, groupName, startID string) error {
	err := client.XGroupCreateMkStream(ctx, streamName, groupName, startID).Err()
	if err != nil {
		if err.Error() != "BUSYGROUP Consumer Group name already exists" {
			return fmt.Errorf("error creating consumer group: %w", err)
		}
	}
	return nil
}
