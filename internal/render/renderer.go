package render

import (
	"strings"

	"github.com/z-d-g/md-cli/v2/internal/config"
	"github.com/z-d-g/md-cli/v2/internal/markdown"

	"charm.land/lipgloss/v2"
)

// lipglossRenderer implements LineRenderer using lipgloss for rendering.
type lipglossRenderer struct {
	styles        *config.EditorStyles
	tableDocCache *tableDoc
	tableDocRaw   []string
	tableDocLines int
	tableVersion  int
	documentLines func() []string
	terminalWidth int
}

// NewLipglossRenderer creates a LineRenderer backed by lipgloss.
func NewLipglossRenderer(styles *config.EditorStyles, width int) LineRenderer {
	return &lipglossRenderer{styles: styles, terminalWidth: width}
}

func (r *lipglossRenderer) SetDocument(lines func() []string) {
	r.documentLines = lines
}

func (r *lipglossRenderer) SetTerminalWidth(w int) {
	r.terminalWidth = w
}

func (r *lipglossRenderer) TerminalWidth() int {
	return r.terminalWidth
}

func (r *lipglossRenderer) RenderLineNumber(text string) string {
	return r.styles.LineNumber.Render(text)
}

func (r *lipglossRenderer) RenderCursorChar(ch string) string {
	return r.styles.Cursor.Render(ch)
}

func (r *lipglossRenderer) RenderSelectionChar(ch string) string {
	return r.styles.Selection.Render(ch)
}

func (r *lipglossRenderer) RenderStyled(text string, lineType int) string {
	switch lineType {
	case LineHeading1:
		return r.styles.H1.Render(text)
	case LineHeading2:
		return r.styles.H2.Render(text)
	case LineHeading3:
		return r.styles.H3.Render(text)
	case LineHeading4:
		return r.styles.H4.Render(text)
	case LineHeading5:
		return r.styles.H5.Render(text)
	case LineHeading6:
		return r.styles.H6.Render(text)
	case LineCodeFence:
		return r.styles.CodeFence.Render(text)
	case LineCodeContent:
		return r.styles.CodeContent.Render(text)
	case LineBlockQuote:
		return r.styles.BlockQuote.Render(text)
	default:
		return text
	}
}

// RenderLink renders text as a clickable hyperlink.
// Applies link styling (foreground, underline, hyperlink) in a single render pass
// to avoid nested ANSI escape corruption.
func (r *lipglossRenderer) RenderLink(text, url string) string {
	linkStyle := r.styles.Link
	s := lipgloss.NewStyle().
		Foreground(linkStyle.GetForeground()).
		Underline(true).
		UnderlineStyle(lipgloss.UnderlineCurly).
		Hyperlink(url)
	return s.Render(text)
}

func (r *lipglossRenderer) TableVersion() int {
	return r.tableVersion
}

// ResetTable invalidates the cached table index. Table rendering itself is
// stateless; this exists so document mutations refresh derived widths.
func (r *lipglossRenderer) ResetTable() {
	r.tableVersion++
	r.tableDocCache = nil
	r.tableDocRaw = nil
	r.tableDocLines = 0
}

// RenderLine renders one document line. idx is its zero-based document row;
// pass -1 when rendering a standalone line with no document context.
func (r *lipglossRenderer) RenderLine(line string, idx int, isInCodeBlock bool) string {
	if markdown.IsCodeFence(line) {
		w := max(r.terminalWidth-4, 10)
		return r.styles.CodeFence.Render(strings.Repeat("┄", w))
	}

	if isInCodeBlock {
		return r.styles.CodeContent.Render(line)
	}

	if markdown.IsBlockquoteLine(line) {
		depth := markdown.CountBlockquoteDepth(line)
		content := strings.TrimSpace(line)
		for range depth {
			content = strings.TrimPrefix(content, ">")
			content = strings.TrimSpace(content)
		}
		elements := ParseInlineElements(content)
		styledContent := r.RenderInline(elements, lipgloss.Style{})
		prefix := strings.Repeat("│ ", depth)
		return r.styles.BlockQuote.Render(prefix) + styledContent
	}

	if markdown.IsHorizontalRule(line) {
		w := max(r.terminalWidth-4, 10)
		return r.styles.HorizontalRule.Render(strings.Repeat("─", w))
	}

	if markdown.IsHeadingLine(line) {
		level := markdown.CountLeadingHashes(line)
		trimmed := strings.TrimSpace(line)
		startIndex := level
		if startIndex < len(trimmed) && trimmed[startIndex] == ' ' {
			startIndex++
		}
		content := strings.TrimSpace(trimmed[startIndex:])

		switch level {
		case 1:
			return r.styles.H1.Render(content)
		case 2:
			return r.styles.H2.Render(content)
		case 3:
			return r.styles.H3.Render(content)
		case 4:
			return r.styles.H4.Render(content)
		case 5:
			return r.styles.H5.Render(content)
		case 6:
			return r.styles.H6.Render(content)
		default:
			return r.styles.H1.Render(content)
		}
	}

	if markdown.IsTableLine(line) {
		return renderTable(line, idx, r)
	}

	if markdown.IsListLine(line) {
		trimmed := strings.TrimSpace(line)

		if isCheckbox(trimmed) {
			return renderCheckbox(trimmed, r)
		}

		if isNumberedList(trimmed) {
			return renderNumberedList(trimmed, r)
		}

		return renderBulletList(trimmed, r)
	}

	elements := ParseInlineElements(line)
	return r.RenderInline(elements, lipgloss.Style{})
}
