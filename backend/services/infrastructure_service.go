package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"padduck/models"
	"padduck/repository"
	"padduck/utils"
)

var ErrNotAssociated = errors.New("not associated")

// InfrastructureService handles devices, racks, locations, and nameservers.
type InfrastructureService struct {
	repo          *repository.Repository
	encryptionKey string
}

func NewInfrastructureService(repo *repository.Repository, encryptionKey string) *InfrastructureService {
	return &InfrastructureService{
		repo:          repo,
		encryptionKey: encryptionKey,
	}
}

// ---- Device types and request types ----

// DeviceCreateRequest holds input for creating a device (used for JSON binding in handlers).
type DeviceCreateRequest = repository.DeviceParams

// DeviceUpdateRequest holds input for updating a device (used for JSON binding in handlers).
type DeviceUpdateRequest = repository.DeviceParams

// DeviceInterfaceRequest holds input for creating or updating a device interface.
type DeviceInterfaceRequest = repository.DeviceInterfaceParams

// ---- Rack request types ----

// RackCreateRequest holds input for creating a rack.
type RackCreateRequest = repository.RackParams

// RackUpdateRequest holds input for updating a rack.
type RackUpdateRequest = repository.RackParams

// ---- Location request types ----

// LocationCreateRequest holds input for creating a location.
type LocationCreateRequest = repository.LocationParams

// LocationUpdateRequest holds input for updating a location.
type LocationUpdateRequest = repository.LocationParams

// ---- Nameserver request types ----

// NameserverCreateRequest holds input for creating a nameserver.
type NameserverCreateRequest = repository.NameserverParams

// NameserverUpdateRequest holds input for updating a nameserver.
type NameserverUpdateRequest = repository.NameserverParams

// ---- Devices ----

// ListDeviceTypes returns all device types.
func (s *InfrastructureService) ListDeviceTypes(ctx context.Context) ([]*models.DeviceType, error) {
	return s.repo.ListDeviceTypes(ctx)
}

func (s *InfrastructureService) ListDevicesWithOptions(ctx context.Context, limit, offset int, opts repository.ListOptions) ([]*models.Device, int64, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	opts.Limit = limit
	opts.Offset = offset
	return s.repo.ListDevicesWithOptions(ctx, opts)
}

// ListAllDevices returns all devices without pagination.
func (s *InfrastructureService) ListAllDevices(ctx context.Context) ([]*models.Device, error) {
	return s.repo.ListAllDevices(ctx)
}

// CreateDevice creates a new device, encrypting SNMP credentials before storage.
func (s *InfrastructureService) CreateDevice(ctx context.Context, req *DeviceCreateRequest) (*models.Device, error) {
	if req.Hostname == "" {
		return nil, fmt.Errorf("hostname is required")
	}

	if err := s.encryptDeviceCredentials(req); err != nil {
		return nil, err
	}

	device, err := s.repo.CreateDevice(ctx, req)
	if err != nil {
		return nil, err
	}

	if req.CustomFields != nil {
		_ = s.setCustomFieldValues(ctx, "device", device.ID, req.CustomFields)
		device.CustomFields, _ = s.repo.GetCustomFieldValues(ctx, "device", device.ID)
	}

	return device, nil
}

// GetDevice retrieves a device by ID (without SNMP credentials).
func (s *InfrastructureService) GetDevice(ctx context.Context, id int64) (*models.Device, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid device ID")
	}
	device, err := s.repo.GetDeviceByID(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("device %d: %w", id, ErrNotFound)
		}
		return nil, err
	}
	device.CustomFields, _ = s.repo.GetCustomFieldValues(ctx, "device", id)
	return device, nil
}

// UpdateDevice updates an existing device, encrypting SNMP credentials before storage.
func (s *InfrastructureService) UpdateDevice(ctx context.Context, id int64, req *DeviceUpdateRequest) (*models.Device, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid device ID")
	}
	if req.Hostname == "" {
		return nil, fmt.Errorf("hostname is required")
	}

	if err := s.encryptDeviceUpdateCredentials(req); err != nil {
		return nil, err
	}

	device, err := s.repo.UpdateDevice(ctx, id, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("device %d: %w", id, ErrNotFound)
		}
		return nil, err
	}

	if req.CustomFields != nil {
		_ = s.setCustomFieldValues(ctx, "device", device.ID, req.CustomFields)
	}
	device.CustomFields, _ = s.repo.GetCustomFieldValues(ctx, "device", device.ID)
	return device, nil
}

