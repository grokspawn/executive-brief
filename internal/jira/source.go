package jira

import (
	"time"

	"github.com/grokspawn/executive-brief/internal/config"
	"github.com/grokspawn/executive-brief/internal/matrix"
)

// JiraSource implements the Source interface
type JiraSource struct{}

// Name returns the source identifier
func (j *JiraSource) Name() string {
	return "jira"
}

// Enabled checks if Jira is enabled in configuration
func (j *JiraSource) Enabled(cfg *config.Config) bool {
	return cfg.Sources.Jira.Enabled
}

// Validate checks if Jira authentication is configured
func (j *JiraSource) Validate(cfg *config.Config) error {
	// Validate by attempting to load the API token
	_, err := LoadAPIToken()
	return err
}

// Query fetches Jira items within the time range
func (j *JiraSource) Query(cfg *config.Config, startTime, endTime time.Time) ([]matrix.Item, error) {
	// Delegate to existing Query function
	return Query(cfg, startTime, endTime)
}
