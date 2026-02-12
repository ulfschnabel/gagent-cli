package docs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestBuilder_EmptyBuild(t *testing.T) {
	b := NewRequestBuilder(1)
	reqs := b.Build()
	assert.Empty(t, reqs)
	assert.Equal(t, int64(1), b.Cursor())
}

func TestRequestBuilder_SingleInsertText(t *testing.T) {
	b := NewRequestBuilder(1)
	b.InsertText("Hello")

	reqs := b.Build()
	require.Len(t, reqs, 1)
	assert.NotNil(t, reqs[0].InsertText)
	assert.Equal(t, "Hello", reqs[0].InsertText.Text)
	assert.Equal(t, int64(1), reqs[0].InsertText.Location.Index)
	assert.Equal(t, int64(6), b.Cursor()) // 1 + len("Hello")
}

func TestRequestBuilder_MultipleInserts_CursorAdvances(t *testing.T) {
	b := NewRequestBuilder(1)
	b.InsertText("Hello")   // cursor: 1 -> 6
	b.InsertText(" World")  // cursor: 6 -> 12

	reqs := b.Build()
	require.Len(t, reqs, 2)
	assert.Equal(t, int64(1), reqs[0].InsertText.Location.Index)
	assert.Equal(t, int64(6), reqs[1].InsertText.Location.Index)
	assert.Equal(t, int64(12), b.Cursor())
}

func TestRequestBuilder_ApplyNamedStyle(t *testing.T) {
	b := NewRequestBuilder(1)
	b.InsertText("Title\n")
	b.ApplyNamedStyle(1, 7, "HEADING_1")

	reqs := b.Build()
	require.Len(t, reqs, 2)

	assert.NotNil(t, reqs[1].UpdateParagraphStyle)
	assert.Equal(t, "HEADING_1", reqs[1].UpdateParagraphStyle.ParagraphStyle.NamedStyleType)
	assert.Equal(t, int64(1), reqs[1].UpdateParagraphStyle.Range.StartIndex)
	assert.Equal(t, int64(7), reqs[1].UpdateParagraphStyle.Range.EndIndex)
}

func TestRequestBuilder_ApplyTextStyle(t *testing.T) {
	b := NewRequestBuilder(1)
	b.InsertText("Bold text\n")
	b.ApplyTextStyle(1, 10, &TextStyleDef{Bold: true}, "bold")

	reqs := b.Build()
	require.Len(t, reqs, 2)

	assert.NotNil(t, reqs[1].UpdateTextStyle)
	assert.True(t, reqs[1].UpdateTextStyle.TextStyle.Bold)
	assert.Equal(t, "bold", reqs[1].UpdateTextStyle.Fields)
}

func TestRequestBuilder_InsertTable(t *testing.T) {
	b := NewRequestBuilder(1)
	b.InsertTable(3, 2)

	reqs := b.Build()
	require.Len(t, reqs, 1)
	assert.NotNil(t, reqs[0].InsertTable)
	assert.Equal(t, int64(3), reqs[0].InsertTable.Rows)
	assert.Equal(t, int64(2), reqs[0].InsertTable.Columns)
	assert.Equal(t, int64(1), reqs[0].InsertTable.Location.Index)
}

func TestRequestBuilder_InsertPageBreak(t *testing.T) {
	b := NewRequestBuilder(10)
	b.InsertPageBreak()

	reqs := b.Build()
	require.Len(t, reqs, 1)
	assert.NotNil(t, reqs[0].InsertPageBreak)
	assert.Equal(t, int64(10), reqs[0].InsertPageBreak.Location.Index)
}

func TestRequestBuilder_CreateList(t *testing.T) {
	b := NewRequestBuilder(1)
	b.InsertText("Item 1\nItem 2\n")
	b.CreateList(1, 15, "BULLET_DISC_CIRCLE_SQUARE")

	reqs := b.Build()
	require.Len(t, reqs, 2)
	assert.NotNil(t, reqs[1].CreateParagraphBullets)
	assert.Equal(t, "BULLET_DISC_CIRCLE_SQUARE", reqs[1].CreateParagraphBullets.BulletPreset)
	assert.Equal(t, int64(1), reqs[1].CreateParagraphBullets.Range.StartIndex)
	assert.Equal(t, int64(15), reqs[1].CreateParagraphBullets.Range.EndIndex)
}

func TestRequestBuilder_ComplexDocument(t *testing.T) {
	b := NewRequestBuilder(1)

	// Title
	titleStart := b.Cursor()
	b.InsertText("My Report\n")
	titleEnd := b.Cursor()
	b.ApplyNamedStyle(titleStart, titleEnd, "TITLE")

	// Section heading
	headingStart := b.Cursor()
	b.InsertText("Introduction\n")
	headingEnd := b.Cursor()
	b.ApplyNamedStyle(headingStart, headingEnd, "HEADING_1")

	// Body text
	b.InsertText("This is the introduction.\n")

	// List
	listStart := b.Cursor()
	b.InsertText("Point 1\nPoint 2\n")
	listEnd := b.Cursor()
	b.CreateList(listStart, listEnd, "BULLET_DISC_CIRCLE_SQUARE")

	reqs := b.Build()
	// 4 inserts + 2 named style + 1 list = 7
	assert.Len(t, reqs, 7)

	// Verify cursor is at end
	expected := int64(1 + len("My Report\n") + len("Introduction\n") + len("This is the introduction.\n") + len("Point 1\nPoint 2\n"))
	assert.Equal(t, expected, b.Cursor())
}

func TestRequestBuilder_InsertHorizontalRule(t *testing.T) {
	b := NewRequestBuilder(1)
	b.InsertHorizontalRule()

	reqs := b.Build()
	require.Len(t, reqs, 1)
	assert.NotNil(t, reqs[0].InsertText)
	assert.Contains(t, reqs[0].InsertText.Text, "___")
}
