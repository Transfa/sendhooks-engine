package queue

import (
	"context"
	"fmt"
	"log"
	"time"
	"webhook/logging"
	"webhook/sender"

	redisClient "webhook/redis"
)

func ProcessWebhooks(ctx context.Context, webhookQueue chan redisClient.WebhookPayload) {
	for payload := range webhookQueue {
		go func(p redisClient.WebhookPayload) {
			backoffTime := time.Second  // starting backoff time
			maxBackoffTime := time.Hour // maximum backoff time
			retries := 0
			maxRetries := 5

			for {
				err := sender.SendWebhook(p.Data, p.Url, p.WebhookId)
				if err == nil {
					break
				}
				logging.WebhookLogger(logging.ErrorType, fmt.Errorf("error sending webhook: %s", err))

				retries++
				if retries >= maxRetries {
					logging.WebhookLogger(logging.WarningType, fmt.Errorf("max retries reached. Giving up on webhook: %s", p.WebhookId))
				}

				time.Sleep(backoffTime)

				// Double the backoff time for the next iteration, capped at the max
				backoffTime *= 2
				log.Println(backoffTime)
				if backoffTime > maxBackoffTime {
					backoffTime = maxBackoffTime
				}
			}
		}(payload)
	}
}