// DeleteDevice deletes a device by ID.
func (s *InfrastructureService) DeleteDevice(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid device ID")
	}
	if err := s.repo.DeleteDevice(ctx, id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("device %d: %w", id, ErrNotFound)
		}
		return err
	}
	return nil
}

// GetDeviceSNMPCredentials retrieves and decrypts SNMP credentials for a device.
func (s *InfrastructureService) GetDeviceSNMPCredentials(ctx context.Context, id int64) (*models.DeviceSNMP, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid device ID")
	}
	creds, err := s.repo.GetDeviceSNMP(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("device %d: %w", id, ErrNotFound)
		}
		return nil, err
	}

	if creds.SNMPCommunity != nil {
		dec, err := utils.DecryptString(*creds.SNMPCommunity, s.encryptionKey)
		if err == nil {
			creds.SNMPCommunity = &dec
		}
	}
	if creds.SNMPV3AuthPass != nil {
		dec, err := utils.DecryptString(*creds.SNMPV3AuthPass, s.encryptionKey)
		if err == nil {
			creds.SNMPV3AuthPass = &dec
		}
	}
	if creds.SNMPV3PrivPass != nil {
		dec, err := utils.DecryptString(*creds.SNMPV3PrivPass, s.encryptionKey)
		if err == nil {
			creds.SNMPV3PrivPass = &dec
		}
	}
	return creds, nil
}

// ListDeviceIPAddresses returns all IP addresses linked to a device.
func (s *InfrastructureService) ListDeviceIPAddresses(ctx context.Context, deviceID int64) ([]*models.IPAddress, error) {
	if deviceID <= 0 {
		return nil, fmt.Errorf("invalid device ID")
	}
	return s.repo.ListIPAddressesByDevice(ctx, deviceID)
}

// AssociateIPToDevice links an IP address to a device with optional interface name and primary flag.
func (s *InfrastructureService) AssociateIPToDevice(ctx context.Context, deviceID, ipID int64, interfaceName *string, isPrimary bool) error {
	if deviceID <= 0 {
		return fmt.Errorf("invalid device ID")
	}
	if ipID <= 0 {
		return fmt.Errorf("invalid IP address ID")
	}
	return s.repo.AssociateIPToDevice(ctx, deviceID, ipID, interfaceName, isPrimary)
}

// UnlinkIPFromDevice removes the association between an IP address and a device.
func (s *InfrastructureService) UnlinkIPFromDevice(ctx context.Context, deviceID, ipID int64) error {
	if deviceID <= 0 {
		return fmt.Errorf("invalid device ID")
	}
	if ipID <= 0 {
		return fmt.Errorf("invalid IP address ID")
	}
	if err := s.repo.UnlinkIPFromDevice(ctx, deviceID, ipID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("ip %d device %d: %w", ipID, deviceID, ErrNotAssociated)
		}
		return err
	}
	return nil
}

// ListDeviceInterfaces returns all interfaces for a device.
func (s *InfrastructureService) ListDeviceInterfaces(ctx context.Context, deviceID int64) ([]*models.DeviceInterface, error) {
	if deviceID <= 0 {
		return nil, fmt.Errorf("invalid device ID")
	}
	return s.repo.ListDeviceInterfaces(ctx, deviceID)
}

// CreateDeviceInterface creates a new interface on a device.
// If connected_to_device_id and connected_to_interface_id are set, also updates the reverse link.
func (s *InfrastructureService) CreateDeviceInterface(ctx context.Context, deviceID int64, req *DeviceInterfaceRequest) (*models.DeviceInterface, error) {
	if deviceID <= 0 {
		return nil, fmt.Errorf("invalid device ID")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("interface name is required")
	}

	iface, err := s.repo.CreateDeviceInterface(ctx, deviceID, req)
	if err != nil {
		return nil, err
	}

	if req.ConnectedToDeviceID != nil && req.ConnectedToInterfaceID != nil {
		_ = s.repo.SetInterfaceConnection(ctx, *req.ConnectedToInterfaceID, deviceID, iface.ID)
	}

	return iface, nil
}

