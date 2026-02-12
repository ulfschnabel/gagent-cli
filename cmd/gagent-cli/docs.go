package main

import (
	"context"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ulfhaga/gagent-cli/internal/auth"
	"github.com/ulfhaga/gagent-cli/internal/config"
	"github.com/ulfhaga/gagent-cli/internal/docs"
	"github.com/ulfhaga/gagent-cli/internal/output"
)

// docsReadService creates a Docs service with read scope.
func docsReadService(ctx context.Context) (*docs.Service, error) {
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

	return docs.NewService(ctx, client)
}

// docsWriteService creates a Docs service with write scope.
func docsWriteService(ctx context.Context) (*docs.Service, error) {
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

	return docs.NewService(ctx, client)
}

func docsListCmd() *cobra.Command {
	var limit int64
	var query string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List documents",
		Long:  "Lists documents from Drive with title, id, last modified.",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := docsReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			documents, _, err := svc.List(docs.ListOptions{
				Query:      query,
				MaxResults: limit,
			})
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"documents": documents,
				"count":     len(documents),
			}, "read")
		},
	}

	cmd.Flags().Int64VarP(&limit, "limit", "n", 10, "Maximum number of documents")
	cmd.Flags().StringVar(&query, "query", "", "Search query")

	return cmd
}

func docsReadCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "read <doc-id>",
		Short: "Read a document",
		Long:  "Returns document content as plain text.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := docsReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			content, err := svc.Read(args[0])
			if err != nil {
				output.NotFoundError("Document", args[0])
				return
			}

			output.Success(content, "read")
		},
	}
}

func docsExportCmd() *cobra.Command {
	var format, outputFile string

	cmd := &cobra.Command{
		Use:   "export <doc-id>",
		Short: "Export a document",
		Long:  "Exports document in specified format (txt, html, pdf, docx).",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := docsReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			data, err := svc.Export(args[0], format)
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
					"document_id": args[0],
					"format":      format,
					"saved_to":    outputFile,
					"size":        len(data),
				}, "read")
				return
			}

			// Return content as string for text formats
			if format == "txt" || format == "text" || format == "html" || format == "md" {
				output.Success(map[string]interface{}{
					"document_id": args[0],
					"format":      format,
					"content":     string(data),
				}, "read")
				return
			}

			// For binary formats, return base64
			output.Success(map[string]interface{}{
				"document_id": args[0],
				"format":      format,
				"size":        len(data),
				"note":        "Use --output to save binary formats to a file",
			}, "read")
		},
	}

	cmd.Flags().StringVar(&format, "format", "txt", "Export format: txt, html, pdf, docx")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path")

	return cmd
}

func docsOutlineCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "outline <doc-id>",
		Short: "Get document outline",
		Long:  "Returns document structure: headings, sections.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := docsReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			outline, err := svc.Outline(args[0])
			if err != nil {
				output.NotFoundError("Document", args[0])
				return
			}

			output.Success(outline, "read")
		},
	}
}

func docsCreateCmd() *cobra.Command {
	var title, content string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a document",
		Long:  "Creates new document, optionally with initial content.",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := docsWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			var result *docs.CreateResult
			if content != "" {
				result, err = svc.CreateWithContent(title, content)
			} else {
				result, err = svc.Create(title)
			}

			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(result, "write")
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "Document title (required)")
	cmd.Flags().StringVar(&content, "content", "", "Initial content")

	cmd.MarkFlagRequired("title")

	return cmd
}

func docsAppendCmd() *cobra.Command {
	var text string

	cmd := &cobra.Command{
		Use:   "append <doc-id>",
		Short: "Append text",
		Long:  "Appends text to end of document.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := docsWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			result, err := svc.Append(args[0], text)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(result, "write")
		},
	}

	cmd.Flags().StringVar(&text, "text", "", "Text to append (required)")
	cmd.MarkFlagRequired("text")

	return cmd
}

func docsPrependCmd() *cobra.Command {
	var text string

	cmd := &cobra.Command{
		Use:   "prepend <doc-id>",
		Short: "Prepend text",
		Long:  "Inserts text at beginning of document.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := docsWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			result, err := svc.Prepend(args[0], text)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(result, "write")
		},
	}

	cmd.Flags().StringVar(&text, "text", "", "Text to prepend (required)")
	cmd.MarkFlagRequired("text")

	return cmd
}

func docsReplaceTextCmd() *cobra.Command {
	var find, replace string
	var matchCase bool

	cmd := &cobra.Command{
		Use:   "replace-text <doc-id>",
		Short: "Find and replace text",
		Long:  "Find and replace all occurrences.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := docsWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			result, err := svc.ReplaceText(args[0], find, replace, matchCase)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"document_id": result.DocumentID,
				"find":        find,
				"replace":     replace,
				"match_case":  matchCase,
			}, "write")
		},
	}

	cmd.Flags().StringVar(&find, "find", "", "Text to find (required)")
	cmd.Flags().StringVar(&replace, "replace", "", "Replacement text (required)")
	cmd.Flags().BoolVar(&matchCase, "match-case", false, "Case-sensitive matching")

	cmd.MarkFlagRequired("find")
	cmd.MarkFlagRequired("replace")

	return cmd
}

