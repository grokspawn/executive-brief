# Executive Brief - Configuration Examples

This document provides real-world examples of how to configure information sources, with emphasis on team-oriented Jira boards and dashboards.

## Table of Contents

- [Minimal Configuration](#minimal-configuration)
- [Team-Focused Jira Boards](#team-focused-jira-boards)
- [Multiple Teams](#multiple-teams)
- [Sprint-Based Workflows](#sprint-based-workflows)
- [Release Management](#release-management)
- [Cross-Organizational Setup](#cross-organizational-setup)
- [Focused Configurations](#focused-configurations)

---

## Minimal Configuration

### Example 1: Just the Basics

No `exec-brief.yaml` needed - relies on defaults.

**teammates.yaml:**
```yaml
user:
  name: Jordan Smith
  github: jsmith
  jira: jordan.smith@company.com
  timezone: America/New_York

teammates:
  - name: Alex Chen
    github: achen
    jira: alex.chen@company.com
    priority: high

  - name: Morgan Lee
    github: mlee
    jira: morgan.lee@company.com
    priority: high
```

**Result:** Queries Jira for issues assigned/reported/watched by you, GitHub PRs, with no custom filters.

---

## Team-Focused Jira Boards

### Example 2: Single Team with Sprint Board

Team uses a Scrum board for sprint planning.

**exec-brief.yaml:**
```yaml
sources:
  jira:
    enabled: true
    server: https://jira.company.com

    # Track the team's sprint board
    boards:
      - url: https://jira.company.com/secure/RapidBoard.jspa?rapidView=1234
        name: "Platform Team Sprint Board"
        description: "Current sprint work for Platform team"
```

**What this captures:**
- All issues in the current sprint
- Backlog items
- Issues in progress
- Identifies which ones involve your teammates

### Example 3: Kanban Board for Continuous Flow

Team uses Kanban instead of Sprints.

**exec-brief.yaml:**
```yaml
sources:
  jira:
    enabled: true
    server: https://jira.example.com

    boards:
      - url: https://jira.example.com/secure/RapidBoard.jspa?rapidView=5678
        name: "Networking Kanban Board"
        description: "Continuous flow board for Networking team"

      - url: https://jira.example.com/secure/RapidBoard.jspa?rapidView=5679
        name: "Support Queue"
        description: "Customer support issues"
```

### Example 4: Team Dashboard with Gadgets

Team has a dashboard with various widgets showing blockers, sprint progress, etc.

**exec-brief.yaml:**
```yaml
sources:
  jira:
    enabled: true
    server: https://jira.company.com

    dashboards:
      - url: https://jira.company.com/secure/Dashboard.jspa?selectPageId=12345
        name: "Team Weekly Dashboard"
        description: "Shows current sprint, blockers, and team velocity"

      - url: https://jira.company.com/secure/Dashboard.jspa?selectPageId=12346
        name: "Release Blockers"
        description: "Critical issues blocking the release"
```

**Dashboard typically includes:**
- Sprint burndown charts
- Blocker issues
- Recently updated issues
- Team workload distribution

---

## Multiple Teams

### Example 5: Manager Tracking Multiple Teams

Engineering manager overseeing 3 teams.

**teammates.yaml:**
```yaml
user:
  name: Sam Taylor
  github: staylor
  jira: sam.taylor@company.com
  timezone: America/New_York

teammates:
  # Team A - Frontend
  - name: Alice Johnson
    github: ajohnson
    jira: alice.johnson@company.com
    slack: ajohnson
    priority: high
    notes: "Frontend team lead"

  - name: Bob Williams
    github: bwilliams
    jira: bob.williams@company.com
    slack: bwilliams
    priority: high
    notes: "Frontend - React specialist"

  # Team B - Backend
  - name: Carol Martinez
    github: cmartinez
    jira: carol.martinez@company.com
    slack: cmartinez
    priority: high
    notes: "Backend team lead"

  - name: David Brown
    github: dbrown
    jira: david.brown@company.com
    slack: dbrown
    priority: high
    notes: "Backend - API design"

  # Team C - Infrastructure
  - name: Emma Davis
    github: edavis
    jira: emma.davis@company.com
    slack: edavis
    priority: high
    notes: "Infrastructure team lead"
```

**exec-brief.yaml:**
```yaml
sources:
  jira:
    enabled: true
    server: https://jira.company.com

    # Each team's board
    boards:
      - url: https://jira.company.com/secure/RapidBoard.jspa?rapidView=100
        name: "Frontend Team Sprint"
        description: "Team A sprint work"

      - url: https://jira.company.com/secure/RapidBoard.jspa?rapidView=101
        name: "Backend Team Sprint"
        description: "Team B sprint work"

      - url: https://jira.company.com/secure/RapidBoard.jspa?rapidView=102
        name: "Infrastructure Team Sprint"
        description: "Team C sprint work"

    # Aggregate dashboard
    dashboards:
      - url: https://jira.company.com/secure/Dashboard.jspa?selectPageId=20001
        name: "Engineering Overview"
        description: "All teams combined view"

      - url: https://jira.company.com/secure/Dashboard.jspa?selectPageId=20002
        name: "Manager - Blockers & Risks"
        description: "Cross-team blockers and risks"

  github:
    enabled: true
    organizations:
      - mycompany
```

**Usage:**
```bash
# See all teams
/exec-brief --daily

# Focus on teammates only (all 5 team members)
/exec-brief --daily --teammates-only
```

---

## Sprint-Based Workflows

### Example 6: Two-Week Sprint Cycle

**exec-brief.yaml:**
```yaml
sources:
  jira:
    enabled: true
    server: https://jira.example.com

    projects:
      - OCPBUGS
      - NETOBSERV

    # Active sprint board
    boards:
      - url: https://jira.example.com/secure/RapidBoard.jspa?rapidView=1234
        name: "Sprint 47 - Observability"
        description: "March 17-28 sprint"

    # Sprint planning dashboard
    dashboards:
      - url: https://jira.example.com/secure/Dashboard.jspa?selectPageId=15000
        name: "Sprint Health"
        description: "Burndown, velocity, and completion forecast"

    # Custom JQL for sprint-related queries
    jql_filters:
      # Current sprint items
      - "project = OCPBUGS AND Sprint = 'Sprint 47' AND status != Closed"

      # Carryover risk (not started yet, sprint ending soon)
      - "project = OCPBUGS AND Sprint = 'Sprint 47' AND status = 'To Do' AND sprint not in (futureSprints())"

      # Blockers in current sprint
      - "project = OCPBUGS AND Sprint = 'Sprint 47' AND priority = Blocker"
```

### Example 7: Planning Next Sprint

Focus on backlog grooming and next sprint planning.

**exec-brief.yaml:**
```yaml
sources:
  jira:
    enabled: true
    server: https://jira.company.com

    boards:
      # Current sprint
      - url: https://jira.company.com/secure/RapidBoard.jspa?rapidView=2000&view=planning
        name: "Current Sprint"

      # Backlog view for grooming
      - url: https://jira.company.com/secure/RapidBoard.jspa?rapidView=2000&view=planning&selectedIssue=PROJ-123
        name: "Backlog - Ready for Sprint"

    jql_filters:
      # Groomed and ready for next sprint
      - "project = PROJ AND status = 'Ready for Dev' AND sprint is EMPTY ORDER BY priority DESC"

      # Needs estimation
      - "project = PROJ AND 'Story Points' is EMPTY AND status = 'To Do'"

      # Technical debt candidates
      - "project = PROJ AND labels = 'tech-debt' AND priority >= Medium"
```

---

## Release Management

### Example 8: Release Tracking Across Teams

**exec-brief.yaml:**
```yaml
sources:
  jira:
    enabled: true
    server: https://jira.example.com

    projects:
      - OCPBUGS
      - RHCLOUD

    # Release-focused dashboards
    dashboards:
      - url: https://jira.example.com/secure/Dashboard.jspa?selectPageId=30000
        name: "4.16 Release Dashboard"
        description: "All issues targeting 4.16 release"

      - url: https://jira.example.com/secure/Dashboard.jspa?selectPageId=30001
        name: "Release Blockers - Critical"
        description: "Must-fix items for release"

      - url: https://jira.example.com/secure/Dashboard.jspa?selectPageId=30002
        name: "Documentation Readiness"
        description: "Release notes and doc status"

    # Release board
    boards:
      - url: https://jira.example.com/secure/RapidBoard.jspa?rapidView=3000
        name: "4.16 Release Board"
        description: "All work items for 4.16"

    jql_filters:
      # Release blockers
      - "project = OCPBUGS AND 'Target Version' = '4.16' AND priority = Blocker AND status != Closed"

      # Features not code complete
      - "project = OCPBUGS AND 'Target Version' = '4.16' AND issuetype = Feature AND status != 'Code Complete'"

      # Documentation incomplete
      - "project = OCPBUGS AND 'Target Version' = '4.16' AND 'Docs Status' != Complete"

      # Test gaps
      - "project = OCPBUGS AND 'Target Version' = '4.16' AND 'Test Coverage' = None"
```

### Example 9: Hotfix and Patch Management

**exec-brief.yaml:**
```yaml
sources:
  jira:
    enabled: true
    server: https://jira.company.com

    dashboards:
      - url: https://jira.company.com/secure/Dashboard.jspa?selectPageId=40000
        name: "Production Issues"
        description: "Live customer-impacting issues"

      - url: https://jira.company.com/secure/Dashboard.jspa?selectPageId=40001
        name: "Hotfix Pipeline"
        description: "Issues queued for emergency patches"

    jql_filters:
      # P0 production issues
      - "project = PROD AND priority = 'P0 - Critical' AND status in ('New', 'In Progress')"

      # Customer escalations
      - "project = SUPPORT AND labels = 'escalation' AND status != Resolved"

      # Security vulnerabilities
      - "project = SECURITY AND 'CVE ID' is not EMPTY AND 'Fix Version' is EMPTY"

      # Hotfixes in flight
      - "labels = 'hotfix' AND status in ('In Progress', 'Code Review', 'Testing')"
```

---

## Cross-Organizational Setup

### Example 10: Multiple Jira Instances

Working across different Jira servers (company + open source).

**teammates.yaml:**
```yaml
user:
  name: Jordan Martinez
  github: jmartinez
  jira: jordan.martinez@company.com  # Primary
  email: jmartinez@opensource.org    # Secondary
  timezone: America/Los_Angeles
```

**exec-brief.yaml:**
```yaml
sources:
  # Company Jira
  jira:
    enabled: true
    server: https://jira.company.com

    boards:
      - url: https://jira.company.com/secure/RapidBoard.jspa?rapidView=100
        name: "Company Product Team"

    jql_filters:
      - "assignee = currentUser() AND updated >= -1d"

  # Note: For multiple Jira instances, you'd configure separate source entries
  # This would require extending the configuration schema
  # Current implementation assumes single Jira server

  github:
    enabled: true

    # Both company and open source orgs
    organizations:
      - mycompany
      - kubernetes
      - openshift
      - prometheus
```

### Example 11: Open Source Maintainer

Maintaining multiple open source projects.

**teammates.yaml:**
```yaml
user:
  name: Alex Rivera
  github: arivera
  jira: alex.rivera@example.com
  timezone: America/New_York

teammates:
  # Core team members across projects
  - name: Chris Anderson
    github: canderson
    priority: high
    notes: "Co-maintainer on kubernetes/sig-network"

  - name: Taylor Kim
    github: tkim
    priority: high
    notes: "Prometheus maintainer"

  - name: Jordan Lee
    github: jlee
    jira: jordan.lee@example.com
    priority: high
    notes: "OpenShift networking lead"
```

**exec-brief.yaml:**
```yaml
sources:
  jira:
    enabled: true
    server: https://jira.example.com

    projects:
      - OCPBUGS
      - NETOBSERV

    boards:
      - url: https://jira.example.com/secure/RapidBoard.jspa?rapidView=7777
        name: "Networking Team Board"

  github:
    enabled: true

    organizations:
      - kubernetes
      - openshift
      - prometheus
      - grafana

    # Filter for specific repos you maintain
    repositories:
      - kubernetes/kubernetes
      - openshift/ovn-kubernetes
      - prometheus/prometheus
      - grafana/loki

    filters:
      review_requested: true
      mentioned: true
      team_prs: true
      assigned: true
```

---

## Focused Configurations

### Example 12: Jira-Only (No GitHub)

Focus exclusively on Jira for a team that doesn't use GitHub heavily.

**exec-brief.yaml:**
```yaml
sources:
  jira:
    enabled: true
    server: https://jira.company.com

    projects:
      - PLATFORM
      - INFRA

    boards:
      - url: https://jira.company.com/secure/RapidBoard.jspa?rapidView=500
        name: "Platform Engineering"

      - url: https://jira.company.com/secure/RapidBoard.jspa?rapidView=501
        name: "Infrastructure"

    dashboards:
      - url: https://jira.company.com/secure/Dashboard.jspa?selectPageId=50000
        name: "Team Capacity"

      - url: https://jira.company.com/secure/Dashboard.jspa?selectPageId=50001
        name: "Blockers & Dependencies"

    jql_filters:
      # Team's active work
      - "project in (PLATFORM, INFRA) AND assignee in (currentUser(), membersOf('platform-team')) AND status in ('In Progress', 'Code Review')"

      # Waiting on other teams
      - "project in (PLATFORM, INFRA) AND status = 'Blocked' AND 'Blocked By' is not EMPTY"

      # Ready for review
      - "project in (PLATFORM, INFRA) AND status = 'Code Review' AND reviewer = currentUser()"

  github:
    enabled: false

  google_docs:
    enabled: false
```

**Usage:**
```bash
/exec-brief --daily --sources jira
```

### Example 13: GitHub-Focused Developer

Individual contributor focusing on code reviews.

**exec-brief.yaml:**
```yaml
sources:
  jira:
    enabled: false

  github:
    enabled: true

    organizations:
      - mycompany

    repositories:
      - mycompany/api-server
      - mycompany/web-client
      - mycompany/mobile-app

    filters:
      review_requested: true   # PRs waiting for your review
      mentioned: true          # PRs/issues mentioning you
      team_prs: true          # PRs from teammates
      assigned: true          # Issues assigned to you
      authored: true          # Your open PRs

    # Check CI status
    check_ci: true

  google_docs:
    enabled: false
```

**Usage:**
```bash
/exec-brief --daily --sources github
```

### Example 14: Component-Specific Team

Team responsible for a specific component across projects.

**exec-brief.yaml:**
```yaml
sources:
  jira:
    enabled: true
    server: https://jira.example.com

    # Filter by component, not project
    jql_filters:
      # All networking issues across projects
      - "component = 'Networking' AND updated >= -1d ORDER BY priority DESC"

      # Blockers in networking
      - "component = 'Networking' AND priority = Blocker AND status != Closed"

      # Customer-reported networking bugs
      - "component = 'Networking' AND reporter in (customersGroup()) AND status in ('New', 'Triaged')"

      # Networking issues assigned to team
      - "component = 'Networking' AND assignee in membersOf('networking-team')"

    # Component-specific board
    boards:
      - url: https://jira.example.com/secure/RapidBoard.jspa?rapidView=8888&quickFilter=10010
        name: "Networking Backlog"
        description: "All networking issues with quick filter for component"

    dashboards:
      - url: https://jira.example.com/secure/Dashboard.jspa?selectPageId=60000
        name: "Networking - Customer Impact"

      - url: https://jira.example.com/secure/Dashboard.jspa?selectPageId=60001
        name: "Networking - Technical Debt"
```

---

## Advanced Patterns

### Example 15: SLA and Priority-Based

Track items by SLA and customer commitments.

**exec-brief.yaml:**
```yaml
sources:
  jira:
    enabled: true
    server: https://jira.company.com

    dashboards:
      - url: https://jira.company.com/secure/Dashboard.jspa?selectPageId=70000
        name: "SLA Compliance"
        description: "Issues approaching or breaching SLA"

    jql_filters:
      # SLA at risk
      - "'Time to Resolution' < 4h AND status != Resolved"

      # Customer commitments
      - "labels = 'customer-commitment' AND 'Due Date' <= 7d AND status != Done"

      # VIP customers
      - "'Customer Tier' = 'Enterprise' AND status in ('New', 'Waiting for Support')"

      # Escalated issues
      - "priority changed to 'P0 - Critical' DURING (-24h, now())"
```

### Example 16: Multi-Region Team

Team distributed across multiple timezones.

**teammates.yaml:**
```yaml
user:
  name: Morgan Taylor
  github: mtaylor
  jira: morgan.taylor@company.com
  timezone: America/New_York  # US East Coast

teammates:
  # US team
  - name: Sam Johnson
    github: sjohnson
    jira: sam.johnson@company.com
    timezone: America/Los_Angeles
    priority: high
    notes: "US West Coast - overlap 12-5pm EST"

  # Europe team
  - name: Emma Schmidt
    github: eschmidt
    jira: emma.schmidt@company.com
    timezone: Europe/Berlin
    priority: high
    notes: "Berlin - overlap 9am-12pm EST"

  # APAC team
  - name: Yuki Tanaka
    github: ytanaka
    jira: yuki.tanaka@company.com
    timezone: Asia/Tokyo
    priority: high
    notes: "Tokyo - async collaboration"
```

**exec-brief.yaml:**
```yaml
sources:
  jira:
    enabled: true
    server: https://jira.company.com

    boards:
      - url: https://jira.company.com/secure/RapidBoard.jspa?rapidView=9000
        name: "Global Team Board"
        description: "All regions combined"

    dashboards:
      - url: https://jira.company.com/secure/Dashboard.jspa?selectPageId=80000
        name: "Handoff Dashboard"
        description: "Items ready for next timezone"

    jql_filters:
      # Items needing US review (morning in US)
      - "labels = 'needs-us-review' AND status = 'Ready for Review'"

      # Items needing EU review
      - "labels = 'needs-eu-review' AND status = 'Ready for Review'"

      # Items needing APAC review
      - "labels = 'needs-apac-review' AND status = 'Ready for Review'"

      # Blockers that crossed timezone
      - "status = 'Blocked' AND updated >= -8h"
```

**Usage with --daily:**
```bash
# Captures activity across all timezones
/exec-brief --daily
```

The `--daily` flag's multi-timezone support ensures you see:
- Activity from EU team (while you were offline)
- Activity from APAC team (overnight)
- Your own timezone activity
- US West Coast activity (your afternoon)

---

## Usage Patterns

### Daily Team Standup

```bash
# Focus on what teammates need help with
/exec-brief --daily --teammates-only --save standup-notes.md
```

### Weekly Planning

```bash
# Full view for sprint planning
/exec-brief --date 2026-03-17 --save week-start.md
```

### Release Preparation

```bash
# Jira-only, focus on release items
/exec-brief --daily --sources jira --save release-prep.md
```

### Manager 1-on-1s

```bash
# Generate brief, then filter for specific teammate in the output
/exec-brief --daily --teammates-only
```

---

## Tips

1. **Start Simple**: Begin with just boards, add JQL filters as needed
2. **Dashboard vs Boards**:
   - Boards: For workflow/sprint tracking
   - Dashboards: For metrics and cross-cutting views
3. **JQL Filters**: Use for specific queries that boards/dashboards don't cover
4. **Teammate Priority**: Set `priority: high` for direct reports or close collaborators
5. **Update Regularly**: Review configuration as team structure changes

---

## Common Jira URL Patterns

### Finding Board IDs

1. Navigate to your board in Jira
2. URL will be: `https://jira.example.com/secure/RapidBoard.jspa?rapidView=1234`
3. Use the full URL in configuration

### Finding Dashboard IDs

1. Navigate to your dashboard
2. URL will be: `https://jira.example.com/secure/Dashboard.jspa?selectPageId=12345`
3. Use the full URL in configuration

### Quick Filters on Boards

Some boards have quick filters (e.g., "Only My Issues", "Blockers"):
```
https://jira.example.com/secure/RapidBoard.jspa?rapidView=1234&quickFilter=10010
```

Include the `quickFilter` parameter to activate specific filters.

---

## Validation

After creating your configuration, validate it:

```bash
./validate_config.py
```

This checks for:
- Valid YAML syntax
- Required fields present
- Proper URL formats
- Teammate identifiers

---

## Next Steps

1. Copy an example that matches your setup
2. Replace URLs with your actual Jira board/dashboard URLs
3. Update teammate information
4. Validate configuration
5. Run `/exec-brief --daily`