// UpdateDeviceInterface updates an existing interface.
// Maintains bidirectional connection links.
func (s *InfrastructureService) UpdateDeviceInterface(ctx context.Context, deviceID, ifaceID int64, req *DeviceInterfaceRequest) (*models.DeviceInterface, error) {
	if deviceID <= 0 {
		return nil, fmt.Errorf("invalid device ID")
	}
	if ifaceID <= 0 {
		return nil, fmt.Errorf("invalid interface ID")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("interface name is required")
	}

	old, err := s.repo.GetDeviceInterface(ctx, ifaceID)
	if err != nil {
		return nil, err
	}

	iface, err := s.repo.UpdateDeviceInterface(ctx, deviceID, ifaceID, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("interface %d: %w", ifaceID, ErrNotFound)
		}
		return nil, err
	}

	if old.ConnectedToInterfaceID != nil {
		oldRevID := *old.ConnectedToInterfaceID
		newRevID := req.ConnectedToInterfaceID
		if newRevID == nil || *newRevID != oldRevID {
			_ = s.repo.ClearInterfaceConnection(ctx, oldRevID)
		}
	}

	if req.ConnectedToDeviceID != nil && req.ConnectedToInterfaceID != nil {
		_ = s.repo.SetInterfaceConnection(ctx, *req.ConnectedToInterfaceID, deviceID, ifaceID)
	}

	return iface, nil
}

// DeleteDeviceInterface deletes an interface and clears any reverse connection link.
func (s *InfrastructureService) DeleteDeviceInterface(ctx context.Context, deviceID, ifaceID int64) error {
	if deviceID <= 0 {
		return fmt.Errorf("invalid device ID")
	}
	if ifaceID <= 0 {
		return fmt.Errorf("invalid interface ID")
	}

	iface, err := s.repo.GetDeviceInterface(ctx, ifaceID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("interface %d: %w", ifaceID, ErrNotFound)
		}
		return err
	}

	if err := s.repo.DeleteDeviceInterface(ctx, deviceID, ifaceID); err != nil {
		return err
	}

	if iface.ConnectedToInterfaceID != nil {
		_ = s.repo.ClearInterfaceConnection(ctx, *iface.ConnectedToInterfaceID)
	}

	return nil
}

// ListDevicesByLocation returns all devices assigned to the given location.
func (s *InfrastructureService) ListDevicesByLocation(ctx context.Context, locationID int64) ([]*models.Device, error) {
	if locationID <= 0 {
		return nil, fmt.Errorf("invalid location ID")
	}
	return s.repo.ListDevicesByLocation(ctx, locationID)
}

// SearchDevices searches devices based on the provided filter criteria.
func (s *InfrastructureService) SearchDevices(ctx context.Context, filter *repository.DeviceSearchFilter, cfFilters ...map[string]string) ([]*models.Device, error) {
	var cf map[string]string
	if len(cfFilters) > 0 {
		cf = cfFilters[0]
	}
	return s.repo.SearchDevicesWithCustomFields(ctx, filter, cf)
}

// encryptDeviceCredentials encrypts SNMP fields in-place on a DeviceCreateRequest.
func (s *InfrastructureService) encryptDeviceCredentials(req *DeviceCreateRequest) error {
	if req.SNMPCommunity != nil && *req.SNMPCommunity != "" {
		enc, err := utils.EncryptString(*req.SNMPCommunity, s.encryptionKey)
		if err != nil {
			return fmt.Errorf("encrypting snmp_community: %w", err)
		}
		req.SNMPCommunity = &enc
	}
	if req.SNMPV3AuthPass != nil && *req.SNMPV3AuthPass != "" {
		enc, err := utils.EncryptString(*req.SNMPV3AuthPass, s.encryptionKey)
		if err != nil {
			return fmt.Errorf("encrypting snmp_v3_auth_pass: %w", err)
		}
		req.SNMPV3AuthPass = &enc
	}
	if req.SNMPV3PrivPass != nil && *req.SNMPV3PrivPass != "" {
		enc, err := utils.EncryptString(*req.SNMPV3PrivPass, s.encryptionKey)
		if err != nil {
			return fmt.Errorf("encrypting snmp_v3_priv_pass: %w", err)
		}
		req.SNMPV3PrivPass = &enc
	}
	return nil
}

