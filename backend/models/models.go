package models

import "time"

// User represents a system user
type User struct {
	ID        int64
	Username  string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Section represents an IP address section/grouping
type Section struct {
	ID        int64
	Name      string
	Description string
	CreatedBy *int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Subnet represents a network subnet
type Subnet struct {
	ID             int64
	SectionID      int64
	NetworkAddress string
	PrefixLength   int
	Description    string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// IPAddress represents an individual IP address
type IPAddress struct {
	ID        int64
	SubnetID  int64
	Address   string
	Hostname  string
	Status    string // available, assigned, reserved
	AssignedTo *string
	AssignedAt *time.Time
	ExpiresAt  *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
