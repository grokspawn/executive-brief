# Architecture

## Pluggable Source Design

Executive-brief uses a pluggable architecture for data sources via the `Source` interface:

```go
type Source interface {
    Name() string
    Enabled(cfg *config.Config) bool
    Query(cfg *config.Config, startTime, endTime time.Time) ([]matrix.Item, error)
}
```

### Key Principles

1. **No source-specific knowledge in common code** - main.go and shared packages interact only through the interface
2. **Sources are self-contained** - each plugin handles its own API, authentication, and teammate identification
3. **Simple registration** - add new sources by implementing the interface and registering in main.go

### Current Sources

- **Jira** - Queries via REST API, identifies teammates by email
- **GitHub** - Queries via GitHub API, identifies teammates by username
- **Slack** - Queries via Slack API, identifies teammates by user ID

### Adding a New Source

1. Create `internal/newsource/` with `client.go` and `source.go`
2. Implement the `Source` interface
3. Add configuration struct to `internal/config/config.go`
4. Register in `main.go`: `registry.Register(&newsource.NewSource{})`
5. Handle teammate identification internally within the source

### Data Flow

```
Registry.QueryAll() → [Jira, GitHub, Slack].Query()
→ []matrix.Item → Categorize() → Output
```

Each source converts platform data to `matrix.Item` and populates `TeammatesInvolved` field.
