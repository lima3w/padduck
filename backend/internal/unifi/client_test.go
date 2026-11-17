package unifi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func envelope(w http.ResponseWriter, data any) {
	_ = json.NewEncoder(w).Encode(map[string]any{"meta": map[string]string{"rc": "ok"}, "data": data})
}

// fakeController serves the classic API under prefix. When auth is "key" it
// requires X-API-KEY; when "cookie" it requires the session cookie set by the
// login handlers (loginPath).
func fakeController(t *testing.T, prefix, auth, loginPath string) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	authed := func(r *http.Request) bool {
		if auth == "key" {
			return r.Header.Get("X-API-KEY") == "secret"
		}
		c, err := r.Cookie("sess")
		return err == nil && c.Value == "ok"
	}
	guard := func(h func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if !authed(r) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			h(w, r)
		}
	}
	if loginPath != "" {
		mux.HandleFunc(loginPath, func(w http.ResponseWriter, r *http.Request) {
			var b map[string]string
			_ = json.NewDecoder(r.Body).Decode(&b)
			if b["username"] != "admin" || b["password"] != "pw" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			http.SetCookie(w, &http.Cookie{Name: "sess", Value: "ok", Path: "/"})
			w.WriteHeader(http.StatusOK)
		})
	}
	mux.HandleFunc(prefix+"/api/self/sites", guard(func(w http.ResponseWriter, _ *http.Request) {
		envelope(w, []Site{{Name: "default", Desc: "Default"}, {Name: "lab", Desc: "Lab"}})
	}))
	mux.HandleFunc(prefix+"/api/s/default/rest/networkconf", guard(func(w http.ResponseWriter, _ *http.Request) {
		envelope(w, []map[string]any{
			{"_id": "n1", "name": "Users", "purpose": "corporate", "vlan_enabled": true, "vlan": "10", "ip_subnet": "10.0.10.1/24", "enabled": true},
			{"_id": "n0", "name": "Default", "purpose": "corporate", "ip_subnet": "10.0.1.1/24", "enabled": true},
		})
	}))
	mux.HandleFunc(prefix+"/api/s/default/stat/sta", guard(func(w http.ResponseWriter, _ *http.Request) {
		envelope(w, []map[string]any{{"mac": "aa:bb:cc:00:00:01", "ip": "10.0.10.5", "hostname": "laptop", "name": "Zed's laptop", "network_id": "n1"}})
	}))
	mux.HandleFunc(prefix+"/api/s/default/stat/device", guard(func(w http.ResponseWriter, _ *http.Request) {
		envelope(w, []map[string]any{{"_id": "d1", "mac": "11:22:33:44:55:66", "name": "sw-01", "model": "USW-24", "version": "7.0.1", "type": "usw", "state": 1}})
	}))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func exercise(t *testing.T, api *API) {
	t.Helper()
	ctx := context.Background()
	require.NoError(t, api.TestConnection(ctx))

	nets, err := api.ListNetworks(ctx)
	require.NoError(t, err)
	require.Len(t, nets, 2)
	assert.Equal(t, 10, nets[0].VLANNumber())
	assert.Equal(t, 0, nets[1].VLANNumber(), "untagged network has no VLAN number")

	clients, err := api.ListClients(ctx)
	require.NoError(t, err)
	require.Len(t, clients, 1)
	assert.Equal(t, "Zed's laptop", clients[0].DisplayName(), "alias wins over reported hostname")

	devs, err := api.ListDevices(ctx)
	require.NoError(t, err)
	require.Len(t, devs, 1)
	assert.True(t, devs[0].Online())
	assert.Equal(t, "sw-01", devs[0].DisplayName())
}

func TestAPIKeyAuth(t *testing.T) {
	srv := fakeController(t, "/proxy/network", "key", "")
	exercise(t, New(Config{BaseURL: srv.URL, APIKey: "secret"}))

	err := New(Config{BaseURL: srv.URL, APIKey: "wrong"}).TestConnection(context.Background())
	assert.ErrorContains(t, err, "rejected the credentials")
}

func TestUniFiOSCookieLogin(t *testing.T) {
	srv := fakeController(t, "/proxy/network", "cookie", "/api/auth/login")
	exercise(t, New(Config{BaseURL: srv.URL, Username: "admin", Password: "pw"}))
}

func TestStandaloneControllerCookieLogin(t *testing.T) {
	// No /api/auth/login (404) so the client falls back to /api/login with no path prefix.
	srv := fakeController(t, "", "cookie", "/api/login")
	exercise(t, New(Config{BaseURL: srv.URL, Username: "admin", Password: "pw"}))
}

func TestBadPasswordIsReported(t *testing.T) {
	srv := fakeController(t, "/proxy/network", "cookie", "/api/auth/login")
	err := New(Config{BaseURL: srv.URL, Username: "admin", Password: "nope"}).TestConnection(context.Background())
	assert.ErrorContains(t, err, "check the username and password")
}

func TestMissingCredentials(t *testing.T) {
	err := New(Config{BaseURL: "http://127.0.0.1:1"}).TestConnection(context.Background())
	assert.ErrorContains(t, err, "API key or a username and password")
}

func TestUnknownSite(t *testing.T) {
	srv := fakeController(t, "/proxy/network", "key", "")
	err := New(Config{BaseURL: srv.URL, APIKey: "secret", Site: "nope"}).TestConnection(context.Background())
	assert.ErrorContains(t, err, `site "nope" not found (available: default, lab)`)
}

func TestNetworkVLANAcceptsStringOrNumber(t *testing.T) {
	var nets []Network
	require.NoError(t, json.Unmarshal([]byte(`[
		{"_id":"a","name":"str","vlan_enabled":true,"vlan":"10"},
		{"_id":"b","name":"num","vlan_enabled":true,"vlan":25},
		{"_id":"c","name":"none"}
	]`), &nets))
	assert.Equal(t, 10, nets[0].VLANNumber())
	assert.Equal(t, 25, nets[1].VLANNumber())
	assert.Equal(t, 0, nets[2].VLANNumber())
}

func TestVLANNumberBounds(t *testing.T) {
	assert.Equal(t, 0, Network{VLANEnabled: true, VLAN: "abc"}.VLANNumber())
	assert.Equal(t, 0, Network{VLANEnabled: true, VLAN: "4095"}.VLANNumber())
	assert.Equal(t, 4094, Network{VLANEnabled: true, VLAN: "4094"}.VLANNumber())
}
