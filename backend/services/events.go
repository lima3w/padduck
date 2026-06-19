package services

import "padduck/models"

// Event is the base interface for all domain events.
type Event interface {
	EventType() string
}

// AuditableEvent is optionally implemented by events that should produce an audit log entry.
type AuditableEvent interface {
	Event
	ToAuditEntry() AuditEntry
}

// --- Network / Subnet events ---

type SubnetCreatedEvent struct {
	Subnet  *models.Subnet
	ActorID int64
	Actor   string
}

func (e SubnetCreatedEvent) EventType() string { return "subnet.created" }

type SubnetUpdatedEvent struct {
	Subnet  *models.Subnet
	ActorID int64
	Actor   string
}

func (e SubnetUpdatedEvent) EventType() string { return "subnet.updated" }

type SubnetDeletedEvent struct {
	SubnetID int64
	ActorID  int64
	Actor    string
}

func (e SubnetDeletedEvent) EventType() string { return "subnet.deleted" }

// --- IP address events ---

type IPAllocatedEvent struct {
	IPAddress *models.IPAddress
	ActorID   int64
	Actor     string
}

func (e IPAllocatedEvent) EventType() string { return "ip.allocated" }

type IPReleasedEvent struct {
	IPAddressID int64
	ActorID     int64
	Actor       string
}

func (e IPReleasedEvent) EventType() string { return "ip.released" }

// --- Scan events ---

type ScanCompletedEvent struct {
	JobID   int64
	AgentID int64
	Found   int
}

func (e ScanCompletedEvent) EventType() string { return "scan.completed" }

// --- Auth events ---

type UserLoggedInEvent struct {
	UserID   int64
	Username string
}

func (e UserLoggedInEvent) EventType() string { return "user.login" }

type UserLoggedOutEvent struct {
	UserID   int64
	Username string
}

func (e UserLoggedOutEvent) EventType() string { return "user.logout" }

// --- Workflow events ---

type RequestCommentAddedEvent struct {
	RequestType string
	RequestID   int64
	AuthorID    int64
	Author      string
	CommentID   int64
}

func (e RequestCommentAddedEvent) EventType() string { return "request.comment_added" }

func (e RequestCommentAddedEvent) ToAuditEntry() AuditEntry {
	return AuditEntry{
		UserID: &e.AuthorID, Username: e.Author,
		Action:       "request_comment_added",
		ResourceType: e.RequestType + "_request",
		ResourceID:   &e.RequestID,
	}
}
