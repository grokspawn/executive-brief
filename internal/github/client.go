package github

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v60/github"
	"github.com/grokspawn/executive-brief/internal/config"
	"github.com/grokspawn/executive-brief/internal/matrix"
)

// Query queries GitHub for PRs and issues
func Query(cfg *config.Config, startTime, endTime time.Time) ([]matrix.Item, error) {
	if !cfg.Sources.GitHub.Enabled {
		return nil, nil
	}

	// Create GitHub client (uses GITHUB_TOKEN env var or gh CLI auth)
	client := github.NewClient(nil)
	ctx := context.Background()

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
					items = append(items, normalizeIssue(pr, "review_requested"))
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
					items = append(items, normalizeIssue(pr, "teammate_pr"))
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
					items = append(items, normalizeIssue(issue, "mentioned"))
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
					items = append(items, normalizeIssue(pr, "authored"))
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
					items = append(items, normalizeIssue(issue, "assigned"))
				}
			}
		}
	}

	return items, nil
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
