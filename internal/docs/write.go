package docs

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"google.golang.org/api/docs/v1"
)

// Create creates a new document.
func (s *Service) Create(title string) (*CreateResult, error) {
	doc := &docs.Document{
		Title: title,
	}

	created, err := s.docs.Documents.Create(doc).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	return &CreateResult{
		DocumentID: created.DocumentId,
		Title:      created.Title,
	}, nil
}

// CreateWithContent creates a new document with initial content.
func (s *Service) CreateWithContent(title, content string) (*CreateResult, error) {
	// First create the document
	result, err := s.Create(title)
	if err != nil {
		return nil, err
	}

	// Then insert content
	if content != "" {
		_, err = s.Append(result.DocumentID, content)
		if err != nil {
			return nil, fmt.Errorf("document created but failed to add content: %w", err)
		}
	}

	return result, nil
}

// Append appends text to the end of a document.
func (s *Service) Append(documentID, text string) (*UpdateResult, error) {
	// Get the document to find the end index
	doc, err := s.Get(documentID)
	if err != nil {
		return nil, err
	}

	// Find the end index (body content ends at Body.Content[last].EndIndex - 1)
	endIndex := int64(1)
	if doc.Body != nil && len(doc.Body.Content) > 0 {
		lastElement := doc.Body.Content[len(doc.Body.Content)-1]
		endIndex = lastElement.EndIndex - 1
	}

	requests := []*docs.Request{
		{
			InsertText: &docs.InsertTextRequest{
				Text: text,
				Location: &docs.Location{
					Index: endIndex,
				},
			},
		},
	}

	_, err = s.docs.Documents.BatchUpdate(documentID, &docs.BatchUpdateDocumentRequest{
		Requests: requests,
	}).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to append text: %w", err)
	}

	return &UpdateResult{
		DocumentID: documentID,
	}, nil
}

// Prepend inserts text at the beginning of a document.
func (s *Service) Prepend(documentID, text string) (*UpdateResult, error) {
	requests := []*docs.Request{
		{
			InsertText: &docs.InsertTextRequest{
				Text: text,
				Location: &docs.Location{
					Index: 1, // Index 1 is the start of body content
				},
			},
		},
	}

	_, err := s.docs.Documents.BatchUpdate(documentID, &docs.BatchUpdateDocumentRequest{
		Requests: requests,
	}).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to prepend text: %w", err)
	}

	return &UpdateResult{
		DocumentID: documentID,
	}, nil
}

// ReplaceText replaces all occurrences of text in a document.
func (s *Service) ReplaceText(documentID, find, replace string, matchCase bool) (*UpdateResult, error) {
	requests := []*docs.Request{
		{
			ReplaceAllText: &docs.ReplaceAllTextRequest{
				ContainsText: &docs.SubstringMatchCriteria{
					Text:      find,
					MatchCase: matchCase,
				},
				ReplaceText: replace,
			},
		},
	}

	_, err := s.docs.Documents.BatchUpdate(documentID, &docs.BatchUpdateDocumentRequest{
		Requests: requests,
	}).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to replace text: %w", err)
	}

	return &UpdateResult{
		DocumentID: documentID,
	}, nil
}

