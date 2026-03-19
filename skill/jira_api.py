#!/usr/bin/env python3
"""
Jira REST API v3 helper for Executive Brief

Uses direct curl calls to Jira Cloud API to bypass MCP limitations.
Implements the working /rest/api/3/search/jql endpoint.
"""

import json
import os
import subprocess
from typing import List, Dict, Any, Optional


class JiraAPI:
    """Direct Jira Cloud REST API v3 client using curl."""

    def __init__(self, server: str, email: str, api_token: Optional[str] = None):
        """
        Initialize Jira API client.

        Args:
            server: Jira server URL (e.g., 'https://jira.example.com')
            email: User email for authentication
            api_token: API token (if None, loads from ~/.claude.json)
        """
        self.server = server.rstrip('/')
        self.email = email
        self.api_token = api_token or self._load_token_from_config()

    def _load_token_from_config(self) -> str:
        """Load Jira API token from ~/.claude.json."""
        config_path = os.path.expanduser('~/.claude.json')

        if not os.path.exists(config_path):
            raise FileNotFoundError(f"Config file not found: {config_path}")

        try:
            result = subprocess.run(
                ['jq', '-r', '.mcpServers.atlassian.env.JIRA_API_TOKEN', config_path],
                capture_output=True,
                text=True,
                check=True
            )
            token = result.stdout.strip()

            if token == 'null' or not token:
                raise ValueError("JIRA_API_TOKEN not found in ~/.claude.json")

            return token
        except subprocess.CalledProcessError as e:
            raise RuntimeError(f"Failed to load Jira token: {e}")

    def search_jql(
        self,
        jql: str,
        max_results: int = 50,
        fields: Optional[List[str]] = None
    ) -> Dict[str, Any]:
        """
        Search Jira issues using JQL.

        Args:
            jql: JQL query string
            max_results: Maximum number of results to return
            fields: List of fields to return (default: common fields)

        Returns:
            Dict with 'total' and 'issues' keys

        Note:
            Email addresses in JQL MUST be quoted: assignee = "user@domain.com"
        """
        if fields is None:
            fields = [
                'key', 'summary', 'status', 'assignee', 'reporter',
                'updated', 'created', 'priority', 'labels', 'duedate'
            ]

        payload = {
            'jql': jql,
            'maxResults': max_results,
            'fields': fields
        }

        return self._post_api('/rest/api/3/search/jql', payload)

    def get_issue(self, issue_key: str, expand: Optional[str] = None) -> Dict[str, Any]:
        """
        Get a specific Jira issue.

        Args:
            issue_key: Issue key (e.g., 'PROJ-123')
            expand: Optional fields to expand (e.g., 'changelog')

        Returns:
            Issue data dict
        """
        url = f'/rest/api/3/issue/{issue_key}'
        if expand:
            url += f'?expand={expand}'

        return self._get_api(url)

    def get_project_issues(
        self,
        project_key: str,
        updated_since_days: int = 1,
        max_results: int = 100
    ) -> List[Dict[str, Any]]:
        """
        Get issues from a project updated in last N days.

        Args:
            project_key: Project key (e.g., 'OPRUN')
            updated_since_days: Number of days to look back
            max_results: Maximum results to return

        Returns:
            List of issue dicts
        """
        jql = f'project = {project_key} AND updated >= -{updated_since_days}d ORDER BY updated DESC'
        result = self.search_jql(jql, max_results=max_results)
        return result.get('issues', [])

    def get_teammate_issues(
        self,
        teammate_emails: List[str],
        updated_since_days: int = 1,
        max_results: int = 100
    ) -> List[Dict[str, Any]]:
        """
        Get issues assigned to teammates.

        Args:
            teammate_emails: List of teammate email addresses
            updated_since_days: Number of days to look back
            max_results: Maximum results to return

        Returns:
            List of issue dicts
        """
        # Quote all emails for JQL
        quoted_emails = [f'"{email}"' for email in teammate_emails]
        emails_str = ', '.join(quoted_emails)

        jql = f'assignee in ({emails_str}) AND updated >= -{updated_since_days}d ORDER BY updated DESC'
        result = self.search_jql(jql, max_results=max_results)
        return result.get('issues', [])

    def _get_api(self, endpoint: str) -> Dict[str, Any]:
        """Make GET request to Jira API."""
        url = f'{self.server}{endpoint}'

        cmd = [
            'curl', '-s',
            '-u', f'{self.email}:{self.api_token}',
            '-H', 'Accept: application/json',
            url
        ]

        result = subprocess.run(cmd, capture_output=True, text=True)

        if result.returncode != 0:
            raise RuntimeError(f"curl failed: {result.stderr}")

        try:
            return json.loads(result.stdout)
        except json.JSONDecodeError as e:
            raise RuntimeError(f"Invalid JSON response: {e}\nResponse: {result.stdout[:500]}")

    def _post_api(self, endpoint: str, payload: Dict[str, Any]) -> Dict[str, Any]:
        """Make POST request to Jira API."""
        url = f'{self.server}{endpoint}'

        cmd = [
            'curl', '-s',
            '-u', f'{self.email}:{self.api_token}',
            '-H', 'Accept: application/json',
            '-H', 'Content-Type: application/json',
            '-X', 'POST',
            url,
            '-d', json.dumps(payload)
        ]

        result = subprocess.run(cmd, capture_output=True, text=True)

        if result.returncode != 0:
            raise RuntimeError(f"curl failed: {result.stderr}")

        try:
            response = json.loads(result.stdout)
        except json.JSONDecodeError as e:
            raise RuntimeError(f"Invalid JSON response: {e}\nResponse: {result.stdout[:500]}")

        # Check for API errors
        if 'errorMessages' in response and response['errorMessages']:
            raise RuntimeError(f"Jira API error: {response['errorMessages']}")

        return response


