# Executive Brief - Daily Assistant Skill

## Overview
This skill functions as an executive assistant that provides daily briefings with information organized into the Eisenhower Matrix (Urgent/Important quadrants). It aggregates data from multiple sources and prioritizes teammate needs first.

## Usage
```
/exec-brief [options]
```

### Options
- `--daily` - Generate brief covering yesterday-to-today (considers EST/EDT and home timezone boundaries)
- `--date YYYY-MM-DD` - Generate brief for specific date (default: today)
- `--sources jira,github,slack` - Specify which sources to include (default: all)
- `--teammates-only` - Show only items involving configured teammates
- `--format markdown|html` - Output format (default: markdown)
- `--save [filename]` - Save brief to file (default: exec-brief-YYYY-MM-DD.md)

### Time Zone Handling
When using `--daily`, the skill considers activity across multiple timezone boundaries:
- **EST/EDT (Eastern Time)**: 00:00 EST yesterday → 23:59 EDT today
- **Home Timezone**: 00:00 local yesterday → 23:59 local today
- This ensures you don't miss updates from teammates in different timezones or systems using EST/EDT

Example: Running at 9:00 AM PST on March 19:
- Captures: March 18 00:00 EST → March 19 09:00 PST
- Why: Catches all East Coast business day activity plus current morning

## Eisenhower Matrix Categories

### Quadrant 1: Urgent & Important (Do First)
- Blocking issues assigned to teammates
- Critical PRs waiting for your review from teammates
- High-priority Jira issues needing immediate attention
- Incidents and production issues
- Deadline-driven tasks due today/tomorrow

### Quadrant 2: Important but Not Urgent (Schedule)
- Strategic planning items
- Feature work in progress
- Important but not time-sensitive reviews
- Technical debt items flagged as important
- Team development activities

### Quadrant 3: Urgent but Not Important (Delegate)
- Non-critical issues needing quick responses
- Routine approvals
- Meeting requests
- Low-impact bugs with high visibility

### Quadrant 4: Neither Urgent nor Important (Eliminate/Review)
- Nice-to-have improvements
- Backlog grooming candidates
- Low-priority notifications
- Information-only updates

## Information Sources

### Jira (MCP: atlassian)
- **Teammate tracking**: Match issues by assignee, reporter, watcher
- **Default Queries**:
  - Issues assigned to you with status changes
  - Issues blocking teammates
  - Issues where you're a watcher
  - Sprint issues requiring attention
  - Release blockers
- **Custom JQL**: Support arbitrary JQL filters (optional)
- **Dashboard/Board Sources**: Can monitor Jira dashboards and agile boards
  - Dashboard URLs automatically extract filters and gadgets
  - Agile board URLs track sprint/backlog items
  - Useful for team-level visibility (inherently team-focused)

### GitHub (via Bash/API)
- **Teammate tracking**: Match PRs by author, reviewer, mentions
- **Queries**:
  - PRs requesting your review
  - PRs from teammates needing review
  - Mentions in issues/PRs
  - CI failures on your PRs
  - Draft PRs becoming ready

### Google Docs (MCP: google-docs)
- **Teammate tracking**: Match by document collaborators, commenters
- **Queries**:
  - Documents with unresolved comments mentioning you
  - Recent documents shared with you
  - Action items in team docs

### Slack/Communication (Future)
- **Teammate tracking**: Match by sender, mentions
- **Queries**:
  - Direct messages
  - Mentions in team channels
  - Threads you're participating in

## Configuration

The skill uses two separate configuration files:
1. **`teammates.yaml`** - Defines you and your teammates with cross-platform identifiers
2. **`.exec-brief.yaml`** - Configures information sources and matrix rules

### 1. Teammate Configuration (`teammates.yaml`)

Maps people across different platforms for identification and prioritization.

```yaml
# Your identity across platforms
user:
  name: Sam Taylor
  github: staylor
  jira: sam.taylor@company.com
  email: sam.taylor@company.com
  slack: staylor
  timezone: America/New_York  # Your home timezone (for --daily)

# Your teammates to prioritize
teammates:
  - name: Alice Smith
    github: asmith
    jira: alice.smith@company.com
    email: alice.smith@company.com
    slack: asmith
    priority: high
    notes: "Team lead"

  - name: Bob Jones
    github: bjones
    jira: bob.jones@company.com
    email: bob.jones@company.com
    slack:
      uid: U87654321
      handle: bjones
    priority: high
    notes: "Direct collaborator"

  - name: Charlie Davis
    github: cdavis
    jira: charlie.davis@company.com
    email: charlie.davis@company.com
    slack:
      handle: cdavis
    priority: medium

  - name: Dana Lee
    github: dlee
    jira: dana.lee@company.com
    email: dana.lee@company.com
    slack:
      uid: U11223344
    priority: medium
```

**Slack Identifier Flexibility**: Use whatever you have available:
- Simple string: `slack: username` (without @ symbol)
- Object with UID only: `slack: {uid: U12345678}`
- Object with handle only: `slack: {handle: username}`
- Object with both: `slack: {uid: U12345678, handle: username}`

The skill matches against whatever identifiers are found in Slack data.

### 2. Source Configuration (`.exec-brief.yaml`)

Defines which information sources to query and how to categorize items.

