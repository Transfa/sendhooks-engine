package sender

import (
	"errors"
	"fmt"
	"sendhooks/adapter"
	"sendhooks/logging"
)

// SendWebhook sends a JSON POST request to the specified URL
func SendWebhook(data interface{}, url string, webhookId string, secretHash string, configuration adapter.Configuration) error {
	jsonBytes, err := marshalJSON(data)
	if err != nil {
		return err
	}

	req, err := prepareRequest(url, jsonBytes, secretHash, configuration)
	if err != nil {
		return err
	}

	resp, err := sendRequest(req)
	if err != nil {

		return err
	}

	defer closeResponse(resp.Body)

	status, respBody, statusCode, err := processResponse(resp)
	if err != nil {
		return err
	}

	message := fmt.Sprintf("webhook sending failed with status: %d, response body: %s", statusCode, string(respBody))

	if status == "failed" {
		logging.WebhookLogger(logging.WarningType, fmt.Errorf(message))
		return errors.New(message)
	}

	return nil
}
