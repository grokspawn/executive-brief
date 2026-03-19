---
name: Executive Brief
description: Generate daily executive briefings with information organized into the Eisenhower Matrix, prioritizing teammate needs first
---

# Executive Brief

This skill functions as an executive assistant that aggregates information from multiple sources (Jira, GitHub, Google Docs, Slack) and organizes items into the Eisenhower Matrix (Urgent/Important quadrants), with special emphasis on prioritizing teammate activities.

## When to Use This Skill

Use this skill when:
- You need a daily briefing of work items and team activities
- You want to prioritize what to work on based on urgency and importance
- You need to identify where teammates need help
- You want a consolidated view across Jira, GitHub, and other platforms
- You need to understand what happened over the past 24 hours (--daily flag)

## Prerequisites

1. **Configuration Files**
   - `teammates.yaml` - Defines you and your teammates (see Configuration section)
   - `exec-brief.yaml` - Optional source configuration (defaults work without it)

2. **Jira Access** (via direct REST API)
   - Jira API token stored in `~/.claude.json` at `mcpServers.atlassian.env.JIRA_API_TOKEN`
   - User email from `teammates.yaml` user.jira field
   - Uses Jira Cloud REST API v3 (`/rest/api/3/search/jql`)
   - Bypasses MCP to avoid deprecated API limitations

3. **GitHub Access**
   - `gh` CLI tool installed and authenticated (for GitHub queries)
   - Or GitHub API access

4. **Optional: Google Docs MCP**
   - `mcp__google-docs__*` tools for Google Docs integration (if enabled)

## Input Parameters

The user will invoke with:
```
/exec-brief [options]
```

### Command-line Options

- `--daily` - Cover yesterday-to-today activity (considers EST/EDT and home timezone)
- `--date YYYY-MM-DD` - Generate brief for specific date (default: today)
- `--sources jira,github,gdocs,slack` - Comma-separated list of sources (default: all enabled)
- `--teammates-only` - Show only items involving configured teammates
- `--format markdown|html` - Output format (default: markdown)
- `--save [filename]` - Save to file (default: exec-brief-YYYY-MM-DD.md)

### Time Zone Handling for --daily

When `--daily` is specified:
1. Calculate yesterday and today boundaries in **EST/EDT** (America/New_York)
2. Calculate yesterday and today boundaries in **user's home timezone** (from teammates.yaml)
3. Use the earliest start time and latest end time to ensure no activities are missed
4. This captures all activity across timezone boundaries

Example at 9:00 AM PST on March 19, 2026:
- EST/EDT range: March 18 00:00 EST → March 19 23:59 EDT
- PST range: March 18 00:00 PST → March 19 09:00 PST
- **Combined range**: March 18 00:00 EST → March 19 09:00 PST

## Configuration Files

### 1. teammates.yaml

Location: `teammates.yaml` in current directory, or `~/.config/exec-brief/teammates.yaml`

```yaml
# Your identity across platforms
user:
  name: Sam Taylor
  github: staylor
  jira: sam.taylor@company.com
  email: sam.taylor@company.com
  slack: staylor
  timezone: America/New_York

# Your teammates to prioritize
teammates:
  - name: Alice Smith
    github: asmith
    jira: alice.smith@company.com
    email: alice.smith@company.com
    slack: asmith
    priority: high

  - name: Bob Jones
    github: bjones
    jira: bob.jones@company.com
    slack:
      uid: U87654321
      handle: bjones
    priority: high
```

**Slack identifier flexibility**:
- Simple string: `slack: username`
- Object with UID: `slack: {uid: U12345678}`
- Object with handle: `slack: {handle: username}`
- Object with both: `slack: {uid: U12345678, handle: username}`
- The skill matches whatever is provided against Slack data

### 2. exec-brief.yaml (Optional)

Location: `exec-brief.yaml` in current directory, or `~/.config/exec-brief/config.yaml`

