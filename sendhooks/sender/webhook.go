package sender

import (
	"errors"
	"fmt"
	"sendhooks/logging"
	redisClient "sendhooks/redis"
)

// SendWebhook sends a JSON POST request to the specified URL
func SendWebhook(data interface{}, url string, webhookId string, secretHash string, configuration redisClient.Configuration) error {
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

	status, respBody, err := processResponse(resp)
	if err != nil {
		return err
	}

	if status == "failed" {
		logging.WebhookLogger(logging.WarningType, fmt.Errorf("sendhooks failed with status: %s, response body: %s", status, string(respBody)))
		return errors.New(status)
	}

	return nil
}
