package docs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gdocs "google.golang.org/api/docs/v1"
)

func TestAnalyzeStructure_EmptyDoc(t *testing.T) {
	doc := &gdocs.Document{
		DocumentId: "test-id",
		Title:      "Empty Doc",
		Body:       &gdocs.Body{Content: []*gdocs.StructuralElement{}},
	}

	s := analyzeStructure(doc)
	assert.Equal(t, "test-id", s.ID)
	assert.Equal(t, "Empty Doc", s.Title)
	assert.Equal(t, 0, s.WordCount)
	assert.Empty(t, s.Headings)
	assert.Empty(t, s.Tables)
	assert.Empty(t, s.Lists)
}

func TestAnalyzeStructure_HeadingsOnly(t *testing.T) {
	doc := &gdocs.Document{
		DocumentId: "test-id",
		Title:      "Headings Doc",
		Body: &gdocs.Body{
			Content: []*gdocs.StructuralElement{
				{
					StartIndex: 0,
					EndIndex:   20,
					Paragraph: &gdocs.Paragraph{
						ParagraphStyle: &gdocs.ParagraphStyle{
							NamedStyleType: "HEADING_1",
						},
						Elements: []*gdocs.ParagraphElement{
							{TextRun: &gdocs.TextRun{Content: "Chapter One\n"}},
						},
					},
				},
				{
					StartIndex: 20,
					EndIndex:   40,
					Paragraph: &gdocs.Paragraph{
						ParagraphStyle: &gdocs.ParagraphStyle{
							NamedStyleType: "HEADING_2",
						},
						Elements: []*gdocs.ParagraphElement{
							{TextRun: &gdocs.TextRun{Content: "Section A\n"}},
						},
					},
				},
			},
		},
	}

	s := analyzeStructure(doc)
	require.Len(t, s.Headings, 2)
	assert.Equal(t, "Chapter One", s.Headings[0].Text)
	assert.Equal(t, 1, s.Headings[0].Level)
	assert.Equal(t, "Section A", s.Headings[1].Text)
	assert.Equal(t, 2, s.Headings[1].Level)
}

func TestAnalyzeStructure_WithTable(t *testing.T) {
	doc := &gdocs.Document{
		DocumentId: "test-id",
		Title:      "Table Doc",
		Body: &gdocs.Body{
			Content: []*gdocs.StructuralElement{
				{
					StartIndex: 0,
					EndIndex:   100,
					Table: &gdocs.Table{
						Rows:    3,
						Columns: 2,
						TableRows: []*gdocs.TableRow{
							{
								TableCells: []*gdocs.TableCell{
									{Content: []*gdocs.StructuralElement{
										{Paragraph: &gdocs.Paragraph{
											Elements: []*gdocs.ParagraphElement{
												{TextRun: &gdocs.TextRun{Content: "Name"}},
											},
										}},
									}},
									{Content: []*gdocs.StructuralElement{
										{Paragraph: &gdocs.Paragraph{
											Elements: []*gdocs.ParagraphElement{
												{TextRun: &gdocs.TextRun{Content: "Value"}},
											},
										}},
									}},
								},
							},
						},
					},
				},
			},
		},
	}

	s := analyzeStructure(doc)
	require.Len(t, s.Tables, 1)
	assert.Equal(t, 3, s.Tables[0].Rows)
	assert.Equal(t, 2, s.Tables[0].Columns)
	assert.Equal(t, int64(0), s.Tables[0].StartIndex)
	assert.Equal(t, []string{"Name", "Value"}, s.Tables[0].FirstRow)
}

