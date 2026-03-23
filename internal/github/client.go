package github

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/go-github/v60/github"
	"github.com/grokspawn/executive-brief/internal/config"
	"github.com/grokspawn/executive-brief/internal/matrix"
	"golang.org/x/oauth2"
)

// newClient creates an authenticated GitHub client using GITHUB_TOKEN
func newClient(ctx context.Context) *github.Client {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		// Return unauthenticated client if no token
		return github.NewClient(nil)
	}

	// Create OAuth2 token source
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

// ValidateAuth validates GitHub authentication
func ValidateAuth() error {
	ctx := context.Background()
	client := newClient(ctx)

	// Test authentication
	_, resp, err := client.Users.Get(ctx, "")
	if err != nil && resp != nil && resp.StatusCode == 401 {
		return fmt.Errorf(`GitHub authentication failed

To fix, choose one option:

Option 1 - Use gh CLI (recommended):
  gh auth login

Option 2 - Set GITHUB_TOKEN environment variable:
  1. Create a Personal Access Token at: https://github.com/settings/tokens
  2. Required scopes: repo, read:user
  3. Set the environment variable:

     export GITHUB_TOKEN=ghp_your_token_here

  Or add to shell profile (~/.bashrc, ~/.zshrc):
     echo 'export GITHUB_TOKEN=ghp_your_token_here' >> ~/.bashrc

The GitHub client will automatically use gh CLI credentials or GITHUB_TOKEN.`)
	}
	return nil
}

// Query queries GitHub for PRs and issues
func Query(cfg *config.Config, startTime, endTime time.Time) ([]matrix.Item, error) {
	if !cfg.Sources.GitHub.Enabled {
		return nil, nil
	}

	ctx := context.Background()
	client := newClient(ctx)

	var items []matrix.Item
	filters := cfg.Sources.GitHub.Filters

	// Format time for GitHub search
	since := startTime.Format("2006-01-02")

	// PRs requesting review
	if filters.ReviewRequested {
		query := fmt.Sprintf("is:pr is:open review-requested:@me updated:>%s", since)
		prs, err := searchIssues(ctx, client, query)
		if err == nil {
			for _, pr := range prs {
				if pr.UpdatedAt.After(startTime) && pr.UpdatedAt.Before(endTime) {
					item := normalizeIssue(pr, "review_requested")
					item.TeammatesInvolved = identifyGitHubTeammates(item, cfg)
					items = append(items, item)
				}
			}
		}
	}

	// PRs from teammates
	if filters.TeamPRs {
		for _, tm := range cfg.Teammates {
			if tm.GitHub == "" {
				continue
			}
			query := fmt.Sprintf("is:pr is:open author:%s updated:>%s", tm.GitHub, since)
			prs, err := searchIssues(ctx, client, query)
			if err != nil {
				continue
			}
			for _, pr := range prs {
				if pr.UpdatedAt.After(startTime) && pr.UpdatedAt.Before(endTime) {
					item := normalizeIssue(pr, "teammate_pr")
					item.TeammatesInvolved = identifyGitHubTeammates(item, cfg)
					items = append(items, item)
				}
			}
		}
	}

	// Issues/PRs mentioning you
	if filters.Mentioned {
		query := fmt.Sprintf("mentions:@me updated:>%s", since)
		issues, err := searchIssues(ctx, client, query)
		if err == nil {
			for _, issue := range issues {
				if issue.UpdatedAt.After(startTime) && issue.UpdatedAt.Before(endTime) {
					item := normalizeIssue(issue, "mentioned")
					item.TeammatesInvolved = identifyGitHubTeammates(item, cfg)
					items = append(items, item)
				}
			}
		}
	}

	// Your PRs
	if filters.Authored {
		query := fmt.Sprintf("is:pr is:open author:@me updated:>%s", since)
		prs, err := searchIssues(ctx, client, query)
		if err == nil {
			for _, pr := range prs {
				if pr.UpdatedAt.After(startTime) && pr.UpdatedAt.Before(endTime) {
					item := normalizeIssue(pr, "authored")
					item.TeammatesInvolved = identifyGitHubTeammates(item, cfg)
					items = append(items, item)
				}
			}
		}
	}

	// Issues assigned to you
	if filters.Assigned {
		query := fmt.Sprintf("is:issue is:open assignee:@me updated:>%s", since)
		issues, err := searchIssues(ctx, client, query)
		if err == nil {
			for _, issue := range issues {
				if issue.UpdatedAt.After(startTime) && issue.UpdatedAt.Before(endTime) {
					item := normalizeIssue(issue, "assigned")
					item.TeammatesInvolved = identifyGitHubTeammates(item, cfg)
					items = append(items, item)
				}
			}
		}
	}

	return items, nil
}

// identifyGitHubTeammates identifies teammates based on GitHub usernames
func identifyGitHubTeammates(item matrix.Item, cfg *config.Config) []string {
	teammates := make(map[string]bool)

	for _, tm := range cfg.Teammates {
		if tm.GitHub != "" {
			if item.Author == tm.GitHub || item.Assignee == tm.GitHub {
				teammates[tm.Name] = true
			}
		}
	}

	result := make([]string, 0, len(teammates))
	for name := range teammates {
		result = append(result, name)
	}
	return result
}

// searchIssues searches for issues/PRs using GitHub Search API
func searchIssues(ctx context.Context, client *github.Client, query string) ([]*github.Issue, error) {
	opts := &github.SearchOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	result, _, err := client.Search.Issues(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("GitHub search failed: %w", err)
	}

	return result.Issues, nil
}

// normalizeIssue converts a GitHub issue/PR to a matrix.Item
func normalizeIssue(issue *github.Issue, reason string) matrix.Item {
	labels := make([]string, 0, len(issue.Labels))
	for _, label := range issue.Labels {
		if label.Name != nil {
			labels = append(labels, *label.Name)
		}
	}

	// Add reason as a pseudo-label
	if reason != "" {
		labels = append(labels, reason)
	}

	itemType := "issue"
	if issue.IsPullRequest() {
		itemType = "pull_request"
	}

	item := matrix.Item{
		ID:        fmt.Sprintf("#%d", *issue.Number),
		Title:     *issue.Title,
		Source:    "github",
		URL:       *issue.HTMLURL,
		Type:      itemType,
		Status:    *issue.State,
		CreatedAt: issue.CreatedAt.Time,
		UpdatedAt: issue.UpdatedAt.Time,
		Labels:    labels,
	}

	if issue.User != nil && issue.User.Login != nil {
		item.Author = *issue.User.Login
	}

	if issue.Assignee != nil && issue.Assignee.Login != nil {
		item.Assignee = *issue.Assignee.Login
	}

	return item
}