// UpdateSection updates the content of a section identified by its heading.
func (s *Service) UpdateSection(documentID, heading, content string) (*UpdateResult, error) {
	// Get the document outline to find the section
	outline, err := s.Outline(documentID)
	if err != nil {
		return nil, err
	}

	// Find the section with the matching heading
	var targetSection *SectionInfo
	var nextSectionStart int64 = -1

	for i, section := range outline.Sections {
		if strings.TrimSpace(section.Heading) == strings.TrimSpace(heading) {
			targetSection = &outline.Sections[i]
			// Find where the next section starts (or end of document)
			if i+1 < len(outline.Sections) {
				nextSectionStart = outline.Sections[i+1].StartIndex
			}
			break
		}
	}

	if targetSection == nil {
		return nil, fmt.Errorf("section not found: %s", heading)
	}

	// Get the document to find the actual content range
	doc, err := s.Get(documentID)
	if err != nil {
		return nil, err
	}

	// Find the end index for deletion
	// We want to delete from after the heading to the next section or end of document
	headingEndIndex := targetSection.EndIndex

	var deleteEndIndex int64
	if nextSectionStart > 0 {
		deleteEndIndex = nextSectionStart
	} else {
		// Use end of document
		if doc.Body != nil && len(doc.Body.Content) > 0 {
			deleteEndIndex = doc.Body.Content[len(doc.Body.Content)-1].EndIndex - 1
		} else {
			deleteEndIndex = headingEndIndex
		}
	}

	// Build requests: delete existing content, then insert new content
	requests := []*docs.Request{}

	// Only delete if there's content to delete
	if deleteEndIndex > headingEndIndex {
		requests = append(requests, &docs.Request{
			DeleteContentRange: &docs.DeleteContentRangeRequest{
				Range: &docs.Range{
					StartIndex: headingEndIndex,
					EndIndex:   deleteEndIndex,
				},
			},
		})
	}

	// Insert new content after the heading
	requests = append(requests, &docs.Request{
		InsertText: &docs.InsertTextRequest{
			Text: "\n" + content + "\n",
			Location: &docs.Location{
				Index: headingEndIndex,
			},
		},
	})

	_, err = s.docs.Documents.BatchUpdate(documentID, &docs.BatchUpdateDocumentRequest{
		Requests: requests,
	}).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to update section: %w", err)
	}

	return &UpdateResult{
		DocumentID: documentID,
	}, nil
}

// BatchUpdate performs a batch update with raw requests.
func (s *Service) BatchUpdate(documentID, requestsJSON string) (*UpdateResult, error) {
	var requests []*docs.Request
	if err := json.Unmarshal([]byte(requestsJSON), &requests); err != nil {
		return nil, fmt.Errorf("failed to parse requests JSON: %w", err)
	}

	_, err := s.docs.Documents.BatchUpdate(documentID, &docs.BatchUpdateDocumentRequest{
		Requests: requests,
	}).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to batch update: %w", err)
	}

	return &UpdateResult{
		DocumentID: documentID,
	}, nil
}

// InsertListOptions contains options for creating a list.
type InsertListOptions struct {
	Type   string   // List type: bullet, numbered, lettered, roman, checklist
	Items  []string // List items
	Indent int      // Indent level (0-9)
}

// InsertList inserts a list at the end of a document.
func (s *Service) InsertList(documentID string, opts InsertListOptions) (*UpdateResult, error) {
	// Validate inputs
	if err := validateListType(opts.Type); err != nil {
		return nil, err
	}
	if err := validateListItems(opts.Items); err != nil {
		return nil, err
	}
	if err := validateIndentLevel(opts.Indent); err != nil {
		return nil, err
	}

	// Get the document to find the end index
	doc, err := s.Get(documentID)
	if err != nil {
		return nil, err
	}

	endIndex := int64(1)
	if doc.Body != nil && len(doc.Body.Content) > 0 {
		lastElement := doc.Body.Content[len(doc.Body.Content)-1]
		endIndex = lastElement.EndIndex - 1
	}

	// Build the list text
	var listText strings.Builder
	for _, item := range opts.Items {
		listText.WriteString(item)
		listText.WriteString("\n")
	}

	glyphType := mapListGlyphType(opts.Type)

	requests := []*docs.Request{
		// Insert text
		{
			InsertText: &docs.InsertTextRequest{
				Text: listText.String(),
				Location: &docs.Location{
					Index: endIndex,
				},
			},
		},
		// Create list
		{
			CreateParagraphBullets: &docs.CreateParagraphBulletsRequest{
				Range: &docs.Range{
					StartIndex: endIndex,
					EndIndex:   endIndex + int64(len(listText.String())),
				},
				BulletPreset: glyphType,
			},
		},
	}

	// Add indent if specified
	if opts.Indent > 0 {
		requests = append(requests, &docs.Request{
			UpdateParagraphStyle: &docs.UpdateParagraphStyleRequest{
				Range: &docs.Range{
					StartIndex: endIndex,
					EndIndex:   endIndex + int64(len(listText.String())),
				},
				ParagraphStyle: &docs.ParagraphStyle{
					IndentStart: &docs.Dimension{
						Magnitude: float64(opts.Indent * 36), // 36 points per indent level
						Unit:      "PT",
					},
				},
				Fields: "indentStart",
			},
		})
	}

	_, err = s.docs.Documents.BatchUpdate(documentID, &docs.BatchUpdateDocumentRequest{
		Requests: requests,
	}).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to insert list: %w", err)
	}

	return &UpdateResult{
		DocumentID: documentID,
	}, nil
}

