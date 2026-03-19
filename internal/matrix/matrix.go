package matrix

import (
	"strings"
	"time"

	"github.com/grokspawn/executive-brief/internal/config"
)

// Item represents a work item from any source
type Item struct {
	ID                 string
	Title              string
	Source             string
	URL                string
	Type               string
	Status             string
	Priority           string
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DueDate            *time.Time
	Labels             []string
	Assignee           string
	AssigneeName       string
	Reporter           string
	Author             string
	TeammatesInvolved  []string
	UrgencyScore       int
	ImportanceScore    int
	Quadrant           int
}

// CategorizedItems represents items organized by quadrant
type CategorizedItems struct {
	Q1 []Item // Urgent & Important
	Q2 []Item // Important, Not Urgent
	Q3 []Item // Urgent, Not Important
	Q4 []Item // Neither
}

// Categorize categorizes items into the Eisenhower Matrix
func Categorize(items []Item, cfg *config.Config) *CategorizedItems {
	result := &CategorizedItems{
		Q1: make([]Item, 0),
		Q2: make([]Item, 0),
		Q3: make([]Item, 0),
		Q4: make([]Item, 0),
	}

	for _, item := range items {
		// Calculate scores
		item.UrgencyScore = calculateUrgency(item, cfg)
		item.ImportanceScore = calculateImportance(item, cfg)

		// Assign quadrant
		isUrgent := item.UrgencyScore >= 3
		isImportant := item.ImportanceScore >= 2

		if isUrgent && isImportant {
			item.Quadrant = 1
			result.Q1 = append(result.Q1, item)
		} else if !isUrgent && isImportant {
			item.Quadrant = 2
			result.Q2 = append(result.Q2, item)
		} else if isUrgent && !isImportant {
			item.Quadrant = 3
			result.Q3 = append(result.Q3, item)
		} else {
			item.Quadrant = 4
			result.Q4 = append(result.Q4, item)
		}
	}

	return result
}

// calculateUrgency calculates urgency score (0-10)
func calculateUrgency(item Item, cfg *config.Config) int {
	score := 0

	// Keyword matching
	titleLower := strings.ToLower(item.Title)
	labelsStr := strings.ToLower(strings.Join(item.Labels, " "))

	for _, keyword := range cfg.MatrixRules.UrgentKeywords {
		if strings.Contains(titleLower, keyword) || strings.Contains(labelsStr, keyword) {
			score += 3
			break
		}
	}

	// Priority field (Jira)
	priorityLower := strings.ToLower(item.Priority)
	if priorityLower == "blocker" || priorityLower == "critical" {
		score += 3
	} else if priorityLower == "major" {
		score += 2
	}

	// Due date
	if item.DueDate != nil {
		daysUntilDue := int(time.Until(*item.DueDate).Hours() / 24)
		if daysUntilDue <= 0 {
			score += 4 // Overdue
		} else if daysUntilDue <= 1 {
			score += 3 // Due within 24h
		} else if daysUntilDue <= 2 {
			score += 2 // Due within 48h
		}
	}

	// Teammate blocked
	if strings.Contains(labelsStr, "blocked") && len(item.TeammatesInvolved) > 0 {
		score += 2
	}

	// CI/CD failing
	if item.Source == "github" && strings.Contains(labelsStr, "ci") {
		if strings.Contains(labelsStr, "fail") {
			score += 2
		}
	}

	// Waiting for review
	if item.Source == "github" && strings.Contains(labelsStr, "review_requested") {
		ageHours := time.Since(item.CreatedAt).Hours()
		if ageHours > 24 {
			score += 1
		}
	}

	if score > 10 {
		score = 10
	}

	return score
}

// calculateImportance calculates importance score (0-10)
func calculateImportance(item Item, cfg *config.Config) int {
	score := 0

	// Teammates involved
	if len(item.TeammatesInvolved) > 0 {
		score += 3
	}

	// Important keywords
	titleLower := strings.ToLower(item.Title)
	labelsStr := strings.ToLower(strings.Join(item.Labels, " "))

	for _, keyword := range cfg.MatrixRules.ImportantKeywords {
		if strings.Contains(titleLower, keyword) || strings.Contains(labelsStr, keyword) {
			score += 2
			break
		}
	}

	// Priority field
	priorityLower := strings.ToLower(item.Priority)
	if priorityLower == "blocker" || priorityLower == "critical" || priorityLower == "major" {
		score += 2
	} else if priorityLower == "high" {
		score += 1
	}

	// Security issues
	if strings.Contains(labelsStr, "security") || strings.Contains(titleLower, "cve") {
		score += 2
	}

	if score > 10 {
		score = 10
	}

	return score
}
