package slides

import (
	"fmt"
	"io"
	"strings"

	"google.golang.org/api/slides/v1"
)

// ListOptions contains options for listing presentations.
type ListOptions struct {
	Query      string
	MaxResults int64
	PageToken  string
}

// List returns a list of presentations from Drive.
func (s *Service) List(opts ListOptions) ([]PresentationSummary, string, error) {
	query := "mimeType='application/vnd.google-apps.presentation'"
	if opts.Query != "" {
		query += fmt.Sprintf(" and name contains '%s'", opts.Query)
	}

	call := s.drive.Files.List().
		Q(query).
		Fields("files(id, name, modifiedTime, webViewLink), nextPageToken").
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
		return nil, "", fmt.Errorf("failed to list presentations: %w", err)
	}

	presentations := make([]PresentationSummary, 0, len(resp.Files))
	for _, file := range resp.Files {
		presentations = append(presentations, PresentationSummary{
			ID:           file.Id,
			Title:        file.Name,
			LastModified: file.ModifiedTime,
			WebViewLink:  file.WebViewLink,
		})
	}

	return presentations, resp.NextPageToken, nil
}

// Info returns metadata about a presentation.
func (s *Service) Info(presentationID string) (*PresentationInfo, error) {
	pres, err := s.slides.Presentations.Get(presentationID).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get presentation: %w", err)
	}

	info := &PresentationInfo{
		ID:         pres.PresentationId,
		Title:      pres.Title,
		SlideCount: len(pres.Slides),
	}

	if pres.PageSize != nil && pres.PageSize.Width != nil && pres.PageSize.Height != nil {
		info.Width = int64(pres.PageSize.Width.Magnitude)
		info.Height = int64(pres.PageSize.Height.Magnitude)
	}

	if pres.Locale != "" {
		info.Locale = pres.Locale
	}

	return info, nil
}

// Get returns the full presentation structure.
func (s *Service) Get(presentationID string) (*slides.Presentation, error) {
	pres, err := s.slides.Presentations.Get(presentationID).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get presentation: %w", err)
	}
	return pres, nil
}

// GetPage returns a specific page/slide.
func (s *Service) GetPage(presentationID, pageID string) (*slides.Page, error) {
	page, err := s.slides.Presentations.Pages.Get(presentationID, pageID).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get page: %w", err)
	}
	return page, nil
}

// Read returns the content of a specific slide.
func (s *Service) Read(presentationID string, slideIndex int) (*SlideContent, error) {
	pres, err := s.Get(presentationID)
	if err != nil {
		return nil, err
	}

	if slideIndex < 1 || slideIndex > len(pres.Slides) {
		return nil, fmt.Errorf("slide index out of range: %d (presentation has %d slides)", slideIndex, len(pres.Slides))
	}

	slide := pres.Slides[slideIndex-1]
	return parseSlideContent(slide, slideIndex), nil
}

// ReadAll returns the content of all slides.
func (s *Service) ReadAll(presentationID string) ([]SlideContent, error) {
	pres, err := s.Get(presentationID)
	if err != nil {
		return nil, err
	}

	contents := make([]SlideContent, 0, len(pres.Slides))
	for i, slide := range pres.Slides {
		contents = append(contents, *parseSlideContent(slide, i+1))
	}

	return contents, nil
}

// Text extracts all text from a presentation.
func (s *Service) Text(presentationID string) (string, error) {
	pres, err := s.Get(presentationID)
	if err != nil {
		return "", err
	}

	var text strings.Builder
	text.WriteString(fmt.Sprintf("Title: %s\n\n", pres.Title))

	for i, slide := range pres.Slides {
		text.WriteString(fmt.Sprintf("--- Slide %d ---\n", i+1))

		for _, element := range slide.PageElements {
			extractTextFromElement(element, &text)
		}

		// Extract speaker notes
		if slide.SlideProperties != nil && slide.SlideProperties.NotesPage != nil {
			for _, element := range slide.SlideProperties.NotesPage.PageElements {
				if element.Shape != nil && element.Shape.ShapeType == "TEXT_BOX" {
					if element.Shape.Text != nil {
						notes := extractTextContent(element.Shape.Text)
						if notes != "" {
							text.WriteString(fmt.Sprintf("\nSpeaker Notes:\n%s\n", notes))
						}
					}
				}
			}
		}

		text.WriteString("\n")
	}

	return text.String(), nil
}

