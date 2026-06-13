package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"padduck/models"
)

var (
	ErrNotFound       = errors.New("not found")
	ErrNotPending     = errors.New("request is not in pending state")
	ErrNotCancellable = errors.New("request cannot be cancelled")
)

// ---- Subnet Requests ----

// SubmitSubnetRequest creates a new pending subnet request.
func (s *Service) SubmitSubnetRequest(ctx context.Context, requesterID, networkID int64, parentSubnetID *int64, prefixLen int, purpose string) (*models.SubnetRequest, error) {
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

	sr, err := s.repository.CreateSubnetRequest(ctx, requesterID, networkID, parentSubnetID, prefixLen, purpose)
	if err != nil {
		return nil, err
	}

	// Notify admins about new request
	s.notifyAdminsSubnetRequest(ctx, sr, requesterID)

	return sr, nil
}

// ListMySubnetRequests returns subnet requests submitted by the given user.
func (s *Service) ListMySubnetRequests(ctx context.Context, requesterID int64) ([]*models.SubnetRequest, error) {
	if requesterID <= 0 {
		return nil, fmt.Errorf("invalid requester ID")
	}
	return s.repository.ListSubnetRequestsByRequester(ctx, requesterID)
}

// ListAllSubnetRequests returns all subnet requests (admin view).
func (s *Service) ListAllSubnetRequests(ctx context.Context) ([]*models.SubnetRequest, error) {
	return s.repository.ListAllSubnetRequests(ctx)
}

// ApproveSubnetRequest approves a subnet request and creates the subnet.
func (s *Service) ApproveSubnetRequest(ctx context.Context, requestID, reviewerID int64, reviewerNote string) (*models.SubnetRequest, error) {
	if requestID <= 0 {
		return nil, fmt.Errorf("invalid request ID")
	}
	if reviewerID <= 0 {
		return nil, fmt.Errorf("invalid reviewer ID")
	}

	// Fetch the request
	sr, err := s.repository.GetSubnetRequestByID(ctx, requestID)
	if err != nil {
		return nil, fmt.Errorf("subnet request %d: %w", requestID, ErrNotFound)
	}
	if sr.Status != "pending" {
		return nil, fmt.Errorf("subnet request %d: %w", requestID, ErrNotPending)
	}

	// Create the subnet — we need a network address; auto-allocate the first free block in the section
	networkAddr, err := s.findFreeSubnetBlock(ctx, sr.NetworkID, sr.ParentSubnetID, sr.RequestedPrefixLen)
	if err != nil {
		return nil, fmt.Errorf("could not allocate subnet block: %w", err)
	}

	subnet, err := s.CreateSubnet(ctx, sr.NetworkID, networkAddr, sr.RequestedPrefixLen, sr.Purpose, nil, false, false, nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create subnet: %w", err)
	}

	// Approve the request
	approved, err := s.repository.ApproveSubnetRequest(ctx, requestID, reviewerID, subnet.ID, reviewerNote)
	if err != nil {
		return nil, err
	}

	// Notify requester
	s.notifyRequesterSubnetApproved(ctx, approved, subnet)

	return approved, nil
}

// RejectSubnetRequest rejects a subnet request.
func (s *Service) RejectSubnetRequest(ctx context.Context, requestID, reviewerID int64, reviewerNote string) (*models.SubnetRequest, error) {
	if requestID <= 0 {
		return nil, fmt.Errorf("invalid request ID")
	}
	if reviewerID <= 0 {
		return nil, fmt.Errorf("invalid reviewer ID")
	}
	if strings.TrimSpace(reviewerNote) == "" {
		return nil, fmt.Errorf("reviewer_note is required when rejecting")
	}

	rejected, err := s.repository.RejectSubnetRequest(ctx, requestID, reviewerID, reviewerNote)
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

	// Notify requester
	s.notifyRequesterSubnetRejected(ctx, rejected)

	return rejected, nil
}

