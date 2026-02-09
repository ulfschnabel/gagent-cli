package docs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestValidateListType tests list type validation.
func TestValidateListType(t *testing.T) {
	tests := []struct {
		name     string
		listType string
		wantErr  bool
	}{
		{"valid bullet", "bullet", false},
		{"valid numbered", "numbered", false},
		{"valid lettered", "lettered", false},
		{"valid roman", "roman", false},
		{"valid checklist", "checklist", false},
		{"empty type", "", true},
		{"invalid type", "invalid", true},
		{"uppercase", "BULLET", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateListType(tt.listType)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateListType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidateListItems tests list items validation.
func TestValidateListItems(t *testing.T) {
	tests := []struct {
		name    string
		items   []string
		wantErr bool
	}{
		{"valid items", []string{"Item 1", "Item 2"}, false},
		{"single item", []string{"Item 1"}, false},
		{"empty list", []string{}, true},
		{"nil list", nil, true},
		{"empty string item", []string{""}, true},
		{"whitespace only", []string{"  "}, true},
		{"mixed valid and empty", []string{"Item 1", ""}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateListItems(tt.items)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateListItems() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidateIndentLevel tests indent level validation.
func TestValidateIndentLevel(t *testing.T) {
	tests := []struct {
		name    string
		indent  int
		wantErr bool
	}{
		{"level 0", 0, false},
		{"level 1", 1, false},
		{"level 5", 5, false},
		{"level 9", 9, false},
		{"negative", -1, true},
		{"too large", 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateIndentLevel(tt.indent)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateIndentLevel() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestMapListGlyphType tests list glyph type mapping.
func TestMapListGlyphType(t *testing.T) {
	tests := []struct {
		name     string
		listType string
		want     string
	}{
		{"bullet", "bullet", "BULLET_DISC_CIRCLE_SQUARE"},
		{"numbered", "numbered", "DECIMAL"},
		{"lettered", "lettered", "ALPHA"},
		{"roman", "roman", "ROMAN"},
		{"checklist", "checklist", "CHECKBOX"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapListGlyphType(tt.listType)
			if got != tt.want {
				t.Errorf("mapListGlyphType(%q) = %q, want %q", tt.listType, got, tt.want)
			}
		})
	}
}

// TestParseTextStyleFlags tests parsing text style flags.
func TestParseTextStyleFlags(t *testing.T) {
	tests := []struct {
		name           string
		bold           bool
		italic         bool
		underline      bool
		strikethrough  bool
		fontSize       int
		color          string
		wantBold       bool
		wantItalic     bool
		wantUnderline  bool
		wantStrike     bool
		wantFontSize   bool
		wantColor      bool
	}{
		{
			name:          "no formatting",
			bold:          false,
			italic:        false,
			underline:     false,
			strikethrough: false,
			fontSize:      0,
			color:         "",
			wantBold:      false,
			wantItalic:    false,
			wantUnderline: false,
			wantStrike:    false,
			wantFontSize:  false,
			wantColor:     false,
		},
		{
			name:          "bold only",
			bold:          true,
			italic:        false,
			underline:     false,
			strikethrough: false,
			fontSize:      0,
			color:         "",
			wantBold:      true,
			wantItalic:    false,
			wantUnderline: false,
			wantStrike:    false,
			wantFontSize:  false,
			wantColor:     false,
		},
		{
			name:          "all formatting",
			bold:          true,
			italic:        true,
			underline:     true,
			strikethrough: true,
			fontSize:      14,
			color:         "#ff0000",
			wantBold:      true,
			wantItalic:    true,
			wantUnderline: true,
			wantStrike:    true,
			wantFontSize:  true,
			wantColor:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := parseTextStyleFlags(tt.bold, tt.italic, tt.underline, tt.strikethrough, tt.fontSize, tt.color)

			assert.Equal(t, tt.wantBold, style.Bold)
			assert.Equal(t, tt.wantItalic, style.Italic)
			assert.Equal(t, tt.wantUnderline, style.Underline)
			assert.Equal(t, tt.wantStrike, style.Strikethrough)

			if tt.wantFontSize {
				assert.NotNil(t, style.FontSize)
			} else {
				assert.Nil(t, style.FontSize)
			}

			if tt.wantColor {
				assert.NotNil(t, style.ForegroundColor)
			} else {
				assert.Nil(t, style.ForegroundColor)
			}
		})
	}
}

// TestValidateColor tests color validation.
func TestValidateColor(t *testing.T) {
	tests := []struct {
		name    string
		color   string
		wantErr bool
	}{
		{"valid hex 6 digits", "#ff0000", false},
		{"valid hex uppercase", "#FF0000", false},
		{"valid hex mixed case", "#Ff0000", false},
		{"empty", "", false}, // Empty is valid (no color)
		{"invalid no hash", "ff0000", true},
		{"invalid short", "#fff", true},
		{"invalid too long", "#ff00000", true},
		{"invalid chars", "#gggggg", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateColor(tt.color)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateColor() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidateFontSize tests font size validation.
func TestValidateFontSize(t *testing.T) {
	tests := []struct {
		name     string
		fontSize int
		wantErr  bool
	}{
		{"valid 12", 12, false},
		{"valid 8", 8, false},
		{"valid 72", 72, false},
		{"zero", 0, false}, // 0 means no size specified
		{"too small", 7, true},
		{"too large", 73, true},
		{"negative", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFontSize(tt.fontSize)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateFontSize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestMapNamedStyle tests named style mapping.
func TestMapNamedStyle(t *testing.T) {
	tests := []struct {
		name      string
		styleName string
		want      string
	}{
		{"heading1", "heading1", "HEADING_1"},
		{"heading2", "heading2", "HEADING_2"},
		{"heading3", "heading3", "HEADING_3"},
		{"heading4", "heading4", "HEADING_4"},
		{"title", "title", "TITLE"},
		{"subtitle", "subtitle", "SUBTITLE"},
		{"normal", "normal", "NORMAL_TEXT"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapNamedStyle(tt.styleName)
			if got != tt.want {
				t.Errorf("mapNamedStyle(%q) = %q, want %q", tt.styleName, got, tt.want)
			}
		})
	}
}

// TestValidateAlignment tests alignment validation.
func TestValidateAlignment(t *testing.T) {
	tests := []struct {
		name      string
		alignment string
		wantErr   bool
	}{
		{"left", "left", false},
		{"center", "center", false},
		{"right", "right", false},
		{"justify", "justify", false},
		{"empty", "", false}, // Empty is valid (no alignment)
		{"invalid", "invalid", true},
		{"uppercase", "LEFT", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAlignment(tt.alignment)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAlignment() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
