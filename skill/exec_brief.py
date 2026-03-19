#!/usr/bin/env python3
"""
Executive Brief Generator

Aggregates information from multiple sources (Jira, GitHub, Google Docs) and
organizes items into the Eisenhower Matrix, prioritizing teammate needs.
"""

import argparse
import json
import os
import sys
import yaml
import subprocess
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Any
import pytz

# Import Jira API helper
from jira_api import JiraAPI, normalize_jira_issue


class Config:
    """Configuration loader for teammates and sources."""

    def __init__(self, teammates_file: Optional[str] = None, config_file: Optional[str] = None):
        self.teammates_file = self._find_file(
            teammates_file,
            ['teammates.yaml', os.path.expanduser('~/.config/exec-brief/teammates.yaml')]
        )
        self.config_file = self._find_file(
            config_file,
            ['exec-brief.yaml', os.path.expanduser('~/.config/exec-brief/config.yaml')]
        )

        self.user = None
        self.teammates = []
        self.sources = {}
        self.matrix_rules = {}

        self._load_teammates()
        self._load_config()
        self._build_maps()

    def _find_file(self, provided: Optional[str], search_paths: List[str]) -> Optional[str]:
        """Find configuration file from provided path or search paths."""
        if provided and os.path.exists(provided):
            return provided

        for path in search_paths:
            if os.path.exists(path):
                return path

        return None

    def _load_teammates(self):
        """Load teammates configuration."""
        if not self.teammates_file:
            raise FileNotFoundError(
                "teammates.yaml not found. Please create it in current directory "
                "or ~/.config/exec-brief/"
            )

        with open(self.teammates_file, 'r') as f:
            data = yaml.safe_load(f)

        self.user = data.get('user', {})
        self.teammates = data.get('teammates', [])

        # Validate user has required fields
        if not self.user.get('name'):
            raise ValueError("User must have 'name' field in teammates.yaml")

    def _load_config(self):
        """Load source and matrix configuration."""
        # Defaults
        self.sources = {
            'jira': {'enabled': True},
            'github': {'enabled': True},
            'google_docs': {'enabled': False},
        }

        self.matrix_rules = {
            'urgent_keywords': ['blocker', 'critical', 'urgent', 'emergency', 'production', 'down', 'failing'],
            'important_keywords': ['feature', 'security', 'performance', 'teammate', 'strategic', 'okr'],
            'time_based': {
                'urgent_within_days': 2,
                'important_within_days': 7
            }
        }

        # Load user overrides if config file exists
        if self.config_file:
            with open(self.config_file, 'r') as f:
                data = yaml.safe_load(f)

            if 'sources' in data:
                self.sources.update(data['sources'])
            if 'matrix_rules' in data:
                self.matrix_rules.update(data['matrix_rules'])

    def _build_maps(self):
        """Build lookup maps for teammate identification."""
        self.github_map = {}
        self.jira_map = {}
        self.slack_map = {}
        self.email_map = {}

        for teammate in self.teammates:
            name = teammate['name']

            if 'github' in teammate:
                self.github_map[teammate['github']] = name

            if 'jira' in teammate:
                self.jira_map[teammate['jira']] = name

            if 'email' in teammate:
                self.email_map[teammate['email']] = name

            if 'slack' in teammate:
                slack_id = teammate['slack']
                if isinstance(slack_id, str):
                    self.slack_map[slack_id] = name
                elif isinstance(slack_id, dict):
                    if 'uid' in slack_id:
                        self.slack_map[slack_id['uid']] = name
                    if 'handle' in slack_id:
                        self.slack_map[slack_id['handle']] = name

        # Add user to maps
        if 'github' in self.user:
            self.github_map[self.user['github']] = self.user['name']
        if 'jira' in self.user:
            self.jira_map[self.user['jira']] = self.user['name']
        if 'email' in self.user:
            self.email_map[self.user['email']] = self.user['name']
        if 'slack' in self.user:
            slack_id = self.user['slack']
            if isinstance(slack_id, str):
                self.slack_map[slack_id] = self.user['name']
            elif isinstance(slack_id, dict):
                if 'uid' in slack_id:
                    self.slack_map[slack_id['uid']] = self.user['name']
                if 'handle' in slack_id:
                    self.slack_map[slack_id['handle']] = self.user['name']


