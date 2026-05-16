package pdns

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestClient(srv *httptest.Server) *Client {
	return NewClient(srv.URL, "test-key")
}

// ---------------------------------------------------------------------------
// TestConnection
// ---------------------------------------------------------------------------

func TestTestConnection_200_ReturnsNil(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-key", r.Header.Get("X-API-Key"))
		assert.Equal(t, "/api/v1/servers/localhost", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	err := newTestClient(srv).TestConnection(context.Background())
	assert.NoError(t, err)
}

func TestTestConnection_500_ReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	err := newTestClient(srv).TestConnection(context.Background())
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// ListZones
// ---------------------------------------------------------------------------

func TestListZones_ReturnsZones(t *testing.T) {
	zones := []Zone{{ID: "example.com.", Name: "example.com.", Kind: "Native"}}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/servers/localhost/zones", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(zones)
	}))
	defer srv.Close()

	got, err := newTestClient(srv).ListZones(context.Background())
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "example.com.", got[0].Name)
}

func TestListZones_ErrorStatus_ReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	_, err := newTestClient(srv).ListZones(context.Background())
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// GetZone
// ---------------------------------------------------------------------------

func TestGetZone_ReturnsZoneDetail(t *testing.T) {
	detail := ZoneDetail{
		Zone:   Zone{ID: "example.com.", Name: "example.com.", Kind: "Native"},
		RRSets: []RRSet{{Name: "www.example.com.", Type: "A", TTL: 300, Records: []Record{{Content: "1.2.3.4"}}}},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/servers/localhost/zones/example.com.", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(detail)
	}))
	defer srv.Close()

	got, err := newTestClient(srv).GetZone(context.Background(), "example.com.")
	require.NoError(t, err)
	assert.Equal(t, "example.com.", got.Name)
	require.Len(t, got.RRSets, 1)
	assert.Equal(t, "1.2.3.4", got.RRSets[0].Records[0].Content)
}

func TestGetZone_NotFound_ReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	_, err := newTestClient(srv).GetZone(context.Background(), "missing.com.")
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// CreateRecord
// ---------------------------------------------------------------------------

func TestCreateRecord_SendsPatchRequest(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]interface{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	err := newTestClient(srv).CreateRecord(context.Background(), "example.com.", "www.example.com.", "A", "1.2.3.4", 300)
	require.NoError(t, err)
	assert.Equal(t, http.MethodPatch, gotMethod)
	assert.Equal(t, "/api/v1/servers/localhost/zones/example.com.", gotPath)
	rrsets, ok := gotBody["rrsets"].([]interface{})
	require.True(t, ok)
	assert.Len(t, rrsets, 1)
}

func TestCreateRecord_ErrorStatus_ReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
	}))
	defer srv.Close()

	err := newTestClient(srv).CreateRecord(context.Background(), "example.com.", "www.example.com.", "A", "1.2.3.4", 300)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// DeleteRecord
// ---------------------------------------------------------------------------

func TestDeleteRecord_SendsPatchWithDeleteChangeType(t *testing.T) {
	var gotBody map[string]interface{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	err := newTestClient(srv).DeleteRecord(context.Background(), "example.com.", "www.example.com.", "A")
	require.NoError(t, err)
	rrsets, ok := gotBody["rrsets"].([]interface{})
	require.True(t, ok)
	first := rrsets[0].(map[string]interface{})
	assert.Equal(t, "DELETE", first["changetype"])
}
