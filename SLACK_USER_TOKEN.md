# Using Slack User Tokens

This guide explains how to use Slack with user tokens (xoxc-) instead of bot tokens (xoxb-).

## When to Use User Tokens

Use user tokens when:
- You don't have permission to create apps in your Slack workspace
- Your workspace admin won't create a bot for your team
- You want to access Slack as yourself rather than a bot

## Warning

⚠️ **User tokens with cookies can be revoked by Slack** if the User-Agent doesn't match your browser session. This is a security feature to prevent stolen credentials from being used. Make sure to use the exact User-Agent from your browser.

## Setup Instructions

### 1. Extract Token from Browser

While logged into Slack in your browser:

1. Open DevTools (F12 or Right-Click > Inspect)
2. Open console tab
3. Type "allow pasting" and press ENTER
4. Paste the following snippet and press ENTER to execute:
`JSON.parse(localStorage.localConfig_v2).teams[document.location.pathname.match(/^\/client\/([A-Z0-9]+)/)[1]].token`
5. Copy the entire value

### 2. Extract Cookie from Browser

In the same Cookies view:

1. Go to **Application** tab (Chrome) or **Storage** tab (Firefox)
2. Expand **Cookies** > `https://slack.com` or your workspace URL
3. Find the cookie named **d**
   - Value starts with `xoxd-`
   - Copy the entire value exactly as shown (may contain %2B, %2F, etc. - the code will decode it automatically)

### 3. Extract User-Agent from Browser

1. Go to **Network** tab in DevTools
2. Refresh the page or click any channel
3. Click on any request in the list
4. Go to **Headers** section
5. Find **User-Agent** in Request Headers
   - Looks like: `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36...`
   - Copy the entire value

### 4. Set Environment Variables

```bash
export SLACK_XOXC_TOKEN='xoxc-your-token-here'
export SLACK_XOXD_TOKEN='xoxd-your-cookie-here'
export SLACK_USER_AGENT='Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36'
```

### 5. Test Authentication

```bash
go run . --sources slack
```

Or add to your shell profile for persistence:

```bash
# Add to ~/.bashrc or ~/.zshrc
export SLACK_XOXC_TOKEN='xoxc-...'
export SLACK_XOXD_TOKEN='xoxd-...'
export SLACK_USER_AGENT='Mozilla/5.0 ...'
```

## Troubleshooting

### Invalid Auth Error

If you get `invalid_auth`:
- Make sure all three values are set (token, cookie, user-agent)
- Verify the cookie is properly copied (may need URL encoding)
- Check that your User-Agent exactly matches your browser

### Token Revoked / Forced Re-login

If Slack forces you to re-login:
- The User-Agent didn't match your browser session
- Extract all three values again from your browser after logging back in
- Make sure you're using the exact User-Agent from the same browser/session

### Best Practices

1. **Use Bot Tokens When Possible**: They're more stable and designed for API access
2. **Keep Tokens Secret**: Don't commit them to git or share them
3. **Refresh Regularly**: User tokens may expire; be prepared to re-extract them
4. **Match User-Agent**: Always use the exact User-Agent from your current browser session

## Configuration File

You can also configure custom environment variable names in `exec-brief.yaml`:

```yaml
slack:
  enabled: true
  token_source: env
  token_env_var: SLACK_XOXC_TOKEN
  cookie_env_var: SLACK_XOXD_COOKIE        # Optional, defaults to SLACK_XOXD_COOKIE
  user_agent_env_var: SLACK_USER_AGENT  # Optional, defaults to SLACK_USER_AGENT
```
