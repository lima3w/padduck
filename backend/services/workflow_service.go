package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"padduck/models"
	"padduck/repository"
)

var (
	ErrNotFound       = errors.New("not found")
	ErrNotPending     = errors.New("request is not in pending state")
	ErrNotCancellable = errors.New("request cannot be cancelled")
)

var validEntityTypes = map[string]bool{
	"subnet": true, "ip_address": true, "device": true,
}

var validFieldTypes = map[string]bool{
	"text": true, "number": true, "textarea": true, "dropdown": true,
	"checkbox": true, "date": true, "url": true, "email": true,
}

type WorkflowService struct {
	repo         *repository.Repository
	ipam         *IPAMService
	dns          *DNSService
	audit        *AuditService
	notification *NotificationService
}

func NewWorkflowService(repo *repository.Repository, ipam *IPAMService, dns *DNSService, audit *AuditService, notification *NotificationService) *WorkflowService {
	return &WorkflowService{repo: repo, ipam: ipam, dns: dns, audit: audit, notification: notification}
}

// ---- Subnet Requests ----

func (s *WorkflowService) SubmitSubnetRequest(ctx context.Context, requesterID, networkID int64, parentSubnetID *int64, prefixLen int, purpose string) (*models.SubnetRequest, error) {
	if requesterID <= 0 {
		return nil, fmt.Errorf("invalid requester ID")
	}
	if networkID <= 0 {
		return nil, fmt.Errorf("invalid section ID")
	}
	if prefixLen < 0 || prefixLen > 32 {
		return nil, fmt.Errorf("requested_prefix_len must be between 0 and 32")
	}
	purpose = strings.TrimSpace(purpose)
	if purpose == "" {
		return nil, fmt.Errorf("purpose is required")
	}

	sr, err := s.repo.CreateSubnetRequest(ctx, requesterID, networkID, parentSubnetID, prefixLen, purpose)
	if err != nil {
		return nil, err
	}
	s.notifyAdminsSubnetRequest(ctx, sr, requesterID)
	return sr, nil
}

func (s *WorkflowService) ListMySubnetRequests(ctx context.Context, requesterID int64) ([]*models.SubnetRequest, error) {
	if requesterID <= 0 {
		return nil, fmt.Errorf("invalid requester ID")
	}
	return s.repo.ListSubnetRequestsByRequester(ctx, requesterID)
}

func (s *WorkflowService) ListAllSubnetRequests(ctx context.Context) ([]*models.SubnetRequest, error) {
	return s.repo.ListAllSubnetRequests(ctx)
}

func (s *WorkflowService) ApproveSubnetRequest(ctx context.Context, requestID, reviewerID int64, reviewerNote string) (*models.SubnetRequest, error) {
	if requestID <= 0 {
		return nil, fmt.Errorf("invalid request ID")
	}
	if reviewerID <= 0 {
		return nil, fmt.Errorf("invalid reviewer ID")
	}

	sr, err := s.repo.GetSubnetRequestByID(ctx, requestID)
	if err != nil {
		return nil, fmt.Errorf("subnet request %d: %w", requestID, ErrNotFound)
	}
	if sr.Status != "pending" {
		return nil, fmt.Errorf("subnet request %d: %w", requestID, ErrNotPending)
	}

	networkAddr, err := s.findFreeSubnetBlock(ctx, sr.NetworkID, sr.ParentSubnetID, sr.RequestedPrefixLen)
	if err != nil {
		return nil, fmt.Errorf("could not allocate subnet block: %w", err)
	}

	subnet, err := s.ipam.CreateSubnet(ctx, sr.NetworkID, networkAddr, sr.RequestedPrefixLen, sr.Purpose, nil, false, false, nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create subnet: %w", err)
	}

	approved, err := s.repo.ApproveSubnetRequest(ctx, requestID, reviewerID, subnet.ID, reviewerNote)
	if err != nil {
		return nil, err
	}
	s.notifyRequesterSubnetApproved(ctx, approved, subnet)
	return approved, nil
}

