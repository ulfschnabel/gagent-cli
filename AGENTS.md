# Agent Guide for gagent-cli

This guide provides specific instructions for AI agents using gagent-cli to interact with Google Workspace APIs.

## Quick Reference for Agents

### When to Use Task vs API Commands

**Use TASK commands** (high-level) when:
- Performing common operations (reading email, scheduling meetings)
- You want simplified, agent-friendly output
- User hasn't specified exact API parameters

**Use API commands** (low-level) when:
- You need full control over API parameters
- Task command doesn't support required fields
- Working with complex nested data structures
- User explicitly requests API-level access

### Command Selection Guide

```
User Request: "Check my inbox"
→ gagent-cli gmail inbox --limit 10

User Request: "Read email 19c1b129dd8ca32e"
→ gagent-cli gmail read 19c1b129dd8ca32e

User Request: "Search for emails from john@example.com"
→ gagent-cli gmail search "from:john@example.com"

User Request: "Reply to that email"
→ gagent-cli gmail reply <message-id> --body "Your reply text"

User Request: "What's on my calendar today?"
→ gagent-cli calendar today

User Request: "Schedule a meeting tomorrow at 2pm"
→ gagent-cli calendar schedule --title "Meeting" --start "2026-02-03T14:00:00Z" --end "2026-02-03T15:00:00Z"

User Request: "Find contact named Sarah"
→ gagent-cli contacts search "Sarah"

User Request: "Delete contact people/c123456"
→ gagent-cli contacts delete people/c123456
```

## Gmail: Agent Best Practices

### Reading Email

```bash
# 1. List recent emails
gagent-cli gmail inbox --limit 10

# 2. Search with Gmail query syntax
gagent-cli gmail search "subject:invoice after:2026/01/01" --limit 20

# 3. Read full message with attachments
gagent-cli gmail read <message-id>

# 4. Get conversation thread
gagent-cli gmail thread <thread-id>
```

### Sending Email

**IMPORTANT FOR THREADING:**
- Use `gmail reply` to keep emails in the same conversation thread
- `gmail send` and `gmail draft` create NEW standalone emails
- `gmail reply` automatically sets In-Reply-To and References headers

```bash
# Reply to maintain thread (RECOMMENDED for conversations)
gagent-cli gmail reply <message-id> --body "Your reply" [--reply-all]

# Send new email (creates new thread)
gagent-cli gmail send --to "user@example.com" --subject "New Topic" --body "Content"

# Create draft for user review
gagent-cli gmail draft --to "user@example.com" --subject "Draft" --body "Content"

# Forward with context
gagent-cli gmail forward <message-id> --to "colleague@example.com" --body "FYI"
```

### Gmail Search Queries

Use Gmail's powerful search syntax:

```bash
# Find unread emails from specific sender
gagent-cli gmail search "from:boss@company.com is:unread"

# Emails with attachments in date range
gagent-cli gmail search "has:attachment after:2026/01/01 before:2026/02/01"

# Important emails with specific subject
gagent-cli gmail search "is:important subject:meeting"

# Exclude certain senders
gagent-cli gmail search "subject:report -from:spam@example.com"
```

**Common search operators:**
- `from:` - Sender email
- `to:` - Recipient email
- `subject:` - Subject line
- `after:`, `before:` - Date filters (YYYY/MM/DD)
- `has:attachment` - Has attachments
- `is:unread`, `is:read` - Read status
- `is:important`, `is:starred` - Labels
- `label:` - Specific label

## Calendar: Agent Best Practices

### Viewing Events

```bash
# Quick views
gagent-cli calendar today
gagent-cli calendar week
gagent-cli calendar upcoming --days 7

# Specific event details
gagent-cli calendar event <event-id>

# Check availability
gagent-cli calendar free-busy --start "2026-02-03T09:00:00Z" --end "2026-02-03T17:00:00Z"
```

### Creating and Managing Events

