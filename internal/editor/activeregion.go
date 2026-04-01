package editor

import (
	"strings"
	"unicode/utf8"

	"github.com/z-d-g/md-cli/internal/markdown"
)

// Syntax types delegate to the shared markdown package.
type SpanType = markdown.SpanType
type SyntaxSpan = markdown.SyntaxSpan

// Span type constants — re-exported from markdown package.
const (
	SpanBold             = markdown.SpanBold
	SpanItalic           = markdown.SpanItalic
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
func FindClosingDelimiter(line string, start int, delimiter string) int {
	return markdown.FindClosingDelimiter(line, start, delimiter)
}

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

// FindBlockRegion determines the active region around the cursor for source-mode rendering.
// This is editor-specific because it depends on GapBuffer.
func FindBlockRegion(buf *GapBuffer, cursorRow int) (int, int, bool) {
	if cursorRow < 0 || cursorRow >= buf.LineCount() {
		return cursorRow, cursorRow, false
	}

	currentLine := buf.LineAt(cursorRow)

	if isInCodeBlock(buf, cursorRow) {
		start, end := findCodeBlockBounds(buf, cursorRow)
		return start, end, true
	}

	if strings.HasPrefix(strings.TrimSpace(currentLine), ">") {
		start, end := findBlockquoteBounds(buf, cursorRow)
		return start, end, true
	}

	if markdown.IsTableLine(currentLine) {
		start, end := findTableBounds(buf, cursorRow)
		return start, end, true
	}

	if markdown.IsListLine(currentLine) {
		start, end := findListBounds(buf, cursorRow)
		return start, end, true
	}

	if markdown.IsHeadingLine(currentLine) {
		return cursorRow, cursorRow, true
	}

	if hasInlineSyntax(currentLine) {
		return cursorRow, cursorRow, true
	}

	return cursorRow, cursorRow, false
}

func isInCodeBlock(buf *GapBuffer, row int) bool {
	trimmed := strings.TrimSpace(buf.LineAt(row))

	insideCodeBlock := false
	for i := 0; i <= row && i < buf.LineCount(); i++ {
		if markdown.IsCodeFence(buf.LineAt(i)) {
			insideCodeBlock = !insideCodeBlock
		}
	}
	return insideCodeBlock || markdown.IsCodeFence(trimmed)
}

func findCodeBlockBounds(buf *GapBuffer, row int) (int, int) {
	fenceCount := 0

	for i := 0; i <= row && i < buf.LineCount(); i++ {
		if markdown.IsCodeFence(buf.LineAt(i)) {
			fenceCount++
			if i == row {
				if fenceCount%2 == 1 {
					return findMatchingCodeFence(buf, row, true)
				} else {
					return findMatchingCodeFence(buf, row, false)
				}
			}
		}
	}

	start := row
	for start >= 0 && !markdown.IsCodeFence(buf.LineAt(start)) {
		start--
	}
	if start < 0 || !markdown.IsCodeFence(buf.LineAt(start)) {
		return row, row
	}

	end := row
	for end < buf.LineCount() && !markdown.IsCodeFence(buf.LineAt(end)) {
		end++
	}
	if end >= buf.LineCount() || !markdown.IsCodeFence(buf.LineAt(end)) {
		return start, buf.LineCount() - 1
	}

	return start, end
}

func findMatchingCodeFence(buf *GapBuffer, fenceRow int, isOpening bool) (int, int) {
	if isOpening {
		start := fenceRow
		end := fenceRow + 1
		for end < buf.LineCount() && !markdown.IsCodeFence(buf.LineAt(end)) {
			end++
		}
		if end >= buf.LineCount() {
			end = buf.LineCount() - 1
		}
		return start, end
	}

	start := fenceRow - 1
	for start >= 0 && !markdown.IsCodeFence(buf.LineAt(start)) {
		start--
	}
	if start < 0 {
		start = 0
	}
	return start, fenceRow
}

func findBlockquoteBounds(buf *GapBuffer, row int) (int, int) {
	start, end := row, row

	for start > 0 && strings.HasPrefix(strings.TrimSpace(buf.LineAt(start-1)), ">") {
		start--
	}
	for end < buf.LineCount()-1 && strings.HasPrefix(strings.TrimSpace(buf.LineAt(end+1)), ">") {
		end++
	}

	return start, end
}

func findTableBounds(buf *GapBuffer, row int) (int, int) {
	start, end := row, row

	for start > 0 && markdown.IsTableLine(buf.LineAt(start-1)) {
		start--
	}
	for end < buf.LineCount()-1 && markdown.IsTableLine(buf.LineAt(end+1)) {
		end++
	}

	return start, end
}

func hasInlineSyntax(line string) bool {
	spans := FindSyntaxSpans(line)
	for _, span := range spans {
		if span.SpanType == SpanBold || span.SpanType == SpanItalic ||
			span.SpanType == SpanCode || span.SpanType == SpanLink ||
			span.SpanType == SpanImage || span.SpanType == SpanStrikethrough {
			return true
		}
	}
	return false
}

func findListBounds(buf *GapBuffer, row int) (int, int) {
	start, end := row, row

	for start > 0 && (markdown.IsListLine(buf.LineAt(start-1)) || markdown.IsEmptyLine(buf.LineAt(start-1))) {
		start--
		if !markdown.IsListLine(buf.LineAt(start)) && !markdown.IsEmptyLine(buf.LineAt(start)) {
			start++
			break
		}
	}

	for end < buf.LineCount()-1 && (markdown.IsListLine(buf.LineAt(end+1)) || markdown.IsEmptyLine(buf.LineAt(end+1))) {
		end++
		if !markdown.IsListLine(buf.LineAt(end)) && !markdown.IsEmptyLine(buf.LineAt(end)) {
			end--
			break
		}
	}

	return start, end
}
