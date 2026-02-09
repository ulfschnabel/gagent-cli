---
name: gagent-cli
description: "Google Workspace CLI for AI agents with dual OAuth scope separation (read/write). Use this skill when working with: (1) Gmail - reading, sending, or searching emails, (2) Calendar - scheduling or viewing events, (3) Contacts - managing contact information, (4) Google Docs - creating or editing documents, (5) Google Sheets - reading or writing spreadsheet data, (6) Google Slides - building or modifying presentations. Provides both high-level task commands and low-level API access."
---

# gagent-cli - Google Workspace for AI Agents

## Overview

gagent-cli provides command-line access to Google Workspace APIs (Gmail, Calendar, Contacts, Docs, Sheets, Slides) with dual OAuth scope separation for security.

**Key Features:**
- Separate read and write authorization
- JSON output for all commands
- Task commands (high-level) + API commands (low-level)
- Full Google Workspace API coverage

## Authentication Workflow

### Check Auth Status

```bash
gagent-cli auth status
```

Returns authorization state for both read and write scopes.

### Initial Setup (One-Time)

```bash
# 1. Setup wizard (configures OAuth credentials)
gagent-cli auth setup

# 2. Authorize read access (opens browser)
gagent-cli auth login --scope read

# 3. Authorize write access when needed (opens browser)
gagent-cli auth login --scope write
```

### Handling Auth Errors

When you encounter auth errors, guide the user:

**Error: `AUTH_REQUIRED`**
→ "I need read permissions. Please run: `gagent-cli auth login --scope read`"

**Error: `SCOPE_INSUFFICIENT`**
→ "I need write permissions to send emails. Please run: `gagent-cli auth login --scope write`"

**Error: `TOKEN_EXPIRED`**
→ "Token expired. Please re-authorize: `gagent-cli auth login --scope [read|write]`"

## Command Selection: Task vs API

**Use TASK commands** (high-level) when:
- Performing common operations
- User hasn't specified exact API parameters
- You want simplified output

**Use API commands** (low-level) when:
- You need full control over API parameters
- Task command doesn't support required fields
- Working with complex nested data

## Quick Command Reference

### Gmail
```bash
gagent-cli gmail inbox --limit 10
gagent-cli gmail search "from:user@example.com"
gagent-cli gmail read <message-id>
gagent-cli gmail reply <message-id> --body "text"  # Maintains thread!
gagent-cli gmail send --to ADDR --subject SUBJ --body BODY  # New thread
```

**CRITICAL**: Use `gmail reply` to keep emails in the same thread, not `gmail send`.

For detailed Gmail commands, see [gmail.md](references/gmail.md).

### Calendar
```bash
gagent-cli calendar today
gagent-cli calendar upcoming --days 7
gagent-cli calendar schedule --title "Meeting" --start "2026-02-05T14:00:00Z" --end "2026-02-05T15:00:00Z"
gagent-cli calendar reschedule <event-id> --start DT --end DT
```

**DateTime format**: Always use ISO 8601: `YYYY-MM-DDTHH:MM:SSZ`

For detailed Calendar commands, see [calendar.md](references/calendar.md).

### Contacts
```bash
gagent-cli contacts search "john smith"
gagent-cli contacts get people/c123456
gagent-cli contacts create --name "Name" --email "email@example.com"
gagent-cli contacts delete people/c123456
```

**Resource format**: Contacts use `people/c123456789` format.

For detailed Contacts commands, see [contacts.md](references/contacts.md).

### Docs, Sheets, Slides
```bash
# Docs
gagent-cli docs read <doc-id>
gagent-cli docs create --title "Title" [--content "text"]
gagent-cli docs append <doc-id> --text "content"

# Sheets
gagent-cli sheets read <spreadsheet-id> --sheet "Sheet1"
gagent-cli sheets write <sheet-id> --sheet "Sheet1" --range "A1" --values '[[1,2],[3,4]]'
gagent-cli sheets append <sheet-id> --sheet "Sheet1" --values '[["row","data"]]'

# Slides
gagent-cli slides info <presentation-id>
gagent-cli slides create --title "Title"
gagent-cli slides add-text <pres-id> --slide 1 --text "text" --x 100 --y 100
```

**CRITICAL for Slides**: Use visual feedback loop! Export to PDF after changes to see actual rendering.

For detailed Docs/Sheets/Slides commands, see [docs-sheets-slides.md](references/docs-sheets-slides.md).

## JSON Output Format

### Success Response

