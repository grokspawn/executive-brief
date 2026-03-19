# Executive Brief Skill

An executive assistant skill for Claude Code that generates daily briefings with information organized into the Eisenhower Matrix, prioritizing teammate needs first.

## Overview

This skill aggregates information from multiple sources (Jira, GitHub, Google Docs) and organizes items into four quadrants based on urgency and importance:

1. **Quadrant 1: Do First** (Urgent & Important) - Critical items requiring immediate attention
2. **Quadrant 2: Schedule** (Important, Not Urgent) - Strategic work to plan for
3. **Quadrant 3: Delegate** (Urgent, Not Important) - Quick tasks that could be delegated
4. **Quadrant 4: Review Later** (Neither) - Low-priority items to review when time permits

The skill **prioritizes teammate needs first**, helping you identify where your colleagues need help before focusing on your own work.

## 📚 Documentation

- **[QUICK_REFERENCE.md](QUICK_REFERENCE.md)** - Command cheat sheet and quick lookups
- **[EXAMPLES.md](EXAMPLES.md)** - 16+ real-world configuration examples
- **[JIRA_INTEGRATION.md](JIRA_INTEGRATION.md)** - Complete guide to Jira boards, dashboards, and JQL
- **[INSTALL.md](../INSTALL.md)** - Installation and setup instructions
- **[SKILL.md](SKILL.md)** - Technical implementation details for Claude Code

## Quick Start

1. **Copy the example configuration**:
   ```bash
   cp teammates.yaml.example teammates.yaml
   ```

2. **Edit `teammates.yaml`** with your information and your teammates' identifiers

3. **Run the skill**:
   ```
   /exec-brief --daily
   ```

**New to this?** Check [QUICK_REFERENCE.md](QUICK_REFERENCE.md) for a command cheat sheet, or browse [EXAMPLES.md](EXAMPLES.md) to find a configuration similar to your setup.

## Usage

### Basic Commands

```bash
# Daily brief (yesterday to today, multi-timezone aware)
/exec-brief --daily

# Brief for a specific date
/exec-brief --date 2026-03-19

# Show only items involving teammates
/exec-brief --daily --teammates-only

# Save to file
/exec-brief --daily --save

# Save to specific file
/exec-brief --daily --save my-brief.md

# Query specific sources only
/exec-brief --daily --sources jira,github
```

### Advanced Usage

```bash
# HTML output (visual matrix layout)
/exec-brief --daily --format html --save brief.html

# Combine multiple options
/exec-brief --daily --teammates-only --sources jira --save team-blockers.md
```

## Configuration

**Need examples?** See **[EXAMPLES.md](EXAMPLES.md)** for 16+ real-world scenarios including:
- Single team with sprint board
- Multiple teams (manager view)
- Release tracking
- Component ownership
- Multi-region teams
- And more...

**Working with Jira?** See **[JIRA_INTEGRATION.md](JIRA_INTEGRATION.md)** for a complete guide to:
- Integrating team boards (Scrum/Kanban)
- Using dashboards for metrics
- Writing custom JQL queries
- Team-oriented patterns

### Required: teammates.yaml

Defines you and your teammates with cross-platform identifiers.

**Location**: Current directory or `~/.config/exec-brief/teammates.yaml`

See `teammates.yaml.example` for a complete template.

**Minimal example**:
```yaml
user:
  name: Your Name
  github: yourusername
  jira: you@company.com
  timezone: America/New_York

teammates:
  - name: Teammate Name
    github: teammate
    jira: teammate@company.com
    priority: high
```

### Optional: exec-brief.yaml

Configures information sources and matrix rules.

**Location**: Current directory or `~/.config/exec-brief/config.yaml`

See `config.yaml.example` for all options.

**The skill works with sensible defaults if this file is not provided.**

## How It Works

### 1. Time Zone Handling

When you use `--daily`, the skill considers both:
- **EST/EDT** (America/New_York) - Where many systems timestamp activities
- **Your home timezone** (from `teammates.yaml`)

This ensures you capture all activity across timezone boundaries.

**Example**: Running at 9:00 AM PST on March 19:
- Captures: March 18 00:00 EST → March 19 09:00 PST
- Why: Catches all East Coast business day activity plus your current morning

### 2. Data Collection

The skill queries configured sources:

**Jira** (via MCP):
- Issues assigned to you
- Issues reported by you
- Issues you're watching
- Custom JQL queries (if configured)
- Dashboard/board items (if configured)

**GitHub** (via `gh` CLI):
- PRs requesting your review
- PRs from teammates
- Issues/PRs mentioning you
- Your open PRs
- CI status checks

**Google Docs** (via MCP):
- Recent documents in team folders
- Documents with unresolved comments
- Documents you have access to

### 3. Teammate Identification

For each item, the skill identifies if teammates are involved by matching:
- Assignees, reporters, authors
- @mentions in descriptions/comments
- Reviewers on PRs
- Watchers on issues

