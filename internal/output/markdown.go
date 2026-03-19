package output

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/grokspawn/executive-brief/internal/config"
	"github.com/grokspawn/executive-brief/internal/matrix"
)

// GenerateMarkdown generates a markdown executive brief
func GenerateMarkdown(items *matrix.CategorizedItems, cfg *config.Config, startTime, endTime time.Time) string {
	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("# Executive Brief - %s\n\n", time.Now().Format("January 2, 2006")))

	// Summary
	total := len(items.Q1) + len(items.Q2) + len(items.Q3) + len(items.Q4)
	teammateCount := countTeammates(items)

	sb.WriteString("## Summary\n")
	sb.WriteString(fmt.Sprintf("- **Total items**: %d\n", total))
	sb.WriteString(fmt.Sprintf("- **Teammates needing help**: %d\n", teammateCount))
	sb.WriteString(fmt.Sprintf("- **Critical blockers**: %d\n", len(items.Q1)))
	sb.WriteString(fmt.Sprintf("- **Time range**: %s - %s\n\n",
		startTime.Format("Jan 2 15:04 MST"),
		endTime.Format("Jan 2 15:04 MST")))

	sb.WriteString("---\n\n")

	// Quadrant 1: Urgent & Important
	sb.WriteString("## 🔥 Quadrant 1: Do First (Urgent & Important)\n\n")
	if len(items.Q1) == 0 {
		sb.WriteString("*No urgent and important items*\n\n")
	} else {
		writeItems(&sb, items.Q1, cfg)
	}
	sb.WriteString("---\n\n")

	// Quadrant 2: Important, Not Urgent
	sb.WriteString("## ⭐ Quadrant 2: Schedule (Important, Not Urgent)\n\n")
	if len(items.Q2) == 0 {
		sb.WriteString("*No important items to schedule*\n\n")
	} else {
		writeItems(&sb, items.Q2, cfg)
	}
	sb.WriteString("---\n\n")

	// Quadrant 3: Urgent, Not Important
	sb.WriteString("## ⚡ Quadrant 3: Delegate (Urgent, Not Important)\n\n")
	if len(items.Q3) == 0 {
		sb.WriteString("*No urgent but unimportant items*\n\n")
	} else {
		writeItems(&sb, items.Q3, cfg)
	}
	sb.WriteString("---\n\n")

	// Quadrant 4: Neither
	sb.WriteString("## 📋 Quadrant 4: Review Later (Neither Urgent nor Important)\n\n")
	if len(items.Q4) == 0 {
		sb.WriteString("*No low priority items*\n\n")
	} else {
		writeItems(&sb, items.Q4, cfg)
	}
	sb.WriteString("---\n\n")

	// Recommended actions (top 5 from Q1)
	sb.WriteString("## 🎯 Recommended Actions\n\n")
	topActions := items.Q1
	if len(topActions) > 5 {
		topActions = topActions[:5]
	}
	if len(topActions) == 0 {
		sb.WriteString("*No urgent actions at this time*\n\n")
	} else {
		for i, item := range topActions {
			sb.WriteString(fmt.Sprintf("%d. **%s**: %s\n", i+1, item.ID, item.Title))
		}
		sb.WriteString("\n")
	}
	sb.WriteString("---\n\n")

	// Metrics
	sb.WriteString("## 📊 Metrics\n\n")
	sb.WriteString("### By Quadrant\n")
	sb.WriteString(fmt.Sprintf("- Q1 (Urgent & Important): %d items\n", len(items.Q1)))
	sb.WriteString(fmt.Sprintf("- Q2 (Important, Not Urgent): %d items\n", len(items.Q2)))
	sb.WriteString(fmt.Sprintf("- Q3 (Urgent, Not Important): %d items\n", len(items.Q3)))
	sb.WriteString(fmt.Sprintf("- Q4 (Neither): %d items\n\n", len(items.Q4)))

	// Source breakdown
	jiraCount, githubCount := countSources(items)
	sb.WriteString("### By Source\n")
	sb.WriteString(fmt.Sprintf("- Jira: %d items\n", jiraCount))
	sb.WriteString(fmt.Sprintf("- GitHub: %d items\n\n", githubCount))

	return sb.String()
}

