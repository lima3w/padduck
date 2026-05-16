package scanner

import (
	"testing"

	"github.com/gosnmp/gosnmp"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// lastOIDComponent — extracts ifIndex from OID string
// ---------------------------------------------------------------------------

func TestLastOIDComponent_ReturnsLastPart(t *testing.T) {
	assert.Equal(t, "5", lastOIDComponent("1.3.6.1.2.1.2.2.1.2.5"))
}

func TestLastOIDComponent_SingleComponent_ReturnsSelf(t *testing.T) {
	assert.Equal(t, "42", lastOIDComponent("42"))
}

// ---------------------------------------------------------------------------
// arpIPFromOID — extracts IP from ipNetToMediaPhysAddress OID
// ---------------------------------------------------------------------------

func TestArpIPFromOID_ValidOID_ReturnsIP(t *testing.T) {
	oid := "1.3.6.1.2.1.4.22.1.2.1.192.168.1.100"
	got := arpIPFromOID(oid)
	assert.Equal(t, "192.168.1.100", got)
}

func TestArpIPFromOID_WithLeadingDot_ReturnsIP(t *testing.T) {
	oid := ".1.3.6.1.2.1.4.22.1.2.2.10.0.0.1"
	got := arpIPFromOID(oid)
	assert.Equal(t, "10.0.0.1", got)
}

func TestArpIPFromOID_TooShort_ReturnsEmpty(t *testing.T) {
	got := arpIPFromOID("1.3.6.1.2.1.4.22.1.2")
	assert.Equal(t, "", got)
}

func TestArpIPFromOID_InvalidIP_ReturnsEmpty(t *testing.T) {
	got := arpIPFromOID("1.3.6.1.2.1.4.22.1.2.1.999.999.999.999")
	assert.Equal(t, "", got)
}

// ---------------------------------------------------------------------------
// snmpString — extracts string value from SNMP PDU
// ---------------------------------------------------------------------------

func TestSnmpString_ByteSlice_ReturnsString(t *testing.T) {
	pdu := gosnmp.SnmpPDU{Value: []byte("router-01")}
	assert.Equal(t, "router-01", snmpString(pdu))
}

func TestSnmpString_String_ReturnsString(t *testing.T) {
	pdu := gosnmp.SnmpPDU{Value: "datacenter"}
	assert.Equal(t, "datacenter", snmpString(pdu))
}

func TestSnmpString_Integer_ReturnsFormattedString(t *testing.T) {
	pdu := gosnmp.SnmpPDU{Value: 42}
	assert.Equal(t, "42", snmpString(pdu))
}
