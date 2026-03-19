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

	// Load configuration
	cfg, err := config.Load(*configPath, *sourceConfigPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Calculate time range
	var startTime, endTime time.Time
	if *daily {
		startTime, endTime = calculateDailyRange(cfg.User.Timezone)
	} else if *date != "" {
		t, err := time.Parse("2006-01-02", *date)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid date format '%s'. Use YYYY-MM-DD\n", *date)
			os.Exit(1)
		}
		startTime = t
		endTime = t.Add(24 * time.Hour)
	} else {
		// Default to last 24 hours
		endTime = time.Now()
		startTime = endTime.Add(-24 * time.Hour)
	}

	// Determine which sources to query
	enabledSources := parseEnabledSources(*sources, cfg)

	// Collect items from all sources
	var items []matrix.Item

	if enabledSources["jira"] {
		jiraItems, err := jira.Query(cfg, startTime, endTime)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not query Jira: %v\n", err)
		} else {
			items = append(items, jiraItems...)
		}
	}

	if enabledSources["github"] {
		githubItems, err := github.Query(cfg, startTime, endTime)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not query GitHub: %v\n", err)
		} else {
			items = append(items, githubItems...)
		}
	}

	// Identify teammate involvement
	for i := range items {
		items[i].TeammatesInvolved = identifyTeammates(&items[i], cfg)
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
func parseEnabledSources(sourcesFlag string, cfg *config.Config) map[string]bool {
	enabled := make(map[string]bool)

	if sourcesFlag != "" {
		// Use explicit list
		for _, src := range strings.Split(sourcesFlag, ",") {
			enabled[strings.TrimSpace(src)] = true
		}
	} else {
		// Use config defaults
		if cfg.Sources.Jira.Enabled {
			enabled["jira"] = true
		}
		if cfg.Sources.GitHub.Enabled {
			enabled["github"] = true
		}
	}

	return enabled
}

// identifyTeammates identifies which teammates are involved in an item
func identifyTeammates(item *matrix.Item, cfg *config.Config) []string {
	teammates := make(map[string]bool)

	for _, tm := range cfg.Teammates {
		switch item.Source {
		case "jira":
			// Check assignee and reporter
			if item.Assignee == tm.Jira || item.Reporter == tm.Jira {
				teammates[tm.Name] = true
			}
		case "github":
			// Check author
			if item.Author == tm.GitHub {
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