class TimeRangeCalculator:
    """Calculate time ranges considering multiple timezones."""

    @staticmethod
    def calculate_daily_range(home_timezone: str) -> tuple:
        """
        Calculate time range for --daily flag.
        Considers both EST/EDT and home timezone to ensure no activities missed.

        Returns: (start_utc, end_utc)
        """
        home_tz = pytz.timezone(home_timezone)
        eastern_tz = pytz.timezone('America/New_York')

        now = datetime.now(home_tz)
        yesterday = now - timedelta(days=1)

        # Calculate boundaries in EST/EDT
        est_yesterday = yesterday.astimezone(eastern_tz).replace(hour=0, minute=0, second=0, microsecond=0)
        est_now = now.astimezone(eastern_tz)

        # Calculate boundaries in home timezone
        home_yesterday = yesterday.replace(hour=0, minute=0, second=0, microsecond=0)
        home_now = now

        # Use earliest start and latest end
        start_time = min(est_yesterday, home_yesterday)
        end_time = max(est_now, home_now)

        # Convert to UTC
        start_utc = start_time.astimezone(pytz.UTC)
        end_utc = end_time.astimezone(pytz.UTC)

        return start_utc, end_utc

    @staticmethod
    def calculate_date_range(date_str: str, timezone: str) -> tuple:
        """
        Calculate time range for specific date.

        Returns: (start_utc, end_utc)
        """
        tz = pytz.timezone(timezone)
        date_obj = datetime.strptime(date_str, '%Y-%m-%d')

        start_time = tz.localize(date_obj.replace(hour=0, minute=0, second=0))
        end_time = tz.localize(date_obj.replace(hour=23, minute=59, second=59))

        start_utc = start_time.astimezone(pytz.UTC)
        end_utc = end_time.astimezone(pytz.UTC)

        return start_utc, end_utc


class ItemCategorizer:
    """Categorize items into Eisenhower Matrix."""

    def __init__(self, config: Config):
        self.config = config

    def identify_teammates(self, item: Dict[str, Any]) -> List[str]:
        """Identify which teammates are involved in this item."""
        teammates = []

        if item['source'] == 'jira':
            assignee = item.get('assignee', '')
            reporter = item.get('reporter', '')

            if assignee in self.config.jira_map:
                teammates.append(self.config.jira_map[assignee])
            if reporter in self.config.jira_map:
                teammates.append(self.config.jira_map[reporter])

        elif item['source'] == 'github':
            author = item.get('author', '')

            if author in self.config.github_map:
                teammates.append(self.config.github_map[author])

        # Remove duplicates and user themselves
        teammates = list(set(teammates))
        if self.config.user['name'] in teammates:
            teammates.remove(self.config.user['name'])

        return teammates

    def calculate_urgency_score(self, item: Dict[str, Any]) -> int:
        """Calculate urgency score (0-10)."""
        score = 0

        # Keyword matching
        text = f"{item['title']} {str(item.get('labels', []))}".lower()
        urgent_keywords = self.config.matrix_rules['urgent_keywords']

        for keyword in urgent_keywords:
            if keyword in text:
                score += 3
                break

        # Priority field (Jira)
        priority = item.get('priority', '')
        if priority in ['Blocker', 'Critical']:
            score += 3
        elif priority == 'Major':
            score += 2

        # Due date
        if item.get('due_date'):
            try:
                due_date = datetime.fromisoformat(item['due_date'].replace('Z', '+00:00'))
                days_until = (due_date - datetime.now(pytz.UTC)).days

                if days_until <= 0:
                    score += 4  # Overdue
                elif days_until <= 1:
                    score += 3  # Due within 24h
                elif days_until <= 2:
                    score += 2  # Due within 48h
            except:
                pass

        # Teammate is blocked
        if 'blocked' in text and item.get('teammates_involved'):
            score += 2

        # CI/CD failing
        if item['source'] == 'github':
            if 'ci' in text and 'fail' in text:
                score += 2

        return min(score, 10)

    def calculate_importance_score(self, item: Dict[str, Any]) -> int:
        """Calculate importance score (0-10)."""
        score = 0

        # Teammates involved
        if item.get('teammates_involved'):
            score += 3

        # Important keywords
        text = f"{item['title']} {str(item.get('labels', []))}".lower()
        important_keywords = self.config.matrix_rules['important_keywords']

        for keyword in important_keywords:
            if keyword in text:
                score += 2
                break

        # Priority field
        priority = item.get('priority', '')
        if priority in ['Blocker', 'Critical', 'Major']:
            score += 2
        elif priority == 'High':
            score += 1

        # Security issues
        if 'security' in text or 'cve' in text:
            score += 2

        return min(score, 10)

    def categorize(self, item: Dict[str, Any]) -> Dict[str, Any]:
        """Categorize item into Eisenhower quadrant."""
        # Identify teammates
        item['teammates_involved'] = self.identify_teammates(item)

        # Calculate scores
        urgency_score = self.calculate_urgency_score(item)
        importance_score = self.calculate_importance_score(item)

        # Determine quadrant
        URGENT_THRESHOLD = 3
        IMPORTANT_THRESHOLD = 2

        is_urgent = urgency_score >= URGENT_THRESHOLD
        is_important = importance_score >= IMPORTANT_THRESHOLD

        if is_urgent and is_important:
            quadrant = 1
        elif not is_urgent and is_important:
            quadrant = 2
        elif is_urgent and not is_important:
            quadrant = 3
        else:
            quadrant = 4

        item['urgent'] = is_urgent
        item['important'] = is_important
        item['quadrant'] = quadrant
        item['urgency_score'] = urgency_score
        item['importance_score'] = importance_score

        return item