```yaml
sources:
  jira:
    enabled: true
    server: https://jira.example.com

    # Optional: specific projects
    projects:
      - OCPBUGS
      - CNTRLPLANE

    # Optional: custom JQL filters
    jql_filters:
      - "priority in (Blocker, Critical) AND status != Closed"

    # Optional: dashboard URLs for team visibility
    dashboards:
      - url: https://jira.example.com/secure/Dashboard.jspa?selectPageId=12345
        name: "Team Dashboard"

    # Optional: agile board URLs
    boards:
      - url: https://jira.example.com/secure/RapidBoard.jspa?rapidView=1234
        name: "Team Sprint Board"

  github:
    enabled: true
    organizations:
      - openshift
      - kubernetes

  google_docs:
    enabled: true
    folders:
      - "Team Documents"

matrix_rules:
  urgent_keywords:
    - blocker
    - critical
    - urgent
    - emergency
    - production

  important_keywords:
    - feature
    - security
    - performance
    - teammate

  time_based:
    urgent_within_days: 2
    important_within_days: 7
```

## Implementation Steps

### Step 1: Load Configuration

1. **Find and load teammates.yaml**
   ```bash
   # Check current directory first, then ~/.config/exec-brief/
   if [ -f "teammates.yaml" ]; then
     TEAMMATES_FILE="teammates.yaml"
   elif [ -f "$HOME/.config/exec-brief/teammates.yaml" ]; then
     TEAMMATES_FILE="$HOME/.config/exec-brief/teammates.yaml"
   else
     echo "Error: teammates.yaml not found"
     exit 1
   fi
   ```

2. **Parse teammates.yaml** using Python or yq
   - Extract user information (name, identifiers, timezone)
   - Extract teammates list with their identifiers
   - Build lookup maps for each platform: github_map, jira_map, slack_map

3. **Load exec-brief.yaml** (optional)
   - Use defaults if not found
   - Merge user config with defaults

### Step 2: Calculate Time Ranges

If `--daily` flag is set:

```python
import pytz
from datetime import datetime, timedelta

# Get user's home timezone from teammates.yaml
home_tz = pytz.timezone(user_timezone)  # e.g., 'America/Los_Angeles'
eastern_tz = pytz.timezone('America/New_York')

# Calculate boundaries
now = datetime.now(home_tz)
yesterday = now - timedelta(days=1)

# EST/EDT boundaries
est_start = yesterday.replace(hour=0, minute=0, second=0).astimezone(eastern_tz)
est_end = now.astimezone(eastern_tz)

# Home timezone boundaries
home_start = yesterday.replace(hour=0, minute=0, second=0)
home_end = now

# Use earliest start and latest end
start_time = min(est_start, home_start)
end_time = max(est_end, home_end)

# Convert to UTC for API queries
start_utc = start_time.astimezone(pytz.UTC)
end_utc = end_time.astimezone(pytz.UTC)
```

### Step 3: Query Data Sources

Query all enabled sources in parallel. For each source, collect items and normalize them to a common format.

#### 3.1 Query Jira

Use direct Jira REST API v3 calls via `curl` (bypasses MCP limitations):

**Authentication**:
- Load from `~/.claude.json`: `mcpServers.atlassian.env.JIRA_API_TOKEN`
- Email from `teammates.yaml`: user.jira field

**API Endpoint**:
- Use `/rest/api/3/search/jql` (POST method with JSON body)
- Old `/rest/api/3/search` endpoint is deprecated

**Default Queries** (if no custom JQL):
```bash
# Query for user and teammate issues
JIRA_TOKEN=$(cat ~/.claude.json | jq -r '.mcpServers.atlassian.env.JIRA_API_TOKEN')
USER_EMAIL="user@company.com"  # From teammates.yaml
TEAMMATE_EMAILS='"teammate1@company.com", "teammate2@company.com"'

curl -s -u "$USER_EMAIL:$JIRA_TOKEN" \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -X POST \
  'https://jira.example.com/rest/api/3/search/jql' \
  -d "{
    \"jql\": \"assignee in ($TEAMMATE_EMAILS, \\\"$USER_EMAIL\\\") AND updated >= -1d ORDER BY updated DESC\",
    \"maxResults\": 50,
    \"fields\": [\"key\", \"summary\", \"status\", \"assignee\", \"updated\", \"priority\", \"labels\", \"duedate\"]
  }"
```

**Board Queries** (if board URLs configured):
```bash
# Extract project from board config, query project issues
PROJECT="OPRUN"  # Extracted from exec-brief.yaml boards config

curl -s -u "$USER_EMAIL:$JIRA_TOKEN" \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -X POST \
  'https://jira.example.com/rest/api/3/search/jql' \
  -d "{
    \"jql\": \"project = $PROJECT AND updated >= -1d ORDER BY updated DESC\",
    \"maxResults\": 100,
    \"fields\": [\"key\", \"summary\", \"status\", \"assignee\", \"updated\", \"priority\", \"labels\"]
  }"
```

