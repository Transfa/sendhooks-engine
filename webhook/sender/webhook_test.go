package sender

import (
	"errors"
	"net/http"
	"testing"
	"webhook/logging"
)

var (
	marshalJSONOrig      = marshalJSON
	prepareRequestOrig   = prepareRequest
	sendRequestOrig      = sendRequest
	processResponseOrig  = processResponse
	webhookLoggerInvoked = false
)

func TestSendWebhook(t *testing.T) {
	logging.WebhookLogger = func(errorType string, errorMessage error) error {
		webhookLoggerInvoked = true
		return nil
	}

	t.Run("Successful webhook sending", func(t *testing.T) {
		resetMocks() // Reset all mocks to original functions

		err := SendWebhook(nil, "http://dummy.com", "webhookId", "secretHash")
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
	})

	t.Run("Failed webhook due to marshaling errors", func(t *testing.T) {
		resetMocks()
		marshalJSON = func(data interface{}) ([]byte, error) {
			return nil, errors.New("marshaling error")
		}

		err := SendWebhook(nil, "http://dummy.com", "webhookId", "secretHash")
		if err == nil || err.Error() != "marshaling error" {
			t.Fatalf("Expected marshaling error, but got: %v", err)
		}
	})

	t.Run("Failed webhook due to request preparation errors", func(t *testing.T) {
		resetMocks()
		prepareRequest = func(url string, jsonBytes []byte, secretHash string) (*http.Request, error) {
			return nil, errors.New("request preparation error")
		}

		err := SendWebhook(nil, "http://dummy.com", "webhookId", "secretHash")
		if err == nil || err.Error() != "request preparation error" {
			t.Fatalf("Expected request preparation error, but got: %v", err)
		}
	})

	t.Run("Failed webhook due to response processing errors", func(t *testing.T) {
		resetMocks()
		processResponse = func(resp *http.Response) (string, []byte, error) {
			return "failed", nil, errors.New("response processing error")
		}

		err := SendWebhook(nil, "http://dummy.com", "webhookId", "secretHash")
		if err == nil || err.Error() != "response processing error" {
			t.Fatalf("Expected response processing error, but got: %v", err)
		}
	})

	t.Run("Logging on failed webhook delivery", func(t *testing.T) {
		resetMocks()
		processResponse = func(resp *http.Response) (string, []byte, error) {
			return "failed", []byte("error body"), nil
		}

		webhookLoggerInvoked = false
		err := SendWebhook(nil, "http://dummy.com", "webhookId", "secretHash")
		if err == nil || err.Error() != "failed" {
			t.Fatalf("Expected failed status, but got: %v", err)
		}

		if !webhookLoggerInvoked {
			t.Fatalf("Expected WebhookLogger to be invoked")
		}
	})
}

func resetMocks() {
	marshalJSON = marshalJSONOrig
	prepareRequest = prepareRequestOrig
	sendRequest = sendRequestOrig
	processResponse = processResponseOrig
}