func (s *WorkflowService) RejectSubnetRequest(ctx context.Context, requestID, reviewerID int64, reviewerNote string) (*models.SubnetRequest, error) {
	if requestID <= 0 {
		return nil, fmt.Errorf("invalid request ID")
	}
	if reviewerID <= 0 {
		return nil, fmt.Errorf("invalid reviewer ID")
	}
	if strings.TrimSpace(reviewerNote) == "" {
		return nil, fmt.Errorf("reviewer_note is required when rejecting")
	}

	rejected, err := s.repo.RejectSubnetRequest(ctx, requestID, reviewerID, reviewerNote)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "not found") {
			return nil, fmt.Errorf("subnet request %d: %w", requestID, ErrNotFound)
		}
		if strings.Contains(msg, "not pending") {
			return nil, fmt.Errorf("subnet request %d: %w", requestID, ErrNotPending)
		}
		return nil, err
	}
	s.notifyRequesterSubnetRejected(ctx, rejected)
	return rejected, nil
}

func (s *WorkflowService) CancelSubnetRequest(ctx context.Context, requestID, requesterID int64) error {
	if requestID <= 0 {
		return fmt.Errorf("invalid request ID")
	}
	if requesterID <= 0 {
		return fmt.Errorf("invalid requester ID")
	}
	if err := s.repo.CancelSubnetRequest(ctx, requestID, requesterID); err != nil {
		msg := err.Error()
		if strings.Contains(msg, "not found") {
			return fmt.Errorf("subnet request %d: %w", requestID, ErrNotFound)
		}
		if strings.Contains(msg, "not cancellable") {
			return fmt.Errorf("subnet request %d: %w", requestID, ErrNotCancellable)
		}
		return err
	}
	return nil
}

func (s *WorkflowService) findFreeSubnetBlock(ctx context.Context, networkID int64, parentSubnetID *int64, prefixLen int) (string, error) {
	existing, err := s.repo.ListSubnetsBySection(ctx, networkID)
	if err != nil {
		return "", err
	}
	if parentSubnetID != nil {
		parent, err := s.repo.GetSubnetByID(ctx, *parentSubnetID)
		if err != nil {
			return "", fmt.Errorf("parent subnet %d: %w", *parentSubnetID, ErrNotFound)
		}
		return findFirstFreeBlock(parent.NetworkAddress, parent.PrefixLength, prefixLen, existing)
	}
	return findFirstFreeBlock("10.0.0.0", 8, prefixLen, existing)
}

// ---- IP Requests ----

func (s *WorkflowService) SubmitIPRequest(ctx context.Context, requesterID, subnetID int64, requestedIP *string, dnsName, purpose string) (*models.IPRequest, error) {
	if requesterID <= 0 {
		return nil, fmt.Errorf("invalid requester ID")
	}
	if subnetID <= 0 {
		return nil, fmt.Errorf("invalid subnet ID")
	}
	purpose = strings.TrimSpace(purpose)
	if purpose == "" {
		return nil, fmt.Errorf("purpose is required")
	}

	ir, err := s.repo.CreateIPRequest(ctx, requesterID, subnetID, requestedIP, strings.TrimSpace(dnsName), purpose)
	if err != nil {
		return nil, err
	}
	s.notifyAdminsIPRequest(ctx, ir, requesterID)
	return ir, nil
}

func (s *WorkflowService) ListMyIPRequests(ctx context.Context, requesterID int64) ([]*models.IPRequest, error) {
	if requesterID <= 0 {
		return nil, fmt.Errorf("invalid requester ID")
	}
	return s.repo.ListIPRequestsByRequester(ctx, requesterID)
}

func (s *WorkflowService) ListAllIPRequests(ctx context.Context) ([]*models.IPRequest, error) {
	return s.repo.ListAllIPRequests(ctx)
}