class OutputGenerator:
    """Generate formatted output."""

    def __init__(self, config: Config):
        self.config = config

    def generate_markdown(self, items: List[Dict[str, Any]], time_range: Optional[tuple] = None) -> str:
        """Generate markdown output."""
        output = []

        # Header
        output.append(f"# Executive Brief - {datetime.now().strftime('%B %d, %Y')}")
        output.append("")

        # Summary
        total_items = len(items)
        teammates_involved = set()
        critical_blockers = 0

        for item in items:
            if item.get('teammates_involved'):
                teammates_involved.update(item['teammates_involved'])
            if item.get('priority') in ['Blocker', 'Critical']:
                critical_blockers += 1

        output.append("## Summary")
        output.append(f"- **Total items**: {total_items}")
        output.append(f"- **Teammates needing help**: {len(teammates_involved)} ({', '.join(sorted(teammates_involved))})")
        output.append(f"- **Critical blockers**: {critical_blockers}")

        if time_range:
            start, end = time_range
            output.append(f"- **Time range**: {start.strftime('%b %d %H:%M %Z')} - {end.strftime('%b %d %H:%M %Z')}")

        output.append("")
        output.append("---")
        output.append("")

        # Organize by quadrant
        quadrants = {
            1: ("Quadrant 1: Do First (Urgent & Important)", "🔥"),
            2: ("Quadrant 2: Schedule (Important, Not Urgent)", "📅"),
            3: ("Quadrant 3: Delegate (Urgent, Not Important)", "👥"),
            4: ("Quadrant 4: Review Later (Neither Urgent nor Important)", "📋")
        }

        for q_num in [1, 2, 3, 4]:
            q_title, q_icon = quadrants[q_num]
            q_items = [item for item in items if item['quadrant'] == q_num]

            if not q_items:
                continue

            output.append(f"## {q_icon} {q_title}")
            output.append("")

            # Separate teammate items
            teammate_items = [item for item in q_items if item.get('teammates_involved')]
            own_items = [item for item in q_items if not item.get('teammates_involved')]

            if teammate_items and q_num <= 2:  # Only show teammate section for Q1 and Q2
                output.append("### Teammate Items")
                for item in teammate_items:
                    output.extend(self._format_item(item))
                output.append("")

            if own_items and q_num <= 2:
                output.append("### Your Items")
                for item in own_items:
                    output.extend(self._format_item(item))
                output.append("")
            elif q_num > 2:
                # Q3 and Q4 don't separate
                for item in q_items:
                    output.extend(self._format_item(item))
                output.append("")

            output.append("---")
            output.append("")

        # Recommended actions
        output.append("## 🎯 Recommended Actions")
        output.append("")

        q1_items = [item for item in items if item['quadrant'] == 1]
        q1_items.sort(key=lambda x: (x['urgency_score'] + x['importance_score']), reverse=True)

        for i, item in enumerate(q1_items[:5], 1):
            teammates_str = f" for @{', @'.join(item['teammates_involved'])}" if item.get('teammates_involved') else ""
            output.append(f"{i}. **{item['id']}**: {item['title']}{teammates_str}")

        output.append("")
        output.append("---")
        output.append("")

        # Metrics
        output.append("## 📊 Metrics")
        output.append("")

        # By source
        source_counts = {}
        for item in items:
            source_counts[item['source']] = source_counts.get(item['source'], 0) + 1

        output.append("### By Source")
        for source, count in sorted(source_counts.items()):
            output.append(f"- {source.title()}: {count} items")

        output.append("")

        # By quadrant
        output.append("### By Quadrant")
        for q_num in [1, 2, 3, 4]:
            count = len([item for item in items if item['quadrant'] == q_num])
            output.append(f"- Q{q_num}: {count} items")

        output.append("")

        # Teammate involvement
        teammate_count = len([item for item in items if item.get('teammates_involved')])
        percentage = (teammate_count / total_items * 100) if total_items > 0 else 0
        output.append("### Teammate Involvement")
        output.append(f"- Items involving teammates: {teammate_count} ({percentage:.0f}%)")

        return "\n".join(output)

    def _format_item(self, item: Dict[str, Any]) -> List[str]:
        """Format a single item for markdown output."""
        lines = []

        teammates_str = f" - @{', @'.join(item['teammates_involved'])}" if item.get('teammates_involved') else ""
        lines.append(f"- [ ] **[{item['id']}] {item['title']}**{teammates_str}")

        # Metadata line
        meta = [f"Source: {item['source'].title()}"]

        if item.get('priority'):
            meta.append(f"Priority: {item['priority']}")

        if item.get('status'):
            meta.append(f"Status: {item['status']}")

        if item.get('updated_at'):
            try:
                updated = datetime.fromisoformat(item['updated_at'].replace('Z', '+00:00'))
                ago = datetime.now(pytz.UTC) - updated
                if ago.days > 0:
                    meta.append(f"Updated: {ago.days}d ago")
                elif ago.seconds > 3600:
                    meta.append(f"Updated: {ago.seconds // 3600}h ago")
            except:
                pass

        lines.append(f"  - {' | '.join(meta)}")
        lines.append(f"  - URL: {item['url']}")

        # Tags
        tags = []
        if item.get('priority') in ['Blocker', 'Critical']:
            tags.append('🔥 Critical')
        if item.get('teammates_involved'):
            tags.append('👥 Teammate')
        if 'security' in str(item.get('labels', [])).lower():
            tags.append('🔒 Security')

        if tags:
            lines.append(f"  - {' | '.join(tags)}")

        lines.append("")

        return lines


