package docs

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
)

// ListOptions contains options for listing documents.
type ListOptions struct {
	Query      string
	MaxResults int64
	PageToken  string
}

// List returns a list of documents from Drive.
func (s *Service) List(opts ListOptions) ([]DocumentSummary, string, error) {
	query := "mimeType='application/vnd.google-apps.document'"
	if opts.Query != "" {
		query += fmt.Sprintf(" and name contains '%s'", opts.Query)
	}

	call := s.drive.Files.List().
		Q(query).
		Fields("files(id, name, modifiedTime, createdTime, webViewLink), nextPageToken").
		OrderBy("modifiedTime desc")

	if opts.MaxResults > 0 {
		call = call.PageSize(opts.MaxResults)
	} else {
		call = call.PageSize(10)
	}

	if opts.PageToken != "" {
		call = call.PageToken(opts.PageToken)
	}

	resp, err := doRetry(func() (*drive.FileList, error) {
		return call.Do()
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to list documents: %w", err)
	}

	documents := make([]DocumentSummary, 0, len(resp.Files))
	for _, file := range resp.Files {
		documents = append(documents, DocumentSummary{
			ID:           file.Id,
			Title:        file.Name,
			LastModified: file.ModifiedTime,
			CreatedTime:  file.CreatedTime,
			WebViewLink:  file.WebViewLink,
		})
	}

	return documents, resp.NextPageToken, nil
}

// Get returns the full document structure.
func (s *Service) Get(documentID string) (*docs.Document, error) {
	doc, err := doRetry(func() (*docs.Document, error) {
		return s.docs.Documents.Get(documentID).Do()
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}
	return doc, nil
}

// Read returns the document content as plain text.
func (s *Service) Read(documentID string) (*DocumentContent, error) {
	doc, err := s.Get(documentID)
	if err != nil {
		return nil, err
	}

	content := extractTextFromDocument(doc)

	return &DocumentContent{
		ID:      doc.DocumentId,
		Title:   doc.Title,
		Content: content,
	}, nil
}

// Outline returns the document structure (headings and sections).
func (s *Service) Outline(documentID string) (*DocumentOutline, error) {
	doc, err := s.Get(documentID)
	if err != nil {
		return nil, err
	}

	sections := extractSections(doc)

	return &DocumentOutline{
		ID:       doc.DocumentId,
		Title:    doc.Title,
		Sections: sections,
	}, nil
}

// Export exports the document in the specified format.
func (s *Service) Export(documentID, format string) ([]byte, error) {
	var mimeType string
	switch format {
	case "txt", "text":
		mimeType = "text/plain"
	case "html":
		mimeType = "text/html"
	case "pdf":
		mimeType = "application/pdf"
	case "docx":
		mimeType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case "md", "markdown":
		// Google doesn't support markdown directly, export as text
		mimeType = "text/plain"
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	resp, err := doRetry(func() (*http.Response, error) {
		return s.drive.Files.Export(documentID, mimeType).Download()
	})
	if err != nil {
		return nil, fmt.Errorf("failed to export document: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read export: %w", err)
	}

	return data, nil
}

// extractTextFromDocument extracts all text content from a document.
func extractTextFromDocument(doc *docs.Document) string {
	var text strings.Builder

	if doc.Body == nil {
		return ""
	}

	for _, element := range doc.Body.Content {
		extractTextFromElement(element, &text)
	}

	return text.String()
}

// extractTextFromElement extracts text from a structural element.
func extractTextFromElement(element *docs.StructuralElement, text *strings.Builder) {
	if element.Paragraph != nil {
		for _, e := range element.Paragraph.Elements {
			if e.TextRun != nil {
				text.WriteString(e.TextRun.Content)
			}
		}
	}

	if element.Table != nil {
		for _, row := range element.Table.TableRows {
			for _, cell := range row.TableCells {
				for _, content := range cell.Content {
					extractTextFromElement(content, text)
				}
			}
		}
	}

	if element.SectionBreak != nil {
		text.WriteString("\n")
	}
}

// extractSections extracts heading sections from a document.
func extractSections(doc *docs.Document) []SectionInfo {
	var sections []SectionInfo

	if doc.Body == nil {
		return sections
	}

	for _, element := range doc.Body.Content {
		if element.Paragraph != nil && element.Paragraph.ParagraphStyle != nil {
			style := element.Paragraph.ParagraphStyle.NamedStyleType
			level := getHeadingLevel(style)

			if level > 0 {
				// Extract heading text
				var heading strings.Builder
				for _, e := range element.Paragraph.Elements {
					if e.TextRun != nil {
						heading.WriteString(e.TextRun.Content)
					}
				}

				sections = append(sections, SectionInfo{
					Heading:    strings.TrimSpace(heading.String()),
					Level:      level,
					StartIndex: element.StartIndex,
					EndIndex:   element.EndIndex,
				})
			}
		}
	}

	return sections
}

// getHeadingLevel returns the heading level from a named style type.
func getHeadingLevel(style string) int {
	switch style {
	case "HEADING_1":
		return 1
	case "HEADING_2":
		return 2
	case "HEADING_3":
		return 3
	case "HEADING_4":
		return 4
	case "HEADING_5":
		return 5
	case "HEADING_6":
		return 6
	default:
		return 0
	}
}

// Structure returns a detailed structural analysis of the document.
func (s *Service) Structure(documentID string) (*DocumentStructure, error) {
	doc, err := s.Get(documentID)
	if err != nil {
		return nil, err
	}
	return analyzeStructure(doc), nil
}

// analyzeStructure walks the document body and extracts structural information.
func analyzeStructure(doc *docs.Document) *DocumentStructure {
	result := &DocumentStructure{
		ID:    doc.DocumentId,
		Title: doc.Title,
	}

	if doc.Body == nil {
		return result
	}

	// Track list item counts
	listCounts := map[string]int{}

	var allText strings.Builder

	for _, element := range doc.Body.Content {
		if element.Paragraph != nil {
			// Extract text for word count
			for _, e := range element.Paragraph.Elements {
				if e.TextRun != nil {
					allText.WriteString(e.TextRun.Content)
				}
			}

			// Check for heading
			if element.Paragraph.ParagraphStyle != nil {
				level := getHeadingLevel(element.Paragraph.ParagraphStyle.NamedStyleType)
				if level > 0 {
					var headingText strings.Builder
					for _, e := range element.Paragraph.Elements {
						if e.TextRun != nil {
							headingText.WriteString(e.TextRun.Content)
						}
					}
					result.Headings = append(result.Headings, HeadingInfo{
						Text:       strings.TrimSpace(headingText.String()),
						Level:      level,
						StartIndex: element.StartIndex,
					})
				}
			}

			// Check for list
			if element.Paragraph.Bullet != nil {
				listID := element.Paragraph.Bullet.ListId
				listCounts[listID]++
			}
		}

		if element.Table != nil {
			ti := TableInfo{
				Rows:       int(element.Table.Rows),
				Columns:    int(element.Table.Columns),
				StartIndex: element.StartIndex,
			}

			// Extract first row text
			if len(element.Table.TableRows) > 0 {
				row := element.Table.TableRows[0]
				for _, cell := range row.TableCells {
					var cellText strings.Builder
					for _, content := range cell.Content {
						if content.Paragraph != nil {
							for _, e := range content.Paragraph.Elements {
								if e.TextRun != nil {
									cellText.WriteString(e.TextRun.Content)
								}
							}
						}
					}
					ti.FirstRow = append(ti.FirstRow, strings.TrimSpace(cellText.String()))
				}
			}

			result.Tables = append(result.Tables, ti)
		}
	}

	// Build list info
	for listID, count := range listCounts {
		li := ListInfo{
			ListID:    listID,
			ItemCount: count,
		}
		// Get glyph type from document lists metadata
		if doc.Lists != nil {
			if listDef, ok := doc.Lists[listID]; ok {
				if listDef.ListProperties != nil && len(listDef.ListProperties.NestingLevels) > 0 {
					li.GlyphType = listDef.ListProperties.NestingLevels[0].GlyphType
				}
			}
		}
		result.Lists = append(result.Lists, li)
	}

	// Word count
	result.WordCount = countWords(allText.String())

	return result
}

// countWords counts words in text, splitting on whitespace.
func countWords(text string) int {
	words := strings.Fields(text)
	return len(words)
}
