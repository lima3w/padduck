package services

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/mail"
	"strconv"
	"strings"
	"time"

	_ "golang.org/x/image/webp"

	"padduck/models"
	"padduck/repository"
	"padduck/utils"
)

// IdentityService handles users, sessions, API tokens, RBAC, security, and Grafana datasource.
type IdentityService struct {
	repo         *repository.Repository
	config       *ConfigService
	email        *EmailService
	mfa          *MFAService
	notification *NotificationService
}

func NewIdentityService(repo *repository.Repository, config *ConfigService, email *EmailService, mfa *MFAService, notification *NotificationService) *IdentityService {
	return &IdentityService{
		repo:         repo,
		config:       config,
		email:        email,
		mfa:          mfa,
		notification: notification,
	}
}

// ---- Constants and errors (security) ----

var (
	ErrAccountLocked      = errors.New("account is temporarily locked due to too many failed login attempts")
	ErrInvalidUnlockToken = errors.New("unlock token is invalid or expired")
)

const (
	maxFailedAttempts  = 5
	bruteForceWindow   = 15 * time.Minute
	notifRateLimit     = 1 * time.Hour
	ipFailureThreshold = 3

	ipThrottleThreshold = 20
	ipThrottleWindow    = 15 * time.Minute
)

// ---- Constants (session) ----

const (
	DefaultIdleTimeoutMinutes   = 60
	DefaultAbsoluteTimeoutHours = 168 // 7 days
	sessionTokenLength          = 32
	DefaultSessionTimeout       = 24 * time.Hour
)

// ---- Constants (auth) ----

const TokenLength = 32

// ---- Constants (users) ----

const maxAvatarDimension = 4096

// ---- Types (auth) ----

// AuthResult is returned by AuthenticateUser; either the full user is set or MFAChallenge is set.
type AuthResult struct {
	User         *models.User
	MFARequired  bool
	MFAChallenge string
}

// ---- Types (users) ----

type BulkUserImportRecord struct {
	Username string
	Email    string
	Role     string
}

type BulkImportResult struct {
	Username     string `json:"username"`
	UserID       int64  `json:"user_id,omitempty"`
	TempPassword string `json:"temp_password,omitempty"`
	Error        string `json:"error,omitempty"`
}

// ---- Permission constants (rbac) ----

const (
	RoleAdmin  = "admin"
	RoleUser   = "user"
	RoleViewer = "viewer"
)

var ValidRoles = []string{RoleAdmin, RoleUser, RoleViewer}

const (
	PermNetworkCreate = "section:create"
	PermNetworkRead   = "section:read"
	PermNetworkUpdate = "section:update"
	PermNetworkDelete = "section:delete"
	PermSubnetCreate  = "subnet:create"
	PermSubnetRead    = "subnet:read"
	PermSubnetUpdate  = "subnet:update"
	PermSubnetDelete  = "subnet:delete"
	PermIPCreate      = "ip:create"
	PermIPRead        = "ip:read"
	PermIPUpdate      = "ip:update"
	PermIPDelete      = "ip:delete"
	PermIPAssign      = "ip:assign"
	PermIPRelease     = "ip:release"
	PermTokenCreate   = "token:create"
	PermTokenRead     = "token:read"
	PermTokenDelete   = "token:delete"
)

const (
	PermV2NetworkList   = "ipam:section:list"
	PermV2NetworkRead   = "ipam:section:read"
	PermV2NetworkWrite  = "ipam:section:write"
	PermV2NetworkDelete = "ipam:section:delete"

	PermV2SubnetList   = "ipam:subnet:list"
	PermV2SubnetRead   = "ipam:subnet:read"
	PermV2SubnetWrite  = "ipam:subnet:write"
	PermV2SubnetDelete = "ipam:subnet:delete"

	PermV2IPList    = "ipam:ip_address:list"
	PermV2IPRead    = "ipam:ip_address:read"
	PermV2IPAssign  = "ipam:ip_address:assign"
	PermV2IPRelease = "ipam:ip_address:release"

	PermV2VRFList   = "ipam:vrf:list"
	PermV2VRFRead   = "ipam:vrf:read"
	PermV2VRFWrite  = "ipam:vrf:write"
	PermV2VRFDelete = "ipam:vrf:delete"

	PermV2VLANList   = "ipam:vlan:list"
	PermV2VLANRead   = "ipam:vlan:read"
	PermV2VLANWrite  = "ipam:vlan:write"
	PermV2VLANDelete = "ipam:vlan:delete"

	PermV2UserList  = "auth:user:list"
	PermV2UserRead  = "auth:user:read"
	PermV2UserWrite = "auth:user:write"
	PermV2AuditRead = "auth:audit:read"

	PermV2DeviceRead   = "devices:read"
	PermV2DeviceWrite  = "devices:write"
	PermV2DeviceDelete = "devices:delete"
	PermV2DeviceAdmin  = "devices:admin"

	PermV2LocationList   = "ipam:location:list"
	PermV2LocationRead   = "ipam:location:read"
	PermV2LocationWrite  = "ipam:location:write"
	PermV2LocationDelete = "ipam:location:delete"

	PermV2NameserverList   = "ipam:nameserver:list"
	PermV2NameserverRead   = "ipam:nameserver:read"
	PermV2NameserverWrite  = "ipam:nameserver:write"
	PermV2NameserverDelete = "ipam:nameserver:delete"

	PermV2SubnetRequestSubmit = "ipam:subnet_request:submit"
	PermV2SubnetRequestReview = "ipam:subnet_request:review"

	PermV2VLANDomainList   = "ipam:vlan_domain:list"
	PermV2VLANDomainRead   = "ipam:vlan_domain:read"
	PermV2VLANDomainWrite  = "ipam:vlan_domain:write"
	PermV2VLANDomainDelete = "ipam:vlan_domain:delete"

	PermV2VLANGroupList   = "ipam:vlan_group:list"
	PermV2VLANGroupRead   = "ipam:vlan_group:read"
	PermV2VLANGroupWrite  = "ipam:vlan_group:write" // #nosec G101 -- permission string, not a credential.
	PermV2VLANGroupDelete = "ipam:vlan_group:delete"

	PermV2AdminRead  = "auth:admin:read"
	PermV2AdminWrite = "auth:admin:write"

	PermV2CustomerList   = "ipam:customer:list"
	PermV2CustomerRead   = "ipam:customer:read"
	PermV2CustomerWrite  = "ipam:customer:write"
	PermV2CustomerDelete = "ipam:customer:delete"

	PermV2ASList   = "ipam:autonomous_system:list"
	PermV2ASRead   = "ipam:autonomous_system:read"
	PermV2ASWrite  = "ipam:autonomous_system:write"
	PermV2ASDelete = "ipam:autonomous_system:delete"

	PermV2NATList   = "ipam:nat:list"
	PermV2NATRead   = "ipam:nat:read"
	PermV2NATWrite  = "ipam:nat:write"
	PermV2NATDelete = "ipam:nat:delete"

	PermV2DHCPList   = "ipam:dhcp:list"
	PermV2DHCPRead   = "ipam:dhcp:read"
	PermV2DHCPWrite  = "ipam:dhcp:write" // #nosec G101 -- permission string, not a credential.
	PermV2DHCPDelete = "ipam:dhcp:delete"

	PermV2CircuitList   = "ipam:circuit:list"
	PermV2CircuitRead   = "ipam:circuit:read"
	PermV2CircuitWrite  = "ipam:circuit:write"
	PermV2CircuitDelete = "ipam:circuit:delete"

	PermV2FirewallList   = "ipam:firewall:list"
	PermV2FirewallRead   = "ipam:firewall:read"
	PermV2FirewallWrite  = "ipam:firewall:write"
	PermV2FirewallDelete = "ipam:firewall:delete"

	PermV2OrgRead      = "auth:org:read"
	PermV2OrgWrite     = "auth:org:write"
	PermV2PlatformAdmin = "auth:platform:admin"
)