// CancelSubnetRequest cancels a pending request by the requester.
func (s *Service) CancelSubnetRequest(ctx context.Context, requestID, requesterID int64) error {
	if requestID <= 0 {
		return fmt.Errorf("invalid request ID")
	}
	if requesterID <= 0 {
		return fmt.Errorf("invalid requester ID")
	}
	if err := s.repository.CancelSubnetRequest(ctx, requestID, requesterID); err != nil {
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

// findFreeSubnetBlock finds the first network address that fits a /<prefixLen> block
// within the given section (optionally within a parent subnet).
func (s *Service) findFreeSubnetBlock(ctx context.Context, networkID int64, parentSubnetID *int64, prefixLen int) (string, error) {
	// Get all existing subnets in the section
	existing, err := s.repository.ListSubnetsBySection(ctx, networkID)
	if err != nil {
		return "", err
	}

	// If parent subnet provided, restrict to that parent's address space
	if parentSubnetID != nil {
		parent, err := s.repository.GetSubnetByID(ctx, *parentSubnetID)
		if err != nil {
			return "", fmt.Errorf("parent subnet %d: %w", *parentSubnetID, ErrNotFound)
		}
		return findFirstFreeBlock(parent.NetworkAddress, parent.PrefixLength, prefixLen, existing)
	}

	// No parent — use RFC1918 private space as fallback (10.0.0.0/8)
	return findFirstFreeBlock("10.0.0.0", 8, prefixLen, existing)
}

// ---- IP Requests ----

// SubmitIPRequest creates a new pending IP request.
func (s *Service) SubmitIPRequest(ctx context.Context, requesterID, subnetID int64, requestedIP *string, dnsName, purpose string) (*models.IPRequest, error) {
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

	ir, err := s.repository.CreateIPRequest(ctx, requesterID, subnetID, requestedIP, strings.TrimSpace(dnsName), purpose)
	if err != nil {
		return nil, err
	}

	// Notify admins
	s.notifyAdminsIPRequest(ctx, ir, requesterID)

	return ir, nil
}

// ListMyIPRequests returns IP requests submitted by the given user.
func (s *Service) ListMyIPRequests(ctx context.Context, requesterID int64) ([]*models.IPRequest, error) {
	if requesterID <= 0 {
		return nil, fmt.Errorf("invalid requester ID")
	}
	return s.repository.ListIPRequestsByRequester(ctx, requesterID)
}

// ListAllIPRequests returns all IP requests (admin view).
func (s *Service) ListAllIPRequests(ctx context.Context) ([]*models.IPRequest, error) {
	return s.repository.ListAllIPRequests(ctx)
}

// ApproveIPRequest approves an IP request and assigns the IP.
func (s *Service) ApproveIPRequest(ctx context.Context, requestID, reviewerID int64, reviewerNote string) (*models.IPRequest, error) {
	if requestID <= 0 {
		return nil, fmt.Errorf("invalid request ID")
	}
	if reviewerID <= 0 {
		return nil, fmt.Errorf("invalid reviewer ID")
	}

	// Fetch the request
	ir, err := s.repository.GetIPRequestByID(ctx, requestID)
	if err != nil {
		return nil, fmt.Errorf("ip request %d: %w", requestID, ErrNotFound)
	}
	if ir.Status != "pending" {
		return nil, fmt.Errorf("ip request %d: %w", requestID, ErrNotPending)
	}

	var ipAddr *models.IPAddress

	if ir.RequestedIP != nil && *ir.RequestedIP != "" {
		// Specific IP requested — check if taken
		existing, lookupErr := s.repository.GetIPAddressBySubnetAndAddress(ctx, ir.SubnetID, *ir.RequestedIP)
		if lookupErr == nil && existing != nil && existing.Status != "available" {
			return nil, &IPAlreadyTakenError{IP: *ir.RequestedIP}
		}
		if lookupErr == nil && existing != nil {
			// Mark existing available IP as assigned
			ipAddr, err = s.repository.UpdateIPAddressStatus(ctx, existing.ID, "assigned", nil)
			if err != nil {
				return nil, fmt.Errorf("failed to assign IP: %w", err)
			}
		} else {
			// Create new IP address record
			dnsName := ir.DNSName
			var dnsNamePtr *string
			if dnsName != "" {
				dnsNamePtr = &dnsName
			}
			ipAddr, err = s.repository.CreateIPAddress(ctx, ir.SubnetID, *ir.RequestedIP, dnsName, "assigned", nil, nil, nil, dnsNamePtr)
			if err != nil {
				return nil, fmt.Errorf("failed to create IP: %w", err)
			}
		}
	} else {
		// Auto-assign next free IP
		ipAddr, err = s.repository.AllocateIPAddress(ctx, ir.SubnetID, nil)
		if err != nil {
			return nil, fmt.Errorf("no available IPs in subnet: %w", err)
		}
	}

	// Update DNS name on the IP if provided, then sync to DNS provider
	if ir.DNSName != "" {
		_ = s.repository.UpdateIPDNSName(ctx, ipAddr.ID, ir.DNSName)
		if updated, err := s.repository.GetIPAddressByID(ctx, ipAddr.ID); err == nil {
			go s.DNS.SyncIPToDNS(ctx, updated)
		}
	}

	// Approve the request
	approved, err := s.repository.ApproveIPRequest(ctx, requestID, reviewerID, ipAddr.ID, reviewerNote)
	if err != nil {
		return nil, err
	}

	// Notify requester
	s.notifyRequesterIPApproved(ctx, approved, ipAddr)

	return approved, nil
}

// RejectIPRequest rejects an IP request.
func (s *Service) RejectIPRequest(ctx context.Context, requestID, reviewerID int64, reviewerNote string) (*models.IPRequest, error) {
	if requestID <= 0 {
		return nil, fmt.Errorf("invalid request ID")
	}
	if reviewerID <= 0 {
		return nil, fmt.Errorf("invalid reviewer ID")
	}
	if strings.TrimSpace(reviewerNote) == "" {
		return nil, fmt.Errorf("reviewer_note is required when rejecting")
	}

	rejected, err := s.repository.RejectIPRequest(ctx, requestID, reviewerID, reviewerNote)
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

	// Notify requester
	s.notifyRequesterIPRejected(ctx, rejected)

	return rejected, nil
}

// CancelIPRequest cancels a pending IP request by the requester.
func (s *Service) CancelIPRequest(ctx context.Context, requestID, requesterID int64) error {
	if requestID <= 0 {
		return fmt.Errorf("invalid request ID")
	}
	if requesterID <= 0 {
		return fmt.Errorf("invalid requester ID")
	}
	if err := s.repository.CancelIPRequest(ctx, requestID, requesterID); err != nil {
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

// GetRequestOwner returns the user ID of the requester who owns the given request.
// requestType must be "subnet" or "ip".
func (s *Service) GetRequestOwner(ctx context.Context, requestType string, requestID int64) (int64, error) {
	switch requestType {
	case "subnet":
		sr, err := s.repository.GetSubnetRequestByID(ctx, requestID)
		if err != nil {
			return 0, fmt.Errorf("subnet request not found")
		}
		return sr.RequesterID, nil
	case "ip":
		ir, err := s.repository.GetIPRequestByID(ctx, requestID)
		if err != nil {
			return 0, fmt.Errorf("ip request not found")
		}
		return ir.RequesterID, nil
	default:
		return 0, fmt.Errorf("invalid request_type: must be 'subnet' or 'ip'")
	}
}

// GetPendingRequestCounts returns the total pending request count.
func (s *Service) GetPendingRequestCounts(ctx context.Context) (subnetCount, ipCount int64, err error) {
	subnetCount, err = s.repository.CountPendingSubnetRequests(ctx)
	if err != nil {
		return
	}
	ipCount, err = s.repository.CountPendingIPRequests(ctx)
	return
}

// ---- Request Comments ----

// AddRequestComment adds a comment to a subnet or IP request.
// Only the requester or an admin (reviewer) may comment.
func (s *Service) AddRequestComment(ctx context.Context, requestType string, requestID, authorID int64, body string) (*models.RequestComment, error) {
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

	comment, err := s.repository.CreateRequestComment(ctx, requestType, requestID, authorID, body)
	if err != nil {
		return nil, err
	}

	// Audit the comment
	s.Audit.Log(ctx, AuditEntry{
		UserID: &authorID, Action: "request_comment_added",
		ResourceType: requestType + "_request", ResourceID: &requestID,
		NewValues: map[string]interface{}{"body": body},
	})

	// Notify the other party
	s.notifyRequestComment(ctx, requestType, requestID, authorID, comment)

	return comment, nil
}

// ListRequestComments returns comments for a given request.
func (s *Service) ListRequestComments(ctx context.Context, requestType string, requestID int64) ([]*models.RequestComment, error) {
	if requestType != "subnet" && requestType != "ip" {
		return nil, fmt.Errorf("invalid request_type: must be 'subnet' or 'ip'")
	}
	if requestID <= 0 {
		return nil, fmt.Errorf("invalid request ID")
	}
	return s.repository.ListRequestComments(ctx, requestType, requestID)
}

// IPAlreadyTakenError is returned when the requested IP is already in use.
type IPAlreadyTakenError struct {
	IP string
}

func (e *IPAlreadyTakenError) Error() string {
	return fmt.Sprintf("IP address %s is already taken", e.IP)
}

// ---- Notification helpers ----

func (s *Service) notifyAdminsSubnetRequest(ctx context.Context, sr *models.SubnetRequest, requesterID int64) {
	admins, err := s.repository.GetUsersByRole(ctx, "admin")
	if err != nil {
		log.Printf("[requests] failed to get admins for notification: %v", err)
		return
	}
	for _, admin := range admins {
		if err := s.Notification.Queue(ctx, admin.ID, NotifRequestSubmitted, map[string]interface{}{
			"RequestType": "Subnet",
			"RequestID":   sr.ID,
			"Purpose":     sr.Purpose,
			"PrefixLen":   sr.RequestedPrefixLen,
		}); err != nil {
			log.Printf("[requests] failed to notify admin %d: %v", admin.ID, err)
		}
	}
}

func (s *Service) notifyAdminsIPRequest(ctx context.Context, ir *models.IPRequest, requesterID int64) {
	admins, err := s.repository.GetUsersByRole(ctx, "admin")
	if err != nil {
		log.Printf("[requests] failed to get admins for notification: %v", err)
		return
	}
	for _, admin := range admins {
		if err := s.Notification.Queue(ctx, admin.ID, NotifRequestSubmitted, map[string]interface{}{
			"RequestType": "IP",
			"RequestID":   ir.ID,
			"Purpose":     ir.Purpose,
			"SubnetID":    ir.SubnetID,
		}); err != nil {
			log.Printf("[requests] failed to notify admin %d: %v", admin.ID, err)
		}
	}
}

func (s *Service) notifyRequesterSubnetApproved(ctx context.Context, sr *models.SubnetRequest, subnet *models.Subnet) {
	cidr := fmt.Sprintf("%s/%d", subnet.NetworkAddress, subnet.PrefixLength)
	if err := s.Notification.Queue(ctx, sr.RequesterID, NotifRequestApprovedSubnet, map[string]interface{}{
		"RequestID":    sr.ID,
		"SubnetCIDR":   cidr,
		"ReviewerNote": sr.ReviewerNote,
	}); err != nil {
		log.Printf("[requests] failed to notify requester %d (subnet approved): %v", sr.RequesterID, err)
	}
}

func (s *Service) notifyRequesterSubnetRejected(ctx context.Context, sr *models.SubnetRequest) {
	if err := s.Notification.Queue(ctx, sr.RequesterID, NotifRequestRejected, map[string]interface{}{
		"RequestType":  "Subnet",
		"RequestID":    sr.ID,
		"ReviewerNote": sr.ReviewerNote,
	}); err != nil {
		log.Printf("[requests] failed to notify requester %d (subnet rejected): %v", sr.RequesterID, err)
	}
}

func (s *Service) notifyRequesterIPApproved(ctx context.Context, ir *models.IPRequest, ipAddr *models.IPAddress) {
	if err := s.Notification.Queue(ctx, ir.RequesterID, NotifRequestApprovedIP, map[string]interface{}{
		"RequestID":    ir.ID,
		"AssignedIP":   ipAddr.Address,
		"ReviewerNote": ir.ReviewerNote,
	}); err != nil {
		log.Printf("[requests] failed to notify requester %d (ip approved): %v", ir.RequesterID, err)
	}
}

func (s *Service) notifyRequesterIPRejected(ctx context.Context, ir *models.IPRequest) {
	if err := s.Notification.Queue(ctx, ir.RequesterID, NotifRequestRejected, map[string]interface{}{
		"RequestType":  "IP",
		"RequestID":    ir.ID,
		"ReviewerNote": ir.ReviewerNote,
	}); err != nil {
		log.Printf("[requests] failed to notify requester %d (ip rejected): %v", ir.RequesterID, err)
	}
}

func (s *Service) notifyRequestComment(ctx context.Context, requestType string, requestID, authorID int64, comment *models.RequestComment) {
	// Find the other party to notify
	var otherUserID int64
	if requestType == "subnet" {
		sr, err := s.repository.GetSubnetRequestByID(ctx, requestID)
		if err != nil {
			return
		}
		if sr.RequesterID == authorID {
			// Author is requester — notify the reviewer (or admins)
			if sr.ReviewerID != nil {
				otherUserID = *sr.ReviewerID
			} else {
				// No reviewer yet — notify admins
				s.notifyAdminsComment(ctx, comment)
				return
			}
		} else {
			otherUserID = sr.RequesterID
		}
	} else {
		ir, err := s.repository.GetIPRequestByID(ctx, requestID)
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
	if err := s.Notification.Queue(ctx, otherUserID, NotifRequestComment, map[string]interface{}{
		"RequestType": requestType,
		"RequestID":   requestID,
		"AuthorName":  comment.AuthorUsername,
		"Body":        comment.Body,
	}); err != nil {
		log.Printf("[requests] failed to notify user %d (comment): %v", otherUserID, err)
	}
}

func (s *Service) notifyAdminsComment(ctx context.Context, comment *models.RequestComment) {
	admins, err := s.repository.GetUsersByRole(ctx, "admin")
	if err != nil {
		return
	}
	for _, admin := range admins {
		_ = s.Notification.Queue(ctx, admin.ID, NotifRequestComment, map[string]interface{}{
			"RequestType": comment.RequestType,
			"RequestID":   comment.RequestID,
			"AuthorName":  comment.AuthorUsername,
			"Body":        comment.Body,
		})
	}
}
