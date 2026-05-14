package services

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"ipam-next/models"
)

// ---------------------------------------------------------------------------
// Permission constant tests
// ---------------------------------------------------------------------------

func TestRequestPermissionConstants(t *testing.T) {
	assert.Equal(t, "ipam:subnet_request:submit", PermV2SubnetRequestSubmit)
	assert.Equal(t, "ipam:subnet_request:review", PermV2SubnetRequestReview)
}

func TestRequestPermissionsInAllPermissions(t *testing.T) {
	found := map[string]bool{}
	for _, p := range AllPermissions {
		found[p] = true
	}
	assert.True(t, found[PermV2SubnetRequestSubmit], "PermV2SubnetRequestSubmit must be in AllPermissions")
	assert.True(t, found[PermV2SubnetRequestReview], "PermV2SubnetRequestReview must be in AllPermissions")
}

// ---------------------------------------------------------------------------
// RBAC: legacy role mapping for request permissions
// ---------------------------------------------------------------------------

func TestLegacyRole_Admin_HasRequestPermissions(t *testing.T) {
	assert.True(t, legacyRoleHasPermission("admin", PermV2SubnetRequestSubmit))
	assert.True(t, legacyRoleHasPermission("admin", PermV2SubnetRequestReview))
}

func TestLegacyRole_User_HasSubmitButNotReview(t *testing.T) {
	assert.True(t, legacyRoleHasPermission("user", PermV2SubnetRequestSubmit), "user should be able to submit requests")
	assert.False(t, legacyRoleHasPermission("user", PermV2SubnetRequestReview), "user should not be able to review requests")
}

func TestLegacyRole_Viewer_HasNoRequestPermissions(t *testing.T) {
	assert.False(t, legacyRoleHasPermission("viewer", PermV2SubnetRequestSubmit), "viewer cannot submit requests")
	assert.False(t, legacyRoleHasPermission("viewer", PermV2SubnetRequestReview), "viewer cannot review requests")
}

// ---------------------------------------------------------------------------
// SubmitSubnetRequest — validation
// ---------------------------------------------------------------------------

func TestSubmitSubnetRequest_InvalidRequesterID(t *testing.T) {
	s := &Service{}
	_, err := s.SubmitSubnetRequest(nil, 0, 1, nil, 24, "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid requester ID")
}

func TestSubmitSubnetRequest_InvalidSectionID(t *testing.T) {
	s := &Service{}
	_, err := s.SubmitSubnetRequest(nil, 1, 0, nil, 24, "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid section ID")
}

func TestSubmitSubnetRequest_InvalidPrefixLen(t *testing.T) {
	s := &Service{}
	_, err := s.SubmitSubnetRequest(nil, 1, 1, nil, 33, "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "requested_prefix_len")
}

func TestSubmitSubnetRequest_EmptyPurpose(t *testing.T) {
	s := &Service{}
	_, err := s.SubmitSubnetRequest(nil, 1, 1, nil, 24, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "purpose")
}

func TestSubmitSubnetRequest_WhitespacePurpose(t *testing.T) {
	s := &Service{}
	_, err := s.SubmitSubnetRequest(nil, 1, 1, nil, 24, "   ")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "purpose")
}

// ---------------------------------------------------------------------------
// SubmitIPRequest — validation
// ---------------------------------------------------------------------------

func TestSubmitIPRequest_InvalidRequesterID(t *testing.T) {
	s := &Service{}
	_, err := s.SubmitIPRequest(nil, 0, 1, nil, "", "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid requester ID")
}

func TestSubmitIPRequest_InvalidSubnetID(t *testing.T) {
	s := &Service{}
	_, err := s.SubmitIPRequest(nil, 1, 0, nil, "", "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid subnet ID")
}

func TestSubmitIPRequest_EmptyPurpose(t *testing.T) {
	s := &Service{}
	_, err := s.SubmitIPRequest(nil, 1, 1, nil, "", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "purpose")
}

// ---------------------------------------------------------------------------
// ApproveSubnetRequest — validation
// ---------------------------------------------------------------------------

