package sender

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
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
	signature := os.Getenv("WEBHOOK_SIGNATURE")

	if signature != "" {
		req.Header.Set("X-Secret-Hash", signature)
	}

	// Send the webhook request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			log.Println("Error closing response body:", err)
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
