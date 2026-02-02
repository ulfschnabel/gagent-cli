package contacts

import (
	"fmt"

	"google.golang.org/api/people/v1"
)

// PersonFields defines which fields to return for contacts.
const PersonFields = "names,emailAddresses,phoneNumbers,addresses,birthdays,organizations,biographies"

// ListOptions contains options for listing contacts.
type ListOptions struct {
	PageSize  int64
	PageToken string
}

// List returns contacts from the user's Google Contacts.
func (s *Service) List(opts ListOptions) ([]ContactSummary, string, error) {
	pageSize := opts.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 1000 {
		pageSize = 1000
	}

	call := s.svc.People.Connections.List("people/me").
		PersonFields(PersonFields).
		PageSize(pageSize)

	if opts.PageToken != "" {
		call = call.PageToken(opts.PageToken)
	}

	resp, err := call.Do()
	if err != nil {
		return nil, "", fmt.Errorf("failed to list contacts: %w", err)
	}

	contacts := make([]ContactSummary, 0, len(resp.Connections))
	for _, person := range resp.Connections {
		contacts = append(contacts, parsePersonToSummary(person))
	}

	return contacts, resp.NextPageToken, nil
}

// Get returns full details for a specific contact.
func (s *Service) Get(resourceName string) (*ContactFull, error) {
	person, err := s.svc.People.Get(resourceName).
		PersonFields(PersonFields).
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}

	return parsePersonToFull(person), nil
}

// Search searches for contacts by query string.
func (s *Service) Search(query string, limit int64) ([]ContactSummary, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 30 {
		limit = 30 // People API search has max 30 results
	}

	resp, err := s.svc.People.SearchContacts().
		Query(query).
		ReadMask(PersonFields).
		PageSize(limit).
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to search contacts: %w", err)
	}

	contacts := make([]ContactSummary, 0, len(resp.Results))
	for _, result := range resp.Results {
		if result.Person != nil {
			contacts = append(contacts, parsePersonToSummary(result.Person))
		}
	}

	return contacts, nil
}

// Groups returns the list of contact groups.
func (s *Service) Groups() ([]GroupInfo, error) {
	resp, err := s.svc.ContactGroups.List().
		PageSize(200).
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list contact groups: %w", err)
	}

	groups := make([]GroupInfo, 0, len(resp.ContactGroups))
	for _, group := range resp.ContactGroups {
		groups = append(groups, GroupInfo{
			ResourceName: group.ResourceName,
			Name:         group.Name,
			MemberCount:  group.MemberCount,
			GroupType:    group.GroupType,
		})
	}

	return groups, nil
}

// parsePersonToSummary converts a People API person to ContactSummary.
func parsePersonToSummary(person *people.Person) ContactSummary {
	summary := ContactSummary{
		ResourceName: person.ResourceName,
	}

	// Get display name
	if len(person.Names) > 0 {
		summary.DisplayName = person.Names[0].DisplayName
	}

	// Get primary email
	if len(person.EmailAddresses) > 0 {
		summary.PrimaryEmail = person.EmailAddresses[0].Value
	}

	// Get primary phone
	if len(person.PhoneNumbers) > 0 {
		summary.PrimaryPhone = person.PhoneNumbers[0].Value
	}

	return summary
}

// parsePersonToFull converts a People API person to ContactFull.
func parsePersonToFull(person *people.Person) *ContactFull {
	contact := &ContactFull{
		ResourceName: person.ResourceName,
		Etag:         person.Etag,
	}

	// Parse names
	if len(person.Names) > 0 {
		name := person.Names[0]
		contact.DisplayName = name.DisplayName
		contact.GivenName = name.GivenName
		contact.FamilyName = name.FamilyName
	}

	// Parse emails
	for _, email := range person.EmailAddresses {
		contact.Emails = append(contact.Emails, EmailInfo{
			Value: email.Value,
			Type:  email.Type,
		})
	}

	// Parse phones
	for _, phone := range person.PhoneNumbers {
		contact.Phones = append(contact.Phones, PhoneInfo{
			Value: phone.Value,
			Type:  phone.Type,
		})
	}

	// Parse addresses
	for _, addr := range person.Addresses {
		contact.Addresses = append(contact.Addresses, AddressInfo{
			FormattedValue: addr.FormattedValue,
			Type:           addr.Type,
			StreetAddress:  addr.StreetAddress,
			City:           addr.City,
			Region:         addr.Region,
			PostalCode:     addr.PostalCode,
			Country:        addr.Country,
		})
	}

	// Parse birthdays
	for _, bday := range person.Birthdays {
		birthday := Birthday{}
		if bday.Date != nil {
			birthday.Date = fmt.Sprintf("%04d-%02d-%02d", bday.Date.Year, bday.Date.Month, bday.Date.Day)
		}
		if bday.Text != "" {
			birthday.Text = bday.Text
		}
		contact.Birthdays = append(contact.Birthdays, birthday)
	}

	// Parse organizations
	for _, org := range person.Organizations {
		contact.Organizations = append(contact.Organizations, OrgInfo{
			Name:  org.Name,
			Title: org.Title,
		})
	}

	// Parse notes/biographies
	if len(person.Biographies) > 0 {
		contact.Notes = person.Biographies[0].Value
	}

	return contact
}
