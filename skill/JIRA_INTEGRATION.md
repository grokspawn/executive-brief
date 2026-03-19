# Jira Integration Guide for Executive Brief

This guide explains how to integrate Jira boards, dashboards, and custom queries into the Executive Brief skill.

## Understanding Jira Information Sources

The skill supports three types of Jira information sources:

### 1. 🎯 Agile Boards
**What**: Scrum/Kanban boards that track team workflow

**When to use**:
- Your team uses sprints or kanban
- You want to track current sprint work
- You need backlog visibility
- You're monitoring team progress

**Example boards**:
- Sprint planning boards
- Kanban flow boards
- Team backlogs
- Release boards

### 2. 📊 Dashboards
**What**: Custom dashboards with gadgets/widgets

**When to use**:
- You have a team dashboard with metrics
- You track blockers across multiple projects
- You need aggregate views
- You want release or sprint health monitoring

**Example dashboards**:
- Team weekly status
- Release blockers
- Sprint burndown
- Cross-team dependencies

### 3. 🔍 Custom JQL Filters
**What**: Jira Query Language for specific searches

**When to use**:
- Boards/dashboards don't capture what you need
- You need component-specific queries
- You want precise filtering
- You're tracking specific issue types

**Example queries**:
- All blockers in your component
- Issues assigned to teammates
- Customer escalations
- Security vulnerabilities

---

## How Boards Work

### Board URL Format

```
https://jira.company.com/secure/RapidBoard.jspa?rapidView=1234
                                                  └─────┬────┘
                                                   Board ID
```

### What the Skill Captures from Boards

1. **Active Sprint Issues** (for Scrum boards)
   - All issues in the current sprint
   - Issue status, assignee, priority
   - Sprint goal and timeline

2. **Backlog Items** (if configured)
   - Ready for development
   - Prioritized backlog
   - Estimated items

3. **Board Filters**
   - Quick filters (if in URL)
   - Saved filters
   - Board configuration

### Configuration Example

```yaml
sources:
  jira:
    enabled: true
    server: https://jira.company.com

    boards:
      # Scrum board - captures current sprint
      - url: https://jira.company.com/secure/RapidBoard.jspa?rapidView=1234
        name: "Platform Team Sprint 47"
        description: "Two-week sprint ending March 28"

      # Kanban board - captures all WIP
      - url: https://jira.company.com/secure/RapidBoard.jspa?rapidView=5678
        name: "Support Queue"
        description: "Customer support tickets"

      # Board with quick filter applied
      - url: https://jira.company.com/secure/RapidBoard.jspa?rapidView=9012&quickFilter=10001
        name: "Critical Issues Only"
        description: "Board filtered to show only critical priority"
```

### Finding Your Board URL

#### Method 1: Navigate to Board
1. Go to **Boards** menu in Jira
2. Click on your team's board
3. Copy the URL from browser address bar
4. Should look like: `https://jira.company.com/secure/RapidBoard.jspa?rapidView=1234`

#### Method 2: From Board Settings
1. Open your board
2. Click **Board** → **Configure**
3. Note the board ID in the URL
4. Construct URL: `https://jira.company.com/secure/RapidBoard.jspa?rapidView=YOUR_ID`

---

## How Dashboards Work

### Dashboard URL Format

```
https://jira.company.com/secure/Dashboard.jspa?selectPageId=12345
                                                └──────┬──────┘
                                                  Dashboard ID
```

### What the Skill Captures from Dashboards

Dashboards contain **gadgets** (widgets) that show different data:

1. **Filter Results Gadget**
   - Issues from saved filters
   - Captures the issues shown

2. **Sprint Burndown**
   - Current sprint progress
   - Remaining work

3. **Issue Statistics**
   - Count by status, priority, assignee
   - Identifies high-volume areas

4. **Activity Stream**
   - Recent updates
   - Comments and transitions

5. **Created vs Resolved**
   - Velocity metrics
   - Trend data

### Configuration Example

```yaml
sources:
  jira:
    enabled: true
    server: https://jira.example.com

    dashboards:
      # Team health dashboard
      - url: https://jira.example.com/secure/Dashboard.jspa?selectPageId=12345
        name: "Engineering Team Health"
        description: "Sprint progress, blockers, and team capacity"
        # This dashboard might have gadgets for:
        # - Current sprint burndown
        # - Blocker issues
        # - Team workload
        # - Recent activity

      # Release tracking dashboard
      - url: https://jira.example.com/secure/Dashboard.jspa?selectPageId=67890
        name: "4.16 Release Dashboard"
        description: "Release blockers and feature completion"
        # This dashboard might track:
        # - Release blocker count
        # - Feature completion status
        # - Documentation readiness
        # - Test coverage

      # Manager overview dashboard
      - url: https://jira.example.com/secure/Dashboard.jspa?selectPageId=11111
        name: "Multi-Team Overview"
        description: "Aggregate view across all teams"
        # Shows:
        # - Issues by team
        # - Cross-team blockers
        # - High-priority items
        # - SLA compliance
```

