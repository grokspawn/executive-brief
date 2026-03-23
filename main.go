package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/grokspawn/executive-brief/internal/config"
	"github.com/grokspawn/executive-brief/internal/github"
	"github.com/grokspawn/executive-brief/internal/jira"
	"github.com/grokspawn/executive-brief/internal/matrix"
	"github.com/grokspawn/executive-brief/internal/output"
	"github.com/grokspawn/executive-brief/internal/slack"
	"github.com/grokspawn/executive-brief/internal/source"
)

func main() {
	// Parse command-line flags
	daily := flag.Bool("daily", false, "Cover yesterday-to-today activity (multi-timezone)")
	date := flag.String("date", "", "Generate brief for specific date (YYYY-MM-DD)")
	sources := flag.String("sources", "", "Comma-separated list of sources (jira,github)")
	teammatesOnly := flag.Bool("teammates-only", false, "Show only items involving teammates")
	format := flag.String("format", "markdown", "Output format (markdown, html)")
	save := flag.String("save", "", "Save to file (optional filename)")
	configPath := flag.String("config", "", "Path to teammates.yaml")
	sourceConfigPath := flag.String("source-config", "", "Path to exec-brief.yaml")

	flag.Parse()

	// Validate flags
	if *format != "markdown" && *format != "html" {
		fmt.Fprintf(os.Stderr, "Invalid format '%s'. Use 'markdown' or 'html'\n", *format)
		os.Exit(1)
	}

	if *date != "" {
		if _, err := time.Parse("2006-01-02", *date); err != nil {
			fmt.Fprintf(os.Stderr, "Invalid date format '%s'. Use YYYY-MM-DD\n", *date)
			os.Exit(1)
		}
	}

	// Load configuration
	cfg, err := config.Load(*configPath, *sourceConfigPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Create source registry and register all sources
	registry := source.NewRegistry()
	registry.Register(&jira.JiraSource{})
	registry.Register(&github.GitHubSource{})
	registry.Register(&slack.SlackSource{})

	// Determine which sources to query
	enabledSources, err := parseEnabledSources(*sources, cfg, registry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Validate all enabled sources before proceeding
	if err := registry.ValidateAll(cfg, enabledSources); err != nil {
		fmt.Fprintf(os.Stderr, "Validation failed: %v\n", err)
		os.Exit(1)
	}

	// Calculate time range
	var startTime, endTime time.Time
	if *daily {
		startTime, endTime = calculateDailyRange(cfg.User.Timezone)
	} else if *date != "" {
		// Date already validated above
		t, _ := time.Parse("2006-01-02", *date)
		startTime = t
		endTime = t.Add(24 * time.Hour)
	} else {
		// Default to last 24 hours
		endTime = time.Now()
		startTime = endTime.Add(-24 * time.Hour)
	}

	// Collect items from all sources (teammates already identified by each source)
	items, err := registry.QueryAll(cfg, startTime, endTime, enabledSources)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Error querying sources: %v\n", err)
	}

	// Filter teammates-only if requested
	if *teammatesOnly {
		filtered := make([]matrix.Item, 0)
		for _, item := range items {
			if len(item.TeammatesInvolved) > 0 {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}

	// Categorize into Eisenhower Matrix
	categorized := matrix.Categorize(items, cfg)

	// Generate output
	var briefContent string
	if *format == "html" {
		briefContent = output.GenerateHTML(categorized, cfg, startTime, endTime)
	} else {
		briefContent = output.GenerateMarkdown(categorized, cfg, startTime, endTime)
	}

	// Save or print
	if *save != "" {
		filename := *save
		if filename == "true" || filename == "" {
			// Default filename
			filename = fmt.Sprintf("exec-brief-%s.md", time.Now().Format("2006-01-02"))
		}
		if err := os.WriteFile(filename, []byte(briefContent), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving to file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Executive brief saved to: %s\n", filename)
	} else {
		fmt.Print(briefContent)
	}
}

// calculateDailyRange calculates the time range for --daily flag
// Considers both EST/EDT and user's home timezone
func calculateDailyRange(userTimezone string) (time.Time, time.Time) {
	// Load timezones
	homeLoc, err := time.LoadLocation(userTimezone)
	if err != nil {
		homeLoc = time.UTC
	}
	eastLoc, _ := time.LoadLocation("America/New_York")

	now := time.Now()

	// Yesterday in home timezone
	homeNow := now.In(homeLoc)
	homeStart := time.Date(homeNow.Year(), homeNow.Month(), homeNow.Day()-1, 0, 0, 0, 0, homeLoc)
	homeEnd := homeNow

	// Yesterday in EST/EDT
	eastNow := now.In(eastLoc)
	eastStart := time.Date(eastNow.Year(), eastNow.Month(), eastNow.Day()-1, 0, 0, 0, 0, eastLoc)
	eastEnd := eastNow

	// Use earliest start and latest end
	startTime := homeStart
	if eastStart.Before(homeStart) {
		startTime = eastStart
	}

	endTime := homeEnd
	if eastEnd.After(homeEnd) {
		endTime = eastEnd
	}

	return startTime, endTime
}

// parseEnabledSources determines which sources to query
// Returns the intersection of requested sources and registered sources
func parseEnabledSources(sourcesFlag string, cfg *config.Config, registry *source.Registry) (map[string]bool, error) {
	// Get all registered sources
	registered := make(map[string]bool)
	for _, name := range registry.ListRegistered() {
		registered[name] = true
	}

	var requested map[string]bool

	if sourcesFlag != "" {
		// Use explicit list from flag
		requested = make(map[string]bool)
		for _, src := range strings.Split(sourcesFlag, ",") {
			name := strings.TrimSpace(src)
			if name != "" {
				// Validate that the source is registered
				if !registered[name] {
					return nil, fmt.Errorf("unknown source '%s'. Available sources: %s",
						name, strings.Join(registry.ListRegistered(), ", "))
				}
				requested[name] = true
			}
		}
	} else {
		// Use config defaults - sources that are both registered AND enabled in config
		requested = make(map[string]bool)
		for name := range registered {
			src, ok := registry.Get(name)
			if ok && src.Enabled(cfg) {
				requested[name] = true
			}
		}
	}

	return requested, nil
}
