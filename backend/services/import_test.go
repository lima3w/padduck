package services

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"padduck/models"
)

// ─────────────────────────────────────────────────────────────────────────────
// Stub repo for ImportService
// ─────────────────────────────────────────────────────────────────────────────

type stubImportRepo struct {
	sections []*models.Network
	subnets  []*models.Subnet
	ips      map[int64][]*models.IPAddress
	vlans    []*models.VLAN
	vrfs     []*models.VRF

	createdSubnets []*models.Subnet
	createdIPs     []*models.IPAddress
	nextSubnetID   int64
	nextIPID       int64
}

func newStubImportRepo() *stubImportRepo {
	return &stubImportRepo{
		ips:          make(map[int64][]*models.IPAddress),
		nextSubnetID: 1,
		nextIPID:     1,
	}
}

func (r *stubImportRepo) ListAllNetworks(_ context.Context) ([]*models.Network, error) {
	return r.sections, nil
}
func (r *stubImportRepo) ListSubnetsBySection(_ context.Context, networkID int64) ([]*models.Subnet, error) {
	var out []*models.Subnet
	for _, s := range r.subnets {
		if s.NetworkID == networkID {
			out = append(out, s)
		}
	}
	return out, nil
}
func (r *stubImportRepo) ListAllSubnets(_ context.Context) ([]*models.Subnet, error) {
	return r.subnets, nil
}
func (r *stubImportRepo) CreateSubnetWithVLAN(_ context.Context, networkID int64, networkAddr string, prefixLen int, description string, gateway *string, autoFirst, autoLast bool, locationID *int64, nameserverID *int64, vlanID *int64) (*models.Subnet, error) {
	sub := &models.Subnet{
		ID:             r.nextSubnetID,
		NetworkID:      networkID,
		NetworkAddress: networkAddr,
		PrefixLength:   prefixLen,
		Description:    description,
		Gateway:        gateway,
		VLANID:         vlanID,
	}
	r.nextSubnetID++
	r.subnets = append(r.subnets, sub)
	r.createdSubnets = append(r.createdSubnets, sub)
	return sub, nil
}
func (r *stubImportRepo) ListIPAddressesBySubnet(_ context.Context, subnetID int64) ([]*models.IPAddress, error) {
	return r.ips[subnetID], nil
}
func (r *stubImportRepo) CreateIPAddress(_ context.Context, subnetID int64, address, hostname, status string, tagID *int64, macAddress, ptrRecord, dnsName *string) (*models.IPAddress, error) {
	ip := &models.IPAddress{
		ID:         r.nextIPID,
		SubnetID:   subnetID,
		Address:    address,
		Hostname:   hostname,
		Status:     status,
		MACAddress: macAddress,
	}
	r.nextIPID++
	r.ips[subnetID] = append(r.ips[subnetID], ip)
	r.createdIPs = append(r.createdIPs, ip)
	return ip, nil
}
func (r *stubImportRepo) ListAllVLANs(_ context.Context) ([]*models.VLAN, error) {
	return r.vlans, nil
}
func (r *stubImportRepo) ListAllVRFs(_ context.Context) ([]*models.VRF, error) {
	return r.vrfs, nil
}

// helper
func newTestImportService(repo *stubImportRepo) *ImportService {
	return NewImportService(repo)
}

// ─────────────────────────────────────────────────────────────────────────────
// ImportSubnetsCSV (#225)
// ─────────────────────────────────────────────────────────────────────────────

func TestImportSubnetsCSV_HappyPath(t *testing.T) {
	t.Parallel()

	repo := newStubImportRepo()
	repo.sections = []*models.Network{
		{ID: 1, Name: "Default"},
	}

	svc := newTestImportService(repo)

	csv := "cidr,description,section,gateway,vlan,vrf,location\n" +
		"10.0.0.0/24,Test subnet,Default,10.0.0.1,,,"

	result, err := svc.ImportSubnetsCSV(context.Background(), strings.NewReader(csv))
	require.NoError(t, err)
	assert.Equal(t, 1, result.Total)
	assert.Equal(t, 1, result.Imported)
	assert.Equal(t, 0, result.Failed)
	assert.Empty(t, result.Errors)
	require.Len(t, repo.createdSubnets, 1)
	assert.Equal(t, "10.0.0.0", repo.createdSubnets[0].NetworkAddress)
	assert.Equal(t, 24, repo.createdSubnets[0].PrefixLength)
}

