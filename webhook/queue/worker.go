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
		go sendWebhookWithRetries(ctx, payload, client, configuration)
	}
}

func sendWebhookWithRetries(ctx context.Context, payload redisClient.WebhookPayload, client *redis.Client, configuration redisClient.Configuration) {
	if err, created := retryWithExponentialBackoff(ctx, payload, client, configuration); err != nil {
		logging.WebhookLogger(logging.WarningType, fmt.Errorf("failed to send webhook after maximum retries. WebhookID : %s", payload.WebhookId))
		err := redisClient.PublishStatus(ctx, payload.WebhookID, payload.URL, created, "", "failed", err.Error(), client)
		if err != nil {
			logging.WebhookLogger(logging.WarningType, fmt.Errorf("error publishing status update: WebhookID : %s ", payload.WebhookID))
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

func retryWithExponentialBackoff(context context.Context,payload redisClient.WebhookPayload, client *redis.Client, configuration redisClient.Configuration) (error, string) {
	retries := 0
	backoffTime := initialBackoff
	var requestError error

	created := time.Now().String()

	for retries < maxRetries {
		err := sender.SendWebhook(payload.Data, payload.URL, payload.WebhookID, payload.SecretHash)

		if err == nil {
			err := redisClient.PublishStatus(context, payload.WebhookID, payload.URL, created, "", "success", "", client, configuration)
			if err != nil {
				logging.WebhookLogger(logging.WarningType, fmt.Errorf("error publishing status update WebhookID : %s ", payload.
					WebhookID))
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
		return requestError, created
	}

	delivered := time.Now().String()

	err := redisClient.PublishStatus(context, payload.WebhookID, payload.URL, created, delivered, "success", "", client)
	if err != nil {
		logging.WebhookLogger(logging.WarningType, fmt.Errorf("error publishing status update: WebhookID : %s ", payload.WebhookID))
	}

	return nil, ""
}
