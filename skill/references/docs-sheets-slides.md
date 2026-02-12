# Google Docs, Sheets, and Slides Commands

## Google Docs

### Building Documents from Markdown (PREFERRED)

**Use `docs from-markdown` to build well-formatted Google Docs.** Write your content as
markdown and this command converts it to native Google Docs formatting (headings, bold,
italic, lists, code blocks, links) in a single atomic API call. Do NOT build documents
piece by piece with multiple append/insert calls.

```bash
# Create a new document from markdown (PREFERRED)
gagent-cli docs from-markdown --title "Project Plan" --text "# Project Plan

## Goals
- Launch by Q2
- **500** active users

## Timeline
1. Design phase: *2 weeks*
2. Implementation: 4 weeks
3. Testing: 1 week"

# Append markdown to an existing document
gagent-cli docs from-markdown <doc-id> --text "## New Section

More **formatted** content here."

# Replace entire document content (for iterating on drafts)
gagent-cli docs from-markdown <doc-id> --replace --text "# Revised version

Completely rewritten with **new content**."

# From a markdown file
gagent-cli docs from-markdown <doc-id> --file report.md

# Preview without applying
gagent-cli docs from-markdown <doc-id> --text "# Test" --preview
```

**Flags:** `--title` creates new doc, `--replace` rewrites existing doc, neither appends.

**Supported markdown:** Headings (h1-h6), **bold**, *italic*, bullet lists,
numbered lists, `inline code`, fenced code blocks (monospace), [links](url),
and horizontal rules (---).

### Reading

```bash
# Read document content
gagent-cli docs read <doc-id>

# Get document outline/structure
gagent-cli docs outline <doc-id>

# Analyze document structure (headings, tables, lists, word count)
gagent-cli docs structure <doc-id>

# Export to different formats
gagent-cli docs export <doc-id> --format txt|html|pdf

# List documents
gagent-cli docs list --limit 10 [--query "search term"]
```

### Other Write Operations

```bash
# Append plain text (use from-markdown instead for formatted content)
gagent-cli docs append <doc-id> --text "plain text"

# Find and replace
gagent-cli docs replace-text <doc-id> --find "old text" --replace "new text"

# Update specific section by heading
gagent-cli docs update-section <doc-id> --heading "Budget" --content "Updated info"
```

### API Commands

```bash
# Get full document structure
gagent-cli docs api get <doc-id>

# Batch update with JSON requests
gagent-cli docs api batch-update <doc-id> --requests-json '[...]'
```

## Google Sheets

### Reading

```bash
# Read spreadsheet data
gagent-cli sheets read <spreadsheet-id> [--sheet "Sheet1"] [--range "A1:Z100"]

# Get spreadsheet info (metadata, sheets)
gagent-cli sheets info <spreadsheet-id>

# Export spreadsheet
gagent-cli sheets export <spreadsheet-id> --format csv|xlsx|pdf

# List spreadsheets
gagent-cli sheets list --limit 10
```

### Writing

```bash
# Create spreadsheet
gagent-cli sheets create --title "Q1 Report"

# Write data (2D array in JSON)
gagent-cli sheets write <spreadsheet-id> \
  --sheet "Sheet1" \
  --range "A1" \
  --values '[["Name","Email"],["John","john@example.com"],["Jane","jane@example.com"]]'

# Append rows
gagent-cli sheets append <spreadsheet-id> \
  --sheet "Sheet1" \
  --values '[["New","Data","Here"]]'

# Clear range
gagent-cli sheets clear <spreadsheet-id> --sheet "Sheet1" --range "A1:Z100"

# Add new sheet
gagent-cli sheets add-sheet <spreadsheet-id> --name "New Sheet"
```

### API Commands

```bash
# Get full spreadsheet metadata
gagent-cli sheets api get <spreadsheet-id>

# Get values from range
gagent-cli sheets api values <spreadsheet-id> --range "Sheet1!A1:Z100"

# Batch update with JSON requests
gagent-cli sheets api batch-update <spreadsheet-id> --requests-json '[...]'
```

## Google Slides

### Reading

```bash
# Get presentation info
gagent-cli slides info <presentation-id>

# Read specific slide
gagent-cli slides read <presentation-id> [--slide 1]

# Extract all text from presentation
gagent-cli slides text <presentation-id>

# Export presentation
gagent-cli slides export <presentation-id> --format pdf|pptx [--output /path/to/file]

# List presentations
gagent-cli slides list --limit 10
```

### Writing

```bash
# Create presentation
gagent-cli slides create --title "Q1 Results"

# Add slide
gagent-cli slides add-slide <presentation-id> [--layout BLANK|TITLE_AND_BODY|...]

# Delete slide
gagent-cli slides delete-slide <presentation-id> --slide 1

# Update text (find and replace)
gagent-cli slides update-text <presentation-id> \
  --slide 1 \
  --find "{{date}}" \
  --replace "2026-02-03"

# Add text box
gagent-cli slides add-text <presentation-id> \
  --slide 1 \
  --text "Title" \
  --x 100 --y 100 \
  --width 600 --height 80
```

### Visual Feedback Loop (CRITICAL for Slides)

**Problem**: When building slides programmatically, you cannot see the visual output. You're doing graphic design blind—guessing at positions, sizes, and styling without knowing if elements overlap or text wraps awkwardly.

**Solution**: Use the export-to-PDF feedback loop:

```bash
# 1. Make changes to slides
gagent-cli slides add-text <pres-id> --slide 1 --text "Title" --x 100 --y 100

# 2. Export to PDF
gagent-cli slides export <pres-id> --format pdf --output /tmp/slides.pdf

# 3. Read the PDF (vision-enabled agents can see the rendered slides)
# Use Read tool to view /tmp/slides.pdf

# 4. Analyze what needs fixing (overlaps, positioning, sizing)

# 5. Apply fixes using batch-update API
gagent-cli slides api batch-update <pres-id> --requests-json '[...]'

# 6. Repeat steps 2-5 until slides look good
```

### API Commands

```bash
# Get full presentation structure with element IDs
gagent-cli slides api get <presentation-id>

# Batch update with JSON requests (for styling, positioning)
gagent-cli slides api batch-update <presentation-id> --requests-json '[...]'
```

### Common Batch Update Operations

**Update text style:**
```json
{
  "updateTextStyle": {
    "objectId": "textbox_id",
    "style": {"fontSize": {"magnitude": 36, "unit": "PT"}, "bold": true},
    "textRange": {"type": "ALL"},
    "fields": "fontSize,bold"
  }
}
```

**Send shape to back:**
```json
{
  "updatePageElementsZOrder": {
    "pageElementObjectIds": ["shape_id"],
    "operation": "SEND_TO_BACK"
  }
}
```

**Apply background color:**
```json
{
  "updatePageProperties": {
    "objectId": "slide_id",
    "pageProperties": {
      "pageBackgroundFill": {
        "solidFill": {"color": {"rgbColor": {"red": 0.95, "green": 0.97, "blue": 1.0}}}
      }
    },
    "fields": "pageBackgroundFill.solidFill.color"
  }
}
```

### Tips for Slide Styling

1. **Get element IDs first**: Use `slides read <pres-id> --slide N` to find textbox IDs
2. **Use batch-update for styling**: The API commands support complex formatting
3. **Positioning units**: Google Slides API uses EMU (English Metric Units). 1 inch = 914400 EMU
4. **Always verify visually**: Export to PDF after changes to see actual rendering
