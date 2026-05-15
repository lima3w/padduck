package export

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeneratePDF_NonEmpty(t *testing.T) {
	data, err := GeneratePDF("Test Report", []string{"Col1", "Col2"}, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, data)
	// Valid PDF starts with %PDF-
	assert.True(t, bytes.HasPrefix(data, []byte("%PDF-")))
}

func TestGeneratePDF_ContainsTitle(t *testing.T) {
	title := "Subnet Utilisation Report"
	data, err := GeneratePDF(title, []string{"Subnet", "Used%"}, [][]string{
		{"10.0.0.0/24", "75.5"},
	})
	require.NoError(t, err)
	assert.Contains(t, string(data), title)
}

func TestGeneratePDF_WithRows(t *testing.T) {
	headers := []string{"CIDR", "Description", "Used", "Total", "Pct"}
	rows := [][]string{
		{"10.0.0.0/24", "Office Network", "50", "254", "19.69"},
		{"192.168.0.0/16", "Private Range", "100", "65534", "0.15"},
	}
	data, err := GeneratePDF("Network Report", headers, rows)
	require.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.True(t, bytes.HasPrefix(data, []byte("%PDF-")))
}

func TestPDFEscape(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"(test)", "\\(test\\)"},
		{"back\\slash", "back\\\\slash"},
		{"a(b)c", "a\\(b\\)c"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expected, pdfEscape(tt.input))
	}
}
