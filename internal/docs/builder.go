package docs

import (
	"strings"

	gdocs "google.golang.org/api/docs/v1"
)

// TextStyleDef defines text style properties for the builder.
type TextStyleDef struct {
	Bold          bool
	Italic        bool
	Underline     bool
	Strikethrough bool
	FontSize      int
	FontFamily    string
	Color         string
	BgColor       string
}

// RequestBuilder accumulates Google Docs API requests with cursor tracking.
type RequestBuilder struct {
	cursor   int64
	requests []*gdocs.Request
}

// NewRequestBuilder creates a builder starting at the given document index.
func NewRequestBuilder(startIndex int64) *RequestBuilder {
	return &RequestBuilder{
		cursor: startIndex,
	}
}

// Cursor returns the current insertion index.
func (b *RequestBuilder) Cursor() int64 {
	return b.cursor
}

// Build returns the accumulated requests.
func (b *RequestBuilder) Build() []*gdocs.Request {
	return b.requests
}

// InsertText inserts text at the current cursor position and advances the cursor.
func (b *RequestBuilder) InsertText(text string) {
	b.requests = append(b.requests, &gdocs.Request{
		InsertText: &gdocs.InsertTextRequest{
			Text: text,
			Location: &gdocs.Location{
				Index: b.cursor,
			},
		},
	})
	b.cursor += int64(len(text))
}

// ApplyNamedStyle applies a named paragraph style to the given range.
func (b *RequestBuilder) ApplyNamedStyle(startIndex, endIndex int64, style string) {
	b.requests = append(b.requests, &gdocs.Request{
		UpdateParagraphStyle: &gdocs.UpdateParagraphStyleRequest{
			Range: &gdocs.Range{
				StartIndex: startIndex,
				EndIndex:   endIndex,
			},
			ParagraphStyle: &gdocs.ParagraphStyle{
				NamedStyleType: style,
			},
			Fields: "namedStyleType",
		},
	})
}

// ApplyTextStyle applies inline text formatting to the given range.
func (b *RequestBuilder) ApplyTextStyle(startIndex, endIndex int64, style *TextStyleDef, fields string) {
	textStyle := &gdocs.TextStyle{
		Bold:          style.Bold,
		Italic:        style.Italic,
		Underline:     style.Underline,
		Strikethrough: style.Strikethrough,
	}

	if style.FontSize > 0 {
		textStyle.FontSize = &gdocs.Dimension{
			Magnitude: float64(style.FontSize),
			Unit:      "PT",
		}
	}
	if style.FontFamily != "" {
		textStyle.WeightedFontFamily = &gdocs.WeightedFontFamily{
			FontFamily: style.FontFamily,
		}
	}
	if style.Color != "" {
		textStyle.ForegroundColor = parseColor(style.Color)
	}
	if style.BgColor != "" {
		textStyle.BackgroundColor = parseColor(style.BgColor)
	}

	b.requests = append(b.requests, &gdocs.Request{
		UpdateTextStyle: &gdocs.UpdateTextStyleRequest{
			Range: &gdocs.Range{
				StartIndex: startIndex,
				EndIndex:   endIndex,
			},
			TextStyle: textStyle,
			Fields:    fields,
		},
	})
}

// InsertTable inserts a table at the current cursor position.
func (b *RequestBuilder) InsertTable(rows, cols int) {
	b.requests = append(b.requests, &gdocs.Request{
		InsertTable: &gdocs.InsertTableRequest{
			Rows:    int64(rows),
			Columns: int64(cols),
			Location: &gdocs.Location{
				Index: b.cursor,
			},
		},
	})
	// Table indices are unknown until inserted; cursor not advanced
}

// InsertPageBreak inserts a page break at the current cursor position.
func (b *RequestBuilder) InsertPageBreak() {
	b.requests = append(b.requests, &gdocs.Request{
		InsertPageBreak: &gdocs.InsertPageBreakRequest{
			Location: &gdocs.Location{
				Index: b.cursor,
			},
		},
	})
	b.cursor += 2 // page break consumes 2 indices
}

// InsertHorizontalRule inserts a text-based horizontal rule.
func (b *RequestBuilder) InsertHorizontalRule() {
	b.InsertText("\n" + strings.Repeat("_", 50) + "\n")
}

// CreateList creates a bulleted/numbered list over the given range.
func (b *RequestBuilder) CreateList(startIndex, endIndex int64, glyphType string) {
	b.requests = append(b.requests, &gdocs.Request{
		CreateParagraphBullets: &gdocs.CreateParagraphBulletsRequest{
			Range: &gdocs.Range{
				StartIndex: startIndex,
				EndIndex:   endIndex,
			},
			BulletPreset: glyphType,
		},
	})
}
