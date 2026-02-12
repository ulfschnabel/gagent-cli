// Package docs provides Google Docs API operations.
package docs

import (
	"context"
	"fmt"
	"net/http"

	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// Service wraps the Docs API service.
type Service struct {
	docs  *docs.Service
	drive *drive.Service
}

// NewService creates a new Docs service.
func NewService(ctx context.Context, client *http.Client) (*Service, error) {
	docsSvc, err := docs.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create Docs service: %w", err)
	}

	driveSvc, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create Drive service: %w", err)
	}

	return &Service{docs: docsSvc, drive: driveSvc}, nil
}

// DocumentSummary represents a summary of a Google Doc.
type DocumentSummary struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	LastModified string `json:"last_modified"`
	CreatedTime  string `json:"created_time,omitempty"`
	WebViewLink  string `json:"web_view_link,omitempty"`
}

// DocumentContent represents the content of a document.
type DocumentContent struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

// DocumentOutline represents the structure of a document.
type DocumentOutline struct {
	ID       string           `json:"id"`
	Title    string           `json:"title"`
	Sections []SectionInfo    `json:"sections"`
}

// SectionInfo represents a section in a document.
type SectionInfo struct {
	Heading    string `json:"heading"`
	Level      int    `json:"level"`
	StartIndex int64  `json:"start_index"`
	EndIndex   int64  `json:"end_index,omitempty"`
}

// CreateResult represents the result of creating a document.
type CreateResult struct {
	DocumentID string `json:"document_id"`
	Title      string `json:"title"`
	RevisionID string `json:"revision_id,omitempty"`
}

// UpdateResult represents the result of updating a document.
type UpdateResult struct {
	DocumentID string `json:"document_id"`
	RevisionID string `json:"revision_id,omitempty"`
}

// DocumentStructure represents a detailed structural analysis of a document.
type DocumentStructure struct {
	ID        string       `json:"id"`
	Title     string       `json:"title"`
	WordCount int          `json:"word_count"`
	Headings  []HeadingInfo `json:"headings,omitempty"`
	Tables    []TableInfo  `json:"tables,omitempty"`
	Lists     []ListInfo   `json:"lists,omitempty"`
}

// HeadingInfo describes a heading in the document.
type HeadingInfo struct {
	Text       string `json:"text"`
	Level      int    `json:"level"`
	StartIndex int64  `json:"start_index"`
}

// TableInfo describes a table in the document.
type TableInfo struct {
	Rows       int      `json:"rows"`
	Columns    int      `json:"columns"`
	StartIndex int64    `json:"start_index"`
	FirstRow   []string `json:"first_row,omitempty"`
}

// ListInfo describes a list in the document.
type ListInfo struct {
	ListID    string `json:"list_id"`
	GlyphType string `json:"glyph_type,omitempty"`
	ItemCount int    `json:"item_count"`
}
