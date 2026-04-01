package render

import "github.com/z-d-g/md-cli/internal/markdown"

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

type StyleFunc func(text string) string

// Compose chains two StyleFuncs: other(s(text)).
func (s StyleFunc) Compose(other StyleFunc) StyleFunc {
	if s == nil {
		return other
	}
	if other == nil {
		return s
	}
	return func(text string) string {
		return other(s(text))
	}
}

// LineRenderer handles rendering of markdown lines to styled terminal output.
type LineRenderer interface {
	RenderLine(line string, isInCodeBlock bool) string
	RenderStyled(text string, lineType int) string
	RenderInline(elements []InlineElement, base StyleFunc) string
	RenderSourceInline(elements []InlineElement, base StyleFunc) string
	RenderLineNumber(text string) string
	RenderCursorChar(ch string) string
	RenderSelectionChar(ch string) string
	RenderLink(text, url string) string
	TableVersion() int
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
