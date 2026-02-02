package contacts

import (
	"encoding/json"
	"fmt"

	"google.golang.org/api/people/v1"
)

// CreateContactOptions contains options for creating a contact.
type CreateContactOptions struct {
	GivenName  string
	FamilyName string
	Email      string
	Phone      string
}

// Create creates a new contact.
func (s *Service) Create(opts CreateContactOptions) (*CreateContactResult, error) {
	person := &people.Person{
		Names: []*people.Name{
			{
				GivenName:  opts.GivenName,
				FamilyName: opts.FamilyName,
			},
		},
	}

	if opts.Email != "" {
		person.EmailAddresses = []*people.EmailAddress{
			{Value: opts.Email},
		}
	}

	if opts.Phone != "" {
		person.PhoneNumbers = []*people.PhoneNumber{
			{Value: opts.Phone},
		}
	}

	created, err := s.svc.People.CreateContact(person).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create contact: %w", err)
	}

	displayName := ""
	if len(created.Names) > 0 {
		displayName = created.Names[0].DisplayName
	}

	return &CreateContactResult{
		ResourceName: created.ResourceName,
		DisplayName:  displayName,
	}, nil
}

// UpdateContactOptions contains options for updating a contact.
type UpdateContactOptions struct {
	GivenName  *string
	FamilyName *string
	Email      *string
	Phone      *string
}

// Update updates an existing contact.
func (s *Service) Update(resourceName string, opts UpdateContactOptions) (*CreateContactResult, error) {
	// First, get the existing contact to get the etag
	existing, err := s.svc.People.Get(resourceName).
		PersonFields(PersonFields).
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get contact for update: %w", err)
	}

	// Build update mask based on what's being updated
	updateFields := []string{}
	person := &people.Person{
		Etag: existing.Etag,
	}

	// Update names if provided
	if opts.GivenName != nil || opts.FamilyName != nil {
		name := &people.Name{}
		if len(existing.Names) > 0 {
			name.GivenName = existing.Names[0].GivenName
			name.FamilyName = existing.Names[0].FamilyName
		}
		if opts.GivenName != nil {
			name.GivenName = *opts.GivenName
		}
		if opts.FamilyName != nil {
			name.FamilyName = *opts.FamilyName
		}
		person.Names = []*people.Name{name}
		updateFields = append(updateFields, "names")
	}

	// Update email if provided
	if opts.Email != nil {
		person.EmailAddresses = []*people.EmailAddress{
			{Value: *opts.Email},
		}
		updateFields = append(updateFields, "emailAddresses")
	}

	// Update phone if provided
	if opts.Phone != nil {
		person.PhoneNumbers = []*people.PhoneNumber{
			{Value: *opts.Phone},
		}
		updateFields = append(updateFields, "phoneNumbers")
	}

	if len(updateFields) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	updated, err := s.svc.People.UpdateContact(resourceName, person).
		UpdatePersonFields(joinFields(updateFields)).
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to update contact: %w", err)
	}

	displayName := ""
	if len(updated.Names) > 0 {
		displayName = updated.Names[0].DisplayName
	}

	return &CreateContactResult{
		ResourceName: updated.ResourceName,
		DisplayName:  displayName,
	}, nil
}

// Delete removes a contact.
func (s *Service) Delete(resourceName string) error {
	_, err := s.svc.People.DeleteContact(resourceName).Do()
	if err != nil {
		return fmt.Errorf("failed to delete contact: %w", err)
	}
	return nil
}

// CreateRaw creates a contact from JSON.
func (s *Service) CreateRaw(personJSON string) (*CreateContactResult, error) {
	var person people.Person
	if err := json.Unmarshal([]byte(personJSON), &person); err != nil {
		return nil, fmt.Errorf("failed to parse contact JSON: %w", err)
	}

	created, err := s.svc.People.CreateContact(&person).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create contact: %w", err)
	}

	displayName := ""
	if len(created.Names) > 0 {
		displayName = created.Names[0].DisplayName
	}

	return &CreateContactResult{
		ResourceName: created.ResourceName,
		DisplayName:  displayName,
	}, nil
}

// UpdateRaw updates a contact from JSON.
func (s *Service) UpdateRaw(resourceName string, personJSON string, updateFields string) (*CreateContactResult, error) {
	// Get the existing contact for the etag
	existing, err := s.svc.People.Get(resourceName).
		PersonFields(PersonFields).
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get contact for update: %w", err)
	}

	var person people.Person
	if err := json.Unmarshal([]byte(personJSON), &person); err != nil {
		return nil, fmt.Errorf("failed to parse contact JSON: %w", err)
	}
	person.Etag = existing.Etag

	updated, err := s.svc.People.UpdateContact(resourceName, &person).
		UpdatePersonFields(updateFields).
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to update contact: %w", err)
	}

	displayName := ""
	if len(updated.Names) > 0 {
		displayName = updated.Names[0].DisplayName
	}

	return &CreateContactResult{
		ResourceName: updated.ResourceName,
		DisplayName:  displayName,
	}, nil
}

// joinFields joins field names with commas.
func joinFields(fields []string) string {
	result := ""
	for i, f := range fields {
		if i > 0 {
			result += ","
		}
		result += f
	}
	return result
}