def normalize_jira_issue(issue: Dict[str, Any], server: str) -> Dict[str, Any]:
    """
    Normalize Jira issue to common format for executive brief.

    Args:
        issue: Raw Jira issue dict from API
        server: Jira server URL

    Returns:
        Normalized issue dict
    """
    fields = issue.get('fields', {})

    # Extract assignee
    assignee_data = fields.get('assignee', {})
    assignee = assignee_data.get('emailAddress') if assignee_data else None
    assignee_name = assignee_data.get('displayName') if assignee_data else None

    # Extract reporter
    reporter_data = fields.get('reporter', {})
    reporter = reporter_data.get('emailAddress') if reporter_data else None

    # Extract status
    status_data = fields.get('status', {})
    status = status_data.get('name') if status_data else None

    # Extract priority
    priority_data = fields.get('priority', {})
    priority = priority_data.get('name') if priority_data else None

    return {
        'id': issue.get('key'),
        'title': fields.get('summary', ''),
        'source': 'jira',
        'url': f"{server}/browse/{issue.get('key')}",
        'type': fields.get('issuetype', {}).get('name', 'Unknown'),
        'status': status,
        'priority': priority,
        'created_at': fields.get('created'),
        'updated_at': fields.get('updated'),
        'due_date': fields.get('duedate'),
        'labels': fields.get('labels', []),
        'assignee': assignee,
        'assignee_name': assignee_name,
        'reporter': reporter,
        'teammates_involved': [],  # Filled in later by categorizer
        'raw_data': issue
    }


# Example usage
if __name__ == '__main__':
    # Test the API
    import sys
    import argparse

    parser = argparse.ArgumentParser(description='Test Jira API')
    parser.add_argument('--server', required=True, help='Jira server URL (e.g., https://jira.example.com)')
    parser.add_argument('--email', required=True, help='User email for authentication')
    parser.add_argument('--project', default='PROJECT', help='Project key to query (default: PROJECT)')
    args = parser.parse_args()

    try:
        api = JiraAPI(
            server=args.server,
            email=args.email
        )

        print("Testing Jira API...")
        print(f"\n1. Get {args.project} issues (last 3 days):")
        issues = api.get_project_issues(args.project, updated_since_days=3, max_results=5)

        for issue in issues:
            normalized = normalize_jira_issue(issue, api.server)
            print(f"  {normalized['id']}: {normalized['title']} [{normalized['assignee_name']}]")

        print(f"\nTotal: {len(issues)} issues")

    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)
