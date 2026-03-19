#!/usr/bin/env python3
"""
Configuration validator for Executive Brief skill.

Validates teammates.yaml and exec-brief.yaml files for correctness.
"""

import argparse
import os
import sys
import yaml
from typing import Dict, Any, List


class ConfigValidator:
    """Validates Executive Brief configuration files."""

    def __init__(self):
        self.errors = []
        self.warnings = []

    def validate_teammates(self, filepath: str) -> bool:
        """Validate teammates.yaml file."""
        print(f"Validating teammates configuration: {filepath}")

        if not os.path.exists(filepath):
            self.errors.append(f"File not found: {filepath}")
            return False

        try:
            with open(filepath, 'r') as f:
                data = yaml.safe_load(f)
        except yaml.YAMLError as e:
            self.errors.append(f"YAML parsing error: {e}")
            return False

        # Validate user section
        if 'user' not in data:
            self.errors.append("Missing required 'user' section")
            return False

        user = data['user']
        self._validate_person(user, "user", required_fields=['name'])

        # Check timezone
        if 'timezone' not in user:
            self.warnings.append("User timezone not set - defaulting to America/New_York")

        # Validate teammates section
        if 'teammates' not in data:
            self.warnings.append("No teammates configured")
        else:
            teammates = data['teammates']
            if not isinstance(teammates, list):
                self.errors.append("'teammates' must be a list")
            else:
                for i, teammate in enumerate(teammates):
                    self._validate_person(teammate, f"teammate[{i}]")

        return len(self.errors) == 0

    def _validate_person(self, person: Dict[str, Any], context: str, required_fields: List[str] = None):
        """Validate a person (user or teammate) configuration."""
        if required_fields is None:
            required_fields = ['name']

        # Check required fields
        for field in required_fields:
            if field not in person:
                self.errors.append(f"{context}: Missing required field '{field}'")

        # Check name
        if 'name' in person and not person['name']:
            self.errors.append(f"{context}: 'name' cannot be empty")

        # Validate platform identifiers
        platform_fields = ['github', 'jira', 'email', 'slack']
        has_platform = any(field in person for field in platform_fields)

        if not has_platform:
            self.warnings.append(f"{context}: No platform identifiers provided")

        # Validate slack format
        if 'slack' in person:
            slack = person['slack']
            if isinstance(slack, str):
                if not slack:
                    self.errors.append(f"{context}: slack string cannot be empty")
            elif isinstance(slack, dict):
                if not slack.get('uid') and not slack.get('handle'):
                    self.errors.append(f"{context}: slack object must have 'uid' or 'handle'")
            else:
                self.errors.append(f"{context}: slack must be string or object")

        # Validate priority (for teammates)
        if context.startswith('teammate') and 'priority' in person:
            priority = person['priority']
            if priority not in ['high', 'medium', 'low']:
                self.warnings.append(f"{context}: priority should be 'high', 'medium', or 'low' (got '{priority}')")

    def validate_source_config(self, filepath: str) -> bool:
        """Validate exec-brief.yaml file."""
        print(f"Validating source configuration: {filepath}")

        if not os.path.exists(filepath):
            print(f"Source config not found (this is OK - defaults will be used)")
            return True

        try:
            with open(filepath, 'r') as f:
                data = yaml.safe_load(f)
        except yaml.YAMLError as e:
            self.errors.append(f"YAML parsing error: {e}")
            return False

        # Validate sources section
        if 'sources' in data:
            sources = data['sources']

            # Validate Jira config
            if 'jira' in sources:
                self._validate_jira_config(sources['jira'])

            # Validate GitHub config
            if 'github' in sources:
                self._validate_github_config(sources['github'])

            # Validate Google Docs config
            if 'google_docs' in sources:
                self._validate_gdocs_config(sources['google_docs'])

        # Validate matrix_rules section
        if 'matrix_rules' in data:
            self._validate_matrix_rules(data['matrix_rules'])

        return len(self.errors) == 0

    def _validate_jira_config(self, config: Dict[str, Any]):
        """Validate Jira source configuration."""
        if 'server' in config and not config['server'].startswith('http'):
            self.errors.append("jira.server must be a valid URL")

        if 'jql_filters' in config:
            if not isinstance(config['jql_filters'], list):
                self.errors.append("jira.jql_filters must be a list")

        if 'dashboards' in config:
            if not isinstance(config['dashboards'], list):
                self.errors.append("jira.dashboards must be a list")
            else:
                for i, dashboard in enumerate(config['dashboards']):
                    if 'url' not in dashboard:
                        self.errors.append(f"jira.dashboards[{i}] missing 'url'")

        if 'boards' in config:
            if not isinstance(config['boards'], list):
                self.errors.append("jira.boards must be a list")
            else:
                for i, board in enumerate(config['boards']):
                    if 'url' not in board:
                        self.errors.append(f"jira.boards[{i}] missing 'url'")

    def _validate_github_config(self, config: Dict[str, Any]):
        """Validate GitHub source configuration."""
        if 'organizations' in config:
            if not isinstance(config['organizations'], list):
                self.errors.append("github.organizations must be a list")

        if 'repositories' in config:
            if not isinstance(config['repositories'], list):
                self.errors.append("github.repositories must be a list")

    def _validate_gdocs_config(self, config: Dict[str, Any]):
        """Validate Google Docs source configuration."""
        if 'folders' in config:
            if not isinstance(config['folders'], list):
                self.errors.append("google_docs.folders must be a list")

    def _validate_matrix_rules(self, rules: Dict[str, Any]):
        """Validate matrix_rules configuration."""
        if 'urgent_keywords' in rules:
            if not isinstance(rules['urgent_keywords'], list):
                self.errors.append("matrix_rules.urgent_keywords must be a list")

        if 'important_keywords' in rules:
            if not isinstance(rules['important_keywords'], list):
                self.errors.append("matrix_rules.important_keywords must be a list")

        if 'time_based' in rules:
            time_based = rules['time_based']
            if 'urgent_within_days' in time_based:
                if not isinstance(time_based['urgent_within_days'], (int, float)):
                    self.errors.append("matrix_rules.time_based.urgent_within_days must be a number")

            if 'important_within_days' in time_based:
                if not isinstance(time_based['important_within_days'], (int, float)):
                    self.errors.append("matrix_rules.time_based.important_within_days must be a number")

    def print_results(self) -> bool:
        """Print validation results."""
        print("\n" + "=" * 60)

        if self.errors:
            print("❌ ERRORS:")
            for error in self.errors:
                print(f"  - {error}")
            print()

        if self.warnings:
            print("⚠️  WARNINGS:")
            for warning in self.warnings:
                print(f"  - {warning}")
            print()

        if not self.errors and not self.warnings:
            print("✅ Configuration is valid!")
        elif not self.errors:
            print("✅ Configuration is valid (with warnings)")
        else:
            print("❌ Configuration has errors")

        print("=" * 60)

        return len(self.errors) == 0


def main():
    parser = argparse.ArgumentParser(description='Validate Executive Brief configuration files')
    parser.add_argument('--teammates', default='teammates.yaml',
                        help='Path to teammates.yaml (default: teammates.yaml)')
    parser.add_argument('--config', default='exec-brief.yaml',
                        help='Path to exec-brief.yaml (default: exec-brief.yaml)')

    args = parser.parse_args()

    validator = ConfigValidator()

    # Validate teammates file
    teammates_valid = validator.validate_teammates(args.teammates)

    # Validate source config (optional)
    config_valid = validator.validate_source_config(args.config)

    # Print results
    is_valid = validator.print_results()

    sys.exit(0 if is_valid else 1)


if __name__ == '__main__':
    main()