### Finding Your Dashboard URL

#### Method 1: From Dashboard Menu
1. Click **Dashboards** in top menu
2. Click **View all dashboards**
3. Click on your dashboard name
4. Copy URL from browser
5. Should look like: `https://jira.company.com/secure/Dashboard.jspa?selectPageId=12345`

#### Method 2: From Dashboard Favorites
1. Click star icon in top-right of dashboard
2. Right-click on dashboard in **Starred** menu
3. Copy link address

---

## How Custom JQL Works

### JQL Filter Format

```yaml
jql_filters:
  - "assignee = currentUser() AND updated >= -1d"
  - "priority in (Blocker, Critical) AND status != Closed"
  - "project = PROJ AND component = 'Networking'"
```

### Common JQL Patterns

#### By Assignment
```jql
# Your issues
assignee = currentUser()

# Team member issues
assignee in (user1, user2, user3)

# Issues assigned to your team
assignee in membersOf("team-name")

# Unassigned issues
assignee is EMPTY
```

#### By Time
```jql
# Updated in last 24 hours
updated >= -1d

# Created this week
created >= startOfWeek()

# Due soon
duedate <= 7d

# Overdue
duedate < now() AND status != Closed
```

#### By Priority/Status
```jql
# Blockers only
priority = Blocker

# Critical or blocker
priority in (Blocker, Critical)

# In progress
status = "In Progress"

# Not closed
status != Closed
```

#### By Project/Component
```jql
# Specific project
project = OCPBUGS

# Multiple projects
project in (OCPBUGS, CNTRLPLANE)

# By component
component = "Networking"

# Multiple components
component in ("Networking", "Storage")
```

#### By Sprint
```jql
# Current sprint
Sprint in openSprints()

# Specific sprint
Sprint = "Sprint 47"

# Not in any sprint
Sprint is EMPTY
```

#### By Labels/Custom Fields
```jql
# Has label
labels = "tech-debt"

# Customer escalation
labels = "escalation"

# Has target version
"Target Version" is not EMPTY

# Security issues
"Security Level" is not EMPTY
```

#### Combined Queries
```jql
# Team blockers updated recently
project = OCPBUGS
  AND assignee in membersOf("platform-team")
  AND priority = Blocker
  AND status != Closed
  AND updated >= -7d
ORDER BY updated DESC

# Carryover risk in current sprint
Sprint in openSprints()
  AND status = "To Do"
  AND sprint not in (futureSprints())

# Customer escalations needing attention
labels = "customer-escalation"
  AND status in ("New", "Triaged")
  AND assignee = currentUser()
  AND created >= -30d
ORDER BY priority DESC, created ASC
```

### Configuration Example

```yaml
sources:
  jira:
    enabled: true
    server: https://jira.company.com

    jql_filters:
      # Your active work
      - "assignee = currentUser() AND status in ('In Progress', 'Code Review') ORDER BY priority DESC"

      # Team blockers
      - "assignee in membersOf('platform-team') AND priority = Blocker AND status != Closed"

      # Issues needing review
      - "status = 'Code Review' AND reviewer = currentUser()"

      # Carryover risk
      - "Sprint in openSprints() AND status = 'To Do' AND duedate <= 3d"

      # Customer escalations
      - "labels = 'escalation' AND created >= -7d AND status != Resolved"

      # Component-specific
      - "component = 'Networking' AND priority in (Blocker, Critical) AND updated >= -1d"

      # Security vulnerabilities
      - "project = SECURITY AND 'CVE ID' is not EMPTY AND 'Fix Version' is EMPTY"
```

---

## Team-Oriented Patterns

### Pattern 1: Team Sprint Board + Team Dashboard

**Use case**: Scrum team with 2-week sprints

```yaml
sources:
  jira:
    enabled: true
    server: https://jira.company.com

    # Sprint work tracking
    boards:
      - url: https://jira.company.com/secure/RapidBoard.jspa?rapidView=1234
        name: "Platform Team Sprint Board"

    # Sprint health metrics
    dashboards:
      - url: https://jira.company.com/secure/Dashboard.jspa?selectPageId=12345
        name: "Sprint Health Dashboard"
        description: "Burndown, velocity, and blockers"

    # Additional sprint-specific queries
    jql_filters:
      # Carryover risk
      - "Sprint in openSprints() AND status = 'To Do' AND remainingEstimate > 8h"
```

