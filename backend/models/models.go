package models

import "time"

// CustomFieldDefinition defines a custom field for an entity type
type CustomFieldDefinition struct {
	ID           int64       `json:"id"`
	EntityType   string      `json:"entity_type"`
	Name         string      `json:"name"`
	Label        string      `json:"label"`
	FieldType    string      `json:"field_type"`
	Options      interface{} `json:"options"`
	IsRequired   bool        `json:"is_required"`
	DefaultValue *string     `json:"default_value"`
	Placeholder  *string     `json:"placeholder"`
	DisplayOrder int         `json:"display_order"`
	IsSearchable bool        `json:"is_searchable"`
	CreatedAt    time.Time   `json:"created_at"`
}

// CustomFieldValue holds a value for a custom field on a specific entity
type CustomFieldValue struct {
	ID           int64   `json:"id"`
	DefinitionID int64   `json:"definition_id"`
	EntityID     int64   `json:"entity_id"`
	EntityType   string  `json:"entity_type"`
	Value        *string `json:"value"`
}

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
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedBy   int64     `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Nameserver represents a set of DNS nameservers that can be assigned to subnets.
type Nameserver struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Server1     string    `json:"server1"`
	Server2     *string   `json:"server2,omitempty"`
	Server3     *string   `json:"server3,omitempty"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Subnet represents a network subnet
