package queue

/*
* This package is used for queuing the webhooks to send using golang channels. In this package,
we also handle the retries and the exponential backoff logic for sending the hooks.
*/

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"time"
	"webhook/logging"
	"webhook/redis_status"
	"webhook/sender"

	redisClient "webhook/redis"
)

const (
	maxRetries int = 5
)

const (
	initialBackoff time.Duration = time.Second
	maxBackoff     time.Duration = time.Hour
)

func ProcessWebhooks(ctx context.Context, webhookQueue chan redisClient.WebhookPayload, client *redis.Client) {

	for payload := range webhookQueue {
		go sendWebhookWithRetries(payload, client)
	}
}

func sendWebhookWithRetries(payload redisClient.WebhookPayload, client *redis.Client) {
	if err := retryWithExponentialBackoff(payload, client); err != nil {
		logging.WebhookLogger(logging.WarningType, fmt.Errorf("failed to send webhook after maximum retries. WebhookID : %s", payload.WebhookId))
		err := redis_status.PublishStatus(payload.WebhookId, "failed", err.Error(), client)
		if err != nil {
			logging.WebhookLogger(logging.WarningType, fmt.Errorf("Error publishing status update:", err))
		}
	}
}

func calculateBackoff(currentBackoff time.Duration) time.Duration {

	nextBackoff := currentBackoff * 2

	if nextBackoff > maxBackoff {
		return maxBackoff
	}

	return nextBackoff
}

func retryWithExponentialBackoff(payload redisClient.WebhookPayload, client *redis.Client) error {
	retries := 0
	backoffTime := initialBackoff

	for retries < maxRetries {
		err := sender.SendWebhook(payload.Data, payload.Url, payload.WebhookId, payload.SecretHash)

		if err == nil {
			err := redis_status.PublishStatus(payload.WebhookId, "success", "", client)
			if err != nil {
				logging.WebhookLogger(logging.WarningType, fmt.Errorf("Error publishing status update:", err))
			}
			// Break the loop if the request has been delivered successfully.
			break
		}

		logging.WebhookLogger(logging.ErrorType, fmt.Errorf("error sending webhook: %s", err))

		backoffTime = calculateBackoff(backoffTime)
		retries++

		time.Sleep(backoffTime)
	}

	logging.WebhookLogger(logging.WarningType, fmt.Errorf("maximum retries reached: %d", maxRetries))

	return nil
}
