package markdown

type InlineType int

const (
	InlineText InlineType = iota
	InlineBold
	InlineItalic
	InlineBoldItalic
	InlineCode
	InlineLink
	InlineImage
	InlineStrikethrough
	InlineUnderline
)

type InlineElement struct {
	Type      InlineType
	Content   string
	URL       string
	Delimiter string
	Children  []InlineElement

	// Start is the byte offset of this element in the source line, or -1 when
	// the element does not map to a source range.
	Start int
}

type SpanType int

const (
	SpanBold SpanType = iota
	SpanItalic
	SpanBoldItalic
	SpanUnderline
	SpanCode
	SpanLink
	SpanImage
	SpanStrikethrough
	SpanHeadingMarker
	SpanListMarker
	SpanBlockquoteMarker
	SpanHR
)

type SyntaxSpan struct {
	Start    int
	End      int
	SpanType SpanType
}

const (
	LineNormal = iota
	LineHeading1
	LineHeading2
	LineHeading3
	LineHeading4
	LineHeading5
	LineHeading6
	LineCodeFence
	LineCodeContent
	LineBlockQuote
)
