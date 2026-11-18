package render

import (
	"github.com/z-d-g/md-cli/v2/internal/markdown"

	"charm.land/lipgloss/v2"
)

type InlineType = markdown.InlineType
type InlineElement = markdown.InlineElement

const (
	InlineText          = markdown.InlineText
	InlineBold          = markdown.InlineBold
	InlineItalic        = markdown.InlineItalic
	InlineBoldItalic    = markdown.InlineBoldItalic
	InlineCode          = markdown.InlineCode
	InlineLink          = markdown.InlineLink
	InlineImage         = markdown.InlineImage
	InlineStrikethrough = markdown.InlineStrikethrough
	InlineUnderline     = markdown.InlineUnderline
)

// LineRenderer handles rendering of markdown lines to styled terminal output.
type LineRenderer interface {
	// RenderLine renders line at document row idx (-1 when context-free).
	RenderLine(line string, idx int, isInCodeBlock bool) string
	RenderStyled(text string, lineType int) string
	RenderInline(elements []InlineElement, base lipgloss.Style) string
	RenderSourceInline(elements []InlineElement, base lipgloss.Style) string
	RenderLineNumber(text string) string
	RenderCursorChar(ch string) string
	RenderSelectionChar(ch string) string
	RenderLink(text, url string) string
	SetDocument(lines func() []string)
	SetTerminalWidth(w int)
	TerminalWidth() int
	TableVersion() int
	ResetTable()
}

const (
	LineNormal      = markdown.LineNormal
	LineHeading1    = markdown.LineHeading1
	LineHeading2    = markdown.LineHeading2
	LineHeading3    = markdown.LineHeading3
	LineHeading4    = markdown.LineHeading4
	LineHeading5    = markdown.LineHeading5
	LineHeading6    = markdown.LineHeading6
	LineCodeFence   = markdown.LineCodeFence
	LineCodeContent = markdown.LineCodeContent
	LineBlockQuote  = markdown.LineBlockQuote
)
