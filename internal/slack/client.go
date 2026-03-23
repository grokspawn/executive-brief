package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/grokspawn/executive-brief/internal/config"
	"github.com/grokspawn/executive-brief/internal/matrix"
	"github.com/rusq/slackdump/v3/auth"
	"github.com/slack-go/slack"
)

// Client represents a Slack API client
type Client struct {
	api *slack.Client
}

// userAgentTransport adds custom User-Agent to HTTP requests
type userAgentTransport struct {
	userAgent string
	base      http.RoundTripper
}

func (t *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.userAgent != "" {
		req.Header.Set("User-Agent", t.userAgent)
	}
	return t.base.RoundTrip(req)
}

// NewClient creates a new Slack API client
// Supports both bot tokens (xoxb-) and user tokens (xoxc-/xoxp-)
// User tokens require SLACK_XOXD_TOKEN and SLACK_USER_AGENT environment variables
func NewClient(token string) (*Client, error) {
	var api *slack.Client

	// User tokens (xoxc-) require a cookie and user-agent for authentication
	if strings.HasPrefix(token, "xoxc-") {
		cookie := os.Getenv("SLACK_XOXD_TOKEN")
		if cookie == "" {
			return nil, fmt.Errorf(`user token (xoxc-) requires SLACK_XOXD_TOKEN environment variable

To fix:
1. Extract the 'd' cookie value from your browser while logged into Slack:
   - Chrome/Firefox: DevTools > Application/Storage > Cookies > slack.com > d
   - The value will look like: xoxd-...
2. Set the environment variable:

   export SLACK_XOXD_TOKEN='your-xoxd-cookie-value-here'

Or add to your shell profile (~/.bashrc, ~/.zshrc)`)
		}

		userAgent := os.Getenv("SLACK_USER_AGENT")
		if userAgent == "" {
			return nil, fmt.Errorf(`user token (xoxc-) requires SLACK_USER_AGENT environment variable

To fix:
1. Extract the User-Agent from your browser while logged into Slack:
   - Chrome: DevTools > Network tab > Select any request > Headers > User-Agent
   - Firefox: DevTools > Network tab > Select any request > Headers > User-Agent
   - Safari: Develop > Show Web Inspector > Network > Select request > Headers
2. Set the environment variable:

   export SLACK_USER_AGENT='your-browser-user-agent-here'

Example User-Agent:
   Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36

WARNING: User tokens with cookies can be revoked by Slack if the user-agent doesn't match
your browser session. Make sure to use the exact user-agent from your browser.`)
		}

		// Use slackdump's auth.ValueAuth to handle cookie encoding properly
		provider, err := auth.NewValueAuth(token, cookie)
		if err != nil {
			return nil, fmt.Errorf("failed to create auth provider: %w", err)
		}

		// Get the HTTP client from the provider (includes proper cookie handling)
		providerClient, err := provider.HTTPClient()
		if err != nil {
			return nil, fmt.Errorf("failed to create HTTP client from auth provider: %w", err)
		}

		// Wrap the provider's transport with our custom User-Agent transport
		customClient := &http.Client{
			Transport: &userAgentTransport{
				userAgent: userAgent,
				base:      providerClient.Transport,
			},
			Timeout: providerClient.Timeout,
		}

		api = slack.New(token, slack.OptionHTTPClient(customClient))
	} else {
		// Bot token or legacy user token (xoxp-) - works without cookie
		api = slack.New(token)
	}

	return &Client{api: api}, nil
}

// ValidateAuth validates Slack authentication
func ValidateAuth(cfg *config.Config) error {
	// Load API token
	token, err := loadAPIToken(cfg)
	if err != nil {
		return err
	}

	// Test authentication
	client, err := NewClient(token)
	if err != nil {
		return err
	}
	ctx := context.Background()
	_, err = client.api.AuthTestContext(ctx)
	if err != nil {
		return fmt.Errorf("Slack authentication failed: %w", err)
	}
	return nil
}