// TextStyleOptions contains options for text styling.
type TextStyleOptions struct {
	Bold          bool
	Italic        bool
	Underline     bool
	Strikethrough bool
	FontSize      int    // Font size in points (8-72)
	FontFamily    string // Font family name
	Color         string // Hex color (#rrggbb)
	BgColor       string // Background hex color (#rrggbb)
	NamedStyle    string // Named style: heading1, heading2, etc.
}

// AppendFormatted appends formatted text to the end of a document.
func (s *Service) AppendFormatted(documentID, text string, opts TextStyleOptions) (*UpdateResult, error) {
	// Validate inputs
	if text == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}
	if err := validateColor(opts.Color); err != nil {
		return nil, err
	}
	if err := validateColor(opts.BgColor); err != nil {
		return nil, err
	}
	if err := validateFontSize(opts.FontSize); err != nil {
		return nil, err
	}

	// Get the document to find the end index
	doc, err := s.Get(documentID)
	if err != nil {
		return nil, err
	}

	endIndex := int64(1)
	if doc.Body != nil && len(doc.Body.Content) > 0 {
		lastElement := doc.Body.Content[len(doc.Body.Content)-1]
		endIndex = lastElement.EndIndex - 1
	}

	requests := []*docs.Request{
		{
			InsertText: &docs.InsertTextRequest{
				Text: text,
				Location: &docs.Location{
					Index: endIndex,
				},
			},
		},
	}

	// Apply named style if specified
	if opts.NamedStyle != "" {
		namedStyleType := mapNamedStyle(opts.NamedStyle)
		requests = append(requests, &docs.Request{
			UpdateParagraphStyle: &docs.UpdateParagraphStyleRequest{
				Range: &docs.Range{
					StartIndex: endIndex,
					EndIndex:   endIndex + int64(len(text)),
				},
				ParagraphStyle: &docs.ParagraphStyle{
					NamedStyleType: namedStyleType,
				},
				Fields: "namedStyleType",
			},
		})
	}

	// Apply text style if any formatting is specified
	textStyle := parseTextStyleFlags(opts.Bold, opts.Italic, opts.Underline, opts.Strikethrough, opts.FontSize, opts.Color)
	if opts.FontFamily != "" {
		textStyle.WeightedFontFamily = &docs.WeightedFontFamily{
			FontFamily: opts.FontFamily,
		}
	}
	if opts.BgColor != "" {
		textStyle.BackgroundColor = parseColor(opts.BgColor)
	}

	// Build fields mask for text style
	fields := buildTextStyleFields(opts)
	if fields != "" {
		requests = append(requests, &docs.Request{
			UpdateTextStyle: &docs.UpdateTextStyleRequest{
				Range: &docs.Range{
					StartIndex: endIndex,
					EndIndex:   endIndex + int64(len(text)),
				},
				TextStyle: textStyle,
				Fields:    fields,
			},
		})
	}

	_, err = s.docs.Documents.BatchUpdate(documentID, &docs.BatchUpdateDocumentRequest{
		Requests: requests,
	}).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to append formatted text: %w", err)
	}

	return &UpdateResult{
		DocumentID: documentID,
	}, nil
}

// ParagraphStyleOptions contains options for paragraph styling.
type ParagraphStyleOptions struct {
	Alignment     string  // left, center, right, justify
	IndentStart   float64 // Indent in points
	IndentEnd     float64 // Right indent in points
	IndentFirst   float64 // First line indent in points
	LineSpacing   float64 // Line spacing (e.g., 1.5)
	SpacingBefore float64 // Space before paragraph in points
	SpacingAfter  float64 // Space after paragraph in points
}

