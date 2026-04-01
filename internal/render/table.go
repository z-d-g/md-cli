package render

import (
	"strings"

	"github.com/z-d-g/md-cli/internal/markdown"

	"charm.land/lipgloss/v2"
)

// tableAlign represents column alignment in a table.
type tableAlign int

const (
	alignLeft tableAlign = iota
	alignCenter
	alignRight
)

// tableContext holds state for a markdown table being rendered.
type tableContext struct {
	active         bool
	alignments     []tableAlign
	widths         []int
	headerRendered bool
}

// renderTable renders a markdown table line with alignment support.
func renderTable(line string, r *lipglossRenderer) string {
	cells := parseTableCells(line)
	if len(cells) == 0 {
		return line
	}

	// First table row — compute widths from the full table block
	if !r.table.active {
		r.computeTableWidths(line)
		r.table.active = true
	}

	if markdown.IsTableSeparatorLine(line) {
		alignments, _ := parseSeparator(cells)
		r.table.alignments = alignments
		r.table.headerRendered = true
		return renderTableSeparator(r.table.widths, r)
	}

	if !r.table.headerRendered {
		return renderTableHeader(cells, r)
	}
	return renderTableRow(cells, r)
}

// computeTableWidths scans ahead to find the table block and pre-computes widths.
func (r *lipglossRenderer) computeTableWidths(currentLine string) {
	currentCells := parseTableCells(currentLine)
	widths := make([]int, len(currentCells))
	for i, cell := range currentCells {
		widths[i] = max(lipgloss.Width(cell), 3)
	}

	if tableLines != nil {
		lines := tableLines()
		foundCurrent := false
		for _, line := range lines {
			if line == currentLine {
				foundCurrent = true
			}
			if !foundCurrent {
				continue
			}
			if !markdown.IsTableLine(line) {
				break
			}
			cells := parseTableCells(line)
			for i, cell := range cells {
				w := max(lipgloss.Width(cell), 3)
				if i >= len(widths) {
					widths = append(widths, w)
				} else if w > widths[i] {
					widths[i] = w
				}
			}
		}
	}

	r.table.widths = widths
}

// ResetTable clears table state between separate tables.
func (r *lipglossRenderer) ResetTable() {
	r.table.active = false
	r.table.alignments = nil
	r.table.widths = nil
	r.table.headerRendered = false
}

// parseTableCells splits a table line into cells.
func parseTableCells(line string) []string {
	trimmed := strings.TrimSpace(line)
	trimmed = strings.TrimPrefix(trimmed, "|")
	trimmed = strings.TrimSuffix(trimmed, "|")
	parts := strings.Split(trimmed, "|")
	cells := make([]string, 0, len(parts))
	for _, p := range parts {
		cells = append(cells, strings.TrimSpace(p))
	}
	return cells
}

// parseSeparator extracts alignment from separator cells.
func parseSeparator(cells []string) ([]tableAlign, []int) {
	alignments := make([]tableAlign, len(cells))
	for i, cell := range cells {
		trimmed := strings.TrimSpace(cell)
		left := strings.HasPrefix(trimmed, ":")
		right := strings.HasSuffix(trimmed, ":")
		switch {
		case left && right:
			alignments[i] = alignCenter
		case right:
			alignments[i] = alignRight
		default:
			alignments[i] = alignLeft
		}
	}
	return alignments, nil
}

// alignText applies alignment to text within a given width.
func alignText(text string, width int, align tableAlign) string {
	textWidth := lipgloss.Width(text)
	if textWidth >= width {
		return text
	}
	pad := width - textWidth
	switch align {
	case alignRight:
		return strings.Repeat(" ", pad) + text
	case alignCenter:
		left := pad / 2
		right := pad - left
		return strings.Repeat(" ", left) + text + strings.Repeat(" ", right)
	default:
		return text + strings.Repeat(" ", pad)
	}
}

// renderTableHeader renders the header row with bold/lilac styling.
func renderTableHeader(cells []string, r *lipglossRenderer) string {
	var b strings.Builder
	b.WriteString(r.styleCache.tableBorderFunc("│"))

	for i, cell := range cells {
		align := alignLeft
		if i < len(r.table.alignments) {
			align = r.table.alignments[i]
		}

		cellWidth := 3
		if i < len(r.table.widths) {
			cellWidth = r.table.widths[i]
		}

		elements := ParseInlineElements(cell)
		styledCell := r.RenderInline(elements, r.styleCache.tableHeaderFunc)
		aligned := alignText(styledCell, cellWidth, align)

		b.WriteString(" ")
		b.WriteString(aligned)
		b.WriteString(" ")
		b.WriteString(r.styleCache.tableBorderFunc("│"))
	}
	return b.String()
}

// renderTableSeparator renders the separator row.
func renderTableSeparator(widths []int, r *lipglossRenderer) string {
	border := r.styleCache.tableBorderFunc

	var b strings.Builder
	b.WriteString(border("├"))
	for i, w := range widths {
		b.WriteString(border(strings.Repeat("─", w+2)))
		if i < len(widths)-1 {
			b.WriteString(border("┼"))
		}
	}
	b.WriteString(border("┤"))
	return b.String()
}

// renderTableRow renders a data row with cell styling.
func renderTableRow(cells []string, r *lipglossRenderer) string {
	var b strings.Builder
	b.WriteString(r.styleCache.tableBorderFunc("│"))

	for i, cell := range cells {
		align := alignLeft
		if i < len(r.table.alignments) {
			align = r.table.alignments[i]
		}

		cellWidth := 3
		if i < len(r.table.widths) {
			cellWidth = r.table.widths[i]
		}

		elements := ParseInlineElements(cell)
		styledCell := r.RenderInline(elements, r.styleCache.tableCellFunc)
		aligned := alignText(styledCell, cellWidth, align)

		b.WriteString(" ")
		b.WriteString(aligned)
		b.WriteString(" ")
		b.WriteString(r.styleCache.tableBorderFunc("│"))
	}
	return b.String()
}