**Custom JQL** (if provided in config):
```python
# Execute each JQL filter from exec-brief.yaml
for jql_filter in config['jql_filters']:
    # Add time constraint to user's JQL
    jql_with_time = f"({jql_filter}) AND updated >= -1d"

    # Make API call
    result = query_jira_api(jql_with_time)
```

**Important JQL Notes**:
- Email addresses MUST be quoted: `"user@domain.com"`
- Use `assignee in ("email1", "email2")` for multiple emails
- Character `@` must be in quotes or escaped as `\\u0040`

**Normalize Jira items**:
```python
{
  'id': issue.key,
  'title': issue.summary,
  'source': 'jira',
  'url': f"{jira_server}/browse/{issue.key}",
  'type': issue.type,
  'status': issue.status,
  'priority': issue.priority,
  'created_at': issue.created,
  'updated_at': issue.updated,
  'due_date': issue.duedate,
  'labels': issue.labels,
  'assignee': issue.assignee,
  'reporter': issue.reporter,
  'teammates_involved': [],  # Fill in Step 4
  'raw_data': issue
}
```

#### 3.2 Query GitHub

Use `gh` CLI or GitHub API:

1. **PRs requesting your review**:
   ```bash
   gh search prs --review-requested=@me --state=open --json number,title,url,author,updatedAt,createdAt,labels
   ```

2. **PRs from teammates**:
   ```bash
   for teammate_gh in "${teammate_github_usernames[@]}"; do
     gh search prs --author=$teammate_gh --state=open --json number,title,url,author,updatedAt,createdAt,labels
   done
   ```

3. **Issues and PRs mentioning you**:
   ```bash
   gh search issues --mentions=@me --json number,title,url,author,updatedAt,createdAt,labels,state
   ```

4. **Check CI failures on your PRs**:
   ```bash
   gh pr list --author=@me --json number,title,url,statusCheckRollup
   ```

**Normalize GitHub items**:
```python
{
  'id': f"PR#{pr.number}" or f"ISSUE#{issue.number}",
  'title': pr.title,
  'source': 'github',
  'url': pr.url,
  'type': 'pull_request' or 'issue',
  'status': pr.state,
  'priority': None,  # Inferred from labels
  'created_at': pr.createdAt,
  'updated_at': pr.updatedAt,
  'due_date': None,
  'labels': pr.labels,
  'author': pr.author.login,
  'teammates_involved': [],  # Fill in Step 4
  'raw_data': pr
}
```

#### 3.3 Query Google Docs

Use `mcp__google-docs__*` tools:

1. **Get recent documents**:
   - Use `mcp__google-docs__getRecentGoogleDocs` with modifiedTime filter

2. **Search for documents in team folders**:
   - Use `mcp__google-docs__listFolderContents` for configured folders
   - Filter by modification time

3. **Check for unresolved comments**:
   - For each document, use `mcp__google-docs__listComments`
   - Filter comments mentioning you or teammates

**Normalize Google Docs items**:
```python
{
  'id': doc.id,
  'title': doc.name,
  'source': 'google_docs',
  'url': f"https://docs.google.com/document/d/{doc.id}",
  'type': 'document',
  'status': 'active',
  'priority': None,
  'created_at': doc.createdTime,
  'updated_at': doc.modifiedTime,
  'due_date': None,
  'labels': [],
  'teammates_involved': [],  # Fill in Step 4
  'raw_data': doc
}
```

### Step 4: Identify Teammate Involvement

For each item collected, check if teammates are involved:

```python
def identify_teammates(item, teammates_config):
    teammates_involved = []

    # Check assignee/author
    if item['source'] == 'jira':
        if item.get('assignee') in jira_teammate_map:
            teammates_involved.append(jira_teammate_map[item['assignee']])
        if item.get('reporter') in jira_teammate_map:
            teammates_involved.append(jira_teammate_map[item['reporter']])

    elif item['source'] == 'github':
        if item.get('author') in github_teammate_map:
            teammates_involved.append(github_teammate_map[item['author']])
        # Check reviewers, mentions in PR/issue
        # Parse from raw_data

    # Check labels/text for teammate names or handles
    # Check comments/descriptions

    item['teammates_involved'] = list(set(teammates_involved))
    return item
```