var AllPermissions = []string{
	PermV2NetworkList, PermV2NetworkRead, PermV2NetworkWrite, PermV2NetworkDelete,
	PermV2SubnetList, PermV2SubnetRead, PermV2SubnetWrite, PermV2SubnetDelete,
	PermV2IPList, PermV2IPRead, PermV2IPAssign, PermV2IPRelease,
	PermV2VRFList, PermV2VRFRead, PermV2VRFWrite, PermV2VRFDelete,
	PermV2VLANList, PermV2VLANRead, PermV2VLANWrite, PermV2VLANDelete,
	PermV2UserList, PermV2UserRead, PermV2UserWrite, PermV2AuditRead,
	PermV2DeviceRead, PermV2DeviceWrite, PermV2DeviceDelete, PermV2DeviceAdmin,
	PermV2LocationList, PermV2LocationRead, PermV2LocationWrite, PermV2LocationDelete,
	PermV2NameserverList, PermV2NameserverRead, PermV2NameserverWrite, PermV2NameserverDelete,
	PermV2SubnetRequestSubmit, PermV2SubnetRequestReview,
	PermV2VLANDomainList, PermV2VLANDomainRead, PermV2VLANDomainWrite, PermV2VLANDomainDelete,
	PermV2VLANGroupList, PermV2VLANGroupRead, PermV2VLANGroupWrite, PermV2VLANGroupDelete,
	PermV2AdminRead, PermV2AdminWrite,
	PermV2CustomerList, PermV2CustomerRead, PermV2CustomerWrite, PermV2CustomerDelete,
	PermV2ASList, PermV2ASRead, PermV2ASWrite, PermV2ASDelete,
	PermV2NATList, PermV2NATRead, PermV2NATWrite, PermV2NATDelete,
	PermV2DHCPList, PermV2DHCPRead, PermV2DHCPWrite, PermV2DHCPDelete,
	PermV2CircuitList, PermV2CircuitRead, PermV2CircuitWrite, PermV2CircuitDelete,
	PermV2FirewallList, PermV2FirewallRead, PermV2FirewallWrite, PermV2FirewallDelete,
	PermV2OrgRead, PermV2OrgWrite,
	PermV2PlatformAdmin,
}

// ResourceScope identifies a resource that a permission check should be scoped to.
type ResourceScope struct {
	Type          string
	ID            int64
	LocationScope *int64
}

func IsValidPermission(p string) bool {
	for _, v := range AllPermissions {
		if v == p {
			return true
		}
	}
	return false
}

func IsValidRole(role string) bool {
	for _, v := range ValidRoles {
		if v == role {
			return true
		}
	}
	return false
}

// ---- Package-level helpers (security) ----

func lockoutDuration(lockoutCount int) time.Duration {
	switch {
	case lockoutCount <= 1:
		return 5 * time.Minute
	case lockoutCount == 2:
		return 15 * time.Minute
	case lockoutCount == 3:
		return 1 * time.Hour
	case lockoutCount == 4:
		return 4 * time.Hour
	case lockoutCount == 5:
		return 24 * time.Hour
	default:
		return 7 * 24 * time.Hour
	}
}

// ---- Package-level helpers (session) ----

func parseDeviceName(userAgent string) string {
	ua := userAgent
	if ua == "" {
		return "Unknown Device"
	}
	switch {
	case strings.Contains(ua, "iPhone"):
		return "iPhone"
	case strings.Contains(ua, "iPad"):
		return "iPad"
	case strings.Contains(ua, "Android"):
		if strings.Contains(ua, "Mobile") {
			return "Android Phone"
		}
		return "Android Tablet"
	case strings.Contains(ua, "Macintosh") || strings.Contains(ua, "Mac OS X"):
		return "Mac"
	case strings.Contains(ua, "Windows"):
		return "Windows PC"
	case strings.Contains(ua, "Linux"):
		return "Linux"
	case strings.Contains(ua, "curl"):
		return "curl"
	default:
		if len(ua) > 50 {
			return ua[:50]
		}
		return ua
	}
}

// ---- Package-level helpers (users) ----

func generateTempPassword() (string, error) {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func validateAvatarImage(dataURL string) (string, error) {
	header, payload, found := strings.Cut(dataURL, ",")
	if !found || !strings.HasPrefix(header, "data:image/") || !strings.HasSuffix(header, ";base64") {
		return "", fmt.Errorf("avatar data must be a base64 image data URL")
	}
	raw, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return "", fmt.Errorf("avatar data is not valid base64")
	}
	cfg, format, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil {
		return "", fmt.Errorf("avatar data is not a recognized image")
	}
	switch format {
	case "jpeg", "png", "gif", "webp":
		// supported
	default:
		return "", fmt.Errorf("unsupported avatar image format: %s", format)
	}
	if cfg.Width <= 0 || cfg.Height <= 0 || cfg.Width > maxAvatarDimension || cfg.Height > maxAvatarDimension {
		return "", fmt.Errorf("avatar image dimensions must be between 1x1 and %dx%d", maxAvatarDimension, maxAvatarDimension)
	}
	return "data:image/" + format + ";base64," + payload, nil
}

// ---- Package-level helpers (rbac) ----

func permMatches(perms []*models.RolePermission, permission string, scopes []ResourceScope) bool {
	for _, p := range perms {
		if p.Permission != permission {
			continue
		}
		if p.ResourceType == nil {
			return true
		}
		if p.ResourceID == nil {
			continue
		}
		for _, s := range scopes {
			if *p.ResourceType == s.Type && *p.ResourceID == s.ID {
				return true
			}
		}
	}
	return false
}

