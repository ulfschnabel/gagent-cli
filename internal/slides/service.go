// Package slides provides Google Slides API operations.
package slides

import (
	"context"
	"fmt"
	"net/http"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"google.golang.org/api/slides/v1"
)

// Service wraps the Slides API service.
type Service struct {
	slides *slides.Service
	drive  *drive.Service
}

// NewService creates a new Slides service.
func NewService(ctx context.Context, client *http.Client) (*Service, error) {
	slidesSvc, err := slides.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create Slides service: %w", err)
	}

	driveSvc, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create Drive service: %w", err)
	}

	return &Service{slides: slidesSvc, drive: driveSvc}, nil
}

// PresentationSummary represents a summary of a presentation.
type PresentationSummary struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	LastModified string `json:"last_modified"`
	WebViewLink  string `json:"web_view_link,omitempty"`
}

// PresentationInfo represents metadata about a presentation.
type PresentationInfo struct {
	ID         string `json:"presentation_id"`
	Title      string `json:"title"`
	SlideCount int    `json:"slide_count"`
	Width      int64  `json:"width_emu,omitempty"`
	Height     int64  `json:"height_emu,omitempty"`
	Locale     string `json:"locale,omitempty"`
}

// SlideContent represents the content of a slide.
type SlideContent struct {
	SlideID    string        `json:"slide_id"`
	Index      int           `json:"index"`
	Elements   []ElementInfo `json:"elements"`
	Notes      string        `json:"notes,omitempty"`
}

// ElementInfo represents information about a page element.
// All coordinates and dimensions are in points (PT), converted from the API's
// internal EMU (English Metric Units) representation.
type ElementInfo struct {
	ID       string  `json:"id"`
	Type     string  `json:"type"`
	Text     string  `json:"text,omitempty"`
	X        float64 `json:"x_pt,omitempty"`      // X coordinate in points
	Y        float64 `json:"y_pt,omitempty"`      // Y coordinate in points
	Width    float64 `json:"width_pt,omitempty"`  // Width in points
	Height   float64 `json:"height_pt,omitempty"` // Height in points
}

// CreateResult represents the result of creating a presentation.
type CreateResult struct {
	PresentationID string `json:"presentation_id"`
	Title          string `json:"title"`
	URL            string `json:"url,omitempty"`
}

// UpdateResult represents the result of an update operation.
type UpdateResult struct {
	PresentationID string `json:"presentation_id"`
}
