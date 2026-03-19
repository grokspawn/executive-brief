# Executive Brief

A Go-based tool that generates daily briefings from Jira and GitHub, organized by the Eisenhower Matrix to help you prioritize teammate needs and your own work.

## What It Does

- 📊 **Aggregates** information from Jira and GitHub
- 🎯 **Organizes** items into the Eisenhower Matrix (Urgent/Important quadrants)
- 👥 **Prioritizes** teammate needs and blockers
- ⏰ **Tracks** multi-timezone activity with `--daily` flag
- 📈 **Generates** actionable briefs in markdown or HTML

## Quick Start

```bash
# 1. Build the binary
make build

# 2. Create configuration
cp examples/teammates.yaml.example teammates.yaml
# Edit with your teammate info

# 3. Run
./exec-brief --daily
```

## Example Output

```markdown
# Executive Brief - March 19, 2026

## Summary
- Total items: 15
- Teammates needing help: 3 (Alice, Bob, Charlie)
- Critical blockers: 2

## 🔥 Quadrant 1: Do First (Urgent & Important)

### Teammate Items
- [ ] [JIRA-123] Production API failing - @Alice Smith
  - Source: Jira | Priority: Blocker | Updated: 2h ago
  - 🔥 Critical | 👥 Teammate

- [ ] [PR#456] Security fix for CVE-2026-1234 - @Bob Jones
  - Source: GitHub | Waiting for review: 1d 3h
  - 🔒 Security | 👥 Teammate

### Your Items
- [ ] [JIRA-789] Release blocker - deployment fails
  - Source: Jira | Priority: Blocker | Due: Today

## 📅 Quadrant 2: Schedule (Important, Not Urgent)
...

## 🎯 Recommended Actions
1. Review PR#456 for @Bob (critical security fix)
2. Help @Alice with JIRA-123 (production blocker)
3. Address JIRA-789 release blocker
```

## Documentation

- **[examples/EXAMPLES.md](examples/EXAMPLES.md)** - Configuration examples
- **[examples/teammates.yaml.example](examples/teammates.yaml.example)** - Teammate config template
- **[examples/config.yaml.example](examples/config.yaml.example)** - Source config template

## Key Features

### 🌍 Multi-Timezone Support

The `--daily` flag considers both EST/EDT and your home timezone:

```bash
./exec-brief --daily
```

Running at 9 AM PST on March 19:
- Captures: March 18 00:00 EST → March 19 09:00 PST
- Why: Gets all East Coast business day + your morning
- Result: No activities missed across timezones

### 👥 Teammate-First Prioritization

Configure your teammates once:

```yaml
teammates:
  - name: Alice Smith
    github: asmith
    jira: alice.smith@company.com
    slack: asmith
    priority: high
```

The brief automatically:
- Identifies items involving teammates
- Shows teammate items first in each quadrant
- Highlights where they need help
- Recommends actions to unblock them

### 📊 Eisenhower Matrix

Items are automatically categorized:

| Quadrant | Description | Action |
|----------|-------------|--------|
| **Q1** 🔥 | Urgent & Important | Do First |
| **Q2** 📅 | Important, Not Urgent | Schedule |
| **Q3** 👥 | Urgent, Not Important | Delegate |
| **Q4** 📋 | Neither | Review Later |

Scoring considers:
- Keywords (blocker, critical, security)
- Due dates and deadlines
- Teammate involvement
- Priority fields
- CI/CD status

### 🔗 Multi-Source Integration

**Jira** (direct API):
- Custom JQL queries
- Issues assigned to you or your teammates
- Component/project tracking

**GitHub** (via `gh` CLI):
- PRs requesting your review
- Team member PRs
- Issues/PRs mentioning you

### 🎨 Flexible Configuration

**Minimal** (teammates.yaml):
```yaml
user:
  name: Your Name
  github: yourusername
  jira: you@company.com
  timezone: America/New_York

teammates:
  - name: Teammate
    github: teammate
    jira: teammate@company.com
    priority: high
```

**Advanced** (exec-brief.yaml):
```yaml
sources:
  jira:
    enabled: true
    jql_filters:
      - "priority = Blocker AND status != Closed"
      - "assignee = currentUser() AND updated >= -1d"
  github:
    enabled: true
```

