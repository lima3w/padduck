package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"

	"padduck/models"
)

type intentRepo interface {
	CreateIntent(ctx context.Context, intent *models.ResourceIntent) (*models.ResourceIntent, error)
	GetIntent(ctx context.Context, id int64) (*models.ResourceIntent, error)
	ListIntents(ctx context.Context, status, resourceType string, resourceID *int64) ([]*models.ResourceIntent, error)
	TransitionPendingIntent(ctx context.Context, id int64, status string, reviewedBy *int64, note *string, reviewedAt time.Time) (*models.ResourceIntent, error)
	UpdateIntentStatus(ctx context.Context, id int64, status string, reviewedBy *int64, note *string, reviewedAt, appliedAt *time.Time) error
}

type intentIPAM interface {
	CreateSubnet(ctx context.Context, networkID int64, networkAddress string, prefixLength int, description string, gateway *string, autoFirst, autoLast bool, locationID *int64, nameserverID *int64, vlanID *int64, customFields ...map[string]*string) (*models.Subnet, error)
	UpdateSubnet(ctx context.Context, id int64, description string, gateway *string, autoFirst, autoLast bool, locationID *int64, nameserverID *int64, vlanID *int64, customFields map[string]*string, technitiumScopeName ...string) (*models.Subnet, error)
	DeleteSubnet(ctx context.Context, id int64) error
	GetSubnet(ctx context.Context, id int64) (*models.Subnet, error)
	GetIPAddress(ctx context.Context, id int64) (*models.IPAddress, error)
	GetVLAN(ctx context.Context, id int64) (*models.VLAN, error)
	CreateIPAddress(ctx context.Context, subnetID int64, address, hostname string, status string, tagID *int64, macAddress, ptrRecord, dnsName *string, customFields ...map[string]*string) (*models.IPAddress, error)
	UpdateIPAddressMeta(ctx context.Context, id int64, hostname string, tagID *int64, macAddress, ptrRecord, dnsName *string, customFields ...map[string]*string) (*models.IPAddress, error)
	DeleteIPAddress(ctx context.Context, id int64) error
	CreateVLAN(ctx context.Context, vrfID *int64, domainID *int64, groupID *int64, vlanID int, name, description string) (*models.VLAN, error)
	UpdateVLAN(ctx context.Context, id int64, domainID *int64, groupID *int64, name, description string) (*models.VLAN, error)
	DeleteVLAN(ctx context.Context, id int64) error
}

type intentInfra interface {
	CreateDevice(ctx context.Context, req *DeviceCreateRequest) (*models.Device, error)
	UpdateDevice(ctx context.Context, id int64, req *DeviceUpdateRequest) (*models.Device, error)
	DeleteDevice(ctx context.Context, id int64) error
	GetDevice(ctx context.Context, id int64) (*models.Device, error)
}

// intentPolicy evaluates automation policies before an intent is applied.
type intentPolicy interface {
	Evaluate(ctx context.Context, req AutomationRequest) (*PolicyDecision, error)
	// RecordApplied runs the matched policies' actions after the change is applied.
	RecordApplied(ctx context.Context, decision *PolicyDecision, resourceType string, resourceID int64, resourceName string)
}

type IntentService struct {
	repo   intentRepo
	config *ConfigService
	ipam   intentIPAM
	infra  intentInfra
	policy intentPolicy
}

func NewIntentService(repo intentRepo, config *ConfigService, ipam intentIPAM, infra intentInfra, policy intentPolicy) *IntentService {
	return &IntentService{repo: repo, config: config, ipam: ipam, infra: infra, policy: policy}
}

type policyBypassKey struct{}

// WithPolicyBypass marks a context so intent policy evaluation is skipped. Use
// it only for callers allowed to bypass policy (platform admins and
// bypass-flagged API tokens), and audit the bypass.
func WithPolicyBypass(ctx context.Context) context.Context {
	return context.WithValue(ctx, policyBypassKey{}, true)
}

