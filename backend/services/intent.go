package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"padduck/models"
)

type intentRepo interface {
	CreateIntent(ctx context.Context, intent *models.ResourceIntent) (*models.ResourceIntent, error)
	GetIntent(ctx context.Context, id int64) (*models.ResourceIntent, error)
	ListIntents(ctx context.Context, orgID *int64, status, resourceType string) ([]*models.ResourceIntent, error)
	UpdateIntentStatus(ctx context.Context, id int64, status string, reviewedBy *int64, note *string, reviewedAt, appliedAt *time.Time) error
}

type intentIPAM interface {
	CreateSubnet(ctx context.Context, networkID int64, networkAddress string, prefixLength int, description string, gateway *string, autoFirst, autoLast bool, locationID *int64, nameserverID *int64, vlanID *int64, customFields ...map[string]*string) (*models.Subnet, error)
	UpdateSubnet(ctx context.Context, id int64, description string, gateway *string, autoFirst, autoLast bool, locationID *int64, nameserverID *int64, vlanID *int64, customFields map[string]*string, technitiumScopeName ...string) (*models.Subnet, error)
	DeleteSubnet(ctx context.Context, id int64) error
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
}

type IntentService struct {
	repo   intentRepo
	config *ConfigService
	ipam   intentIPAM
	infra  intentInfra
}

func NewIntentService(repo intentRepo, config *ConfigService, ipam intentIPAM, infra intentInfra) *IntentService {
	return &IntentService{repo: repo, config: config, ipam: ipam, infra: infra}
}

func (s *IntentService) isAutoApprove(ctx context.Context) bool {
	v, err := s.config.GetCtx(ctx, "intent_auto_approve")
	if err != nil || v == "" || v == "true" {
		return true
	}
	return false
}

// SubmitIntent records a desired-state change. When auto-approve is enabled the
// intent is applied synchronously and transitions straight to "applied".
func (s *IntentService) SubmitIntent(ctx context.Context, orgID *int64, resourceType string, resourceID *int64, operation string, desiredState map[string]any, submittedBy *int64) (*models.ResourceIntent, error) {
	if resourceType == "" {
		return nil, fmt.Errorf("resource_type is required")
	}
	switch operation {
	case "create", "update", "delete":
	default:
		return nil, fmt.Errorf("operation must be create, update, or delete")
	}
	if operation != "create" && resourceID == nil {
		return nil, fmt.Errorf("resource_id is required for %s", operation)
	}

	intent := &models.ResourceIntent{
		OrganizationID: orgID,
		ResourceType:   resourceType,
		ResourceID:     resourceID,
		Operation:      operation,
		DesiredState:   desiredState,
		Status:         "pending",
		SubmittedBy:    submittedBy,
	}

	created, err := s.repo.CreateIntent(ctx, intent)
	if err != nil {
		return nil, err
	}

	if s.isAutoApprove(ctx) {
		return s.approve(ctx, created, submittedBy, nil)
	}
	return created, nil
}

// ApproveIntent transitions an intent to approved and applies the change.
func (s *IntentService) ApproveIntent(ctx context.Context, id int64, reviewerID *int64, note string) (*models.ResourceIntent, error) {
	intent, err := s.repo.GetIntent(ctx, id)
	if err != nil {
		return nil, err
	}
	if intent.Status != "pending" {
		return nil, fmt.Errorf("intent is not pending (status: %s)", intent.Status)
	}
	var notePtr *string
	if note != "" {
		notePtr = &note
	}
	return s.approve(ctx, intent, reviewerID, notePtr)
}

// RejectIntent transitions a pending intent to rejected.
func (s *IntentService) RejectIntent(ctx context.Context, id int64, reviewerID *int64, note string) (*models.ResourceIntent, error) {
	intent, err := s.repo.GetIntent(ctx, id)
	if err != nil {
		return nil, err
	}
	if intent.Status != "pending" {
		return nil, fmt.Errorf("intent is not pending (status: %s)", intent.Status)
	}
	now := time.Now()
	var notePtr *string
	if note != "" {
		notePtr = &note
	}
	if err := s.repo.UpdateIntentStatus(ctx, id, "rejected", reviewerID, notePtr, &now, nil); err != nil {
		return nil, err
	}
	return s.repo.GetIntent(ctx, id)
}

// GetIntent returns a single intent by ID.
func (s *IntentService) GetIntent(ctx context.Context, id int64) (*models.ResourceIntent, error) {
	return s.repo.GetIntent(ctx, id)
}

// ListIntents returns intents filtered by optional org, status, and resource_type.
func (s *IntentService) ListIntents(ctx context.Context, orgID *int64, status, resourceType string) ([]*models.ResourceIntent, error) {
	items, err := s.repo.ListIntents(ctx, orgID, status, resourceType)
	if err != nil {
		return nil, err
	}
	if items == nil {
		return []*models.ResourceIntent{}, nil
	}
	return items, nil
}

func (s *IntentService) approve(ctx context.Context, intent *models.ResourceIntent, reviewerID *int64, note *string) (*models.ResourceIntent, error) {
	now := time.Now()
	if err := s.applyIntent(ctx, intent); err != nil {
		_ = s.repo.UpdateIntentStatus(ctx, intent.ID, "failed", reviewerID, ptrStr(err.Error()), &now, nil)
		return nil, fmt.Errorf("apply failed: %w", err)
	}
	if err := s.repo.UpdateIntentStatus(ctx, intent.ID, "applied", reviewerID, note, &now, &now); err != nil {
		return nil, err
	}
	return s.repo.GetIntent(ctx, intent.ID)
}