```json
{
  "success": true,
  "data": { ... },
  "metadata": {
    "scope_used": "read",
    "timestamp": "2026-02-03T10:30:00Z",
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

**Parsing logic:**
1. Check `success` field
2. If `true`, extract `data`
3. If `false`, read `error.code` and `error.message`
4. Use `metadata.scope_used` to understand permissions used

## Common Error Codes

| Code | Meaning | Agent Action |
|------|---------|--------------|
| `AUTH_REQUIRED` | No token found | Ask user to run `auth login` |
| `SCOPE_INSUFFICIENT` | Need write but have only read | Ask user to run `auth login --scope write` |
| `TOKEN_EXPIRED` | Token refresh failed | Ask user to re-run `auth login` |
| `NOT_FOUND` | Resource doesn't exist | Verify ID and inform user |
| `RATE_LIMITED` | API quota exceeded | Wait and retry, or inform user |
| `INVALID_INPUT` | Bad command arguments | Check command syntax |

## Best Practices

### 1. Always Specify Limits

```bash
# Good
gagent-cli gmail inbox --limit 10

# Avoid (returns too much data)
gagent-cli gmail inbox
```

### 2. Confirm Before Destructive Operations

Always confirm with user before:
- Deleting emails, contacts, or calendar events
- Sending emails (unless explicitly instructed)
- Modifying documents, sheets, or slides

### 3. Use Appropriate Search Queries

```bash
# Good: Specific query
gagent-cli gmail search "from:boss@company.com subject:urgent after:2026/01/01" --limit 5

# Less optimal: Broad query
gagent-cli gmail search "urgent" --limit 100
```

### 4. Handle DateTime Carefully

- Always use ISO 8601 format
- Ask user about timezone if unclear
- For all-day events, use date-only: `2026-02-03`

### 5. Preserve Email Threading

- Use `gmail reply` to respond in thread
- Use `gmail send` only for new conversations

### 6. Visual Feedback for Slides

When building presentations:
1. Make changes
2. Export to PDF: `gagent-cli slides export <pres-id> --format pdf --output /tmp/check.pdf`
3. Read the PDF to see rendered output
4. Fix issues (overlaps, positioning, sizing)
5. Repeat until satisfied

## Common Workflows

### Process Unread Emails

```bash
# 1. Get unread emails
gagent-cli gmail search "is:unread" --limit 20

# 2. Read each email
gagent-cli gmail read <message-id>

# 3. Take action (reply, forward, etc.)
gagent-cli gmail reply <message-id> --body "Response"
```

### Schedule Meeting

```bash
# 1. Check availability
gagent-cli calendar free-busy --start "2026-02-05T14:00:00Z" --end "2026-02-05T16:00:00Z"

# 2. If free, schedule
gagent-cli calendar schedule \
  --title "Team Sync" \
  --start "2026-02-05T14:00:00Z" \
  --end "2026-02-05T15:00:00Z" \
  --attendees "team@example.com"
```

### Build Styled Slide Deck

```bash
# 1. Create presentation
gagent-cli slides create --title "Project Overview"
# Save presentation_id from response

# 2. Add slides with content
gagent-cli slides add-slide <pres-id> --layout BLANK
gagent-cli slides add-text <pres-id> --slide 1 --text "Title" --x 100 --y 50

# 3. Export to PDF and check visually
gagent-cli slides export <pres-id> --format pdf --output /tmp/check.pdf
# Read /tmp/check.pdf to see rendering

# 4. Fix issues found using batch-update
gagent-cli slides api batch-update <pres-id> --requests-json '[...]'

# 5. Repeat steps 3-4 until satisfied
```

## Security Considerations

**Never:**
- Store tokens (gagent-cli handles token storage)
- Share OAuth credentials
- Bypass user confirmation for destructive operations

**Always:**
- Use read scope when possible
- Ask permission for write operations
- Confirm destructive operations twice
- Preview changes before applying

## Reference Files

For detailed command syntax and examples:

- **[gmail.md](references/gmail.md)** - Email operations, search syntax, threading
- **[calendar.md](references/calendar.md)** - Event management, datetime formatting
- **[contacts.md](references/contacts.md)** - Contact operations, resource names
- **[docs-sheets-slides.md](references/docs-sheets-slides.md)** - Document creation/editing, visual feedback loop

## Configuration

Config directory: `~/.config/gagent-cli/`
- `config.json` - OAuth credentials and preferences
- `token_read.json` - Read OAuth token
- `token_write.json` - Write OAuth token

All files have 0600 permissions (user read/write only).

## Rate Limits

Google APIs have rate limits:
- **Gmail**: 250 quota units/user/second
- **Calendar**: 1,000,000 queries/day
- **Contacts**: 600 requests/minute/user
- **Docs/Sheets/Slides**: 300 requests/minute/user

If rate limited, wait 60 seconds and retry.