```bash
# Schedule meeting with attendees
gagent-cli calendar schedule \
  --title "Team Sync" \
  --start "2026-02-05T14:00:00Z" \
  --end "2026-02-05T15:00:00Z" \
  --attendees "alice@example.com,bob@example.com"

# Reschedule existing event
gagent-cli calendar reschedule <event-id> \
  --start "2026-02-05T15:00:00Z" \
  --end "2026-02-05T16:00:00Z"

# Cancel with notification
gagent-cli calendar cancel <event-id> --notify

# Respond to invitation
gagent-cli calendar respond <event-id> --status accepted
```

### DateTime Format

Always use ISO 8601 format: `YYYY-MM-DDTHH:MM:SSZ`

Examples:
- `2026-02-03T14:00:00Z` - 2pm UTC on Feb 3, 2026
- `2026-02-03T14:00:00-05:00` - 2pm EST on Feb 3, 2026
- `2026-02-03` - All-day event on Feb 3, 2026

## Contacts: Agent Best Practices

### Finding Contacts

```bash
# List all contacts (paginated)
gagent-cli contacts list --limit 50

# Search by name, email, or phone
gagent-cli contacts search "john smith"
gagent-cli contacts search "john@example.com"
gagent-cli contacts search "+1234567890"

# Get specific contact
gagent-cli contacts get people/c123456789

# List contact groups
gagent-cli contacts groups
```

### Managing Contacts

```bash
# Create simple contact
gagent-cli contacts create \
  --name "John Smith" \
  --email "john@example.com" \
  --phone "+1-555-0100"

# Update contact (use API command for complex updates)
gagent-cli contacts api update people/c123456 \
  --person-json '{"names":[{"givenName":"John","familyName":"Doe"}],"emailAddresses":[{"value":"john.doe@example.com"}]}' \
  --etag "etag-value-from-get"

# Delete contact
gagent-cli contacts delete people/c123456
```

### Important Notes on Contacts

1. **Resource Names**: Contacts use format `people/c123456789`
2. **ETags Required**: Updates require the current etag (get it with `contacts get`)
3. **Birthday Events**: Deleting a contact also removes associated calendar birthday events
4. **Search is Fuzzy**: Search matches partial names, emails, and phone numbers

## Docs, Sheets, Slides: Agent Best Practices

### Google Docs

```bash
# Read document content
gagent-cli docs read <doc-id>

# Get document outline/structure
gagent-cli docs outline <doc-id>

# Create new document
gagent-cli docs create --title "Meeting Notes" --content "Initial content"

# Append to document
gagent-cli docs append <doc-id> --text "\n\nNew section content"

# Find and replace
gagent-cli docs replace-text <doc-id> --find "old text" --replace "new text"

# Update specific section
gagent-cli docs update-section <doc-id> --heading "Budget" --content "Updated budget info"
```

### Google Sheets

```bash
# Read spreadsheet
gagent-cli sheets read <spreadsheet-id> --sheet "Sheet1" --range "A1:Z100"

# Create spreadsheet
gagent-cli sheets create --title "Q1 Report"

# Write data (2D array in JSON)
gagent-cli sheets write <spreadsheet-id> \
  --sheet "Sheet1" \
  --range "A1" \
  --values '[[" Name","Email"],["John","john@example.com"],["Jane","jane@example.com"]]'

# Append rows
gagent-cli sheets append <spreadsheet-id> \
  --sheet "Sheet1" \
  --values '[["New","Data","Here"]]'

# Clear range
gagent-cli sheets clear <spreadsheet-id> --sheet "Sheet1" --range "A1:Z100"
```

### Google Slides

```bash
# Read presentation
gagent-cli slides info <presentation-id>
gagent-cli slides read <presentation-id> --slide 1

# Extract all text
gagent-cli slides text <presentation-id>

# Create presentation
gagent-cli slides create --title "Q1 Results"

# Add slide
gagent-cli slides add-slide <presentation-id> --layout "TITLE_AND_BODY"

# Update text
gagent-cli slides update-text <presentation-id> \
  --slide 1 \
  --find "{{date}}" \
  --replace "2026-02-03"
```

