package editor

import (
	"unicode/utf8"

	"github.com/z-d-g/md-cli/v2/internal/markdown"
)

// Syntax types delegate to the shared markdown package.
type SpanType = markdown.SpanType
type SyntaxSpan = markdown.SyntaxSpan

// Span type constants — re-exported from markdown package.
const (
	SpanBold             = markdown.SpanBold
	SpanItalic           = markdown.SpanItalic
	SpanBoldItalic       = markdown.SpanBoldItalic
	SpanUnderline        = markdown.SpanUnderline
	SpanCode             = markdown.SpanCode
	SpanLink             = markdown.SpanLink
	SpanImage            = markdown.SpanImage
	SpanStrikethrough    = markdown.SpanStrikethrough
	SpanHeadingMarker    = markdown.SpanHeadingMarker
	SpanListMarker       = markdown.SpanListMarker
	SpanBlockquoteMarker = markdown.SpanBlockquoteMarker
	SpanHR               = markdown.SpanHR
)

// FindSyntaxSpans delegates to the shared markdown package implementation.
func FindSyntaxSpans(line string) []SyntaxSpan {
	return markdown.FindSyntaxSpans(line)
}

// FindClosingDelimiter delegates to the shared markdown package implementation.

func IsCursorOnSyntax(line string, col int) (bool, SyntaxSpan) {
	spans := FindSyntaxSpans(line)
	byteOffset := runeToByteOffset(line, col)

	for _, span := range spans {
		if byteOffset >= span.Start && byteOffset < span.End {
			return true, span
		}
	}

	return false, SyntaxSpan{}
}

func runeToByteOffset(s string, runeIndex int) int {
	if runeIndex <= 0 {
		return 0
	}

	byteOffset := 0
	runeIdx := 0
	for byteOffset < len(s) && runeIdx < runeIndex {
		_, size := utf8.DecodeRuneInString(s[byteOffset:])
		byteOffset += size
		runeIdx++
	}
	return byteOffset
}

// FindBlockRegion reports the inclusive line range shown as raw source when the
// cursor sits on row. The doc is the parsed document snapshot for this frame.
func FindBlockRegion(doc *markdown.Doc, row int) (int, int, bool) {
	if doc == nil {
		return row, row, false
	}
	if row < 0 || row >= doc.Len() {
		return row, row, false
	}
	kind := doc.LineKind(row)
	if kind == markdown.LineCodeFence || doc.InCode(row) {
		start, end := doc.Bounds(row)
		return start, end, true
	}
	if kind == markdown.LineBlockQuote || kind >= markdown.LineHeading1 && kind <= markdown.LineHeading6 {
		start, end := doc.Bounds(row)
		return start, end, true
	}
	if start, end := doc.Bounds(row); start != row || end != row {
		return start, end, true
	}
	if hasInlineSyntax(doc.Lines()[row]) {
		return row, row, true
	}
	return row, row, false
}

func hasInlineSyntax(line string) bool {
	spans := FindSyntaxSpans(line)
	for _, span := range spans {
		if span.SpanType == SpanBold || span.SpanType == SpanItalic ||
			span.SpanType == SpanBoldItalic || span.SpanType == SpanUnderline ||
			span.SpanType == SpanCode || span.SpanType == SpanLink ||
			span.SpanType == SpanImage || span.SpanType == SpanStrikethrough {
			return true
		}
	}
	return false
}
