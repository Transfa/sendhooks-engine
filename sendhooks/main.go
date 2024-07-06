// main.go
package main

import (
	"context"
	"fmt"
	"log"

	"sendhooks/adapter"
	"sendhooks/adapter/adapter_manager"

	redisadapter "sendhooks/adapter/redis_adapter"
	"sendhooks/logging"
)

func main() {
	adapter_manager.LoadConfiguration("config.json")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := logging.WebhookLogger(logging.EventType, "starting sendhooks engine")
	if err != nil {
		log.Fatalf("Failed to log sendhooks event: %v", err)
	}

	var queueAdapter adapter.Adapter
	conf := adapter_manager.GetConfig()

	if conf.Broker == "redis" {
		queueAdapter = redisadapter.NewRedisAdapter(conf)
	}

	err = queueAdapter.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	webhookQueue := make(chan adapter.WebhookPayload, 100)

	go queueAdapter.ProcessWebhooks(ctx, webhookQueue)

	err = queueAdapter.SubscribeToQueue(ctx, webhookQueue)
	if err != nil {
		logging.WebhookLogger(logging.ErrorType, fmt.Errorf("error initializing connection: %s", err))
		log.Fatalf("error initializing connection: %v", err)
		return
	}

	select {}
}