func legacyRoleHasPermission(role, permission string) bool {
	switch role {
	case "admin":
		return true
	case "user":
		adminOnly := map[string]bool{
			PermV2UserWrite: true, PermV2AuditRead: true, PermV2DeviceAdmin: true,
			PermV2SubnetRequestReview: true, PermV2AdminRead: true, PermV2AdminWrite: true,
			PermV2CustomerWrite: true, PermV2CustomerDelete: true,
			PermV2ASWrite: true, PermV2ASDelete: true,
			PermV2NATWrite: true, PermV2NATDelete: true,
			PermV2DHCPWrite: true, PermV2DHCPDelete: true,
			PermV2CircuitWrite: true, PermV2CircuitDelete: true,
			PermV2FirewallWrite: true, PermV2FirewallDelete: true,
			PermV2OrgRead: true, PermV2OrgWrite: true, PermV2PlatformAdmin: true,
		}
		return !adminOnly[permission]
	case "viewer":
		readPerms := map[string]bool{
			PermV2NetworkList: true, PermV2NetworkRead: true,
			PermV2SubnetList: true, PermV2SubnetRead: true,
			PermV2IPList: true, PermV2IPRead: true,
			PermV2VRFList: true, PermV2VRFRead: true,
			PermV2VLANList: true, PermV2VLANRead: true,
			PermV2UserList: true, PermV2UserRead: true,
			PermV2DeviceRead:   true,
			PermV2LocationList: true, PermV2LocationRead: true,
			PermV2NameserverList: true, PermV2NameserverRead: true,
			PermV2VLANDomainList: true, PermV2VLANDomainRead: true,
			PermV2VLANGroupList: true, PermV2VLANGroupRead: true,
			PermV2CustomerList: true, PermV2CustomerRead: true,
			PermV2ASList: true, PermV2ASRead: true,
			PermV2NATList: true, PermV2NATRead: true,
			PermV2DHCPList: true, PermV2DHCPRead: true,
			PermV2CircuitList: true, PermV2CircuitRead: true,
			PermV2FirewallList: true, PermV2FirewallRead: true,
		}
		return readPerms[permission]
	}
	return false
}

func getRolePermissions(role string) []string {
	switch role {
	case RoleAdmin:
		return []string{
			PermNetworkCreate, PermNetworkRead, PermNetworkUpdate, PermNetworkDelete,
			PermSubnetCreate, PermSubnetRead, PermSubnetUpdate, PermSubnetDelete,
			PermIPCreate, PermIPRead, PermIPUpdate, PermIPDelete, PermIPAssign, PermIPRelease,
			PermTokenCreate, PermTokenRead, PermTokenDelete,
		}
	case RoleUser:
		return []string{
			PermNetworkRead, PermNetworkCreate, PermNetworkUpdate, PermNetworkDelete,
			PermSubnetRead, PermSubnetCreate, PermSubnetUpdate, PermSubnetDelete,
			PermIPRead, PermIPCreate, PermIPUpdate, PermIPDelete, PermIPAssign, PermIPRelease,
			PermTokenCreate, PermTokenRead, PermTokenDelete,
		}
	case RoleViewer:
		return []string{PermNetworkRead, PermSubnetRead, PermIPRead, PermTokenRead}
	}
	return nil
}

func isValidRoleName(name string) bool {
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
	}
	return true
}

// ============================================================
// User management methods
// ============================================================

func (s *IdentityService) CreateUser(ctx context.Context, username, email string) (*models.User, error) {
	if username == "" {
		return nil, fmt.Errorf("username is required")
	}
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, fmt.Errorf("invalid email format: %w", err)
	}
	return s.repo.CreateUser(ctx, username, email)
}

func (s *IdentityService) GetUser(ctx context.Context, id int64) (*models.User, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid user ID")
	}
	return s.repo.GetUserByID(ctx, id)
}

func (s *IdentityService) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid user ID")
	}
	return s.repo.GetUserByID(ctx, id)
}

func (s *IdentityService) ListAllUsers(ctx context.Context) ([]*models.User, error) {
	return s.repo.ListAllUsers(ctx)
}

func (s *IdentityService) CreateUserWithPassword(ctx context.Context, username, email, passwordHash, role string) (*models.User, error) {
	if username == "" {
		return nil, fmt.Errorf("username is required")
	}
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, fmt.Errorf("invalid email format: %w", err)
	}
	if role != "admin" && role != "user" && role != "viewer" {
		return nil, fmt.Errorf("invalid role")
	}
	return s.repo.CreateUserWithPassword(ctx, username, email, passwordHash, role)
}

func (s *IdentityService) UpdateUserRole(ctx context.Context, userID int64, role string) (*models.User, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("invalid user ID")
	}
	if role != "admin" && role != "user" && role != "viewer" {
		return nil, fmt.Errorf("invalid role")
	}
	return s.repo.UpdateUserRole(ctx, userID, role)
}

func (s *IdentityService) DeleteUser(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return fmt.Errorf("invalid user ID")
	}
	return s.repo.DeleteUser(ctx, userID)
}

func (s *IdentityService) UpdateUserEmail(ctx context.Context, userID int64, email string) error {
	if userID <= 0 {
		return fmt.Errorf("invalid user ID")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return fmt.Errorf("invalid email format")
	}
	return s.repo.UpdateUserEmail(ctx, userID, email)
}

func (s *IdentityService) SuspendUser(ctx context.Context, userID, adminID int64, reason string) error {
	if userID <= 0 {
		return fmt.Errorf("invalid user ID")
	}
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}
	if user.Role == "admin" {
		return fmt.Errorf("cannot suspend admin users")
	}
	return s.repo.SuspendUser(ctx, userID, adminID, reason)
}

func (s *IdentityService) UnsuspendUser(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return fmt.Errorf("invalid user ID")
	}
	return s.repo.UnsuspendUser(ctx, userID)
}

func (s *IdentityService) BulkSuspendUsers(ctx context.Context, userIDs []int64, adminID int64, reason string) (int64, error) {
	filtered := make([]int64, 0, len(userIDs))
	for _, id := range userIDs {
		u, err := s.repo.GetUserByID(ctx, id)
		if err == nil && u.Role != "admin" {
			filtered = append(filtered, id)
		}
	}
	if len(filtered) == 0 {
		return 0, nil
	}
	return s.repo.BulkUpdateUserState(ctx, filtered, "suspended")
}

func (s *IdentityService) BulkActivateUsers(ctx context.Context, userIDs []int64) (int64, error) {
	return s.repo.BulkUpdateUserState(ctx, userIDs, "active")
}

func (s *IdentityService) BulkDeleteUsers(ctx context.Context, userIDs []int64) (int64, error) {
	remainingAdmins, err := s.repo.CountAdminsExcluding(ctx, userIDs)
	if err != nil {
		return 0, fmt.Errorf("failed to check admin count: %w", err)
	}
	if remainingAdmins == 0 {
		return 0, fmt.Errorf("cannot delete all admins")
	}
	return s.repo.BulkDeleteUsers(ctx, userIDs)
}

