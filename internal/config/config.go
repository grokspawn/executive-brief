package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the complete configuration
type Config struct {
	User        User              `yaml:"user"`
	Teammates   []Teammate        `yaml:"teammates"`
	Sources     Sources           `yaml:"sources"`
	MatrixRules MatrixRules       `yaml:"matrix_rules"`
	Output      OutputPreferences `yaml:"output"`
}

// User represents the current user
type User struct {
	Name     string `yaml:"name"`
	GitHub   string `yaml:"github"`
	Jira     string `yaml:"jira"`
	Email    string `yaml:"email"`
	Slack    string `yaml:"slack"`
	Timezone string `yaml:"timezone"`
}

// Teammate represents a team member
type Teammate struct {
	Name     string      `yaml:"name"`
	GitHub   string      `yaml:"github"`
	Jira     string      `yaml:"jira"`
	Email    string      `yaml:"email"`
	Slack    interface{} `yaml:"slack"` // Can be string or SlackInfo
	Priority string      `yaml:"priority"`
	Notes    string      `yaml:"notes"`
}

// SlackInfo represents Slack identifier details
type SlackInfo struct {
	UID    string `yaml:"uid"`
	Handle string `yaml:"handle"`
}

// Sources represents all data sources
type Sources struct {
	Jira       JiraSource   `yaml:"jira"`
	GitHub     GitHubSource `yaml:"github"`
	Slack      SlackSource  `yaml:"slack"`
	GoogleDocs GoogleDocs   `yaml:"google_docs"`
}

// SlackSource represents Slack configuration
type SlackSource struct {
	Enabled          bool     `yaml:"enabled"`
	TokenSource      string   `yaml:"token_source"`    // "env" or "file"
	TokenEnvVar      string   `yaml:"token_env_var"`   // Environment variable for token (default: SLACK_XOXC_TOKEN)
	TokenFile        string   `yaml:"token_file"`      // File path for token
	CookieEnvVar     string   `yaml:"cookie_env_var"`  // Environment variable for cookie (default: SLACK_XOXD_TOKEN, required for xoxc- tokens)
	UserAgentEnvVar  string   `yaml:"user_agent_env_var"` // Environment variable for user-agent (default: SLACK_USER_AGENT, required for xoxc- tokens)
	Channels         []string `yaml:"channels"`
	IncludeDMs       bool     `yaml:"include_dms"`
	IncludeMentions  bool     `yaml:"include_mentions"`
	IncludeThreads   bool     `yaml:"include_threads"`
	MinReactionCount int      `yaml:"min_reaction_count"`
}

// JiraSource represents Jira configuration
type JiraSource struct {
	Enabled     bool           `yaml:"enabled"`
	Server      string         `yaml:"server"`
	Projects    []string       `yaml:"projects"`
	JQLFilters  []string       `yaml:"jql_filters"`
	Dashboards  []JiraResource `yaml:"dashboards"`
	Boards      []JiraResource `yaml:"boards"`
}