def query_jira(config: Config, time_range: Optional[tuple]) -> List[Dict[str, Any]]:
    """Query Jira for issues using direct API."""
    items = []

    try:
        jira_config = config.sources.get('jira', {})
        jira_server = jira_config.get('server', 'https://jira.example.com')
        user_email = config.user.get('jira')

        if not user_email:
            print("Warning: No Jira email configured in teammates.yaml", file=sys.stderr)
            return items

        api = JiraAPI(server=jira_server, email=user_email)

        # Get issues from configured projects
        projects = jira_config.get('projects', [])
        if projects:
            for project in projects:
                try:
                    issues = api.get_project_issues(project, updated_since_days=1, max_results=100)
                    for issue in issues:
                        items.append(normalize_jira_issue(issue, jira_server))
                except Exception as e:
                    print(f"Warning: Failed to query {project}: {e}", file=sys.stderr)

        # Get issues assigned to teammates and user
        teammate_emails = [t.get('jira') for t in config.teammates if t.get('jira')]
        if config.user.get('jira'):
            teammate_emails.append(config.user.get('jira'))

        if teammate_emails:
            try:
                issues = api.get_teammate_issues(teammate_emails, updated_since_days=7, max_results=100)
                for issue in issues:
                    normalized = normalize_jira_issue(issue, jira_server)
                    # Avoid duplicates
                    if not any(i['id'] == normalized['id'] for i in items):
                        items.append(normalized)
            except Exception as e:
                print(f"Warning: Failed to query teammate issues: {e}", file=sys.stderr)

    except Exception as e:
        print(f"Error querying Jira: {e}", file=sys.stderr)

    return items