func (s *IdentityService) BulkImportUsers(ctx context.Context, records []BulkUserImportRecord) ([]BulkImportResult, error) {
	results := make([]BulkImportResult, 0, len(records))
	for _, rec := range records {
		role := rec.Role
		if role == "" {
			role = "user"
		}
		if role != "admin" && role != "user" && role != "viewer" {
			results = append(results, BulkImportResult{Username: rec.Username, Error: "invalid role"})
			continue
		}
		if rec.Username == "" || rec.Email == "" {
			results = append(results, BulkImportResult{Username: rec.Username, Error: "username and email required"})
			continue
		}
		tempPassword, err := generateTempPassword()
		if err != nil {
			results = append(results, BulkImportResult{Username: rec.Username, Error: "failed to generate password"})
			continue
		}
		hash, err := utils.HashPassword(tempPassword)
		if err != nil {
			results = append(results, BulkImportResult{Username: rec.Username, Error: "password hash error"})
			continue
		}
		user, err := s.repo.CreateUserWithPassword(ctx, rec.Username, rec.Email, hash, role)
		if err != nil {
			results = append(results, BulkImportResult{Username: rec.Username, Error: err.Error()})
			continue
		}
		results = append(results, BulkImportResult{Username: rec.Username, UserID: user.ID, TempPassword: tempPassword})
	}
	return results, nil
}

func (s *IdentityService) AcceptPrivacyPolicy(ctx context.Context, userID int64) error {
	version, err := s.config.GetCtx(ctx, "privacy_policy_version")
	if err != nil || version == "" {
		version = "1.0"
	}
	return s.repo.UpdatePrivacyConsent(ctx, userID, version)
}

func (s *IdentityService) GetPrivacyPolicyVersion(ctx context.Context) string {
	v, err := s.config.GetCtx(ctx, "privacy_policy_version")
	if err != nil || v == "" {
		return "1.0"
	}
	return v
}

func (s *IdentityService) RequestAccountDeletion(ctx context.Context, userID int64) error {
	return s.repo.RequestDeletion(ctx, userID)
}

func (s *IdentityService) GDPRDeleteUser(ctx context.Context, userID int64) error {
	_ = s.repo.DeleteAllUserSessions(ctx, userID)
	return s.repo.AnonymizeUser(ctx, userID)
}

func (s *IdentityService) ExportUserData(ctx context.Context, userID int64) (map[string]interface{}, error) {
	return s.repo.GetUserAllData(ctx, userID)
}

func (s *IdentityService) StartImpersonation(ctx context.Context, targetUserID, adminID int64, ipAddress, userAgent string) (string, error) {
	target, err := s.repo.GetUserByID(ctx, targetUserID)
	if err != nil {
		return "", fmt.Errorf("target user not found")
	}
	if target.Role == "admin" {
		return "", fmt.Errorf("cannot impersonate admin users")
	}
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	rawToken := hex.EncodeToString(tokenBytes)
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])
	expiry := time.Now().UTC().Add(1 * time.Hour)
	_, err = s.repo.CreateImpersonationSession(ctx, targetUserID, adminID, tokenHash,
		"impersonation", ipAddress, userAgent, expiry)
	if err != nil {
		return "", err
	}
	return rawToken, nil
}

func (s *IdentityService) GetUserAvatarData(ctx context.Context, userID int64) (*string, error) {
	return s.repo.GetUserAvatarData(ctx, userID)
}

func (s *IdentityService) UpdateUserAvatar(ctx context.Context, userID int64, source string, data *string) error {
	if source != "gravatar" && source != "custom" {
		return fmt.Errorf("invalid avatar source: must be 'gravatar' or 'custom'")
	}
	if source == "custom" && (data == nil || *data == "") {
		return fmt.Errorf("avatar data is required when source is 'custom'")
	}
	if data != nil && len(*data) > 3*1024*1024 {
		return fmt.Errorf("avatar data exceeds maximum allowed size")
	}
	if source == "gravatar" {
		data = nil
	}
	if data != nil {
		normalized, err := validateAvatarImage(*data)
		if err != nil {
			return err
		}
		data = &normalized
	}
	return s.repo.UpdateUserAvatar(ctx, userID, source, data)
}

// ============================================================
// RBAC methods
// ============================================================

func (s *IdentityService) CheckPermission(ctx context.Context, userID int64, permission string, scopes ...ResourceScope) error {
	if userID <= 0 {
		return fmt.Errorf("permission denied")
	}
	count, err := s.repo.CountUserRoles(ctx, userID)
	if err == nil && count > 0 {
		perms, err := s.repo.GetUserPermissions(ctx, userID)
		if err != nil {
			return fmt.Errorf("permission denied")
		}
		if permMatches(perms, permission, scopes) {
			for _, scope := range scopes {
				if scope.LocationScope == nil {
					continue
				}
				_, hasGlobal, err := s.repo.GetUserRoleLocationIDs(ctx, userID)
				if err != nil || hasGlobal {
					return nil
				}
				allowed, err := s.isLocationAllowed(ctx, userID, *scope.LocationScope)
				if err != nil || !allowed {
					return fmt.Errorf("permission denied: location not in scope")
				}
			}
			return nil
		}
		return fmt.Errorf("permission denied: %s", permission)
	}
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("permission denied")
	}
	if legacyRoleHasPermission(user.Role, permission) {
		return nil
	}
	// Check direct role grants (global first, then any matching scope).
	if ok, _ := s.repo.UserHasGrant(ctx, userID, permission, nil, nil); ok {
		return nil
	}
	for i := range scopes {
		if ok, _ := s.repo.UserHasGrant(ctx, userID, permission, &scopes[i].Type, &scopes[i].ID); ok {
			return nil
		}
	}
	return fmt.Errorf("permission denied: %s", permission)
}

func (s *IdentityService) isLocationAllowed(ctx context.Context, userID int64, locationID int64) (bool, error) {
	ancestors, err := s.repo.GetLocationAncestors(ctx, locationID)
	if err != nil {
		return false, err
	}
	scopedIDs, _, err := s.repo.GetUserRoleLocationIDs(ctx, userID)
	if err != nil {
		return false, err
	}
	for _, anc := range ancestors {
		for _, allowed := range scopedIDs {
			if anc == allowed {
				return true, nil
			}
		}
	}
	return false, nil
}

func (s *IdentityService) HasPermission(user *models.User, permission string) bool {
	if user == nil {
		return false
	}
	if user.Role == RoleAdmin {
		return true
	}
	for _, p := range getRolePermissions(user.Role) {
		if p == permission {
			return true
		}
	}
	return false
}

func (s *IdentityService) CanAccessResource(user *models.User, action string) bool {
	return s.HasPermission(user, action)
}

func (s *IdentityService) RequirePermission(user *models.User, permission string) error {
	if !s.HasPermission(user, permission) {
		return fmt.Errorf("user does not have permission: %s", permission)
	}
	return nil
}

func (s *IdentityService) CreateRole(ctx context.Context, name, description string) (*models.Role, error) {
	if name == "" {
		return nil, fmt.Errorf("role name is required")
	}
	if !isValidRoleName(name) {
		return nil, fmt.Errorf("role name may only contain letters, numbers, hyphens, and underscores")
	}
	return s.repo.CreateRole(ctx, name, description, false)
}

