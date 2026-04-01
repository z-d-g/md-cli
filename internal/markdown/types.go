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
}

type SpanType int

const (
	SpanBold SpanType = iota
	SpanItalic
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