func TestAnalyzeStructure_WithList(t *testing.T) {
	doc := &gdocs.Document{
		DocumentId: "test-id",
		Title:      "List Doc",
		Lists: map[string]gdocs.List{
			"list-1": {ListProperties: &gdocs.ListProperties{
				NestingLevels: []*gdocs.NestingLevel{
					{GlyphType: "DECIMAL"},
				},
			}},
		},
		Body: &gdocs.Body{
			Content: []*gdocs.StructuralElement{
				{
					Paragraph: &gdocs.Paragraph{
						Bullet: &gdocs.Bullet{ListId: "list-1"},
						Elements: []*gdocs.ParagraphElement{
							{TextRun: &gdocs.TextRun{Content: "Item 1\n"}},
						},
					},
				},
				{
					Paragraph: &gdocs.Paragraph{
						Bullet: &gdocs.Bullet{ListId: "list-1"},
						Elements: []*gdocs.ParagraphElement{
							{TextRun: &gdocs.TextRun{Content: "Item 2\n"}},
						},
					},
				},
			},
		},
	}

	s := analyzeStructure(doc)
	require.Len(t, s.Lists, 1)
	assert.Equal(t, 2, s.Lists[0].ItemCount)
	assert.Equal(t, "list-1", s.Lists[0].ListID)
}

func TestAnalyzeStructure_WordCount(t *testing.T) {
	doc := &gdocs.Document{
		DocumentId: "test-id",
		Title:      "Word Count Doc",
		Body: &gdocs.Body{
			Content: []*gdocs.StructuralElement{
				{
					Paragraph: &gdocs.Paragraph{
						Elements: []*gdocs.ParagraphElement{
							{TextRun: &gdocs.TextRun{Content: "Hello world this is a test\n"}},
						},
					},
				},
				{
					Paragraph: &gdocs.Paragraph{
						Elements: []*gdocs.ParagraphElement{
							{TextRun: &gdocs.TextRun{Content: "Another line here\n"}},
						},
					},
				},
			},
		},
	}

	s := analyzeStructure(doc)
	assert.Equal(t, 9, s.WordCount)
}

func TestAnalyzeStructure_ComplexDocument(t *testing.T) {
	doc := &gdocs.Document{
		DocumentId: "complex-id",
		Title:      "Complex Doc",
		Lists: map[string]gdocs.List{
			"list-a": {ListProperties: &gdocs.ListProperties{
				NestingLevels: []*gdocs.NestingLevel{
					{GlyphType: "DECIMAL"},
				},
			}},
		},
		Body: &gdocs.Body{
			Content: []*gdocs.StructuralElement{
				{
					StartIndex: 0,
					EndIndex:   15,
					Paragraph: &gdocs.Paragraph{
						ParagraphStyle: &gdocs.ParagraphStyle{NamedStyleType: "HEADING_1"},
						Elements: []*gdocs.ParagraphElement{
							{TextRun: &gdocs.TextRun{Content: "Title\n"}},
						},
					},
				},
				{
					StartIndex: 15,
					EndIndex:   40,
					Paragraph: &gdocs.Paragraph{
						Elements: []*gdocs.ParagraphElement{
							{TextRun: &gdocs.TextRun{Content: "Some body text\n"}},
						},
					},
				},
				{
					StartIndex: 40,
					EndIndex:   60,
					Paragraph: &gdocs.Paragraph{
						Bullet: &gdocs.Bullet{ListId: "list-a"},
						Elements: []*gdocs.ParagraphElement{
							{TextRun: &gdocs.TextRun{Content: "Item\n"}},
						},
					},
				},
				{
					StartIndex: 60,
					EndIndex:   160,
					Table: &gdocs.Table{
						Rows:    2,
						Columns: 2,
						TableRows: []*gdocs.TableRow{
							{
								TableCells: []*gdocs.TableCell{
									{Content: []*gdocs.StructuralElement{
										{Paragraph: &gdocs.Paragraph{
											Elements: []*gdocs.ParagraphElement{
												{TextRun: &gdocs.TextRun{Content: "A"}},
											},
										}},
									}},
									{Content: []*gdocs.StructuralElement{
										{Paragraph: &gdocs.Paragraph{
											Elements: []*gdocs.ParagraphElement{
												{TextRun: &gdocs.TextRun{Content: "B"}},
											},
										}},
									}},
								},
							},
						},
					},
				},
			},
		},
	}

	s := analyzeStructure(doc)
	assert.Equal(t, "complex-id", s.ID)
	assert.Len(t, s.Headings, 1)
	assert.Len(t, s.Tables, 1)
	assert.Len(t, s.Lists, 1)
	assert.Greater(t, s.WordCount, 0)
}