func policyBypassed(ctx context.Context) bool {
	v, _ := ctx.Value(policyBypassKey{}).(bool)
	return v
}

// PolicyDeniedError is returned when an automation policy blocks an intent.
type PolicyDeniedError struct{ Message string }

func (e *PolicyDeniedError) Error() string { return e.Message }

// intentPolicyRequest builds the policy workflow, action and values for an
// intent. The names match the direct-write paths so one policy covers both.
func intentPolicyRequest(i *models.ResourceIntent) AutomationRequest {
	action := i.Operation
	if i.ResourceType == "device" && i.Operation == "create" {
		action = "register"
	}
	values := map[string]string{}
	for k, v := range i.DesiredState {
		switch t := v.(type) {
		case nil:
		case string:
			values[k] = t
		case float64:
			values[k] = strconv.FormatFloat(t, 'f', -1, 64)
		case bool:
			values[k] = strconv.FormatBool(t)
		}
	}
	if v, ok := values["prefix_length"]; ok {
		values["prefix_len"] = v
	}
	if i.ResourceID != nil {
		values["resource_id"] = strconv.FormatInt(*i.ResourceID, 10)
	}
	return AutomationRequest{Workflow: i.ResourceType, Action: action, Values: values, TriggeredBy: "api", OrgID: i.OrganizationID}
}

func (s *IntentService) isAutoApprove(ctx context.Context) bool {
	v, err := s.config.GetCtx(ctx, "intent_auto_approve")
	if err != nil || v == "" || v == "true" {
		return true
	}
	return false
}

// ErrIntentNotPending is returned when approving or rejecting an intent that is
// missing, belongs to another organization, or has already been reviewed.
var ErrIntentNotPending = errors.New("intent is not pending")

// IntentResourceTypes lists the resource types the intent model can apply.
var IntentResourceTypes = []string{"subnet", "ip_address", "device", "vlan"}

// IntentPermission returns the permission a caller needs to propose the given
// operation on a resource type. Applying an intent never grants more than the
// submitter could do directly.
func IntentPermission(resourceType, operation string) (string, bool) {
	del := operation == "delete"
	switch resourceType {
	case "subnet":
		if del {
			return PermV2SubnetDelete, true
		}
		return PermV2SubnetWrite, true
	case "ip_address":
		if del {
			return PermV2IPRelease, true
		}
		return PermV2IPAssign, true
	case "vlan":
		if del {
			return PermV2VLANDelete, true
		}
		return PermV2VLANWrite, true
	case "device":
		if del {
			return PermV2DeviceDelete, true
		}
		return PermV2DeviceWrite, true
	}
	return "", false
}

// IntentReadPermission returns the permission needed to read a resource type's
// intent history.
func IntentReadPermission(resourceType string) (string, bool) {
	switch resourceType {
	case "subnet":
		return PermV2SubnetRead, true
	case "ip_address":
		return PermV2IPRead, true
	case "vlan":
		return PermV2VLANRead, true
	case "device":
		return PermV2DeviceRead, true
	}
	return "", false
}

// SubmitIntent records a desired-state change in the caller's organization.
// When auto-approve is enabled the intent is applied synchronously and
// transitions straight to "applied". Pass forceManual=true to skip auto-approve
// and always leave the intent pending (used when a policy decision requires
// manual review).
func (s *IntentService) SubmitIntent(ctx context.Context, resourceType string, resourceID *int64, operation string, desiredState map[string]any, submittedBy *int64, forceManual ...bool) (*models.ResourceIntent, error) {
	if _, ok := IntentPermission(resourceType, operation); !ok {
		return nil, fmt.Errorf("resource_type must be one of subnet, ip_address, device, vlan")
	}
	switch operation {
	case "create", "update", "delete":
	default:
		return nil, fmt.Errorf("operation must be create, update, or delete")
	}
	if operation != "create" && resourceID == nil {
		return nil, fmt.Errorf("resource_id is required for %s", operation)
	}
	if desiredState == nil {
		desiredState = map[string]any{}
	}

	intent := &models.ResourceIntent{
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Operation:    operation,
		DesiredState: desiredState,
		Status:       "pending",
		SubmittedBy:  submittedBy,
	}

	created, err := s.repo.CreateIntent(ctx, intent)
	if err != nil {
		return nil, err
	}

	manual := len(forceManual) > 0 && forceManual[0]
	if !manual && s.isAutoApprove(ctx) {
		return s.approve(ctx, created.ID, submittedBy, nil, false)
	}
	return created, nil
}

