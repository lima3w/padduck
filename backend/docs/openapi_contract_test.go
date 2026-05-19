package docs

import (
	"bytes"
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestOpenAPIContractStableV1(t *testing.T) {
	spec := readOpenAPISpec(t)
	info := spec["info"].(map[string]any)
	if got := info["version"]; got != "1.26.0" {
		t.Fatalf("info.version = %v, want 1.26.0", got)
	}
	if got := info["x-api-contract"]; got != "stable-v1" {
		t.Fatalf("info.x-api-contract = %v, want stable-v1", got)
	}

	paths := spec["paths"].(map[string]any)
	requiredPaths := []string{
		"/api/v1/admin/webhooks/sample-payload",
		"/api/v1/automation/ip-addresses/allocate",
		"/api/v1/automation/ip-addresses/reserve",
		"/api/v1/automation/ip-addresses/{id}/release",
		"/api/v1/automation/dns/update",
		"/api/v1/automation/devices/register",
	}
	for _, path := range requiredPaths {
		if _, ok := paths[path]; !ok {
			t.Fatalf("missing stable API path %s", path)
		}
	}
}

func TestAutomationWritePathsDeclareIdempotencyKey(t *testing.T) {
	spec := readOpenAPISpec(t)
	paths := spec["paths"].(map[string]any)
	writePaths := []string{
		"/api/v1/automation/ip-addresses/allocate",
		"/api/v1/automation/ip-addresses/reserve",
		"/api/v1/automation/ip-addresses/{id}/release",
		"/api/v1/automation/dns/update",
		"/api/v1/automation/devices/register",
	}
	for _, path := range writePaths {
		post := paths[path].(map[string]any)["post"].(map[string]any)
		params, ok := post["parameters"].([]any)
		if !ok {
			t.Fatalf("%s post is missing parameters", path)
		}
		found := false
		for _, param := range params {
			ref, _ := param.(map[string]any)["$ref"].(string)
			if ref == "#/components/parameters/idempotencyKey" {
				found = true
			}
		}
		if !found {
			t.Fatalf("%s post does not declare idempotencyKey", path)
		}
	}
}

func TestOpenAPICopiesStayInSync(t *testing.T) {
	root, err := os.ReadFile("../../docs/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	local, err := os.ReadFile("openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(root, local) {
		t.Fatal("docs/openapi.yaml and backend/docs/openapi.yaml differ")
	}
}

func readOpenAPISpec(t *testing.T) map[string]any {
	t.Helper()
	data, err := os.ReadFile("openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var spec map[string]any
	if err := yaml.Unmarshal(data, &spec); err != nil {
		t.Fatal(err)
	}
	return spec
}
