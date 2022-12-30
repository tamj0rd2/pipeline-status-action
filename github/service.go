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

func (s Service) WaitForChecksToSucceed(ctx context.Context, timeout time.Duration, owner string, repo string, sha string, checkNames []string) ([]Status, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	const sleepTimeSeconds = 30
	statusTracker := newStatusTracker(checkNames)

	for {
		fmt.Println(statusTracker)
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
		var checksInProgressName []string
		for _, status := range checksInProgress {
			checksInProgressName = append(checksInProgressName, status.Name)
		}

		log.Printf(
			"waiting for some checks to start and/or complete - %s. will check again in %d seconds\n",
			strings.Join(checksInProgressName, ", "),
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

	for name, status := range statusTracker {
		if status.Finished {
			continue
		}

		for _, gitStatus := range combinedStatus.Statuses {
			if gitStatus.GetContext() == name {
				stat := statusTracker[name]
				stat.Finished = true
				stat.Url = gitStatus.GetTargetURL()
				switch gitStatus.GetState() {
				case "success":
					stat.Succeeded = true
					statusTracker[name] = stat
				case "error":
					stat.Succeeded = false
					statusTracker[name] = stat
				case "failure":
					stat.Succeeded = false
					statusTracker[name] = stat
				}
				break
			}
		}
	}

	return nil
}

type statusTracker map[string]Status

type Status struct {
	Name      string
	Succeeded bool
	Finished  bool
	Url       string
}

func newStatus(name string) Status {
	return Status{Name: name, Succeeded: false, Finished: false, Url: ""}
}

// type statusTracker struct {
// 	statuses []Status
// }

func newStatusTracker(checkNames []string) statusTracker {
	tracker := make(statusTracker)
	for _, name := range checkNames {
		tracker[name] = newStatus(name)
	}
	return tracker
}

func (t statusTracker) GetFailedChecks() []Status {
	var failedChecks []Status
	for _, status := range t {
		if !status.Succeeded && status.Finished {
			failedChecks = append(failedChecks, status)
		}
	}
	return failedChecks
}

func (t statusTracker) GetIncompleteChecks() []Status {
	var incompleteChecks []Status
	for _, status := range t {
		if !status.Succeeded && !status.Finished {
			incompleteChecks = append(incompleteChecks, status)
		}
	}
	return incompleteChecks
}

func (t statusTracker) AllCompletedSuccessfully() bool {
	for _, status := range t {
		if !status.Succeeded {
			return false
		}
	}

	return true
}
