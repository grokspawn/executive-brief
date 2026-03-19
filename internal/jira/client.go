package jira

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/grokspawn/executive-brief/internal/config"
	"github.com/grokspawn/executive-brief/internal/matrix"
)

// Issue represents a Jira issue from the API
type Issue struct {
	Key    string `json:"key"`
	Fields struct {
		Summary  string   `json:"summary"`
		Created  string   `json:"created"`
		Updated  string   `json:"updated"`
		DueDate  *string  `json:"duedate"`
		Labels   []string `json:"labels"`
		Status   *struct {
			Name string `json:"name"`
		} `json:"status"`
		Priority *struct {
			Name string `json:"name"`
		} `json:"priority"`
		IssueType *struct {
			Name string `json:"name"`
		} `json:"issuetype"`
		Assignee *struct {
			DisplayName  string `json:"displayName"`
			EmailAddress string `json:"emailAddress"`
		} `json:"assignee"`
		Reporter *struct {
			DisplayName  string `json:"displayName"`
			EmailAddress string `json:"emailAddress"`
		} `json:"reporter"`
	} `json:"fields"`
}

// SearchResult represents Jira search API response
type SearchResult struct {
	Total  int     `json:"total"`
	Issues []Issue `json:"issues"`
}

// Client represents a Jira API client
type Client struct {
	Server   string
	Email    string
	APIToken string
	client   *http.Client
}

// NewClient creates a new Jira API client
func NewClient(server, email, apiToken string) *Client {
	return &Client{
		Server:   server,
		Email:    email,
		APIToken: apiToken,
		client:   &http.Client{Timeout: 30 * time.Second},
	}
}

