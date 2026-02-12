package docs

import (
	"fmt"
	"strings"

	gdocs "google.golang.org/api/docs/v1"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// MarkdownConverter converts Markdown to Google Docs API requests.
type MarkdownConverter struct {
	builder *RequestBuilder
	source  []byte
}

// NewMarkdownConverter creates a converter starting at the given document index.
func NewMarkdownConverter(startIndex int64) *MarkdownConverter {
	return &MarkdownConverter{
		builder: NewRequestBuilder(startIndex),
	}
}

// Convert parses markdown and returns Google Docs API requests.
func (mc *MarkdownConverter) Convert(markdown string) ([]*gdocs.Request, error) {
	if strings.TrimSpace(markdown) == "" {
		return nil, nil
	}

	mc.source = []byte(markdown)

	md := goldmark.New()
	reader := text.NewReader(mc.source)
	doc := md.Parser().Parse(reader)

	if err := mc.walkNode(doc); err != nil {
		return nil, fmt.Errorf("failed to convert markdown: %w", err)
	}

	return mc.builder.Build(), nil
}

func (mc *MarkdownConverter) walkNode(n ast.Node) error {
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		if err := mc.processNode(child); err != nil {
			return err
		}
	}
	return nil
}

func (mc *MarkdownConverter) processNode(n ast.Node) error {
	switch node := n.(type) {
	case *ast.Heading:
		return mc.processHeading(node)
	case *ast.Paragraph:
		return mc.processParagraph(node)
	case *ast.List:
		return mc.processList(node)
	case *ast.FencedCodeBlock:
		return mc.processCodeBlock(node)
	case *ast.CodeBlock:
		return mc.processCodeBlock2(node)
	case *ast.ThematicBreak:
		mc.builder.InsertHorizontalRule()
		return nil
	default:
		// Walk children for unknown block types
		return mc.walkNode(n)
	}
}

func (mc *MarkdownConverter) processHeading(node *ast.Heading) error {
	start := mc.builder.Cursor()
	text := mc.extractInlineText(node)
	mc.builder.InsertText(text + "\n")
	end := mc.builder.Cursor()

	style := fmt.Sprintf("HEADING_%d", node.Level)
	mc.builder.ApplyNamedStyle(start, end, style)

	// Apply inline formatting (bold, italic, etc.)
	mc.applyInlineStyles(node, start)

	return nil
}

func (mc *MarkdownConverter) processParagraph(node *ast.Paragraph) error {
	start := mc.builder.Cursor()

	// Check if this paragraph contains only inline content
	text := mc.extractInlineText(node)
	if text == "" {
		return nil
	}

	mc.builder.InsertText(text + "\n")

	// Apply inline formatting
	mc.applyInlineStyles(node, start)

	return nil
}

func (mc *MarkdownConverter) processList(node *ast.List) error {
	start := mc.builder.Cursor()

	// Collect all list items
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if li, ok := child.(*ast.ListItem); ok {
			text := mc.extractListItemText(li)
			mc.builder.InsertText(text + "\n")
		}
	}

	end := mc.builder.Cursor()

	// Apply bullet or numbered style
	if node.IsOrdered() {
		mc.builder.CreateList(start, end, "NUMBERED_DECIMAL_ALPHA_ROMAN")
	} else {
		mc.builder.CreateList(start, end, "BULLET_DISC_CIRCLE_SQUARE")
	}

	return nil
}

func (mc *MarkdownConverter) processCodeBlock(node *ast.FencedCodeBlock) error {
	start := mc.builder.Cursor()

	// Extract code text from lines
	var code strings.Builder
	for i := 0; i < node.Lines().Len(); i++ {
		line := node.Lines().At(i)
		code.Write(line.Value(mc.source))
	}

	text := code.String()
	if text == "" {
		return nil
	}

	mc.builder.InsertText(text)
	end := mc.builder.Cursor()

	// Apply monospace font
	mc.builder.ApplyTextStyle(start, end, &TextStyleDef{
		FontFamily: "Courier New",
	}, "weightedFontFamily")

	// Add trailing newline after code block
	mc.builder.InsertText("\n")

	return nil
}

