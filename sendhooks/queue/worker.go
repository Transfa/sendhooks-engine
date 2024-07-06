package queue

/*
* This package is used for queuing the webhooks to send using golang channels. In this package,
we also handle the retries and the exponential backoff logic for sending the hooks.
*/

import (
	"context"
	"fmt"
	"sendhooks/adapter"
	"sendhooks/adapter/adapter_manager"
	"sendhooks/logging"
	"sendhooks/sender"
	"time"

	redisadapter "sendhooks/adapter/redis_adapter"
)

const (
	maxRetries int = 5
)

const (
	initialBackoff time.Duration = time.Second
	maxBackoff     time.Duration = time.Hour
)

func ProcessWebhooks(ctx context.Context, webhookQueue chan adapter.WebhookPayload, configuration adapter.Configuration) {

	for payload := range webhookQueue {
		go sendWebhookWithRetries(ctx, payload, configuration)
	}
}

func sendWebhookWithRetries(ctx context.Context, payload adapter.WebhookPayload, configuration adapter.Configuration) {
	conf := adapter_manager.GetConfig()

	var queueAdapter adapter.Adapter

	if conf.Broker == "redis" {
		queueAdapter = redisadapter.NewRedisAdapter(conf)
	}
	if err, created := retryWithExponentialBackoff(ctx, payload, configuration); err != nil {
		logging.WebhookLogger(logging.WarningType, fmt.Errorf("failed to send sendhooks after maximum retries. WebhookID : %s", payload.WebhookID))
		err := queueAdapter.PublishStatus(ctx, payload.WebhookID, payload.URL, created, "", "failed", err.Error())
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

func retryWithExponentialBackoff(context context.Context, payload adapter.WebhookPayload, configuration adapter.Configuration) (error, string) {
	retries := 0
	backoffTime := initialBackoff
	var requestError error

	created := time.Now().String()

	for retries < maxRetries {
		err := sender.SendWebhook(payload.Data, payload.URL, payload.WebhookID, payload.SecretHash, configuration)

		if err == nil {
			if err != nil {
				logging.WebhookLogger(logging.WarningType, fmt.Errorf("error publishing status update WebhookID : %s ", payload.
					WebhookID))
			}
			// Break the loop if the request has been delivered successfully.
			requestError = nil
			break
		}

		logging.WebhookLogger(logging.ErrorType, fmt.Errorf("error sending sendhooks: %s", err))

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

	conf := adapter_manager.GetConfig()

	var queueAdapter adapter.Adapter

	if conf.Broker == "redis" {
		queueAdapter = redisadapter.NewRedisAdapter(conf)
	}

	err := queueAdapter.PublishStatus(context, payload.WebhookID, payload.URL, created, delivered, "success", "")
	if err != nil {
		logging.WebhookLogger(logging.WarningType, fmt.Errorf("error publishing status update: WebhookID : %s ", payload.WebhookID))
	}

	return nil, ""
}
