package sender

import (
	"errors"
	"fmt"
	"webhook/logging"
)

// SendWebhook sends a JSON POST request to the specified URL and updates the event status in the database
func SendWebhook(data interface{}, url string, webhookId string, secretHash string) error {
	jsonBytes, err := marshalJSON(data)
	if err != nil {
		return err
	}

	req, err := prepareRequest(url, jsonBytes, secretHash)
	if err != nil {
		return err
	}

	resp, err := sendRequest(req)
	if err != nil {

		return err
	}

	defer closeResponse(resp.Body)

	status, respBody, err := processResponse(resp)
	if err != nil {
		return err
	}

	if status == "failed" {
		logging.WebhookLogger(logging.WarningType, fmt.Errorf("webhook failed with status: %s, response body: %s", status, string(respBody)))
		return errors.New(status)
	}

	return nil
}