func (s *IdentityService) GetRole(ctx context.Context, id int64) (*models.Role, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid role ID")
	}
	return s.repo.GetRoleByID(ctx, id)
}

func (s *IdentityService) ListRoles(ctx context.Context) ([]*models.Role, error) {
	roles, err := s.repo.ListRoles(ctx)
	if err != nil {
		return nil, err
	}
	if roles == nil {
		roles = []*models.Role{}
	}
	return roles, nil
}

func (s *IdentityService) UpdateRole(ctx context.Context, id int64, name, description string) (*models.Role, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid role ID")
	}
	if name == "" {
		return nil, fmt.Errorf("role name is required")
	}
	return s.repo.UpdateRole(ctx, id, name, description)
}

func (s *IdentityService) DeleteRole(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid role ID")
	}
	return s.repo.DeleteRole(ctx, id)
}

func (s *IdentityService) AddPermissionToRole(ctx context.Context, roleID int64, permission string, resourceType *string, resourceID *int64) (*models.RolePermission, error) {
	if roleID <= 0 {
		return nil, fmt.Errorf("invalid role ID")
	}
	if !IsValidPermission(permission) {
		return nil, fmt.Errorf("unknown permission: %s", permission)
	}
	if resourceType != nil && *resourceType != "" && resourceID == nil {
		return nil, fmt.Errorf("resource ID is required when resource type is set")
	}
	return s.repo.AddPermissionToRole(ctx, roleID, permission, resourceType, resourceID)
}

func (s *IdentityService) RemovePermissionFromRole(ctx context.Context, permissionID int64) error {
	if permissionID <= 0 {
		return fmt.Errorf("invalid permission ID")
	}
	return s.repo.RemovePermissionFromRole(ctx, permissionID)
}

func (s *IdentityService) GetUserRoles(ctx context.Context, userID int64) ([]*models.Role, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("invalid user ID")
	}
	roles, err := s.repo.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}
	if roles == nil {
		roles = []*models.Role{}
	}
	return roles, nil
}

func (s *IdentityService) AssignRoleToUser(ctx context.Context, userID, roleID int64, locationID ...*int64) error {
	if userID <= 0 {
		return fmt.Errorf("invalid user ID")
	}
	if roleID <= 0 {
		return fmt.Errorf("invalid role ID")
	}
	var locID *int64
	if len(locationID) > 0 {
		locID = locationID[0]
	}
	return s.repo.AssignRoleToUserWithLocation(ctx, userID, roleID, locID)
}

func (s *IdentityService) RemoveRoleFromUser(ctx context.Context, userID, roleID int64) error {
	if userID <= 0 {
		return fmt.Errorf("invalid user ID")
	}
	if roleID <= 0 {
		return fmt.Errorf("invalid role ID")
	}
	return s.repo.RemoveRoleFromUser(ctx, userID, roleID)
}

// ============================================================
// Security methods
// ============================================================

func (s *IdentityService) ProcessFailedLogin(ctx context.Context, userID int64, username, ipAddress, userAgent, failureReason string) {
	_ = s.repo.CreateLoginAttempt(ctx, username, ipAddress, userAgent, false, failureReason)
	if userID == 0 {
		return
	}
	since := time.Now().UTC().Add(-bruteForceWindow)
	count, err := s.repo.CountRecentFailedAttemptsByUsername(ctx, username, since)
	if err != nil {
		return
	}
	if count >= maxFailedAttempts {
		_ = s.lockAccount(ctx, userID, username, count)
	} else if count >= ipFailureThreshold {
		_ = s.sendFailedLoginAlert(ctx, userID, username, ipAddress, count)
	}
}

func (s *IdentityService) IsIPThrottled(ctx context.Context, ipAddress string) (bool, error) {
	since := time.Now().UTC().Add(-ipThrottleWindow)
	count, err := s.repo.CountRecentFailedAttemptsByIPOnly(ctx, ipAddress, since)
	if err != nil {
		return false, err
	}
	return count >= ipThrottleThreshold, nil
}

func (s *IdentityService) lockAccount(ctx context.Context, userID int64, username string, failCount int) error {
	lockoutCount, _ := s.repo.CountUserLockouts(ctx, userID)
	lockoutCount++
	duration := lockoutDuration(lockoutCount)
	unlockAt := time.Now().UTC().Add(duration)
	reason := fmt.Sprintf("%d failed login attempts within %s", failCount, bruteForceWindow)
	lockout, err := s.repo.CreateAccountLockout(ctx, userID, unlockAt, reason, lockoutCount)
	if err != nil {
		return err
	}
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return err
	}
	rawToken := hex.EncodeToString(tokenBytes)
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])
	tokenExpiry := time.Now().UTC().Add(24 * time.Hour)
	if err := s.repo.SetUnlockToken(ctx, lockout.ID, tokenHash, tokenExpiry); err != nil {
		return err
	}
	appURL, _ := s.config.GetCtx(ctx, "app_url")
	if appURL == "" {
		appURL = "http://localhost:3000"
	}
	unlockURL := fmt.Sprintf("%s/unlock-account?token=%s", appURL, rawToken)
	_ = s.notification.Queue(ctx, userID, NotifAccountLocked, map[string]interface{}{
		"UnlockURL": unlockURL,
		"Duration":  duration.String(),
	})
	_ = s.repo.CreateSecurityNotification(ctx, userID, "account_locked", "")
	return nil
}

func (s *IdentityService) sendFailedLoginAlert(ctx context.Context, userID int64, username, ipAddress string, count int) error {
	since := time.Now().UTC().Add(-notifRateLimit)
	recent, err := s.repo.CountRecentSecurityNotifications(ctx, userID, "failed_login_alert", since)
	if err != nil || recent > 0 {
		return err
	}
	if err := s.notification.Queue(ctx, userID, NotifLoginFailed, map[string]interface{}{
		"IP":    ipAddress,
		"Count": count,
	}); err != nil {
		return err
	}
	return s.repo.CreateSecurityNotification(ctx, userID, "failed_login_alert", ipAddress)
}

func (s *IdentityService) IsAccountLocked(ctx context.Context, userID int64) (bool, *models.AccountLockout, error) {
	lockout, err := s.repo.GetActiveAccountLockout(ctx, userID)
	if err != nil {
		return false, nil, nil
	}
	return true, lockout, nil
}

func (s *IdentityService) RecordSuccessfulLogin(ctx context.Context, username, ipAddress, userAgent string) {
	_ = s.repo.CreateLoginAttempt(ctx, username, ipAddress, userAgent, true, "")
}

func (s *IdentityService) GetLoginHistory(ctx context.Context, userID int64, limit int) ([]*models.LoginAttempt, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repo.GetLoginHistory(ctx, user.Username, limit)
}