// Query queries Slack for items within the time range
func Query(cfg *config.Config, startTime, endTime time.Time) ([]matrix.Item, error) {
	if !cfg.Sources.Slack.Enabled {
		return nil, nil
	}

	// Load API token
	token, err := loadAPIToken(cfg)
	if err != nil {
		return nil, fmt.Errorf("error loading Slack API token: %w", err)
	}

	client, err := NewClient(token)
	if err != nil {
		return nil, fmt.Errorf("error creating Slack client: %w", err)
	}
	ctx := context.Background()

	var items []matrix.Item

	// Get current user info
	authTest, err := client.api.AuthTestContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("error authenticating with Slack: %w", err)
	}
	currentUserID := authTest.UserID

	// Query configured channels
	if len(cfg.Sources.Slack.Channels) > 0 {
		channelItems, err := client.queryChannels(ctx, cfg, currentUserID, startTime, endTime)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to query Slack channels: %v\n", err)
		} else {
			items = append(items, channelItems...)
		}
	}

	// Query direct messages
	if cfg.Sources.Slack.IncludeDMs {
		dmItems, err := client.queryDirectMessages(ctx, cfg, currentUserID, startTime, endTime)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to query Slack DMs: %v\n", err)
		} else {
			items = append(items, dmItems...)
		}
	}

	return items, nil
}

// queryChannels queries messages in configured channels
func (c *Client) queryChannels(ctx context.Context, cfg *config.Config, currentUserID string, startTime, endTime time.Time) ([]matrix.Item, error) {
	var items []matrix.Item

	// Get all channels
	channels, _, err := c.api.GetConversationsContext(ctx, &slack.GetConversationsParameters{
		Types: []string{"public_channel", "private_channel"},
		Limit: 1000,
	})
	if err != nil {
		return nil, fmt.Errorf("error fetching channels: %w", err)
	}

	// Filter to configured channels
	channelMap := make(map[string]slack.Channel)
	for _, channel := range channels {
		for _, configChan := range cfg.Sources.Slack.Channels {
			if channel.Name == configChan || channel.ID == configChan {
				channelMap[channel.ID] = channel
				break
			}
		}
	}

	// Query each channel for messages
	for channelID, channel := range channelMap {
		channelItems, err := c.queryConversation(ctx, cfg, channel.Name, channelID, currentUserID, startTime, endTime)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to query channel %s: %v\n", channel.Name, err)
			continue
		}
		items = append(items, channelItems...)
	}

	return items, nil
}

// queryDirectMessages queries direct message conversations
func (c *Client) queryDirectMessages(ctx context.Context, cfg *config.Config, currentUserID string, startTime, endTime time.Time) ([]matrix.Item, error) {
	var items []matrix.Item

	// Get IM conversations
	conversations, _, err := c.api.GetConversationsContext(ctx, &slack.GetConversationsParameters{
		Types: []string{"im", "mpim"},
		Limit: 1000,
	})
	if err != nil {
		return nil, fmt.Errorf("error fetching DM conversations: %w", err)
	}

	// Query each conversation
	for _, conv := range conversations {
		convItems, err := c.queryConversation(ctx, cfg, "DM", conv.ID, currentUserID, startTime, endTime)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to query DM %s: %v\n", conv.ID, err)
			continue
		}
		items = append(items, convItems...)
	}

	return items, nil
}