// FormatParagraph applies paragraph formatting to a range of text.
func (s *Service) FormatParagraph(documentID string, startIndex, endIndex int64, opts ParagraphStyleOptions) (*UpdateResult, error) {
	if err := validateAlignment(opts.Alignment); err != nil {
		return nil, err
	}

	if startIndex < 0 || endIndex < startIndex {
		return nil, fmt.Errorf("invalid range: startIndex=%d, endIndex=%d", startIndex, endIndex)
	}

	paragraphStyle := &docs.ParagraphStyle{}
	fields := []string{}

	if opts.Alignment != "" {
		paragraphStyle.Alignment = strings.ToUpper(opts.Alignment)
		fields = append(fields, "alignment")
	}
	if opts.IndentStart > 0 {
		paragraphStyle.IndentStart = &docs.Dimension{
			Magnitude: opts.IndentStart,
			Unit:      "PT",
		}
		fields = append(fields, "indentStart")
	}
	if opts.IndentEnd > 0 {
		paragraphStyle.IndentEnd = &docs.Dimension{
			Magnitude: opts.IndentEnd,
			Unit:      "PT",
		}
		fields = append(fields, "indentEnd")
	}
	if opts.IndentFirst != 0 {
		paragraphStyle.IndentFirstLine = &docs.Dimension{
			Magnitude: opts.IndentFirst,
			Unit:      "PT",
		}
		fields = append(fields, "indentFirstLine")
	}
	if opts.LineSpacing > 0 {
		paragraphStyle.LineSpacing = opts.LineSpacing * 100 // API expects percentage
		fields = append(fields, "lineSpacing")
	}
	if opts.SpacingBefore > 0 {
		paragraphStyle.SpaceAbove = &docs.Dimension{
			Magnitude: opts.SpacingBefore,
			Unit:      "PT",
		}
		fields = append(fields, "spaceAbove")
	}
	if opts.SpacingAfter > 0 {
		paragraphStyle.SpaceBelow = &docs.Dimension{
			Magnitude: opts.SpacingAfter,
			Unit:      "PT",
		}
		fields = append(fields, "spaceBelow")
	}

	if len(fields) == 0 {
		return nil, fmt.Errorf("no formatting options specified")
	}

	requests := []*docs.Request{
		{
			UpdateParagraphStyle: &docs.UpdateParagraphStyleRequest{
				Range: &docs.Range{
					StartIndex: startIndex,
					EndIndex:   endIndex,
				},
				ParagraphStyle: paragraphStyle,
				Fields:         strings.Join(fields, ","),
			},
		},
	}

	_, err := s.docs.Documents.BatchUpdate(documentID, &docs.BatchUpdateDocumentRequest{
		Requests: requests,
	}).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to format paragraph: %w", err)
	}

	return &UpdateResult{
		DocumentID: documentID,
	}, nil
}

// TableOptions contains options for creating a table.
type TableOptions struct {
	Rows    int      // Number of rows
	Columns int      // Number of columns
	Headers []string // Optional header row
	Data    [][]string
}

// InsertTable inserts a table at the end of a document.
func (s *Service) InsertTable(documentID string, opts TableOptions) (*UpdateResult, error) {
	if opts.Rows < 1 || opts.Columns < 1 {
		return nil, fmt.Errorf("rows and columns must be at least 1")
	}

	// Get the document to find the end index
	doc, err := s.Get(documentID)
	if err != nil {
		return nil, err
	}

	endIndex := int64(1)
	if doc.Body != nil && len(doc.Body.Content) > 0 {
		lastElement := doc.Body.Content[len(doc.Body.Content)-1]
		endIndex = lastElement.EndIndex - 1
	}

	requests := []*docs.Request{
		{
			InsertTable: &docs.InsertTableRequest{
				Rows:    int64(opts.Rows),
				Columns: int64(opts.Columns),
				Location: &docs.Location{
					Index: endIndex,
				},
			},
		},
	}

	_, err = s.docs.Documents.BatchUpdate(documentID, &docs.BatchUpdateDocumentRequest{
		Requests: requests,
	}).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to insert table: %w", err)
	}

	// If headers or data provided, insert them
	if len(opts.Headers) > 0 || len(opts.Data) > 0 {
		// Get the updated document to find table cells
		doc, err = s.Get(documentID)
		if err != nil {
			return nil, err
		}

		// Populate table with data
		err = s.populateTableData(documentID, doc, opts)
		if err != nil {
			return nil, fmt.Errorf("table created but failed to populate: %w", err)
		}
	}

	return &UpdateResult{
		DocumentID: documentID,
	}, nil
}

