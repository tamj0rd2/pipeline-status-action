package slack

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

func AlertThatChecksFailed(ctx context.Context, webhookURL string, owner string, repo string, sha string, message string) error {
	commitURL := fmt.Sprintf("https://github.com/%s/%s/commit/%s", owner, repo, sha)

	requestBody := fmt.Sprintf(`{
		"blocks": [
			{
				"type": "header",
				"text": {
					"type": "plain_text",
					"text": "Commit status checks failed",
					"emoji": true
				}
			},
			{
				"type": "section",
				"text": {
					"type": "mrkdwn",
					"text": "%s"
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
	}`, message, commitURL)

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
