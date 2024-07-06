package adapter_manager

import (
	"encoding/json"
	"log"
	"os"
	"sendhooks/adapter"
	"sync"

	redisadapter "sendhooks/adapter/redis_adapter"
)

var (
	instance adapter.Adapter
	once     sync.Once
)

var (
	config adapter.Configuration
)

// LoadConfiguration loads the configuration from a file
func LoadConfiguration(filename string) {
	once.Do(func() {
		file, err := os.Open(filename)
		if err != nil {
			log.Fatalf("Failed to open config file: %v", err)
		}
		defer file.Close()

		decoder := json.NewDecoder(file)
		err = decoder.Decode(&config)
		if err != nil {
			log.Fatalf("Failed to decode config file: %v", err)
		}
	})
}

// GetConfig returns the loaded configuration
func GetConfig() adapter.Configuration {
	return config
}

// Initialize initializes the appropriate adapter based on the configuration.
func Initialize() {
	once.Do(func() {
		conf := GetConfig()
		switch conf.Broker {
		case "redis":
			instance = redisadapter.NewRedisAdapter(conf)
		default:
			log.Fatalf("Unsupported broker type: %v", conf.Broker)
		}

		err := instance.Connect()
		if err != nil {
			log.Fatalf("Failed to connect to broker: %v", err)
		}
	})
}

// GetAdapter returns the singleton instance of the adapter.
func GetAdapter() adapter.Adapter {
	return instance
}