func TestImportSubnetsCSV_MultipleRows(t *testing.T) {
	repo := newStubImportRepo()
	repo.sections = []*models.Network{
		{ID: 1, Name: "Primary"},
	}

	svc := newTestImportService(repo)

	csv := "cidr,description,section,gateway,vlan,vrf,location\n" +
		"192.168.0.0/24,LAN,Primary,,,, \n" +
		"172.16.0.0/16,WAN,Primary,,,,"

	result, err := svc.ImportSubnetsCSV(context.Background(), strings.NewReader(csv))
	require.NoError(t, err)
	assert.Equal(t, 2, result.Total)
	assert.Equal(t, 2, result.Imported)
	assert.Equal(t, 0, result.Failed)
}

func TestImportSubnetsCSV_MissingCIDR(t *testing.T) {
	repo := newStubImportRepo()
	repo.sections = []*models.Network{{ID: 1, Name: "Default"}}
	svc := newTestImportService(repo)

	csv := "cidr,description,section,gateway,vlan,vrf,location\n" +
		",Empty CIDR,Default,,,,\n" +
		"10.0.1.0/24,Valid,Default,,,,"

	result, err := svc.ImportSubnetsCSV(context.Background(), strings.NewReader(csv))
	require.NoError(t, err)
	assert.Equal(t, 2, result.Total)
	assert.Equal(t, 1, result.Imported)
	assert.Equal(t, 1, result.Failed)
	assert.Equal(t, 1, len(result.Errors))
	assert.Equal(t, 2, result.Errors[0].Row)
}

func TestImportSubnetsCSV_InvalidCIDR(t *testing.T) {
	repo := newStubImportRepo()
	repo.sections = []*models.Network{{ID: 1, Name: "Default"}}
	svc := newTestImportService(repo)

	csv := "cidr,description,section,gateway,vlan,vrf,location\n" +
		"not-a-cidr,Bad,Default,,,,"

	result, err := svc.ImportSubnetsCSV(context.Background(), strings.NewReader(csv))
	require.NoError(t, err)
	assert.Equal(t, 1, result.Total)
	assert.Equal(t, 0, result.Imported)
	assert.Equal(t, 1, result.Failed)
}

func TestImportSubnetsCSV_SectionNotFound(t *testing.T) {
	repo := newStubImportRepo()
	// No sections at all.
	svc := newTestImportService(repo)

	csv := "cidr,description,section,gateway,vlan,vrf,location\n" +
		"10.0.0.0/24,Test,NonExistentSection,,,,"

	result, err := svc.ImportSubnetsCSV(context.Background(), strings.NewReader(csv))
	require.NoError(t, err)
	assert.Equal(t, 1, result.Total)
	assert.Equal(t, 1, result.Failed)
	assert.Contains(t, result.Errors[0].Message, "not found")
}

func TestImportSubnetsCSV_WithVLAN(t *testing.T) {
	repo := newStubImportRepo()
	repo.sections = []*models.Network{{ID: 1, Name: "Default"}}
	vlanID := int64(10)
	repo.vlans = []*models.VLAN{{ID: vlanID, Name: "VLAN10"}}
	svc := newTestImportService(repo)

	csv := "cidr,description,section,gateway,vlan,vrf,location\n" +
		"10.10.0.0/24,With VLAN,Default,,VLAN10,,"

	result, err := svc.ImportSubnetsCSV(context.Background(), strings.NewReader(csv))
	require.NoError(t, err)
	assert.Equal(t, 1, result.Imported)
	require.Len(t, repo.createdSubnets, 1)
	assert.Equal(t, &vlanID, repo.createdSubnets[0].VLANID)
}

