package app

import (
	"strings"

	"charm.land/lipgloss/v2"
)

type helpEntry struct {
	key  string
	desc string
}

type helpGroup struct {
	title string
	items []helpEntry
}

var helpGroups = []helpGroup{
	{"File", []helpEntry{{"Ctrl+S", "Save"}, {"Ctrl+Q", "Quit"}}},
	{"Cursor", []helpEntry{{"↑↓←→", "Move"}, {"Home/End", "Line"}, {"Ctrl+←→", "Word"}, {"PgUp/PgDn", "Page"}, {"Ctrl+Home/End", "Doc"}}},
	{"Select", []helpEntry{{"Shift+↑↓←→", "Extend"}, {"Ctrl+A", "All"}, {"Ctrl+D", "Word"}, {"Ctrl+L", "Line"}}},
	{"Edit", []helpEntry{{"Ctrl+C", "Copy"}, {"Ctrl+X", "Cut"}, {"Ctrl+V", "Paste"}, {"Ctrl+Z/Y", "Undo/Redo"}, {"Bksp/Del", "Delete"}, {"Alt+Bksp/Del", "Word del"}}},
}

// helpDialogLines renders the help dialog for the current width. It is cached
// per width because the dialog only depends on the width and the theme.
func (m *Model) helpDialogLines() []string {
	if len(m.dialogLines) > 0 && m.Width == m.dialogWidth {
		return m.dialogLines
	}
	m.dialogWidth = m.Width
	dialog := m.renderHelpDialogContent(
		m.Config.HelpKeyStyle, m.Config.HelpDescStyle,
		m.Config.HelpSectionStyle, m.Config.InfoStyle,
	)
	if dialog == "" {
		m.dialogLines = nil
		return nil
	}
	m.dialogLines = strings.Split(strings.TrimRight(dialog, "\n"), "\n")
	return m.dialogLines
}

// renderHelpOverlay draws the help dialog sliding over the editor using a
// lipgloss canvas, so the dialog keeps its own layer instead of being spliced
// into the editor's text.
func (m Model) renderHelpOverlay(editorContent string) string {
	if m.Editor == nil || m.Width < 30 || m.Height <= 0 {
		return editorContent
	}
	dialogLines := m.helpDialogLines()
	if len(dialogLines) == 0 {
		return editorContent
	}

	viewport := m.Height - m.reservedRows()
	topOffset := int((1.0 - easeOutQuart(m.HelpAnimOffset)) * float64(len(dialogLines)))
	if topOffset >= len(dialogLines) {
		return editorContent
	}
	visible := dialogLines[topOffset:]
	if len(visible) > viewport {
		visible = visible[:viewport]
	}

	dialog := strings.Join(visible, "\n")
	editorLines := strings.Split(editorContent, "\n")
	width := max(m.Width, lipgloss.Width(dialog))
	height := max(viewport, len(editorLines))
	canvas := lipgloss.NewCanvas(width, height)
	canvas.Compose(lipgloss.NewLayer(editorContent).X(0).Y(0))
	canvas.Compose(lipgloss.NewLayer(dialog).X(0).Y(0))
	return canvas.Render()
}

func (m Model) renderHelpDialogContent(key, desc, section, info lipgloss.Style) string {
	padL, padR, borderW := 2, 1, 2
	contentWidth := m.Width - borderW - padL - padR
	if contentWidth <= 0 {
		return ""
	}

	maxKeyW := 0
	for _, g := range helpGroups {
		for _, item := range g.items {
			maxKeyW = max(maxKeyW, lipgloss.Width(key.Render(item.key)))
		}
	}

	pair := func(k, d string) string {
		kRendered := key.Render(k)
		pad := maxKeyW - lipgloss.Width(kRendered)
		return kRendered + strings.Repeat(" ", max(1, pad+1)) + desc.Render(strings.ToLower(d))
	}

	var contentLines []string
	switch {
	case contentWidth >= 83:
		contentLines = m.helpColumns(4, contentWidth, pair, section)
	case contentWidth >= 55:
		contentLines = m.helpColumns(2, contentWidth, pair, section)
	default:
		contentLines = m.helpStacked(padL, padR, pair, section)
	}

	footerText := " esc or f1 to close "
	dashW := max(contentWidth-lipgloss.Width(footerText), 0)
	footer := info.Render(strings.Repeat("─", dashW/2) + footerText + strings.Repeat("─", dashW-dashW/2))
	contentLines = append(contentLines, strings.Repeat(" ", padL)+footer+strings.Repeat(" ", padR))

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), false, true, true, true).
		BorderForeground(m.Config.ModalBorderStyle.GetBorderBottomForeground()).
		Render(strings.Join(contentLines, "\n"))
}

func (m Model) helpStacked(padL, padR int, pair func(string, string) string, section lipgloss.Style) []string {
	var lines []string
	for _, g := range helpGroups {
		tag := " " + g.title + " "
		lines = append(lines, strings.Repeat(" ", padL)+section.Render("─"+tag+"─")+strings.Repeat(" ", padR))
		for _, item := range g.items {
			lines = append(lines, strings.Repeat(" ", padL)+pair(item.key, item.desc)+strings.Repeat(" ", padR))
		}
		lines = append(lines, strings.Repeat(" ", padL+padR))
	}
	return lines
}

func (m Model) helpColumns(cols, contentWidth int, pair func(string, string) string, section lipgloss.Style) []string {
	var lines []string
	for _, indices := range chunkGroups(len(helpGroups), cols) {
		colW := (contentWidth - (len(indices) - 1)) / len(indices)

		var header strings.Builder
		for i, gi := range indices {
			if i > 0 {
				header.WriteString(" ")
			}
			tag := " " + helpGroups[gi].title + " "
			h := section.Render("─" + tag + "─")
			header.WriteString(h)
			if p := colW - lipgloss.Width(h); p > 0 {
				header.WriteString(strings.Repeat(" ", p))
			}
		}
		lines = append(lines, "  "+header.String()+" ")

		maxItems := 0
		for _, gi := range indices {
			maxItems = max(maxItems, len(helpGroups[gi].items))
		}
		for row := 0; row < maxItems; row++ {
			var line strings.Builder
			for i, gi := range indices {
				if i > 0 {
					line.WriteString(" ")
				}
				if row < len(helpGroups[gi].items) {
					cell := pair(helpGroups[gi].items[row].key, helpGroups[gi].items[row].desc)
					line.WriteString(cell)
					if p := colW - lipgloss.Width(cell); p > 0 {
						line.WriteString(strings.Repeat(" ", p))
					}
				} else {
					line.WriteString(strings.Repeat(" ", colW))
				}
			}
			lines = append(lines, "  "+line.String()+" ")
		}
		lines = append(lines, "   ")
	}
	return lines
}

func chunkGroups(n, cols int) [][]int {
	rows := (n + cols - 1) / cols
	out := make([][]int, 0, rows)
	for r := range rows {
		start := r * cols
		end := min(start+cols, n)
		chunk := make([]int, end-start)
		for i := start; i < end; i++ {
			chunk[i-start] = i
		}
		out = append(out, chunk)
	}
	return out
}
