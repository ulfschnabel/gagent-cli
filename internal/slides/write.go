package slides

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/api/slides/v1"
)

// Create creates a new presentation.
func (s *Service) Create(title string) (*CreateResult, error) {
	pres := &slides.Presentation{
		Title: title,
	}

	created, err := s.slides.Presentations.Create(pres).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create presentation: %w", err)
	}

	return &CreateResult{
		PresentationID: created.PresentationId,
		Title:          created.Title,
	}, nil
}

// AddSlide adds a new slide to a presentation.
func (s *Service) AddSlide(presentationID, layout string) (*UpdateResult, error) {
	// Map layout names to predefined layout IDs
	layoutID := mapLayoutName(layout)

	requests := []*slides.Request{
		{
			CreateSlide: &slides.CreateSlideRequest{
				SlideLayoutReference: &slides.LayoutReference{
					PredefinedLayout: layoutID,
				},
			},
		},
	}

	_, err := s.slides.Presentations.BatchUpdate(presentationID, &slides.BatchUpdatePresentationRequest{
		Requests: requests,
	}).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to add slide: %w", err)
	}

	return &UpdateResult{
		PresentationID: presentationID,
	}, nil
}

// DeleteSlide deletes a slide from a presentation.
func (s *Service) DeleteSlide(presentationID string, slideIndex int) (*UpdateResult, error) {
	// Get the presentation to find the slide ID
	pres, err := s.Get(presentationID)
	if err != nil {
		return nil, err
	}

	if slideIndex < 1 || slideIndex > len(pres.Slides) {
		return nil, fmt.Errorf("slide index out of range: %d (presentation has %d slides)", slideIndex, len(pres.Slides))
	}

	slideID := pres.Slides[slideIndex-1].ObjectId

	requests := []*slides.Request{
		{
			DeleteObject: &slides.DeleteObjectRequest{
				ObjectId: slideID,
			},
		},
	}

	_, err = s.slides.Presentations.BatchUpdate(presentationID, &slides.BatchUpdatePresentationRequest{
		Requests: requests,
	}).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to delete slide: %w", err)
	}

	return &UpdateResult{
		PresentationID: presentationID,
	}, nil
}

// UpdateText replaces text on a specific slide.
func (s *Service) UpdateText(presentationID string, slideIndex int, find, replace string) (*UpdateResult, error) {
	// Get the presentation to find the slide ID
	pres, err := s.Get(presentationID)
	if err != nil {
		return nil, err
	}

	if slideIndex < 1 || slideIndex > len(pres.Slides) {
		return nil, fmt.Errorf("slide index out of range: %d (presentation has %d slides)", slideIndex, len(pres.Slides))
	}

	slideID := pres.Slides[slideIndex-1].ObjectId

	requests := []*slides.Request{
		{
			ReplaceAllText: &slides.ReplaceAllTextRequest{
				ContainsText: &slides.SubstringMatchCriteria{
					Text:      find,
					MatchCase: true,
				},
				ReplaceText: replace,
				PageObjectIds: []string{slideID},
			},
		},
	}

	_, err = s.slides.Presentations.BatchUpdate(presentationID, &slides.BatchUpdatePresentationRequest{
		Requests: requests,
	}).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to update text: %w", err)
	}

	return &UpdateResult{
		PresentationID: presentationID,
	}, nil
}

