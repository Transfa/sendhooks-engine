package sender

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"webhook/logging"
)

// SendWebhook sends a JSON POST request to the specified URL and updates the event status in the database
func SendWebhook(data interface{}, url string, webhookId string) error {
	// Marshal the data into JSON
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Prepare the webhook request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	// Send the webhook request
	client := &http.Client{}
	resp, err := client.Do(req)
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
	if resp.StatusCode == http.StatusOK {
		status = "delivered"
	}

	log.Println(status)

	if status == "failed" {
		return errors.New(status)
	}

	return nil
}
