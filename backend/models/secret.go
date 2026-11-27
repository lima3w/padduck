package models

import "time"

// Secret is the metadata of a stored secret. The value is never part of this
// type: it is only reachable through SecretsService.Resolve.
type Secret struct {
	ID             int64      `json:"id"`
	OrganizationID int64      `json:"organization_id"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	Version        int        `json:"version"`
	Exists         bool       `json:"exists"`
	CreatedBy      *int64     `json:"created_by,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	RotatedAt      *time.Time `json:"rotated_at,omitempty"`
}

// SecretVersion is the metadata of one stored version of a secret.
type SecretVersion struct {
	Version   int       `json:"version"`
	Current   bool      `json:"current"`
	CreatedBy *int64    `json:"created_by,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