### Pattern 2: Multi-Team Manager View

**Use case**: Manager overseeing 3 teams

```yaml
sources:
  jira:
    enabled: true
    server: https://jira.company.com

    # Each team's board
    boards:
      - url: https://jira.company.com/secure/RapidBoard.jspa?rapidView=100
        name: "Frontend Team"
      - url: https://jira.company.com/secure/RapidBoard.jspa?rapidView=101
        name: "Backend Team"
      - url: https://jira.company.com/secure/RapidBoard.jspa?rapidView=102
        name: "Infrastructure Team"

    # Aggregate dashboards
    dashboards:
      - url: https://jira.company.com/secure/Dashboard.jspa?selectPageId=20000
        name: "Engineering Overview"
        description: "All teams combined"

      - url: https://jira.company.com/secure/Dashboard.jspa?selectPageId=20001
        name: "Cross-Team Blockers"
        description: "Dependencies and blockers"

    # Manager-specific queries
    jql_filters:
      # All team blockers
      - "assignee in membersOf('frontend-team', 'backend-team', 'infra-team') AND priority = Blocker"

      # Cross-team dependencies
      - "labels = 'cross-team' AND status in ('Blocked', 'Waiting')"
```

### Pattern 3: Component Ownership

**Use case**: Team owns a component across multiple projects

```yaml
sources:
  jira:
    enabled: true
    server: https://jira.example.com

    # Component-filtered board
    boards:
      - url: https://jira.example.com/secure/RapidBoard.jspa?rapidView=7777&quickFilter=10001
        name: "Networking - All Issues"
        description: "Board with quick filter for Networking component"

    # Component dashboard
    dashboards:
      - url: https://jira.example.com/secure/Dashboard.jspa?selectPageId=30000
        name: "Networking Component Health"

    # Component-specific queries
    jql_filters:
      # All networking issues across projects
      - "component = 'Networking' AND updated >= -1d ORDER BY priority DESC"

      # Networking blockers
      - "component = 'Networking' AND priority = Blocker AND status != Closed"

      # Customer-reported networking issues
      - "component = 'Networking' AND labels = 'customer-reported' AND created >= -30d"
```

### Pattern 4: Release Tracking

**Use case**: Release manager tracking a specific release

```yaml
sources:
  jira:
    enabled: true
    server: https://jira.example.com

    # Release board
    boards:
      - url: https://jira.example.com/secure/RapidBoard.jspa?rapidView=8888
        name: "4.16 Release Board"

    # Release dashboards
    dashboards:
      - url: https://jira.example.com/secure/Dashboard.jspa?selectPageId=40000
        name: "4.16 Release Overview"

      - url: https://jira.example.com/secure/Dashboard.jspa?selectPageId=40001
        name: "4.16 Blockers & Risks"

      - url: https://jira.example.com/secure/Dashboard.jspa?selectPageId=40002
        name: "4.16 Documentation Status"

    # Release-specific queries
    jql_filters:
      # Release blockers
      - "'Target Version' = '4.16' AND priority = Blocker AND status != Closed"

      # Features not code complete
      - "'Target Version' = '4.16' AND issuetype = Feature AND status != 'Code Complete'"

      # Documentation gaps
      - "'Target Version' = '4.16' AND 'Docs Status' in (None, 'In Progress')"

      # Test coverage gaps
      - "'Target Version' = '4.16' AND 'Test Coverage' = None"
```

---

## How the Skill Processes Jira Data

### Data Collection Flow

```
1. Load Configuration
   ├── Read exec-brief.yaml
   ├── Extract board URLs
   ├── Extract dashboard URLs
   └── Extract JQL filters

2. Query Jira (via MCP)
   ├── For each board:
   │   ├── Get board issues
   │   ├── Get sprint info (if Scrum)
   │   └── Apply quick filters
   │
   ├── For each dashboard:
   │   ├── Get dashboard configuration
   │   ├── For each gadget:
   │   │   ├── Extract filter
   │   │   └── Get issues
   │   └── Aggregate results
   │
   └── For each JQL filter:
       ├── Execute query
       └── Get matching issues

3. Normalize Data
   ├── Convert to common format
   ├── Extract: id, title, url, status, priority, assignee
   ├── Parse dates (created, updated, due)
   └── Extract labels and custom fields

4. Identify Teammates
   ├── Match assignee against teammate map
   ├── Match reporter against teammate map
   ├── Check watchers list
   └── Flag items with teammate involvement

5. Categorize to Matrix
   ├── Calculate urgency score
   ├── Calculate importance score
   └── Assign to quadrant (Q1, Q2, Q3, Q4)

6. Generate Output
   ├── Sort items by quadrant
   ├── Separate teammate items
   └── Generate markdown/HTML
```

