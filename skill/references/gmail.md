# Gmail Commands

## Reading Email

```bash
# List recent emails
gagent-cli gmail inbox --limit 10 [--unread-only]

# Search with Gmail query syntax
gagent-cli gmail search "from:boss@company.com is:unread" --limit 20
gagent-cli gmail search "has:attachment after:2026/01/01"
gagent-cli gmail search "subject:invoice before:2026/02/01"

# Read full message
gagent-cli gmail read <message-id>

# Get conversation thread
gagent-cli gmail thread <thread-id>
```

## Sending Email

**CRITICAL: Email Threading**
- Use `gmail reply` to keep emails in the same thread
- Use `gmail send` only for NEW conversations
- `gmail reply` automatically sets proper headers (In-Reply-To, References)

```bash
# Reply to maintain thread (RECOMMENDED)
gagent-cli gmail reply <message-id> --body "Your reply" [--reply-all]

# Send new email (creates new thread)
gagent-cli gmail send --to "user@example.com" --subject "Topic" --body "Content"

# Create draft for user review
gagent-cli gmail draft --to "user@example.com" --subject "Draft" --body "Content"

# Forward with context
gagent-cli gmail forward <message-id> --to "colleague@example.com" --body "FYI"
```

## Search Operators

Common Gmail search syntax:
- `from:` - Sender email
- `to:` - Recipient email
- `subject:` - Subject line
- `after:`, `before:` - Date filters (YYYY/MM/DD format)
- `has:attachment` - Has attachments
- `is:unread`, `is:read` - Read status
- `is:important`, `is:starred` - Labels
- `label:` - Specific label
- `-` prefix - Exclude (e.g., `-from:spam@example.com`)

## API Commands (Low-Level)

```bash
# List messages with advanced filtering
gagent-cli gmail api list --label INBOX --query "is:unread"

# Get raw message data
gagent-cli gmail api get <message-id>

# List all labels
gagent-cli gmail api labels

# Download attachment
gagent-cli gmail api attachment <message-id> <attachment-id> --save-to /path/to/file
```
