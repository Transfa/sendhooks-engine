package sender

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"webhook/logging"
)

var HTTPClient = &http.Client{}

// SendWebhook sends a JSON POST request to the specified URL and updates the event status in the database
func SendWebhook(data interface{}, url string, webhookId string, secretHash string) error {
	// Marshal the data into JSON
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		logging.WebhookLogger(logging.ErrorType, fmt.Errorf("error marshaling JSON: %s", err))
		return err
	}

	// Prepare the webhook request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		logging.WebhookLogger(logging.WarningType, fmt.Errorf("error during the webhook request prepariton"))
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	if secretHash != "" {
		req.Header.Set("X-Secret-Hash", secretHash)
	}

	// Send the webhook request
	resp, err := HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			logging.WebhookLogger(logging.WarningType, fmt.Errorf("error closing response body: %s", err))
		}
	}(resp.Body)

	// Determine the status based on the response code
	status := "failed"

	var respBody []byte
	respBody, err = io.ReadAll(resp.Body)
	if err != nil {
		status = "failed"
		logging.WebhookLogger(logging.WarningType, fmt.Errorf("error reading response body: %s", err))
	}

	if resp.StatusCode == http.StatusOK {
		status = "delivered"
	} else {
		logging.WebhookLogger(logging.WarningType, fmt.Errorf("HTTP request failed with status code: %d, response body: %s", resp.StatusCode, string(respBody)))
	}

	if status == "failed" {
		logging.WebhookLogger(logging.ErrorType, fmt.Errorf("webhook failed with status: %s, response body: %s", status, string(respBody)))
		return errors.New(status)
	}

	return nil
}