// ApproveIntent moves a pending intent to approved and applies the change. Only
// one caller can win the transition, so a change is never applied twice.
func (s *IntentService) ApproveIntent(ctx context.Context, id int64, reviewerID *int64, note string) (*models.ResourceIntent, error) {
	return s.approve(ctx, id, reviewerID, optNote(note), true)
}

// RejectIntent transitions a pending intent to rejected without applying it.
func (s *IntentService) RejectIntent(ctx context.Context, id int64, reviewerID *int64, note string) (*models.ResourceIntent, error) {
	intent, err := s.repo.TransitionPendingIntent(ctx, id, "rejected", reviewerID, optNote(note), time.Now())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrIntentNotPending
		}
		return nil, err
	}
	return intent, nil
}

// GetIntent returns a single intent by ID.
func (s *IntentService) GetIntent(ctx context.Context, id int64) (*models.ResourceIntent, error) {
	return s.repo.GetIntent(ctx, id)
}

// ListIntents returns the caller's organization's intents, filtered by optional
// status, resource type and resource ID.
func (s *IntentService) ListIntents(ctx context.Context, status, resourceType string, resourceID *int64) ([]*models.ResourceIntent, error) {
	items, err := s.repo.ListIntents(ctx, status, resourceType, resourceID)
	if err != nil {
		return nil, err
	}
	if items == nil {
		return []*models.ResourceIntent{}, nil
	}
	return items, nil
}

// IntentFieldChange is one field whose desired value differs from the current one.
type IntentFieldChange struct {
	Field   string `json:"field"`
	Current any    `json:"current"`
	Desired any    `json:"desired"`
}

// CurrentVsDesired pairs a resource's authoritative state with its latest
// pending intent for diff display.
type CurrentVsDesired struct {
	Current map[string]any         `json:"current"`
	Pending *models.ResourceIntent `json:"pending,omitempty"`
	Changes []IntentFieldChange    `json:"changes"`
}

// GetCurrentVsDesired returns the resource as it is now, its latest pending
// intent, and the fields that intent would change.
func (s *IntentService) GetCurrentVsDesired(ctx context.Context, resourceType string, resourceID int64) (*CurrentVsDesired, error) {
	var (
		cur any
		err error
	)
	switch resourceType {
	case "subnet":
		cur, err = s.ipam.GetSubnet(ctx, resourceID)
	case "ip_address":
		cur, err = s.ipam.GetIPAddress(ctx, resourceID)
	case "vlan":
		cur, err = s.ipam.GetVLAN(ctx, resourceID)
	case "device":
		cur, err = s.infra.GetDevice(ctx, resourceID)
	default:
		return nil, fmt.Errorf("unsupported resource_type: %s", resourceType)
	}
	if err != nil {
		return nil, err
	}
	out := &CurrentVsDesired{Changes: []IntentFieldChange{}}
	if err := remarshal(cur, &out.Current); err != nil {
		return nil, err
	}

	pending, err := s.repo.ListIntents(ctx, "pending", resourceType, &resourceID)
	if err != nil {
		return nil, err
	}
	if len(pending) == 0 {
		return out, nil
	}
	out.Pending = pending[0]
	if out.Pending.Operation == "update" {
		keys := make([]string, 0, len(out.Pending.DesiredState))
		for k := range out.Pending.DesiredState {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			want := out.Pending.DesiredState[k]
			have := out.Current[k]
			if !jsonEqual(have, want) {
				out.Changes = append(out.Changes, IntentFieldChange{Field: k, Current: have, Desired: want})
			}
		}
	}
	return out, nil
}