func docsUpdateSectionCmd() *cobra.Command {
	var heading, content string

	cmd := &cobra.Command{
		Use:   "update-section <doc-id>",
		Short: "Update a section",
		Long:  "Finds section by heading, replaces its content.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := docsWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			result, err := svc.UpdateSection(args[0], heading, content)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"document_id": result.DocumentID,
				"heading":     heading,
				"updated":     true,
			}, "write")
		},
	}

	cmd.Flags().StringVar(&heading, "heading", "", "Section heading (required)")
	cmd.Flags().StringVar(&content, "content", "", "New section content (required)")

	cmd.MarkFlagRequired("heading")
	cmd.MarkFlagRequired("content")

	return cmd
}

// Docs API commands
func docsAPICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api",
		Short: "Low-level Docs API commands",
		Long:  "Direct access to Docs API operations.",
	}

	cmd.AddCommand(docsAPIGetCmd())
	cmd.AddCommand(docsAPIBatchUpdateCmd())
	cmd.AddCommand(docsAPICreateCmd())

	return cmd
}

func docsAPIGetCmd() *cobra.Command {
	var suggestionsView string

	cmd := &cobra.Command{
		Use:   "get <doc-id>",
		Short: "Get document JSON structure",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := docsReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			doc, err := svc.Get(args[0])
			if err != nil {
				output.NotFoundError("Document", args[0])
				return
			}

			output.Success(doc, "read")
		},
	}

	cmd.Flags().StringVar(&suggestionsView, "suggestions-view", "DEFAULT", "Suggestions view: PREVIEW, SUGGESTIONS_INLINE, DEFAULT")

	return cmd
}

func docsAPIBatchUpdateCmd() *cobra.Command {
	var requestsJSON string

	cmd := &cobra.Command{
		Use:   "batch-update <doc-id>",
		Short: "Batch update document",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := docsWriteService(ctx)
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

func docsAPICreateCmd() *cobra.Command {
	var title string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create empty document",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := docsWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			result, err := svc.Create(title)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(result, "write")
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "Document title (required)")
	cmd.MarkFlagRequired("title")

	return cmd
}

// docsInsertListCmd inserts a list into a document.
func docsInsertListCmd() *cobra.Command {
	var listType string
	var items []string
	var itemsStr string
	var indent int

	cmd := &cobra.Command{
		Use:   "insert-list <doc-id>",
		Short: "Insert a list",
		Long:  "Inserts a bullet, numbered, or other type of list.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := docsWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			// Parse items from comma-separated string if provided
			if itemsStr != "" {
				items = parseListItems(itemsStr)
			}

			result, err := svc.InsertList(args[0], docs.InsertListOptions{
				Type:   listType,
				Items:  items,
				Indent: indent,
			})
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"document_id": result.DocumentID,
				"list_type":   listType,
				"item_count":  len(items),
			}, "write")
		},
	}

	cmd.Flags().StringVar(&listType, "type", "bullet", "List type: bullet, numbered, lettered, roman, checklist")
	cmd.Flags().StringSliceVar(&items, "items", []string{}, "List items (can be repeated)")
	cmd.Flags().StringVar(&itemsStr, "items-string", "", "List items as comma-separated string")
	cmd.Flags().IntVar(&indent, "indent", 0, "Indent level (0-9)")

	return cmd
}

// docsAppendFormattedCmd appends formatted text to a document.
func docsAppendFormattedCmd() *cobra.Command {
	var text string
	var bold, italic, underline, strikethrough bool
	var fontSize int
	var fontFamily, color, bgColor, namedStyle string

	cmd := &cobra.Command{
		Use:   "append-formatted <doc-id>",
		Short: "Append formatted text",
		Long:  "Appends text with formatting (bold, italic, colors, etc.).",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := docsWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			result, err := svc.AppendFormatted(args[0], text, docs.TextStyleOptions{
				Bold:          bold,
				Italic:        italic,
				Underline:     underline,
				Strikethrough: strikethrough,
				FontSize:      fontSize,
				FontFamily:    fontFamily,
				Color:         color,
				BgColor:       bgColor,
				NamedStyle:    namedStyle,
			})
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(result, "write")
		},
	}

	cmd.Flags().StringVar(&text, "text", "", "Text to append (required)")
	cmd.Flags().BoolVar(&bold, "bold", false, "Bold text")
	cmd.Flags().BoolVar(&italic, "italic", false, "Italic text")
	cmd.Flags().BoolVar(&underline, "underline", false, "Underline text")
	cmd.Flags().BoolVar(&strikethrough, "strikethrough", false, "Strikethrough text")
	cmd.Flags().IntVar(&fontSize, "font-size", 0, "Font size (8-72)")
	cmd.Flags().StringVar(&fontFamily, "font-family", "", "Font family (e.g., Arial, Times New Roman)")
	cmd.Flags().StringVar(&color, "color", "", "Text color (#rrggbb)")
	cmd.Flags().StringVar(&bgColor, "bg-color", "", "Background color (#rrggbb)")
	cmd.Flags().StringVar(&namedStyle, "style", "", "Named style: heading1, heading2, heading3, heading4, title, subtitle, normal")

	cmd.MarkFlagRequired("text")

	return cmd
}

