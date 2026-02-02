package docs

import (
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/api/docs/v1"
)

// Create creates a new document.
func (s *Service) Create(title string) (*CreateResult, error) {
	doc := &docs.Document{
		Title: title,
	}

	created, err := s.docs.Documents.Create(doc).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	return &CreateResult{
		DocumentID: created.DocumentId,
		Title:      created.Title,
	}, nil
}

// CreateWithContent creates a new document with initial content.
func (s *Service) CreateWithContent(title, content string) (*CreateResult, error) {
	// First create the document
	result, err := s.Create(title)
	if err != nil {
		return nil, err
	}

	// Then insert content
	if content != "" {
		_, err = s.Append(result.DocumentID, content)
		if err != nil {
			return nil, fmt.Errorf("document created but failed to add content: %w", err)
		}
	}

	return result, nil
}

// Append appends text to the end of a document.
func (s *Service) Append(documentID, text string) (*UpdateResult, error) {
	// Get the document to find the end index
	doc, err := s.Get(documentID)
	if err != nil {
		return nil, err
	}

	// Find the end index (body content ends at Body.Content[last].EndIndex - 1)
	endIndex := int64(1)
	if doc.Body != nil && len(doc.Body.Content) > 0 {
		lastElement := doc.Body.Content[len(doc.Body.Content)-1]
		endIndex = lastElement.EndIndex - 1
	}

	requests := []*docs.Request{
		{
			InsertText: &docs.InsertTextRequest{
				Text: text,
				Location: &docs.Location{
					Index: endIndex,
				},
			},
		},
	}

	_, err = s.docs.Documents.BatchUpdate(documentID, &docs.BatchUpdateDocumentRequest{
		Requests: requests,
	}).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to append text: %w", err)
	}

	return &UpdateResult{
		DocumentID: documentID,
	}, nil
}

// Prepend inserts text at the beginning of a document.
func (s *Service) Prepend(documentID, text string) (*UpdateResult, error) {
	requests := []*docs.Request{
		{
			InsertText: &docs.InsertTextRequest{
				Text: text,
				Location: &docs.Location{
					Index: 1, // Index 1 is the start of body content
				},
			},
		},
	}

	_, err := s.docs.Documents.BatchUpdate(documentID, &docs.BatchUpdateDocumentRequest{
		Requests: requests,
	}).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to prepend text: %w", err)
	}

	return &UpdateResult{
		DocumentID: documentID,
	}, nil
}

// ReplaceText replaces all occurrences of text in a document.
func (s *Service) ReplaceText(documentID, find, replace string, matchCase bool) (*UpdateResult, error) {
	requests := []*docs.Request{
		{
			ReplaceAllText: &docs.ReplaceAllTextRequest{
				ContainsText: &docs.SubstringMatchCriteria{
					Text:      find,
					MatchCase: matchCase,
				},
				ReplaceText: replace,
			},
		},
	}

	_, err := s.docs.Documents.BatchUpdate(documentID, &docs.BatchUpdateDocumentRequest{
		Requests: requests,
	}).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to replace text: %w", err)
	}

	return &UpdateResult{
		DocumentID: documentID,
	}, nil
}

// UpdateSection updates the content of a section identified by its heading.
func (s *Service) UpdateSection(documentID, heading, content string) (*UpdateResult, error) {
	// Get the document outline to find the section
	outline, err := s.Outline(documentID)
	if err != nil {
		return nil, err
	}

	// Find the section with the matching heading
	var targetSection *SectionInfo
	var nextSectionStart int64 = -1

	for i, section := range outline.Sections {
		if strings.TrimSpace(section.Heading) == strings.TrimSpace(heading) {
			targetSection = &outline.Sections[i]
			// Find where the next section starts (or end of document)
			if i+1 < len(outline.Sections) {
				nextSectionStart = outline.Sections[i+1].StartIndex
			}
			break
		}
	}

	if targetSection == nil {
		return nil, fmt.Errorf("section not found: %s", heading)
	}

	// Get the document to find the actual content range
	doc, err := s.Get(documentID)
	if err != nil {
		return nil, err
	}

	// Find the end index for deletion
	// We want to delete from after the heading to the next section or end of document
	headingEndIndex := targetSection.EndIndex

	var deleteEndIndex int64
	if nextSectionStart > 0 {
		deleteEndIndex = nextSectionStart
	} else {
		// Use end of document
		if doc.Body != nil && len(doc.Body.Content) > 0 {
			deleteEndIndex = doc.Body.Content[len(doc.Body.Content)-1].EndIndex - 1
		} else {
			deleteEndIndex = headingEndIndex
		}
	}

	// Build requests: delete existing content, then insert new content
	requests := []*docs.Request{}

	// Only delete if there's content to delete
	if deleteEndIndex > headingEndIndex {
		requests = append(requests, &docs.Request{
			DeleteContentRange: &docs.DeleteContentRangeRequest{
				Range: &docs.Range{
					StartIndex: headingEndIndex,
					EndIndex:   deleteEndIndex,
				},
			},
		})
	}

	// Insert new content after the heading
	requests = append(requests, &docs.Request{
		InsertText: &docs.InsertTextRequest{
			Text: "\n" + content + "\n",
			Location: &docs.Location{
				Index: headingEndIndex,
			},
		},
	})

	_, err = s.docs.Documents.BatchUpdate(documentID, &docs.BatchUpdateDocumentRequest{
		Requests: requests,
	}).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to update section: %w", err)
	}

	return &UpdateResult{
		DocumentID: documentID,
	}, nil
}

// BatchUpdate performs a batch update with raw requests.
func (s *Service) BatchUpdate(documentID, requestsJSON string) (*UpdateResult, error) {
	var requests []*docs.Request
	if err := json.Unmarshal([]byte(requestsJSON), &requests); err != nil {
		return nil, fmt.Errorf("failed to parse requests JSON: %w", err)
	}

	_, err := s.docs.Documents.BatchUpdate(documentID, &docs.BatchUpdateDocumentRequest{
		Requests: requests,
	}).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to batch update: %w", err)
	}

	return &UpdateResult{
		DocumentID: documentID,
	}, nil
}
