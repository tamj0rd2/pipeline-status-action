package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/tamj0rd2/pipeline-status-action/slack"
	"log"
	"os"
	"strings"
	"time"

	"github.com/tamj0rd2/pipeline-status-action/github"
)

func main() {
	config, err := parseArgs()
	if err != nil {
		log.Println(err)
		log.Println(`Usage: main.go -token=<token> -repository=<repository> -sha=<sha>`)
		flag.PrintDefaults()
		os.Exit(1)
	}

	ctx := context.Background()

	github := github.NewService(ctx, config.token)

	if err := github.WaitForChecksToSucceed(ctx, time.Hour, config.owner, config.repoName, config.sha, config.checkNames); err != nil {
		log.Println(err)

		if err := slack.AlertThatChecksFailed(ctx, config.slackWebhookURL, config.owner, config.repoName, config.sha, err.Error()); err != nil {
			log.Fatal(err)
		}

		log.Println("slack alert sent")
		os.Exit(0)
	}

	fmt.Println("all status checks completed successfully")
}

type config struct {
	token, sha      string
	owner           string
	repoName        string
	checkNames      []string
	slackWebhookURL string
}

func parseArgs() (config, error) {
	var token, repo, sha, checkNames, slackWebhookURL string

	flag.StringVar(&token, "token", "", "GitHub token")
	flag.StringVar(&repo, "repository", "", "GitHub repository")
	flag.StringVar(&sha, "sha", "", "Commit SHA")
	flag.StringVar(&checkNames, "checkNames", "", "A comma separated list of the checks to run, e.g check1,check2,check3")
	flag.StringVar(&slackWebhookURL, "slackWebhookURL", "", "The slack webhook URL")
	flag.Parse()

	if token == "" {
		return config{}, fmt.Errorf("token is required")
	}

	if repo == "" {
		return config{}, fmt.Errorf("repository is required")
	}

	if sha == "" {
		return config{}, fmt.Errorf("sha is required")
	}

	if checkNames == "" {
		return config{}, fmt.Errorf("checkNames is required")
	}

	if slackWebhookURL == "" {
		return config{}, fmt.Errorf("slackWebhookURL is required")
	}

	splitRepo := strings.SplitN(repo, "/", 2)
	owner := splitRepo[0]
	repoName := splitRepo[1]

	return config{
		token:           token,
		sha:             sha,
		owner:           owner,
		repoName:        repoName,
		checkNames:      strings.Split(checkNames, ","),
		slackWebhookURL: slackWebhookURL,
	}, nil
}