// InsertTableFromCSV inserts a table from CSV data.
func (s *Service) InsertTableFromCSV(documentID string, csvData string, hasHeaders bool) (*UpdateResult, error) {
	reader := csv.NewReader(strings.NewReader(csvData))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("CSV data is empty")
	}

	opts := TableOptions{
		Columns: len(records[0]),
	}

	if hasHeaders {
		opts.Headers = records[0]
		opts.Data = records[1:]
		opts.Rows = len(records) // Including header
	} else {
		opts.Data = records
		opts.Rows = len(records)
	}

	return s.InsertTable(documentID, opts)
}

// InsertPageBreak inserts a page break at the end of a document.
func (s *Service) InsertPageBreak(documentID string) (*UpdateResult, error) {
	doc, err := s.Get(documentID)
	if err != nil {
		return nil, err
	}

	endIndex := int64(1)
	if doc.Body != nil && len(doc.Body.Content) > 0 {
		lastElement := doc.Body.Content[len(doc.Body.Content)-1]
		endIndex = lastElement.EndIndex - 1
	}

	requests := []*docs.Request{
		{
			InsertPageBreak: &docs.InsertPageBreakRequest{
				Location: &docs.Location{
					Index: endIndex,
				},
			},
		},
	}

	_, err = s.docs.Documents.BatchUpdate(documentID, &docs.BatchUpdateDocumentRequest{
		Requests: requests,
	}).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to insert page break: %w", err)
	}

	return &UpdateResult{
		DocumentID: documentID,
	}, nil
}

// InsertHorizontalRule inserts a horizontal rule at the end of a document.
func (s *Service) InsertHorizontalRule(documentID string) (*UpdateResult, error) {
	// Google Docs doesn't have a direct HR API, so we insert a text-based rule
	_, err := s.Append(documentID, "\n"+strings.Repeat("_", 50)+"\n")
	if err != nil {
		return nil, fmt.Errorf("failed to insert horizontal rule: %w", err)
	}

	return &UpdateResult{
		DocumentID: documentID,
	}, nil
}

// InsertTOC inserts a table of contents at the beginning of a document.
func (s *Service) InsertTOC(documentID string) (*UpdateResult, error) {
	// Insert a placeholder for TOC
	// The Google Docs API doesn't have direct TOC insertion in the public API
	// Users need to manually insert it through the UI or use Apps Script
	_, err := s.Prepend(documentID, "[Table of Contents]\n\n")
	if err != nil {
		return nil, fmt.Errorf("failed to insert table of contents placeholder: %w", err)
	}

	return &UpdateResult{
		DocumentID: documentID,
	}, nil
}

// populateTableData fills a table with data.
func (s *Service) populateTableData(documentID string, doc *docs.Document, opts TableOptions) error {
	// Find the table in the document
	var table *docs.Table
	for _, element := range doc.Body.Content {
		if element.Table != nil {
			table = element.Table
			break
		}
	}

	if table == nil {
		return fmt.Errorf("table not found in document")
	}

	requests := []*docs.Request{}

	// Insert headers
	if len(opts.Headers) > 0 && len(table.TableRows) > 0 {
		row := table.TableRows[0]
		for colIdx, header := range opts.Headers {
			if colIdx >= len(row.TableCells) {
				break
			}
			cell := row.TableCells[colIdx]
			if cell.Content != nil && len(cell.Content) > 0 {
				startIndex := cell.Content[0].StartIndex
				requests = append(requests, &docs.Request{
					InsertText: &docs.InsertTextRequest{
						Text: header,
						Location: &docs.Location{
							Index: startIndex,
						},
					},
				})
			}
		}
	}

	// Insert data rows
	dataStartRow := 0
	if len(opts.Headers) > 0 {
		dataStartRow = 1
	}

	for rowIdx, rowData := range opts.Data {
		tableRowIdx := dataStartRow + rowIdx
		if tableRowIdx >= len(table.TableRows) {
			break
		}
		row := table.TableRows[tableRowIdx]
		for colIdx, cellData := range rowData {
			if colIdx >= len(row.TableCells) {
				break
			}
			cell := row.TableCells[colIdx]
			if cell.Content != nil && len(cell.Content) > 0 {
				startIndex := cell.Content[0].StartIndex
				requests = append(requests, &docs.Request{
					InsertText: &docs.InsertTextRequest{
						Text: cellData,
						Location: &docs.Location{
							Index: startIndex,
						},
					},
				})
			}
		}
	}

	if len(requests) > 0 {
		_, err := s.docs.Documents.BatchUpdate(documentID, &docs.BatchUpdateDocumentRequest{
			Requests: requests,
		}).Do()
		if err != nil {
			return err
		}
	}

	return nil
}

