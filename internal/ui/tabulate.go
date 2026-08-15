package ui

import (
	"fmt"
	"strings"
)

// FormatTable formats 2D rows into plain aligned text table (matching python tabulate tablefmt="plain")
func FormatTable(headers []string, rows [][]string) string {
	if len(rows) == 0 {
		return ""
	}

	numCols := len(headers)
	widths := make([]int, numCols)

	for i, h := range headers {
		widths[i] = len(h)
	}

	for _, row := range rows {
		for i, val := range row {
			if i < numCols && len(val) > widths[i] {
				widths[i] = len(val)
			}
		}
	}

	var sb strings.Builder

	// Header line
	for i, h := range headers {
		sb.WriteString(fmt.Sprintf("%-*s", widths[i], h))
		if i < numCols-1 {
			sb.WriteString("  ")
		}
	}
	sb.WriteString("\n")

	// Data lines
	for _, row := range rows {
		for i, val := range row {
			if i < numCols {
				sb.WriteString(fmt.Sprintf("%-*s", widths[i], val))
				if i < numCols-1 {
					sb.WriteString("  ")
				}
			}
		}
		sb.WriteString("\n")
	}

	return strings.TrimRight(sb.String(), "\n")
}
