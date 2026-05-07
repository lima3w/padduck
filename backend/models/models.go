package models

import "time"

// User represents a system user
type User struct {
	ID                     int64
	Username               string
	Email                  string
	PasswordHash           string
	Role                   string // admin, user, viewer
	State                  string // active, pending_email_verification, pending_admin_approval, rejected, disabled, suspended
	LastLoginAt            *time.Time
	SuspendedAt            *time.Time
	SuspendedBy            *int64
	SuspensionReason       *string
	PrivacyAcceptedAt      *time.Time
	PrivacyAcceptedVersion *string
	DeletionRequestedAt    *time.Time
	AnonymizedAt           *time.Time
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

// Config represents an application configuration key-value pair
type Config struct {
	Key       string
	Value     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// EmailVerification represents an email verification token
type EmailVerification struct {
	ID        int64
	UserID    int64
	TokenHash string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UserApproval represents an admin approval record for a pending user
type UserApproval struct {
	ID              int64
	UserID          int64
	Status          string // pending, approved, rejected
	ReviewedBy      *int64
	ReviewedAt      *time.Time
	RejectionReason *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Session represents an authenticated web session
type Session struct {
	ID                int64
	UserID            int64
	TokenHash         string
	DeviceName        string
	IPAddress         string
	UserAgent         string
	LastUsedAt        time.Time
	AbsoluteExpiresAt time.Time
	IsImpersonation   bool
	ImpersonatedBy    *int64
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// APIToken represents an API authentication token
type APIToken struct {
	ID                     int64      `json:"id"`
	UserID                 int64      `json:"user_id"`
	TokenHash              string     `json:"-"`
	Name                   string     `json:"name"`
	Scope                  string     `json:"scope"`
	UsageCount             int64      `json:"usage_count"`
	LastUsedAt             *time.Time `json:"last_used_at,omitempty"`
	LastUsedIP             *string    `json:"last_used_ip,omitempty"`
	ExpiresAt              *time.Time `json:"expires_at,omitempty"`
	RotationGraceExpiresAt *time.Time `json:"rotation_grace_expires_at,omitempty"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

// PasswordReset represents a password reset request
type PasswordReset struct {
	ID        int64
	UserID    int64
	TokenHash string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Section represents an IP address section/grouping
type Section struct {
	ID        int64
	Name      string
	Description string
	CreatedBy int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Subnet represents a network subnet
type Subnet struct {
	ID               int64
	SectionID        int64
	NetworkAddress   string
	PrefixLength     int
	Description      string
	Gateway          *string
	AutoReserveFirst bool
	AutoReserveLast  bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// IPTag represents a named, colour-coded label for IP addresses
type IPTag struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Colour      string    `json:"colour"`
	Description *string   `json:"description,omitempty"`
	IsSystem    bool      `json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
}

// IPAddress represents an individual IP address
type IPAddress struct {
	ID         int64
	SubnetID   int64
	Address    string
	Hostname   string
	Status     string // available, assigned, reserved
	AssignedTo *string
	AssignedAt *time.Time
	ExpiresAt  *time.Time
	TagID      *int64
	Tag        *IPTag
	LastSeen   *time.Time
	MACAddress *string
	PTRRecord  *string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Role represents a named set of permissions
type Role struct {
	ID          int64
	Name        string
	Description string
	IsSystem    bool
	Permissions []*RolePermission
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// RolePermission is a single permission entry within a role
type RolePermission struct {
	ID           int64
	RoleID       int64
	Permission   string
	ResourceType *string // nil = all resources
	ResourceID   *int64  // nil = all resources of that type
	CreatedAt    time.Time
}

// UserRole links a user to a role
type UserRole struct {
	ID        int64
	UserID    int64
	RoleID    int64
	CreatedAt time.Time
}

// VRF represents a Virtual Routing and Forwarding instance
type VRF struct {
	ID                 int64
	Name               string
	RouteDistinguisher string
	Description        string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// VLAN represents a Virtual LAN segment
type VLAN struct {
	ID          int64
	VRFID       *int64
	VlanID      int
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// UserMFASettings tracks MFA status for a user
type UserMFASettings struct {
	ID                    int64
	UserID                int64
	TOTPEnabled           bool
	BackupCodesGeneratedAt *time.Time
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// UserTOTPSecret stores the encrypted TOTP secret
type UserTOTPSecret struct {
	ID              int64
	UserID          int64
	EncryptedSecret []byte
	Verified        bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// UserBackupCode is a single hashed recovery code
type UserBackupCode struct {
	ID        int64
	UserID    int64
	CodeHash  string
	Used      bool
	UsedAt    *time.Time
	CreatedAt time.Time
}

// MFAChallenge is a short-lived token issued after password auth when MFA is required
type MFAChallenge struct {
	ID            int64
	UserID        int64
	ChallengeHash string
	ExpiresAt     time.Time
	CompletedAt   *time.Time
	CreatedAt     time.Time
}

// LoginAttempt records each login attempt for brute force detection and audit
type LoginAttempt struct {
	ID            int64
	Username      string
	IPAddress     string
	UserAgent     string
	Success       bool
	FailureReason string
	CreatedAt     time.Time
}

// AccountLockout records an account lockout event with unlock metadata
type AccountLockout struct {
	ID                   int64
	UserID               int64
	LockedAt             time.Time
	UnlockAt             time.Time
	UnlockTokenHash      *string
	UnlockTokenExpiresAt *time.Time
	UnlockTokenUsedAt    *time.Time
	Reason               string
	LockoutCount         int
	UnlockedAt           *time.Time
	UnlockedBy           *int64
	CreatedAt            time.Time
}

// SecurityNotification tracks sent security alerts for rate limiting
type SecurityNotification struct {
	ID               int64
	UserID           int64
	NotificationType string
	IPAddress        string
	SentAt           time.Time
}

// AuditLog records all significant user actions for compliance and security review
type AuditLog struct {
	ID           int64
	UserID       *int64
	Username     string
	Action       string
	ResourceType string
	ResourceID   *int64
	ResourceName string
	OldValues    *string // JSON
	NewValues    *string // JSON
	IPAddress    string
	UserAgent    string
	Status       string // success, failure
	ErrorMessage string
	CreatedAt    time.Time
}

// AuditLogFilter defines search criteria for querying audit logs
type AuditLogFilter struct {
	UserID       *int64
	Username     string
	Action       string
	ResourceType string
	IPAddress    string
	Status       string
	Since        *time.Time
	Until        *time.Time
	Limit        int
	Offset       int
}

// NotificationPreferences stores per-user email notification opt-in settings
type NotificationPreferences struct {
	ID              int64     `json:"id"`
	UserID          int64     `json:"user_id"`
	LoginSuccess    bool      `json:"login_success"`
	LoginFailed     bool      `json:"login_failed"`
	AccountLocked   bool      `json:"account_locked"`
	PasswordChanged bool      `json:"password_changed"`
	MFAChanges      bool      `json:"mfa_changes"`
	APITokenChanges bool      `json:"api_token_changes"`
	RoleChanges     bool      `json:"role_changes"`
	SessionRevoked  bool      `json:"session_revoked"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// ScanJob represents a scheduled or one-time network discovery scan
type ScanJob struct {
	ID           int64      `json:"id"`
	Name         string     `json:"name"`
	SubnetIDs    []int64    `json:"subnet_ids"`
	ScheduleCron *string    `json:"schedule_cron,omitempty"`
	IsActive     bool       `json:"is_active"`
	LastRunAt    *time.Time `json:"last_run_at,omitempty"`
	NextRunAt    *time.Time `json:"next_run_at,omitempty"`
	CreatedBy    int64      `json:"created_by"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// ScanResult records the outcome of scanning a single IP address
type ScanResult struct {
	ID             int64     `json:"id"`
	JobID          int64     `json:"job_id"`
	SubnetID       int64     `json:"subnet_id"`
	IPAddressID    *int64    `json:"ip_address_id,omitempty"`
	IPAddress      string    `json:"ip_address"`
	IsAlive        bool      `json:"is_alive"`
	ResponseTimeMs *int64    `json:"response_time_ms,omitempty"`
	ScannedAt      time.Time `json:"scanned_at"`
}

// DashboardSummary holds aggregate IPAM statistics for the dashboard
type DashboardSummary struct {
	TotalSections    int64              `json:"total_sections"`
	TotalSubnets     int64              `json:"total_subnets"`
	TotalIPs         int64              `json:"total_ips"`
	UsedIPs          int64              `json:"used_ips"`
	UtilisationPct   float64            `json:"utilisation_pct"`
	TopSubnets       []SubnetUtilisation `json:"top_subnets"`
}

// SubnetUtilisation holds utilisation data for a single subnet
type SubnetUtilisation struct {
	ID              int64   `json:"id"`
	CIDR            string  `json:"cidr"`
	Description     string  `json:"description"`
	Used            int64   `json:"used"`
	Total           int64   `json:"total"`
	UtilisationPct  float64 `json:"utilisation_pct"`
}

// DashboardActivity is a single recent activity entry from audit_logs
type DashboardActivity struct {
	ID          int64   `json:"id"`
	Action      string  `json:"action"`
	EntityType  string  `json:"entity_type"`
	EntityID    *int64  `json:"entity_id,omitempty"`
	UserID      *int64  `json:"user_id,omitempty"`
	Username    string  `json:"username"`
	Description string  `json:"description"`
	CreatedAt   string  `json:"created_at"`
}

// SubnetTreeNode represents a subnet in the tree hierarchy view
type SubnetTreeNode struct {
	ID             int64            `json:"id"`
	CIDR           string           `json:"cidr"`
	Description    string           `json:"description"`
	Used           int64            `json:"used"`
	Total          int64            `json:"total"`
	UtilisationPct float64          `json:"utilisation_pct"`
	Children       []SubnetTreeNode `json:"children"`
}

// NotificationQueue is a durable outbox for outbound email notifications
type NotificationQueue struct {
	ID          int64      `json:"id"`
	UserID      int64      `json:"user_id"`
	Email       string     `json:"email"`
	Template    string     `json:"template"`
	Data        string     `json:"data"` // JSON string
	Status      string     `json:"status"`
	RetryCount  int        `json:"retry_count"`
	NextRetryAt *time.Time `json:"next_retry_at,omitempty"`
	SentAt      *time.Time `json:"sent_at,omitempty"`
	ErrorMsg    *string    `json:"error_msg,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