func (s *IdentityService) RequestUnlockEmail(ctx context.Context, username string) error {
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil
	}
	locked, lockout, err := s.IsAccountLocked(ctx, user.ID)
	if err != nil || !locked {
		return nil
	}
	if lockout.UnlockTokenHash != nil && lockout.UnlockTokenExpiresAt != nil && lockout.UnlockTokenExpiresAt.After(time.Now()) {
		return nil
	}
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return err
	}
	rawToken := hex.EncodeToString(tokenBytes)
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])
	tokenExpiry := time.Now().UTC().Add(24 * time.Hour)
	if err := s.repo.SetUnlockToken(ctx, lockout.ID, tokenHash, tokenExpiry); err != nil {
		return err
	}
	appURL, _ := s.config.GetCtx(ctx, "app_url")
	if appURL == "" {
		appURL = "http://localhost:3000"
	}
	unlockURL := fmt.Sprintf("%s/unlock-account?token=%s", appURL, rawToken)
	duration := time.Until(lockout.UnlockAt)
	_ = s.email.SendAccountLockedEmail(user.Email, user.Username, unlockURL, duration)
	return nil
}

func (s *IdentityService) UnlockAccountByToken(ctx context.Context, rawToken string) error {
	if rawToken == "" {
		return ErrInvalidUnlockToken
	}
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])
	lockout, err := s.repo.GetLockoutByUnlockToken(ctx, tokenHash)
	if err != nil {
		return ErrInvalidUnlockToken
	}
	if lockout.UnlockTokenUsedAt != nil {
		return ErrInvalidUnlockToken
	}
	if lockout.UnlockTokenExpiresAt != nil && lockout.UnlockTokenExpiresAt.Before(time.Now()) {
		return ErrInvalidUnlockToken
	}
	if lockout.UnlockedAt != nil {
		return nil
	}
	if err := s.repo.UnlockAccount(ctx, lockout.ID, nil); err != nil {
		return err
	}
	return s.repo.MarkUnlockTokenUsed(ctx, lockout.ID)
}

func (s *IdentityService) UnlockAccountByAdmin(ctx context.Context, userID, adminID int64) error {
	locked, lockout, err := s.IsAccountLocked(ctx, userID)
	if err != nil {
		return err
	}
	if !locked {
		return nil
	}
	return s.repo.UnlockAccount(ctx, lockout.ID, &adminID)
}

// ============================================================
// Session methods
// ============================================================

func (s *IdentityService) sessionIdleTimeout(ctx context.Context) time.Duration {
	if val, err := s.config.GetCtx(ctx, "session_idle_timeout_minutes"); err == nil && val != "" {
		var mins int
		if _, err := fmt.Sscanf(val, "%d", &mins); err == nil && mins > 0 {
			return time.Duration(mins) * time.Minute
		}
	}
	return DefaultIdleTimeoutMinutes * time.Minute
}

func (s *IdentityService) sessionAbsoluteTimeout(ctx context.Context) time.Duration {
	if val, err := s.config.GetCtx(ctx, "session_absolute_timeout_hours"); err == nil && val != "" {
		var hrs int
		if _, err := fmt.Sscanf(val, "%d", &hrs); err == nil && hrs > 0 {
			return time.Duration(hrs) * time.Hour
		}
	}
	return DefaultAbsoluteTimeoutHours * time.Hour
}

func (s *IdentityService) CreateWebSession(ctx context.Context, userID int64, ipAddress, userAgent string) (string, error) {
	if userID <= 0 {
		return "", fmt.Errorf("invalid user ID")
	}
	tokenBytes := make([]byte, sessionTokenLength)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("failed to generate session token: %w", err)
	}
	rawToken := hex.EncodeToString(tokenBytes)
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])
	deviceName := parseDeviceName(userAgent)
	absoluteExpiry := time.Now().UTC().Add(s.sessionAbsoluteTimeout(ctx))
	_, err := s.repo.CreateSession(ctx, userID, tokenHash, deviceName, ipAddress, userAgent, absoluteExpiry)
	if err != nil {
		return "", fmt.Errorf("failed to store session: %w", err)
	}
	return rawToken, nil
}

func (s *IdentityService) ValidateSession(ctx context.Context, rawToken string) (*models.User, *models.Session, error) {
	if rawToken == "" {
		return nil, nil, fmt.Errorf("token is required")
	}
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])
	session, err := s.repo.GetSessionByHash(ctx, tokenHash)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid session")
	}
	now := time.Now()
	if now.After(session.AbsoluteExpiresAt) {
		_ = s.repo.DeleteSession(ctx, session.ID)
		return nil, nil, fmt.Errorf("session expired")
	}
	idleTimeout := s.sessionIdleTimeout(ctx)
	if now.After(session.LastUsedAt.Add(idleTimeout)) {
		_ = s.repo.DeleteSession(ctx, session.ID)
		return nil, nil, fmt.Errorf("session expired due to inactivity")
	}
	_ = s.repo.UpdateSessionLastUsed(ctx, session.ID)
	user, err := s.repo.GetUserByID(ctx, session.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("user not found")
	}
	if user.State == "suspended" || user.State == "deleted" {
		_ = s.repo.DeleteSession(ctx, session.ID)
		return nil, nil, fmt.Errorf("account is %s", user.State)
	}
	return user, session, nil
}

func (s *IdentityService) RevokeSession(ctx context.Context, userID int64, rawToken string) error {
	if rawToken == "" {
		return fmt.Errorf("token is required")
	}
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])
	session, err := s.repo.GetSessionByHash(ctx, tokenHash)
	if err != nil {
		return fmt.Errorf("session not found")
	}
	if session.UserID != userID {
		return fmt.Errorf("session does not belong to this user")
	}
	return s.repo.DeleteSession(ctx, session.ID)
}

func (s *IdentityService) RevokeSessionByID(ctx context.Context, userID, sessionID int64) error {
	if sessionID <= 0 {
		return fmt.Errorf("invalid session ID")
	}
	sessions, err := s.repo.ListSessionsByUser(ctx, userID)
	if err != nil {
		return err
	}
	for _, sess := range sessions {
		if sess.ID == sessionID {
			return s.repo.DeleteSession(ctx, sessionID)
		}
	}
	return fmt.Errorf("session not found")
}

func (s *IdentityService) RevokeAllSessions(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return fmt.Errorf("invalid user ID")
	}
	return s.repo.DeleteAllUserSessions(ctx, userID)
}

func (s *IdentityService) ListUserSessions(ctx context.Context, userID int64) ([]*models.Session, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("invalid user ID")
	}
	return s.repo.ListSessionsByUser(ctx, userID)
}

func (s *IdentityService) UpdateLastLogin(ctx context.Context, userID int64) error {
	return s.repo.UpdateLastLogin(ctx, userID)
}

func (s *IdentityService) IsSessionExpired(lastLoginAt *time.Time) bool {
	if lastLoginAt == nil {
		return false
	}
	return time.Now().After(lastLoginAt.Add(DefaultSessionTimeout))
}

// ============================================================
// API token / auth methods
// ============================================================

