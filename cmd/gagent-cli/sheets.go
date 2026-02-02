package main

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ulfhaga/gagent-cli/internal/auth"
	"github.com/ulfhaga/gagent-cli/internal/config"
	"github.com/ulfhaga/gagent-cli/internal/output"
	"github.com/ulfhaga/gagent-cli/internal/sheets"
)

// sheetsReadService creates a Sheets service with read scope.
func sheetsReadService(ctx context.Context) (*sheets.Service, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	configDir, err := config.GetConfigDir()
	if err != nil {
		return nil, err
	}

	if err := auth.RequireScope(configDir, auth.ScopeRead); err != nil {
		return nil, err
	}

	client, err := auth.GetClient(ctx, configDir, cfg.ClientID, cfg.ClientSecret, auth.ScopeRead)
	if err != nil {
		return nil, err
	}

	return sheets.NewService(ctx, client)
}

// sheetsWriteService creates a Sheets service with write scope.
func sheetsWriteService(ctx context.Context) (*sheets.Service, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	configDir, err := config.GetConfigDir()
	if err != nil {
		return nil, err
	}

	if err := auth.RequireScope(configDir, auth.ScopeWrite); err != nil {
		return nil, err
	}

	client, err := auth.GetClient(ctx, configDir, cfg.ClientID, cfg.ClientSecret, auth.ScopeWrite)
	if err != nil {
		return nil, err
	}

	return sheets.NewService(ctx, client)
}

func sheetsListCmd() *cobra.Command {
	var limit int64
	var query string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List spreadsheets",
		Long:  "Lists spreadsheets from Drive with title, id, last modified.",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := sheetsReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			spreadsheets, _, err := svc.List(sheets.ListOptions{
				Query:      query,
				MaxResults: limit,
			})
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"spreadsheets": spreadsheets,
				"count":        len(spreadsheets),
			}, "read")
		},
	}

	cmd.Flags().Int64VarP(&limit, "limit", "n", 10, "Maximum number of spreadsheets")
	cmd.Flags().StringVar(&query, "query", "", "Search query")

	return cmd
}

func sheetsReadCmd() *cobra.Command {
	var sheet, rangeStr string

	cmd := &cobra.Command{
		Use:   "read <spreadsheet-id>",
		Short: "Read spreadsheet values",
		Long:  "Returns cell values as 2D array.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := sheetsReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			result, err := svc.Read(args[0], sheet, rangeStr)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(result, "read")
		},
	}

	cmd.Flags().StringVar(&sheet, "sheet", "", "Sheet name")
	cmd.Flags().StringVar(&rangeStr, "range", "", "Cell range (e.g., A1:Z100)")

	return cmd
}

func sheetsInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info <spreadsheet-id>",
		Short: "Get spreadsheet metadata",
		Long:  "Returns spreadsheet metadata: sheets list, properties.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := sheetsReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			info, err := svc.Info(args[0])
			if err != nil {
				output.NotFoundError("Spreadsheet", args[0])
				return
			}

			output.Success(info, "read")
		},
	}
}

func sheetsExportCmd() *cobra.Command {
	var format, sheet, outputFile string

	cmd := &cobra.Command{
		Use:   "export <spreadsheet-id>",
		Short: "Export spreadsheet",
		Long:  "Exports spreadsheet in specified format (csv, xlsx, pdf).",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := sheetsReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			data, err := svc.Export(args[0], format, sheet)
			if err != nil {
				output.APIError(err)
				return
			}

			if outputFile != "" {
				if err := os.WriteFile(outputFile, data, 0644); err != nil {
					output.FailureFromError(output.ErrInternal, err)
					return
				}
				output.Success(map[string]interface{}{
					"spreadsheet_id": args[0],
					"format":         format,
					"saved_to":       outputFile,
					"size":           len(data),
				}, "read")
				return
			}

			if format == "csv" {
				output.Success(map[string]interface{}{
					"spreadsheet_id": args[0],
					"format":         format,
					"content":        string(data),
				}, "read")
				return
			}

			output.Success(map[string]interface{}{
				"spreadsheet_id": args[0],
				"format":         format,
				"size":           len(data),
				"note":           "Use --output to save binary formats to a file",
			}, "read")
		},
	}

	cmd.Flags().StringVar(&format, "format", "csv", "Export format: csv, xlsx, pdf")
	cmd.Flags().StringVar(&sheet, "sheet", "", "Sheet name (for CSV)")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path")

	return cmd
}