// docsFormatParagraphCmd formats a paragraph.
func docsFormatParagraphCmd() *cobra.Command {
	var startIndex, endIndex int64
	var alignment string
	var indentStart, indentEnd, indentFirst float64
	var lineSpacing, spacingBefore, spacingAfter float64

	cmd := &cobra.Command{
		Use:   "format-paragraph <doc-id>",
		Short: "Format a paragraph",
		Long:  "Applies paragraph formatting (alignment, indentation, spacing).",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := docsWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			result, err := svc.FormatParagraph(args[0], startIndex, endIndex, docs.ParagraphStyleOptions{
				Alignment:     alignment,
				IndentStart:   indentStart,
				IndentEnd:     indentEnd,
				IndentFirst:   indentFirst,
				LineSpacing:   lineSpacing,
				SpacingBefore: spacingBefore,
				SpacingAfter:  spacingAfter,
			})
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(result, "write")
		},
	}

	cmd.Flags().Int64Var(&startIndex, "start", 0, "Start index (required)")
	cmd.Flags().Int64Var(&endIndex, "end", 0, "End index (required)")
	cmd.Flags().StringVar(&alignment, "align", "", "Alignment: left, center, right, justify")
	cmd.Flags().Float64Var(&indentStart, "indent-start", 0, "Left indent in points")
	cmd.Flags().Float64Var(&indentEnd, "indent-end", 0, "Right indent in points")
	cmd.Flags().Float64Var(&indentFirst, "indent-first", 0, "First line indent in points")
	cmd.Flags().Float64Var(&lineSpacing, "line-spacing", 0, "Line spacing (e.g., 1.5)")
	cmd.Flags().Float64Var(&spacingBefore, "spacing-before", 0, "Space before paragraph in points")
	cmd.Flags().Float64Var(&spacingAfter, "spacing-after", 0, "Space after paragraph in points")

	cmd.MarkFlagRequired("start")
	cmd.MarkFlagRequired("end")

	return cmd
}

// docsInsertTableCmd inserts a table.
func docsInsertTableCmd() *cobra.Command {
	var rows, columns int
	var headers []string
	var csvData string
	var hasHeaders bool

	cmd := &cobra.Command{
		Use:   "insert-table <doc-id>",
		Short: "Insert a table",
		Long:  "Inserts a table with optional CSV data.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := docsWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			var result *docs.UpdateResult

			// If CSV data provided, use that
			if csvData != "" {
				result, err = svc.InsertTableFromCSV(args[0], csvData, hasHeaders)
			} else {
				// Otherwise create empty table or with headers
				result, err = svc.InsertTable(args[0], docs.TableOptions{
					Rows:    rows,
					Columns: columns,
					Headers: headers,
				})
			}

			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"document_id": result.DocumentID,
				"rows":        rows,
				"columns":     columns,
			}, "write")
		},
	}

	cmd.Flags().IntVar(&rows, "rows", 3, "Number of rows")
	cmd.Flags().IntVar(&columns, "cols", 3, "Number of columns")
	cmd.Flags().StringSliceVar(&headers, "headers", []string{}, "Header row (optional)")
	cmd.Flags().StringVar(&csvData, "csv", "", "CSV data")
	cmd.Flags().BoolVar(&hasHeaders, "has-headers", false, "CSV has header row")

	return cmd
}

// docsInsertPageBreakCmd inserts a page break.
func docsInsertPageBreakCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "insert-pagebreak <doc-id>",
		Short: "Insert a page break",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := docsWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			result, err := svc.InsertPageBreak(args[0])
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(result, "write")
		},
	}
}

// docsInsertHRCmd inserts a horizontal rule.
func docsInsertHRCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "insert-hr <doc-id>",
		Short: "Insert a horizontal rule",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := docsWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			result, err := svc.InsertHorizontalRule(args[0])
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(result, "write")
		},
	}
}

// docsInsertTOCCmd inserts a table of contents.
func docsInsertTOCCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "insert-toc <doc-id>",
		Short: "Insert a table of contents placeholder",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := docsWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			result, err := svc.InsertTOC(args[0])
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(result, "write")
		},
	}
}

