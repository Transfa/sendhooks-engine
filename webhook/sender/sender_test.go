package sender

/*
This package contains mostly test functions for the utils used to send webhooks. If these utils works,
we can ensure that the send webhook function will function normally too. For a whole test suite for the sender function,
check out the webhook_test.go file.
*/

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockClient struct {
	MockDo func(req *http.Request) (*http.Response, error)
}

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return m.MockDo(req)
}

// Testing the MarshalJSON method
func TestMarshalJSON(t *testing.T) {
	data := map[string]string{
		"key": "value",
	}
	expectedJSON := `{"key":"value"}`

	jsonBytes, err := marshalJSON(data)

	assert.NoError(t, err)

	assert.Equal(t, expectedJSON, string(jsonBytes))
}

// Test for prepareRequest
func TestPrepareRequest(t *testing.T) {
	url := "http://example.com/webhook"
	jsonBytes := []byte(`{"key":"value"}`)
	secretHash := "secret123"

	req, err := prepareRequest(url, jsonBytes, secretHash)

	assert.NoError(t, err)

	assert.Equal(t, "application/json", req.Header.Get("Content-Type"))

	assert.Equal(t, secretHash, req.Header.Get("X-Secret-Hash"))
}

func TestSendRequest(t *testing.T) {
	HTTPClient = &MockClient{
		MockDo: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString("OK")),
			}, nil
		},
	}

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	resp, err := sendRequest(req)
	
	assert.NoError(t, err)
	
	body, _ := io.ReadAll(resp.Body)

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "OK", string(body))
}
