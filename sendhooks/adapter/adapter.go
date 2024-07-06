package adapter

import (
	"context"
)

// WebhookPayload represents the structure of the data from Redis.
type WebhookPayload struct {
	URL        string                 `json:"url"`
	WebhookID  string                 `json:"webhookId"`
	MessageID  string                 `json:"messageId"`
	Data       map[string]interface{} `json:"data"`
	SecretHash string                 `json:"secretHash"`
	MetaData   map[string]interface{} `json:"metaData"`
}

type WebhookDeliveryStatus struct {
	WebhookID     string `json:"webhookId"`
	Status        string `json:"status"`
	DeliveryError string `json:"deliveryError"`
	URL           string `json:"url"`
	Created       string `json:"created"`
	PayloadSize   int    `json:"payloadSize"`
	NumberOfTries int    `json:"numberOfTries"`
	Delivered     string `json:"delivered"`
}

type RedisConfig struct {
	RedisAddress          string `json:"redisAddress"`
	RedisPassword         string `json:"redisPassword"`
	RedisDb               string `json:"redisDb"`
	RedisSsl              string `json:"redisSsl"`
	RedisCaCert           string `json:"redisCaCert"`
	RedisClientCert       string `json:"redisClientCert"`
	RedisClientKey        string `json:"redisClientKey"`
	RedisStreamName       string `json:"redisStreamName"`
	RedisStreamStatusName string `json:"redisStreamStatusName"`
}

type Configuration struct {
	Redis                RedisConfig `json:"redis"`
	SecretHashHeaderName string      `json:"secretHashHeaderName"`
	Broker               string      `json:"broker"`
}

// Adapter defines methods for interacting with different queue systems.
type Adapter interface {
	Connect() error
	SubscribeToQueue(ctx context.Context, queue chan<- WebhookPayload) error
	ProcessWebhooks(ctx context.Context, queue chan WebhookPayload, queueAdapter Adapter)
	PublishStatus(ctx context.Context, webhookID, url, created, delivered, status, deliveryError string, payloadSize int, numberOfTries int) error
}
