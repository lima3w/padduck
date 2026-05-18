package technitium

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// fake server helpers
// ---------------------------------------------------------------------------

// okEnvelope wraps a response payload in the Technitium {status:ok, response:...} envelope.
func okEnvelope(t *testing.T, response any) string {
	t.Helper()
	inner, err := json.Marshal(response)
	require.NoError(t, err)
	outer, err := json.Marshal(map[string]any{
		"status":   "ok",
		"response": json.RawMessage(inner),
	})
	require.NoError(t, err)
	return string(outer)
}

func okBare(t *testing.T) string {
	t.Helper()
	b, err := json.Marshal(map[string]string{"status": "ok"})
	require.NoError(t, err)
	return string(b)
}

func errEnvelope(t *testing.T, msg string) string {
	t.Helper()
	b, err := json.Marshal(map[string]string{"status": "error", "errorMessage": msg})
	require.NoError(t, err)
	return string(b)
}

// reqCapture holds the query params and Authorization header from the last request.
type reqCapture struct {
	params url.Values
	auth   string
}

// fakeServer starts an httptest.Server that records the last request and
// responds with body. Returns the server, client, and a pointer to captured request data.
func fakeServer(t *testing.T, path, body string) (*httptest.Server, *Client, *reqCapture) {
	t.Helper()
	cap := &reqCapture{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != path {
			http.NotFound(w, r)
			return
		}
		cap.params = r.URL.Query()
		cap.auth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	c := NewClient(srv.URL, "test-token", false)
	return srv, c, cap
}

// ---------------------------------------------------------------------------
// TestConnection
// ---------------------------------------------------------------------------

func TestConnection_Success(t *testing.T) {
	_, c, cap := fakeServer(t, "/api/zones/list", okBare(t))
	err := c.TestConnection(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "Bearer test-token", cap.auth)
	assert.Empty(t, cap.params.Get("token"), "token must not appear in query string")
}

func TestConnection_APIError(t *testing.T) {
	_, c, _ := fakeServer(t, "/api/zones/list", errEnvelope(t, "invalid token"))
	err := c.TestConnection(context.Background())
	assert.ErrorContains(t, err, "invalid token")
}

func TestConnection_HTTP404(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(srv.Close)
	c := NewClient(srv.URL, "tok", false)
	err := c.TestConnection(context.Background())
	assert.ErrorContains(t, err, "404")
}

// ---------------------------------------------------------------------------
// ListZones
// ---------------------------------------------------------------------------

func TestListZones_ReturnsZones(t *testing.T) {
	body := okEnvelope(t, map[string]any{
		"zones": []map[string]any{
			{"name": "example.com", "type": "Primary", "disabled": false},
			{"name": "internal.local", "type": "Primary", "disabled": true},
		},
	})
	_, c, cap := fakeServer(t, "/api/zones/list", body)
	zones, err := c.ListZones(context.Background())
	require.NoError(t, err)
	require.Len(t, zones, 2)
	assert.Equal(t, "example.com", zones[0].Name)
	assert.Equal(t, "Primary", zones[0].Type)
	assert.Equal(t, "Bearer test-token", cap.auth)
	assert.Empty(t, cap.params.Get("token"), "token must not appear in query string")
}

func TestListZones_Empty(t *testing.T) {
	body := okEnvelope(t, map[string]any{"zones": []any{}})
	_, c, _ := fakeServer(t, "/api/zones/list", body)
	zones, err := c.ListZones(context.Background())
	assert.NoError(t, err)
	assert.Empty(t, zones)
}

// ---------------------------------------------------------------------------
// GetZoneRecords — query parameters
// ---------------------------------------------------------------------------

func TestGetZoneRecords_SendsRequiredParams(t *testing.T) {
	body := okEnvelope(t, map[string]any{"records": []any{}})
	_, c, cap := fakeServer(t, "/api/zones/records/get", body)
	_, err := c.GetZoneRecords(context.Background(), "example.com")
	require.NoError(t, err)
	assert.Equal(t, "example.com", cap.params.Get("zone"))
	assert.Equal(t, "example.com", cap.params.Get("domain"), "domain param is required by Technitium API")
	assert.Equal(t, "true", cap.params.Get("listZone"), "listZone=true is required to get all records")
}

func TestGetZoneRecords_ParsesARecord(t *testing.T) {
	body := okEnvelope(t, map[string]any{
		"records": []map[string]any{
			{"name": "host.example.com", "type": "A", "ttl": 300,
				"rData": map[string]string{"ipAddress": "192.168.1.10"}},
		},
	})
	_, c, _ := fakeServer(t, "/api/zones/records/get", body)
	records, err := c.GetZoneRecords(context.Background(), "example.com")
	require.NoError(t, err)
	require.Len(t, records, 1)
	assert.Equal(t, "host.example.com", records[0].Name)
	assert.Equal(t, "A", records[0].Type)
	assert.Equal(t, 300, records[0].TTL)
	assert.Equal(t, "192.168.1.10", records[0].Content())
}

// ---------------------------------------------------------------------------
// AddRecord
// ---------------------------------------------------------------------------

func TestAddRecord_SendsCorrectParams(t *testing.T) {
	_, c, cap := fakeServer(t, "/api/zones/records/add", okBare(t))
	err := c.AddRecord(context.Background(), "example.com", "host.example.com", "10.0.0.1")
	require.NoError(t, err)
	assert.Equal(t, "example.com", cap.params.Get("zone"))
	assert.Equal(t, "host.example.com", cap.params.Get("domain"))
	assert.Equal(t, "A", cap.params.Get("type"))
	assert.Equal(t, "10.0.0.1", cap.params.Get("ipAddress"))
}

func TestAddRecord_APIError(t *testing.T) {
	_, c, _ := fakeServer(t, "/api/zones/records/add", errEnvelope(t, "Parameter 'domain' missing."))
	err := c.AddRecord(context.Background(), "example.com", "host.example.com", "10.0.0.1")
	assert.ErrorContains(t, err, "Parameter 'domain' missing.")
}

// ---------------------------------------------------------------------------
// DeleteRecord
// ---------------------------------------------------------------------------

func TestDeleteRecord_SendsCorrectParams(t *testing.T) {
	_, c, cap := fakeServer(t, "/api/zones/records/delete", okBare(t))
	err := c.DeleteRecord(context.Background(), "example.com", "host.example.com", "")
	require.NoError(t, err)
	assert.Equal(t, "example.com", cap.params.Get("zone"))
	assert.Equal(t, "host.example.com", cap.params.Get("domain"))
	assert.Equal(t, "A", cap.params.Get("type"))
}

// ---------------------------------------------------------------------------
// Record.Content() — one case per record type
// ---------------------------------------------------------------------------

func rdata(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return json.RawMessage(b)
}

func TestRecordContent(t *testing.T) {
	cases := []struct {
		rtype   string
		rdata   json.RawMessage
		want    string
	}{
		{"A", rdata(map[string]string{"ipAddress": "1.2.3.4"}), "1.2.3.4"},
		{"AAAA", rdata(map[string]string{"ipAddress": "::1"}), "::1"},
		{"PTR", rdata(map[string]string{"ptrdName": "host.example.com"}), "host.example.com"},
		{"CNAME", rdata(map[string]string{"cname": "alias.example.com"}), "alias.example.com"},
		{"MX", rdata(map[string]string{"exchange": "mail.example.com"}), "mail.example.com"},
		{"NS", rdata(map[string]string{"nsDomainName": "ns1.example.com"}), "ns1.example.com"},
		{"TXT", rdata(map[string]string{"text": "v=spf1 include:example.com ~all"}), "v=spf1 include:example.com ~all"},
		{"SOA", rdata(map[string]string{"primaryNameServer": "ns1.example.com"}), "ns1.example.com"},
		{"UNKNOWN", rdata(map[string]string{"ipAddress": "9.9.9.9"}), "9.9.9.9"}, // default branch falls through to ipAddress
	}
	for _, tc := range cases {
		t.Run(tc.rtype+"_"+tc.want, func(t *testing.T) {
			r := &Record{Type: tc.rtype, RData: tc.rdata}
			assert.Equal(t, tc.want, r.Content())
		})
	}
}
