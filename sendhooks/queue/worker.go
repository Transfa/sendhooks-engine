package queue

/*
* This package is used for queuing the webhooks to send using golang channels. In this package,
we also handle the retries and the exponential backoff logic for sending the hooks.
*/

import (
	"context"
	"fmt"
	"sendhooks/adapter"
	"sendhooks/logging"
	"sendhooks/sender"
	"time"
	"unsafe"
)

// Function to measure the size of a map in bytes
func SizeofMap(m map[string]interface{}) int {
	size := int(unsafe.Sizeof(m))
	for k, v := range m {
		size += int(unsafe.Sizeof(k)) + len(k)
		size += SizeofValue(v)
	}
	return size
}

// Function to measure the size of a value in bytes
func SizeofValue(v interface{}) int {
	switch v := v.(type) {
	case int:
		return int(unsafe.Sizeof(v))
	case float64:
		return int(unsafe.Sizeof(v))
	case string:
		return int(unsafe.Sizeof(v)) + len(v)
	case []int:
		return int(unsafe.Sizeof(v)) + len(v)*int(unsafe.Sizeof(v[0]))
	// Add more cases as needed for other types
	default:
		return 0 // Unsupported type
	}
}

const (
	maxRetries int = 5
)

const (
	initialBackoff time.Duration = time.Second
	maxBackoff     time.Duration = time.Hour
)

func ProcessWebhooks(ctx context.Context, webhookQueue chan adapter.WebhookPayload, configuration adapter.Configuration, queueAdapter adapter.Adapter) {

	for payload := range webhookQueue {
		go sendWebhookWithRetries(ctx, payload, configuration, queueAdapter)
	}
}

func sendWebhookWithRetries(ctx context.Context, payload adapter.WebhookPayload, configuration adapter.Configuration, queueAdapter adapter.Adapter) {

	if err, created, retries := retryWithExponentialBackoff(ctx, payload, configuration, queueAdapter); err != nil {
		logging.WebhookLogger(logging.WarningType, fmt.Errorf("failed to send sendhooks after maximum retries. WebhookID : %s", payload.WebhookID))
		err := queueAdapter.PublishStatus(ctx, payload.WebhookID, payload.URL, created, "", "failed", err.Error(), SizeofMap(payload.Data), retries)
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

func retryWithExponentialBackoff(context context.Context, payload adapter.WebhookPayload, configuration adapter.Configuration, queueAdapter adapter.Adapter) (error, string, int) {
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
		return requestError, created, retries
	}

	delivered := time.Now().String()

	err := queueAdapter.PublishStatus(context, payload.WebhookID, payload.URL, created, delivered, "success", "", SizeofMap(payload.Data), retries)
	if err != nil {
		logging.WebhookLogger(logging.WarningType, fmt.Errorf("error publishing status update: WebhookID : %s ", payload.WebhookID))
	}

	return nil, "", retries
}