// SearchJQL executes a JQL query using the v3 API
func (c *Client) SearchJQL(ctx context.Context, jql string, maxResults int) (*SearchResult, error) {
	url := fmt.Sprintf("%s/rest/api/3/search/jql", c.Server)

	payload := map[string]interface{}{
		"jql":        jql,
		"maxResults": maxResults,
		"fields": []string{
			"key", "summary", "status", "assignee", "reporter",
			"updated", "created", "priority", "labels", "duedate", "issuetype",
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.SetBasicAuth(c.Email, c.APIToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Jira API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &result, nil
}

// Query queries Jira for items within the time range
func Query(cfg *config.Config, startTime, endTime time.Time) ([]matrix.Item, error) {
	if !cfg.Sources.Jira.Enabled {
		return nil, nil
	}

	// Load API token
	apiToken, err := loadAPIToken()
	if err != nil {
		return nil, fmt.Errorf("error loading Jira API token: %w", err)
	}

	server := cfg.Sources.Jira.Server
	if server == "" {
		server = "https://jira.example.com"
	}

	client := NewClient(server, cfg.User.Jira, apiToken)
	ctx := context.Background()

	var items []matrix.Item

	// Build JQL queries
	daysSince := int(time.Since(startTime).Hours() / 24)
	if daysSince < 1 {
		daysSince = 1
	}

	// Query for user and teammates
	emails := []string{fmt.Sprintf(`"%s"`, cfg.User.Jira)}
	for _, tm := range cfg.Teammates {
		if tm.Jira != "" {
			emails = append(emails, fmt.Sprintf(`"%s"`, tm.Jira))
		}
	}

	jql := fmt.Sprintf("assignee in (%s) AND updated >= -%dd ORDER BY updated DESC",
		joinStrings(emails, ", "), daysSince)

	result, err := client.SearchJQL(ctx, jql, 100)
	if err != nil {
		return nil, err
	}
	for _, issue := range result.Issues {
		items = append(items, normalizeIssue(issue, server))
	}

	// Execute custom JQL filters if configured
	for _, jqlFilter := range cfg.Sources.Jira.JQLFilters {
		customJQL := fmt.Sprintf("(%s) AND updated >= -%dd", jqlFilter, daysSince)
		result, err := client.SearchJQL(ctx, customJQL, 100)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: JQL filter failed: %v\n", err)
			continue
		}
		for _, issue := range result.Issues {
			items = append(items, normalizeIssue(issue, server))
		}
	}

	// Query projects if configured
	for _, project := range cfg.Sources.Jira.Projects {
		projectJQL := fmt.Sprintf("project = %s AND updated >= -%dd ORDER BY updated DESC",
			project, daysSince)
		result, err := client.SearchJQL(ctx, projectJQL, 100)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Project query failed for %s: %v\n", project, err)
			continue
		}
		for _, issue := range result.Issues {
			items = append(items, normalizeIssue(issue, server))
		}
	}

	return items, nil
}

// normalizeIssue converts a Jira issue to a matrix.Item
func normalizeIssue(issue Issue, server string) matrix.Item {
	item := matrix.Item{
		ID:     issue.Key,
		Title:  issue.Fields.Summary,
		Source: "jira",
		URL:    fmt.Sprintf("%s/browse/%s", server, issue.Key),
		Labels: issue.Fields.Labels,
	}

	// Parse Jira timestamps (format: "2006-01-02T15:04:05.000-0700")
	if issue.Fields.Created != "" {
		if t, err := parseJiraTime(issue.Fields.Created); err == nil {
			item.CreatedAt = t
		}
	}
	if issue.Fields.Updated != "" {
		if t, err := parseJiraTime(issue.Fields.Updated); err == nil {
			item.UpdatedAt = t
		}
	}

	if issue.Fields.Status != nil {
		item.Status = issue.Fields.Status.Name
	}

	if issue.Fields.Priority != nil {
		item.Priority = issue.Fields.Priority.Name
	}

	if issue.Fields.IssueType != nil {
		item.Type = issue.Fields.IssueType.Name
	}

	if issue.Fields.Assignee != nil {
		item.Assignee = issue.Fields.Assignee.EmailAddress
		item.AssigneeName = issue.Fields.Assignee.DisplayName
	}

	if issue.Fields.Reporter != nil {
		item.Reporter = issue.Fields.Reporter.EmailAddress
	}

	if issue.Fields.DueDate != nil && *issue.Fields.DueDate != "" {
		if dueDate, err := time.Parse("2006-01-02", *issue.Fields.DueDate); err == nil {
			item.DueDate = &dueDate
		}
	}

	return item
}

// loadAPIToken loads the Jira API token from ~/.claude.json
func loadAPIToken() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting home directory: %w", err)
	}

	claudePath := filepath.Join(home, ".claude.json")
	data, err := os.ReadFile(claudePath)
	if err != nil {
		return "", fmt.Errorf("error reading ~/.claude.json: %w", err)
	}

	var claudeConfig map[string]interface{}
	if err := json.Unmarshal(data, &claudeConfig); err != nil {
		return "", fmt.Errorf("error parsing ~/.claude.json: %w", err)
	}

	// Navigate to mcpServers.atlassian.env.JIRA_API_TOKEN
	mcpServers, ok := claudeConfig["mcpServers"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("mcpServers not found in ~/.claude.json")
	}

	atlassian, ok := mcpServers["atlassian"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("mcpServers.atlassian not found in ~/.claude.json")
	}

	env, ok := atlassian["env"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("mcpServers.atlassian.env not found in ~/.claude.json")
	}

	token, ok := env["JIRA_API_TOKEN"].(string)
	if !ok || token == "" {
		return "", fmt.Errorf("JIRA_API_TOKEN not found in ~/.claude.json")
	}

	return token, nil
}

// parseJiraTime parses Jira's timestamp format
func parseJiraTime(s string) (time.Time, error) {
	// Jira uses format like "2026-02-08T23:30:17.213+0000"
	// Try multiple formats
	formats := []string{
		"2006-01-02T15:04:05.000-0700",
		"2006-01-02T15:04:05.000+0000",
		time.RFC3339,
		time.RFC3339Nano,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time: %s", s)
}

// joinStrings joins a slice of strings with a separator
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
