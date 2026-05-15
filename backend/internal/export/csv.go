package export

import (
	"bytes"
	"encoding/csv"
)

// GenerateCSV converts a slice of maps (column->value) to CSV bytes with UTF-8 BOM.
// headers defines column order.
func GenerateCSV(headers []string, rows []map[string]string) ([]byte, error) {
	var buf bytes.Buffer
	// Write UTF-8 BOM for Excel compatibility
	buf.Write([]byte{0xEF, 0xBB, 0xBF})

	w := csv.NewWriter(&buf)

	if err := w.Write(headers); err != nil {
		return nil, err
	}

	for _, row := range rows {
		record := make([]string, len(headers))
		for i, h := range headers {
			record[i] = row[h]
		}
		if err := w.Write(record); err != nil {
			return nil, err
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