```yaml
sources:
  jira:
    enabled: true
    server: https://issues.redhat.com

    # Optional: Specific projects to monitor (if omitted, uses default queries)
    projects:
      - OCPBUGS
      - CNTRLPLANE

    # Optional: Custom JQL filters (not required - defaults work without these)
    jql_filters:
      - "assignee = currentUser() OR reporter = currentUser()"
      - "status changed DURING (startOfDay(), endOfDay())"
      - "priority in (Blocker, Critical) AND status != Closed"

    # Optional: Dashboard URLs for team-level visibility
    dashboards:
      - url: https://issues.redhat.com/secure/Dashboard.jspa?selectPageId=12345
        name: "Team Dashboard"
      - url: https://issues.redhat.com/secure/Dashboard.jspa?selectPageId=67890
        name: "Release Dashboard"

    # Optional: Agile board URLs for sprint/backlog tracking
    boards:
      - url: https://issues.redhat.com/secure/RapidBoard.jspa?rapidView=1234
        name: "Team Sprint Board"
      - url: https://issues.redhat.com/secure/RapidBoard.jspa?rapidView=5678
        name: "Platform Board"

  github:
    enabled: true
    organizations:
      - openshift
      - kubernetes
    filters:
      review_requested: true
      mentioned: true
      team_prs: true

  google_docs:
    enabled: true
    folders:
      - "Team Documents"
      - "Projects"
    track_comments: true

matrix_rules:
  # Rules to categorize items into Eisenhower quadrants
  urgent_keywords:
    - blocker
    - critical
    - urgent
    - emergency
    - production
    - down
    - failing

  important_keywords:
    - feature
    - security
    - performance
    - teammate
    - team

  time_based:
    urgent_within_days: 2
    important_within_days: 7
```

## Implementation Details

### Data Collection Flow
1. Load teammate and source configuration
2. Query each enabled source in parallel
3. Normalize data into common format
4. Apply teammate filters (prioritize teammate items)
5. Categorize into Eisenhower matrix
6. Generate formatted output

### Data Model
```javascript
{
  item: {
    id: string,
    title: string,
    source: 'jira' | 'github' | 'gdocs',
    url: string,
    type: string,
    status: string,
    priority: 'high' | 'medium' | 'low',
    urgent: boolean,
    important: boolean,
    quadrant: 1 | 2 | 3 | 4,
    teammates_involved: string[],
    created_at: date,
    updated_at: date,
    due_date?: date,
    labels: string[],
    context: string
  }
}
```

### Categorization Logic

#### Urgency Factors (Higher = More Urgent)
- Has "blocker", "critical", "urgent" labels: +3
- Due date within 24 hours: +3
- Due date within 48 hours: +2
- Teammate is blocked: +2
- Production incident: +3
- CI/CD failing: +1
- Waiting for your review > 24h: +1

#### Importance Factors (Higher = More Important)
- Involves teammates: +3
- High priority label: +2
- Security issue: +2
- Feature work: +1
- Has dependencies: +1
- Strategic/OKR related: +2
- Number of people affected: +1 per person (max +3)

#### Threshold
- Urgent: score >= 3
- Important: score >= 2
- Quadrant assignment:
  - Q1: urgent=true AND important=true
  - Q2: urgent=false AND important=true
  - Q3: urgent=true AND important=false
  - Q4: urgent=false AND important=false

## Output Format

### Daily Brief Structure
```markdown
# Executive Brief - [Date]

## Summary
- Total items: X
- Teammates needing help: Y
- Critical blockers: Z

## Quadrant 1: Do First (Urgent & Important)
### Teammate Items
- [ ] [JIRA-123] Fix production blocker - @teammate (2h ago)
- [ ] [PR#456] Review critical security fix - @teammate (waiting 1d)

### Your Items
- [ ] [JIRA-789] Release blocker assigned to you

## Quadrant 2: Schedule (Important, Not Urgent)
### Teammate Items
- [ ] [PR#234] Review feature implementation - @teammate

### Your Items
- [ ] [JIRA-456] Implement new feature

## Quadrant 3: Delegate (Urgent, Not Important)
- [ ] [ISSUE#789] Routine approval needed

## Quadrant 4: Review Later (Neither Urgent nor Important)
- [ ] [JIRA-999] Backlog refinement

## Actions Recommended
1. Review PR#456 for @teammate (critical security fix)
2. Unblock @teammate on JIRA-123
3. Address JIRA-789 release blocker

## Metrics
- Items by source: Jira (5), GitHub (3), GDocs (2)
- Items by quadrant: Q1 (2), Q2 (4), Q3 (1), Q4 (3)
- Teammate involvement: 60%
```

## Integration with Claude Code

### As a Daily Hook
Configure in settings.json to run automatically:
```json
{
  "hooks": {
    "on_session_start": {
      "command": "/exec-brief --teammates-only --save",
      "enabled": true,
      "schedule": "daily at 9:00 AM"
    }
  }
}
```

### With Other Skills
- Combine with `/jira:status-rollup` for deeper Jira insights
- Use with `/utils:gh-attention` for GitHub focus
- Pair with `/git:summary` for repository context

## Future Enhancements
- [ ] Email digest support
- [ ] Slack integration for team notifications
- [ ] Calendar integration for meeting prep
- [ ] AI-powered priority recommendations
- [ ] Historical trending analysis
- [ ] Customizable matrix rules per user
- [ ] Mobile-friendly HTML output
- [ ] Integration with task management tools
