package main

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"github.com/ulfhaga/gagent-cli/internal/auth"
	"github.com/ulfhaga/gagent-cli/internal/config"
	"github.com/ulfhaga/gagent-cli/internal/output"
	"github.com/ulfhaga/gagent-cli/internal/slides"
)

// slidesReadService creates a Slides service with read scope.
func slidesReadService(ctx context.Context) (*slides.Service, error) {
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

	return slides.NewService(ctx, client)
}

// slidesWriteService creates a Slides service with write scope.
func slidesWriteService(ctx context.Context) (*slides.Service, error) {
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

	return slides.NewService(ctx, client)
}

func slidesListCmd() *cobra.Command {
	var limit int64
	var query string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List presentations",
		Long:  "Lists presentations from Drive with title, id, last modified.",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := slidesReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			presentations, _, err := svc.List(slides.ListOptions{
				Query:      query,
				MaxResults: limit,
			})
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"presentations": presentations,
				"count":         len(presentations),
			}, "read")
		},
	}

	cmd.Flags().Int64VarP(&limit, "limit", "n", 10, "Maximum number of presentations")
	cmd.Flags().StringVar(&query, "query", "", "Search query")

	return cmd
}

func slidesInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info <presentation-id>",
		Short: "Get presentation metadata",
		Long:  "Returns presentation metadata: slide count, dimensions.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := slidesReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			info, err := svc.Info(args[0])
			if err != nil {
				output.NotFoundError("Presentation", args[0])
				return
			}

			output.Success(info, "read")
		},
	}
}

func slidesReadCmd() *cobra.Command {
	var slideNum int

	cmd := &cobra.Command{
		Use:   "read <presentation-id>",
		Short: "Read slide content",
		Long:  "Returns slide content: text elements, shapes, images.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := slidesReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			if slideNum > 0 {
				content, err := svc.Read(args[0], slideNum)
				if err != nil {
					output.APIError(err)
					return
				}
				output.Success(content, "read")
				return
			}

			// Read all slides
			contents, err := svc.ReadAll(args[0])
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"presentation_id": args[0],
				"slides":          contents,
				"count":           len(contents),
			}, "read")
		},
	}

	cmd.Flags().IntVar(&slideNum, "slide", 0, "Specific slide number (1-indexed)")

	return cmd
}

