package queue

import (
	"context"
	"log"
	"time"
	"webhook/sender"

	redisClient "webhook/redis"
)

func ProcessWebhooks(ctx context.Context, webhookQueue chan redisClient.WebhookPayload) {
	for payload := range webhookQueue {
		go sendWebhookWithRetries(payload)
	}
}

func sendWebhookWithRetries(payload redisClient.WebhookPayload) {
	if err := retryWithExponentialBackoff(payload, 5, 1*time.Second, time.Hour); err != nil {
		log.Println("Failed to send webhook after maximum retries. WebhookID:", payload.WebhookId)
	}
}

func retryWithExponentialBackoff(payload redisClient.WebhookPayload, maxRetries int, initialBackoff, maxBackoff time.Duration) error {
	retries := 0
	backoffTime := initialBackoff

	for retries < maxRetries {
		err := sender.SendWebhook(payload.Data, payload.Url, payload.WebhookId)
		if err == nil {
			return nil
		}

		log.Println("Error sending webhook:", err)

		backoffTime = calculateBackoff(backoffTime, maxBackoff)
		retries++

		time.Sleep(backoffTime)
	}

	return log.Println("Maximum retries reached")
}

func calculateBackoff(currentBackoff, maxBackoff time.Duration) time.Duration {

	nextBackoff := currentBackoff * 2

	if nextBackoff > maxBackoff {
		return maxBackoff
	}

	return nextBackoff
}
