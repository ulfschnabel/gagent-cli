# Calendar Commands

## Viewing Events

```bash
# Quick views
gagent-cli calendar today [--calendar ID]
gagent-cli calendar week [--calendar ID]
gagent-cli calendar upcoming --days 7

# Specific event details
gagent-cli calendar event <event-id>

# Check availability (free/busy)
gagent-cli calendar free-busy --start "2026-02-03T09:00:00Z" --end "2026-02-03T17:00:00Z"
```

## Creating and Managing Events

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
gagent-cli calendar respond <event-id> --status accepted|declined|tentative
```

## DateTime Format

**Always use ISO 8601 format: `YYYY-MM-DDTHH:MM:SSZ`**

Examples:
- `2026-02-03T14:00:00Z` - 2pm UTC on Feb 3, 2026
- `2026-02-03T14:00:00-05:00` - 2pm EST on Feb 3, 2026
- `2026-02-03` - All-day event on Feb 3, 2026

**Important:** Ask user about timezone if unclear. Default to UTC unless specified.

## API Commands (Low-Level)

```bash
# List calendars
gagent-cli calendar api calendars

# List events with custom filtering
gagent-cli calendar api events --time-min "2026-02-01T00:00:00Z" --time-max "2026-02-28T23:59:59Z"
```

## Common Workflow: Schedule Meeting

```bash
# 1. Check availability
gagent-cli calendar free-busy \
  --start "2026-02-05T14:00:00Z" \
  --end "2026-02-05T16:00:00Z"

# 2. If free, schedule
gagent-cli calendar schedule \
  --title "Team Sync" \
  --start "2026-02-05T14:00:00Z" \
  --end "2026-02-05T15:00:00Z" \
  --attendees "team@example.com"
```
