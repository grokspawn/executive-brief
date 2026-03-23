package slack

import (
	"time"

	"github.com/grokspawn/executive-brief/internal/config"
	"github.com/grokspawn/executive-brief/internal/matrix"
)

// SlackSource implements the Source interface
type SlackSource struct{}

// Name returns the source identifier
func (s *SlackSource) Name() string {
	return "slack"
}

// Enabled checks if Slack is enabled in configuration
func (s *SlackSource) Enabled(cfg *config.Config) bool {
	return cfg.Sources.Slack.Enabled
}

// Validate checks if Slack authentication is configured
func (s *SlackSource) Validate(cfg *config.Config) error {
	// Validate by testing authentication
	return ValidateAuth(cfg)
}

// Query fetches Slack items within the time range
func (s *SlackSource) Query(cfg *config.Config, startTime, endTime time.Time) ([]matrix.Item, error) {
	// Delegate to existing Query function
	return Query(cfg, startTime, endTime)
}