func (s *WorkflowService) ApproveIPRequest(ctx context.Context, requestID, reviewerID int64, reviewerNote string) (*models.IPRequest, error) {
	if requestID <= 0 {
		return nil, fmt.Errorf("invalid request ID")
	}
	if reviewerID <= 0 {
		return nil, fmt.Errorf("invalid reviewer ID")
	}

	ir, err := s.repo.GetIPRequestByID(ctx, requestID)
	if err != nil {
		return nil, fmt.Errorf("ip request %d: %w", requestID, ErrNotFound)
	}
	if ir.Status != "pending" {
		return nil, fmt.Errorf("ip request %d: %w", requestID, ErrNotPending)
	}

	var ipAddr *models.IPAddress

	if ir.RequestedIP != nil && *ir.RequestedIP != "" {
		existing, lookupErr := s.repo.GetIPAddressBySubnetAndAddress(ctx, ir.SubnetID, *ir.RequestedIP)
		if lookupErr == nil && existing != nil && existing.Status != "available" {
			return nil, &IPAlreadyTakenError{IP: *ir.RequestedIP}
		}
		if lookupErr == nil && existing != nil {
			ipAddr, err = s.repo.UpdateIPAddressStatus(ctx, existing.ID, "assigned", nil)
			if err != nil {
				return nil, fmt.Errorf("failed to assign IP: %w", err)
			}
		} else {
			dnsName := ir.DNSName
			var dnsNamePtr *string
			if dnsName != "" {
				dnsNamePtr = &dnsName
			}
			ipAddr, err = s.repo.CreateIPAddress(ctx, ir.SubnetID, *ir.RequestedIP, dnsName, "assigned", nil, nil, nil, dnsNamePtr)
			if err != nil {
				return nil, fmt.Errorf("failed to create IP: %w", err)
			}
		}
	} else {
		ipAddr, err = s.repo.AllocateIPAddress(ctx, ir.SubnetID, nil)
		if err != nil {
			return nil, fmt.Errorf("no available IPs in subnet: %w", err)
		}
	}

	if ir.DNSName != "" {
		_ = s.repo.UpdateIPDNSName(ctx, ipAddr.ID, ir.DNSName)
		if updated, err := s.repo.GetIPAddressByID(ctx, ipAddr.ID); err == nil {
			go s.dns.SyncIPToDNS(ctx, updated)
		}
	}

	approved, err := s.repo.ApproveIPRequest(ctx, requestID, reviewerID, ipAddr.ID, reviewerNote)
	if err != nil {
		return nil, err
	}
	s.notifyRequesterIPApproved(ctx, approved, ipAddr)
	return approved, nil
}

func (s *WorkflowService) RejectIPRequest(ctx context.Context, requestID, reviewerID int64, reviewerNote string) (*models.IPRequest, error) {
	if requestID <= 0 {
		return nil, fmt.Errorf("invalid request ID")
	}
	if reviewerID <= 0 {
		return nil, fmt.Errorf("invalid reviewer ID")
	}
	if strings.TrimSpace(reviewerNote) == "" {
		return nil, fmt.Errorf("reviewer_note is required when rejecting")
	}

	rejected, err := s.repo.RejectIPRequest(ctx, requestID, reviewerID, reviewerNote)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "not found") {
			return nil, fmt.Errorf("ip request %d: %w", requestID, ErrNotFound)
		}
		if strings.Contains(msg, "not pending") {
			return nil, fmt.Errorf("ip request %d: %w", requestID, ErrNotPending)
		}
		return nil, err
	}
	s.notifyRequesterIPRejected(ctx, rejected)
	return rejected, nil
}

func (s *WorkflowService) CancelIPRequest(ctx context.Context, requestID, requesterID int64) error {
	if requestID <= 0 {
		return fmt.Errorf("invalid request ID")
	}
	if requesterID <= 0 {
		return fmt.Errorf("invalid requester ID")
	}
	if err := s.repo.CancelIPRequest(ctx, requestID, requesterID); err != nil {
		msg := err.Error()
		if strings.Contains(msg, "not found") {
			return fmt.Errorf("ip request %d: %w", requestID, ErrNotFound)
		}
		if strings.Contains(msg, "not cancellable") {
			return fmt.Errorf("ip request %d: %w", requestID, ErrNotCancellable)
		}
		return err
	}
	return nil
}