### Step 5: Categorize into Eisenhower Matrix

For each item, calculate urgency and importance scores, then assign to quadrant.

#### Scoring Algorithm

**Urgency Score** (0-10):
```python
urgency_score = 0

# Keyword matching in title, labels, description
urgent_keywords = ['blocker', 'critical', 'urgent', 'emergency', 'production', 'down', 'failing']
for keyword in urgent_keywords:
    if keyword in item['title'].lower() or keyword in str(item['labels']).lower():
        urgency_score += 3
        break

# Priority field (Jira)
if item.get('priority') in ['Blocker', 'Critical']:
    urgency_score += 3
elif item.get('priority') == 'Major':
    urgency_score += 2

# Due date
if item.get('due_date'):
    days_until_due = (item['due_date'] - datetime.now()).days
    if days_until_due <= 0:
        urgency_score += 4  # Overdue
    elif days_until_due <= 1:
        urgency_score += 3  # Due within 24h
    elif days_until_due <= 2:
        urgency_score += 2  # Due within 48h

# Teammate is blocked
if 'blocked' in str(item['labels']).lower() and item['teammates_involved']:
    urgency_score += 2

# CI/CD failing
if item['source'] == 'github' and 'ci' in str(item.get('statusCheckRollup', '')).lower():
    if 'fail' in str(item.get('statusCheckRollup', '')).lower():
        urgency_score += 2

# Waiting for your review
if item['source'] == 'github' and item.get('review_requested'):
    age_hours = (datetime.now() - item['created_at']).total_seconds() / 3600
    if age_hours > 24:
        urgency_score += 1

urgency_score = min(urgency_score, 10)  # Cap at 10
```

**Importance Score** (0-10):
```python
importance_score = 0

# Teammates involved
if item['teammates_involved']:
    importance_score += 3

# Important keywords
important_keywords = ['feature', 'security', 'performance', 'strategic', 'okr']
for keyword in important_keywords:
    if keyword in item['title'].lower() or keyword in str(item['labels']).lower():
        importance_score += 2
        break

# Priority field
if item.get('priority') in ['Blocker', 'Critical', 'Major']:
    importance_score += 2
elif item.get('priority') == 'High':
    importance_score += 1

# Security issues
if 'security' in str(item['labels']).lower() or 'cve' in item['title'].lower():
    importance_score += 2

# Number of people affected/involved
# Check comments, watchers, reviewers
affected_count = len(item.get('watchers', [])) + len(item.get('reviewers', []))
importance_score += min(affected_count, 3)

importance_score = min(importance_score, 10)  # Cap at 10
```

#### Quadrant Assignment

```python
# Thresholds
URGENT_THRESHOLD = 3
IMPORTANT_THRESHOLD = 2

is_urgent = urgency_score >= URGENT_THRESHOLD
is_important = importance_score >= IMPORTANT_THRESHOLD

if is_urgent and is_important:
    quadrant = 1  # Do First
elif not is_urgent and is_important:
    quadrant = 2  # Schedule
elif is_urgent and not is_important:
    quadrant = 3  # Delegate
else:
    quadrant = 4  # Eliminate/Review

item['urgent'] = is_urgent
item['important'] = is_important
item['quadrant'] = quadrant
item['urgency_score'] = urgency_score
item['importance_score'] = importance_score
```

### Step 6: Generate Output

#### Output Structure

