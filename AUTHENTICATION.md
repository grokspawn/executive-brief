# Authentication Setup

All data sources require API tokens set as environment variables.

## Jira

1. **Create API token**: https://id.atlassian.com/manage-profile/security/api-tokens
2. **Set environment variable**:
   ```bash
   export JIRA_API_TOKEN=your-token-here
   ```

**Note**: Use your email address as the username with the API token (not your password).

## GitHub

Choose one option:

### Option 1: gh CLI (Recommended)
```bash
gh auth login
```

### Option 2: Personal Access Token
1. **Create token**: https://github.com/settings/tokens
   - Required scopes: `repo`, `read:user`
2. **Set environment variable**:
   ```bash
   export GITHUB_TOKEN=ghp_your_token_here
   ```

## Slack

### Option 1: Bot Token (Recommended)

1. **Create Slack App**: https://api.slack.com/apps
2. **Add OAuth Scopes**:
   - `channels:history`
   - `channels:read`
   - `groups:history`
   - `groups:read`
   - `im:history`
   - `im:read`
   - `mpim:history`
   - `mpim:read`
   - `users:read`
3. **Install app to workspace** and copy the Bot User OAuth Token (starts with `xoxb-`)
4. **Set environment variable**:
   ```bash
   export SLACK_XOXC_TOKEN=xoxb-your-token-here
   ```

### Option 2: User Token (For workspaces where you can't create apps)

⚠️ **Warning**: User tokens may be revoked by Slack if the user-agent doesn't match your browser session.

1. **Extract from browser** (see [SLACK_USER_TOKEN.md](SLACK_USER_TOKEN.md) for detailed instructions):
   - Token (xoxc-) from cookies
   - Cookie (xoxd-) from cookies
   - User-Agent from network requests

2. **Set environment variables**:
   ```bash
   export SLACK_XOXC_TOKEN='xoxc-your-token-here'
   export SLACK_XOXD_TOKEN='xoxd-your-cookie-here'
   export SLACK_USER_AGENT='Mozilla/5.0 (Macintosh; ...) ...'
   ```

### Alternative: File-based token

Configure in `exec-brief.yaml`:
```yaml
slack:
  token_source: file
  token_file: ~/.slack-token
```

Then save your token:
```bash
echo 'xoxb-your-token-here' > ~/.slack-token
chmod 600 ~/.slack-token
```

## Making Tokens Persistent

Add to your shell profile (`~/.bashrc`, `~/.zshrc`, etc.):

```bash
# Executive Brief API tokens
export JIRA_API_TOKEN=your-jira-token
export GITHUB_TOKEN=ghp_your-github-token
export SLACK_XOXC_TOKEN=xoxb-your-slack-token
```

Then reload:
```bash
source ~/.bashrc  # or ~/.zshrc
```

## Verification

Check that tokens are set:
```bash
echo $JIRA_API_TOKEN
echo $GITHUB_TOKEN
echo $SLACK_XOXC_TOKEN
```

## Troubleshooting

If authentication fails, the tool will display helpful error messages with:
- What's missing
- How to create the token
- How to set the environment variable
- Links to token creation pages

To disable a source temporarily:
```bash
./exec-brief --daily --sources jira,github  # Skip Slack
```
