# gagent-cli

A CLI tool that gives AI agents safe, scoped access to Google Workspace APIs (Gmail, Calendar, Contacts, Docs, Sheets, and Slides).

## Features

- **Dual OAuth Scopes**: Read and write operations require separate authorization
- **JSON Output**: Structured JSON output for agent consumption
- **Task + API Commands**: High-level task commands and low-level API access
- **Cross-Platform**: Linux and macOS support
- **Full Google Workspace API Coverage**: Gmail, Calendar, Contacts, Docs, Sheets, and Slides

## For AI Agents

**→ See [AGENTS.md](./AGENTS.md) for detailed agent-specific guidance, best practices, and common workflows.**

## Installation

```bash
# Build from source
go build -o gagent-cli ./cmd/gagent-cli
```

## Quick Start

### 1. Setup

Run the interactive setup wizard to configure your Google Cloud OAuth credentials:

```bash
gagent-cli auth setup
```

### 2. Authorize

Authorize read access (opens browser):

```bash
gagent-cli auth login --scope read
```

Optionally authorize write access:

```bash
gagent-cli auth login --scope write
```

### 3. Use

```bash
# Check inbox
gagent-cli gmail inbox --limit 5

# Read a message
gagent-cli gmail read <message-id>

# List today's events
gagent-cli calendar today

# Read a document
gagent-cli docs read <doc-id>

# Read spreadsheet data
gagent-cli sheets read <spreadsheet-id> --sheet "Sheet1"

# Get presentation info
gagent-cli slides info <presentation-id>
```

## Commands

### Authentication

```bash
gagent-cli auth setup                  # Interactive setup wizard
gagent-cli auth login --scope read     # Authorize read access
gagent-cli auth login --scope write    # Authorize write access
gagent-cli auth status                 # Show current auth status
gagent-cli auth revoke --scope read    # Revoke read token
gagent-cli auth revoke --scope write   # Revoke write token
```

### Gmail

```bash
# Task commands (high-level)
gagent-cli gmail inbox [--limit N] [--unread-only]
gagent-cli gmail read <message-id>
gagent-cli gmail search <query> [--limit N]
gagent-cli gmail thread <thread-id>
gagent-cli gmail send --to ADDR --subject SUBJ --body BODY
gagent-cli gmail reply <message-id> --body BODY [--reply-all]
gagent-cli gmail forward <message-id> --to ADDR [--body BODY]
gagent-cli gmail draft --to ADDR --subject SUBJ --body BODY

# API commands (low-level)
gagent-cli gmail api list [--label LABEL] [--query QUERY]
gagent-cli gmail api get <message-id>
gagent-cli gmail api labels
```

### Calendar

```bash
# Task commands
gagent-cli calendar today [--calendar ID]
gagent-cli calendar week [--calendar ID]
gagent-cli calendar upcoming [--days N]
gagent-cli calendar event <event-id>
gagent-cli calendar free-busy --start DATETIME --end DATETIME
gagent-cli calendar schedule --title TITLE --start DT --end DT [--attendees EMAILS]
gagent-cli calendar reschedule <event-id> --start DT --end DT
gagent-cli calendar cancel <event-id> [--notify]
gagent-cli calendar respond <event-id> --status accepted|declined|tentative

# API commands
gagent-cli calendar api calendars
gagent-cli calendar api events [--time-min DT] [--time-max DT]
```

### Contacts

```bash
# Task commands (simplified, user-friendly)
gagent-cli contacts list [--limit N] [--query QUERY]
gagent-cli contacts get <resource-name>
gagent-cli contacts search <query> [--limit N]
gagent-cli contacts groups
gagent-cli contacts create --name NAME [--email EMAIL] [--phone PHONE]
gagent-cli contacts update <resource-name> --json '{"displayName":"New Name",...}'
gagent-cli contacts delete <resource-name>

# API commands (low-level, full People API access)
gagent-cli contacts api list [--page-size N] [--page-token TOKEN]
gagent-cli contacts api get <resource-name>
gagent-cli contacts api create --person-json '{...}'
gagent-cli contacts api update <resource-name> --person-json '{...}' --etag ETAG
gagent-cli contacts api delete <resource-name>
```