func (mc *MarkdownConverter) processCodeBlock2(node *ast.CodeBlock) error {
	start := mc.builder.Cursor()

	var code strings.Builder
	for i := 0; i < node.Lines().Len(); i++ {
		line := node.Lines().At(i)
		code.Write(line.Value(mc.source))
	}

	text := code.String()
	if text == "" {
		return nil
	}

	mc.builder.InsertText(text)
	end := mc.builder.Cursor()

	mc.builder.ApplyTextStyle(start, end, &TextStyleDef{
		FontFamily: "Courier New",
	}, "weightedFontFamily")

	mc.builder.InsertText("\n")

	return nil
}

// extractInlineText extracts plain text from inline nodes.
func (mc *MarkdownConverter) extractInlineText(n ast.Node) string {
	var text strings.Builder
	mc.walkInlineText(n, &text)
	return text.String()
}

func (mc *MarkdownConverter) walkInlineText(n ast.Node, text *strings.Builder) {
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		switch node := child.(type) {
		case *ast.Text:
			text.Write(node.Segment.Value(mc.source))
			if node.SoftLineBreak() {
				text.WriteString(" ")
			}
		case *ast.CodeSpan:
			mc.walkInlineText(node, text)
		case *ast.Emphasis:
			mc.walkInlineText(node, text)
		case *ast.Link:
			mc.walkInlineText(node, text)
		default:
			mc.walkInlineText(child, text)
		}
	}
}

// extractListItemText extracts text from a list item.
func (mc *MarkdownConverter) extractListItemText(li *ast.ListItem) string {
	var text strings.Builder
	for child := li.FirstChild(); child != nil; child = child.NextSibling() {
		if p, ok := child.(*ast.Paragraph); ok {
			text.WriteString(mc.extractInlineText(p))
		}
	}
	return text.String()
}

// applyInlineStyles walks inline nodes and applies bold/italic/code/link styles.
func (mc *MarkdownConverter) applyInlineStyles(n ast.Node, blockStart int64) {
	// Track character offset within the block text
	offset := int64(0)
	mc.walkInlineStyles(n, blockStart, &offset)
}

func (mc *MarkdownConverter) walkInlineStyles(n ast.Node, blockStart int64, offset *int64) {
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		switch node := child.(type) {
		case *ast.Text:
			segLen := int64(len(node.Segment.Value(mc.source)))
			*offset += segLen
			if node.SoftLineBreak() {
				*offset += 1 // space
			}

		case *ast.Emphasis:
			emphStart := blockStart + *offset
			mc.walkInlineStyles(node, blockStart, offset)
			emphEnd := blockStart + *offset

			if emphEnd > emphStart {
				if node.Level == 2 {
					mc.builder.ApplyTextStyle(emphStart, emphEnd, &TextStyleDef{Bold: true}, "bold")
				} else {
					mc.builder.ApplyTextStyle(emphStart, emphEnd, &TextStyleDef{Italic: true}, "italic")
				}
			}

		case *ast.CodeSpan:
			codeStart := blockStart + *offset
			mc.walkInlineStyles(node, blockStart, offset)
			codeEnd := blockStart + *offset

			if codeEnd > codeStart {
				mc.builder.ApplyTextStyle(codeStart, codeEnd, &TextStyleDef{
					FontFamily: "Courier New",
				}, "weightedFontFamily")
			}

		case *ast.Link:
			linkStart := blockStart + *offset
			mc.walkInlineStyles(node, blockStart, offset)
			linkEnd := blockStart + *offset

			if linkEnd > linkStart {
				url := string(node.Destination)
				mc.builder.requests = append(mc.builder.requests, &gdocs.Request{
					UpdateTextStyle: &gdocs.UpdateTextStyleRequest{
						Range: &gdocs.Range{
							StartIndex: linkStart,
							EndIndex:   linkEnd,
						},
						TextStyle: &gdocs.TextStyle{
							Link: &gdocs.Link{Url: url},
						},
						Fields: "link",
					},
				})
			}

		default:
			mc.walkInlineStyles(child, blockStart, offset)
		}
	}
}