func (s *WorkflowService) GetRequestOwner(ctx context.Context, requestType string, requestID int64) (int64, error) {
	switch requestType {
	case "subnet":
		sr, err := s.repo.GetSubnetRequestByID(ctx, requestID)
		if err != nil {
			return 0, fmt.Errorf("subnet request not found")
		}
		return sr.RequesterID, nil
	case "ip":
		ir, err := s.repo.GetIPRequestByID(ctx, requestID)
		if err != nil {
			return 0, fmt.Errorf("ip request not found")
		}
		return ir.RequesterID, nil
	default:
		return 0, fmt.Errorf("invalid request_type: must be 'subnet' or 'ip'")
	}
}

func (s *WorkflowService) GetPendingRequestCounts(ctx context.Context) (subnetCount, ipCount int64, err error) {
	subnetCount, err = s.repo.CountPendingSubnetRequests(ctx)
	if err != nil {
		return
	}
	ipCount, err = s.repo.CountPendingIPRequests(ctx)
	return
}

// ---- Request Comments ----

func (s *WorkflowService) AddRequestComment(ctx context.Context, requestType string, requestID, authorID int64, body string) (*models.RequestComment, error) {
	if requestType != "subnet" && requestType != "ip" {
		return nil, fmt.Errorf("invalid request_type: must be 'subnet' or 'ip'")
	}
	if requestID <= 0 {
		return nil, fmt.Errorf("invalid request ID")
	}
	if authorID <= 0 {
		return nil, fmt.Errorf("invalid author ID")
	}
	body = strings.TrimSpace(body)
	if body == "" {
		return nil, fmt.Errorf("comment body is required")
	}

	comment, err := s.repo.CreateRequestComment(ctx, requestType, requestID, authorID, body)
	if err != nil {
		return nil, err
	}

	s.audit.Log(ctx, AuditEntry{
		UserID: &authorID, Action: "request_comment_added",
		ResourceType: requestType + "_request", ResourceID: &requestID,
		NewValues: map[string]interface{}{"body": body},
	})
	s.notifyRequestComment(ctx, requestType, requestID, authorID, comment)
	return comment, nil
}

func (s *WorkflowService) ListRequestComments(ctx context.Context, requestType string, requestID int64) ([]*models.RequestComment, error) {
	if requestType != "subnet" && requestType != "ip" {
		return nil, fmt.Errorf("invalid request_type: must be 'subnet' or 'ip'")
	}
	if requestID <= 0 {
		return nil, fmt.Errorf("invalid request ID")
	}
	return s.repo.ListRequestComments(ctx, requestType, requestID)
}

// IPAlreadyTakenError is returned when the requested IP is already in use.
type IPAlreadyTakenError struct {
	IP string
}

func (e *IPAlreadyTakenError) Error() string {
	return fmt.Sprintf("IP address %s is already taken", e.IP)
}

// ---- Notification helpers ----

func (s *WorkflowService) notifyAdminsSubnetRequest(ctx context.Context, sr *models.SubnetRequest, requesterID int64) {
	admins, err := s.repo.GetUsersByRole(ctx, "admin")
	if err != nil {
		slog.Error("requests: failed to get admins for notification", "error", err)
		return
	}
	for _, admin := range admins {
		if err := s.notification.Queue(ctx, admin.ID, NotifRequestSubmitted, map[string]interface{}{
			"RequestType": "Subnet",
			"RequestID":   sr.ID,
			"Purpose":     sr.Purpose,
			"PrefixLen":   sr.RequestedPrefixLen,
		}); err != nil {
			slog.Error("requests: failed to notify admin", "admin_id", admin.ID, "error", err)
		}
	}
}

