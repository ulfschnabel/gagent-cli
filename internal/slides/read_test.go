package slides

import (
	"testing"

	"google.golang.org/api/slides/v1"
)

// TestEMUPerPointConstant verifies the conversion constant is correct.
func TestEMUPerPointConstant(t *testing.T) {
	// According to Google Slides API: 1 point = 12,700 EMUs
	expected := 12700.0
	if EMUPerPoint != expected {
		t.Errorf("EMUPerPoint = %f, want %f", EMUPerPoint, expected)
	}
}

// TestParsePageElement_CoordinateConversion tests EMU to PT conversion.
func TestParsePageElement_CoordinateConversion(t *testing.T) {
	tests := []struct {
		name           string
		element        *slides.PageElement
		wantX          float64
		wantY          float64
		wantWidth      float64
		wantHeight     float64
	}{
		{
			name: "standard coordinates",
			element: &slides.PageElement{
				ObjectId: "test-shape-1",
				Shape: &slides.Shape{
					ShapeType: "TEXT_BOX",
				},
				Transform: &slides.AffineTransform{
					TranslateX: 1270000, // 100 points in EMUs
					TranslateY: 2540000, // 200 points in EMUs
				},
				Size: &slides.Size{
					Width:  &slides.Dimension{Magnitude: 2540000}, // 200 points
					Height: &slides.Dimension{Magnitude: 635000},  // 50 points
				},
			},
			wantX:      100.0,
			wantY:      200.0,
			wantWidth:  200.0,
			wantHeight: 50.0,
		},
		{
			name: "zero coordinates",
			element: &slides.PageElement{
				ObjectId: "test-shape-2",
				Shape: &slides.Shape{
					ShapeType: "TEXT_BOX",
				},
				Transform: &slides.AffineTransform{
					TranslateX: 0,
					TranslateY: 0,
				},
				Size: &slides.Size{
					Width:  &slides.Dimension{Magnitude: 1270000},
					Height: &slides.Dimension{Magnitude: 1270000},
				},
			},
			wantX:      0.0,
			wantY:      0.0,
			wantWidth:  100.0,
			wantHeight: 100.0,
		},
		{
			name: "large coordinates",
			element: &slides.PageElement{
				ObjectId: "test-shape-3",
				Shape: &slides.Shape{
					ShapeType: "TEXT_BOX",
				},
				Transform: &slides.AffineTransform{
					TranslateX: 63500000, // 5000 points
					TranslateY: 63500000, // 5000 points
				},
				Size: &slides.Size{
					Width:  &slides.Dimension{Magnitude: 12700000}, // 1000 points
					Height: &slides.Dimension{Magnitude: 12700000}, // 1000 points
				},
			},
			wantX:      5000.0,
			wantY:      5000.0,
			wantWidth:  1000.0,
			wantHeight: 1000.0,
		},
		{
			name: "fractional conversion",
			element: &slides.PageElement{
				ObjectId: "test-shape-4",
				Shape: &slides.Shape{
					ShapeType: "TEXT_BOX",
				},
				Transform: &slides.AffineTransform{
					TranslateX: 1270001, // 100.00007874... points
					TranslateY: 1270001,
				},
				Size: &slides.Size{
					Width:  &slides.Dimension{Magnitude: 1270001},
					Height: &slides.Dimension{Magnitude: 1270001},
				},
			},
			wantX:      100.00007874015748,
			wantY:      100.00007874015748,
			wantWidth:  100.00007874015748,
			wantHeight: 100.00007874015748,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := parsePageElement(tt.element)

			if info == nil {
				t.Fatal("parsePageElement returned nil")
			}

			// Use a small epsilon for floating point comparison
			epsilon := 0.00001

			if abs(info.X-tt.wantX) > epsilon {
				t.Errorf("X = %f, want %f", info.X, tt.wantX)
			}
			if abs(info.Y-tt.wantY) > epsilon {
				t.Errorf("Y = %f, want %f", info.Y, tt.wantY)
			}
			if abs(info.Width-tt.wantWidth) > epsilon {
				t.Errorf("Width = %f, want %f", info.Width, tt.wantWidth)
			}
			if abs(info.Height-tt.wantHeight) > epsilon {
				t.Errorf("Height = %f, want %f", info.Height, tt.wantHeight)
			}
		})
	}
}

