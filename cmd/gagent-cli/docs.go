package main

import (
	"context"
	"os"

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