func (s *WorkflowService) notifyAdminsIPRequest(ctx context.Context, ir *models.IPRequest, requesterID int64) {
	admins, err := s.repo.GetUsersByRole(ctx, "admin")
	if err != nil {
		slog.Error("requests: failed to get admins for notification", "error", err)
		return
	}
	for _, admin := range admins {
		if err := s.notification.Queue(ctx, admin.ID, NotifRequestSubmitted, map[string]interface{}{
			"RequestType": "IP",
			"RequestID":   ir.ID,
			"Purpose":     ir.Purpose,
			"SubnetID":    ir.SubnetID,
		}); err != nil {
			slog.Error("requests: failed to notify admin", "admin_id", admin.ID, "error", err)
		}
	}
}

func (s *WorkflowService) notifyRequesterSubnetApproved(ctx context.Context, sr *models.SubnetRequest, subnet *models.Subnet) {
	cidr := fmt.Sprintf("%s/%d", subnet.NetworkAddress, subnet.PrefixLength)
	if err := s.notification.Queue(ctx, sr.RequesterID, NotifRequestApprovedSubnet, map[string]interface{}{
		"RequestID":    sr.ID,
		"SubnetCIDR":   cidr,
		"ReviewerNote": sr.ReviewerNote,
	}); err != nil {
		slog.Error("requests: failed to notify requester (subnet approved)", "requester_id", sr.RequesterID, "error", err)
	}
}

func (s *WorkflowService) notifyRequesterSubnetRejected(ctx context.Context, sr *models.SubnetRequest) {
	if err := s.notification.Queue(ctx, sr.RequesterID, NotifRequestRejected, map[string]interface{}{
		"RequestType":  "Subnet",
		"RequestID":    sr.ID,
		"ReviewerNote": sr.ReviewerNote,
	}); err != nil {
		slog.Error("requests: failed to notify requester (subnet rejected)", "requester_id", sr.RequesterID, "error", err)
	}
}

func (s *WorkflowService) notifyRequesterIPApproved(ctx context.Context, ir *models.IPRequest, ipAddr *models.IPAddress) {
	if err := s.notification.Queue(ctx, ir.RequesterID, NotifRequestApprovedIP, map[string]interface{}{
		"RequestID":    ir.ID,
		"AssignedIP":   ipAddr.Address,
		"ReviewerNote": ir.ReviewerNote,
	}); err != nil {
		slog.Error("requests: failed to notify requester (ip approved)", "requester_id", ir.RequesterID, "error", err)
	}
}

func (s *WorkflowService) notifyRequesterIPRejected(ctx context.Context, ir *models.IPRequest) {
	if err := s.notification.Queue(ctx, ir.RequesterID, NotifRequestRejected, map[string]interface{}{
		"RequestType":  "IP",
		"RequestID":    ir.ID,
		"ReviewerNote": ir.ReviewerNote,
	}); err != nil {
		slog.Error("requests: failed to notify requester (ip rejected)", "requester_id", ir.RequesterID, "error", err)
	}
}

func (s *WorkflowService) notifyRequestComment(ctx context.Context, requestType string, requestID, authorID int64, comment *models.RequestComment) {
	var otherUserID int64
	if requestType == "subnet" {
		sr, err := s.repo.GetSubnetRequestByID(ctx, requestID)
		if err != nil {
			return
		}
		if sr.RequesterID == authorID {
			if sr.ReviewerID != nil {
				otherUserID = *sr.ReviewerID
			} else {
				s.notifyAdminsComment(ctx, comment)
				return
			}
		} else {
			otherUserID = sr.RequesterID
		}
	} else {
		ir, err := s.repo.GetIPRequestByID(ctx, requestID)
		if err != nil {
			return
		}
		if ir.RequesterID == authorID {
			if ir.ReviewerID != nil {
				otherUserID = *ir.ReviewerID
			} else {
				s.notifyAdminsComment(ctx, comment)
				return
			}
		} else {
			otherUserID = ir.RequesterID
		}
	}

	if otherUserID <= 0 {
		return
	}
	if err := s.notification.Queue(ctx, otherUserID, NotifRequestComment, map[string]interface{}{
		"RequestType": requestType,
		"RequestID":   requestID,
		"AuthorName":  comment.AuthorUsername,
		"Body":        comment.Body,
	}); err != nil {
		slog.Error("requests: failed to notify user (comment)", "user_id", otherUserID, "error", err)
	}
}