// Validation functions

func validateListType(listType string) error {
	validTypes := map[string]bool{
		"bullet":    true,
		"numbered":  true,
		"lettered":  true,
		"roman":     true,
		"checklist": true,
	}

	if listType == "" {
		return fmt.Errorf("list type cannot be empty")
	}

	if !validTypes[listType] {
		return fmt.Errorf("invalid list type: %s (valid: bullet, numbered, lettered, roman, checklist)", listType)
	}

	return nil
}

func validateListItems(items []string) error {
	if items == nil || len(items) == 0 {
		return fmt.Errorf("list items cannot be empty")
	}

	for i, item := range items {
		if strings.TrimSpace(item) == "" {
			return fmt.Errorf("list item %d is empty or whitespace only", i)
		}
	}

	return nil
}

func validateIndentLevel(indent int) error {
	if indent < 0 {
		return fmt.Errorf("indent level cannot be negative")
	}
	if indent > 9 {
		return fmt.Errorf("indent level cannot exceed 9")
	}
	return nil
}

func mapListGlyphType(listType string) string {
	glyphTypes := map[string]string{
		"bullet":    "BULLET_DISC_CIRCLE_SQUARE",
		"numbered":  "DECIMAL",
		"lettered":  "ALPHA",
		"roman":     "ROMAN",
		"checklist": "CHECKBOX",
	}
	return glyphTypes[listType]
}

func validateColor(color string) error {
	if color == "" {
		return nil // Empty is valid (no color)
	}

	hexPattern := regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
	if !hexPattern.MatchString(color) {
		return fmt.Errorf("invalid color format: %s (expected #rrggbb)", color)
	}

	return nil
}

func validateFontSize(fontSize int) error {
	if fontSize == 0 {
		return nil // 0 means no size specified
	}
	if fontSize < 8 || fontSize > 72 {
		return fmt.Errorf("font size must be between 8 and 72 points")
	}
	return nil
}

func mapNamedStyle(styleName string) string {
	styles := map[string]string{
		"heading1": "HEADING_1",
		"heading2": "HEADING_2",
		"heading3": "HEADING_3",
		"heading4": "HEADING_4",
		"title":    "TITLE",
		"subtitle": "SUBTITLE",
		"normal":   "NORMAL_TEXT",
	}
	return styles[styleName]
}

func validateAlignment(alignment string) error {
	if alignment == "" {
		return nil // Empty is valid (no alignment)
	}

	validAlignments := map[string]bool{
		"left":    true,
		"center":  true,
		"right":   true,
		"justify": true,
	}

	if !validAlignments[alignment] {
		return fmt.Errorf("invalid alignment: %s (valid: left, center, right, justify)", alignment)
	}

	return nil
}

func parseTextStyleFlags(bold, italic, underline, strikethrough bool, fontSize int, color string) *docs.TextStyle {
	style := &docs.TextStyle{}

	style.Bold = bold
	style.Italic = italic
	style.Underline = underline
	style.Strikethrough = strikethrough

	if fontSize > 0 {
		style.FontSize = &docs.Dimension{
			Magnitude: float64(fontSize),
			Unit:      "PT",
		}
	}
	if color != "" {
		style.ForegroundColor = parseColor(color)
	}

	return style
}

func parseColor(hexColor string) *docs.OptionalColor {
	if hexColor == "" {
		return nil
	}

	// Remove # prefix
	hexColor = strings.TrimPrefix(hexColor, "#")

	// Parse RGB values
	var r, g, b int
	fmt.Sscanf(hexColor, "%02x%02x%02x", &r, &g, &b)

	return &docs.OptionalColor{
		Color: &docs.Color{
			RgbColor: &docs.RgbColor{
				Red:   float64(r) / 255.0,
				Green: float64(g) / 255.0,
				Blue:  float64(b) / 255.0,
			},
		},
	}
}