### Docs

```bash
# Task commands
gagent-cli docs list [--limit N] [--query QUERY]
gagent-cli docs read <doc-id>
gagent-cli docs export <doc-id> --format txt|html|pdf
gagent-cli docs outline <doc-id>
gagent-cli docs create --title TITLE [--content TEXT]
gagent-cli docs append <doc-id> --text TEXT
gagent-cli docs replace-text <doc-id> --find TEXT --replace TEXT
gagent-cli docs update-section <doc-id> --heading HEADING --content TEXT

# Rich formatting commands
gagent-cli docs append-formatted <doc-id> --text TEXT [--bold] [--italic] [--underline] [--color COLOR] [--style STYLE]
gagent-cli docs insert-list <doc-id> --type TYPE --items ITEM1,ITEM2 [--indent N]
gagent-cli docs format-paragraph <doc-id> --start INDEX --end INDEX [--align ALIGN] [--line-spacing N]
gagent-cli docs insert-table <doc-id> --rows N --cols N [--headers H1,H2] [--csv DATA]
gagent-cli docs insert-pagebreak <doc-id>
gagent-cli docs insert-hr <doc-id>
gagent-cli docs insert-toc <doc-id>
gagent-cli docs format-template <doc-id> --template-file FILE.json

# API commands
gagent-cli docs api get <doc-id>
gagent-cli docs api batch-update <doc-id> --requests-json JSON
```

#### Rich Formatting Examples

**Formatted text with bold and color:**
```bash
gagent-cli docs append-formatted <doc-id> \
  --text "TOTAL AMOUNT: €3,523.59" \
  --bold --color "#ff0000" --font-size 14
```

**Insert a heading:**
```bash
gagent-cli docs append-formatted <doc-id> \
  --text "Financial Summary" \
  --style heading1
```

**Create a bulleted list:**
```bash
gagent-cli docs insert-list <doc-id> \
  --type bullet \
  --items "First item,Second item,Third item"
```

**Insert a table with headers:**
```bash
gagent-cli docs insert-table <doc-id> \
  --rows 3 --cols 3 \
  --headers "Item,Amount,Total"
```

**Format from JSON template:**
```json
{
  "title": {
    "text": "Tax Document 2024",
    "style": {
      "namedStyle": "title"
    }
  },
  "sections": [
    {
      "heading": "Income Summary",
      "style": "heading1",
      "content": [
        {
          "type": "text",
          "text": "Total income for the year:",
          "style": {"bold": true}
        },
        {
          "type": "table",
          "rows": 3,
          "columns": 2,
          "headers": ["Category", "Amount"],
          "table_data": [
            ["Salary", "€50,000"],
            ["Investments", "€5,000"]
          ]
        }
      ]
    }
  ]
}
```

```bash
gagent-cli docs format-template <doc-id> --template-file tax-doc.json
```

### Sheets

```bash
# Task commands
gagent-cli sheets list [--limit N]
gagent-cli sheets read <spreadsheet-id> [--sheet NAME] [--range A1:Z100]
gagent-cli sheets info <spreadsheet-id>
gagent-cli sheets export <spreadsheet-id> --format csv|xlsx|pdf
gagent-cli sheets create --title TITLE
gagent-cli sheets write <spreadsheet-id> --sheet NAME --range A1 --values '[[1,2],[3,4]]'
gagent-cli sheets append <spreadsheet-id> --sheet NAME --values '[[1,2,3]]'
gagent-cli sheets clear <spreadsheet-id> --sheet NAME --range A1:Z100
gagent-cli sheets add-sheet <spreadsheet-id> --name "New Sheet"

# API commands
gagent-cli sheets api get <spreadsheet-id>
gagent-cli sheets api values <spreadsheet-id> --range "Sheet1!A1:Z100"
gagent-cli sheets api batch-update <spreadsheet-id> --requests-json JSON
```

### Slides

