package render

import (
	"strings"

	"github.com/z-d-g/md-cli/internal/config"
	"github.com/z-d-g/md-cli/internal/markdown"

	"charm.land/lipgloss/v2"
)

// styleCache holds pre-built StyleFuncs for inline rendering performance.
type styleCache struct {
	boldFunc          StyleFunc
	italicFunc        StyleFunc
	underlineFunc     StyleFunc
	strikethroughFunc StyleFunc
	codeSpanFunc      StyleFunc
	codeContentFunc   StyleFunc
	linkFunc          StyleFunc
	linkURLFunc       StyleFunc
	imageFunc         StyleFunc
	blockquoteFunc    StyleFunc
	bulletFunc        StyleFunc
	hrFunc            StyleFunc
	codeFenceFunc     StyleFunc
	h1Func            StyleFunc
	h2Func            StyleFunc
	h3Func            StyleFunc
	h4Func            StyleFunc
	h5Func            StyleFunc
	h6Func            StyleFunc
	tableBorderFunc   StyleFunc
	tableHeaderFunc   StyleFunc
	tableCellFunc     StyleFunc
	selectionFunc     StyleFunc
	cursorFunc        StyleFunc
}

// lipglossRenderer implements LineRenderer using lipgloss for rendering.
type lipglossRenderer struct {
	styles         *config.EditorStyles
	styleCache     styleCache
	lineNumberFunc StyleFunc
	table          tableContext
	tableVersion   int
	documentLines  func() []string
	terminalWidth  int
}

// NewLipglossRenderer creates a LineRenderer backed by lipgloss.
func NewLipglossRenderer(styles *config.EditorStyles, width int) LineRenderer {
	r := &lipglossRenderer{
		styles:        styles,
		terminalWidth: width,
	}

	r.styleCache = styleCache{
		boldFunc:          func(t string) string { return styles.Bold.Render(t) },
		italicFunc:        func(t string) string { return styles.Italic.Render(t) },
		underlineFunc:     func(t string) string { return styles.Underline.Render(t) },
		strikethroughFunc: func(t string) string { return styles.Strikethrough.Render(t) },
		codeSpanFunc:      func(t string) string { return styles.CodeSpan.Render(t) },
		codeContentFunc:   func(t string) string { return styles.CodeContent.Render(t) },
		linkFunc:          func(t string) string { return styles.Link.Render(t) },
		linkURLFunc:       func(t string) string { return styles.LinkURL.Render(t) },
		imageFunc:         func(t string) string { return styles.Image.Render(t) },
		blockquoteFunc:    func(t string) string { return styles.BlockQuote.Render(t) },
		bulletFunc:        func(t string) string { return styles.Bullet.Render(t) },
		hrFunc:            func(t string) string { return styles.HorizontalRule.Render(t) },
		codeFenceFunc:     func(t string) string { return styles.CodeFence.Render(t) },
		h1Func:            func(t string) string { return styles.H1.Render(t) },
		h2Func:            func(t string) string { return styles.H2.Render(t) },
		h3Func:            func(t string) string { return styles.H3.Render(t) },
		h4Func:            func(t string) string { return styles.H4.Render(t) },
		h5Func:            func(t string) string { return styles.H5.Render(t) },
		h6Func:            func(t string) string { return styles.H6.Render(t) },
		tableBorderFunc:   func(t string) string { return styles.TableBorder.Render(t) },
		tableHeaderFunc:   func(t string) string { return styles.TableHeader.Render(t) },
		tableCellFunc:     func(t string) string { return styles.TableCell.Render(t) },
		selectionFunc:     func(t string) string { return styles.Selection.Render(t) },
		cursorFunc:        func(t string) string { return styles.Cursor.Render(t) },
	}

	r.lineNumberFunc = func(t string) string { return styles.LineNumber.Render(t) }

	return r
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
	return r.lineNumberFunc(text)
}

func (r *lipglossRenderer) RenderCursorChar(ch string) string {
	return r.styleCache.cursorFunc(ch)
}

func (r *lipglossRenderer) RenderSelectionChar(ch string) string {
	return r.styleCache.selectionFunc(ch)
}

func (r *lipglossRenderer) RenderStyled(text string, lineType int) string {
	switch lineType {
	case LineHeading1:
		return r.styleCache.h1Func(text)
	case LineHeading2:
		return r.styleCache.h2Func(text)
	case LineHeading3:
		return r.styleCache.h3Func(text)
	case LineHeading4:
		return r.styleCache.h4Func(text)
	case LineHeading5:
		return r.styleCache.h5Func(text)
	case LineHeading6:
		return r.styleCache.h6Func(text)
	case LineCodeFence:
		return r.styleCache.codeFenceFunc(text)
	case LineCodeContent:
		return r.styleCache.codeContentFunc(text)
	case LineBlockQuote:
		return r.styleCache.blockquoteFunc(text)
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

func (r *lipglossRenderer) RenderLine(line string, isInCodeBlock bool) string {
	if markdown.IsCodeFence(line) {
		w := max(r.terminalWidth-4, 10)
		return r.styleCache.codeFenceFunc(strings.Repeat("┄", w))
	}

	if isInCodeBlock {
		return r.styleCache.codeContentFunc(line)
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
		return r.styleCache.blockquoteFunc(prefix) + styledContent
	}

	if markdown.IsHorizontalRule(line) {
		w := max(r.terminalWidth-4, 10)
		return r.styleCache.hrFunc(strings.Repeat("─", w))
	}

	if markdown.IsHeadingLine(line) {
		level := markdown.CountLeadingHashes(line)
		startIndex := level
		if startIndex < len(line) && line[startIndex] == ' ' {
			startIndex++
		}
		content := strings.TrimSpace(line[startIndex:])

		switch level {
		case 1:
			return r.styleCache.h1Func(content)
		case 2:
			return r.styleCache.h2Func(content)
		case 3:
			return r.styleCache.h3Func(content)
		case 4:
			return r.styleCache.h4Func(content)
		case 5:
			return r.styleCache.h5Func(content)
		case 6:
			return r.styleCache.h6Func(content)
		default:
			return r.styleCache.h1Func(content)
		}
	}

	if markdown.IsTableLine(line) {
		return renderTable(line, r)
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

	// Non-table line: reset table state if was active
	if r.table.active {
		r.ResetTable()
	}

	elements := ParseInlineElements(line)
	return r.RenderInline(elements, lipgloss.Style{})
}