func sheetsQueryCmd() *cobra.Command {
	var sheet, where string

	cmd := &cobra.Command{
		Use:   "query <spreadsheet-id>",
		Short: "Query spreadsheet",
		Long:  "Simple query syntax for filtering rows.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := sheetsReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			result, err := svc.Query(args[0], sheet, where)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(result, "read")
		},
	}

	cmd.Flags().StringVar(&sheet, "sheet", "", "Sheet name (required)")
	cmd.Flags().StringVar(&where, "where", "", "Filter condition")

	cmd.MarkFlagRequired("sheet")

	return cmd
}

func sheetsCreateCmd() *cobra.Command {
	var title, sheetNames string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create spreadsheet",
		Long:  "Creates new spreadsheet with optional sheet names.",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := sheetsWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			var sheets []string
			if sheetNames != "" {
				sheets = strings.Split(sheetNames, ",")
			}

			result, err := svc.Create(title, sheets)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(result, "write")
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "Spreadsheet title (required)")
	cmd.Flags().StringVar(&sheetNames, "sheets", "", "Sheet names (comma-separated)")

	cmd.MarkFlagRequired("title")

	return cmd
}

func sheetsWriteCmd() *cobra.Command {
	var sheet, rangeStr, values string

	cmd := &cobra.Command{
		Use:   "write <spreadsheet-id>",
		Short: "Write values",
		Long:  "Writes values to range (JSON 2D array format).",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := sheetsWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			var vals [][]any
			if err := json.Unmarshal([]byte(values), &vals); err != nil {
				output.InvalidInputError("Invalid values JSON: " + err.Error())
				return
			}

			result, err := svc.Write(args[0], sheet, rangeStr, vals)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(result, "write")
		},
	}

	cmd.Flags().StringVar(&sheet, "sheet", "", "Sheet name (required)")
	cmd.Flags().StringVar(&rangeStr, "range", "A1", "Starting cell")
	cmd.Flags().StringVar(&values, "values", "", "Values as JSON 2D array (required)")

	cmd.MarkFlagRequired("sheet")
	cmd.MarkFlagRequired("values")

	return cmd
}

func sheetsAppendCmd() *cobra.Command {
	var sheet, values string

	cmd := &cobra.Command{
		Use:   "append <spreadsheet-id>",
		Short: "Append rows",
		Long:  "Appends row(s) to the end of data in sheet.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := sheetsWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			var vals [][]any
			if err := json.Unmarshal([]byte(values), &vals); err != nil {
				output.InvalidInputError("Invalid values JSON: " + err.Error())
				return
			}

			result, err := svc.Append(args[0], sheet, vals)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(result, "write")
		},
	}

	cmd.Flags().StringVar(&sheet, "sheet", "", "Sheet name (required)")
	cmd.Flags().StringVar(&values, "values", "", "Values as JSON 2D array (required)")

	cmd.MarkFlagRequired("sheet")
	cmd.MarkFlagRequired("values")

	return cmd
}

func sheetsClearCmd() *cobra.Command {
	var sheet, rangeStr string

	cmd := &cobra.Command{
		Use:   "clear <spreadsheet-id>",
		Short: "Clear values",
		Long:  "Clears cell values in range (preserves formatting).",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := sheetsWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			result, err := svc.Clear(args[0], sheet, rangeStr)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"spreadsheet_id": result.SpreadsheetID,
				"cleared_range":  result.UpdatedRange,
			}, "write")
		},
	}

	cmd.Flags().StringVar(&sheet, "sheet", "", "Sheet name (required)")
	cmd.Flags().StringVar(&rangeStr, "range", "", "Range to clear (required)")

	cmd.MarkFlagRequired("sheet")
	cmd.MarkFlagRequired("range")

	return cmd
}

func sheetsAddSheetCmd() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "add-sheet <spreadsheet-id>",
		Short: "Add a sheet",
		Long:  "Adds a new sheet to existing spreadsheet.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := sheetsWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			result, err := svc.AddSheet(args[0], name)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"spreadsheet_id": result.SpreadsheetID,
				"sheet_name":     name,
				"added":          true,
			}, "write")
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Sheet name (required)")
	cmd.MarkFlagRequired("name")

	return cmd
}

