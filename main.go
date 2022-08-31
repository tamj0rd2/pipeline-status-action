package main

import (
	"context"
	"flag"
	"fmt"
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
		log.Fatal(err)
	}

	fmt.Println("all status checks completed successfully")
}

type config struct {
	token, sha string
	owner      string
	repoName   string
	checkNames []string
}

func parseArgs() (config, error) {
	var token, repo, sha, checkNames string

	flag.StringVar(&token, "token", "", "GitHub token")
	flag.StringVar(&repo, "repository", "", "GitHub repository")
	flag.StringVar(&sha, "sha", "", "Commit SHA")
	flag.StringVar(&checkNames, "checkNames", "", "A comma separated list of the checks to run, e.g check1,check2,check3")
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

	splitRepo := strings.SplitN(repo, "/", 2)
	owner := splitRepo[0]
	repoName := splitRepo[1]

	return config{
		token:      token,
		sha:        sha,
		owner:      owner,
		repoName:   repoName,
		checkNames: strings.Split(checkNames, ","),
	}, nil
}
