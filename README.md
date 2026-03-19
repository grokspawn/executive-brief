# Executive Brief - Daily Assistant Skill for Claude Code

An intelligent executive assistant that generates daily briefings organized by the Eisenhower Matrix, helping you **prioritize teammate needs first** before your own work.

## What Is This?

Executive Brief is a Claude Code skill that:

- 📊 **Aggregates** information from Jira, GitHub, and Google Docs
- 🎯 **Organizes** items into the Eisenhower Matrix (Urgent/Important quadrants)
- 👥 **Prioritizes** teammate needs and blockers
- ⏰ **Tracks** multi-timezone activity with `--daily` flag
- 📈 **Generates** actionable daily briefs in markdown or HTML

Perfect for:
- Team leads tracking multiple team members
- Individual contributors helping teammates
- Managers overseeing multiple teams
- Anyone who wants to stay organized and teammate-focused

## Quick Start

```bash
# 1. Install dependencies
cd skill/
pip install -r requirements.txt

# 2. Create configuration
cp teammates.yaml.example teammates.yaml
# Edit with your info

# 3. Validate
./validate_config.py

# 4. Run in Claude Code
/exec-brief --daily
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

### Getting Started
- **[INSTALL.md](INSTALL.md)** - Installation and setup guide
- **[skill/README.md](skill/README.md)** - Full user documentation
- **[skill/QUICK_REFERENCE.md](skill/QUICK_REFERENCE.md)** - Command cheat sheet

### Configuration Help
- **[skill/EXAMPLES.md](skill/EXAMPLES.md)** - 16+ real-world configuration examples
- **[skill/JIRA_INTEGRATION.md](skill/JIRA_INTEGRATION.md)** - Complete Jira boards/dashboards guide
- **[skill/teammates.yaml.example](skill/teammates.yaml.example)** - Teammate config template
- **[skill/config.yaml.example](skill/config.yaml.example)** - Source config template

### Technical Details
- **[exec-brief.md](exec-brief.md)** - Design document and rationale
- **[skill/SKILL.md](skill/SKILL.md)** - Claude Code skill implementation spec

## Key Features

### 🌍 Multi-Timezone Support

The `--daily` flag considers both EST/EDT and your home timezone:

```bash
/exec-brief --daily
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

**Jira** (via MCP):
- Team sprint boards (Scrum/Kanban)
- Dashboards with metrics
- Custom JQL queries
- Component/project tracking

**GitHub** (via `gh` CLI):
- PRs requesting your review
- Team member PRs
- Issues/PRs mentioning you
- CI status checks

**Google Docs** (via MCP):
- Team documents
- Unresolved comments
- Recent updates

See [JIRA_INTEGRATION.md](skill/JIRA_INTEGRATION.md) for Jira-specific patterns.

### 🎨 Flexible Configuration

**Minimal** (just teammates):
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

**Advanced** (team boards, dashboards, custom queries):
```yaml
sources:
  jira:
    boards:
      - url: https://jira.company.com/secure/RapidBoard.jspa?rapidView=1234
        name: "Team Sprint"
    dashboards:
      - url: https://jira.company.com/secure/Dashboard.jspa?selectPageId=12345
        name: "Team Health"
    jql_filters:
      - "priority = Blocker AND status != Closed"
```

See [EXAMPLES.md](skill/EXAMPLES.md) for 16+ configuration examples.

## Usage Examples

```bash
# Daily brief (yesterday to today, multi-timezone)
/exec-brief --daily

# Focus on teammates only
/exec-brief --daily --teammates-only

# Save to file
/exec-brief --daily --save

# Specific sources only
/exec-brief --daily --sources jira,github

# HTML output
/exec-brief --daily --format html --save brief.html

# Specific date
/exec-brief --date 2026-03-19

# Combine options
/exec-brief --daily --teammates-only --sources jira --save team-blockers.md
```

See [QUICK_REFERENCE.md](skill/QUICK_REFERENCE.md) for all commands.

## Common Use Cases

### Daily Standup Prep
```bash
/exec-brief --daily --teammates-only --save standup.md
```
Shows what teammates need help with.

### Sprint Planning
```bash
/exec-brief --daily --sources jira --save sprint-planning.md
```
Focus on Jira items for planning.

### Code Review Focus
```bash
/exec-brief --daily --sources github
```
See PRs needing review from teammates.

### Manager Overview
```bash
/exec-brief --daily --save team-status.md
```
Full team status across all sources.

## Configuration Examples

Browse [EXAMPLES.md](skill/EXAMPLES.md) for scenarios like:

1. **Single Team with Sprint Board** - Basic Scrum team setup
2. **Multiple Teams** - Manager overseeing 3 teams
3. **Component Ownership** - Team responsible for component across projects
4. **Release Management** - Release tracking with blockers
5. **Multi-Region Team** - Distributed team across timezones
6. **Open Source Maintainer** - Multiple projects and repositories
7. **And 10 more...**

## File Structure

