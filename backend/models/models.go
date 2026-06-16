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
	ExternalAuthProvider   *string
	ExternalAuthID         *string
	AvatarSource           string // "gravatar" or "custom"
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

// LDAPConfig holds the LDAP / Active Directory authentication configuration.
type LDAPConfig struct {
	ID              int64
	Enabled         bool
	Host            string
	Port            int
	BindDN          string
	BindPasswordEnc []byte // AES-GCM encrypted bytes; empty slice means no bind password
	BaseDN          string
	UserFilter      string
	UsernameAttr    string
	EmailAttr       string
	TLSMode         string // "none", "starttls", "tls"
	TLSSkipVerify   bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// LDAPGroupRoleMapping maps an LDAP group DN to a local RBAC role.
type LDAPGroupRoleMapping struct {
	ID          int64
	LDAPGroupDN string
	RoleID      int64
	CreatedAt   time.Time
}

// OAuth2Config holds the OAuth2 / OIDC authentication configuration.
type OAuth2Config struct {
	ID               int64
	Enabled          bool
	ProviderName     string
	ClientID         string
	ClientSecretEnc  []byte // AES-GCM encrypted bytes
	DiscoveryURL     string // OIDC auto-discovery endpoint
	AuthorizationURL string
	TokenURL         string
	UserinfoURL      string
	Scopes           string
	RedirectURI      string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// SAMLConfig holds the SAML 2.0 Service Provider configuration.
type SAMLConfig struct {
	ID             int64
	Enabled        bool
	IDPMetadataURL string
	IDPMetadataXML string
	SPCertPEM      string
	SPKeyPEM       string
	EntityID       string
	ACSURL         string
	NameIDFormat   string
	CreatedAt      time.Time
	UpdatedAt      time.Time
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
	Username          string // populated by join queries (e.g. ListAllActiveSessions)
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
	Username               string     `json:"username,omitempty"`
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

// Network represents an IP address section/grouping
type Network struct {
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
	ID                 int64              `json:"id"`
	NetworkID          int64              `json:"network_id"`
	NetworkAddress     string             `json:"network_address"`
	PrefixLength       int                `json:"prefix_length"`
	Description        string             `json:"description"`
	Gateway            *string            `json:"gateway,omitempty"`
	AutoReserveFirst   bool               `json:"auto_reserve_first"`
	AutoReserveLast    bool               `json:"auto_reserve_last"`
	LocationID         *int64             `json:"location_id,omitempty"`
	NameserverID       *int64             `json:"nameserver_id,omitempty"`
	Nameserver         *Nameserver        `json:"nameserver,omitempty"`
	VLANID             *int64             `json:"vlan_id,omitempty"`
	ParentSubnetID     *int64             `json:"parent_subnet_id,omitempty"`
	IsContainer        bool               `json:"is_container"`
	AlertThresholdPct  *int               `json:"alert_threshold_pct,omitempty"`
	AlertEmailOverride *string            `json:"alert_email_override,omitempty"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
	CustomFields       map[string]*string `json:"custom_fields,omitempty"`
	ScanProfileID        *int64             `json:"scan_profile_id,omitempty"`
	TechnitiumScopeName  string             `json:"technitium_scope_name"`
}

// IPv6Delegation represents a delegated IPv6 prefix assigned to a device or description
type IPv6Delegation struct {
	ID                     int64      `json:"id"`
	ParentSubnetID         int64      `json:"parent_subnet_id"`
	DelegatedPrefix        string     `json:"delegated_prefix"`
	DelegatedToDeviceID    *int64     `json:"delegated_to_device_id,omitempty"`
	DelegatedToDescription *string    `json:"delegated_to_description,omitempty"`
	ValidLifetimeSec       *int       `json:"valid_lifetime_sec,omitempty"`
	PreferredLifetimeSec   *int       `json:"preferred_lifetime_sec,omitempty"`
	ExpiresAt              *time.Time `json:"expires_at,omitempty"`
	IsExpired              bool       `json:"is_expired"`
	CreatedAt              time.Time  `json:"created_at"`
}

// TopologyNode represents a subnet node in the network topology graph
type TopologyNode struct {
	ID          int64   `json:"id"`
	Label       string  `json:"label"`
	CIDR        string  `json:"cidr"`
	PrefixLen   int     `json:"prefix_len"`
	IsContainer bool    `json:"is_container"`
	ParentID    *int64  `json:"parent_id,omitempty"`
	VLANID      *int64  `json:"vlan_id,omitempty"`
	Utilization float64 `json:"utilization"`
}

// TopologyEdge represents a directed edge in the network topology graph
type TopologyEdge struct {
	Source int64  `json:"source"`
	Target int64  `json:"target"`
	Type   string `json:"type"` // "parent_child" or "subnet_vlan"
}

// NetworkTopology holds the full topology graph for a section
type NetworkTopology struct {
	Nodes []*TopologyNode `json:"nodes"`
	Edges []*TopologyEdge `json:"edges"`
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

// DeviceSummary is a minimal device representation embedded in IP address responses.
type DeviceSummary struct {
	ID       int64  `json:"id"`
	Hostname string `json:"hostname"`
}

// IPAddress represents an individual IP address
type IPAddress struct {
	ID             int64              `json:"id"`
	SubnetID       int64              `json:"subnet_id"`
	Address        string             `json:"address"`
	Hostname       string             `json:"hostname"`
	Status         string             `json:"status"`
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
	Device         *DeviceSummary     `json:"device,omitempty"`
	InterfaceName  *string            `json:"interface_name,omitempty"`
	IsPrimary      bool               `json:"is_primary"`
	PortOpen       map[string]bool    `json:"port_open,omitempty"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
	CustomFields   map[string]*string `json:"custom_fields,omitempty"`
	Virtual        bool               `json:"virtual,omitempty"`
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
	ID          int64     `json:"id"`
	VRFID       *int64    `json:"vrf_id,omitempty"`
	DomainID    *int64    `json:"domain_id,omitempty"`
	GroupID     *int64    `json:"group_id,omitempty"`
	VlanID      int       `json:"vlan_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UserMFASettings tracks MFA status for a user
type UserMFASettings struct {
	ID                     int64
	UserID                 int64
	TOTPEnabled            bool
	BackupCodesGeneratedAt *time.Time
	CreatedAt              time.Time
	UpdatedAt              time.Time
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
	ResourceID   *int64
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
	ID              int64      `json:"id"`
	Name            string     `json:"name"`
	SubnetIDs       []int64    `json:"subnet_ids"`
	ScheduleCron    *string    `json:"schedule_cron,omitempty"`
	IsActive        bool       `json:"is_active"`
	LastRunAt       *time.Time `json:"last_run_at,omitempty"`
	NextRunAt       *time.Time `json:"next_run_at,omitempty"`
	CreatedBy       int64      `json:"created_by"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	PingConcurrency int        `json:"ping_concurrency"`
	NotifyOnChange  bool       `json:"notify_on_change"`
	ScanType        string     `json:"scan_type"`
	AgentID         *int64     `json:"agent_id,omitempty"`
	AutoAddIPs      bool       `json:"auto_add_ips"`
	DiscoverDNS     bool       `json:"discover_dns"`
	DNSOverwrite    bool       `json:"dns_overwrite"`
	// Transient: populated from scan profile at runtime, not stored in DB.
	SNMPCommunityOverride string `json:"-"`
	SNMPVersionOverride   string `json:"-"`
}

// ScanRun records a single execution of a scan job
type ScanRun struct {
	ID           int64      `json:"id"`
	ScanJobID    int64      `json:"scan_job_id"`
	StartedAt    time.Time  `json:"started_at"`
	FinishedAt   *time.Time `json:"finished_at,omitempty"`
	NewCount     int        `json:"new_count"`
	GoneCount    int        `json:"gone_count"`
	ChangedCount int        `json:"changed_count"`
}

// ScanRunChange records a single IP change detected during a scan run
type ScanRunChange struct {
	ID         int64     `json:"id"`
	RunID      int64     `json:"run_id"`
	IPAddress  string    `json:"ip_address"`
	ChangeType string    `json:"change_type"`
	ScannedAt  time.Time `json:"scanned_at"`
}

// ScanAgent represents a remote scanning agent
type ScanAgent struct {
	ID           int64      `json:"id"`
	Name         string     `json:"name"`
	TokenHash    string     `json:"-"`
	LastSeen     *time.Time `json:"last_seen,omitempty"`
	IsActive     bool       `json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	Version      *string    `json:"version,omitempty"`
	Capabilities []string   `json:"capabilities,omitempty"`
	Status       string     `json:"status"`
	LastError    *string    `json:"last_error,omitempty"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
}

// ScanProfile holds reusable scan configuration that can be referenced by scan jobs or subnets.
type ScanProfile struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	Description     *string   `json:"description,omitempty"`
	ScanType        string    `json:"scan_type"`
	PingConcurrency int       `json:"ping_concurrency"`
	TCPPorts        *string   `json:"tcp_ports,omitempty"`
	DNSLookup       bool      `json:"dns_lookup"`
	SNMPCommunity   *string   `json:"snmp_community,omitempty"`
	SNMPVersion     string    `json:"snmp_version"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
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
	PTRRecord      *string   `json:"ptr_record,omitempty"`
	FwdRevMismatch bool      `json:"fwd_rev_mismatch"`
	ScannedAt      time.Time `json:"scanned_at"`
}

// DashboardSummary holds aggregate IPAM statistics for the dashboard
type DashboardSummary struct {
	TotalNetworks         int64               `json:"total_networks"`
	TotalSubnets          int64               `json:"total_subnets"`
	TotalIPs              int64               `json:"total_ips"`
	UsedIPs               int64               `json:"used_ips"`
	UtilizationPct        float64             `json:"utilization_pct"`
	TopSubnets            []SubnetUtilization `json:"top_subnets"`
	PendingSubnetRequests int64               `json:"pending_subnet_requests"`
	PendingIPRequests     int64               `json:"pending_ip_requests"`
}

// SubnetUtilization holds utilization data for a single subnet
type SubnetUtilization struct {
	ID             int64   `json:"id"`
	CIDR           string  `json:"cidr"`
	Description    string  `json:"description"`
	Used           int64   `json:"used"`
	Total          int64   `json:"total"`
	UtilizationPct float64 `json:"utilization_pct"`
}

// DashboardActivity is a single recent activity entry from audit_logs
type DashboardActivity struct {
	ID          int64  `json:"id"`
	Action      string `json:"action"`
	EntityType  string `json:"entity_type"`
	EntityID    *int64 `json:"entity_id,omitempty"`
	UserID      *int64 `json:"user_id,omitempty"`
	Username    string `json:"username"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
}

// SubnetTreeNode represents a subnet in the tree hierarchy view
type SubnetTreeNode struct {
	ID             int64            `json:"id"`
	CIDR           string           `json:"cidr"`
	Description    string           `json:"description"`
	Used           int64            `json:"used"`
	Total          int64            `json:"total"`
	UtilizationPct float64          `json:"utilization_pct"`
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

// WebhookEndpoint configures an outbound webhook receiver.
type WebhookEndpoint struct {
	ID               int64             `json:"id"`
	Name             string            `json:"name"`
	URL              string            `json:"url"`
	Secret           string            `json:"-"`
	Events           []string          `json:"events"`
	ObjectTypes      []string          `json:"object_types"`
	TagFilters       []string          `json:"tag_filters"`
	FilterConditions map[string]string `json:"filter_conditions,omitempty"`
	IsActive         bool              `json:"is_active"`
	CreatedBy        *int64            `json:"created_by,omitempty"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}

// WebhookDelivery is one queued outbound webhook delivery attempt.
type WebhookDelivery struct {
	ID             int64      `json:"id"`
	EndpointID     int64      `json:"endpoint_id"`
	EventType      string     `json:"event_type"`
	Payload        string     `json:"payload"`
	Status         string     `json:"status"`
	RetryCount     int        `json:"retry_count"`
	NextRetryAt    *time.Time `json:"next_retry_at,omitempty"`
	DeliveredAt    *time.Time `json:"delivered_at,omitempty"`
	ResponseStatus *int       `json:"response_status,omitempty"`
	ErrorMsg       *string    `json:"error_msg,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// WebhookFailureGroup summarizes failed/retrying deliveries that share a cause.
type WebhookFailureGroup struct {
	EndpointID     int64      `json:"endpoint_id"`
	EventType      string     `json:"event_type"`
	Status         string     `json:"status"`
	ErrorMsg       string     `json:"error_msg"`
	Count          int64      `json:"count"`
	LastOccurredAt time.Time  `json:"last_occurred_at"`
	LastDeliveryID int64      `json:"last_delivery_id"`
	ResponseStatus *int       `json:"response_status,omitempty"`
	NextRetryAt    *time.Time `json:"next_retry_at,omitempty"`
}

// AutomationPolicy is a lightweight rule evaluated before automation writes.
type AutomationPolicy struct {
	ID         int64             `json:"id"`
	Name       string            `json:"name"`
	Workflow   string            `json:"workflow"`
	Action     string            `json:"action"`
	Effect     string            `json:"effect"`
	Enabled    bool              `json:"enabled"`
	Conditions map[string]string `json:"conditions,omitempty"`
	Message    string            `json:"message,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

// IntegrationTemplate documents an automation platform recipe.
type IntegrationTemplate struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Category    string   `json:"category"`
	Description string   `json:"description"`
	Steps       []string `json:"steps"`
	Endpoints   []string `json:"endpoints"`
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
	ID            int64              `json:"id"`
	Hostname      string             `json:"hostname"`
	Description   *string            `json:"description,omitempty"`
	TypeID        *int64             `json:"type_id,omitempty"`
	Type          *DeviceType        `json:"type,omitempty"`
	NetworkID     *int64             `json:"network_id,omitempty"`
	Vendor        *string            `json:"vendor,omitempty"`
	Model         *string            `json:"model,omitempty"`
	OSVersion     *string            `json:"os_version,omitempty"`
	IsOnline      bool               `json:"is_online"`
	LastPingAt    *time.Time         `json:"last_ping_at,omitempty"`
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

// DeviceFingerprint holds automatically derived identity data for a device.
type DeviceFingerprint struct {
	ID              int64     `json:"id"`
	DeviceID        int64     `json:"device_id"`
	OpenPorts       []int     `json:"open_ports"`
	OSGuess         *string   `json:"os_guess,omitempty"`
	VendorGuess     *string   `json:"vendor_guess,omitempty"`
	ConfidenceScore float64   `json:"confidence_score"`
	Evidence        []string  `json:"evidence"`
	LastUpdatedAt   time.Time `json:"last_updated_at"`
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
	ID           int64     `json:"id"`
	ParentID     *int64    `json:"parent_id,omitempty"`
	Name         string    `json:"name"`
	Type         string    `json:"type"`
	Address      *string   `json:"address,omitempty"`
	City         *string   `json:"city,omitempty"`
	Region       *string   `json:"region,omitempty"`
	Country      *string   `json:"country,omitempty"`
	FacilityCode *string   `json:"facility_code,omitempty"`
	TimeZone     *string   `json:"time_zone,omitempty"`
	ContactName  *string   `json:"contact_name,omitempty"`
	ContactEmail *string   `json:"contact_email,omitempty"`
	ContactPhone *string   `json:"contact_phone,omitempty"`
	Status       string    `json:"status"`
	Lat          *float64  `json:"lat,omitempty"`
	Lng          *float64  `json:"lng,omitempty"`
	Description  *string   `json:"description,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
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
	NetworkID          int64     `json:"network_id"`
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
	ID             int64     `json:"id"`
	RequestType    string    `json:"request_type"`
	RequestID      int64     `json:"request_id"`
	AuthorID       int64     `json:"author_id"`
	AuthorUsername string    `json:"author_username,omitempty"`
	Body           string    `json:"body"`
	CreatedAt      time.Time `json:"created_at"`
}

// VLANUsageEntry holds per-VLAN metrics for the usage report
type VLANUsageEntry struct {
	VLANID         int64   `json:"vlan_id"`
	VLANName       string  `json:"vlan_name"`
	VLANTag        int     `json:"vlan_tag"`
	SubnetCount    int64   `json:"subnet_count"`
	IPCount        int64   `json:"ip_count"`
	TotalIPs       int64   `json:"total_ips"`
	UtilizationPct float64 `json:"utilization_pct"`
}

// VLANUsageReport is the top-level report returned by the usage-report endpoint
type VLANUsageReport struct {
	Entries     []*VLANUsageEntry `json:"entries"`
	GeneratedAt string            `json:"generated_at"`
}

// DeviceInterface represents a network interface on a device
type DeviceInterface struct {
	ID                     int64     `json:"id"`
	DeviceID               int64     `json:"device_id"`
	Name                   string    `json:"name"`
	Description            *string   `json:"description,omitempty"`
	SpeedMbps              *int      `json:"speed_mbps,omitempty"`
	MediaType              *string   `json:"media_type,omitempty"`
	VLANID                 *int64    `json:"vlan_id,omitempty"`
	IPAddressID            *int64    `json:"ip_address_id,omitempty"`
	ConnectedToDeviceID    *int64    `json:"connected_to_device_id,omitempty"`
	ConnectedToInterfaceID *int64    `json:"connected_to_interface_id,omitempty"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

// SubnetUtilizationPoint is a single time-series data point for subnet utilization history.
type SubnetUtilizationPoint struct {
	RecordedAt     time.Time `json:"recorded_at"`
	UsedCount      int       `json:"used_count"`
	TotalCount     int       `json:"total_count"`
	UtilizationPct float64   `json:"utilization_pct"`
}

// SubnetUtilizationTrend holds current utilization and the delta vs 7 days ago for a subnet.
type SubnetUtilizationTrend struct {
	SubnetID    int64   `json:"subnet_id"`
	CIDR        string  `json:"cidr"`
	Description string  `json:"description"`
	CurrentPct  float64 `json:"current_pct"`
	WeekAgoPct  float64 `json:"week_ago_pct"`
	DeltaPct    float64 `json:"delta_pct"`
}

// AlertCooldown tracks when a threshold alert was last sent for a subnet.
type AlertCooldown struct {
	ID         int64     `json:"id"`
	SubnetID   int64     `json:"subnet_id"`
	AlertedAt  time.Time `json:"alerted_at"`
	AlertedPct float64   `json:"alerted_pct"`
}

// ScheduledReport defines a recurring report that is emailed to recipients.
type ScheduledReport struct {
	ID              int64          `json:"id"`
	Name            string         `json:"name"`
	ReportType      string         `json:"report_type"`
	ScheduleCron    string         `json:"schedule_cron"`
	RecipientEmails []string       `json:"recipient_emails"`
	Filters         map[string]any `json:"filters"`
	Format          string         `json:"format"`
	LastRunAt       *time.Time     `json:"last_run_at,omitempty"`
	CreatedBy       int64          `json:"created_by"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

// InactiveIPReport describes an IP address that has not been seen recently.
type InactiveIPReport struct {
	IPID         int64      `json:"ip_id"`
	IPAddress    string     `json:"ip_address"`
	Hostname     string     `json:"hostname"`
	SubnetCIDR   string     `json:"subnet_cidr"`
	NetworkName  string     `json:"section_name"`
	DeviceID     *int64     `json:"device_id,omitempty"`
	LastSeen     *time.Time `json:"last_seen,omitempty"`
	DaysInactive int        `json:"days_inactive"`
}

// InactiveDeviceReport describes a device that has not been pinged recently.
type InactiveDeviceReport struct {
	DeviceID     int64      `json:"device_id"`
	Hostname     string     `json:"hostname"`
	Vendor       string     `json:"vendor"`
	Model        string     `json:"model"`
	LastPingAt   *time.Time `json:"last_ping_at,omitempty"`
	DaysInactive int        `json:"days_inactive"`
}

// FailedScanJobReport describes a scan job that has never run or whose last run is overdue.
type FailedScanJobReport struct {
	JobID        int64      `json:"job_id"`
	JobName      string     `json:"job_name"`
	ScheduleCron string     `json:"schedule_cron"`
	LastRunAt    *time.Time `json:"last_run_at,omitempty"`
	DaysSinceRun int        `json:"days_since_run"`
	IsActive     bool       `json:"is_active"`
}

// Customer represents an organisation that owns or uses IP space.
type Customer struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
	Notes       string    `json:"notes"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NATRule represents a first-class NAT mapping between internal and external addresses.
type NATRule struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Type         string    `json:"type"`
	InternalCIDR string    `json:"internal_cidr"`
	ExternalCIDR string    `json:"external_cidr"`
	Protocol     string    `json:"protocol"`
	InternalPort *int      `json:"internal_port,omitempty"`
	ExternalPort *int      `json:"external_port,omitempty"`
	DeviceID     *int64    `json:"device_id,omitempty"`
	CustomerID   *int64    `json:"customer_id,omitempty"`
	Description  string    `json:"description"`
	Status       string    `json:"status"`
	CustomerName *string   `json:"customer_name,omitempty"`
	DeviceName   *string   `json:"device_name,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// DHCPServer tracks DHCP service ownership and operational metadata.
type DHCPServer struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Address      string    `json:"address"`
	Vendor       string    `json:"vendor"`
	Version      string    `json:"version"`
	LocationID   *int64    `json:"location_id,omitempty"`
	Description  string    `json:"description"`
	Status       string    `json:"status"`
	LocationName *string   `json:"location_name,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// DHCPLease records a DHCP lease observed from, or managed by, a DHCP server.
type DHCPLease struct {
	ID           int64      `json:"id"`
	ServerID     int64      `json:"server_id"`
	IPAddress    string     `json:"ip_address"`
	MACAddress   string     `json:"mac_address"`
	Hostname     string     `json:"hostname"`
	SubnetID     *int64     `json:"subnet_id,omitempty"`
	IPID         *int64     `json:"ip_id,omitempty"`
	CustomerID   *int64     `json:"customer_id,omitempty"`
	StartsAt     *time.Time `json:"starts_at,omitempty"`
	EndsAt       *time.Time `json:"ends_at,omitempty"`
	State        string     `json:"state"`
	ServerName   *string    `json:"server_name,omitempty"`
	CustomerName *string    `json:"customer_name,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// CircuitProvider represents a carrier or internal circuit provider.
type CircuitProvider struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	AccountNo    string    `json:"account_no"`
	SupportEmail string    `json:"support_email"`
	SupportPhone string    `json:"support_phone"`
	PortalURL    string    `json:"portal_url"`
	Notes        string    `json:"notes"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// PhysicalCircuit represents an installed physical circuit.
type PhysicalCircuit struct {
	ID            int64      `json:"id"`
	ProviderID    int64      `json:"provider_id"`
	CircuitID     string     `json:"circuit_id"`
	Name          string     `json:"name"`
	Type          string     `json:"type"`
	Status        string     `json:"status"`
	BandwidthMbps *int       `json:"bandwidth_mbps,omitempty"`
	LocationAID   *int64     `json:"location_a_id,omitempty"`
	LocationBID   *int64     `json:"location_b_id,omitempty"`
	CustomerID    *int64     `json:"customer_id,omitempty"`
	InstallDate   *time.Time `json:"install_date,omitempty"`
	ProviderName  *string    `json:"provider_name,omitempty"`
	LocationAName *string    `json:"location_a_name,omitempty"`
	LocationBName *string    `json:"location_b_name,omitempty"`
	CustomerName  *string    `json:"customer_name,omitempty"`
	Notes         string     `json:"notes"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// LogicalCircuit represents an overlay or service carried over physical circuits.
type LogicalCircuit struct {
	ID                  int64     `json:"id"`
	PhysicalCircuitID   *int64    `json:"physical_circuit_id,omitempty"`
	Name                string    `json:"name"`
	ServiceID           string    `json:"service_id"`
	Type                string    `json:"type"`
	Status              string    `json:"status"`
	VLANID              *int64    `json:"vlan_id,omitempty"`
	VRFID               *int64    `json:"vrf_id,omitempty"`
	CustomerID          *int64    `json:"customer_id,omitempty"`
	BandwidthMbps       *int      `json:"bandwidth_mbps,omitempty"`
	PhysicalCircuitName *string   `json:"physical_circuit_name,omitempty"`
	CustomerName        *string   `json:"customer_name,omitempty"`
	Notes               string    `json:"notes"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// CustomerAssociation links a customer to another IPAM object.
type CustomerAssociation struct {
	ID           int64     `json:"id"`
	CustomerID   int64     `json:"customer_id"`
	ObjectType   string    `json:"object_type"`
	ObjectID     int64     `json:"object_id"`
	ObjectName   string    `json:"object_name"`
	Relationship string    `json:"relationship"`
	Notes        string    `json:"notes"`
	CustomerName *string   `json:"customer_name,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// FirewallZone represents a documented network security zone.
type FirewallZone struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Color       string    `json:"color"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// FirewallZoneMapping links a security zone to an IPAM object or CIDR.
type FirewallZoneMapping struct {
	ID          int64     `json:"id"`
	ZoneID      int64     `json:"zone_id"`
	ObjectType  string    `json:"object_type"`
	ObjectID    *int64    `json:"object_id,omitempty"`
	CIDR        string    `json:"cidr"`
	Direction   string    `json:"direction"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	ZoneName    *string   `json:"zone_name,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// DuplicateHostname describes a hostname that appears on more than one device.
type DuplicateHostname struct {
	Hostname  string  `json:"hostname"`
	Count     int     `json:"count"`
	DeviceIDs []int64 `json:"device_ids"`
}

// ConflictingIP describes an IP address that is assigned to more than one entry.
type ConflictingIP struct {
	IPAddress  string   `json:"ip_address"`
	SubnetCIDR string   `json:"subnet_cidr"`
	Count      int      `json:"count"`
	Hostnames  []string `json:"hostnames"`
}

// DuplicatesReport is the aggregate response for GET /admin/reports/duplicates.
type DuplicatesReport struct {
	DuplicateHostnames []DuplicateHostname `json:"duplicate_hostnames"`
	ConflictingIPs     []ConflictingIP     `json:"conflicting_ips"`
}

// AutonomousSystem represents a BGP Autonomous System entry.
type AutonomousSystem struct {
	ID          int64     `json:"id"`
	ASN         int64     `json:"asn"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	RIR         string    `json:"rir"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ScanRetentionSettings controls how long raw scan history is kept.
type ScanRetentionSettings struct {
	ID              int64     `json:"id"`
	RawHistoryDays  int       `json:"raw_history_days"`
	RollupEnabled   bool      `json:"rollup_enabled"`
	RollupAfterDays int       `json:"rollup_after_days"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// DiscoveryConflict represents a pending conflict between discovered and manually-entered data.
type DiscoveryConflict struct {
	ID              int64      `json:"id"`
	DeviceID        int64      `json:"device_id"`
	FieldName       string     `json:"field_name"`
	CurrentValue    *string    `json:"current_value,omitempty"`
	DiscoveredValue string     `json:"discovered_value"`
	ConfidenceScore float64    `json:"confidence_score"`
	Source          string     `json:"source"`
	Status          string     `json:"status"`
	ReviewedBy      *string    `json:"reviewed_by,omitempty"`
	ReviewedAt      *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

// TopologyHint represents a suggested relationship between two inventory entities,
// derived from discovery or inventory data.
type TopologyHint struct {
	ID              int64     `json:"id"`
	SourceType      string    `json:"source_type"`
	SourceID        int64     `json:"source_id"`
	TargetType      string    `json:"target_type"`
	TargetID        int64     `json:"target_id"`
	HintType        string    `json:"hint_type"`
	ConfidenceScore float64   `json:"confidence_score"`
	Evidence        *string   `json:"evidence,omitempty"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// PrivacyPolicyVersion represents a versioned privacy policy record.
type PrivacyPolicyVersion struct {
	ID            int64     `json:"id"`
	Version       string    `json:"version"`
	EffectiveDate string    `json:"effective_date"`
	Summary       *string   `json:"summary,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// UserConsentStatus holds privacy consent info for a single user (admin report).
type UserConsentStatus struct {
	UserID             int64      `json:"user_id"`
	Username           string     `json:"username"`
	Email              string     `json:"email"`
	PrivacyAcceptedAt  *time.Time `json:"privacy_accepted_at,omitempty"`
	PrivacyAcceptedVer *string    `json:"privacy_accepted_version,omitempty"`
	HasConsent         bool       `json:"has_consent"`
}

// BreakGlassSession represents an emergency admin break-glass access session.
type BreakGlassSession struct {
	ID                int64      `json:"id"`
	InitiatedByUserID int64      `json:"initiated_by_user_id"`
	Justification     string     `json:"justification"`
	ExpiresAt         time.Time  `json:"expires_at"`
	EndedAt           *time.Time `json:"ended_at,omitempty"`
	EndedByUserID     *int64     `json:"ended_by_user_id,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	IsActive          bool       `json:"is_active"`
}

// AuditRetentionSettings controls how long audit logs are retained.
type AuditRetentionSettings struct {
	ID             int64     `json:"id"`
	RetentionDays  int       `json:"retention_days"`
	ArchiveEnabled bool      `json:"archive_enabled"`
	UpdatedAt      time.Time `json:"updated_at"`
}