func slidesExportCmd() *cobra.Command {
	var format, outputFile string

	cmd := &cobra.Command{
		Use:   "export <presentation-id>",
		Short: "Export presentation",
		Long: `Exports presentation in specified format (pdf, pptx).

TIP FOR AI AGENTS: Use this command to create a visual feedback loop when
building slides. Export to PDF, read the PDF to see the rendered output,
then fix any issues (overlaps, positioning) via 'slides api batch-update'.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := slidesReadService(ctx)
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
					"presentation_id": args[0],
					"format":          format,
					"saved_to":        outputFile,
					"size":            len(data),
				}, "read")
				return
			}

			output.Success(map[string]interface{}{
				"presentation_id": args[0],
				"format":          format,
				"size":            len(data),
				"note":            "Use --output to save to a file",
			}, "read")
		},
	}

	cmd.Flags().StringVar(&format, "format", "pdf", "Export format: pdf, pptx")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path")

	return cmd
}

func slidesTextCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "text <presentation-id>",
		Short: "Extract text",
		Long:  "Extracts all text content from presentation.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := slidesReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			text, err := svc.Text(args[0])
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"presentation_id": args[0],
				"text":            text,
			}, "read")
		},
	}
}

func slidesCreateCmd() *cobra.Command {
	var title string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create presentation",
		Long:  "Creates new blank presentation.",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := slidesWriteService(ctx)
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

	cmd.Flags().StringVar(&title, "title", "", "Presentation title (required)")
	cmd.MarkFlagRequired("title")

	return cmd
}

func slidesAddSlideCmd() *cobra.Command {
	var layout string

	cmd := &cobra.Command{
		Use:   "add-slide <presentation-id>",
		Short: "Add a slide",
		Long:  "Adds new slide with specified layout.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := slidesWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			result, err := svc.AddSlide(args[0], layout)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"presentation_id": result.PresentationID,
				"layout":          layout,
				"added":           true,
			}, "write")
		},
	}

	cmd.Flags().StringVar(&layout, "layout", "blank", "Layout: blank, title, title_body, section, etc.")

	return cmd
}

func slidesDeleteSlideCmd() *cobra.Command {
	var slideNum int

	cmd := &cobra.Command{
		Use:   "delete-slide <presentation-id>",
		Short: "Delete a slide",
		Long:  "Deletes slide at position N (1-indexed).",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := slidesWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			result, err := svc.DeleteSlide(args[0], slideNum)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"presentation_id": result.PresentationID,
				"slide_number":    slideNum,
				"deleted":         true,
			}, "write")
		},
	}

	cmd.Flags().IntVar(&slideNum, "slide", 0, "Slide number to delete (required, 1-indexed)")
	cmd.MarkFlagRequired("slide")

	return cmd
}

func slidesUpdateTextCmd() *cobra.Command {
	var slideNum int
	var find, replace string

	cmd := &cobra.Command{
		Use:   "update-text <presentation-id>",
		Short: "Find and replace text",
		Long:  "Find and replace text on specific slide.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := slidesWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			result, err := svc.UpdateText(args[0], slideNum, find, replace)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"presentation_id": result.PresentationID,
				"slide_number":    slideNum,
				"find":            find,
				"replace":         replace,
			}, "write")
		},
	}

	cmd.Flags().IntVar(&slideNum, "slide", 0, "Slide number (required, 1-indexed)")
	cmd.Flags().StringVar(&find, "find", "", "Text to find (required)")
	cmd.Flags().StringVar(&replace, "replace", "", "Replacement text (required)")

	cmd.MarkFlagRequired("slide")
	cmd.MarkFlagRequired("find")
	cmd.MarkFlagRequired("replace")

	return cmd
}

func slidesAddTextCmd() *cobra.Command {
	var slideNum int
	var text string
	var x, y, width, height float64

	cmd := &cobra.Command{
		Use:   "add-text <presentation-id>",
		Short: "Add text box",
		Long:  "Adds text box at position.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := slidesWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			result, err := svc.AddText(args[0], slideNum, text, x, y, width, height)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"presentation_id": result.PresentationID,
				"slide_number":    slideNum,
				"added":           true,
			}, "write")
		},
	}

	cmd.Flags().IntVar(&slideNum, "slide", 0, "Slide number (required, 1-indexed)")
	cmd.Flags().StringVar(&text, "text", "", "Text content (required)")
	cmd.Flags().Float64Var(&x, "x", 100, "X position in points")
	cmd.Flags().Float64Var(&y, "y", 100, "Y position in points")
	cmd.Flags().Float64Var(&width, "width", 200, "Width in points")
	cmd.Flags().Float64Var(&height, "height", 50, "Height in points")

	cmd.MarkFlagRequired("slide")
	cmd.MarkFlagRequired("text")

	return cmd
}

func slidesAddImageCmd() *cobra.Command {
	var slideNum int
	var url string
	var x, y, width, height float64

	cmd := &cobra.Command{
		Use:   "add-image <presentation-id>",
		Short: "Add image",
		Long:  "Adds image from URL at position.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := slidesWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			result, err := svc.AddImage(args[0], slideNum, url, x, y, width, height)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"presentation_id": result.PresentationID,
				"slide_number":    slideNum,
				"added":           true,
			}, "write")
		},
	}

	cmd.Flags().IntVar(&slideNum, "slide", 0, "Slide number (required, 1-indexed)")
	cmd.Flags().StringVar(&url, "url", "", "Image URL (required)")
	cmd.Flags().Float64Var(&x, "x", 100, "X position in points")
	cmd.Flags().Float64Var(&y, "y", 100, "Y position in points")
	cmd.Flags().Float64Var(&width, "width", 300, "Width in points")
	cmd.Flags().Float64Var(&height, "height", 200, "Height in points")

	cmd.MarkFlagRequired("slide")
	cmd.MarkFlagRequired("url")

	return cmd
}

// Slides API commands
func slidesAPICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api",
		Short: "Low-level Slides API commands",
		Long:  "Direct access to Slides API operations.",
	}

	cmd.AddCommand(slidesAPIGetCmd())
	cmd.AddCommand(slidesAPIBatchUpdateCmd())
	cmd.AddCommand(slidesAPICreateCmd())

	return cmd
}

func slidesAPIGetCmd() *cobra.Command {
	var pageID string

	cmd := &cobra.Command{
		Use:   "get <presentation-id>",
		Short: "Get presentation structure",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := slidesReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			if pageID != "" {
				page, err := svc.GetPage(args[0], pageID)
				if err != nil {
					output.NotFoundError("Page", pageID)
					return
				}
				output.Success(page, "read")
				return
			}

			pres, err := svc.Get(args[0])
			if err != nil {
				output.NotFoundError("Presentation", args[0])
				return
			}

			output.Success(pres, "read")
		},
	}

	cmd.Flags().StringVar(&pageID, "page-id", "", "Specific page/slide ID")

	return cmd
}

func slidesAPIBatchUpdateCmd() *cobra.Command {
	var requestsJSON string

	cmd := &cobra.Command{
		Use:   "batch-update <presentation-id>",
		Short: "Batch update presentation",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := slidesWriteService(ctx)
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

func slidesAPICreateCmd() *cobra.Command {
	var title string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create empty presentation",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := slidesWriteService(ctx)
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

	cmd.Flags().StringVar(&title, "title", "", "Presentation title (required)")
	cmd.MarkFlagRequired("title")

	return cmd
}