func jsonEqual(a, b any) bool {
	x, _ := json.Marshal(a)
	y, _ := json.Marshal(b)
	return string(x) == string(y)
}

func optNote(note string) *string {
	if note == "" {
		return nil
	}
	return &note
}

// approve runs the policy pre-flight, then claims the intent and applies it.
// human is true when a reviewer approves explicitly, which satisfies a
// manual_review decision; an automatic approval instead leaves the intent
// pending for that review. A deny decision always blocks the change.
func (s *IntentService) approve(ctx context.Context, id int64, reviewerID *int64, note *string, human bool) (*models.ResourceIntent, error) {
	now := time.Now()
	var decision *PolicyDecision
	if s.policy != nil && !policyBypassed(ctx) {
		pending, err := s.repo.GetIntent(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, ErrIntentNotPending
			}
			return nil, err
		}
		if pending.Status != "pending" {
			return nil, ErrIntentNotPending
		}
		decision, err = s.policy.Evaluate(ctx, intentPolicyRequest(pending))
		if err != nil {
			return nil, fmt.Errorf("policy evaluation failed: %w", err)
		}
		switch {
		case decision.ReviewNeeded:
			if !human {
				return pending, nil
			}
		case !decision.Allowed:
			msg := decision.Message
			if msg == "" {
				msg = "request denied by automation policy"
			}
			if _, err := s.repo.TransitionPendingIntent(ctx, id, "rejected", reviewerID, ptrStr("blocked by policy: "+msg), now); err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return nil, err
			}
			return nil, &PolicyDeniedError{Message: msg}
		}
	}

	intent, err := s.repo.TransitionPendingIntent(ctx, id, "approved", reviewerID, note, now)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrIntentNotPending
		}
		return nil, err
	}
	res, err := s.applyIntent(ctx, intent)
	if err != nil {
		_ = s.repo.UpdateIntentStatus(ctx, id, "failed", reviewerID, ptrStr(err.Error()), &now, nil)
		return nil, fmt.Errorf("apply failed: %w", err)
	}
	applied := time.Now()
	if err := s.repo.UpdateIntentStatus(ctx, id, "applied", reviewerID, note, &now, &applied); err != nil {
		return nil, err
	}
	// The change is in: run the actions of the policies it matched, in the
	// background so they never delay or fail the approval.
	if decision != nil && len(decision.Matched) > 0 {
		go s.policy.RecordApplied(context.WithoutCancel(ctx), decision, intent.ResourceType, res.ID, res.Name)
	}
	return s.repo.GetIntent(ctx, id)
}

// appliedResource identifies the resource an applied intent created or changed.
type appliedResource struct {
	ID   int64
	Name string
}

func (s *IntentService) applyIntent(ctx context.Context, intent *models.ResourceIntent) (appliedResource, error) {
	switch intent.ResourceType {
	case "subnet":
		return s.applySubnet(ctx, intent)
	case "ip_address":
		return s.applyIPAddress(ctx, intent)
	case "device":
		return s.applyDevice(ctx, intent)
	case "vlan":
		return s.applyVLAN(ctx, intent)
	default:
		return appliedResource{}, fmt.Errorf("unsupported resource_type: %s", intent.ResourceType)
	}
}

// -- subnet --

type subnetDesiredState struct {
	NetworkID      int64              `json:"network_id"`
	NetworkAddress string             `json:"network_address"`
	PrefixLength   int                `json:"prefix_length"`
	Description    string             `json:"description"`
	Gateway        *string            `json:"gateway"`
	AutoFirst      bool               `json:"auto_first"`
	AutoLast       bool               `json:"auto_last"`
	LocationID     *int64             `json:"location_id"`
	NameserverID   *int64             `json:"nameserver_id"`
	VLANID         *int64             `json:"vlan_id"`
	CustomFields   map[string]*string `json:"custom_fields"`
}