Items involving teammates are **prioritized and highlighted**.

### 4. Categorization

Each item receives:
- **Urgency Score** (0-10) based on:
  - Keywords (blocker, critical, urgent, emergency)
  - Priority field
  - Due date proximity
  - Teammates blocked
  - CI/CD failures

- **Importance Score** (0-10) based on:
  - Teammate involvement (+3)
  - Security/strategic keywords
  - Priority field
  - Number of people affected

Items are assigned to quadrants using thresholds:
- Urgent: score ≥ 3
- Important: score ≥ 2

### 5. Output Generation

The brief includes:
- **Summary** with key metrics
- **Four quadrants** with items organized by priority
- **Teammate items** shown separately and first
- **Recommended actions** (top 5 from Q1)
- **Metrics** by source, quadrant, and teammate involvement

## Customization

### Custom JQL Filters

Add to `exec-brief.yaml`:

```yaml
sources:
  jira:
    jql_filters:
      - "project = MYPROJ AND status = 'In Progress'"
      - "component = 'My Component' AND priority = Blocker"
```

### Team Dashboards

Track Jira dashboards/boards:

```yaml
sources:
  jira:
    dashboards:
      - url: https://jira.example.com/secure/Dashboard.jspa?selectPageId=12345
        name: "Team Dashboard"

    boards:
      - url: https://jira.example.com/secure/RapidBoard.jspa?rapidView=1234
        name: "Sprint Board"
```

### Matrix Rules

Customize urgency/importance detection:

```yaml
matrix_rules:
  urgent_keywords:
    - production
    - outage
    - p0

  important_keywords:
    - okr
    - strategic
    - customer
```

## Output Format

### Markdown (default)

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

### Your Items
- [ ] [JIRA-789] Release blocker - deployment fails
  - Source: Jira | Priority: Blocker | Due: Today

...
```

### HTML Output

With `--format html`, generates a visual 2x2 matrix with:
- Color-coded quadrants
- Interactive checkboxes
- Clickable links
- Teammate badges
- Responsive design

## Integration

### Daily Morning Routine

Add to Claude Code hooks:

```json
{
  "hooks": {
    "on_session_start": {
      "command": "/exec-brief --daily --teammates-only --save",
      "enabled": true
    }
  }
}
```

### With Other Skills

Combine with existing skills:

```bash
# Deep dive into Jira items from brief
/exec-brief --daily --sources jira
/jira:status-rollup JIRA-123

# Check GitHub PRs from brief
/exec-brief --daily --sources github
/utils:gh-attention

# Get repository context
/git:summary
/exec-brief --daily
```

## Troubleshooting

### "teammates.yaml not found"

Create the file in the current directory or `~/.config/exec-brief/`:

```bash
mkdir -p ~/.config/exec-brief
cp teammates.yaml.example ~/.config/exec-brief/teammates.yaml
# Edit with your information
```

### No items returned

1. Check if sources are enabled in `exec-brief.yaml`
2. Verify MCP servers are connected (Jira, Google Docs)
3. Verify `gh` CLI is authenticated for GitHub
4. Try a wider date range: `--date 2026-03-01`

### Items in wrong quadrant

1. Enable score display: `--show-scores` (if implemented)
2. Check keyword matching in titles/labels
3. Customize rules in `exec-brief.yaml`

### Teammates not identified

1. Verify identifiers in `teammates.yaml` match exactly
2. Check for typos in usernames/emails
3. Some systems use different formats (e.g., email vs username)

## Examples

### Morning Brief for Team Leads

```bash
/exec-brief --daily --teammates-only --save team-brief.md
```

Shows only items where teammates need help, saved to file for review.

### Weekly Planning

```bash
/exec-brief --date 2026-03-17 --save monday.md
/exec-brief --date 2026-03-18 --save tuesday.md
...
```

Generate briefs for each day to review the week.

### Focus on Critical Items

```bash
/exec-brief --daily --sources jira --teammates-only
```

See only Jira blockers/critical items involving teammates.

### Full Overview

```bash
/exec-brief --daily --save --format html
```

Complete brief across all sources, saved as interactive HTML.

## Requirements

- **Python 3.8+** with `pyyaml`, `pytz`
- **MCP Servers**: Jira (mcp__atlassian), Google Docs (mcp__google-docs)
- **GitHub CLI**: `gh` authenticated with your account
- **Configuration**: `teammates.yaml` in current directory or `~/.config/exec-brief/`

## Future Enhancements

- [ ] Slack integration for team notifications
- [ ] Email digest support
- [ ] Calendar integration for meeting prep
- [ ] AI-powered priority recommendations
- [ ] Historical trending and analytics
- [ ] Mobile app integration
- [ ] Customizable templates
- [ ] Team-wide aggregate briefs

## Contributing

This skill is designed to be customized for your workflow. Feel free to:
- Add new data sources
- Customize categorization logic
- Create new output formats
- Share your configurations

## License

MIT
