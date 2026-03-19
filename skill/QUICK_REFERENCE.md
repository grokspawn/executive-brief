# Executive Brief - Quick Reference

## Command Cheat Sheet

```bash
# Daily brief (yesterday to today, multi-timezone)
/exec-brief --daily

# Specific date
/exec-brief --date 2026-03-19

# Only teammate items
/exec-brief --daily --teammates-only

# Save to file
/exec-brief --daily --save
/exec-brief --daily --save my-brief.md

# Specific sources
/exec-brief --daily --sources jira
/exec-brief --daily --sources jira,github
/exec-brief --daily --sources github

# HTML output
/exec-brief --daily --format html --save brief.html

# Combine options
/exec-brief --daily --teammates-only --sources jira --save team-jira.md
```

## Common Configuration Patterns

### Minimal Setup (No exec-brief.yaml)

Just create `teammates.yaml`:

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

Run: `/exec-brief --daily`

### Team Sprint Board Only

**exec-brief.yaml:**
```yaml
sources:
  jira:
    enabled: true
    server: https://jira.company.com
    boards:
      - url: https://jira.company.com/secure/RapidBoard.jspa?rapidView=1234
        name: "Team Sprint"
```

### Team Dashboard Only

**exec-brief.yaml:**
```yaml
sources:
  jira:
    enabled: true
    server: https://jira.company.com
    dashboards:
      - url: https://jira.company.com/secure/Dashboard.jspa?selectPageId=12345
        name: "Team Dashboard"
```

### Custom JQL Filters

**exec-brief.yaml:**
```yaml
sources:
  jira:
    enabled: true
    server: https://jira.company.com
    jql_filters:
      - "assignee = currentUser() AND updated >= -1d"
      - "priority = Blocker AND status != Closed"
      - "component = 'MyComponent' AND status = 'In Progress'"
```

### Multiple Boards (Multi-Team)

**exec-brief.yaml:**
```yaml
sources:
  jira:
    enabled: true
    server: https://jira.company.com
    boards:
      - url: https://jira.company.com/secure/RapidBoard.jspa?rapidView=100
        name: "Team A"
      - url: https://jira.company.com/secure/RapidBoard.jspa?rapidView=101
        name: "Team B"
      - url: https://jira.company.com/secure/RapidBoard.jspa?rapidView=102
        name: "Team C"
```

### GitHub Focus

**exec-brief.yaml:**
```yaml
sources:
  jira:
    enabled: false

  github:
    enabled: true
    organizations:
      - mycompany
      - opensource-org
    filters:
      review_requested: true
      mentioned: true
      team_prs: true
```

## Slack Identifier Formats

All of these are valid:

```yaml
# Simple string (just handle)
slack: jsmith

# Object with handle only
slack:
  handle: jsmith

# Object with UID only
slack:
  uid: U12345678

# Object with both
slack:
  uid: U12345678
  handle: jsmith
```

## Finding Jira URLs

### For Boards
1. Go to your board in Jira
2. Copy URL from browser
3. Should look like: `https://jira.company.com/secure/RapidBoard.jspa?rapidView=1234`
4. Paste entire URL into config

### For Dashboards
1. Go to your dashboard in Jira
2. Copy URL from browser
3. Should look like: `https://jira.company.com/secure/Dashboard.jspa?selectPageId=12345`
4. Paste entire URL into config

## Eisenhower Matrix Quadrants

| Quadrant | Urgency | Importance | Action |
|----------|---------|------------|--------|
| Q1 | ✅ Urgent | ✅ Important | **Do First** - Critical items |
| Q2 | ❌ Not Urgent | ✅ Important | **Schedule** - Plan these |
| Q3 | ✅ Urgent | ❌ Not Important | **Delegate** - Quick tasks |
| Q4 | ❌ Not Urgent | ❌ Not Important | **Review Later** - Low priority |

## Scoring Reference

### Urgency (+3 each)
- Keywords: blocker, critical, urgent, emergency, production
- Priority: Blocker/Critical
- Due: Overdue or < 24h
- Teammate blocked
- CI failing

### Importance (+3 for teammates)
- Teammate involved
- Security/strategic keywords
- High priority
- Multiple people affected

### Thresholds
- **Urgent**: Score ≥ 3
- **Important**: Score ≥ 2

## Time Zone Behavior