// ─────────────────────────────────────────────────────────────────────────────
// ImportIPsCSV (#226)
// ─────────────────────────────────────────────────────────────────────────────

func TestImportIPsCSV_HappyPath(t *testing.T) {
	t.Parallel()

	repo := newStubImportRepo()
	repo.subnets = []*models.Subnet{
		{ID: 1, NetworkAddress: "10.0.0.0", PrefixLength: 24},
	}
	svc := newTestImportService(repo)

	csv := "address,hostname,status,subnet_cidr,assigned_to,mac_address\n" +
		"10.0.0.10,server1,assigned,10.0.0.0/24,alice,aa:bb:cc:dd:ee:ff"

	result, err := svc.ImportIPsCSV(context.Background(), strings.NewReader(csv))
	require.NoError(t, err)
	assert.Equal(t, 1, result.Total)
	assert.Equal(t, 1, result.Imported)
	assert.Equal(t, 0, result.Failed)
	require.Len(t, repo.createdIPs, 1)
	assert.Equal(t, "10.0.0.10", repo.createdIPs[0].Address)
	assert.Equal(t, "assigned", repo.createdIPs[0].Status)
}

func TestImportIPsCSV_DefaultStatus(t *testing.T) {
	t.Parallel()

	repo := newStubImportRepo()
	repo.subnets = []*models.Subnet{
		{ID: 1, NetworkAddress: "10.0.0.0", PrefixLength: 24},
	}
	svc := newTestImportService(repo)

	// No status column value — should default to "available".
	csv := "address,hostname,status,subnet_cidr,assigned_to,mac_address\n" +
		"10.0.0.11,server2,,10.0.0.0/24,,"

	result, err := svc.ImportIPsCSV(context.Background(), strings.NewReader(csv))
	require.NoError(t, err)
	assert.Equal(t, 1, result.Imported)
	assert.Equal(t, "available", repo.createdIPs[0].Status)
}

func TestImportIPsCSV_InvalidStatus(t *testing.T) {
	t.Parallel()

	repo := newStubImportRepo()
	repo.subnets = []*models.Subnet{
		{ID: 1, NetworkAddress: "10.0.0.0", PrefixLength: 24},
	}
	svc := newTestImportService(repo)

	csv := "address,hostname,status,subnet_cidr,assigned_to,mac_address\n" +
		"10.0.0.12,server3,bogus,10.0.0.0/24,,"

	result, err := svc.ImportIPsCSV(context.Background(), strings.NewReader(csv))
	require.NoError(t, err)
	assert.Equal(t, 1, result.Total)
	assert.Equal(t, 1, result.Failed)
	assert.Contains(t, result.Errors[0].Message, "invalid status")
	assert.Empty(t, repo.createdIPs)
}

func TestImportIPsCSV_RejectsLegacyStatusesBeforeDBInsert(t *testing.T) {
	t.Parallel()

	repo := newStubImportRepo()
	repo.subnets = []*models.Subnet{
		{ID: 1, NetworkAddress: "10.0.0.0", PrefixLength: 24},
	}
	svc := newTestImportService(repo)

	csv := "address,hostname,status,subnet_cidr,assigned_to,mac_address\n" +
		"10.0.0.13,server4,active,10.0.0.0/24,,"

	result, err := svc.ImportIPsCSV(context.Background(), strings.NewReader(csv))
	require.NoError(t, err)
	assert.Equal(t, 1, result.Total)
	assert.Equal(t, 0, result.Imported)
	assert.Equal(t, 1, result.Failed)
	assert.Contains(t, result.Errors[0].Message, `invalid status "active"`)
	assert.Empty(t, repo.createdIPs)
}

func TestImportIPsCSV_MissingAddress(t *testing.T) {
	repo := newStubImportRepo()
	repo.subnets = []*models.Subnet{{ID: 1, NetworkAddress: "10.0.0.0", PrefixLength: 24}}
	svc := newTestImportService(repo)

	csv := "address,hostname,status,subnet_cidr,assigned_to,mac_address\n" +
		",,active,10.0.0.0/24,,"

	result, err := svc.ImportIPsCSV(context.Background(), strings.NewReader(csv))
	require.NoError(t, err)
	assert.Equal(t, 1, result.Failed)
	assert.Equal(t, 0, result.Imported)
}

