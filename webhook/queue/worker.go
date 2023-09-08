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

func ProcessWebhooks(ctx context.Context, webhookQueue chan redisClient.WebhookPayload) {
	for payload := range webhookQueue {
		go sendWebhookWithRetries(payload)
	}
}

func sendWebhookWithRetries(payload redisClient.WebhookPayload) {
	if err := retryWithExponentialBackoff(payload); err != nil {
		logging.WebhookLogger(logging.WarningType, fmt.Errorf("failed to send webhook after maximum retries. WebhookID : %s", payload.WebhookId))
	}
}

func calculateBackoff(currentBackoff time.Duration) time.Duration {

	nextBackoff := currentBackoff * 2

	if nextBackoff > maxBackoff {
		return maxBackoff
	}

	return nextBackoff
}

func retryWithExponentialBackoff(payload redisClient.WebhookPayload) error {
	retries := 0
	backoffTime := initialBackoff

	for retries < maxRetries {
		err := sender.SendWebhook(payload.Data, payload.Url, payload.WebhookId, payload.SecretHash)

		logging.WebhookLogger(logging.ErrorType, fmt.Errorf("error sending webhook: %s", err))

		backoffTime = calculateBackoff(backoffTime)
		retries++

		time.Sleep(backoffTime)
	}

	logging.WebhookLogger(logging.WarningType, fmt.Errorf("maximum retries reached: %s", maxRetries))

	return nil
}

