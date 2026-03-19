# Executive Brief Skill - Installation Guide

## Overview

The Executive Brief skill has been implemented with the following components:

```
executive-brief/
├── exec-brief.md              # Original design document
├── INSTALL.md                 # This file
└── skill/                     # Skill implementation
    ├── SKILL.md               # Skill instructions for Claude Code
    ├── README.md              # User documentation
    ├── exec_brief.py          # Core Python implementation
    ├── validate_config.py     # Configuration validator
    ├── requirements.txt       # Python dependencies
    ├── teammates.yaml.example # Example teammates config
    └── config.yaml.example    # Example source config
```

## Installation Steps

### 1. Install Python Dependencies

```bash
cd skill/
pip install -r requirements.txt
```

Required packages:
- `pyyaml>=6.0` - YAML configuration parsing
- `pytz>=2024.1` - Timezone handling for --daily flag

### 2. Create Your Configuration

#### Create teammates.yaml

```bash
cp teammates.yaml.example teammates.yaml
# Edit with your information
nano teammates.yaml
```

**Minimal configuration**:
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

#### (Optional) Create .exec-brief.yaml

```bash
cp config.yaml.example .exec-brief.yaml
# Customize as needed
nano .exec-brief.yaml
```

The skill works with defaults if this file is not provided.

### 3. Validate Configuration

```bash
./validate_config.py
```

This checks for:
- Required fields
- Valid YAML syntax
- Proper identifier formats
- Configuration errors

Expected output:
```
Validating teammates configuration: teammates.yaml
Validating source configuration: .exec-brief.yaml
============================================================
✅ Configuration is valid!
============================================================
```

### 4. (Optional) Install to Claude Code Plugins

To make the skill available globally in Claude Code:

```bash
# Option 1: Create symlink in Claude Code plugins directory
ln -s $(pwd) ~/.claude/plugins/local/executive-brief

# Option 2: Copy to Claude Code plugins directory
cp -r $(pwd) ~/.claude/plugins/local/executive-brief
```

### 5. Test the Skill

#### Test the Python helper directly:

```bash
./exec_brief.py --daily
```

Note: This will show empty results since data source querying is implemented via Claude Code MCP tools, not in the Python script directly. The script provides categorization and output logic.

#### Test via Claude Code:

```
/exec-brief --daily
```

## Usage

### Basic Commands

```bash
# Daily brief (yesterday to today, multi-timezone aware)
/exec-brief --daily

# Brief for specific date
/exec-brief --date 2026-03-19

# Show only teammate items
/exec-brief --daily --teammates-only

# Save to file
/exec-brief --daily --save

# Specific sources only
/exec-brief --daily --sources jira,github
```

## How It Works

The skill operates in two modes:

### 1. Via Claude Code (Recommended)

When invoked with `/exec-brief`, Claude reads `SKILL.md` and:
1. Loads configuration from `teammates.yaml` and `.exec-brief.yaml`
2. Queries data sources using available MCP tools:
   - Jira via `mcp__atlassian__jira_*` tools
   - GitHub via `gh` CLI
   - Google Docs via `mcp__google-docs__*` tools
3. Uses `exec_brief.py` for categorization logic
4. Generates formatted output organized by Eisenhower Matrix

### 2. Standalone Python Script

The `exec_brief.py` script provides:
- Configuration loading (`Config` class)
- Time range calculation (`TimeRangeCalculator` class)
- Item categorization (`ItemCategorizer` class)
- Output generation (`OutputGenerator` class)

The script is designed to be called by Claude Code, not used standalone.

## Configuration Files

### teammates.yaml (Required)

Defines you and your teammates with cross-platform identifiers.

**Location priority**:
1. `./teammates.yaml` (current directory)
2. `~/.config/exec-brief/teammates.yaml`

See `teammates.yaml.example` for full format.

### .exec-brief.yaml (Optional)

Configures information sources and matrix rules.

**Location priority**:
1. `./.exec-brief.yaml` (current directory)
2. `~/.config/exec-brief/config.yaml`

See `config.yaml.example` for all options.

**The skill works with sensible defaults without this file.**

## Eisenhower Matrix Categorization

Items are scored for urgency and importance, then assigned to quadrants:

### Urgency Score (0-10)
- Keywords: blocker, critical, urgent, emergency (+3)
- Priority: Blocker/Critical (+3), Major (+2)
- Due date: Overdue (+4), <24h (+3), <48h (+2)
- Teammate blocked (+2)
- CI failing (+2)

### Importance Score (0-10)
- Teammate involved (+3)
- Keywords: feature, security, strategic (+2)
- Priority: Blocker/Critical/Major (+2), High (+1)
- Security/CVE (+2)

### Quadrants
- **Q1 (Urgent & Important)**: urgency ≥ 3 AND importance ≥ 2
- **Q2 (Important, Not Urgent)**: urgency < 3 AND importance ≥ 2
- **Q3 (Urgent, Not Important)**: urgency ≥ 3 AND importance < 2
- **Q4 (Neither)**: urgency < 3 AND importance < 2

## Timezone Handling

The `--daily` flag considers activity across multiple timezones:

**EST/EDT** (America/New_York):
- Many systems timestamp activities in Eastern time
- Ensures East Coast business day is captured

**Home Timezone** (from `teammates.yaml`):
- Your configured timezone
- Ensures your local day is captured

**Combined Range**:
- Uses earliest start time and latest end time
- Ensures no activities are missed

**Example** (9:00 AM PST on March 19):
- EST range: March 18 00:00 EST → March 19 23:59 EDT
- PST range: March 18 00:00 PST → March 19 09:00 PST
- **Combined**: March 18 00:00 EST → March 19 09:00 PST

## Prerequisites

### Required
- **Python 3.8+**
- **teammates.yaml** configuration file

### Optional (for data sources)
- **Jira MCP Server** - For Jira integration
- **Google Docs MCP Server** - For Google Docs integration
- **GitHub CLI (`gh`)** - For GitHub integration

### Check Prerequisites

```bash
# Python version
python3 --version

# GitHub CLI
gh --version
gh auth status

# MCP servers (check in Claude Code)
# Available tools include: mcp__atlassian__*, mcp__google-docs__*
```

## Troubleshooting

### Configuration Issues

```bash
# Validate configuration
./validate_config.py

# Check file locations
ls -l teammates.yaml
ls -l .exec-brief.yaml
```

### Missing Dependencies

```bash
# Install dependencies
pip install -r requirements.txt

# Or individually
pip install pyyaml pytz
```

### No Data Sources Available

The skill requires at least one data source:
- Jira: Ensure MCP server is connected
- GitHub: Ensure `gh` CLI is authenticated
- Google Docs: Ensure MCP server is connected

Check available tools in Claude Code:
```
List available MCP tools
```

### Permission Issues

```bash
# Make scripts executable
chmod +x exec_brief.py validate_config.py
```

## Next Steps

1. **Review the design**: Read `exec-brief.md` for full design rationale
2. **Read user docs**: See `README.md` for usage examples
3. **Customize config**: Edit `.exec-brief.yaml` for your workflow
4. **Try it out**: Run `/exec-brief --daily` in Claude Code
5. **Integrate**: Add to your daily routine or hooks

## Support

For issues or questions:
1. Check `README.md` for usage examples
2. Validate configuration with `./validate_config.py`
3. Review `SKILL.md` for implementation details
4. Check Claude Code logs for errors

## Future Enhancements

See `README.md` for planned features:
- Slack integration
- Email digests
- Calendar integration
- AI-powered recommendations
- Historical trending