```bash
# Task commands
gagent-cli slides list [--limit N]
gagent-cli slides info <presentation-id>
gagent-cli slides read <presentation-id> [--slide N]
gagent-cli slides export <presentation-id> --format pdf|pptx
gagent-cli slides text <presentation-id>
gagent-cli slides create --title TITLE
gagent-cli slides add-slide <presentation-id> [--layout BLANK|TITLE|...]
gagent-cli slides delete-slide <presentation-id> --slide N
gagent-cli slides update-text <presentation-id> --slide N --find TEXT --replace TEXT
gagent-cli slides add-text <presentation-id> --slide N --text TEXT --x 100 --y 100

# API commands
gagent-cli slides api get <presentation-id>
gagent-cli slides api batch-update <presentation-id> --requests-json JSON
```

## JSON Output Format

### Success Response

```json
{
  "success": true,
  "data": { ... },
  "metadata": {
    "scope_used": "read",
    "timestamp": "2024-01-15T10:30:00Z",
    "request_id": "uuid"
  }
}
```

### Error Response

```json
{
  "success": false,
  "error": {
    "code": "AUTH_REQUIRED",
    "message": "Write scope not authorized. Run: gagent-cli auth login --scope write",
    "details": { ... }
  }
}
```

### Error Codes

- `AUTH_REQUIRED` - Need to run auth login
- `SCOPE_INSUFFICIENT` - Have read but need write (or vice versa)
- `TOKEN_EXPIRED` - Token refresh failed, re-auth needed
- `RATE_LIMITED` - Google API rate limit hit
- `NOT_FOUND` - Resource doesn't exist
- `INVALID_INPUT` - Bad command arguments
- `API_ERROR` - Google API error

## Configuration

Configuration is stored in `~/.config/gagent-cli/`:

- `config.json` - OAuth credentials and preferences
- `token_read.json` - Read-only OAuth token
- `token_write.json` - Write OAuth token

### Config Options

```bash
gagent-cli config set redirect_url "http://localhost:12345/oauth2callback"
gagent-cli config set default_calendar "work@group.calendar.google.com"
gagent-cli config set audit_log true
gagent-cli config get redirect_url
gagent-cli config get default_calendar
```

**Note on redirect_url**: If you encounter OAuth redirect_uri_mismatch errors, configure a custom redirect URL that matches what's registered in your Google Cloud Console. The redirect URL must include the full host, port, and path (e.g., `http://localhost:12345/oauth2callback`). If not set, the CLI will use a dynamic port with `http://127.0.0.1:<random-port>/callback`.

## Safety Features

- **Scope Separation**: Read and write require separate authorization
- **Dry Run Mode**: Use `--dry-run` flag for write operations to preview changes
- **File Permissions**: All config and token files use 0600 permissions

## Agent Skill

An agent skill is available for AI assistants like Claude to use gagent-cli effectively. The skill includes:

- Authentication workflow guidance
- Command selection (task vs API)
- Domain-specific references (Gmail, Calendar, Contacts, Docs, Sheets, Slides)
- Best practices and common workflows
- Visual feedback loop for Slides

**Install the skill:**
```bash
# Install the packaged skill
cp gagent-cli.skill ~/.claude/skills/
```

**Build from source:**
```bash
cd skill && ./package.sh
```

See **[skill/](skill/)** directory for source files.

## Development

```bash
# Run tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Build locally
go build -o gagent-cli ./cmd/gagent-cli
```

## Releases

Releases are managed with [GoReleaser](https://goreleaser.com/). To create a new release:

```bash
# Install GoReleaser (if not already installed)
brew install goreleaser

# Tag the release
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0

# Create the release (requires GITHUB_TOKEN env var)
goreleaser release --clean
```

GoReleaser will automatically:
- Build binaries for Linux, macOS, and Windows (amd64 and arm64)
- Create archives with proper naming (e.g., `gagent-cli_v0.1.0_Darwin_arm64.tar.gz`)
- Upload to GitHub Releases
- Generate checksums

For testing the release process locally without publishing:
```bash
goreleaser release --snapshot --clean
```

## License

MIT
