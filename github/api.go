package github

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	gh "github.com/google/go-github/v74/github"
)

type PullRequest struct {
	Number     int
	Title      string
	URL        string
	Repository string
	Draft      bool
}

type GhOperations struct {
	client *gh.Client
}

func NewGithubOperations(token string) *GhOperations {
	client := gh.NewClient(http.DefaultClient).WithAuthToken(token)
	return &GhOperations{
		client: client,
	}
}

func (ops *GhOperations) GetSelf() (string, error) {
	ctx := context.Background()
	user, _, err := ops.client.Users.Get(ctx, "")
	if err != nil {
		return "", err
	}
	return user.GetLogin(), nil
}

func (ops *GhOperations) CreatedOpenPRs() ([]PullRequest, error) {
	prs, err := ops.searchIssues("is:pr is:open author:@me archived:false",
		gh.SearchOptions{Sort: "updated", Order: "desc", ListOptions: gh.ListOptions{PerPage: 100}})
	if err != nil {
		return nil, fmt.Errorf("failed to find open pull requests created by the user: %e", err)
	}
	return prs, nil
}

func (ops *GhOperations) ReviewerOpenPRs() ([]PullRequest, error) {
	prs, err := ops.searchIssues("is:pr is:open review-requested:@me archived:false",
		gh.SearchOptions{Sort: "updated", Order: "desc", ListOptions: gh.ListOptions{PerPage: 100}})
	if err != nil {
		return nil, fmt.Errorf("failed to find open pull requests assigned to the user: %e", err)
	}
	return prs, nil
}

func (ops *GhOperations) GetAllSelfOpenPRs() ([]PullRequest, error) {
	prs, err := ops.searchIssues("is:pr is:open (review-requested:@me or author:@me) archived:false",
		gh.SearchOptions{Sort: "updated", Order: "desc", ListOptions: gh.ListOptions{PerPage: 100}})
	if err != nil {
		return nil, fmt.Errorf("failed to find open pull requests: %e", err)
	}
	return prs, nil
}

func (ops *GhOperations) SearchIssues(query string) ([]PullRequest, error) {
	prs, err := ops.searchIssues(query, gh.SearchOptions{Sort: "updated", Order: "desc", ListOptions: gh.ListOptions{PerPage: 100}})
	if err != nil {
		return nil, fmt.Errorf("failed to search github issues from query %s: %w", query, err)
	}
	return prs, err
}

func (ops *GhOperations) searchIssues(query string, options gh.SearchOptions) ([]PullRequest, error) {
	client := ops.client
	ctx := context.Background()
	items, _, err := client.Search.Issues(ctx, query, &options)

	if err != nil {
		return nil, fmt.Errorf("failed to find issues: %e", err)
	}
	createdPRs := make([]PullRequest, 0)
	for _, issuePR := range items.Issues {
		if issuePR.IsPullRequest() {
			repoName := repositoryNameFromGhURL(issuePR.GetRepositoryURL())
			createdPRs = append(createdPRs, PullRequest{
				Number:     issuePR.GetNumber(),
				Title:      issuePR.GetTitle(),
				Repository: repoName,
				URL:        issuePR.GetHTMLURL(),
				Draft:      issuePR.GetDraft(),
			})
		}
	}
	return createdPRs, nil
}

func repositoryNameFromGhURL(url string) string {
	if url[len(url)-1] == '/' {
		url = url[:len(url)-1] // Remove trailing slash if present
	}
	parts := strings.Split(url, "/")
	if len(parts) < 2 {
		return ""
	}
	return strings.Join(parts[len(parts)-2:], "/")
}