### Teammate Identification

For each Jira issue, the skill checks:

```python
# From teammates.yaml
jira_map = {
  "alice.smith@company.com": "Alice Smith",
  "bob.jones@company.com": "Bob Jones",
  ...
}

# Check assignee
if issue.assignee in jira_map:
    teammates_involved.append(jira_map[issue.assignee])

# Check reporter
if issue.reporter in jira_map:
    teammates_involved.append(jira_map[issue.reporter])

# Result: issue tagged with teammate names
issue['teammates_involved'] = ['Alice Smith']
```

### Matrix Categorization

```python
# Urgency factors
if 'blocker' in issue.priority.lower():
    urgency_score += 3

if issue.due_date and days_until_due < 1:
    urgency_score += 3

if 'blocked' in issue.status.lower() and issue.teammates_involved:
    urgency_score += 2

# Importance factors
if issue.teammates_involved:
    importance_score += 3

if 'security' in issue.labels:
    importance_score += 2

# Assign to quadrant
if urgency_score >= 3 and importance_score >= 2:
    quadrant = 1  # Do First
elif urgency_score < 3 and importance_score >= 2:
    quadrant = 2  # Schedule
...
```

---

## Best Practices

### 1. Start with Existing Structures

Use boards and dashboards your team already maintains:
- ✅ Leverage existing sprint boards
- ✅ Use team dashboards everyone watches
- ✅ Add JQL only when boards/dashboards aren't enough

### 2. Layer Information Sources

Combine multiple approaches:
```yaml
boards:
  - url: ... # Team sprint board (workflow)
dashboards:
  - url: ... # Team health (metrics)
jql_filters:
  - "..." # Specific queries (edge cases)
```

### 3. Descriptive Names

Use clear, meaningful names:
```yaml
# Good
- name: "Platform Team Sprint 47 (Mar 17-28)"
- name: "Customer Escalations - Last 30 Days"

# Not as helpful
- name: "Board 1"
- name: "Dashboard"
```

### 4. Keep JQL Readable

Break complex queries into multiple filters:
```yaml
# Good - separate concerns
jql_filters:
  - "priority = Blocker AND status != Closed"
  - "component = 'Networking' AND updated >= -1d"

# Harder to maintain
jql_filters:
  - "priority = Blocker AND status != Closed OR component = 'Networking' AND updated >= -1d"
```

### 5. Document Team Conventions

Add notes to your configuration:
```yaml
boards:
  - url: https://jira.company.com/secure/RapidBoard.jspa?rapidView=1234
    name: "Platform Sprint"
    description: "2-week sprints, starting Mondays, team uses story points"
```

---

## Troubleshooting

### No Issues Found from Board

**Check**:
1. Board ID is correct in URL
2. You have permission to view the board
3. Board has active sprint (for Scrum boards)
4. Board filter includes your items

**Fix**:
```bash
# Verify board URL in browser first
# Make sure you can see issues when logged into Jira
```

### Dashboard Returns Too Many Items

**Check**:
1. Dashboard has many gadgets with large filters
2. Filters aren't scoped to your team

**Fix**:
```yaml
# Use more specific JQL instead
jql_filters:
  - "project = PROJ AND assignee in membersOf('your-team')"
```

### JQL Query Fails

**Check**:
1. JQL syntax is valid (test in Jira's issue navigator)
2. Field names are correct (case-sensitive)
3. You have permission to query those fields

**Fix**:
```bash
# Test JQL in Jira first:
# 1. Go to Issues → Search for Issues
# 2. Switch to Advanced (JQL)
# 3. Paste your query
# 4. Click Search
# 5. Fix any errors
# 6. Copy working JQL to config
```

---

## Examples Summary

See `EXAMPLES.md` for 16 detailed configuration examples including:

1. **Single Team**: Sprint board only
2. **Multi-Team Manager**: 3 teams, multiple boards
3. **Release Manager**: Release dashboard + blocker tracking
4. **Component Owner**: Component-specific queries
5. **Global Team**: Multi-timezone coordination
6. **And more...**

---

## Quick Start Checklist

- [ ] Find your team's Jira board URL
- [ ] Add board URL to `exec-brief.yaml`
- [ ] (Optional) Add team dashboard URL
- [ ] (Optional) Add custom JQL for specific needs
- [ ] Run `./validate_config.py`
- [ ] Test with `/exec-brief --daily`
- [ ] Review output, adjust configuration as needed

---

## Next Steps

1. **Review examples**: See `EXAMPLES.md` for your use case
2. **Configure sources**: Edit `exec-brief.yaml`
3. **Validate**: Run `./validate_config.py`
4. **Test**: Try `/exec-brief --daily`
5. **Refine**: Adjust based on results