See [examples/EXAMPLES.md](examples/EXAMPLES.md) for more configuration examples.

## Usage Examples

```bash
# Daily brief (yesterday to today, multi-timezone)
./exec-brief --daily

# Focus on teammates only
./exec-brief --daily --teammates-only

# Save to file
./exec-brief --daily --save team-brief.md

# Specific sources only
./exec-brief --daily --sources jira,github

# HTML output
./exec-brief --daily --format html --save brief.html

# Specific date
./exec-brief --date 2026-03-19

# Combine options
./exec-brief --daily --teammates-only --sources jira --save team-blockers.md
```

## Common Use Cases

**Daily Standup Prep:**
```bash
./exec-brief --daily --teammates-only
```

**Code Review Focus:**
```bash
./exec-brief --daily --sources github
```

**Team Status:**
```bash
./exec-brief --daily --save team-status.md
```

## File Structure

```
executive-brief/
├── README.md                  # This file
├── go.mod                     # Go module definition
├── main.go                    # Main entry point
├── Makefile                   # Build commands
├── internal/                  # Go packages
│   ├── config/                # Configuration loading
│   ├── jira/                  # Jira API client
│   ├── github/                # GitHub integration
│   ├── matrix/                # Eisenhower matrix logic
│   └── output/                # Output generation (markdown/HTML)
└── examples/                  # Configuration examples
    ├── EXAMPLES.md            # Configuration scenarios
    ├── teammates.yaml.example # Teammate config template
    └── config.yaml.example    # Source config template
```

## Prerequisites

**Required:**
- Go 1.21+ (for building)

**Optional (for data sources):**
- **Jira:** API token configured (see Jira client implementation)
- **GitHub:** `gh` CLI installed and authenticated

## Installation

```bash
# 1. Clone the repository
git clone <repo-url>
cd executive-brief

# 2. Build the binary
make build

# 3. Create your configuration
cp examples/teammates.yaml.example teammates.yaml
# Edit teammates.yaml with your info

# 4. (Optional) Create source configuration
cp examples/config.yaml.example exec-brief.yaml
# Edit exec-brief.yaml to customize sources

# 5. Run
./exec-brief --daily
```

## How It Works

1. **Time Range:** Calculates yesterday→today across timezones (EST/EDT + your timezone) to capture all activity
2. **Data Collection:** Queries Jira (via API) and GitHub (via `gh` CLI) for relevant items
3. **Teammate Matching:** Identifies items involving your configured teammates (assignee, author, reporter)
4. **Categorization:** Scores items for urgency and importance, assigns to Eisenhower quadrants (Q1-Q4)
5. **Output:** Generates organized brief with teammate items prioritized, in markdown or HTML

## Customization

Create `exec-brief.yaml` to customize sources and rules:

```yaml
sources:
  jira:
    enabled: true
    jql_filters:
      - "assignee = currentUser() AND updated >= -1d"
      - "priority = Blocker AND status != Closed"
      - "component = 'Networking' AND status = 'In Progress'"
  github:
    enabled: true

matrix_rules:
  urgent_keywords:
    - blocker
    - critical
    - p0
    - production
  important_keywords:
    - feature
    - security
    - teammate
    - customer
  time_based:
    urgent_within_days: 2
    important_within_days: 7
```

See [examples/EXAMPLES.md](examples/EXAMPLES.md) for more configuration patterns.

## Integration

### With Cron

Run automatically every morning:

```bash
# Add to crontab
0 9 * * * cd /path/to/executive-brief && ./exec-brief --daily --save
```

### With Other Tools

```bash
# Get brief, then review items
./exec-brief --daily > brief.md
gh pr list

# Email yourself the brief
./exec-brief --daily --save | mail -s "Daily Brief" you@company.com
```

## Troubleshooting

| Problem | Solution |
|---------|----------|
| "teammates.yaml not found" | Create in current directory |
| No items returned | Check Jira API token, `gh` auth, source enablement |
| Items in wrong quadrant | Customize `matrix_rules` in `exec-brief.yaml` |
| Teammates not identified | Verify identifiers match exactly (case-sensitive) |

## Future Enhancements

- Additional data sources (Slack, calendar, email)
- Jira board/dashboard support
- Google Docs integration
- AI-powered priority scoring
- Historical trending
- Team-wide aggregate views

## License

MIT