func (s *WorkflowService) notifyAdminsComment(ctx context.Context, comment *models.RequestComment) {
	admins, err := s.repo.GetUsersByRole(ctx, "admin")
	if err != nil {
		return
	}
	for _, admin := range admins {
		_ = s.notification.Queue(ctx, admin.ID, NotifRequestComment, map[string]interface{}{
			"RequestType": comment.RequestType,
			"RequestID":   comment.RequestID,
			"AuthorName":  comment.AuthorUsername,
			"Body":        comment.Body,
		})
	}
}

// ---- Custom Field Definitions ----

func (s *WorkflowService) ListCustomFieldDefinitions(ctx context.Context, entityType string) ([]*models.CustomFieldDefinition, error) {
	if entityType != "" && !validEntityTypes[entityType] {
		return nil, fmt.Errorf("invalid entity_type: must be subnet, ip_address, or device")
	}
	return s.repo.ListCustomFieldDefinitions(ctx, entityType)
}

func (s *WorkflowService) CreateCustomFieldDefinition(ctx context.Context, p *repository.CustomFieldDefinitionParams) (*models.CustomFieldDefinition, error) {
	if p.EntityType == "" {
		return nil, fmt.Errorf("entity_type is required")
	}
	if !validEntityTypes[p.EntityType] {
		return nil, fmt.Errorf("invalid entity_type: must be subnet, ip_address, or device")
	}
	if p.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if p.Label == "" {
		return nil, fmt.Errorf("label is required")
	}
	if !validFieldTypes[p.FieldType] {
		return nil, fmt.Errorf("invalid field_type: must be text, number, textarea, dropdown, checkbox, date, url, or email")
	}
	return s.repo.CreateCustomFieldDefinition(ctx, p)
}

func (s *WorkflowService) GetCustomFieldDefinition(ctx context.Context, id int64) (*models.CustomFieldDefinition, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid id")
	}
	def, err := s.repo.GetCustomFieldDefinition(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("custom field definition %d: %w", id, ErrNotFound)
		}
		return nil, err
	}
	return def, nil
}

func (s *WorkflowService) UpdateCustomFieldDefinition(ctx context.Context, id int64, p *repository.CustomFieldDefinitionParams) (*models.CustomFieldDefinition, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid id")
	}
	if p.EntityType == "" {
		return nil, fmt.Errorf("entity_type is required")
	}
	if !validEntityTypes[p.EntityType] {
		return nil, fmt.Errorf("invalid entity_type: must be subnet, ip_address, or device")
	}
	if p.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if p.Label == "" {
		return nil, fmt.Errorf("label is required")
	}
	if !validFieldTypes[p.FieldType] {
		return nil, fmt.Errorf("invalid field_type: must be text, number, textarea, dropdown, checkbox, date, url, or email")
	}
	def, err := s.repo.UpdateCustomFieldDefinition(ctx, id, p)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("custom field definition %d: %w", id, ErrNotFound)
		}
		return nil, err
	}
	return def, nil
}

func (s *WorkflowService) DeleteCustomFieldDefinition(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid id")
	}
	if err := s.repo.DeleteCustomFieldDefinition(ctx, id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("custom field definition %d: %w", id, ErrNotFound)
		}
		return err
	}
	return nil
}

func (s *WorkflowService) ReorderCustomFieldDefinitions(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return fmt.Errorf("ids is required")
	}
	return s.repo.ReorderCustomFieldDefinitions(ctx, ids)
}

func (s *WorkflowService) GetCustomFieldValues(ctx context.Context, entityType string, entityID int64) (map[string]*string, error) {
	if entityID <= 0 {
		return nil, fmt.Errorf("invalid entity ID")
	}
	return s.repo.GetCustomFieldValues(ctx, entityType, entityID)
}

func (s *WorkflowService) SetCustomFieldValues(ctx context.Context, entityType string, entityID int64, values map[string]*string) error {
	if entityID <= 0 {
		return fmt.Errorf("invalid entity ID")
	}
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