func (s *IntentService) applyIntent(ctx context.Context, intent *models.ResourceIntent) error {
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
		return fmt.Errorf("unsupported resource_type: %s", intent.ResourceType)
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

func (s *IntentService) applySubnet(ctx context.Context, intent *models.ResourceIntent) error {
	switch intent.Operation {
	case "create":
		var ds subnetDesiredState
		if err := remarshal(intent.DesiredState, &ds); err != nil {
			return err
		}
		_, err := s.ipam.CreateSubnet(ctx, ds.NetworkID, ds.NetworkAddress, ds.PrefixLength,
			ds.Description, ds.Gateway, ds.AutoFirst, ds.AutoLast,
			ds.LocationID, ds.NameserverID, ds.VLANID, ds.CustomFields)
		return err
	case "update":
		if intent.ResourceID == nil {
			return fmt.Errorf("resource_id required for subnet update")
		}
		var ds subnetDesiredState
		if err := remarshal(intent.DesiredState, &ds); err != nil {
			return err
		}
		_, err := s.ipam.UpdateSubnet(ctx, *intent.ResourceID, ds.Description,
			ds.Gateway, ds.AutoFirst, ds.AutoLast,
			ds.LocationID, ds.NameserverID, ds.VLANID, ds.CustomFields)
		return err
	case "delete":
		if intent.ResourceID == nil {
			return fmt.Errorf("resource_id required for subnet delete")
		}
		return s.ipam.DeleteSubnet(ctx, *intent.ResourceID)
	}
	return fmt.Errorf("unknown operation: %s", intent.Operation)
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

func (s *IntentService) applyIPAddress(ctx context.Context, intent *models.ResourceIntent) error {
	switch intent.Operation {
	case "create":
		var ds ipDesiredState
		if err := remarshal(intent.DesiredState, &ds); err != nil {
			return err
		}
		if ds.Status == "" {
			ds.Status = "active"
		}
		_, err := s.ipam.CreateIPAddress(ctx, ds.SubnetID, ds.Address, ds.Hostname,
			ds.Status, ds.TagID, ds.MacAddress, ds.PtrRecord, ds.DNSName, ds.CustomFields)
		return err
	case "update":
		if intent.ResourceID == nil {
			return fmt.Errorf("resource_id required for ip_address update")
		}
		var ds ipDesiredState
		if err := remarshal(intent.DesiredState, &ds); err != nil {
			return err
		}
		_, err := s.ipam.UpdateIPAddressMeta(ctx, *intent.ResourceID, ds.Hostname,
			ds.TagID, ds.MacAddress, ds.PtrRecord, ds.DNSName, ds.CustomFields)
		return err
	case "delete":
		if intent.ResourceID == nil {
			return fmt.Errorf("resource_id required for ip_address delete")
		}
		return s.ipam.DeleteIPAddress(ctx, *intent.ResourceID)
	}
	return fmt.Errorf("unknown operation: %s", intent.Operation)
}

// -- device --

func (s *IntentService) applyDevice(ctx context.Context, intent *models.ResourceIntent) error {
	switch intent.Operation {
	case "create":
		var req DeviceCreateRequest
		if err := remarshal(intent.DesiredState, &req); err != nil {
			return err
		}
		_, err := s.infra.CreateDevice(ctx, &req)
		return err
	case "update":
		if intent.ResourceID == nil {
			return fmt.Errorf("resource_id required for device update")
		}
		var req DeviceUpdateRequest
		if err := remarshal(intent.DesiredState, &req); err != nil {
			return err
		}
		_, err := s.infra.UpdateDevice(ctx, *intent.ResourceID, &req)
		return err
	case "delete":
		if intent.ResourceID == nil {
			return fmt.Errorf("resource_id required for device delete")
		}
		return s.infra.DeleteDevice(ctx, *intent.ResourceID)
	}
	return fmt.Errorf("unknown operation: %s", intent.Operation)
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

func (s *IntentService) applyVLAN(ctx context.Context, intent *models.ResourceIntent) error {
	switch intent.Operation {
	case "create":
		var ds vlanDesiredState
		if err := remarshal(intent.DesiredState, &ds); err != nil {
			return err
		}
		_, err := s.ipam.CreateVLAN(ctx, ds.VRFID, ds.DomainID, ds.GroupID, ds.VlanID, ds.Name, ds.Description)
		return err
	case "update":
		if intent.ResourceID == nil {
			return fmt.Errorf("resource_id required for vlan update")
		}
		var ds vlanDesiredState
		if err := remarshal(intent.DesiredState, &ds); err != nil {
			return err
		}
		_, err := s.ipam.UpdateVLAN(ctx, *intent.ResourceID, ds.DomainID, ds.GroupID, ds.Name, ds.Description)
		return err
	case "delete":
		if intent.ResourceID == nil {
			return fmt.Errorf("resource_id required for vlan delete")
		}
		return s.ipam.DeleteVLAN(ctx, *intent.ResourceID)
	}
	return fmt.Errorf("unknown operation: %s", intent.Operation)
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