func (s *IdentityService) GenerateAPIToken(ctx context.Context, userID int64, tokenName, scope string, expiresInDays int) (string, error) {
	if userID <= 0 {
		return "", fmt.Errorf("invalid user ID")
	}
	if tokenName == "" {
		return "", fmt.Errorf("token name is required")
	}
	switch scope {
	case "read", "write", "admin":
	default:
		scope = "write"
	}
	var expiresAt *time.Time
	if expiresInDays == 0 {
		defaultDaysStr, _ := s.config.GetCtx(ctx, "api_token_default_expiration_days")
		if n, err := strconv.Atoi(defaultDaysStr); err == nil && n > 0 {
			t := time.Now().UTC().Add(time.Duration(n) * 24 * time.Hour)
			expiresAt = &t
		}
	} else if expiresInDays > 0 {
		t := time.Now().UTC().Add(time.Duration(expiresInDays) * 24 * time.Hour)
		expiresAt = &t
	}
	tokenBytes := make([]byte, TokenLength)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(tokenBytes)
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])
	if _, err := s.repo.CreateAPITokenFull(ctx, userID, tokenHash, tokenName, scope, expiresAt); err != nil {
		return "", err
	}
	return token, nil
}

func (s *IdentityService) ValidateAPIToken(ctx context.Context, token, ip string) (*models.User, *models.APIToken, error) {
	if token == "" {
		return nil, nil, fmt.Errorf("token is required")
	}
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])
	apiToken, err := s.repo.GetAPITokenByHash(ctx, tokenHash)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid token")
	}
	if apiToken.ExpiresAt != nil && apiToken.ExpiresAt.Before(time.Now()) {
		return nil, nil, fmt.Errorf("token has expired")
	}
	if apiToken.RotationGraceExpiresAt != nil && apiToken.RotationGraceExpiresAt.Before(time.Now()) {
		return nil, nil, fmt.Errorf("token has been rotated and grace period has expired")
	}
	_ = s.repo.UpdateAPITokenLastUsed(ctx, apiToken.ID, ip)
	user, err := s.repo.GetUserByID(ctx, apiToken.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("user not found")
	}
	return user, apiToken, nil
}

func (s *IdentityService) RotateAPIToken(ctx context.Context, tokenID, userID int64) (newToken string, graceExpiresAt time.Time, err error) {
	oldToken, err := s.repo.GetAPITokenByID(ctx, tokenID)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("token not found")
	}
	if oldToken.UserID != userID {
		return "", time.Time{}, fmt.Errorf("token does not belong to this user")
	}
	gracePeriod := 24 * time.Hour
	if graceHoursStr, err2 := s.config.GetCtx(ctx, "api_token_rotation_grace_period_hours"); err2 == nil {
		if n, err3 := strconv.Atoi(graceHoursStr); err3 == nil && n > 0 {
			gracePeriod = time.Duration(n) * time.Hour
		}
	}
	graceExpiresAt = time.Now().UTC().Add(gracePeriod)
	if err = s.repo.MarkAPITokenRotated(ctx, tokenID, graceExpiresAt); err != nil {
		return "", time.Time{}, fmt.Errorf("failed to rotate token: %w", err)
	}
	var expiresInDays int
	if oldToken.ExpiresAt != nil {
		remaining := time.Until(*oldToken.ExpiresAt)
		if remaining < 24*time.Hour {
			remaining = 24 * time.Hour
		}
		expiresInDays = int(remaining.Hours() / 24)
	}
	newRawToken, err := s.GenerateAPIToken(ctx, userID, oldToken.Name, oldToken.Scope, expiresInDays)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to create replacement token: %w", err)
	}
	return newRawToken, graceExpiresAt, nil
}

func (s *IdentityService) ExtendAPIToken(ctx context.Context, tokenID, userID int64, days int) (*models.APIToken, error) {
	existing, err := s.repo.GetAPITokenByID(ctx, tokenID)
	if err != nil {
		return nil, fmt.Errorf("token not found")
	}
	if existing.UserID != userID {
		return nil, fmt.Errorf("token does not belong to this user")
	}
	if days == 0 {
		defaultDaysStr, _ := s.config.GetCtx(ctx, "api_token_default_expiration_days")
		if n, err2 := strconv.Atoi(defaultDaysStr); err2 == nil && n > 0 {
			days = n
		} else {
			days = 30
		}
	}
	newExpiresAt := time.Now().UTC().Add(time.Duration(days) * 24 * time.Hour)
	return s.repo.ExtendAPIToken(ctx, tokenID, userID, newExpiresAt)
}

func (s *IdentityService) CleanupExpiredTokens(ctx context.Context) error {
	return s.repo.DeleteExpiredAPITokens(ctx)
}

func (s *IdentityService) RevokeAPIToken(ctx context.Context, tokenID int64) error {
	if tokenID <= 0 {
		return fmt.Errorf("invalid token ID")
	}
	return s.repo.DeleteAPIToken(ctx, tokenID)
}

func (s *IdentityService) ListUserTokens(ctx context.Context, userID int64) ([]*models.APIToken, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("invalid user ID")
	}
	return s.repo.ListAPITokensByUser(ctx, userID)
}

func (s *IdentityService) AuthenticateUser(ctx context.Context, username, password, ipAddress, userAgent string) (*AuthResult, error) {
	if username == "" || password == "" {
		return nil, fmt.Errorf("username and password required")
	}
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		utils.DummyVerifyPassword(password)
		s.ProcessFailedLogin(ctx, 0, username, ipAddress, userAgent, "user not found")
		return nil, fmt.Errorf("user not found")
	}
	if user.PasswordHash == "" {
		utils.DummyVerifyPassword(password)
		s.ProcessFailedLogin(ctx, user.ID, username, ipAddress, userAgent, "no password set")
		return nil, fmt.Errorf("user has no password set")
	}
	if !utils.VerifyPassword(user.PasswordHash, password) {
		s.ProcessFailedLogin(ctx, user.ID, username, ipAddress, userAgent, "invalid password")
		return nil, fmt.Errorf("invalid password")
	}
	if locked, lockout, _ := s.IsAccountLocked(ctx, user.ID); locked {
		_ = s.repo.CreateLoginAttempt(ctx, username, ipAddress, userAgent, false, "account locked")
		return nil, fmt.Errorf("%w; locked until %s", ErrAccountLocked, lockout.UnlockAt.Format(time.RFC3339))
	}
	switch user.State {
	case "pending_email_verification":
		return nil, ErrEmailNotVerified
	case "pending_admin_approval":
		return nil, ErrPendingApproval
	case "rejected":
		return nil, ErrAccountRejected
	case "disabled":
		return nil, ErrAccountDisabled
	case "suspended":
		return nil, fmt.Errorf("account is suspended")
	}
	if s.mfa.IsMFAEnabled(ctx, user.ID) {
		challenge, err := s.mfa.CreateChallenge(ctx, user.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to create MFA challenge: %w", err)
		}
		return &AuthResult{MFARequired: true, MFAChallenge: challenge}, nil
	}
	s.RecordSuccessfulLogin(ctx, username, ipAddress, userAgent)
	return &AuthResult{User: user}, nil
}

