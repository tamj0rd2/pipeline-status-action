package slack

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

func AlertThatStatusFailed(
	ctx context.Context,
	webhookURL string,
	commitURL, commitAuthor, commitMessage string,
	errorMessage string,
	failedStatuses []string,
) error {
	errorBody := fmt.Sprintf("*Error*: %s\n*Failed statuses*: %s", errorMessage, strings.Join(failedStatuses, ", "))
	commitDetails := fmt.Sprintf("*Commit author*: %s\n*Commit message*: %s", commitAuthor, commitMessage)

	requestBody := fmt.Sprintf(`{
		"blocks": [
			{
				"type": "header",
				"text": {
					"type": "plain_text",
					"text": ":x: Commit statuses failed",
					"emoji": true
				}
			},
			{
				"type": "section",
				"text": {
					"type": "mrkdwn",
					"text": "%s",
				}
			},
			{
				"type": "divider",
			},
			{
				"type": "section",
				"text": {
					"type": "mrkdwn",
					"text": "%s",
				}
			},
			{
				"type": "actions",
				"elements": [
					{
						"type": "button",
						"text": {
							"type": "plain_text",
							"text": "Github commit"
						},
						"url": "%s"
					}
				]
			}
		]
	}`, errorBody, commitDetails, commitURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, strings.NewReader(requestBody))
	if err != nil {
		return err
	}
	req.Header.Add("Content-type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK:
		return nil
	default:
		log.Println("Request body:", requestBody)

		body, _ := io.ReadAll(res.Body)
		log.Println("Response body:", string(body))

		return fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}
}
