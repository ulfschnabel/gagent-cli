package docs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarkdownConvert_EmptyInput(t *testing.T) {
	mc := NewMarkdownConverter(1)
	reqs, err := mc.Convert("")
	require.NoError(t, err)
	assert.Empty(t, reqs)
}

func TestMarkdownConvert_Heading1(t *testing.T) {
	mc := NewMarkdownConverter(1)
	reqs, err := mc.Convert("# Hello\n")
	require.NoError(t, err)
	require.NotEmpty(t, reqs)

	// Should have InsertText + ApplyNamedStyle(HEADING_1)
	hasInsert := false
	hasHeading := false
	for _, r := range reqs {
		if r.InsertText != nil && r.InsertText.Text == "Hello\n" {
			hasInsert = true
		}
		if r.UpdateParagraphStyle != nil &&
			r.UpdateParagraphStyle.ParagraphStyle.NamedStyleType == "HEADING_1" {
			hasHeading = true
		}
	}
	assert.True(t, hasInsert, "should insert heading text")
	assert.True(t, hasHeading, "should apply HEADING_1 style")
}

func TestMarkdownConvert_HeadingLevels(t *testing.T) {
	tests := []struct {
		md    string
		style string
	}{
		{"# H1\n", "HEADING_1"},
		{"## H2\n", "HEADING_2"},
		{"### H3\n", "HEADING_3"},
		{"#### H4\n", "HEADING_4"},
		{"##### H5\n", "HEADING_5"},
		{"###### H6\n", "HEADING_6"},
	}

	for _, tt := range tests {
		t.Run(tt.style, func(t *testing.T) {
			mc := NewMarkdownConverter(1)
			reqs, err := mc.Convert(tt.md)
			require.NoError(t, err)

			found := false
			for _, r := range reqs {
				if r.UpdateParagraphStyle != nil &&
					r.UpdateParagraphStyle.ParagraphStyle.NamedStyleType == tt.style {
					found = true
				}
			}
			assert.True(t, found, "should have %s style", tt.style)
		})
	}
}

func TestMarkdownConvert_Paragraph(t *testing.T) {
	mc := NewMarkdownConverter(1)
	reqs, err := mc.Convert("Hello world\n")
	require.NoError(t, err)

	hasInsert := false
	for _, r := range reqs {
		if r.InsertText != nil && r.InsertText.Text == "Hello world\n" {
			hasInsert = true
		}
	}
	assert.True(t, hasInsert, "should insert paragraph text")
}

func TestMarkdownConvert_Bold(t *testing.T) {
	mc := NewMarkdownConverter(1)
	reqs, err := mc.Convert("**bold text**\n")
	require.NoError(t, err)

	hasBold := false
	for _, r := range reqs {
		if r.UpdateTextStyle != nil && r.UpdateTextStyle.TextStyle.Bold {
			hasBold = true
		}
	}
	assert.True(t, hasBold, "should have bold style")
}

func TestMarkdownConvert_Italic(t *testing.T) {
	mc := NewMarkdownConverter(1)
	reqs, err := mc.Convert("*italic text*\n")
	require.NoError(t, err)

	hasItalic := false
	for _, r := range reqs {
		if r.UpdateTextStyle != nil && r.UpdateTextStyle.TextStyle.Italic {
			hasItalic = true
		}
	}
	assert.True(t, hasItalic, "should have italic style")
}

func TestMarkdownConvert_BulletList(t *testing.T) {
	mc := NewMarkdownConverter(1)
	reqs, err := mc.Convert("- Item 1\n- Item 2\n")
	require.NoError(t, err)

	hasBullets := false
	for _, r := range reqs {
		if r.CreateParagraphBullets != nil {
			hasBullets = true
		}
	}
	assert.True(t, hasBullets, "should create bullet list")
}

func TestMarkdownConvert_NumberedList(t *testing.T) {
	mc := NewMarkdownConverter(1)
	reqs, err := mc.Convert("1. First\n2. Second\n")
	require.NoError(t, err)

	hasNumbered := false
	for _, r := range reqs {
		if r.CreateParagraphBullets != nil &&
			r.CreateParagraphBullets.BulletPreset == "NUMBERED_DECIMAL_ALPHA_ROMAN" {
			hasNumbered = true
		}
	}
	assert.True(t, hasNumbered, "should create numbered list")
}

func TestMarkdownConvert_CodeBlock(t *testing.T) {
	mc := NewMarkdownConverter(1)
	reqs, err := mc.Convert("```\nfmt.Println(\"hello\")\n```\n")
	require.NoError(t, err)

	hasMonospace := false
	for _, r := range reqs {
		if r.UpdateTextStyle != nil && r.UpdateTextStyle.TextStyle.WeightedFontFamily != nil &&
			r.UpdateTextStyle.TextStyle.WeightedFontFamily.FontFamily == "Courier New" {
			hasMonospace = true
		}
	}
	assert.True(t, hasMonospace, "code block should have monospace font")
}

func TestMarkdownConvert_InlineCode(t *testing.T) {
	mc := NewMarkdownConverter(1)
	reqs, err := mc.Convert("Use `fmt.Println` to print\n")
	require.NoError(t, err)

	hasMonospace := false
	for _, r := range reqs {
		if r.UpdateTextStyle != nil && r.UpdateTextStyle.TextStyle.WeightedFontFamily != nil &&
			r.UpdateTextStyle.TextStyle.WeightedFontFamily.FontFamily == "Courier New" {
			hasMonospace = true
		}
	}
	assert.True(t, hasMonospace, "inline code should have monospace font")
}

func TestMarkdownConvert_Link(t *testing.T) {
	mc := NewMarkdownConverter(1)
	reqs, err := mc.Convert("[Google](https://google.com)\n")
	require.NoError(t, err)

	hasLink := false
	for _, r := range reqs {
		if r.UpdateTextStyle != nil && r.UpdateTextStyle.TextStyle.Link != nil {
			hasLink = true
			assert.Equal(t, "https://google.com", r.UpdateTextStyle.TextStyle.Link.Url)
		}
	}
	assert.True(t, hasLink, "should have link style")
}

func TestMarkdownConvert_ThematicBreak(t *testing.T) {
	mc := NewMarkdownConverter(1)
	reqs, err := mc.Convert("---\n")
	require.NoError(t, err)

	hasHR := false
	for _, r := range reqs {
		if r.InsertText != nil && len(r.InsertText.Text) > 10 {
			hasHR = true
		}
	}
	assert.True(t, hasHR, "should insert horizontal rule text")
}

func TestMarkdownConvert_ComplexDocument(t *testing.T) {
	md := `# Report Title

## Introduction

This is the **introduction** with *emphasis*.

- Point one
- Point two

### Code Example

` + "```" + `
func main() {}
` + "```" + `

---

[Link](https://example.com)
`

	mc := NewMarkdownConverter(1)
	reqs, err := mc.Convert(md)
	require.NoError(t, err)
	assert.NotEmpty(t, reqs)

	// Verify we have a mix of request types
	hasInsert := false
	hasStyle := false
	hasBullets := false
	for _, r := range reqs {
		if r.InsertText != nil {
			hasInsert = true
		}
		if r.UpdateParagraphStyle != nil {
			hasStyle = true
		}
		if r.CreateParagraphBullets != nil {
			hasBullets = true
		}
	}
	assert.True(t, hasInsert)
	assert.True(t, hasStyle)
	assert.True(t, hasBullets)
}
