// main.go
package main

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"sync"

	"sendhooks/adapter"
	"sendhooks/adapter/adapter_manager"
	redisadapter "sendhooks/adapter/redis_adapter"
	"sendhooks/logging"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
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

	// Define the size of the channel and the number of workers
	webhookQueue := make(chan adapter.WebhookPayload, 1000)
	numWorkers := 50

	// Start the worker pool
	var wg sync.WaitGroup
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()
			queueAdapter.ProcessWebhooks(ctx, webhookQueue, queueAdapter)
		}()
	}

	err = queueAdapter.SubscribeToQueue(ctx, webhookQueue)
	if err != nil {
		logging.WebhookLogger(logging.ErrorType, fmt.Errorf("error initializing connection: %s", err))
		log.Fatalf("error initializing connection: %v", err)
		return
	}

	// Wait for all workers to finish
	wg.Wait()

	select {}
}