`--daily` flag considers **both**:
- **EST/EDT** (America/New_York) - system timestamps
- **Your timezone** (from teammates.yaml)

**Example** at 9 AM PST on March 19:
- Captures: March 18 00:00 EST → March 19 09:00 PST
- Why: Gets all East Coast activity + your morning

## Validation

```bash
# Validate configuration
cd skill/
./validate_config.py

# Expected output
✅ Configuration is valid!
```

## Troubleshooting

| Problem | Solution |
|---------|----------|
| "teammates.yaml not found" | Create in current dir or `~/.config/exec-brief/` |
| No items returned | Check source enablement, MCP connections, `gh` auth |
| Items in wrong quadrant | Customize `matrix_rules` in `exec-brief.yaml` |
| Teammates not identified | Verify identifiers match exactly (case-sensitive) |

## File Locations

Priority order for finding configs:

### teammates.yaml
1. `./teammates.yaml` (current directory)
2. `~/.config/exec-brief/teammates.yaml`

### exec-brief.yaml
1. `./exec-brief.yaml` (current directory)
2. `~/.config/exec-brief/config.yaml`

## Common Use Cases

### Daily Standup Prep
```bash
/exec-brief --daily --teammates-only --save standup.md
```

### Sprint Planning
```bash
/exec-brief --daily --sources jira --save sprint-planning.md
```

### Code Review Focus
```bash
/exec-brief --daily --sources github
```

### Manager Overview
```bash
/exec-brief --daily --save team-status.md
```

### Weekly Summary
```bash
/exec-brief --date 2026-03-17 --save monday.md
```

## Prerequisites

### Required
- ✅ Python 3.8+
- ✅ `teammates.yaml` configuration

### Optional (for data sources)
- Jira: MCP server (`mcp__atlassian__*`)
- GitHub: `gh` CLI authenticated
- Google Docs: MCP server (`mcp__google-docs__*`)

### Install Python Deps
```bash
pip install pyyaml pytz
# or
pip install -r requirements.txt
```

## Integration Examples

### Claude Code Hooks

Run automatically on session start:

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

```bash
# Get brief, then dive into Jira
/exec-brief --daily
/jira:status-rollup JIRA-123

# Get brief, then check GitHub
/exec-brief --daily
/utils:gh-attention

# Get repository context first
/git:summary
/exec-brief --daily
```

## Matrix Rules Customization

Override urgency/importance keywords:

**exec-brief.yaml:**
```yaml
matrix_rules:
  urgent_keywords:
    - blocker
    - critical
    - p0
    - production
    - outage

  important_keywords:
    - feature
    - security
    - okr
    - strategic
    - customer

  time_based:
    urgent_within_days: 2
    important_within_days: 7
```

## Full Configuration Template

Comprehensive example with all options:

**exec-brief.yaml:**
```yaml
sources:
  jira:
    enabled: true
    server: https://jira.company.com
    projects:
      - PROJ1
      - PROJ2
    jql_filters:
      - "assignee = currentUser() AND updated >= -1d"
    dashboards:
      - url: https://jira.company.com/secure/Dashboard.jspa?selectPageId=12345
        name: "Team Dashboard"
    boards:
      - url: https://jira.company.com/secure/RapidBoard.jspa?rapidView=1234
        name: "Sprint Board"

  github:
    enabled: true
    organizations:
      - myorg
    repositories:
      - myorg/repo1
    filters:
      review_requested: true
      mentioned: true
      team_prs: true
      assigned: true
      authored: true
    check_ci: true

  google_docs:
    enabled: false
    folders:
      - "Team Docs"
    track_comments: true

matrix_rules:
  urgent_keywords:
    - blocker
    - critical
  important_keywords:
    - feature
    - security
  time_based:
    urgent_within_days: 2
    important_within_days: 7
```

## Getting Help

1. **Check examples**: `EXAMPLES.md` has 16+ real-world scenarios
2. **Validate config**: Run `./validate_config.py`
3. **Read full docs**: See `README.md` for detailed documentation
4. **Review skill**: Check `SKILL.md` for implementation details

## Quick Links

- Examples: `EXAMPLES.md`
- Full docs: `README.md`
- Installation: `INSTALL.md`
- Skill spec: `SKILL.md`
- Design doc: `../exec-brief.md`