// TestParsePageElement_NilHandling tests nil pointer handling.
func TestParsePageElement_NilHandling(t *testing.T) {
	tests := []struct {
		name    string
		element *slides.PageElement
	}{
		{
			name: "nil transform",
			element: &slides.PageElement{
				ObjectId: "test-1",
				Shape:    &slides.Shape{ShapeType: "TEXT_BOX"},
				Size: &slides.Size{
					Width:  &slides.Dimension{Magnitude: 1270000},
					Height: &slides.Dimension{Magnitude: 1270000},
				},
			},
		},
		{
			name: "nil size",
			element: &slides.PageElement{
				ObjectId: "test-2",
				Shape:    &slides.Shape{ShapeType: "TEXT_BOX"},
				Transform: &slides.AffineTransform{
					TranslateX: 1270000,
					TranslateY: 1270000,
				},
			},
		},
		{
			name: "nil width dimension",
			element: &slides.PageElement{
				ObjectId: "test-3",
				Shape:    &slides.Shape{ShapeType: "TEXT_BOX"},
				Size: &slides.Size{
					Height: &slides.Dimension{Magnitude: 1270000},
				},
			},
		},
		{
			name: "nil height dimension",
			element: &slides.PageElement{
				ObjectId: "test-4",
				Shape:    &slides.Shape{ShapeType: "TEXT_BOX"},
				Size: &slides.Size{
					Width: &slides.Dimension{Magnitude: 1270000},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			info := parsePageElement(tt.element)
			if info == nil {
				t.Fatal("parsePageElement returned nil")
			}
		})
	}
}

// TestParsePageElement_ElementTypes tests different element types.
func TestParsePageElement_ElementTypes(t *testing.T) {
	tests := []struct {
		name     string
		element  *slides.PageElement
		wantType string
	}{
		{
			name: "shape element",
			element: &slides.PageElement{
				ObjectId: "test-1",
				Shape:    &slides.Shape{ShapeType: "TEXT_BOX"},
			},
			wantType: "shape",
		},
		{
			name: "image element",
			element: &slides.PageElement{
				ObjectId: "test-2",
				Image:    &slides.Image{},
			},
			wantType: "image",
		},
		{
			name: "table element",
			element: &slides.PageElement{
				ObjectId: "test-3",
				Table:    &slides.Table{},
			},
			wantType: "table",
		},
		{
			name: "line element",
			element: &slides.PageElement{
				ObjectId: "test-4",
				Line:     &slides.Line{},
			},
			wantType: "line",
		},
		{
			name: "video element",
			element: &slides.PageElement{
				ObjectId: "test-5",
				Video:    &slides.Video{},
			},
			wantType: "video",
		},
		{
			name: "chart element",
			element: &slides.PageElement{
				ObjectId:    "test-6",
				SheetsChart: &slides.SheetsChart{},
			},
			wantType: "chart",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := parsePageElement(tt.element)
			if info == nil {
				t.Fatal("parsePageElement returned nil")
			}
			if info.Type != tt.wantType {
				t.Errorf("Type = %s, want %s", info.Type, tt.wantType)
			}
		})
	}
}

// TestParsePageElement_UnknownElement tests unknown element type handling.
func TestParsePageElement_UnknownElement(t *testing.T) {
	element := &slides.PageElement{
		ObjectId: "test-unknown",
		// No specific element type set
	}

	info := parsePageElement(element)
	if info != nil {
		t.Errorf("parsePageElement should return nil for unknown element type, got %v", info)
	}
}

// abs returns the absolute value of a float64.
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