// writeItems writes a list of items to the string builder
func writeItems(sb *strings.Builder, items []matrix.Item, cfg *config.Config) {
	// Separate teammate items from user items
	teammateItems := make([]matrix.Item, 0)
	userItems := make([]matrix.Item, 0)

	for _, item := range items {
		if len(item.TeammatesInvolved) > 0 {
			teammateItems = append(teammateItems, item)
		} else {
			userItems = append(userItems, item)
		}
	}

	// Sort by urgency/importance score
	sort.Slice(teammateItems, func(i, j int) bool {
		return teammateItems[i].UrgencyScore+teammateItems[i].ImportanceScore >
			teammateItems[j].UrgencyScore+teammateItems[j].ImportanceScore
	})
	sort.Slice(userItems, func(i, j int) bool {
		return userItems[i].UrgencyScore+userItems[i].ImportanceScore >
			userItems[j].UrgencyScore+userItems[j].ImportanceScore
	})

	// Write teammate items first
	if len(teammateItems) > 0 {
		sb.WriteString("### Teammate Items\n")
		for _, item := range teammateItems {
			writeItem(sb, item, cfg)
		}
		sb.WriteString("\n")
	}

	// Write user items
	if len(userItems) > 0 {
		sb.WriteString("### Your Items\n")
		for _, item := range userItems {
			writeItem(sb, item, cfg)
		}
		sb.WriteString("\n")
	}
}

// writeItem writes a single item
func writeItem(sb *strings.Builder, item matrix.Item, cfg *config.Config) {
	// Checkbox and title
	sb.WriteString(fmt.Sprintf("- [ ] **[%s] %s**", item.ID, item.Title))

	// Teammate indicator
	if len(item.TeammatesInvolved) > 0 {
		sb.WriteString(fmt.Sprintf(" - @%s", strings.Join(item.TeammatesInvolved, ", @")))
	}
	sb.WriteString("\n")

	// Metadata line
	metadata := make([]string, 0)
	// Capitalize first letter
	source := item.Source
	if len(source) > 0 {
		source = strings.ToUpper(source[:1]) + source[1:]
	}
	metadata = append(metadata, fmt.Sprintf("Source: %s", source))

	if item.Priority != "" {
		metadata = append(metadata, fmt.Sprintf("Priority: %s", item.Priority))
	}

	if item.Status != "" {
		metadata = append(metadata, fmt.Sprintf("Status: %s", item.Status))
	}

	// Time info
	timeAgo := formatTimeAgo(item.UpdatedAt)
	metadata = append(metadata, fmt.Sprintf("Updated: %s", timeAgo))

	if item.DueDate != nil {
		daysUntil := int(time.Until(*item.DueDate).Hours() / 24)
		if daysUntil < 0 {
			metadata = append(metadata, "📅 OVERDUE")
		} else if daysUntil == 0 {
			metadata = append(metadata, "📅 Due today")
		} else if daysUntil <= 2 {
			metadata = append(metadata, fmt.Sprintf("📅 Due in %d days", daysUntil))
		}
	}

	sb.WriteString(fmt.Sprintf("  - %s\n", strings.Join(metadata, " | ")))
	sb.WriteString(fmt.Sprintf("  - URL: %s\n", item.URL))

	// Labels/tags
	if len(item.Labels) > 0 {
		tags := make([]string, 0)
		for _, label := range item.Labels {
			emoji := getEmoji(label, cfg)
			if emoji != "" {
				tags = append(tags, fmt.Sprintf("%s %s", emoji, label))
			}
		}
		if len(tags) > 0 {
			sb.WriteString(fmt.Sprintf("  - %s\n", strings.Join(tags, " | ")))
		}
	}

	sb.WriteString("\n")
}

// formatTimeAgo formats a time as "2h ago", "3d ago", etc.
func formatTimeAgo(t time.Time) string {
	duration := time.Since(t)
	hours := int(duration.Hours())

	if hours < 1 {
		return fmt.Sprintf("%dm ago", int(duration.Minutes()))
	} else if hours < 24 {
		return fmt.Sprintf("%dh ago", hours)
	} else {
		days := hours / 24
		return fmt.Sprintf("%dd ago", days)
	}
}

// getEmoji returns an emoji for a label/keyword
func getEmoji(label string, cfg *config.Config) string {
	labelLower := strings.ToLower(label)

	for key, emoji := range cfg.Output.Emojis {
		if strings.Contains(labelLower, key) {
			return emoji
		}
	}

	return ""
}

// countTeammates counts unique teammates across all items
func countTeammates(items *matrix.CategorizedItems) int {
	teammates := make(map[string]bool)

	for _, item := range append(append(append(items.Q1, items.Q2...), items.Q3...), items.Q4...) {
		for _, tm := range item.TeammatesInvolved {
			teammates[tm] = true
		}
	}

	return len(teammates)
}

// countSources counts items by source
func countSources(items *matrix.CategorizedItems) (jira, github int) {
	for _, item := range append(append(append(items.Q1, items.Q2...), items.Q3...), items.Q4...) {
		switch item.Source {
		case "jira":
			jira++
		case "github":
			github++
		}
	}
	return
}