## Authentication Flow for Agents

### Initial Setup (One-Time)

```bash
# 1. User runs setup wizard
gagent-cli auth setup

# 2. Authorize read access
gagent-cli auth login --scope read

# 3. Authorize write access (optional, when needed)
gagent-cli auth login --scope write
```

### Handling Auth Errors

When you encounter auth errors, guide the user:

```json
{
  "success": false,
  "error": {
    "code": "AUTH_REQUIRED",
    "message": "Write scope not authorized. Run: gagent-cli auth login --scope write"
  }
}
```

**Agent Response:**
"I need write permissions to send emails. Please run: `gagent-cli auth login --scope write`"

### Checking Auth Status

```bash
gagent-cli auth status
```

Returns current authorization state for both read and write scopes.

## Error Handling for Agents

### Common Error Codes

| Code | Meaning | Agent Action |
|------|---------|--------------|
| `AUTH_REQUIRED` | No token found | Ask user to run `auth login` |
| `SCOPE_INSUFFICIENT` | Need write but have only read | Ask user to run `auth login --scope write` |
| `TOKEN_EXPIRED` | Token refresh failed | Ask user to re-run `auth login` |
| `NOT_FOUND` | Resource doesn't exist | Verify ID and inform user |
| `RATE_LIMITED` | API quota exceeded | Wait and retry, or inform user |
| `INVALID_INPUT` | Bad command arguments | Check command syntax |

### Parsing JSON Output

All commands return structured JSON:

```json
{
  "success": true,
  "data": {
    // Command-specific data
  },
  "metadata": {
    "scope_used": "read",
    "timestamp": "2026-02-03T10:30:00Z",
    "request_id": "uuid"
  }
}
```

**Agent parsing logic:**
1. Check `success` field
2. If `true`, extract data from `data` field
3. If `false`, read `error.code` and `error.message`
4. Use `metadata.scope_used` to understand what permission was used

## Best Practices for Agents

### 1. Always Specify Limits

```bash
# Good: Limit results
gagent-cli gmail inbox --limit 10

# Avoid: No limit (returns too much data)
gagent-cli gmail inbox
```

### 2. Use Appropriate Search Queries

```bash
# Good: Specific query
gagent-cli gmail search "from:boss@company.com subject:urgent after:2026/01/01" --limit 5

# Less optimal: Broad query
gagent-cli gmail search "urgent" --limit 100
```

### 3. Confirm Before Destructive Operations

Always confirm with user before:
- Deleting emails, contacts, or calendar events
- Sending emails (unless explicitly instructed)
- Modifying documents, sheets, or slides

### 4. Handle Date/Time Carefully

- Always use ISO 8601 format for datetime
- Consider user's timezone (ask if unclear)
- For all-day events, use date-only format: `2026-02-03`

### 5. Preserve Threading in Email

- Use `gmail reply` to respond to emails in thread
- Use `gmail send` only for new conversations
- Explain the difference if user seems confused

### 6. Resource IDs vs Resource Names

- **Gmail**: Uses message IDs (e.g., `19c1b129dd8ca32e`)
- **Calendar**: Uses event IDs (e.g., `eventid123_20260203`)
- **Contacts**: Uses resource names (e.g., `people/c123456789`)
- **Docs/Sheets/Slides**: Uses document IDs (e.g., from URL)

### 7. Working with Attachments

```bash
# Download attachment
gagent-cli gmail api attachment <message-id> <attachment-id> --save-to /path/to/file

# Read PDFs and images
# Use other tools to process the downloaded files
```

## Common Agent Workflows

### Workflow: Process Unread Emails

```bash
# 1. Get unread emails
gagent-cli gmail search "is:unread" --limit 20

# 2. For each email, read content
gagent-cli gmail read <message-id>

# 3. Take action (reply, forward, archive, etc.)
gagent-cli gmail reply <message-id> --body "Response"
```

