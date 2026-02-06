package slides

import (
	"testing"
)

// TestValidateDimensions tests dimension validation.
func TestValidateDimensions(t *testing.T) {
	tests := []struct {
		name    string
		width   float64
		height  float64
		wantErr bool
	}{
		{
			name:    "valid positive dimensions",
			width:   100.0,
			height:  50.0,
			wantErr: false,
		},
		{
			name:    "zero width",
			width:   0.0,
			height:  50.0,
			wantErr: true,
		},
		{
			name:    "zero height",
			width:   100.0,
			height:  0.0,
			wantErr: true,
		},
		{
			name:    "negative width",
			width:   -100.0,
			height:  50.0,
			wantErr: true,
		},
		{
			name:    "negative height",
			width:   100.0,
			height:  -50.0,
			wantErr: true,
		},
		{
			name:    "both zero",
			width:   0.0,
			height:  0.0,
			wantErr: true,
		},
		{
			name:    "both negative",
			width:   -100.0,
			height:  -50.0,
			wantErr: true,
		},
		{
			name:    "very small positive",
			width:   0.1,
			height:  0.1,
			wantErr: false,
		},
		{
			name:    "very large dimensions",
			width:   10000.0,
			height:  10000.0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDimensions(tt.width, tt.height)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDimensions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidateCoordinates tests coordinate validation.
func TestValidateCoordinates(t *testing.T) {
	tests := []struct {
		name    string
		x       float64
		y       float64
		wantErr bool
	}{
		{
			name:    "positive coordinates",
			x:       100.0,
			y:       100.0,
			wantErr: false,
		},
		{
			name:    "zero coordinates",
			x:       0.0,
			y:       0.0,
			wantErr: false,
		},
		{
			name:    "negative x (valid for slides)",
			x:       -50.0,
			y:       100.0,
			wantErr: false,
		},
		{
			name:    "negative y (valid for slides)",
			x:       100.0,
			y:       -50.0,
			wantErr: false,
		},
		{
			name:    "both negative (valid)",
			x:       -100.0,
			y:       -100.0,
			wantErr: false,
		},
		{
			name:    "extremely large positive",
			x:       1000000.0,
			y:       1000000.0,
			wantErr: false, // API will handle boundary checks
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCoordinates(tt.x, tt.y)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateCoordinates() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestMapLayoutName tests layout name mapping.
func TestMapLayoutName(t *testing.T) {
	tests := []struct {
		name   string
		layout string
		want   string
	}{
		{"blank", "blank", "BLANK"},
		{"title", "title", "TITLE"},
		{"title_body", "title_body", "TITLE_AND_BODY"},
		{"title_and_body", "title_and_body", "TITLE_AND_BODY"},
		{"one_column", "one_column", "ONE_COLUMN_TEXT"},
		{"main_point", "main_point", "MAIN_POINT"},
		{"section", "section", "SECTION_HEADER"},
		{"section_header", "section_header", "SECTION_HEADER"},
		{"title_only", "title_only", "TITLE_ONLY"},
		{"big_number", "big_number", "BIG_NUMBER"},
		{"caption", "caption", "CAPTION_ONLY"},
		{"unknown", "unknown_layout", "BLANK"}, // Default to BLANK
		{"empty", "", "BLANK"},                 // Default to BLANK
		{"uppercase", "BLANK", "BLANK"},        // Pass through uppercase
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapLayoutName(tt.layout)
			if got != tt.want {
				t.Errorf("mapLayoutName(%q) = %q, want %q", tt.layout, got, tt.want)
			}
		})
	}
}

// TestValidateSlideIndex tests slide index validation.
func TestValidateSlideIndex(t *testing.T) {
	tests := []struct {
		name       string
		slideIndex int
		slideCount int
		wantErr    bool
	}{
		{
			name:       "valid index",
			slideIndex: 1,
			slideCount: 5,
			wantErr:    false,
		},
		{
			name:       "last slide",
			slideIndex: 5,
			slideCount: 5,
			wantErr:    false,
		},
		{
			name:       "zero index",
			slideIndex: 0,
			slideCount: 5,
			wantErr:    true,
		},
		{
			name:       "negative index",
			slideIndex: -1,
			slideCount: 5,
			wantErr:    true,
		},
		{
			name:       "index too large",
			slideIndex: 6,
			slideCount: 5,
			wantErr:    true,
		},
		{
			name:       "empty presentation",
			slideIndex: 1,
			slideCount: 0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSlideIndex(tt.slideIndex, tt.slideCount)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSlideIndex() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidateText tests text validation.
func TestValidateText(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		wantErr bool
	}{
		{
			name:    "valid text",
			text:    "Hello, World!",
			wantErr: false,
		},
		{
			name:    "empty text",
			text:    "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			text:    "   ",
			wantErr: true,
		},
		{
			name:    "newlines",
			text:    "Line 1\nLine 2",
			wantErr: false,
		},
		{
			name:    "unicode",
			text:    "Hello 世界 🌍",
			wantErr: false,
		},
		{
			name:    "very long text",
			text:    string(make([]byte, 10000)),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateText(tt.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateText() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidateURL tests URL validation.
func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid http url",
			url:     "http://example.com/image.png",
			wantErr: false,
		},
		{
			name:    "valid https url",
			url:     "https://example.com/image.png",
			wantErr: false,
		},
		{
			name:    "empty url",
			url:     "",
			wantErr: true,
		},
		{
			name:    "invalid scheme",
			url:     "ftp://example.com/image.png",
			wantErr: true,
		},
		{
			name:    "no scheme",
			url:     "example.com/image.png",
			wantErr: true,
		},
		{
			name:    "malformed url",
			url:     "ht!tp://invalid",
			wantErr: true,
		},
		{
			name:    "localhost url",
			url:     "http://localhost:8080/image.png",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