func (s *IdentityService) RevokeSessionToken(ctx context.Context, userID int64, token string) error {
	if token == "" {
		return fmt.Errorf("token is required")
	}
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])
	apiToken, err := s.repo.GetAPITokenByHash(ctx, tokenHash)
	if err != nil {
		return fmt.Errorf("token not found")
	}
	if apiToken.UserID != userID {
		return fmt.Errorf("token does not belong to this user")
	}
	return s.repo.DeleteAPIToken(ctx, apiToken.ID)
}

func (s *IdentityService) CreatePasswordResetToken(ctx context.Context, email string) (token string, err error) {
	if email == "" {
		return "", fmt.Errorf("email is required")
	}
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", fmt.Errorf("user not found")
	}
	tokenBytes := make([]byte, TokenLength)
	_, err = rand.Read(tokenBytes)
	if err != nil {
		return "", err
	}
	token = hex.EncodeToString(tokenBytes)
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])
	_, err = s.repo.CreatePasswordReset(ctx, user.ID, tokenHash)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *IdentityService) SendPasswordResetEmail(ctx context.Context, email string) error {
	token, err := s.CreatePasswordResetToken(ctx, email)
	if err != nil {
		return err
	}
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}
	return s.email.SendPasswordResetEmail(user.Email, user.Username, token)
}

func (s *IdentityService) SendPasswordResetEmailByID(ctx context.Context, userID int64) error {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}
	token, err := s.CreatePasswordResetToken(ctx, user.Email)
	if err != nil {
		return err
	}
	return s.email.SendPasswordResetEmail(user.Email, user.Username, token)
}

func (s *IdentityService) ResetPasswordWithToken(ctx context.Context, token, newPasswordHash string) (int64, error) {
	if token == "" || newPasswordHash == "" {
		return 0, fmt.Errorf("token and password hash required")
	}
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])
	resetRecord, err := s.repo.GetPasswordResetByToken(ctx, tokenHash)
	if err != nil {
		return 0, fmt.Errorf("invalid reset token")
	}
	if resetRecord.ExpiresAt.Before(time.Now()) {
		return 0, fmt.Errorf("reset token has expired")
	}
	if resetRecord.UsedAt != nil {
		return 0, fmt.Errorf("reset token has already been used")
	}
	if err := s.repo.UpdateUserPassword(ctx, resetRecord.UserID, newPasswordHash); err != nil {
		return 0, err
	}
	if err := s.repo.MarkPasswordResetAsUsed(ctx, resetRecord.ID); err != nil {
		return 0, err
	}
	if err := s.repo.DeleteAllUserSessions(ctx, resetRecord.UserID); err != nil {
		return 0, fmt.Errorf("password reset but failed to revoke sessions: %w", err)
	}
	return resetRecord.UserID, nil
}

func (s *IdentityService) InitAdminPassword(ctx context.Context, password string) (bool, error) {
	hash, err := utils.HashPassword(password)
	if err != nil {
		return false, fmt.Errorf("hashing admin password: %w", err)
	}
	return s.repo.InitAdminPassword(ctx, hash)
}

func (s *IdentityService) ChangePassword(ctx context.Context, userID int64, currentPassword, newPassword, keepSessionToken string) error {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}
	if user.PasswordHash == "" || !utils.VerifyPassword(user.PasswordHash, currentPassword) {
		return fmt.Errorf("current password is incorrect")
	}
	hash, err := utils.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password")
	}
	if err := s.repo.UpdateUserPassword(ctx, userID, hash); err != nil {
		return err
	}
	if keepSessionToken == "" {
		if err := s.repo.DeleteAllUserSessions(ctx, userID); err != nil {
			return fmt.Errorf("password changed but failed to revoke sessions: %w", err)
		}
		return nil
	}
	keepHash := sha256.Sum256([]byte(keepSessionToken))
	if err := s.repo.DeleteUserSessionsExcept(ctx, userID, hex.EncodeToString(keepHash[:])); err != nil {
		return fmt.Errorf("password changed but failed to revoke other sessions: %w", err)
	}
	return nil
}

func (s *IdentityService) ForceResetAdminPassword(ctx context.Context, password string) error {
	hash, err := utils.HashPassword(password)
	if err != nil {
		return fmt.Errorf("hashing admin password: %w", err)
	}
	return s.repo.ForceSetAdminPassword(ctx, hash)
}

// ============================================================
// Grafana datasource methods
// ============================================================

func (s *IdentityService) GrafanaSubnetUtilization(ctx context.Context) ([]repository.GrafanaSubnetRow, error) {
	return s.repo.GrafanaGetSubnetUtilization(ctx)
}

func (s *IdentityService) GrafanaIPCountsByStatus(ctx context.Context) ([]repository.GrafanaIPStatusRow, error) {
	return s.repo.GrafanaGetIPCountsByStatus(ctx)
}

func (s *IdentityService) GrafanaNetworkSummary(ctx context.Context) ([]repository.GrafanaSectionRow, error) {
	return s.repo.GrafanaGetSectionSummary(ctx)
}

// CreateGrantRequest carries the parameters for granting a direct permission.
type CreateGrantRequest struct {
	UserID     int64   `json:"user_id"`
	Permission string  `json:"permission"`
	ScopeType  *string `json:"scope_type"`
	ScopeID    *int64  `json:"scope_id"`
}

// CreateGrant grants a permission to a user. The grantor must themselves hold
// the permission to prevent privilege escalation.
func (s *IdentityService) CreateGrant(ctx context.Context, grantorID int64, req CreateGrantRequest) (*models.RoleGrant, error) {
	if !IsValidPermission(req.Permission) {
		return nil, fmt.Errorf("unknown permission: %s", req.Permission)
	}
	if err := s.CheckPermission(ctx, grantorID, req.Permission); err != nil {
		return nil, fmt.Errorf("cannot grant a permission you do not hold: %s", req.Permission)
	}
	target, err := s.repo.GetUserByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	if target.OrganizationID == nil {
		return nil, fmt.Errorf("target user has no organization")
	}
	return s.repo.CreateRoleGrant(ctx, *target.OrganizationID, req.UserID, req.Permission, req.ScopeType, req.ScopeID, &grantorID)
}

// RevokeGrant removes a direct permission grant by ID.
func (s *IdentityService) RevokeGrant(ctx context.Context, grantID int64) error {
	return s.repo.DeleteRoleGrant(ctx, grantID)
}

// ListUserGrants returns all direct permission grants for a user.
func (s *IdentityService) ListUserGrants(ctx context.Context, userID int64) ([]*models.RoleGrant, error) {
	grants, err := s.repo.ListUserGrants(ctx, userID)
	if err != nil {
		return nil, err
	}
	if grants == nil {
		grants = []*models.RoleGrant{}
	}
	return grants, nil
}