// queryConversation queries messages in a specific conversation
func (c *Client) queryConversation(ctx context.Context, cfg *config.Config, channelName, channelID, currentUserID string, startTime, endTime time.Time) ([]matrix.Item, error) {
	var items []matrix.Item

	// Convert times to Slack timestamps
	oldest := fmt.Sprintf("%d.000000", startTime.Unix())
	latest := fmt.Sprintf("%d.000000", endTime.Unix())

	// Fetch conversation history
	params := &slack.GetConversationHistoryParameters{
		ChannelID: channelID,
		Oldest:    oldest,
		Latest:    latest,
		Limit:     100,
	}

	history, err := c.api.GetConversationHistoryContext(ctx, params)
	if err != nil {
		return nil, err
	}

	// Get teammate Slack IDs
	teammateIDs := make(map[string]string) // ID -> Name
	for _, tm := range cfg.Teammates {
		if tm.Slack != nil {
			switch v := tm.Slack.(type) {
			case string:
				if v != "" {
					teammateIDs[v] = tm.Name
				}
			case map[string]interface{}:
				if uid, ok := v["uid"].(string); ok && uid != "" {
					teammateIDs[uid] = tm.Name
				}
			}
		}
	}

	// Process messages
	for _, msg := range history.Messages {
		// Skip bot messages and system messages
		if msg.BotID != "" || msg.SubType == "channel_join" || msg.SubType == "channel_leave" {
			continue
		}

		// Parse timestamp
		msgTime, err := parseSlackTimestamp(msg.Timestamp)
		if err != nil {
			continue
		}

		// Check if message is in time range
		if msgTime.Before(startTime) || msgTime.After(endTime) {
			continue
		}

		// Check for mentions
		hasMention := false
		if cfg.Sources.Slack.IncludeMentions {
			hasMention = strings.Contains(msg.Text, fmt.Sprintf("<@%s>", currentUserID))
		}

		// Check for reactions
		hasReactions := false
		reactionCount := 0
		if msg.Reactions != nil {
			for _, reaction := range msg.Reactions {
				reactionCount += reaction.Count
			}
			hasReactions = reactionCount >= cfg.Sources.Slack.MinReactionCount
		}

		// Check if from teammate
		isTeammate := false
		teammateName := ""
		if name, ok := teammateIDs[msg.User]; ok {
			isTeammate = true
			teammateName = name
		}

		// Check if in thread
		isThread := msg.ThreadTimestamp != "" && msg.ThreadTimestamp != msg.Timestamp

		// Include if: mentioned, has reactions, from teammate, or is thread participant
		shouldInclude := hasMention || hasReactions || isTeammate
		if cfg.Sources.Slack.IncludeThreads && isThread {
			shouldInclude = true
		}

		if !shouldInclude {
			continue
		}

		// Create item
		labels := []string{}
		if hasMention {
			labels = append(labels, "mention")
		}
		if hasReactions {
			labels = append(labels, "reactions")
		}
		if isThread {
			labels = append(labels, "thread")
		}
		if channelName == "DM" {
			labels = append(labels, "direct_message")
		}

		// Get permalink
		permalink, _ := c.api.GetPermalinkContext(ctx, &slack.PermalinkParameters{
			Channel: channelID,
			Ts:      msg.Timestamp,
		})

		// Truncate message text for title
		title := msg.Text
		if len(title) > 100 {
			title = title[:97] + "..."
		}
		title = fmt.Sprintf("[%s] %s", channelName, title)

		item := matrix.Item{
			ID:        msg.Timestamp,
			Title:     title,
			Source:    "slack",
			URL:       permalink,
			Type:      "message",
			CreatedAt: msgTime,
			UpdatedAt: msgTime,
			Labels:    labels,
			Author:    msg.User,
		}

		// Set teammate involvement
		if isTeammate {
			item.TeammatesInvolved = []string{teammateName}
		}

		items = append(items, item)
	}

	return items, nil
}

