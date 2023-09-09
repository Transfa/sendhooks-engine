package sender

import (
	"bytes"
	"io"
	"net/http"
	"testing"
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
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	if string(jsonBytes) != expectedJSON {
		t.Fatalf("Expected %s, but got %s", expectedJSON, string(jsonBytes))
	}
}

// Test for prepareRequest
func TestPrepareRequest(t *testing.T) {
	url := "http://example.com/webhook"
	jsonBytes := []byte(`{"key":"value"}`)
	secretHash := "secret123"

	req, err := prepareRequest(url, jsonBytes, secretHash)
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	if req.Header.Get("Content-Type") != "application/json" {
		t.Fatalf("Expected header Content-Type to be application/json but got %s", req.Header.Get("Content-Type"))
	}

	if req.Header.Get("X-Secret-Hash") != secretHash {
		t.Fatalf("Expected header X-Secret-Hash to be %s but got %s", secretHash, req.Header.Get("X-Secret-Hash"))
	}
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
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "OK" {
		t.Fatalf("Expected body to be OK but got %s", string(body))
	}
}
