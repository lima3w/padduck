package services

import (
	"context"
	"fmt"

	"ipam-next/models"
	"ipam-next/repository"
	"ipam-next/utils"
)

// ListDeviceTypes returns all device types.
func (s *Service) ListDeviceTypes(ctx context.Context) ([]*models.DeviceType, error) {
	return s.repository.ListDeviceTypes(ctx)
}

// ListDevices returns a paginated list of devices.
func (s *Service) ListDevices(ctx context.Context, limit, offset int) ([]*models.Device, int64, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repository.ListDevices(ctx, limit, offset)
}

// CreateDevice creates a new device, encrypting SNMP credentials before storage.
func (s *Service) CreateDevice(ctx context.Context, req *DeviceCreateRequest) (*models.Device, error) {
	if req.Hostname == "" {
		return nil, fmt.Errorf("hostname is required")
	}

	if err := s.encryptDeviceCredentials(req); err != nil {
		return nil, err
	}

	device, err := s.repository.CreateDevice(ctx, req)
	if err != nil {
		return nil, err
	}

	if req.CustomFields != nil {
		_ = s.SetCustomFieldValues(ctx, "device", device.ID, req.CustomFields)
		device.CustomFields, _ = s.repository.GetCustomFieldValues(ctx, "device", device.ID)
	}

	return device, nil
}

// GetDevice retrieves a device by ID (without SNMP credentials).
func (s *Service) GetDevice(ctx context.Context, id int64) (*models.Device, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid device ID")
	}
	device, err := s.repository.GetDeviceByID(ctx, id)
	if err != nil {
		return nil, err
	}
	device.CustomFields, _ = s.repository.GetCustomFieldValues(ctx, "device", id)
	return device, nil
}

// UpdateDevice updates an existing device, encrypting SNMP credentials before storage.
func (s *Service) UpdateDevice(ctx context.Context, id int64, req *DeviceUpdateRequest) (*models.Device, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid device ID")
	}
	if req.Hostname == "" {
		return nil, fmt.Errorf("hostname is required")
	}

	if err := s.encryptDeviceUpdateCredentials(req); err != nil {
		return nil, err
	}

	device, err := s.repository.UpdateDevice(ctx, id, req)
	if err != nil {
		return nil, err
	}

	if req.CustomFields != nil {
		_ = s.SetCustomFieldValues(ctx, "device", device.ID, req.CustomFields)
	}
	device.CustomFields, _ = s.repository.GetCustomFieldValues(ctx, "device", device.ID)
	return device, nil
}

// DeleteDevice deletes a device by ID.
func (s *Service) DeleteDevice(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid device ID")
	}
	return s.repository.DeleteDevice(ctx, id)
}