// AddText adds a text box to a slide.
func (s *Service) AddText(presentationID string, slideIndex int, text string, x, y, width, height float64) (*UpdateResult, error) {
	// Get the presentation to find the slide ID
	pres, err := s.Get(presentationID)
	if err != nil {
		return nil, err
	}

	if slideIndex < 1 || slideIndex > len(pres.Slides) {
		return nil, fmt.Errorf("slide index out of range: %d (presentation has %d slides)", slideIndex, len(pres.Slides))
	}

	slideID := pres.Slides[slideIndex-1].ObjectId
	elementID := fmt.Sprintf("textbox_%s", uuid.New().String()[:8])

	requests := []*slides.Request{
		{
			CreateShape: &slides.CreateShapeRequest{
				ObjectId:  elementID,
				ShapeType: "TEXT_BOX",
				ElementProperties: &slides.PageElementProperties{
					PageObjectId: slideID,
					Size: &slides.Size{
						Width:  &slides.Dimension{Magnitude: width, Unit: "PT"},
						Height: &slides.Dimension{Magnitude: height, Unit: "PT"},
					},
					Transform: &slides.AffineTransform{
						ScaleX:     1,
						ScaleY:     1,
						TranslateX: x,
						TranslateY: y,
						Unit:       "PT",
					},
				},
			},
		},
		{
			InsertText: &slides.InsertTextRequest{
				ObjectId: elementID,
				Text:     text,
			},
		},
	}

	_, err = s.slides.Presentations.BatchUpdate(presentationID, &slides.BatchUpdatePresentationRequest{
		Requests: requests,
	}).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to add text: %w", err)
	}

	return &UpdateResult{
		PresentationID: presentationID,
	}, nil
}

// AddImage adds an image to a slide.
func (s *Service) AddImage(presentationID string, slideIndex int, imageURL string, x, y, width, height float64) (*UpdateResult, error) {
	// Get the presentation to find the slide ID
	pres, err := s.Get(presentationID)
	if err != nil {
		return nil, err
	}

	if slideIndex < 1 || slideIndex > len(pres.Slides) {
		return nil, fmt.Errorf("slide index out of range: %d (presentation has %d slides)", slideIndex, len(pres.Slides))
	}

	slideID := pres.Slides[slideIndex-1].ObjectId
	elementID := fmt.Sprintf("image_%s", uuid.New().String()[:8])

	requests := []*slides.Request{
		{
			CreateImage: &slides.CreateImageRequest{
				ObjectId: elementID,
				Url:      imageURL,
				ElementProperties: &slides.PageElementProperties{
					PageObjectId: slideID,
					Size: &slides.Size{
						Width:  &slides.Dimension{Magnitude: width, Unit: "PT"},
						Height: &slides.Dimension{Magnitude: height, Unit: "PT"},
					},
					Transform: &slides.AffineTransform{
						ScaleX:     1,
						ScaleY:     1,
						TranslateX: x,
						TranslateY: y,
						Unit:       "PT",
					},
				},
			},
		},
	}

	_, err = s.slides.Presentations.BatchUpdate(presentationID, &slides.BatchUpdatePresentationRequest{
		Requests: requests,
	}).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to add image: %w", err)
	}

	return &UpdateResult{
		PresentationID: presentationID,
	}, nil
}

// BatchUpdate performs a batch update with raw requests.
func (s *Service) BatchUpdate(presentationID, requestsJSON string) (*UpdateResult, error) {
	var requests []*slides.Request
	if err := json.Unmarshal([]byte(requestsJSON), &requests); err != nil {
		return nil, fmt.Errorf("failed to parse requests JSON: %w", err)
	}

	_, err := s.slides.Presentations.BatchUpdate(presentationID, &slides.BatchUpdatePresentationRequest{
		Requests: requests,
	}).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to batch update: %w", err)
	}

	return &UpdateResult{
		PresentationID: presentationID,
	}, nil
}

// mapLayoutName maps user-friendly layout names to API predefined layout IDs.
func mapLayoutName(layout string) string {
	layouts := map[string]string{
		"blank":          "BLANK",
		"title":          "TITLE",
		"title_body":     "TITLE_AND_BODY",
		"title_and_body": "TITLE_AND_BODY",
		"one_column":     "ONE_COLUMN_TEXT",
		"main_point":     "MAIN_POINT",
		"section":        "SECTION_HEADER",
		"section_header": "SECTION_HEADER",
		"title_only":     "TITLE_ONLY",
		"big_number":     "BIG_NUMBER",
		"caption":        "CAPTION_ONLY",
	}

	if mapped, ok := layouts[layout]; ok {
		return mapped
	}
	return "BLANK"
}
