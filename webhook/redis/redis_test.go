package redis

import (
	"context"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

// Mocking the Redis Client using interfaces would be ideal.
// For simplicity, this example won't include that part.

func TestGetRedisChannelName(t *testing.T) {
	assert := assert.New(t)

	// Test default value when environment variable is not set
	channel := getRedisChannelName()
	assert.Equal("hooks", channel, "Default channel name should be 'hooks'")
}

func TestProcessMessage(t *testing.T) {
	assert := assert.New(t)

	ctx := context.Background()

	// Mock the PubSub client and the message here.
	// This example doesn't include that, but you'd ideally want to mock the pubSub and the message being received.
	pubSub := &redis.PubSub{} // placeholder, you'd want a mock here

	webhookQueue := make(chan WebhookPayload, 1)
	defer close(webhookQueue)

	err := processMessage(ctx, pubSub, webhookQueue)
	assert.Nil(err, "Expected no error from processMessage")

	// You can add more tests to simulate different scenarios, like:
	// - Error during message reception from Redis.
	// - Error during JSON unmarshalling.
	// - Webhook queue being full.
}

// You can add more tests for the other methods and various scenarios.