// loadAPIToken loads the Slack API token from configuration
func loadAPIToken(cfg *config.Config) (string, error) {
	switch cfg.Sources.Slack.TokenSource {
	case "env":
		envVar := cfg.Sources.Slack.TokenEnvVar
		if envVar == "" {
			envVar = "SLACK_XOXC_TOKEN"
		}
		token := os.Getenv(envVar)
		if token == "" {
			return "", fmt.Errorf(`Slack authentication failed: environment variable %s not set

Option 1 - Bot Token (Recommended):
  1. Create a Slack App at: https://api.slack.com/apps
  2. Add OAuth scopes: channels:history, channels:read, groups:history, groups:read,
     im:history, im:read, mpim:history, mpim:read, users:read
  3. Install app to workspace and copy the Bot User OAuth Token (starts with xoxb-)
  4. Set the environment variable:
     export %s=xoxb-your-token-here

Option 2 - User Token (For workspaces where you can't create apps):
  1. Extract from browser while logged into Slack:
     - Token: DevTools > Application > Cookies > slack.com > token (xoxc-...)
     - Cookie: DevTools > Application > Cookies > slack.com > d (xoxd-...)
     - User-Agent: DevTools > Network > any request > Headers > User-Agent
  2. Set environment variables:
     export %s=xoxc-your-token-here
     export SLACK_XOXD_TOKEN=xoxd-your-cookie-here
     export SLACK_USER_AGENT='Mozilla/5.0 ...'

WARNING: User tokens may be revoked by Slack if the user-agent doesn't match your browser

Alternatively, use token_source: file in exec-brief.yaml`, envVar, envVar, envVar)
		}
		return token, nil

	case "file":
		if cfg.Sources.Slack.TokenFile == "" {
			return "", fmt.Errorf(`Slack authentication failed: token_file not configured

To fix:
1. Create a Slack App at: https://api.slack.com/apps
2. Add OAuth scopes and install to workspace
3. Save your Bot User OAuth Token to a file:

   echo 'xoxb-your-token-here' > ~/.slack-token

4. Configure in exec-brief.yaml:

   slack:
     token_source: file
     token_file: ~/.slack-token`)
		}
		data, err := os.ReadFile(cfg.Sources.Slack.TokenFile)
		if err != nil {
			return "", fmt.Errorf(`Slack authentication failed: cannot read token file %s

To fix:
1. Create a Slack App at: https://api.slack.com/apps
2. Add OAuth scopes and install to workspace
3. Save your Bot User OAuth Token to the file:

   echo 'xoxb-your-token-here' > %s

Error: %w`, cfg.Sources.Slack.TokenFile, cfg.Sources.Slack.TokenFile, err)
		}
		token := strings.TrimSpace(string(data))
		if token == "" {
			return "", fmt.Errorf(`Slack authentication failed: token file %s is empty

The file should contain your Slack Bot User OAuth Token (starts with xoxb-)`, cfg.Sources.Slack.TokenFile)
		}
		return token, nil

	default:
		return "", fmt.Errorf(`Slack authentication failed: unknown token_source "%s"

Valid options are:
- "env" (reads from environment variable)
- "file" (reads from file path)

Configure in exec-brief.yaml:
  slack:
    token_source: env
    token_env_var: SLACK_XOXC_TOKEN`, cfg.Sources.Slack.TokenSource)
	}
}

// parseSlackTimestamp parses a Slack timestamp string to time.Time
func parseSlackTimestamp(ts string) (time.Time, error) {
	// Slack timestamps are like "1234567890.123456"
	var sec, nsec int64
	_, err := fmt.Sscanf(ts, "%d.%d", &sec, &nsec)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(sec, nsec*1000), nil
}

// GetSlackID returns the Slack ID for a teammate
func GetSlackID(tm config.Teammate) string {
	if tm.Slack == nil {
		return ""
	}

	switch v := tm.Slack.(type) {
	case string:
		return v
	case map[string]interface{}:
		if uid, ok := v["uid"].(string); ok {
			return uid
		}
		// Try parsing as JSON
		data, err := json.Marshal(v)
		if err == nil {
			var slackInfo config.SlackInfo
			if err := json.Unmarshal(data, &slackInfo); err == nil {
				return slackInfo.UID
			}
		}
	}
	return ""
}
