package source

import (
	"fmt"
	"os"
	"time"

	"github.com/grokspawn/executive-brief/internal/config"
	"github.com/grokspawn/executive-brief/internal/matrix"
)

// Source represents a pluggable data source
type Source interface {
	// Name returns the source identifier (e.g., "jira", "github", "slack")
	Name() string

	// Enabled checks if this source is enabled in the configuration
	Enabled(cfg *config.Config) bool

	// Validate checks if the source can authenticate and is properly configured
	// This is called at startup before any queries are made
	Validate(cfg *config.Config) error

	// Query fetches items from this source within the time range
	Query(cfg *config.Config, startTime, endTime time.Time) ([]matrix.Item, error)
}

// Registry manages all registered sources
type Registry struct {
	sources map[string]Source
}

// NewRegistry creates a new source registry
func NewRegistry() *Registry {
	return &Registry{
		sources: make(map[string]Source),
	}
}

// Register adds a source to the registry
func (r *Registry) Register(s Source) {
	r.sources[s.Name()] = s
}

// Get retrieves a source by name
func (r *Registry) Get(name string) (Source, bool) {
	s, ok := r.sources[name]
	return s, ok
}

// QueryAll queries all enabled sources and aggregates results
func (r *Registry) QueryAll(cfg *config.Config, startTime, endTime time.Time, enabledSources map[string]bool) ([]matrix.Item, error) {
	var allItems []matrix.Item

	for name, source := range r.sources {
		// Skip if explicitly disabled by flag
		if len(enabledSources) > 0 {
			if !enabledSources[name] {
				continue
			}
		}

		// Skip if disabled in config
		if !source.Enabled(cfg) {
			continue
		}

		items, err := source.Query(cfg, startTime, endTime)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %s query failed: %v\n", name, err)
			continue
		}

		allItems = append(allItems, items...)
	}

	return allItems, nil
}

// ListEnabled returns names of all enabled sources
func (r *Registry) ListEnabled(cfg *config.Config) []string {
	var enabled []string
	for name, source := range r.sources {
		if source.Enabled(cfg) {
			enabled = append(enabled, name)
		}
	}
	return enabled
}

// ListRegistered returns names of all registered sources
func (r *Registry) ListRegistered() []string {
	var registered []string
	for name := range r.sources {
		registered = append(registered, name)
	}
	return registered
}

// ValidateAll validates all enabled sources
func (r *Registry) ValidateAll(cfg *config.Config, enabledSources map[string]bool) error {
	for name, source := range r.sources {
		// Skip if explicitly disabled by flag
		if len(enabledSources) > 0 {
			if !enabledSources[name] {
				continue
			}
		}

		// Skip if disabled in config
		if !source.Enabled(cfg) {
			continue
		}

		// Validate the source
		if err := source.Validate(cfg); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
	}

	return nil
}
