package queue

/*
* This package is used for queuing the webhooks to send using golang channels. In this package,
we also handle the retries and the exponential backoff logic for sending the hooks.
*/

import (
	"context"
	"fmt"
	"time"
	"webhook/logging"
	"webhook/redis_status"
	"webhook/sender"

	"github.com/go-redis/redis/v8"

	redisClient "webhook/redis"
)

const (
	maxRetries int = 5
)

const (
	initialBackoff time.Duration = time.Second
	maxBackoff     time.Duration = time.Hour
)

func ProcessWebhooks(ctx context.Context, webhookQueue chan redisClient.WebhookPayload, client *redis.Client, configuration redisClient.Configuration) {

	for payload := range webhookQueue {
		go sendWebhookWithRetries(payload, client, configuration)
	}
}

func sendWebhookWithRetries(payload redisClient.WebhookPayload, client *redis.Client, configuration redisClient.Configuration) {
	if err := retryWithExponentialBackoff(payload, client, configuration); err != nil {
		logging.WebhookLogger(logging.WarningType, fmt.Errorf("failed to send webhook after maximum retries. WebhookID : %s", payload.WebhookId))
		err := redis_status.PublishStatus(payload.WebhookId, "failed", err.Error(), client, configuration)
		if err != nil {
			logging.WebhookLogger(logging.WarningType, fmt.Errorf("error publishing status update: WebhookID : %s ", payload.WebhookId))
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

func retryWithExponentialBackoff(payload redisClient.WebhookPayload, client *redis.Client, configuration redisClient.Configuration) error {
	retries := 0
	backoffTime := initialBackoff
	var requestError error

	for retries < maxRetries {
		err := sender.SendWebhook(payload.Data, payload.Url, payload.WebhookId, payload.SecretHash)

		if err == nil {
			err := redis_status.PublishStatus(payload.WebhookId, "success", "", client, configuration)
			if err != nil {
				logging.WebhookLogger(logging.WarningType, fmt.Errorf("error publishing status update WebhookID : %s ", payload.WebhookId))
			}
			// Break the loop if the request has been delivered successfully.
			requestError = nil
			break
		}

		logging.WebhookLogger(logging.ErrorType, fmt.Errorf("error sending webhook: %s", err))

		backoffTime = calculateBackoff(backoffTime)
		retries++
		requestError = err
		time.Sleep(backoffTime)
	}

	logging.WebhookLogger(logging.WarningType, fmt.Errorf("maximum retries reached: %d", retries))

	if requestError != nil {
		return requestError
	}

	return nil
}
