package export

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateCSV_EmptyRows(t *testing.T) {
	headers := []string{"col1", "col2"}
	data, err := GenerateCSV(headers, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, data)
	// Should contain the UTF-8 BOM
	assert.True(t, bytes.HasPrefix(data, []byte{0xEF, 0xBB, 0xBF}))
}

func TestGenerateCSV_WithRows(t *testing.T) {
	headers := []string{"subnet", "description", "used", "total"}
	rows := []map[string]string{
		{"subnet": "10.0.0.0/24", "description": "Test subnet", "used": "50", "total": "254"},
		{"subnet": "192.168.1.0/24", "description": "Another subnet", "used": "10", "total": "254"},
	}
	data, err := GenerateCSV(headers, rows)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Content should include header and both row values
	content := string(data[3:]) // skip BOM
	assert.Contains(t, content, "subnet")
	assert.Contains(t, content, "10.0.0.0/24")
	assert.Contains(t, content, "192.168.1.0/24")
}

func TestGenerateCSV_ColumnOrder(t *testing.T) {
	headers := []string{"b", "a", "c"}
	rows := []map[string]string{
		{"a": "alpha", "b": "beta", "c": "gamma"},
	}
	data, err := GenerateCSV(headers, rows)
	require.NoError(t, err)

	content := string(data[3:])
	// Header row should be in order b, a, c
	assert.Contains(t, content, "b,a,c")
	// Data row should follow same order
	assert.Contains(t, content, "beta,alpha,gamma")
}