### Workflow: Schedule Meeting

```bash
# 1. Check availability
gagent-cli calendar free-busy \
  --start "2026-02-05T14:00:00Z" \
  --end "2026-02-05T16:00:00Z"

# 2. If free, schedule meeting
gagent-cli calendar schedule \
  --title "Team Sync" \
  --start "2026-02-05T14:00:00Z" \
  --end "2026-02-05T15:00:00Z" \
  --attendees "team@example.com"
```

### Workflow: Contact Cleanup

```bash
# 1. List all contacts
gagent-cli contacts list --limit 100

# 2. For each contact, check if it should be kept
# Agent analyzes the data

# 3. Delete unwanted contacts
gagent-cli contacts delete people/c123456

# 4. Update contacts that need changes
gagent-cli contacts api update people/c789 --person-json '{...}' --etag "..."
```

### Workflow: Document Collaboration

```bash
# 1. Read current document
gagent-cli docs read <doc-id>

# 2. Make changes
gagent-cli docs append <doc-id> --text "\n\n## New Section\nContent here"

# 3. Or update specific sections
gagent-cli docs update-section <doc-id> \
  --heading "Budget" \
  --content "Updated Q1 budget: $50k"
```

## Security Considerations

### What Agents Should Never Do

1. **Never store tokens** - Let gagent-cli handle token storage
2. **Never share OAuth credentials** - Each user has their own
3. **Never bypass user confirmation** for:
   - Sending emails
   - Deleting data
   - Sharing documents
   - Granting calendar access

### Safe Patterns

- Read operations are generally safe (use `read` scope)
- Preview changes before applying (describe what will happen)
- Ask permission for write operations
- Confirm destructive operations twice

## Troubleshooting Guide for Agents

### User Reports: "Command not working"

1. Check auth status: `gagent-cli auth status`
2. Verify command syntax in this guide
3. Check for error in JSON output
4. Look at `error.code` for specific issue

### User Reports: "Can't send email"

- Verify write scope: `gagent-cli auth status`
- If not authorized, guide user: `gagent-cli auth login --scope write`
- Check recipient email format
- Ensure subject and body are provided

### User Reports: "Contact not found"

- Verify resource name format: `people/c123456789`
- Try searching: `gagent-cli contacts search "name"`
- Contact might have been deleted

### User Reports: "Email not in thread"

- Did you use `gmail send` instead of `gmail reply`?
- Explain: "To reply in thread, use: `gagent-cli gmail reply <message-id> --body "text"`"

## File Paths and Configuration

- Config directory: `~/.config/gagent-cli/`
- Read token: `~/.config/gagent-cli/token_read.json`
- Write token: `~/.config/gagent-cli/token_write.json`
- OAuth config: `~/.config/gagent-cli/config.json`

All files have 0600 permissions (user read/write only).

## API Rate Limits

Google APIs have rate limits. If you hit them:

- **Gmail**: 250 quota units/user/second, 1 billion/day
- **Calendar**: 1,000,000 queries/day
- **Contacts**: 600 requests/minute/user
- **Docs/Sheets/Slides**: 300 requests/minute/user

If rate limited, wait 60 seconds and retry. Inform user if persistent.

## Getting Help

- README: General usage and command reference
- This file (AGENTS.md): Detailed agent guidance
- Command help: `gagent-cli <command> --help`
- Issues: https://github.com/ulfschnabel/gagent-cli/issues

## Summary: Quick Agent Checklist

✅ Use task commands for common operations
✅ Use API commands for advanced features
✅ Always parse JSON output (`success` field)
✅ Handle auth errors gracefully
✅ Confirm destructive operations
✅ Use `gmail reply` for threaded conversations
✅ Specify `--limit` for list operations
✅ Use ISO 8601 for all datetime values
✅ Check resource ID formats (message ID vs resource name)
✅ Preserve user data and privacy

Happy automating! 🤖