```markdown
# Executive Brief - March 19, 2026

## Summary
- **Total items**: 15
- **Teammates needing help**: 3 (Alice, Bob, Charlie)
- **Critical blockers**: 2
- **Time range**: Mar 18 00:00 EST - Mar 19 09:00 PST

---

## Quadrant 1: Do First (Urgent & Important)

### Teammate Items
- [ ] **[JIRA-123] Production API failing for customers** - @Alice Smith
  - Source: Jira | Priority: Blocker | Updated: 2h ago
  - URL: https://jira.example.com/browse/JIRA-123
  - 🔥 Blocker | 👥 Teammate involved

- [ ] **[PR#456] Security fix for CVE-2026-1234** - @Bob Jones
  - Source: GitHub | Waiting for review: 1d 3h
  - URL: https://github.com/org/repo/pull/456
  - 🔒 Security | ⏰ Waiting 1d | 👥 Teammate involved

### Your Items
- [ ] **[JIRA-789] Release blocker - deployment fails**
  - Source: Jira | Priority: Blocker | Due: Today
  - URL: https://jira.example.com/browse/JIRA-789
  - 🔥 Blocker | 📅 Due today

---

## Quadrant 2: Schedule (Important, Not Urgent)

### Teammate Items
- [ ] **[PR#234] Implement feature X** - @Alice Smith
  - Source: GitHub | Updated: 5h ago
  - URL: https://github.com/org/repo/pull/234
  - ⭐ Feature | 👥 Teammate involved

### Your Items
- [ ] **[JIRA-456] Refactor authentication module**
  - Source: Jira | Priority: Major | Due: Mar 25
  - URL: https://jira.example.com/browse/JIRA-456
  - ⭐ Feature | 📅 Due in 6 days

---

## Quadrant 3: Delegate (Urgent, Not Important)

- [ ] **[ISSUE#789] Update documentation**
  - Source: GitHub | Mentioned: @you
  - URL: https://github.com/org/repo/issues/789

---

## Quadrant 4: Review Later (Neither Urgent nor Important)

- [ ] **[JIRA-999] Backlog grooming - nice-to-have feature**
  - Source: Jira | Priority: Low
  - URL: https://jira.example.com/browse/JIRA-999

---

## 🎯 Recommended Actions

1. **Urgent**: Review PR#456 for @Bob Jones (critical security fix)
2. **Urgent**: Help @Alice Smith with JIRA-123 (production blocker)
3. **Important**: Address JIRA-789 release blocker (due today)
4. **Important**: Review PR#234 for @Alice Smith (feature implementation)

---

## 📊 Metrics

### By Source
- Jira: 8 items
- GitHub: 5 items
- Google Docs: 2 items

### By Quadrant
- Q1 (Urgent & Important): 3 items
- Q2 (Important, Not Urgent): 6 items
- Q3 (Urgent, Not Important): 2 items
- Q4 (Neither): 4 items

### Teammate Involvement
- Items involving teammates: 9 (60%)
- Teammates needing help: Alice (3 items), Bob (2 items), Charlie (1 item)
```

#### HTML Output Format (if --format html)

Generate an HTML file with:
- CSS styling for the Eisenhower matrix (2x2 grid layout)
- Color coding: Red (Q1), Yellow (Q2), Blue (Q3), Gray (Q4)
- Interactive checkboxes
- Clickable links to source items
- Teammate avatars/badges
- Timeline view of activity

### Step 7: Save Output

If `--save` flag is provided:

```python
import os
from datetime import datetime

# Determine filename
if args.save:
    if args.save is True or args.save == '':
        filename = f"exec-brief-{datetime.now().strftime('%Y-%m-%d')}.md"
    else:
        filename = args.save
else:
    # Just print to stdout
    print(output)
    return

# Save to file
output_path = os.path.join(os.getcwd(), filename)
with open(output_path, 'w') as f:
    f.write(output)

print(f"Executive brief saved to: {output_path}")
```

## Error Handling

### Missing Configuration
```python
if not os.path.exists(teammates_file):
    print("Error: teammates.yaml not found")
    print("Please create teammates.yaml in current directory or ~/.config/exec-brief/")
    print("\nExample:")
    print("""
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
    """)
    exit(1)
```

### Source Unavailable
```python
try:
    jira_items = query_jira()
except Exception as e:
    print(f"Warning: Could not query Jira: {e}")
    jira_items = []
    # Continue with other sources
```

### Invalid Date Format
```python
if args.date:
    try:
        date_obj = datetime.strptime(args.date, '%Y-%m-%d')
    except ValueError:
        print(f"Error: Invalid date format '{args.date}'. Use YYYY-MM-DD")
        exit(1)
```

## Testing

Test with different scenarios:

1. **No teammates involved**:
   - Should still categorize items correctly
   - Should show "0 teammates needing help"

2. **Empty results**:
   - Should show "No items found for this time period"

3. **Multiple timezones**:
   - Verify --daily captures correct time range
   - Test with user in PST, teammates in EST

4. **All quadrants**:
   - Ensure items are distributed correctly
   - Verify scoring algorithm works

5. **Teammate-only filter**:
   - `--teammates-only` should only show items involving teammates

## Future Enhancements

- Email digest support
- Slack notifications for teammate items
- Calendar integration
- AI-powered priority recommendations
- Historical trending
- Mobile app integration
