package redis

/*
TestSubscribe is a unit test for the Subscribe function within the redis package.

Purpose:
- To ensure that the Subscribe function correctly initiates a subscription to the desired Redis channel ("hooks" in this case).
- To verify that the function enters its main loop and starts listening for messages from the channel.

Details:
1. A mock Redis client is used to simulate interactions with an actual Redis server.
2. The test expects the Subscribe function to call the Subscribe method of the Redis client with the channel name "hooks".
3. The Subscribe function runs in a separate goroutine to emulate its continuous listening nature in a real-world scenario.
4. The function is expected to signal that it has started its main loop within a certain timeout (2 seconds in this test). Failure to do so results in the test being marked as failed.

By performing this test, we aim to catch potential initialization and listening issues in the Subscribe function before they manifest in production or other stages of the development lifecycle.
*/

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	gomock "go.uber.org/mock/gomock"
)

func TestSubscribe(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRedisClient := NewMockRedisClient(ctrl)

	// Setup your expectations
	// For example, if you want to mock that Subscribe is called with the "hooks" channel and return a mock PubSub:
	mockPubSub := &redis.PubSub{} // You would ideally want to mock this too
	mockRedisClient.EXPECT().Subscribe(gomock.Any(), "hooks").Return(mockPubSub)

	ctx := context.Background()
	webhookQueue := make(chan WebhookPayload, 1)
	defer close(webhookQueue)

	loopStarted := make(chan bool)

	go func() {
		// Call your function in a goroutine to not block the test
		_ = Subscribe(ctx, mockRedisClient, webhookQueue, loopStarted)
	}()

	select {
	case <-loopStarted:
		// Loop has started
	case <-time.After(2 * time.Second):
		t.Fatal("Expected the Subscribe function to enter its loop within 2 seconds")
	}

}