// GetDeviceSNMPCredentials retrieves and decrypts SNMP credentials for a device.
func (s *Service) GetDeviceSNMPCredentials(ctx context.Context, id int64) (*models.DeviceSNMP, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid device ID")
	}
	creds, err := s.repository.GetDeviceSNMP(ctx, id)
	if err != nil {
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
func (s *Service) ListDeviceIPAddresses(ctx context.Context, deviceID int64) ([]*models.IPAddress, error) {
	if deviceID <= 0 {
		return nil, fmt.Errorf("invalid device ID")
	}
	return s.repository.ListIPAddressesByDevice(ctx, deviceID)
}

// AssociateIPToDevice links an IP address to a device with optional interface name and primary flag.
func (s *Service) AssociateIPToDevice(ctx context.Context, deviceID, ipID int64, interfaceName *string, isPrimary bool) error {
	if deviceID <= 0 {
		return fmt.Errorf("invalid device ID")
	}
	if ipID <= 0 {
		return fmt.Errorf("invalid IP address ID")
	}
	return s.repository.AssociateIPToDevice(ctx, deviceID, ipID, interfaceName, isPrimary)
}

// UnlinkIPFromDevice removes the association between an IP address and a device.
func (s *Service) UnlinkIPFromDevice(ctx context.Context, deviceID, ipID int64) error {
	if deviceID <= 0 {
		return fmt.Errorf("invalid device ID")
	}
	if ipID <= 0 {
		return fmt.Errorf("invalid IP address ID")
	}
	return s.repository.UnlinkIPFromDevice(ctx, deviceID, ipID)
}

// ListDeviceInterfaces returns all interfaces for a device.
func (s *Service) ListDeviceInterfaces(ctx context.Context, deviceID int64) ([]*models.DeviceInterface, error) {
	if deviceID <= 0 {
		return nil, fmt.Errorf("invalid device ID")
	}
	return s.repository.ListDeviceInterfaces(ctx, deviceID)
}

// CreateDeviceInterface creates a new interface on a device.
// If connected_to_device_id and connected_to_interface_id are set, also updates the reverse link.
func (s *Service) CreateDeviceInterface(ctx context.Context, deviceID int64, req *DeviceInterfaceRequest) (*models.DeviceInterface, error) {
	if deviceID <= 0 {
		return nil, fmt.Errorf("invalid device ID")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("interface name is required")
	}

	iface, err := s.repository.CreateDeviceInterface(ctx, deviceID, req)
	if err != nil {
		return nil, err
	}

	// Update reverse link if connecting to another interface
	if req.ConnectedToDeviceID != nil && req.ConnectedToInterfaceID != nil {
		_ = s.repository.SetInterfaceConnection(ctx, *req.ConnectedToInterfaceID, deviceID, iface.ID)
	}

	return iface, nil
}

// UpdateDeviceInterface updates an existing interface.
// Maintains bidirectional connection links.
func (s *Service) UpdateDeviceInterface(ctx context.Context, deviceID, ifaceID int64, req *DeviceInterfaceRequest) (*models.DeviceInterface, error) {
	if deviceID <= 0 {
		return nil, fmt.Errorf("invalid device ID")
	}
	if ifaceID <= 0 {
		return nil, fmt.Errorf("invalid interface ID")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("interface name is required")
	}

	// Get old state to clear previous reverse link if needed
	old, err := s.repository.GetDeviceInterface(ctx, ifaceID)
	if err != nil {
		return nil, err
	}

	iface, err := s.repository.UpdateDeviceInterface(ctx, deviceID, ifaceID, req)
	if err != nil {
		return nil, err
	}

	// Clear old reverse link if it existed and connection changed
	if old.ConnectedToInterfaceID != nil {
		oldRevID := *old.ConnectedToInterfaceID
		newRevID := req.ConnectedToInterfaceID
		if newRevID == nil || *newRevID != oldRevID {
			_ = s.repository.ClearInterfaceConnection(ctx, oldRevID)
		}
	}

	// Set new reverse link
	if req.ConnectedToDeviceID != nil && req.ConnectedToInterfaceID != nil {
		_ = s.repository.SetInterfaceConnection(ctx, *req.ConnectedToInterfaceID, deviceID, ifaceID)
	}

	return iface, nil
}

// DeleteDeviceInterface deletes an interface and clears any reverse connection link.
func (s *Service) DeleteDeviceInterface(ctx context.Context, deviceID, ifaceID int64) error {
	if deviceID <= 0 {
		return fmt.Errorf("invalid device ID")
	}
	if ifaceID <= 0 {
		return fmt.Errorf("invalid interface ID")
	}

	// Get the interface to find any reverse link
	iface, err := s.repository.GetDeviceInterface(ctx, ifaceID)
	if err != nil {
		return err
	}

	if err := s.repository.DeleteDeviceInterface(ctx, deviceID, ifaceID); err != nil {
		return err
	}

	// Clear reverse link on connected interface
	if iface.ConnectedToInterfaceID != nil {
		_ = s.repository.ClearInterfaceConnection(ctx, *iface.ConnectedToInterfaceID)
	}

	return nil
}

// SearchDevices searches devices based on the provided filter criteria.
func (s *Service) SearchDevices(ctx context.Context, filter *repository.DeviceSearchFilter, cfFilters ...map[string]string) ([]*models.Device, error) {
	var cf map[string]string
	if len(cfFilters) > 0 {
		cf = cfFilters[0]
	}
	return s.repository.SearchDevicesWithCustomFields(ctx, filter, cf)
}

// encryptDeviceCredentials encrypts SNMP fields in-place on a DeviceCreateRequest.
func (s *Service) encryptDeviceCredentials(req *DeviceCreateRequest) error {
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
func (s *Service) encryptDeviceUpdateCredentials(req *DeviceUpdateRequest) error {
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

// DeviceCreateRequest holds input for creating a device (used for JSON binding in handlers).
type DeviceCreateRequest = repository.DeviceParams

// DeviceUpdateRequest holds input for updating a device (used for JSON binding in handlers).
type DeviceUpdateRequest = repository.DeviceParams

// DeviceInterfaceRequest holds input for creating or updating a device interface.
type DeviceInterfaceRequest = repository.DeviceInterfaceParams