// encryptDeviceUpdateCredentials encrypts SNMP fields in-place on a DeviceUpdateRequest.
func (s *InfrastructureService) encryptDeviceUpdateCredentials(req *DeviceUpdateRequest) error {
	if req.SNMPCommunity != nil && *req.SNMPCommunity != "" {
		enc, err := utils.EncryptString(*req.SNMPCommunity, s.encryptionKey)
		if err != nil {
			return fmt.Errorf("encrypting snmp_community: %w", err)
		}
		req.SNMPCommunity = &enc
	}
	if req.SNMPV3AuthPass != nil && *req.SNMPV3AuthPass != "" {
		enc, err := utils.EncryptString(*req.SNMPV3AuthPass, s.encryptionKey)
		if err != nil {
			return fmt.Errorf("encrypting snmp_v3_auth_pass: %w", err)
		}
		req.SNMPV3AuthPass = &enc
	}
	if req.SNMPV3PrivPass != nil && *req.SNMPV3PrivPass != "" {
		enc, err := utils.EncryptString(*req.SNMPV3PrivPass, s.encryptionKey)
		if err != nil {
			return fmt.Errorf("encrypting snmp_v3_priv_pass: %w", err)
		}
		req.SNMPV3PrivPass = &enc
	}
	return nil
}

// setCustomFieldValues validates required fields and writes custom field values for an entity.
// Mirrors the logic of Service.SetCustomFieldValues but uses s.repo directly.
func (s *InfrastructureService) setCustomFieldValues(ctx context.Context, entityType string, entityID int64, values map[string]*string) error {
	defs, err := s.repo.ListCustomFieldDefinitions(ctx, entityType)
	if err != nil {
		return err
	}
	for _, def := range defs {
		if def.IsRequired {
			val, ok := values[def.Name]
			if !ok || val == nil || *val == "" {
				return fmt.Errorf("field %q is required", def.Name)
			}
		}
	}
	return s.repo.SetCustomFieldValues(ctx, entityType, entityID, defs, values)
}

// ---- Racks ----

// CreateRack creates a new rack.
func (s *InfrastructureService) CreateRack(ctx context.Context, req *RackCreateRequest) (*models.Rack, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("rack name is required")
	}
	if req.SizeU <= 0 {
		req.SizeU = 42
	}
	return s.repo.CreateRack(ctx, req)
}

// GetRack retrieves a rack by ID.
func (s *InfrastructureService) GetRack(ctx context.Context, id int64) (*models.Rack, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid rack ID")
	}
	rack, err := s.repo.GetRackByID(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("rack %d: %w", id, ErrNotFound)
		}
		return nil, err
	}
	return rack, nil
}

// ListRacks returns all racks, optionally filtered by location.
func (s *InfrastructureService) ListRacks(ctx context.Context, locationID *int64) ([]*models.Rack, error) {
	return s.repo.ListRacks(ctx, locationID)
}

// UpdateRack updates an existing rack.
func (s *InfrastructureService) UpdateRack(ctx context.Context, id int64, req *RackUpdateRequest) (*models.Rack, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid rack ID")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("rack name is required")
	}
	if req.SizeU <= 0 {
		req.SizeU = 42
	}
	rack, err := s.repo.UpdateRack(ctx, id, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("rack %d: %w", id, ErrNotFound)
		}
		return nil, err
	}
	return rack, nil
}

// DeleteRack deletes a rack by ID.
func (s *InfrastructureService) DeleteRack(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid rack ID")
	}
	if err := s.repo.DeleteRack(ctx, id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("rack %d: %w", id, ErrNotFound)
		}
		return err
	}
	return nil
}

// ListDevicesInRack returns all devices assigned to a rack, ordered by rack_unit_start.
func (s *InfrastructureService) ListDevicesInRack(ctx context.Context, rackID int64) ([]*models.Device, error) {
	if rackID <= 0 {
		return nil, fmt.Errorf("invalid rack ID")
	}
	return s.repo.ListDevicesInRack(ctx, rackID)
}

// ---- Locations ----

// CreateLocation creates a new location.
func (s *InfrastructureService) CreateLocation(ctx context.Context, req *LocationCreateRequest) (*models.Location, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("location name is required")
	}
	if req.Type == "" {
		req.Type = "other"
	}
	if req.Status == "" {
		req.Status = "active"
	}
	return s.repo.CreateLocation(ctx, req)
}

// GetLocation retrieves a location by ID.
func (s *InfrastructureService) GetLocation(ctx context.Context, id int64) (*models.Location, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid location ID")
	}
	loc, err := s.repo.GetLocationByID(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("location %d: %w", id, ErrNotFound)
		}
		return nil, err
	}
	return loc, nil
}