// Export exports the presentation in the specified format.
func (s *Service) Export(presentationID, format string) ([]byte, error) {
	var mimeType string
	switch format {
	case "pdf":
		mimeType = "application/pdf"
	case "pptx":
		mimeType = "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	case "odp":
		mimeType = "application/vnd.oasis.opendocument.presentation"
	case "png":
		// For PNG, we'd need to export individual slides via thumbnail API
		// This is a simplified implementation
		mimeType = "application/pdf"
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	resp, err := s.drive.Files.Export(presentationID, mimeType).Download()
	if err != nil {
		return nil, fmt.Errorf("failed to export presentation: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read export: %w", err)
	}

	return data, nil
}

// parseSlideContent extracts content from a slide.
func parseSlideContent(slide *slides.Page, index int) *SlideContent {
	content := &SlideContent{
		SlideID:  slide.ObjectId,
		Index:    index,
		Elements: make([]ElementInfo, 0),
	}

	for _, element := range slide.PageElements {
		info := parsePageElement(element)
		if info != nil {
			content.Elements = append(content.Elements, *info)
		}
	}

	// Extract speaker notes
	if slide.SlideProperties != nil && slide.SlideProperties.NotesPage != nil {
		var notes strings.Builder
		for _, element := range slide.SlideProperties.NotesPage.PageElements {
			if element.Shape != nil && element.Shape.Text != nil {
				notes.WriteString(extractTextContent(element.Shape.Text))
			}
		}
		content.Notes = strings.TrimSpace(notes.String())
	}

	return content
}

// parsePageElement extracts information from a page element.
func parsePageElement(element *slides.PageElement) *ElementInfo {
	info := &ElementInfo{
		ID: element.ObjectId,
	}

	// Get position and size
	if element.Transform != nil {
		info.X = element.Transform.TranslateX
		info.Y = element.Transform.TranslateY
	}
	if element.Size != nil {
		if element.Size.Width != nil {
			info.Width = element.Size.Width.Magnitude
		}
		if element.Size.Height != nil {
			info.Height = element.Size.Height.Magnitude
		}
	}

	switch {
	case element.Shape != nil:
		info.Type = "shape"
		if element.Shape.Text != nil {
			info.Text = extractTextContent(element.Shape.Text)
		}
	case element.Image != nil:
		info.Type = "image"
	case element.Table != nil:
		info.Type = "table"
		info.Text = extractTableText(element.Table)
	case element.Line != nil:
		info.Type = "line"
	case element.Video != nil:
		info.Type = "video"
	case element.SheetsChart != nil:
		info.Type = "chart"
	default:
		return nil
	}

	return info
}

// extractTextFromElement extracts text from a page element and writes to builder.
func extractTextFromElement(element *slides.PageElement, text *strings.Builder) {
	if element.Shape != nil && element.Shape.Text != nil {
		content := extractTextContent(element.Shape.Text)
		if content != "" {
			text.WriteString(content)
			text.WriteString("\n")
		}
	}

	if element.Table != nil {
		tableText := extractTableText(element.Table)
		if tableText != "" {
			text.WriteString(tableText)
			text.WriteString("\n")
		}
	}
}

// extractTextContent extracts plain text from TextContent.
func extractTextContent(tc *slides.TextContent) string {
	if tc == nil {
		return ""
	}

	var text strings.Builder
	for _, elem := range tc.TextElements {
		if elem.TextRun != nil {
			text.WriteString(elem.TextRun.Content)
		}
	}
	return strings.TrimSpace(text.String())
}

// extractTableText extracts text from a table.
func extractTableText(table *slides.Table) string {
	var text strings.Builder
	for _, row := range table.TableRows {
		for i, cell := range row.TableCells {
			if cell.Text != nil {
				text.WriteString(extractTextContent(cell.Text))
			}
			if i < len(row.TableCells)-1 {
				text.WriteString("\t")
			}
		}
		text.WriteString("\n")
	}
	return strings.TrimSpace(text.String())
}