// docsFormatFromTemplateCmd applies a template to a document.
func docsFormatFromTemplateCmd() *cobra.Command {
	var templateJSON string
	var templateFile string

	cmd := &cobra.Command{
		Use:   "format-template <doc-id>",
		Short: "Apply formatting template",
		Long:  "Applies a JSON template with multiple formatting operations.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := docsWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			// Read template from file if specified
			if templateFile != "" {
				data, err := os.ReadFile(templateFile)
				if err != nil {
					output.FailureFromError(output.ErrInvalidInput, err)
					return
				}
				templateJSON = string(data)
			}

			result, err := svc.FormatFromTemplate(args[0], templateJSON)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"document_id": result.DocumentID,
				"applied":     true,
			}, "write")
		},
	}

	cmd.Flags().StringVar(&templateJSON, "template", "", "JSON template string")
	cmd.Flags().StringVar(&templateFile, "template-file", "", "Path to JSON template file")

	return cmd
}

// docsFromMarkdownCmd creates or appends to a document from markdown.
// This is the PREFERRED way to build well-formatted Google Docs.
func docsFromMarkdownCmd() *cobra.Command {
	var filePath string
	var text string
	var title string
	var replace bool
	var preview bool

	cmd := &cobra.Command{
		Use:   "from-markdown [doc-id]",
		Short: "Create or update a Google Doc from markdown (PREFERRED for building docs)",
		Long: `Converts markdown to a clean, well-formatted Google Doc in a single atomic operation.
This is the best way to build documents — write your content as markdown and
let this command handle all the Google Docs formatting (headings, bold, italic,
lists, code blocks, links, horizontal rules).

Creates a new document when --title is provided (doc-id is optional).
Appends to an existing document when doc-id is provided without --title.
Use --replace to clear the document and rewrite it (for iterating on content).

Supports: headings (h1-h6), bold, italic, bullet lists, numbered lists,
code blocks (monospace), inline code, links, and horizontal rules.`,
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var markdown string

			switch {
			case text != "":
				markdown = text
			case filePath != "":
				data, err := os.ReadFile(filePath)
				if err != nil {
					output.FailureFromError(output.ErrInvalidInput, err)
					return
				}
				markdown = string(data)
			default:
				output.Failure(output.ErrInvalidInput, "provide markdown via --text or --file", nil)
				return
			}

			if preview {
				mc := docs.NewMarkdownConverter(1)
				reqs, err := mc.Convert(markdown)
				if err != nil {
					output.APIError(err)
					return
				}
				output.Success(map[string]interface{}{
					"preview":       true,
					"request_count": len(reqs),
				}, "read")
				return
			}

			ctx := context.Background()
			svc, err := docsWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			// Determine document ID: create new or use existing
			var documentID string
			if title != "" {
				// Create a new document
				createResult, err := svc.Create(title)
				if err != nil {
					output.APIError(err)
					return
				}
				documentID = createResult.DocumentID
			} else if len(args) > 0 {
				documentID = args[0]
			} else {
				output.Failure(output.ErrInvalidInput, "provide a doc-id or use --title to create a new document", nil)
				return
			}

			var result *docs.UpdateResult
			if replace {
				result, err = svc.ReplaceFromMarkdown(documentID, markdown)
			} else {
				result, err = svc.FromMarkdown(documentID, markdown)
			}
			if err != nil {
				output.APIError(err)
				return
			}

			resp := map[string]interface{}{
				"document_id": result.DocumentID,
				"applied":     true,
			}
			if title != "" {
				resp["created"] = true
				resp["title"] = title
			}

			output.Success(resp, "write")
		},
	}

	cmd.Flags().StringVar(&text, "text", "", "Markdown content to convert")
	cmd.Flags().StringVar(&filePath, "file", "", "Path to a markdown file")
	cmd.Flags().StringVar(&title, "title", "", "Create a new document with this title")
	cmd.Flags().BoolVar(&replace, "replace", false, "Clear existing content and replace with markdown")
	cmd.Flags().BoolVar(&preview, "preview", false, "Preview without applying")

	return cmd
}

// docsStructureCmd returns a detailed structural analysis of a document.
func docsStructureCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "structure <doc-id>",
		Short: "Analyze document structure",
		Long:  "Returns structural analysis: headings, tables, lists, word count.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := docsReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			structure, err := svc.Structure(args[0])
			if err != nil {
				output.NotFoundError("Document", args[0])
				return
			}

			output.Success(structure, "read")
		},
	}
}

// parseListItems parses comma-separated list items.
func parseListItems(itemsStr string) []string {
	var items []string
	for _, item := range strings.Split(itemsStr, ",") {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			items = append(items, trimmed)
		}
	}
	return items
}
