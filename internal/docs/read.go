package docs

import (
	"fmt"
	"io"
	"strings"

	"google.golang.org/api/docs/v1"
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

	resp, err := call.Do()
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
	doc, err := s.docs.Documents.Get(documentID).Do()
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

	resp, err := s.drive.Files.Export(documentID, mimeType).Download()
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
