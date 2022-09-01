package github

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/go-github/v42/github"
	"golang.org/x/oauth2"
)

type Service struct {
	client *github.Client
}

func NewService(ctx context.Context, githubToken string) *Service {
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: githubToken}))
	client := github.NewClient(httpClient)

	return &Service{
		client: client,
	}
}

func (s Service) WaitForChecksToSucceed(ctx context.Context, timeout time.Duration, owner string, repo string, sha string, checkNames []string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	const sleepTimeSeconds = 30
	statusTracker := newStatusTracker(checkNames)

	for {
		if err := ctx.Err(); err != nil {
			return statusTracker.GetIncompleteChecks(), fmt.Errorf("timed out waiting for checks to start/complete: %w", err)
		}

		if err := s.check(ctx, owner, repo, sha, statusTracker); err != nil {
			return nil, fmt.Errorf("failed to get statuses for commit - %w", err)
		}

		if failedChecks := statusTracker.GetFailedChecks(); len(failedChecks) > 0 {
			return statusTracker.GetFailedChecks(), errors.New("one or more checks failed")
		}

		if statusTracker.AllCompletedSuccessfully() {
			return nil, nil
		}

		checksInProgress := statusTracker.GetIncompleteChecks()
		log.Printf(
			"waiting for some checks to start and/or complete - %s. will check again in %d seconds\n",
			strings.Join(checksInProgress, ", "),
			sleepTimeSeconds,
		)
		time.Sleep(time.Second * time.Duration(sleepTimeSeconds))
	}
}

func (s Service) GetCommitInfo(ctx context.Context, owner, repo, sha string) (string, string, string) {
	c, _, err := s.client.Repositories.GetCommit(ctx, owner, repo, sha, nil)
	if err != nil {
		return "failed to retrieve", "failed to retrieve", fmt.Sprintf("https://github.com/%s/%s/commit/%s", owner, repo, sha)
	}

	trimmedCommitMessage := c.Commit.GetMessage()
	if len(trimmedCommitMessage) > 45 {
		trimmedCommitMessage = trimmedCommitMessage[:45] + "..."
	}

	return c.Commit.GetAuthor().GetName(), trimmedCommitMessage, c.GetHTMLURL()
}

func (s Service) check(ctx context.Context, owner string, repo string, sha string, statusTracker statusTracker) error {
	combinedStatus, _, err := s.client.Repositories.GetCombinedStatus(ctx, owner, repo, sha, nil)
	if err != nil {
		return err
	}

	for name, currentRecordedStatus := range statusTracker {
		if currentRecordedStatus != nil {
			continue
		}

		for _, status := range combinedStatus.Statuses {
			if status.GetContext() == name {
				switch status.GetState() {
				case "success":
					statusTracker[name] = github.Bool(true)
				case "error":
					statusTracker[name] = github.Bool(false)
				case "failure":
					statusTracker[name] = github.Bool(false)
				}
				break
			}
		}
	}

	return nil
}

type statusTracker map[string]*bool

func newStatusTracker(checkNames []string) statusTracker {
	tracker := make(map[string]*bool)
	for _, name := range checkNames {
		tracker[name] = nil
	}
	return tracker
}

func (t statusTracker) GetFailedChecks() []string {
	var failedChecks []string
	for name, status := range t {
		if status != nil && !*status {
			failedChecks = append(failedChecks, name)
		}
	}
	return failedChecks
}

func (t statusTracker) GetIncompleteChecks() []string {
	var incompleteChecks []string
	for name, status := range t {
		if status == nil {
			incompleteChecks = append(incompleteChecks, name)
		}
	}
	return incompleteChecks
}

func (t statusTracker) AllCompletedSuccessfully() bool {
	for _, wasSuccessful := range t {
		if wasSuccessful == nil {
			return false
		}

		if !*wasSuccessful {
			return false
		}
	}

	return true
}