func TestImportIPsCSV_SubnetNotFound(t *testing.T) {
	repo := newStubImportRepo()
	// No subnets.
	svc := newTestImportService(repo)

	csv := "address,hostname,status,subnet_cidr,assigned_to,mac_address\n" +
		"10.0.0.10,host,active,10.0.0.0/24,,"

	result, err := svc.ImportIPsCSV(context.Background(), strings.NewReader(csv))
	require.NoError(t, err)
	assert.Equal(t, 1, result.Failed)
	assert.Contains(t, result.Errors[0].Message, "not found")
}

func TestImportIPsCSV_AllValidStatuses(t *testing.T) {
	t.Parallel()

	repo := newStubImportRepo()
	repo.subnets = []*models.Subnet{{ID: 1, NetworkAddress: "10.0.0.0", PrefixLength: 24}}
	svc := newTestImportService(repo)

	csv := "address,hostname,status,subnet_cidr,assigned_to,mac_address\n" +
		"10.0.0.1,h1,available,10.0.0.0/24,,\n" +
		"10.0.0.2,h2,reserved,10.0.0.0/24,,\n" +
		"10.0.0.3,h3,assigned,10.0.0.0/24,,"

	result, err := svc.ImportIPsCSV(context.Background(), strings.NewReader(csv))
	require.NoError(t, err)
	assert.Equal(t, 3, result.Imported)
	assert.Equal(t, 0, result.Failed)
}

// ─────────────────────────────────────────────────────────────────────────────
// ImportFromPHPIpam (#227)
// ─────────────────────────────────────────────────────────────────────────────

func TestImportFromPHPIpam_Subnets_HappyPath(t *testing.T) {
	repo := newStubImportRepo()
	repo.sections = []*models.Network{{ID: 1, Name: "Main"}}
	svc := newTestImportService(repo)

	csv := "subnet,mask,description,sectionName\n" +
		"192.168.1.0,255.255.255.0,PHPIpam subnet,Main"

	result, err := svc.ImportFromPHPIpam(context.Background(), strings.NewReader(csv), "subnets")
	require.NoError(t, err)
	assert.Equal(t, 1, result.Total)
	assert.Equal(t, 1, result.Imported)
	assert.Equal(t, 0, result.Failed)
	require.Len(t, repo.createdSubnets, 1)
	assert.Equal(t, "192.168.1.0", repo.createdSubnets[0].NetworkAddress)
	assert.Equal(t, 24, repo.createdSubnets[0].PrefixLength)
}

func TestImportFromPHPIpam_Subnets_MaskAsInt(t *testing.T) {
	repo := newStubImportRepo()
	repo.sections = []*models.Network{{ID: 1, Name: "Main"}}
	svc := newTestImportService(repo)

	csv := "subnet,mask,description,sectionName\n" +
		"10.0.0.0,16,Corp,Main"

	result, err := svc.ImportFromPHPIpam(context.Background(), strings.NewReader(csv), "subnets")
	require.NoError(t, err)
	assert.Equal(t, 1, result.Imported)
	assert.Equal(t, 16, repo.createdSubnets[0].PrefixLength)
}

func TestImportFromPHPIpam_IPs_HappyPath(t *testing.T) {
	repo := newStubImportRepo()
	repo.subnets = []*models.Subnet{{ID: 1, NetworkAddress: "10.0.0.0", PrefixLength: 24}}
	svc := newTestImportService(repo)

	csv := "ip,hostname,description,subnetIp,subnetMask,state\n" +
		"10.0.0.5,myhost,Some IP,10.0.0.0,255.255.255.0,1"

	result, err := svc.ImportFromPHPIpam(context.Background(), strings.NewReader(csv), "ips")
	require.NoError(t, err)
	assert.Equal(t, 1, result.Imported)
	require.Len(t, repo.createdIPs, 1)
	assert.Equal(t, "10.0.0.5", repo.createdIPs[0].Address)
	assert.Equal(t, "active", repo.createdIPs[0].Status)
}