```
executive-brief/
├── README.md                  # This file - project overview
├── INSTALL.md                 # Installation guide
├── exec-brief.md              # Design document
└── skill/                     # Skill implementation
    ├── SKILL.md               # Claude Code skill spec
    ├── README.md              # User documentation
    ├── QUICK_REFERENCE.md     # Command cheat sheet
    ├── EXAMPLES.md            # 16+ configuration examples
    ├── JIRA_INTEGRATION.md    # Jira boards/dashboards guide
    ├── exec_brief.py          # Core Python implementation
    ├── validate_config.py     # Configuration validator
    ├── requirements.txt       # Python dependencies
    ├── teammates.yaml.example # Example teammate config
    └── config.yaml.example    # Example source config
```

## Prerequisites

### Required
- Python 3.8+
- `teammates.yaml` configuration

### Optional (for data sources)
- Jira: MCP server (`mcp__atlassian__*`)
- GitHub: `gh` CLI authenticated
- Google Docs: MCP server (`mcp__google-docs__*`)

Install Python dependencies:
```bash
pip install -r skill/requirements.txt
```

## Installation

See [INSTALL.md](INSTALL.md) for detailed setup instructions.

Quick version:
1. Clone/download this repository
2. Install Python dependencies
3. Copy and edit `teammates.yaml`
4. Optionally copy and edit `.exec-brief.yaml`
5. Validate with `./validate_config.py`
6. Run `/exec-brief --daily` in Claude Code

## How It Works

### 1. Time Range Calculation
- Uses `--daily` to calculate yesterday→today in both EST/EDT and your timezone
- Ensures no activities missed across timezone boundaries

### 2. Data Collection
- Queries Jira boards, dashboards, and custom JQL
- Queries GitHub for PRs, issues, mentions
- Queries Google Docs for recent changes

### 3. Teammate Identification
- Matches assignees, authors, reviewers against teammate config
- Tags items with teammate involvement

### 4. Categorization
- Calculates urgency score (keywords, due dates, priority)
- Calculates importance score (teammates, security, impact)
- Assigns to Eisenhower quadrant (Q1-Q4)

### 5. Output Generation
- Organizes by quadrant
- Shows teammate items first
- Generates markdown or HTML
- Includes summary and recommended actions

See [SKILL.md](skill/SKILL.md) for implementation details.

## Customization

### Customize Matrix Rules

Edit `.exec-brief.yaml`:

```yaml
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

### Add Team Jira Boards

```yaml
sources:
  jira:
    boards:
      - url: https://jira.company.com/secure/RapidBoard.jspa?rapidView=1234
        name: "Team Sprint Board"
```

See [JIRA_INTEGRATION.md](skill/JIRA_INTEGRATION.md) for comprehensive Jira setup.

### Custom JQL Queries

```yaml
sources:
  jira:
    jql_filters:
      - "assignee = currentUser() AND updated >= -1d"
      - "priority = Blocker AND status != Closed"
      - "component = 'Networking' AND status = 'In Progress'"
```

See [EXAMPLES.md](skill/EXAMPLES.md) for more patterns.

## Integration

### With Claude Code Hooks

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
# Get brief, then dive deeper
/exec-brief --daily
/jira:status-rollup JIRA-123

# GitHub focus
/exec-brief --daily --sources github
/utils:gh-attention
```

## Troubleshooting

### Common Issues

| Problem | Solution |
|---------|----------|
| "teammates.yaml not found" | Create in current dir or `~/.config/exec-brief/` |
| No items returned | Check MCP connections, `gh` auth, source enablement |
| Items in wrong quadrant | Customize `matrix_rules` in `.exec-brief.yaml` |
| Teammates not identified | Verify identifiers match exactly (case-sensitive) |

### Validation

```bash
cd skill/
./validate_config.py
```

Checks for:
- Required fields
- Valid YAML syntax
- Proper identifier formats
- Configuration errors

See [INSTALL.md](INSTALL.md) for detailed troubleshooting.

## Contributing

Contributions welcome! Areas for enhancement:

- Additional data sources (Slack, email, calendar)
- AI-powered priority recommendations
- Historical trending and analytics
- Mobile-friendly HTML templates
- Team-wide aggregate briefs

## Future Enhancements

- [ ] Slack integration for notifications
- [ ] Email digest support
- [ ] Calendar integration for meeting prep
- [ ] AI-powered recommendations
- [ ] Historical analysis and trending
- [ ] Customizable output templates
- [ ] Team-wide aggregate views
- [ ] Mobile app integration

## License

MIT

## Support

- **Questions?** Check [QUICK_REFERENCE.md](skill/QUICK_REFERENCE.md)
- **Need examples?** See [EXAMPLES.md](skill/EXAMPLES.md)
- **Jira setup?** Read [JIRA_INTEGRATION.md](skill/JIRA_INTEGRATION.md)
- **Installation issues?** Review [INSTALL.md](INSTALL.md)

---

**Built with ❤️ for teams that prioritize helping each other first.**