func sheetsDeleteSheetCmd() *cobra.Command {
	var sheet string

	cmd := &cobra.Command{
		Use:   "delete-sheet <spreadsheet-id>",
		Short: "Delete a sheet",
		Long:  "Deletes a sheet from spreadsheet.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := sheetsWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			result, err := svc.DeleteSheet(args[0], sheet)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"spreadsheet_id": result.SpreadsheetID,
				"sheet_name":     sheet,
				"deleted":        true,
			}, "write")
		},
	}

	cmd.Flags().StringVar(&sheet, "sheet", "", "Sheet name (required)")
	cmd.MarkFlagRequired("sheet")

	return cmd
}

// Sheets API commands
func sheetsAPICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api",
		Short: "Low-level Sheets API commands",
		Long:  "Direct access to Sheets API operations.",
	}

	cmd.AddCommand(sheetsAPIGetCmd())
	cmd.AddCommand(sheetsAPIValuesCmd())
	cmd.AddCommand(sheetsAPIUpdateCmd())
	cmd.AddCommand(sheetsAPIBatchUpdateCmd())
	cmd.AddCommand(sheetsAPIAppendCmd())

	return cmd
}

func sheetsAPIGetCmd() *cobra.Command {
	var includeGridData bool
	var ranges string

	cmd := &cobra.Command{
		Use:   "get <spreadsheet-id>",
		Short: "Get spreadsheet structure",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := sheetsReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			var rangeList []string
			if ranges != "" {
				rangeList = strings.Split(ranges, ",")
			}

			info, err := svc.GetRaw(args[0], includeGridData, rangeList)
			if err != nil {
				output.NotFoundError("Spreadsheet", args[0])
				return
			}

			output.Success(info, "read")
		},
	}

	cmd.Flags().BoolVar(&includeGridData, "include-grid-data", false, "Include cell data")
	cmd.Flags().StringVar(&ranges, "ranges", "", "Ranges to include (comma-separated)")

	return cmd
}

func sheetsAPIValuesCmd() *cobra.Command {
	var rangeStr string

	cmd := &cobra.Command{
		Use:   "values <spreadsheet-id>",
		Short: "Get values from range",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := sheetsReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			result, err := svc.ValuesGet(args[0], rangeStr)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(result, "read")
		},
	}

	cmd.Flags().StringVar(&rangeStr, "range", "", "Range (required, e.g., Sheet1!A1:Z100)")
	cmd.MarkFlagRequired("range")

	return cmd
}

func sheetsAPIUpdateCmd() *cobra.Command {
	var rangeStr, valuesJSON string

	cmd := &cobra.Command{
		Use:   "update <spreadsheet-id>",
		Short: "Update values",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := sheetsWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			result, err := svc.ValuesUpdate(args[0], rangeStr, valuesJSON)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(result, "write")
		},
	}

	cmd.Flags().StringVar(&rangeStr, "range", "", "Range (required)")
	cmd.Flags().StringVar(&valuesJSON, "values-json", "", "Values JSON (required)")

	cmd.MarkFlagRequired("range")
	cmd.MarkFlagRequired("values-json")

	return cmd
}

func sheetsAPIBatchUpdateCmd() *cobra.Command {
	var requestsJSON string

	cmd := &cobra.Command{
		Use:   "batch-update <spreadsheet-id>",
		Short: "Batch update spreadsheet",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := sheetsWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			result, err := svc.BatchUpdate(args[0], requestsJSON)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(result, "write")
		},
	}

	cmd.Flags().StringVar(&requestsJSON, "requests-json", "", "Array of request objects (required)")
	cmd.MarkFlagRequired("requests-json")

	return cmd
}

func sheetsAPIAppendCmd() *cobra.Command {
	var rangeStr, valuesJSON string

	cmd := &cobra.Command{
		Use:   "append <spreadsheet-id>",
		Short: "Append values",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := sheetsWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			result, err := svc.ValuesAppend(args[0], rangeStr, valuesJSON)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(result, "write")
		},
	}

	cmd.Flags().StringVar(&rangeStr, "range", "", "Range (required)")
	cmd.Flags().StringVar(&valuesJSON, "values-json", "", "Values JSON (required)")

	cmd.MarkFlagRequired("range")
	cmd.MarkFlagRequired("values-json")

	return cmd
}
