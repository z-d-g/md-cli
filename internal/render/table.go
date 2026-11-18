package render

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/z-d-g/md-cli/v2/internal/markdown"
)

// tableAlign represents column alignment in a table.
type tableAlign int

const (
	alignLeft tableAlign = iota
	alignCenter
	alignRight
)

// tableDoc is the parsed view of the whole document that table rendering needs.
type tableDoc struct {
	lines []string
	// bounds[i] is the inclusive [start,end] line range of the table containing
	// line i, or nil when line i is not in a table.
	bounds [][2]int
	widths [][]int
	aligns [][]tableAlign
	header []bool
	sep    []bool
}

func newTableDoc(lines []string) *tableDoc {
	d := &tableDoc{lines: lines}
	d.index()
	return d
}

func (d *tableDoc) index() {
	n := len(d.lines)
	d.bounds = make([][2]int, n)
	d.widths = make([][]int, n)
	d.aligns = make([][]tableAlign, n)
	d.header = make([]bool, n)
	d.sep = make([]bool, n)

	doc := markdown.NewDoc(strings.Join(d.lines, "\n"))
	for i := 0; i < n; i++ {
		if !doc.IsTable(i) {
			continue
		}
		start, end := i, i
		for end+1 < n && doc.IsTable(end+1) && strings.TrimSpace(d.lines[end+1]) != "" {
			end++
		}
		d.addBlock(start, end, doc)
		i = end
	}
}

func (d *tableDoc) addBlock(start, end int, doc *markdown.Doc) {
	var widths []int
	var aligns []tableAlign
	headerLine := -1
	for i := start; i <= end; i++ {
		cells := parseTableCells(d.lines[i])
		sep := doc.IsTableSeparator(i)
		if sep {
			aligns, _ = parseSeparator(cells)
			headerLine = i - 1
			continue
		}
		for c, cell := range cells {
			w := max(ansi.StringWidth(cell), 3)
			if c >= len(widths) {
				widths = append(widths, w)
				continue
			}
			widths[c] = max(widths[c], w)
		}
	}
	for i := start; i <= end; i++ {
		d.bounds[i] = [2]int{start, end}
		d.widths[i] = widths
		d.aligns[i] = aligns
		d.header[i] = i == headerLine
		d.sep[i] = doc.IsTableSeparator(i)
	}
}

// tableDocFor returns the table index for the renderer's current document,
// rebuilding it only when the document snapshot changes.
func (r *lipglossRenderer) tableDocFor() *tableDoc {
	lines := r.documentLines()
	if r.tableDocCache != nil && r.tableDocLines == len(lines) && sameLines(r.tableDocRaw, lines) {
		return r.tableDocCache
	}
	d := newTableDoc(lines)
	r.tableDocCache = d
	r.tableDocRaw = lines
	r.tableDocLines = len(lines)
	return d
}

func sameLines(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// renderTable renders the table line at document row idx. Widths and
// alignment come from the table block that row belongs to, so two tables with
// identical separator text never share state.
func renderTable(line string, idx int, r *lipglossRenderer) string {
	d := r.tableDocFor()
	if idx < 0 || idx >= len(d.widths) {
		return line
	}
	widths, aligns := d.widths[idx], d.aligns[idx]
	if d.sep[idx] {
		return renderTableSeparator(widths, r)
	}
	return renderTableRow(parseTableCells(line), r, widths, aligns, d.header[idx])
}

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
		left := strings.HasPrefix(cell, ":")
		right := strings.HasSuffix(cell, ":")
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
	textWidth := ansi.StringWidth(text)
	if textWidth >= width {
		return text
	}
	pad := width - textWidth
	switch align {
	case alignRight:
		return strings.Repeat(" ", pad) + text
	case alignCenter:
		left := pad / 2
		return strings.Repeat(" ", left) + text + strings.Repeat(" ", pad-left)
	default:
		return text + strings.Repeat(" ", pad)
	}
}

func cellAlign(aligns []tableAlign, i int) tableAlign {
	if i < len(aligns) {
		return aligns[i]
	}
	return alignLeft
}

func cellWidth(widths []int, i int) int {
	if i < len(widths) {
		return widths[i]
	}
	return 3
}

// renderTableRow renders one table row with cell styling and alignment.
func renderTableRow(cells []string, r *lipglossRenderer, widths []int, aligns []tableAlign, header bool) string {
	var b strings.Builder
	b.WriteString(r.styles.TableBorder.Render("│"))

	for i, cell := range cells {
		elements := ParseInlineElements(cell)
		var styled string
		if header {
			styled = r.RenderInline(elements, r.styles.TableHeader)
		} else {
			styled = r.RenderInline(elements, r.styles.TableCell)
		}
		b.WriteString(" ")
		b.WriteString(alignText(styled, cellWidth(widths, i), cellAlign(aligns, i)))
		b.WriteString(" ")
		b.WriteString(r.styles.TableBorder.Render("│"))
	}
	return b.String()
}

// renderTableSeparator renders the separator row.
func renderTableSeparator(widths []int, r *lipglossRenderer) string {
	border := r.styles.TableBorder.Render
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