def query_github(config: Config, time_range: Optional[tuple]) -> List[Dict[str, Any]]:
    """Query GitHub for PRs and issues using gh CLI."""
    items = []

    try:
        # Get PRs requesting review
        result = subprocess.run(
            ['gh', 'search', 'prs', '--review-requested=@me', '--state=open',
             '--json', 'number,title,url,repository,author,updatedAt,createdAt,labels',
             '--limit', '20'],
            capture_output=True,
            text=True,
            check=True
        )
        review_prs = json.loads(result.stdout)

        for pr in review_prs:
            items.append({
                'id': f"PR#{pr['number']}",
                'title': pr.get('title', ''),
                'source': 'github',
                'url': pr.get('url', ''),
                'type': 'pull_request',
                'status': 'review_requested',
                'priority': None,
                'created_at': pr.get('createdAt'),
                'updated_at': pr.get('updatedAt'),
                'due_date': None,
                'labels': [l.get('name', '') for l in pr.get('labels', [])],
                'assignee': None,
                'reporter': pr.get('author', {}).get('login'),
                'repo': pr.get('repository', {}).get('nameWithOwner', ''),
                'teammates_involved': []
            })

        # Get teammate PRs
        teammate_github = [t.get('github') for t in config.teammates if t.get('github')]
        for teammate in teammate_github[:5]:  # Limit to avoid too many queries
            try:
                result = subprocess.run(
                    ['gh', 'search', 'prs', f'--author={teammate}', '--state=open',
                     '--json', 'number,title,url,repository,author,updatedAt',
                     '--limit', '5'],
                    capture_output=True,
                    text=True,
                    check=True,
                    timeout=10
                )
                teammate_prs = json.loads(result.stdout)

                for pr in teammate_prs:
                    # Avoid duplicates
                    pr_id = f"PR#{pr['number']}"
                    if any(i['id'] == pr_id for i in items):
                        continue

                    items.append({
                        'id': pr_id,
                        'title': pr.get('title', ''),
                        'source': 'github',
                        'url': pr.get('url', ''),
                        'type': 'pull_request',
                        'status': 'open',
                        'priority': None,
                        'created_at': None,
                        'updated_at': pr.get('updatedAt'),
                        'due_date': None,
                        'labels': [],
                        'assignee': None,
                        'reporter': pr.get('author', {}).get('login'),
                        'repo': pr.get('repository', {}).get('nameWithOwner', ''),
                        'teammates_involved': []
                    })
            except Exception as e:
                print(f"Warning: Failed to get PRs for {teammate}: {e}", file=sys.stderr)

    except subprocess.CalledProcessError as e:
        print(f"Error querying GitHub: {e}", file=sys.stderr)
    except FileNotFoundError:
        print("Error: gh CLI not found. Install with: brew install gh", file=sys.stderr)

    return items


def main():
    parser = argparse.ArgumentParser(description='Executive Brief Generator')
    parser.add_argument('--daily', action='store_true', help='Cover yesterday-to-today activity')
    parser.add_argument('--date', help='Generate brief for specific date (YYYY-MM-DD)')
    parser.add_argument('--sources', help='Comma-separated list of sources (jira,github,gdocs)')
    parser.add_argument('--teammates-only', action='store_true', help='Show only items involving teammates')
    parser.add_argument('--format', choices=['markdown', 'html'], default='markdown', help='Output format')
    parser.add_argument('--save', nargs='?', const=True, help='Save to file (optional filename)')
    parser.add_argument('--config', help='Path to teammates.yaml')
    parser.add_argument('--source-config', help='Path to exec-brief.yaml')

    args = parser.parse_args()

    try:
        # Load configuration
        config = Config(teammates_file=args.config, config_file=args.source_config)

        # Calculate time range
        time_range = None
        if args.daily:
            timezone = config.user.get('timezone', 'America/New_York')
            time_range = TimeRangeCalculator.calculate_daily_range(timezone)
            print(f"Time range: {time_range[0]} to {time_range[1]}", file=sys.stderr)
        elif args.date:
            timezone = config.user.get('timezone', 'America/New_York')
            time_range = TimeRangeCalculator.calculate_date_range(args.date, timezone)

        # Query data sources
        items = []

        # Query Jira
        if config.sources.get('jira', {}).get('enabled', True):
            items.extend(query_jira(config, time_range))

        # Query GitHub
        if config.sources.get('github', {}).get('enabled', True):
            items.extend(query_github(config, time_range))

        # Categorize items
        categorizer = ItemCategorizer(config)
        categorized_items = [categorizer.categorize(item) for item in items]

        # Filter if teammates-only
        if args.teammates_only:
            categorized_items = [item for item in categorized_items if item.get('teammates_involved')]

        # Generate output
        generator = OutputGenerator(config)
        output = generator.generate_markdown(categorized_items, time_range)

        # Save or print
        if args.save:
            if args.save is True:
                filename = f"exec-brief-{datetime.now().strftime('%Y-%m-%d')}.md"
            else:
                filename = args.save

            with open(filename, 'w') as f:
                f.write(output)

            print(f"Executive brief saved to: {filename}", file=sys.stderr)
        else:
            print(output)

    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == '__main__':
    main()
