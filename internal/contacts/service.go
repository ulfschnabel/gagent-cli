// Package contacts provides Google People API operations for contacts.
package contacts

import (
	"context"
	"fmt"
	"net/http"

	"google.golang.org/api/option"
	"google.golang.org/api/people/v1"
)

// Service wraps the People API service.
type Service struct {
	svc *people.Service
}

// NewService creates a new People API service.
func NewService(ctx context.Context, client *http.Client) (*Service, error) {
	svc, err := people.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create People service: %w", err)
	}
	return &Service{svc: svc}, nil
}

// ContactSummary represents a summary of a contact.
type ContactSummary struct {
	ResourceName string `json:"resource_name"`
	DisplayName  string `json:"display_name"`
	PrimaryEmail string `json:"primary_email,omitempty"`
	PrimaryPhone string `json:"primary_phone,omitempty"`
}

// ContactFull represents full contact details.
type ContactFull struct {
	ResourceName string        `json:"resource_name"`
	Etag         string        `json:"etag,omitempty"`
	DisplayName  string        `json:"display_name"`
	GivenName    string        `json:"given_name,omitempty"`
	FamilyName   string        `json:"family_name,omitempty"`
	Emails       []EmailInfo   `json:"emails,omitempty"`
	Phones       []PhoneInfo   `json:"phones,omitempty"`
	Addresses    []AddressInfo `json:"addresses,omitempty"`
	Birthdays    []Birthday    `json:"birthdays,omitempty"`
	Organizations []OrgInfo    `json:"organizations,omitempty"`
	Notes        string        `json:"notes,omitempty"`
}

// EmailInfo represents an email address.
type EmailInfo struct {
	Value string `json:"value"`
	Type  string `json:"type,omitempty"`
}

// PhoneInfo represents a phone number.
type PhoneInfo struct {
	Value string `json:"value"`
	Type  string `json:"type,omitempty"`
}

// AddressInfo represents a postal address.
type AddressInfo struct {
	FormattedValue string `json:"formatted_value,omitempty"`
	Type           string `json:"type,omitempty"`
	StreetAddress  string `json:"street_address,omitempty"`
	City           string `json:"city,omitempty"`
	Region         string `json:"region,omitempty"`
	PostalCode     string `json:"postal_code,omitempty"`
	Country        string `json:"country,omitempty"`
}

// Birthday represents a birthday.
type Birthday struct {
	Date string `json:"date,omitempty"`
	Text string `json:"text,omitempty"`
}

// OrgInfo represents organization information.
type OrgInfo struct {
	Name  string `json:"name,omitempty"`
	Title string `json:"title,omitempty"`
}

// GroupInfo represents a contact group.
type GroupInfo struct {
	ResourceName string `json:"resource_name"`
	Name         string `json:"name"`
	MemberCount  int64  `json:"member_count"`
	GroupType    string `json:"group_type,omitempty"`
}

// CreateContactResult represents the result of creating a contact.
type CreateContactResult struct {
	ResourceName string `json:"resource_name"`
	DisplayName  string `json:"display_name"`
}
