package redisadapter

import (
	"context"
	"crypto/tls"
	"strconv"
	"strings"

	"sendhooks/adapter"
	"sendhooks/utils"

	"github.com/go-redis/redis/v8"
)

type RedisAdapter struct {
	client       *redis.Client
	config       adapter.Configuration
	streamName   string
	statusStream string
}

// PublishStatus implements adapter.Adapter.
func (r *RedisAdapter) PublishStatus(ctx context.Context, webhookID string, url string, created string, delivered string, status string, deliveryError string) error {
	panic("unimplemented")
}

// SubscribeToQueue implements adapter.Adapter.
func (r *RedisAdapter) SubscribeToQueue(ctx context.Context, queue chan<- adapter.WebhookPayload) error {
	panic("unimplemented")
}

func NewRedisAdapter(config adapter.Configuration) *RedisAdapter {
	return &RedisAdapter{
		config:       config,
		streamName:   config.Redis.RedisStreamName,
		statusStream: config.Redis.RedisStreamStatusName,
	}
}

func (r *RedisAdapter) Connect() error {
	redisAddress := r.config.Redis.RedisAddress
	if redisAddress == "" {
		redisAddress = "localhost:6379" // Default address
	}

	redisDB := r.config.Redis.RedisDb
	if redisDB == "" {
		redisDB = "0" // Default database
	}

	redisDBInt, _ := strconv.Atoi(redisDB)
	redisPassword := r.config.Redis.RedisPassword

	useSSL := strings.ToLower(r.config.Redis.RedisSsl) == "true"
	var tlsConfig *tls.Config

	if useSSL {
		caCertPath := r.config.Redis.RedisCaCert
		clientCertPath := r.config.Redis.RedisClientCert
		clientKeyPath := r.config.Redis.RedisClientKey

		var err error
		tlsConfig, err = utils.CreateTLSConfig(caCertPath, clientCertPath, clientKeyPath)
		if err != nil {
			return err
		}
	}

	r.client = redis.NewClient(&redis.Options{
		Addr:      redisAddress,
		Password:  redisPassword,
		DB:        redisDBInt,
		TLSConfig: tlsConfig,
	})

	return nil
}

func (r *RedisAdapter) SubscribeToStream(ctx context.Context, streamName string, queue chan<- adapter.WebhookPayload) error {
	stream := streamName
	if stream == "" {
		stream = r.streamName
	}
	// Implement subscription logic here
	// Example:
	// for {
	//     streams, err := r.client.XRead(&redis.XReadArgs{
	//         Streams: []string{stream, "0"},
	//         Count:   1,
	//         Block:   0,
	//     }).Result()
	//     if err != nil {
	//         return err
	//     }
	//     for _, message := range streams[0].Messages {
	//         queue <- adapter.WebhookPayload{ID: message.ID, Payload: message.Values["data"].(string)}
	//     }
	// }
	return nil
}

func (r *RedisAdapter) ProcessWebhooks(ctx context.Context, queue <-chan adapter.WebhookPayload) error {
	// Implement webhook processing logic here
	return nil
}
