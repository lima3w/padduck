package export

import (
	"bytes"
	"fmt"
	"strings"
)

// GeneratePDF creates a minimal PDF document with a title, headers, and rows.
// It produces a standards-compliant PDF 1.4 without external dependencies.
// Layout: A4 portrait, title at top, plain text table.
func GeneratePDF(title string, headers []string, rows [][]string) ([]byte, error) {
	var body bytes.Buffer

	// We build a simple PDF manually.
	// Object offsets for the cross-reference table.
	offsets := make([]int, 0, 8)

	// Helper to write an object and record its offset.
	writeObj := func(id int, content string) {
		offsets = append(offsets, body.Len())
		fmt.Fprintf(&body, "%d 0 obj\n%s\nendobj\n", id, content)
	}

	// Build page content stream.
	var stream bytes.Buffer
	y := 800.0
	stream.WriteString("BT\n")
	fmt.Fprintf(&stream, "/F1 16 Tf\n")
	fmt.Fprintf(&stream, "50 %.0f Td\n", y)
	escapedTitle := pdfEscape(title)
	fmt.Fprintf(&stream, "(%s) Tj\n", escapedTitle)
	stream.WriteString("ET\n")
	y -= 30

	// Headers
	stream.WriteString("BT\n")
	fmt.Fprintf(&stream, "/F1 10 Tf\n")
	fmt.Fprintf(&stream, "50 %.0f Td\n", y)
	headerLine := strings.Join(headers, "  |  ")
	fmt.Fprintf(&stream, "(%s) Tj\n", pdfEscape(headerLine))
	stream.WriteString("ET\n")
	y -= 6

	// Divider line
	fmt.Fprintf(&stream, "50 %.0f m 550 %.0f l S\n", y, y)
	y -= 14

	// Rows
	for _, row := range rows {
		if y < 50 {
			// Simple page overflow protection — truncate.
			break
		}
		// Alternating shading
		stream.WriteString("BT\n")
		fmt.Fprintf(&stream, "/F1 9 Tf\n")
		fmt.Fprintf(&stream, "50 %.0f Td\n", y)
		rowLine := strings.Join(row, "  |  ")
		fmt.Fprintf(&stream, "(%s) Tj\n", pdfEscape(rowLine))
		stream.WriteString("ET\n")
		y -= 12
	}

	streamBytes := stream.Bytes()
	streamLen := len(streamBytes)

	// Object 1: Catalog
	writeObj(1, "<< /Type /Catalog /Pages 2 0 R >>")

	// Object 2: Pages
	writeObj(2, "<< /Type /Pages /Kids [3 0 R] /Count 1 >>")

	// Object 3: Page
	writeObj(3, fmt.Sprintf(
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>",
	))

	// Object 4: Content stream
	writeObj(4, fmt.Sprintf(
		"<< /Length %d >>\nstream\n%sendstream",
		streamLen, string(streamBytes),
	))

	// Object 5: Font
	writeObj(5, "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica /Encoding /WinAnsiEncoding >>")

	// Cross-reference table
	xrefOffset := body.Len()
	numObjs := 5
	fmt.Fprintf(&body, "xref\n0 %d\n", numObjs+1)
	fmt.Fprintf(&body, "0000000000 65535 f \n")
	for _, off := range offsets {
		fmt.Fprintf(&body, "%010d 00000 n \n", off)
	}

	fmt.Fprintf(&body, "trailer\n<< /Size %d /Root 1 0 R >>\n", numObjs+1)
	fmt.Fprintf(&body, "startxref\n%d\n%%%%EOF\n", xrefOffset)

	// Prepend PDF header
	var out bytes.Buffer
	out.WriteString("%PDF-1.4\n")
	out.Write(body.Bytes())

	return out.Bytes(), nil
}

// pdfEscape escapes special PDF string characters.
func pdfEscape(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "(", "\\(")
	s = strings.ReplaceAll(s, ")", "\\)")
	return s
}
