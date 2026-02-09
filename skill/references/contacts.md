# Contacts Commands

## Finding Contacts

```bash
# List all contacts (paginated)
gagent-cli contacts list --limit 50 [--query "filter"]

# Search by name, email, or phone
gagent-cli contacts search "john smith"
gagent-cli contacts search "john@example.com"
gagent-cli contacts search "+1234567890"

# Get specific contact
gagent-cli contacts get people/c123456789

# List contact groups
gagent-cli contacts groups
```

## Managing Contacts

```bash
# Create simple contact
gagent-cli contacts create \
  --name "John Smith" \
  --email "john@example.com" \
  --phone "+1-555-0100"

# Update contact (requires etag)
gagent-cli contacts update people/c123456 \
  --json '{"displayName":"New Name","emailAddresses":[{"value":"new@example.com"}]}'

# Delete contact
gagent-cli contacts delete people/c123456
```

## Important Notes

1. **Resource Names**: Contacts use format `people/c123456789`
2. **ETags Required**: Updates require the current etag (get it with `contacts get`)
3. **Birthday Events**: Deleting a contact also removes associated calendar birthday events
4. **Search is Fuzzy**: Search matches partial names, emails, and phone numbers

## API Commands (Low-Level)

For complex updates with full control:

```bash
# Get contact with etag
gagent-cli contacts api get people/c123456

# Update with full Person resource
gagent-cli contacts api update people/c123456 \
  --person-json '{"names":[{"givenName":"John","familyName":"Doe"}],"emailAddresses":[{"value":"john.doe@example.com"}]}' \
  --etag "etag-value-from-get"

# Create with full control
gagent-cli contacts api create \
  --person-json '{"names":[{"givenName":"Jane","familyName":"Smith"}]}'

# Delete
gagent-cli contacts api delete people/c123456
```
