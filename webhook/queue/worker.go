package queue

import (
	"context"
	"fmt"
	"log"
	"time"
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
		log.Println("Failed to send webhook after maximum retries. WebhookID:", payload.WebhookId)
	}
}

func retryWithExponentialBackoff(payload redisClient.WebhookPayload) error {
	retries := 0
	backoffTime := initialBackoff

	for retries < maxRetries {
		err := sender.SendWebhook(payload.Data, payload.Url, payload.WebhookId)

		fmt.Errorf("Error sending webhook:", err)

		backoffTime = calculateBackoff(backoffTime)
		retries++

		time.Sleep(backoffTime)
	}

	fmt.Errorf("Maximum retries reached:", maxRetries)

}

func calculateBackoff(currentBackoff time.Duration) time.Duration {

	nextBackoff := currentBackoff * 2

	if nextBackoff > maxBackoff {
		return maxBackoff
	}

	return nextBackoff
}