func (s *IntentService) applySubnet(ctx context.Context, intent *models.ResourceIntent) (appliedResource, error) {
	switch intent.Operation {
	case "create":
		var ds subnetDesiredState
		if err := remarshal(intent.DesiredState, &ds); err != nil {
			return appliedResource{}, err
		}
		sub, err := s.ipam.CreateSubnet(ctx, ds.NetworkID, ds.NetworkAddress, ds.PrefixLength,
			ds.Description, ds.Gateway, ds.AutoFirst, ds.AutoLast,
			ds.LocationID, ds.NameserverID, ds.VLANID, ds.CustomFields)
		if err != nil {
			return appliedResource{}, err
		}
		return appliedResource{sub.ID, fmt.Sprintf("%s/%d", sub.NetworkAddress, sub.PrefixLength)}, nil
	case "update":
		if intent.ResourceID == nil {
			return appliedResource{}, fmt.Errorf("resource_id required for subnet update")
		}
		var ds subnetDesiredState
		if err := remarshal(intent.DesiredState, &ds); err != nil {
			return appliedResource{}, err
		}
		sub, err := s.ipam.UpdateSubnet(ctx, *intent.ResourceID, ds.Description,
			ds.Gateway, ds.AutoFirst, ds.AutoLast,
			ds.LocationID, ds.NameserverID, ds.VLANID, ds.CustomFields)
		if err != nil {
			return appliedResource{}, err
		}
		return appliedResource{sub.ID, fmt.Sprintf("%s/%d", sub.NetworkAddress, sub.PrefixLength)}, nil
	case "delete":
		if intent.ResourceID == nil {
			return appliedResource{}, fmt.Errorf("resource_id required for subnet delete")
		}
		return appliedResource{ID: *intent.ResourceID}, s.ipam.DeleteSubnet(ctx, *intent.ResourceID)
	}
	return appliedResource{}, fmt.Errorf("unknown operation: %s", intent.Operation)
}

// -- ip_address --

type ipDesiredState struct {
	SubnetID     int64              `json:"subnet_id"`
	Address      string             `json:"address"`
	Hostname     string             `json:"hostname"`
	Status       string             `json:"status"`
	TagID        *int64             `json:"tag_id"`
	MacAddress   *string            `json:"mac_address"`
	PtrRecord    *string            `json:"ptr_record"`
	DNSName      *string            `json:"dns_name"`
	CustomFields map[string]*string `json:"custom_fields"`
}

func (s *IntentService) applyIPAddress(ctx context.Context, intent *models.ResourceIntent) (appliedResource, error) {
	switch intent.Operation {
	case "create":
		var ds ipDesiredState
		if err := remarshal(intent.DesiredState, &ds); err != nil {
			return appliedResource{}, err
		}
		if ds.Status == "" {
			ds.Status = "active"
		}
		ip, err := s.ipam.CreateIPAddress(ctx, ds.SubnetID, ds.Address, ds.Hostname,
			ds.Status, ds.TagID, ds.MacAddress, ds.PtrRecord, ds.DNSName, ds.CustomFields)
		if err != nil {
			return appliedResource{}, err
		}
		return appliedResource{ip.ID, ip.Address}, nil
	case "update":
		if intent.ResourceID == nil {
			return appliedResource{}, fmt.Errorf("resource_id required for ip_address update")
		}
		var ds ipDesiredState
		if err := remarshal(intent.DesiredState, &ds); err != nil {
			return appliedResource{}, err
		}
		ip, err := s.ipam.UpdateIPAddressMeta(ctx, *intent.ResourceID, ds.Hostname,
			ds.TagID, ds.MacAddress, ds.PtrRecord, ds.DNSName, ds.CustomFields)
		if err != nil {
			return appliedResource{}, err
		}
		return appliedResource{ip.ID, ip.Address}, nil
	case "delete":
		if intent.ResourceID == nil {
			return appliedResource{}, fmt.Errorf("resource_id required for ip_address delete")
		}
		return appliedResource{ID: *intent.ResourceID}, s.ipam.DeleteIPAddress(ctx, *intent.ResourceID)
	}
	return appliedResource{}, fmt.Errorf("unknown operation: %s", intent.Operation)
}

