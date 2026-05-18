package scanner

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/gosnmp/gosnmp"
)

// SNMPInterface holds data for a single network interface discovered via SNMP.
type SNMPInterface struct {
	Name       string
	IPAddress  string
	MACAddress string
}

// ARPEntry represents a single ARP table entry discovered via SNMP.
type ARPEntry struct {
	IPAddress  string
	MACAddress string
}

// SNMPResult holds all data retrieved from a single SNMP target.
type SNMPResult struct {
	SysName        string
	SysDescription string
	SysLocation    string
	Interfaces     []SNMPInterface
	ARPEntries     []ARPEntry
}

// SNMPv3Params holds credentials for SNMPv3 User-based Security Model (USM).
// AuthProto and PrivProto accept "MD5", "SHA" and "DES", "AES" respectively.
// Security level is derived automatically: NoAuthNoPriv when AuthPass is empty,
// AuthNoPriv when only AuthPass is set, AuthPriv when both passwords are set.
type SNMPv3Params struct {
	User      string
	AuthProto string
	AuthPass  string
	PrivProto string
	PrivPass  string
}

// ScanSNMP performs an SNMP query against ip. version must be "2c" or "3".
// For version "3", v3 must be non-nil and contain at least a User; community
// is ignored in that case. timeout applies per request.
func ScanSNMP(ctx context.Context, ip, community, version string, v3 *SNMPv3Params, timeout time.Duration) (*SNMPResult, error) {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	if community == "" {
		community = "public"
	}

	params := &gosnmp.GoSNMP{
		Target:  ip,
		Port:    161,
		Timeout: timeout,
		Retries: 1,
		MaxOids: gosnmp.MaxOids,
	}

	if version == "3" && v3 != nil {
		params.Version = gosnmp.Version3
		params.SecurityModel = gosnmp.UserSecurityModel
		usp := &gosnmp.UsmSecurityParameters{UserName: v3.User}
		switch {
		case v3.AuthPass != "" && v3.PrivPass != "":
			params.MsgFlags = gosnmp.AuthPriv
			usp.AuthenticationProtocol = snmpAuthProto(v3.AuthProto)
			usp.AuthenticationPassphrase = v3.AuthPass
			usp.PrivacyProtocol = snmpPrivProto(v3.PrivProto)
			usp.PrivacyPassphrase = v3.PrivPass
		case v3.AuthPass != "":
			params.MsgFlags = gosnmp.AuthNoPriv
			usp.AuthenticationProtocol = snmpAuthProto(v3.AuthProto)
			usp.AuthenticationPassphrase = v3.AuthPass
		default:
			params.MsgFlags = gosnmp.NoAuthNoPriv
		}
		params.SecurityParameters = usp
	} else {
		params.Version = gosnmp.Version2c
		params.Community = community
	}

	// Respect context cancellation.
	done := make(chan struct{})
	var connectErr error
	go func() {
		defer close(done)
		connectErr = params.Connect()
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-done:
	}
	if connectErr != nil {
		return nil, fmt.Errorf("snmp connect: %w", connectErr)
	}
	defer params.Conn.Close()

	result := &SNMPResult{}

	// --- System info (SNMPv2-MIB) ---
	sysOIDs := []string{
		"1.3.6.1.2.1.1.1.0", // sysDescr
		"1.3.6.1.2.1.1.5.0", // sysName
		"1.3.6.1.2.1.1.6.0", // sysLocation
	}
	pkt, err := params.Get(sysOIDs)
	if err == nil && pkt != nil {
		for _, v := range pkt.Variables {
			switch v.Name {
			case ".1.3.6.1.2.1.1.1.0", "1.3.6.1.2.1.1.1.0":
				result.SysDescription = snmpString(v)
			case ".1.3.6.1.2.1.1.5.0", "1.3.6.1.2.1.1.5.0":
				result.SysName = snmpString(v)
			case ".1.3.6.1.2.1.1.6.0", "1.3.6.1.2.1.1.6.0":
				result.SysLocation = snmpString(v)
			}
		}
	}

	// --- Interfaces (IF-MIB ifDescr 1.3.6.1.2.1.2.2.1.2) ---
	ifDescrs := map[string]string{} // ifIndex -> ifDescr
	err = params.Walk("1.3.6.1.2.1.2.2.1.2", func(pdu gosnmp.SnmpPDU) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		idx := lastOIDComponent(pdu.Name)
		ifDescrs[idx] = snmpString(pdu)
		return nil
	})
	if err != nil && ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// ifPhysAddress (1.3.6.1.2.1.2.2.1.6) for MAC
	ifMACs := map[string]string{}
	_ = params.Walk("1.3.6.1.2.1.2.2.1.6", func(pdu gosnmp.SnmpPDU) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		idx := lastOIDComponent(pdu.Name)
		if b, ok := pdu.Value.([]byte); ok && len(b) == 6 {
			ifMACs[idx] = fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", b[0], b[1], b[2], b[3], b[4], b[5])
		}
		return nil
	})

	for idx, descr := range ifDescrs {
		result.Interfaces = append(result.Interfaces, SNMPInterface{
			Name:       descr,
			MACAddress: ifMACs[idx],
		})
	}

	// --- ARP table (IP-MIB ipNetToMediaPhysAddress 1.3.6.1.2.1.4.22.1.2) ---
	// OID suffix encodes: ifIndex.a.b.c.d  — extract the last 4 components as IP
	_ = params.Walk("1.3.6.1.2.1.4.22.1.2", func(pdu gosnmp.SnmpPDU) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		oidIP := arpIPFromOID(pdu.Name)
		if oidIP == "" {
			return nil
		}
		if b, ok := pdu.Value.([]byte); ok && len(b) == 6 {
			mac := fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", b[0], b[1], b[2], b[3], b[4], b[5])
			result.ARPEntries = append(result.ARPEntries, ARPEntry{
				IPAddress:  oidIP,
				MACAddress: mac,
			})
		}
		return nil
	})

	return result, nil
}