// JiraResource represents a Jira board or dashboard
type JiraResource struct {
	URL         string `yaml:"url"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// GitHubSource represents GitHub configuration
type GitHubSource struct {
	Enabled       bool              `yaml:"enabled"`
	Organizations []string          `yaml:"organizations"`
	Repositories  []string          `yaml:"repositories"`
	Filters       GitHubFilters     `yaml:"filters"`
	CheckCI       bool              `yaml:"check_ci"`
}

// GitHubFilters represents GitHub query filters
type GitHubFilters struct {
	ReviewRequested bool `yaml:"review_requested"`
	Mentioned       bool `yaml:"mentioned"`
	TeamPRs         bool `yaml:"team_prs"`
	Assigned        bool `yaml:"assigned"`
	Authored        bool `yaml:"authored"`
}

// GoogleDocs represents Google Docs configuration
type GoogleDocs struct {
	Enabled       bool     `yaml:"enabled"`
	Folders       []string `yaml:"folders"`
	TrackComments bool     `yaml:"track_comments"`
	EditableOnly  bool     `yaml:"editable_only"`
}

// MatrixRules represents Eisenhower matrix categorization rules
type MatrixRules struct {
	UrgentKeywords    []string  `yaml:"urgent_keywords"`
	ImportantKeywords []string  `yaml:"important_keywords"`
	TimeBased         TimeBased `yaml:"time_based"`
}

// TimeBased represents time-based urgency/importance rules
type TimeBased struct {
	UrgentWithinDays    int `yaml:"urgent_within_days"`
	ImportantWithinDays int `yaml:"important_within_days"`
}

// OutputPreferences represents output formatting preferences
type OutputPreferences struct {
	Emojis          map[string]string `yaml:"emojis"`
	GroupBy         string            `yaml:"group_by"`
	SortBy          string            `yaml:"sort_by"`
	MaxPerQuadrant  int               `yaml:"max_per_quadrant"`
	ShowScores      bool              `yaml:"show_scores"`
}

// Load loads configuration from teammates.yaml and optional exec-brief.yaml
func Load(teammatesPath, sourcesPath string) (*Config, error) {
	// Find teammates.yaml
	if teammatesPath == "" {
		var err error
		teammatesPath, err = findConfig("teammates.yaml")
		if err != nil {
			return nil, fmt.Errorf("teammates.yaml not found: %w", err)
		}
	}

	// Load teammates.yaml
	data, err := os.ReadFile(teammatesPath)
	if err != nil {
		return nil, fmt.Errorf("error reading teammates.yaml: %w", err)
	}

	cfg := &Config{
		// Set defaults
		Sources: Sources{
			Jira: JiraSource{
				Enabled: true,
				Server:  "https://jira.example.com",
			},
			GitHub: GitHubSource{
				Enabled: true,
				Filters: GitHubFilters{
					ReviewRequested: true,
					Mentioned:       true,
					TeamPRs:         true,
					Assigned:        true,
					Authored:        true,
				},
			},
			Slack: SlackSource{
				Enabled:          true,
				TokenSource:      "env",
				TokenEnvVar:      "SLACK_XOXC_TOKEN",
				IncludeDMs:       true,
				IncludeMentions:  true,
				IncludeThreads:   true,
				MinReactionCount: 2,
			},
		},
		MatrixRules: MatrixRules{
			UrgentKeywords:    []string{"blocker", "critical", "urgent", "emergency", "production"},
			ImportantKeywords: []string{"feature", "security", "performance", "teammate"},
			TimeBased: TimeBased{
				UrgentWithinDays:    2,
				ImportantWithinDays: 7,
			},
		},
		Output: OutputPreferences{
			Emojis: map[string]string{
				"blocker":       "🔥",
				"security":      "🔒",
				"feature":       "⭐",
				"bug":           "🐛",
				"documentation": "📝",
				"teammate":      "👥",
				"waiting":       "⏰",
			},
			GroupBy:        "teammate",
			SortBy:         "urgency",
			MaxPerQuadrant: 20,
			ShowScores:     false,
		},
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("error parsing teammates.yaml: %w", err)
	}

	// Default timezone if not set
	if cfg.User.Timezone == "" {
		cfg.User.Timezone = "America/New_York"
	}

	// Load exec-brief.yaml if provided or found
	if sourcesPath == "" {
		sourcesPath, _ = findConfig("exec-brief.yaml")
	}

	if sourcesPath != "" {
		data, err := os.ReadFile(sourcesPath)
		if err == nil {
			// Merge with existing config
			if err := yaml.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("error parsing exec-brief.yaml: %w", err)
			}
		}
	}

	return cfg, nil
}

// findConfig searches for a config file in current directory and ~/.config/exec-brief/
func findConfig(filename string) (string, error) {
	// Check current directory
	if _, err := os.Stat(filename); err == nil {
		return filename, nil
	}

	// Check ~/.config/exec-brief/
	home, err := os.UserHomeDir()
	if err == nil {
		configPath := filepath.Join(home, ".config", "exec-brief", filename)
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}
	}

	return "", fmt.Errorf("file not found: %s", filename)
}