func TestImportFromPHPIpam_IPs_StateMapping(t *testing.T) {
	tests := []struct {
		state  string
		expect string
	}{
		{"1", "active"},
		{"used", "active"},
		{"active", "active"},
		{"2", "reserved"},
		{"reserved", "reserved"},
		{"3", "dhcp"},
		{"dhcp", "dhcp"},
		{"0", "inactive"},
		{"unknown", "inactive"},
	}

	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			repo := newStubImportRepo()
			repo.subnets = []*models.Subnet{{ID: 1, NetworkAddress: "10.0.0.0", PrefixLength: 24}}
			svc := newTestImportService(repo)

			csv := "ip,hostname,description,subnetIp,subnetMask,state\n" +
				"10.0.0.10,h," + tt.state + ",10.0.0.0,24," + tt.state

			result, err := svc.ImportFromPHPIpam(context.Background(), strings.NewReader(csv), "ips")
			require.NoError(t, err)
			require.Equal(t, 1, result.Imported, "state=%s", tt.state)
			assert.Equal(t, tt.expect, repo.createdIPs[0].Status, "state=%s", tt.state)
		})
	}
}

func TestImportFromPHPIpam_InvalidKind(t *testing.T) {
	repo := newStubImportRepo()
	svc := newTestImportService(repo)

	_, err := svc.ImportFromPHPIpam(context.Background(), strings.NewReader(""), "devices")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown kind")
}

func TestImportFromPHPIpam_Subnets_SectionNotFound(t *testing.T) {
	repo := newStubImportRepo()
	svc := newTestImportService(repo)

	csv := "subnet,mask,description,sectionName\n" +
		"10.0.0.0,24,test,NoSection"

	result, err := svc.ImportFromPHPIpam(context.Background(), strings.NewReader(csv), "subnets")
	require.NoError(t, err)
	assert.Equal(t, 1, result.Failed)
}

// ─────────────────────────────────────────────────────────────────────────────
// ExportFullData (#228)
// ─────────────────────────────────────────────────────────────────────────────

func TestExportFullData_CSV(t *testing.T) {
	repo := newStubImportRepo()
	repo.subnets = []*models.Subnet{
		{ID: 1, NetworkID: 1, NetworkAddress: "10.0.0.0", PrefixLength: 24, Description: "Corp"},
	}
	repo.ips[1] = []*models.IPAddress{
		{ID: 1, SubnetID: 1, Address: "10.0.0.10", Hostname: "server1", Status: "active"},
	}
	svc := newTestImportService(repo)

	data, filename, ct, err := svc.ExportFullData(context.Background(), "csv")
	require.NoError(t, err)
	assert.Equal(t, "text/csv", ct)
	assert.Contains(t, filename, "ipam-export")
	assert.Contains(t, filename, ".csv")
	body := string(data)
	assert.Contains(t, body, "10.0.0.0/24")
	assert.Contains(t, body, "10.0.0.10")
}

func TestExportFullData_JSON(t *testing.T) {
	repo := newStubImportRepo()
	repo.subnets = []*models.Subnet{
		{ID: 2, NetworkID: 1, NetworkAddress: "192.168.1.0", PrefixLength: 24, Description: "LAN"},
	}
	repo.ips[2] = []*models.IPAddress{
		{ID: 10, SubnetID: 2, Address: "192.168.1.100", Hostname: "laptop", Status: "active"},
	}
	svc := newTestImportService(repo)

	data, filename, ct, err := svc.ExportFullData(context.Background(), "json")
	require.NoError(t, err)
	assert.Equal(t, "application/json", ct)
	assert.Contains(t, filename, ".json")
	body := string(data)
	assert.Contains(t, body, "192.168.1.0/24")
	assert.Contains(t, body, "192.168.1.100")
	assert.Contains(t, body, `"subnets"`)
	assert.Contains(t, body, `"ip_addresses"`)
}