// snmpString extracts a string value from an SNMP PDU.
func snmpString(pdu gosnmp.SnmpPDU) string {
	switch v := pdu.Value.(type) {
	case []byte:
		return string(v)
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

// lastOIDComponent returns the last numeric component of an OID string (the ifIndex).
func lastOIDComponent(oid string) string {
	for i := len(oid) - 1; i >= 0; i-- {
		if oid[i] == '.' {
			return oid[i+1:]
		}
	}
	return oid
}

// arpIPFromOID extracts the IP address from an ipNetToMediaPhysAddress OID.
// OID format: 1.3.6.1.2.1.4.22.1.2.<ifIndex>.<a>.<b>.<c>.<d>
func arpIPFromOID(oid string) string {
	// strip leading dot if present
	if len(oid) > 0 && oid[0] == '.' {
		oid = oid[1:]
	}
	// prefix to strip: 1.3.6.1.2.1.4.22.1.2.
	const prefix = "1.3.6.1.2.1.4.22.1.2."
	if len(oid) <= len(prefix) {
		return ""
	}
	rest := oid[len(prefix):]
	// rest = <ifIndex>.<a>.<b>.<c>.<d>
	// find second dot to skip ifIndex
	dotIdx := -1
	for i, c := range rest {
		if c == '.' {
			dotIdx = i
			break
		}
	}
	if dotIdx < 0 {
		return ""
	}
	ipPart := rest[dotIdx+1:]
	// parse dotted-decimal IP
	parsed := net.ParseIP(ipPart)
	if parsed == nil {
		return ""
	}
	return parsed.String()
}

// snmpAuthProto maps a protocol name string to the gosnmp constant.
// Defaults to MD5 for unrecognised values.
func snmpAuthProto(s string) gosnmp.SnmpV3AuthProtocol {
	switch s {
	case "SHA", "sha":
		return gosnmp.SHA
	default:
		return gosnmp.MD5
	}
}

// snmpPrivProto maps a protocol name string to the gosnmp constant.
// Defaults to DES for unrecognised values.
func snmpPrivProto(s string) gosnmp.SnmpV3PrivProtocol {
	switch s {
	case "AES", "aes":
		return gosnmp.AES
	default:
		return gosnmp.DES
	}
}