func buildTextStyleFields(opts TextStyleOptions) string {
	fields := []string{}

	if opts.Bold {
		fields = append(fields, "bold")
	}
	if opts.Italic {
		fields = append(fields, "italic")
	}
	if opts.Underline {
		fields = append(fields, "underline")
	}
	if opts.Strikethrough {
		fields = append(fields, "strikethrough")
	}
	if opts.FontSize > 0 {
		fields = append(fields, "fontSize")
	}
	if opts.FontFamily != "" {
		fields = append(fields, "weightedFontFamily")
	}
	if opts.Color != "" {
		fields = append(fields, "foregroundColor")
	}
	if opts.BgColor != "" {
		fields = append(fields, "backgroundColor")
	}

	return strings.Join(fields, ",")
}

// DocumentTemplate represents a template for batch formatting.
type DocumentTemplate struct {
	Title    *TemplateText       `json:"title,omitempty"`
	Sections []TemplateSection   `json:"sections,omitempty"`
	Elements []TemplateElement   `json:"elements,omitempty"`
}

// TemplateText represents formatted text.
type TemplateText struct {
	Text  string           `json:"text"`
	Style TextStyleOptions `json:"style,omitempty"`
}

// TemplateSection represents a document section.
type TemplateSection struct {
	Heading  string            `json:"heading"`
	Style    string            `json:"style,omitempty"` // heading1, heading2, etc.
	Content  []TemplateElement `json:"content,omitempty"`
}

// TemplateElement represents a document element.
type TemplateElement struct {
	Type      string            `json:"type"` // text, list, table, pagebreak, hr
	Text      string            `json:"text,omitempty"`
	Style     TextStyleOptions  `json:"style,omitempty"`
	ListType  string            `json:"list_type,omitempty"`
	Items     []string          `json:"items,omitempty"`
	Indent    int               `json:"indent,omitempty"`
	Rows      int               `json:"rows,omitempty"`
	Columns   int               `json:"columns,omitempty"`
	Headers   []string          `json:"headers,omitempty"`
	TableData [][]string        `json:"table_data,omitempty"`
}

// FormatFromTemplate applies a template to a document.
func (s *Service) FormatFromTemplate(documentID string, templateJSON string) (*UpdateResult, error) {
	var template DocumentTemplate
	if err := json.Unmarshal([]byte(templateJSON), &template); err != nil {
		return nil, fmt.Errorf("failed to parse template JSON: %w", err)
	}

	// Add title if specified
	if template.Title != nil {
		_, err := s.AppendFormatted(documentID, template.Title.Text+"\n\n", template.Title.Style)
		if err != nil {
			return nil, fmt.Errorf("failed to add title: %w", err)
		}
	}

	// Add sections
	for _, section := range template.Sections {
		// Add heading
		headingStyle := TextStyleOptions{
			NamedStyle: section.Style,
		}
		if headingStyle.NamedStyle == "" {
			headingStyle.NamedStyle = "heading2"
		}
		_, err := s.AppendFormatted(documentID, section.Heading+"\n", headingStyle)
		if err != nil {
			return nil, fmt.Errorf("failed to add section heading: %w", err)
		}

		// Add section content
		for _, element := range section.Content {
			if err := s.addTemplateElement(documentID, element); err != nil {
				return nil, fmt.Errorf("failed to add section element: %w", err)
			}
		}
	}

	// Add standalone elements
	for _, element := range template.Elements {
		if err := s.addTemplateElement(documentID, element); err != nil {
			return nil, fmt.Errorf("failed to add element: %w", err)
		}
	}

	return &UpdateResult{
		DocumentID: documentID,
	}, nil
}

// addTemplateElement adds a template element to a document.
func (s *Service) addTemplateElement(documentID string, element TemplateElement) error {
	switch element.Type {
	case "text":
		_, err := s.AppendFormatted(documentID, element.Text+"\n", element.Style)
		return err

	case "list":
		_, err := s.InsertList(documentID, InsertListOptions{
			Type:   element.ListType,
			Items:  element.Items,
			Indent: element.Indent,
		})
		return err

	case "table":
		_, err := s.InsertTable(documentID, TableOptions{
			Rows:    element.Rows,
			Columns: element.Columns,
			Headers: element.Headers,
			Data:    element.TableData,
		})
		return err

	case "pagebreak":
		_, err := s.InsertPageBreak(documentID)
		return err

	case "hr":
		_, err := s.InsertHorizontalRule(documentID)
		return err

	default:
		return fmt.Errorf("unknown element type: %s", element.Type)
	}
}