// -- device --

func (s *IntentService) applyDevice(ctx context.Context, intent *models.ResourceIntent) (appliedResource, error) {
	switch intent.Operation {
	case "create":
		var req DeviceCreateRequest
		if err := remarshal(intent.DesiredState, &req); err != nil {
			return appliedResource{}, err
		}
		d, err := s.infra.CreateDevice(ctx, &req)
		if err != nil {
			return appliedResource{}, err
		}
		return appliedResource{d.ID, d.Hostname}, nil
	case "update":
		if intent.ResourceID == nil {
			return appliedResource{}, fmt.Errorf("resource_id required for device update")
		}
		var req DeviceUpdateRequest
		if err := remarshal(intent.DesiredState, &req); err != nil {
			return appliedResource{}, err
		}
		d, err := s.infra.UpdateDevice(ctx, *intent.ResourceID, &req)
		if err != nil {
			return appliedResource{}, err
		}
		return appliedResource{d.ID, d.Hostname}, nil
	case "delete":
		if intent.ResourceID == nil {
			return appliedResource{}, fmt.Errorf("resource_id required for device delete")
		}
		return appliedResource{ID: *intent.ResourceID}, s.infra.DeleteDevice(ctx, *intent.ResourceID)
	}
	return appliedResource{}, fmt.Errorf("unknown operation: %s", intent.Operation)
}

// -- vlan --

type vlanDesiredState struct {
	VRFID       *int64 `json:"vrf_id"`
	DomainID    *int64 `json:"domain_id"`
	GroupID     *int64 `json:"group_id"`
	VlanID      int    `json:"vlan_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (s *IntentService) applyVLAN(ctx context.Context, intent *models.ResourceIntent) (appliedResource, error) {
	switch intent.Operation {
	case "create":
		var ds vlanDesiredState
		if err := remarshal(intent.DesiredState, &ds); err != nil {
			return appliedResource{}, err
		}
		v, err := s.ipam.CreateVLAN(ctx, ds.VRFID, ds.DomainID, ds.GroupID, ds.VlanID, ds.Name, ds.Description)
		if err != nil {
			return appliedResource{}, err
		}
		return appliedResource{v.ID, v.Name}, nil
	case "update":
		if intent.ResourceID == nil {
			return appliedResource{}, fmt.Errorf("resource_id required for vlan update")
		}
		var ds vlanDesiredState
		if err := remarshal(intent.DesiredState, &ds); err != nil {
			return appliedResource{}, err
		}
		v, err := s.ipam.UpdateVLAN(ctx, *intent.ResourceID, ds.DomainID, ds.GroupID, ds.Name, ds.Description)
		if err != nil {
			return appliedResource{}, err
		}
		return appliedResource{v.ID, v.Name}, nil
	case "delete":
		if intent.ResourceID == nil {
			return appliedResource{}, fmt.Errorf("resource_id required for vlan delete")
		}
		return appliedResource{ID: *intent.ResourceID}, s.ipam.DeleteVLAN(ctx, *intent.ResourceID)
	}
	return appliedResource{}, fmt.Errorf("unknown operation: %s", intent.Operation)
}

// remarshal round-trips a map[string]any into a typed struct via JSON.
func remarshal(src any, dst any) error {
	b, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dst)
}

func ptrStr(s string) *string { return &s }
