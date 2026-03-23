package github

import (
	"time"

	"github.com/grokspawn/executive-brief/internal/config"
	"github.com/grokspawn/executive-brief/internal/matrix"
)

// GitHubSource implements the Source interface
type GitHubSource struct{}

// Name returns the source identifier
func (g *GitHubSource) Name() string {
	return "github"
}

// Enabled checks if GitHub is enabled in configuration
func (g *GitHubSource) Enabled(cfg *config.Config) bool {
	return cfg.Sources.GitHub.Enabled
}

// Validate checks if GitHub authentication is configured
func (g *GitHubSource) Validate(cfg *config.Config) error {
	// Validate by testing authentication
	return ValidateAuth()
}

// Query fetches GitHub items within the time range
func (g *GitHubSource) Query(cfg *config.Config, startTime, endTime time.Time) ([]matrix.Item, error) {
	// Delegate to existing Query function
	return Query(cfg, startTime, endTime)
}
