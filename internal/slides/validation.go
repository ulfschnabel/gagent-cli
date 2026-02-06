package slides

import (
	"fmt"
	"net/url"
	"strings"
)

// validateDimensions checks if width and height are positive.
func validateDimensions(width, height float64) error {
	if width <= 0 {
		return fmt.Errorf("width must be positive, got %f", width)
	}
	if height <= 0 {
		return fmt.Errorf("height must be positive, got %f", height)
	}
	return nil
}

// validateCoordinates checks if coordinates are valid.
// Note: Negative coordinates are allowed in Google Slides (elements can be positioned
// outside visible area), so we don't restrict them here.
func validateCoordinates(x, y float64) error {
	// Currently no restrictions on coordinates
	// Google Slides API will handle boundary validation
	return nil
}

// validateSlideIndex checks if slide index is valid (1-indexed).
func validateSlideIndex(slideIndex, slideCount int) error {
	if slideIndex < 1 {
		return fmt.Errorf("slide index must be at least 1, got %d", slideIndex)
	}
	if slideIndex > slideCount {
		return fmt.Errorf("slide index %d out of range (presentation has %d slides)", slideIndex, slideCount)
	}
	return nil
}

// validateText checks if text is non-empty.
func validateText(text string) error {
	if strings.TrimSpace(text) == "" {
		return fmt.Errorf("text cannot be empty or whitespace only")
	}
	return nil
}

// validateURL checks if URL is valid and uses http/https scheme.
func validateURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("url cannot be empty")
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid url: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("url must use http or https scheme, got %s", parsedURL.Scheme)
	}

	return nil
}
