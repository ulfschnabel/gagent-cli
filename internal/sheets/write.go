package sheets

import (
	"encoding/json"
	"fmt"

	"google.golang.org/api/sheets/v4"
)

// Create creates a new spreadsheet.
func (s *Service) Create(title string, sheetNames []string) (*CreateResult, error) {
	ss := &sheets.Spreadsheet{
		Properties: &sheets.SpreadsheetProperties{
			Title: title,
		},
	}

	if len(sheetNames) > 0 {
		ss.Sheets = make([]*sheets.Sheet, 0, len(sheetNames))
		for i, name := range sheetNames {
			ss.Sheets = append(ss.Sheets, &sheets.Sheet{
				Properties: &sheets.SheetProperties{
					Title: name,
					Index: int64(i),
				},
			})
		}
	}

	created, err := s.sheets.Spreadsheets.Create(ss).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create spreadsheet: %w", err)
	}

	return &CreateResult{
		SpreadsheetID: created.SpreadsheetId,
		Title:         created.Properties.Title,
		URL:           created.SpreadsheetUrl,
	}, nil
}

// Write writes values to a range.
func (s *Service) Write(spreadsheetID, sheetName, rangeStr string, values [][]any) (*UpdateResult, error) {
	writeRange := rangeStr
	if sheetName != "" {
		writeRange = fmt.Sprintf("%s!%s", sheetName, rangeStr)
	}

	vr := &sheets.ValueRange{
		Values: values,
	}

	resp, err := s.sheets.Spreadsheets.Values.Update(spreadsheetID, writeRange, vr).
		ValueInputOption("USER_ENTERED").
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to write values: %w", err)
	}

	return &UpdateResult{
		SpreadsheetID: resp.SpreadsheetId,
		UpdatedRange:  resp.UpdatedRange,
		UpdatedRows:   resp.UpdatedRows,
		UpdatedCells:  resp.UpdatedCells,
	}, nil
}

// Append appends values to the end of data in a sheet.
func (s *Service) Append(spreadsheetID, sheetName string, values [][]any) (*UpdateResult, error) {
	appendRange := sheetName
	if appendRange == "" {
		appendRange = "Sheet1"
	}

	vr := &sheets.ValueRange{
		Values: values,
	}

	resp, err := s.sheets.Spreadsheets.Values.Append(spreadsheetID, appendRange, vr).
		ValueInputOption("USER_ENTERED").
		InsertDataOption("INSERT_ROWS").
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to append values: %w", err)
	}

	return &UpdateResult{
		SpreadsheetID: spreadsheetID,
		UpdatedRange:  resp.Updates.UpdatedRange,
		UpdatedRows:   resp.Updates.UpdatedRows,
		UpdatedCells:  resp.Updates.UpdatedCells,
	}, nil
}

// Clear clears values in a range (preserves formatting).
func (s *Service) Clear(spreadsheetID, sheetName, rangeStr string) (*UpdateResult, error) {
	clearRange := rangeStr
	if sheetName != "" {
		clearRange = fmt.Sprintf("%s!%s", sheetName, rangeStr)
	}

	_, err := s.sheets.Spreadsheets.Values.Clear(spreadsheetID, clearRange, &sheets.ClearValuesRequest{}).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to clear values: %w", err)
	}

	return &UpdateResult{
		SpreadsheetID: spreadsheetID,
		UpdatedRange:  clearRange,
	}, nil
}

// AddSheet adds a new sheet to an existing spreadsheet.
func (s *Service) AddSheet(spreadsheetID, sheetName string) (*UpdateResult, error) {
	req := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				AddSheet: &sheets.AddSheetRequest{
					Properties: &sheets.SheetProperties{
						Title: sheetName,
					},
				},
			},
		},
	}

	_, err := s.sheets.Spreadsheets.BatchUpdate(spreadsheetID, req).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to add sheet: %w", err)
	}

	return &UpdateResult{
		SpreadsheetID: spreadsheetID,
	}, nil
}

// DeleteSheet deletes a sheet from a spreadsheet.
func (s *Service) DeleteSheet(spreadsheetID, sheetName string) (*UpdateResult, error) {
	// First, get the sheet ID
	ss, err := s.sheets.Spreadsheets.Get(spreadsheetID).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get spreadsheet: %w", err)
	}

	var sheetID int64 = -1
	for _, sheet := range ss.Sheets {
		if sheet.Properties.Title == sheetName {
			sheetID = sheet.Properties.SheetId
			break
		}
	}

	if sheetID == -1 {
		return nil, fmt.Errorf("sheet not found: %s", sheetName)
	}

	req := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				DeleteSheet: &sheets.DeleteSheetRequest{
					SheetId: sheetID,
				},
			},
		},
	}

	_, err = s.sheets.Spreadsheets.BatchUpdate(spreadsheetID, req).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to delete sheet: %w", err)
	}

	return &UpdateResult{
		SpreadsheetID: spreadsheetID,
	}, nil
}

// ValuesUpdate updates values using the raw API.
func (s *Service) ValuesUpdate(spreadsheetID, rangeStr, valuesJSON string) (*UpdateResult, error) {
	var values [][]any
	if err := json.Unmarshal([]byte(valuesJSON), &values); err != nil {
		return nil, fmt.Errorf("failed to parse values JSON: %w", err)
	}

	vr := &sheets.ValueRange{
		Values: values,
	}

	resp, err := s.sheets.Spreadsheets.Values.Update(spreadsheetID, rangeStr, vr).
		ValueInputOption("USER_ENTERED").
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to update values: %w", err)
	}

	return &UpdateResult{
		SpreadsheetID: resp.SpreadsheetId,
		UpdatedRange:  resp.UpdatedRange,
		UpdatedRows:   resp.UpdatedRows,
		UpdatedCells:  resp.UpdatedCells,
	}, nil
}

// BatchUpdate performs a batch update with raw requests.
func (s *Service) BatchUpdate(spreadsheetID, requestsJSON string) (*UpdateResult, error) {
	var requests []*sheets.Request
	if err := json.Unmarshal([]byte(requestsJSON), &requests); err != nil {
		return nil, fmt.Errorf("failed to parse requests JSON: %w", err)
	}

	req := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: requests,
	}

	_, err := s.sheets.Spreadsheets.BatchUpdate(spreadsheetID, req).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to batch update: %w", err)
	}

	return &UpdateResult{
		SpreadsheetID: spreadsheetID,
	}, nil
}

// ValuesAppend appends values using the raw API.
func (s *Service) ValuesAppend(spreadsheetID, rangeStr, valuesJSON string) (*UpdateResult, error) {
	var values [][]any
	if err := json.Unmarshal([]byte(valuesJSON), &values); err != nil {
		return nil, fmt.Errorf("failed to parse values JSON: %w", err)
	}

	vr := &sheets.ValueRange{
		Values: values,
	}

	resp, err := s.sheets.Spreadsheets.Values.Append(spreadsheetID, rangeStr, vr).
		ValueInputOption("USER_ENTERED").
		InsertDataOption("INSERT_ROWS").
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to append values: %w", err)
	}

	return &UpdateResult{
		SpreadsheetID: spreadsheetID,
		UpdatedRange:  resp.Updates.UpdatedRange,
		UpdatedRows:   resp.Updates.UpdatedRows,
		UpdatedCells:  resp.Updates.UpdatedCells,
	}, nil
}
