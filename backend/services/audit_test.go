package services

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedactSensitiveValueRedactsSNMPCommunity(t *testing.T) {
	redacted := redactSensitiveValue(map[string]interface{}{
		"hostname":       "router-1",
		"snmp_community": "public",
		"nested": map[string]interface{}{
			"api_token": "secret-token",
		},
	}).(map[string]interface{})

	assert.Equal(t, "router-1", redacted["hostname"])
	assert.Equal(t, redactedValue, redacted["snmp_community"])
	nested := redacted["nested"].(map[string]interface{})
	assert.Equal(t, redactedValue, nested["api_token"])
}

func TestRedactSensitiveJSON(t *testing.T) {
	raw := `{"snmp_community":"public","name":"scan"}`
	redacted := RedactSensitiveJSON(&raw)
	require.NotNil(t, redacted)

	var decoded map[string]string
	require.NoError(t, json.Unmarshal([]byte(*redacted), &decoded))
	assert.Equal(t, redactedValue, decoded["snmp_community"])
	assert.Equal(t, "scan", decoded["name"])
}
