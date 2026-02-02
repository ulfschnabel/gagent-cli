package sheets

import (
	"fmt"
	"io"
)

// ListOptions contains options for listing spreadsheets.
type ListOptions struct {
	Query      string
	MaxResults int64
	PageToken  string
}

// List returns a list of spreadsheets from Drive.
func (s *Service) List(opts ListOptions) ([]SpreadsheetSummary, string, error) {
	query := "mimeType='application/vnd.google-apps.spreadsheet'"
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
		return nil, "", fmt.Errorf("failed to list spreadsheets: %w", err)
	}

	spreadsheets := make([]SpreadsheetSummary, 0, len(resp.Files))
	for _, file := range resp.Files {
		spreadsheets = append(spreadsheets, SpreadsheetSummary{
			ID:           file.Id,
			Title:        file.Name,
			LastModified: file.ModifiedTime,
			WebViewLink:  file.WebViewLink,
		})
	}

	return spreadsheets, resp.NextPageToken, nil
}

// Info returns metadata about a spreadsheet.
func (s *Service) Info(spreadsheetID string) (*SpreadsheetInfo, error) {
	ss, err := s.sheets.Spreadsheets.Get(spreadsheetID).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get spreadsheet: %w", err)
	}

	sheets := make([]SheetInfo, 0, len(ss.Sheets))
	for _, sheet := range ss.Sheets {
		props := sheet.Properties
		grid := props.GridProperties
		sheets = append(sheets, SheetInfo{
			ID:          props.SheetId,
			Name:        props.Title,
			Index:       props.Index,
			RowCount:    grid.RowCount,
			ColumnCount: grid.ColumnCount,
		})
	}

	return &SpreadsheetInfo{
		ID:       ss.SpreadsheetId,
		Title:    ss.Properties.Title,
		Sheets:   sheets,
		Locale:   ss.Properties.Locale,
		TimeZone: ss.Properties.TimeZone,
	}, nil
}

// Read returns cell values from a spreadsheet.
func (s *Service) Read(spreadsheetID, sheetName, rangeStr string) (*ValuesResult, error) {
	// Build the range
	readRange := rangeStr
	if sheetName != "" {
		if rangeStr != "" {
			readRange = fmt.Sprintf("%s!%s", sheetName, rangeStr)
		} else {
			readRange = sheetName
		}
	}
	if readRange == "" {
		readRange = "A:ZZ" // Default to all columns
	}

	resp, err := s.sheets.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to read values: %w", err)
	}

	return &ValuesResult{
		Range:  resp.Range,
		Values: resp.Values,
	}, nil
}

// GetRaw returns the full spreadsheet structure.
func (s *Service) GetRaw(spreadsheetID string, includeGridData bool, ranges []string) (*SpreadsheetInfo, error) {
	call := s.sheets.Spreadsheets.Get(spreadsheetID)

	if includeGridData {
		call = call.IncludeGridData(true)
	}

	if len(ranges) > 0 {
		call = call.Ranges(ranges...)
	}

	ss, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get spreadsheet: %w", err)
	}

	sheets := make([]SheetInfo, 0, len(ss.Sheets))
	for _, sheet := range ss.Sheets {
		props := sheet.Properties
		grid := props.GridProperties
		sheets = append(sheets, SheetInfo{
			ID:          props.SheetId,
			Name:        props.Title,
			Index:       props.Index,
			RowCount:    grid.RowCount,
			ColumnCount: grid.ColumnCount,
		})
	}

	return &SpreadsheetInfo{
		ID:       ss.SpreadsheetId,
		Title:    ss.Properties.Title,
		Sheets:   sheets,
		Locale:   ss.Properties.Locale,
		TimeZone: ss.Properties.TimeZone,
	}, nil
}

// ValuesGet returns values from a range using the raw API.
func (s *Service) ValuesGet(spreadsheetID, rangeStr string) (*ValuesResult, error) {
	resp, err := s.sheets.Spreadsheets.Values.Get(spreadsheetID, rangeStr).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get values: %w", err)
	}

	return &ValuesResult{
		Range:  resp.Range,
		Values: resp.Values,
	}, nil
}

// Export exports the spreadsheet in the specified format.
func (s *Service) Export(spreadsheetID, format, sheetName string) ([]byte, error) {
	var mimeType string
	switch format {
	case "csv":
		mimeType = "text/csv"
	case "xlsx":
		mimeType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case "pdf":
		mimeType = "application/pdf"
	case "ods":
		mimeType = "application/vnd.oasis.opendocument.spreadsheet"
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	// Note: For CSV, Google exports only the first sheet by default
	// To export a specific sheet, we'd need to use a different approach
	resp, err := s.drive.Files.Export(spreadsheetID, mimeType).Download()
	if err != nil {
		return nil, fmt.Errorf("failed to export spreadsheet: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read export: %w", err)
	}

	return data, nil
}

// Query performs a simple query on spreadsheet data.
// This is a simplified implementation that filters rows based on column values.
func (s *Service) Query(spreadsheetID, sheetName, where string) (*ValuesResult, error) {
	// Read all data from the sheet
	result, err := s.Read(spreadsheetID, sheetName, "")
	if err != nil {
		return nil, err
	}

	// For now, return all data - a full implementation would parse the where clause
	// and filter rows accordingly
	return result, nil
}