type Subnet struct {
	ID               int64              `json:"id"`
	SectionID        int64              `json:"section_id"`
	NetworkAddress   string             `json:"network_address"`
	PrefixLength     int                `json:"prefix_length"`
	Description      string             `json:"description"`
	Gateway          *string            `json:"gateway,omitempty"`
	AutoReserveFirst bool               `json:"auto_reserve_first"`
	AutoReserveLast  bool               `json:"auto_reserve_last"`
	LocationID       *int64             `json:"location_id,omitempty"`
	NameserverID     *int64             `json:"nameserver_id,omitempty"`
	Nameserver       *Nameserver        `json:"nameserver,omitempty"`
	VLANID           *int64             `json:"vlan_id,omitempty"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
	CustomFields     map[string]*string `json:"custom_fields,omitempty"`
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
	ID             int64              `json:"id"`
	SubnetID       int64              `json:"subnet_id"`
	Address        string             `json:"address"`
	Hostname       string             `json:"hostname"`
	Status         string             `json:"status"`
	AssignedTo     *string            `json:"assigned_to,omitempty"`
	AssignedAt     *time.Time         `json:"assigned_at,omitempty"`
	ExpiresAt      *time.Time         `json:"expires_at,omitempty"`
	TagID          *int64             `json:"tag_id,omitempty"`
	Tag            *IPTag             `json:"tag,omitempty"`
	LastSeen       *time.Time         `json:"last_seen,omitempty"`
	MACAddress     *string            `json:"mac_address,omitempty"`
	PTRRecord      *string            `json:"ptr_record,omitempty"`
	DNSName        *string            `json:"dns_name,omitempty"`
	DNSRecords     *string            `json:"dns_records,omitempty"` // JSON string
	DNSLastChecked *time.Time         `json:"dns_last_checked,omitempty"`
	DeviceID       *int64             `json:"device_id,omitempty"`
	InterfaceName  *string            `json:"interface_name,omitempty"`
	IsPrimary      bool               `json:"is_primary"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
	CustomFields   map[string]*string `json:"custom_fields,omitempty"`
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
	ID         int64
	UserID     int64
	RoleID     int64
	LocationID *int64 `json:"location_id,omitempty"`
	CreatedAt  time.Time
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

// VLANDomain represents an L2 domain that groups VLANs
type VLANDomain struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// VLANGroup represents a named grouping/category of VLANs
type VLANGroup struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Colour      *string   `json:"colour,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// VLAN represents a Virtual LAN segment
type VLAN struct {
	ID          int64
	VRFID       *int64
	DomainID    *int64
	GroupID     *int64
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
	TotalSections          int64               `json:"total_sections"`
	TotalSubnets           int64               `json:"total_subnets"`
	TotalIPs               int64               `json:"total_ips"`
	UsedIPs                int64               `json:"used_ips"`
	UtilisationPct         float64             `json:"utilisation_pct"`
	TopSubnets             []SubnetUtilisation `json:"top_subnets"`
	PendingSubnetRequests  int64               `json:"pending_subnet_requests"`
	PendingIPRequests      int64               `json:"pending_ip_requests"`
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

// DeviceType represents a category of network device
type DeviceType struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Icon        string    `json:"icon"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Device represents a network device in the inventory
type Device struct {
	ID           int64              `json:"id"`
	Hostname     string             `json:"hostname"`
	Description  *string            `json:"description,omitempty"`
	TypeID       *int64             `json:"type_id,omitempty"`
	Type         *DeviceType        `json:"type,omitempty"`
	SectionID    *int64             `json:"section_id,omitempty"`
	Vendor       *string            `json:"vendor,omitempty"`
	Model        *string            `json:"model,omitempty"`
	OSVersion    *string            `json:"os_version,omitempty"`
	IsOnline     bool               `json:"is_online"`
	LastPingAt   *time.Time         `json:"last_ping_at,omitempty"`
	LocationID    *int64             `json:"location_id,omitempty"`
	RackID        *int64             `json:"rack_id,omitempty"`
	RackUnitStart *int               `json:"rack_unit_start,omitempty"`
	RackUnitSize  int                `json:"rack_unit_size"`
	IPCount       int                `json:"ip_count"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
	CustomFields  map[string]*string `json:"custom_fields,omitempty"`
	// NOTE: SNMP fields intentionally omitted — use DeviceSNMP for credentials endpoint
}

// DeviceSNMP holds the SNMP credentials for a device (only returned via privileged endpoint)
type DeviceSNMP struct {
	DeviceID        int64   `json:"device_id"`
	SNMPCommunity   *string `json:"snmp_community,omitempty"`
	SNMPVersion     string  `json:"snmp_version"`
	SNMPV3User      *string `json:"snmp_v3_user,omitempty"`
	SNMPV3AuthProto *string `json:"snmp_v3_auth_proto,omitempty"`
	SNMPV3AuthPass  *string `json:"snmp_v3_auth_pass,omitempty"`
	SNMPV3PrivProto *string `json:"snmp_v3_priv_proto,omitempty"`
	SNMPV3PrivPass  *string `json:"snmp_v3_priv_pass,omitempty"`
}

// Rack represents a physical equipment rack in a location
type Rack struct {
	ID          int64     `json:"id"`
	LocationID  *int64    `json:"location_id,omitempty"`
	Name        string    `json:"name"`
	SizeU       int       `json:"size_u"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Location represents a physical place in the location hierarchy (site, building, floor, room, cage, etc.)
type Location struct {
	ID          int64      `json:"id"`
	ParentID    *int64     `json:"parent_id,omitempty"`
	Name        string     `json:"name"`
	Type        string     `json:"type"`
	Address     *string    `json:"address,omitempty"`
	Lat         *float64   `json:"lat,omitempty"`
	Lng         *float64   `json:"lng,omitempty"`
	Description *string    `json:"description,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// LocationTreeNode is a Location with nested children for tree responses
type LocationTreeNode struct {
	Location
	Children []*LocationTreeNode `json:"children"`
}

// SubnetRequest represents a user request for a new subnet allocation
type SubnetRequest struct {
	ID                 int64     `json:"id"`
	RequesterID        int64     `json:"requester_id"`
	RequesterUsername  string    `json:"requester_username,omitempty"`
	SectionID          int64     `json:"section_id"`
	ParentSubnetID     *int64    `json:"parent_subnet_id,omitempty"`
	RequestedPrefixLen int       `json:"requested_prefix_len"`
	Purpose            string    `json:"purpose"`
	Status             string    `json:"status"`
	ReviewerID         *int64    `json:"reviewer_id,omitempty"`
	ReviewerUsername   string    `json:"reviewer_username,omitempty"`
	ReviewerNote       string    `json:"reviewer_note"`
	SubnetID           *int64    `json:"subnet_id,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// IPRequest represents a user request for an IP address assignment
type IPRequest struct {
	ID                int64     `json:"id"`
	RequesterID       int64     `json:"requester_id"`
	RequesterUsername string    `json:"requester_username,omitempty"`
	SubnetID          int64     `json:"subnet_id"`
	RequestedIP       *string   `json:"requested_ip,omitempty"`
	DNSName           string    `json:"dns_name"`
	Purpose           string    `json:"purpose"`
	Status            string    `json:"status"`
	ReviewerID        *int64    `json:"reviewer_id,omitempty"`
	ReviewerUsername  string    `json:"reviewer_username,omitempty"`
	ReviewerNote      string    `json:"reviewer_note"`
	IPAddressID       *int64    `json:"ip_address_id,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// RequestComment represents a comment on a subnet or IP request
type RequestComment struct {
	ID          int64     `json:"id"`
	RequestType string    `json:"request_type"`
	RequestID   int64     `json:"request_id"`
	AuthorID    int64     `json:"author_id"`
	AuthorUsername string `json:"author_username,omitempty"`
	Body        string    `json:"body"`
	CreatedAt   time.Time `json:"created_at"`
}

// DeviceInterface represents a network interface on a device
type DeviceInterface struct {
	ID                     int64   `json:"id"`
	DeviceID               int64   `json:"device_id"`
	Name                   string  `json:"name"`
	Description            *string `json:"description,omitempty"`
	SpeedMbps              *int    `json:"speed_mbps,omitempty"`
	MediaType              *string `json:"media_type,omitempty"`
	VLANID                 *int64  `json:"vlan_id,omitempty"`
	IPAddressID            *int64  `json:"ip_address_id,omitempty"`
	ConnectedToDeviceID    *int64  `json:"connected_to_device_id,omitempty"`
	ConnectedToInterfaceID *int64  `json:"connected_to_interface_id,omitempty"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}