// ListLocations returns all locations.
func (s *InfrastructureService) ListLocations(ctx context.Context) ([]*models.Location, error) {
	return s.repo.ListLocations(ctx)
}

// ListLocationsPaginated returns a paginated list of locations.
func (s *InfrastructureService) ListLocationsPaginated(ctx context.Context, page, limit int) ([]*models.Location, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 25
	}
	offset := (page - 1) * limit
	return s.repo.ListLocationsPaginated(ctx, limit, offset)
}

// UpdateLocation updates an existing location.
func (s *InfrastructureService) UpdateLocation(ctx context.Context, id int64, req *LocationUpdateRequest) (*models.Location, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid location ID")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("location name is required")
	}
	if req.Type == "" {
		req.Type = "other"
	}
	if req.Status == "" {
		req.Status = "active"
	}
	loc, err := s.repo.UpdateLocation(ctx, id, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("location %d: %w", id, ErrNotFound)
		}
		return nil, err
	}
	return loc, nil
}

// DeleteLocation deletes a location by ID.
func (s *InfrastructureService) DeleteLocation(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid location ID")
	}
	if err := s.repo.DeleteLocation(ctx, id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("location %d: %w", id, ErrNotFound)
		}
		return err
	}
	return nil
}

// GetLocationTree returns all locations assembled into a nested tree.
func (s *InfrastructureService) GetLocationTree(ctx context.Context) ([]*models.LocationTreeNode, error) {
	locs, err := s.repo.GetLocationTree(ctx)
	if err != nil {
		return nil, err
	}
	return buildLocationTree(locs), nil
}

// buildLocationTree assembles a flat location list into a nested tree.
func buildLocationTree(locs []*models.Location) []*models.LocationTreeNode {
	nodeMap := make(map[int64]*models.LocationTreeNode, len(locs))
	for _, l := range locs {
		nodeMap[l.ID] = &models.LocationTreeNode{Location: *l, Children: []*models.LocationTreeNode{}}
	}
	roots := make([]*models.LocationTreeNode, 0)
	for _, l := range locs {
		node := nodeMap[l.ID]
		if l.ParentID == nil {
			roots = append(roots, node)
		} else {
			if parent, ok := nodeMap[*l.ParentID]; ok {
				parent.Children = append(parent.Children, node)
			}
		}
	}
	return roots
}

// ---- Nameservers ----

// CreateNameserver creates a new nameserver entry.
func (s *InfrastructureService) CreateNameserver(ctx context.Context, req *NameserverCreateRequest) (*models.Nameserver, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("nameserver name is required")
	}
	if req.Server1 == "" {
		return nil, fmt.Errorf("server1 is required")
	}
	return s.repo.CreateNameserver(ctx, req)
}

// GetNameserver retrieves a nameserver by ID.
func (s *InfrastructureService) GetNameserver(ctx context.Context, id int64) (*models.Nameserver, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid nameserver ID")
	}
	ns, err := s.repo.GetNameserverByID(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("nameserver %d: %w", id, ErrNotFound)
		}
		return nil, err
	}
	return ns, nil
}

// ListNameservers returns all nameservers.
func (s *InfrastructureService) ListNameservers(ctx context.Context) ([]*models.Nameserver, error) {
	ns, err := s.repo.ListNameservers(ctx)
	if err != nil {
		return nil, err
	}
	if ns == nil {
		ns = []*models.Nameserver{}
	}
	return ns, nil
}

// UpdateNameserver updates an existing nameserver.
func (s *InfrastructureService) UpdateNameserver(ctx context.Context, id int64, req *NameserverUpdateRequest) (*models.Nameserver, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid nameserver ID")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("nameserver name is required")
	}
	if req.Server1 == "" {
		return nil, fmt.Errorf("server1 is required")
	}
	ns, err := s.repo.UpdateNameserver(ctx, id, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("nameserver %d: %w", id, ErrNotFound)
		}
		return nil, err
	}
	return ns, nil
}

// DeleteNameserver deletes a nameserver by ID.
func (s *InfrastructureService) DeleteNameserver(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid nameserver ID")
	}
	if err := s.repo.DeleteNameserver(ctx, id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("nameserver %d: %w", id, ErrNotFound)
		}
		return err
	}
	return nil
}
