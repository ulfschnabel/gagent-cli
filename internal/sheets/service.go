// Package sheets provides Google Sheets API operations.
package sheets

import (
	"context"
	"fmt"
	"net/http"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// Service wraps the Sheets API service.
type Service struct {
	sheets *sheets.Service
	drive  *drive.Service
}

// NewService creates a new Sheets service.
func NewService(ctx context.Context, client *http.Client) (*Service, error) {
	sheetsSvc, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create Sheets service: %w", err)
	}

	driveSvc, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create Drive service: %w", err)
	}

	return &Service{sheets: sheetsSvc, drive: driveSvc}, nil
}

// SpreadsheetSummary represents a summary of a spreadsheet.
type SpreadsheetSummary struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	LastModified string `json:"last_modified"`
	WebViewLink  string `json:"web_view_link,omitempty"`
}

// SpreadsheetInfo represents metadata about a spreadsheet.
type SpreadsheetInfo struct {
	ID         string      `json:"spreadsheet_id"`
	Title      string      `json:"title"`
	Sheets     []SheetInfo `json:"sheets"`
	Locale     string      `json:"locale,omitempty"`
	TimeZone   string      `json:"time_zone,omitempty"`
}

// SheetInfo represents information about a sheet within a spreadsheet.
type SheetInfo struct {
	ID          int64  `json:"sheet_id"`
	Name        string `json:"name"`
	Index       int64  `json:"index"`
	RowCount    int64  `json:"row_count"`
	ColumnCount int64  `json:"column_count"`
}

// ValuesResult represents the result of reading values.
type ValuesResult struct {
	Range  string     `json:"range"`
	Values [][]any `json:"values"`
}

// UpdateResult represents the result of an update operation.
type UpdateResult struct {
	SpreadsheetID string `json:"spreadsheet_id"`
	UpdatedRange  string `json:"updated_range,omitempty"`
	UpdatedRows   int64  `json:"updated_rows,omitempty"`
	UpdatedCells  int64  `json:"updated_cells,omitempty"`
}

// CreateResult represents the result of creating a spreadsheet.
type CreateResult struct {
	SpreadsheetID string `json:"spreadsheet_id"`
	Title         string `json:"title"`
	URL           string `json:"url"`
}
