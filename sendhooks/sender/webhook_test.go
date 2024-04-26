package sender

import (
	"errors"
	"net/http"
	"sendhooks/logging"
	"sendhooks/redis"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	marshalJSONOrig      = marshalJSON
	prepareRequestOrig   = prepareRequest
	sendRequestOrig      = sendRequest
	processResponseOrig  = processResponse
	webhookLoggerInvoked = false
)

func TestSendWebhook(t *testing.T) {
	logging.WebhookLogger = func(errorType string, errorMessage interface{}) error {
		webhookLoggerInvoked = true
		return nil
	}

	t.Run("Successful sendhooks sending", func(t *testing.T) {
		resetMocks() // Reset all mocks to original functions

		err := SendWebhook(nil, "http://dummy.com", "webhookId", "secretHash", redis.Configuration{})

		assert.NoError(t, err)
	})

	t.Run("Failed sendhooks due to marshaling errors", func(t *testing.T) {
		resetMocks()
		marshalJSON = func(data interface{}) ([]byte, error) {
			return nil, errors.New("marshaling error")
		}

		err := SendWebhook(nil, "http://dummy.com", "webhookId", "secretHash", redis.Configuration{})

		assert.EqualError(t, err, "marshaling error")
	})

	t.Run("Failed sendhooks due to request preparation errors", func(t *testing.T) {
		resetMocks()
		prepareRequest = func(url string, jsonBytes []byte, secretHash string, configuration redis.Configuration) (*http.Request, error) {
			return nil, errors.New("request preparation error")
		}

		err := SendWebhook(nil, "http://dummy.com", "webhookId", "secretHash", redis.Configuration{})

		assert.EqualError(t, err, "request preparation error")
	})

	t.Run("Failed sendhooks due to response processing errors", func(t *testing.T) {
		resetMocks()
		processResponse = func(resp *http.Response) (string, []byte, error) {
			return "failed", nil, errors.New("response processing error")
		}

		err := SendWebhook(nil, "http://dummy.com", "webhookId", "secretHash", redis.Configuration{})

		assert.EqualError(t, err, "response processing error")
	})

	t.Run("Logging on failed sendhooks delivery", func(t *testing.T) {
		resetMocks()
		processResponse = func(resp *http.Response) (string, []byte, error) {
			return "failed", []byte("error body"), nil
		}

		err := SendWebhook(nil, "http://dummy.com", "webhookId", "secretHash", redis.Configuration{})
		if !webhookLoggerInvoked {
			assert.Fail(t, "Expected WebhookLogger to be invoked")
		}

		assert.EqualError(t, err, "failed")
	})
}

func resetMocks() {
	marshalJSON = marshalJSONOrig
	prepareRequest = prepareRequestOrig
	sendRequest = sendRequestOrig
	processResponse = processResponseOrig
}