func TestApproveSubnetRequest_InvalidRequestID(t *testing.T) {
	s := &Service{}
	_, err := s.ApproveSubnetRequest(nil, 0, 1, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid request ID")
}

func TestApproveSubnetRequest_InvalidReviewerID(t *testing.T) {
	s := &Service{}
	_, err := s.ApproveSubnetRequest(nil, 1, 0, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid reviewer ID")
}

// ---------------------------------------------------------------------------
// RejectSubnetRequest — validation
// ---------------------------------------------------------------------------

func TestRejectSubnetRequest_InvalidRequestID(t *testing.T) {
	s := &Service{}
	_, err := s.RejectSubnetRequest(nil, 0, 1, "reason")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid request ID")
}

func TestRejectSubnetRequest_MissingNote(t *testing.T) {
	s := &Service{}
	_, err := s.RejectSubnetRequest(nil, 1, 1, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reviewer_note")
}

// ---------------------------------------------------------------------------
// CancelSubnetRequest — validation
// ---------------------------------------------------------------------------

func TestCancelSubnetRequest_InvalidRequestID(t *testing.T) {
	s := &Service{}
	err := s.CancelSubnetRequest(nil, 0, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid request ID")
}

func TestCancelSubnetRequest_InvalidRequesterID(t *testing.T) {
	s := &Service{}
	err := s.CancelSubnetRequest(nil, 1, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid requester ID")
}

// ---------------------------------------------------------------------------
// ApproveIPRequest — validation
// ---------------------------------------------------------------------------

func TestApproveIPRequest_InvalidRequestID(t *testing.T) {
	s := &Service{}
	_, err := s.ApproveIPRequest(nil, 0, 1, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid request ID")
}

func TestApproveIPRequest_InvalidReviewerID(t *testing.T) {
	s := &Service{}
	_, err := s.ApproveIPRequest(nil, 1, 0, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid reviewer ID")
}

// ---------------------------------------------------------------------------
// RejectIPRequest — validation
// ---------------------------------------------------------------------------

func TestRejectIPRequest_InvalidRequestID(t *testing.T) {
	s := &Service{}
	_, err := s.RejectIPRequest(nil, 0, 1, "reason")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid request ID")
}

func TestRejectIPRequest_MissingNote(t *testing.T) {
	s := &Service{}
	_, err := s.RejectIPRequest(nil, 1, 1, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reviewer_note")
}

// ---------------------------------------------------------------------------
// CancelIPRequest — validation
// ---------------------------------------------------------------------------

func TestCancelIPRequest_InvalidRequestID(t *testing.T) {
	s := &Service{}
	err := s.CancelIPRequest(nil, 0, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid request ID")
}

func TestCancelIPRequest_InvalidRequesterID(t *testing.T) {
	s := &Service{}
	err := s.CancelIPRequest(nil, 1, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid requester ID")
}

// ---------------------------------------------------------------------------
// AddRequestComment — validation
// ---------------------------------------------------------------------------

func TestAddRequestComment_InvalidType(t *testing.T) {
	s := &Service{}
	_, err := s.AddRequestComment(nil, "invalid", 1, 1, "body")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid request_type")
}

func TestAddRequestComment_InvalidRequestID(t *testing.T) {
	s := &Service{}
	_, err := s.AddRequestComment(nil, "subnet", 0, 1, "body")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid request ID")
}

func TestAddRequestComment_InvalidAuthorID(t *testing.T) {
	s := &Service{}
	_, err := s.AddRequestComment(nil, "subnet", 1, 0, "body")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid author ID")
}

func TestAddRequestComment_EmptyBody(t *testing.T) {
	s := &Service{}
	_, err := s.AddRequestComment(nil, "subnet", 1, 1, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "body is required")
}

func TestAddRequestComment_WhitespaceBody(t *testing.T) {
	s := &Service{}
	_, err := s.AddRequestComment(nil, "ip", 1, 1, "   ")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "body is required")
}

// ---------------------------------------------------------------------------
// ListRequestComments — validation
// ---------------------------------------------------------------------------

func TestListRequestComments_InvalidType(t *testing.T) {
	s := &Service{}
	_, err := s.ListRequestComments(nil, "bad", 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid request_type")
}

func TestListRequestComments_InvalidID(t *testing.T) {
	s := &Service{}
	_, err := s.ListRequestComments(nil, "subnet", 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid request ID")
}

// ---------------------------------------------------------------------------
// IPAlreadyTakenError
// ---------------------------------------------------------------------------

func TestIPAlreadyTakenError_Message(t *testing.T) {
	err := &IPAlreadyTakenError{IP: "10.0.0.1"}
	assert.Contains(t, err.Error(), "10.0.0.1")
	assert.Contains(t, err.Error(), "already taken")
}

// ---------------------------------------------------------------------------
// findFirstFreeBlock — unit tests
// ---------------------------------------------------------------------------

func TestFindFirstFreeBlock_NoExisting(t *testing.T) {
	addr, err := findFirstFreeBlock("10.0.0.0", 8, 24, nil)
	assert.NoError(t, err)
	assert.Equal(t, "10.0.0.0", addr)
}

func TestFindFirstFreeBlock_SkipsExisting(t *testing.T) {
	existing := makeTestSubnets([]string{"10.0.0.0/24", "10.0.1.0/24"})
	addr, err := findFirstFreeBlock("10.0.0.0", 8, 24, existing)
	assert.NoError(t, err)
	assert.Equal(t, "10.0.2.0", addr)
}

func TestFindFirstFreeBlock_PrefixTooSmall(t *testing.T) {
	_, err := findFirstFreeBlock("10.0.0.0", 24, 8, nil)
	assert.Error(t, err)
}

func TestFindFirstFreeBlock_NoSpace(t *testing.T) {
	// Fill the entire /30 with /32s (4 addresses)
	existing := makeTestSubnets([]string{"10.0.0.0/32", "10.0.0.1/32", "10.0.0.2/32", "10.0.0.3/32"})
	_, err := findFirstFreeBlock("10.0.0.0", 30, 32, existing)
	assert.Error(t, err)
}

// makeTestSubnets creates subnet model stubs from CIDR strings.
func makeTestSubnets(cidrs []string) []*models.Subnet {
	var result []*models.Subnet
	for _, cidr := range cidrs {
		parts := splitCIDR(cidr)
		if len(parts) != 2 {
			continue
		}
		pl := 0
		_, err := fmt.Sscanf(parts[1], "%d", &pl)
		if err != nil {
			continue
		}
		result = append(result, &models.Subnet{
			NetworkAddress: parts[0],
			PrefixLength:   pl,
		})
	}
	return result
}

func splitCIDR(cidr string) []string {
	for i, c := range cidr {
		if c == '/' {
			return []string{cidr[:i], cidr[i+1:]}
		}
	}
	return []string{cidr}
}