func TestExportFullData_DefaultIsCsv(t *testing.T) {
	repo := newStubImportRepo()
	repo.subnets = []*models.Subnet{
		{ID: 1, NetworkAddress: "10.0.0.0", PrefixLength: 24},
	}
	svc := newTestImportService(repo)

	_, _, ct, err := svc.ExportFullData(context.Background(), "")
	require.NoError(t, err)
	assert.Equal(t, "text/csv", ct)
}

func TestExportFullData_EmptyDatabase(t *testing.T) {
	repo := newStubImportRepo()
	svc := newTestImportService(repo)

	data, filename, ct, err := svc.ExportFullData(context.Background(), "csv")
	require.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.NotEmpty(t, filename)
	assert.Equal(t, "text/csv", ct)
}

func TestExportV2MigrationBundle_ZipContents(t *testing.T) {
	repo := newStubImportRepo()
	repo.subnets = []*models.Subnet{
		{ID: 1, NetworkID: 1, NetworkAddress: "10.0.0.0", PrefixLength: 24, Description: "Corp"},
	}
	repo.ips[1] = []*models.IPAddress{
		{ID: 1, SubnetID: 1, Address: "10.0.0.10", Hostname: "server1", Status: "active"},
	}
	svc := newTestImportService(repo)

	data, filename, ct, err := svc.ExportV2MigrationBundle(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "application/zip", ct)
	assert.Contains(t, filename, "padduck-v2-migration-bundle")
	assert.Contains(t, filename, ".zip")

	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	require.NoError(t, err)

	files := map[string]string{}
	for _, f := range zr.File {
		rc, err := f.Open()
		require.NoError(t, err)
		body, err := io.ReadAll(rc)
		require.NoError(t, err)
		require.NoError(t, rc.Close())
		files[f.Name] = string(body)
	}

	assert.Contains(t, files, "manifest.json")
	assert.Contains(t, files, "data/padduck-v1-export.json")
	assert.Contains(t, files, "data/padduck-v1-export.csv")
	assert.Contains(t, files, "README.md")
	assert.Contains(t, files["data/padduck-v1-export.json"], "10.0.0.10")
	assert.Contains(t, files["data/padduck-v1-export.csv"], "10.0.0.0/24")

	var manifest migrationBundleManifest
	require.NoError(t, json.Unmarshal([]byte(files["manifest.json"]), &manifest))
	assert.Equal(t, "v1-pre-v2", manifest.BundleVersion)
	assert.Equal(t, "padduck-v2", manifest.Target)
	require.Len(t, manifest.Files, 3)
	assert.Equal(t, "data/padduck-v1-export.json", manifest.Files[0].Path)
	assert.NotEmpty(t, manifest.Files[0].SHA256)
}

// ─────────────────────────────────────────────────────────────────────────────
// Helper function unit tests
// ─────────────────────────────────────────────────────────────────────────────

func TestParseCIDR(t *testing.T) {
	tests := []struct {
		cidr      string
		wantNet   string
		wantPfx   int
		wantError bool
	}{
		{"10.0.0.0/24", "10.0.0.0", 24, false},
		{"192.168.1.0/16", "192.168.0.0", 16, false},
		{"not-a-cidr", "", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.cidr, func(t *testing.T) {
			net, pfx, err := parseCIDR(tt.cidr)
			if tt.wantError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantNet, net)
			assert.Equal(t, tt.wantPfx, pfx)
		})
	}
}

func TestMaskToPrefixLen(t *testing.T) {
	tests := []struct {
		mask      string
		wantPfx   int
		wantError bool
	}{
		{"24", 24, false},
		{"16", 16, false},
		{"255.255.255.0", 24, false},
		{"255.255.0.0", 16, false},
		{"255.0.0.0", 8, false},
		{"bad-mask", 0, true},
		{"256.0.0.0", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.mask, func(t *testing.T) {
			pfx, err := maskToPrefixLen(tt.mask)
			if tt.wantError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantPfx, pfx)
		})
	}
}
