package markdown

import "strings"

// CodeFenceChar returns the fence character of line, or 0 when it is not a fence.
func CodeFenceChar(line string) byte {
	if f := ParseFence(line); f.Valid {
		return f.Char
	}
	return 0
}

// IsBlockquoteLine reports whether line starts a blockquote.
func IsBlockquoteLine(line string) bool {
	return strings.HasPrefix(strings.TrimSpace(line), ">")
}

// IsCodeFence reports whether line is a fenced code block delimiter.
func IsCodeFence(line string) bool { return ParseFence(line).Valid }

// ParseFence describes the fenced-code delimiter on line, if any. A fence is at
// least three identical ` or ~ characters. Backtick fences accept no backtick
// in the info string.
func ParseFence(line string) Fence {
	trimmed := strings.TrimLeft(line, " \t")
	if len(trimmed) < 3 {
		return Fence{}
	}
	c := trimmed[0]
	if c != '`' && c != '~' {
		return Fence{}
	}
	n := 0
	for n < len(trimmed) && trimmed[n] == c {
		n++
	}
	if n < 3 {
		return Fence{}
	}
	info := strings.TrimSpace(trimmed[n:])
	if c == '`' && strings.ContainsRune(info, '`') {
		return Fence{}
	}
	return Fence{Char: c, Len: n, Info: info, Valid: true}
}

// IsHorizontalRule reports whether line is a thematic break.
func IsHorizontalRule(line string) bool {
	trimmed := strings.TrimSpace(line)
	if len(trimmed) < 3 {
		return false
	}
	c := trimmed[0]
	if c != '-' && c != '*' && c != '_' {
		return false
	}
	count := 0
	for i := 0; i < len(trimmed); i++ {
		switch {
		case trimmed[i] == c:
			count++
		case trimmed[i] == ' ' || trimmed[i] == '\t':
		default:
			return false
		}
	}
	return count >= 3
}

// IsListLine reports whether line starts a list item.
func IsListLine(line string) bool { return listMarkerStart(line) >= 0 }

// IsHeadingLine reports whether line is an ATX heading.
func IsHeadingLine(line string) bool {
	n := CountLeadingHashes(line)
	return n > 0 && n < len(strings.TrimSpace(line)) && strings.TrimSpace(line)[n] == ' '
}

// IsTableLine reports whether line is part of a table. It requires at least two
// pipes, with a leading or trailing pipe, so shell pipelines stay text.
func IsTableLine(line string) bool {
	if strings.Count(line, "|") < 2 {
		return false
	}
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, "|") || strings.HasSuffix(trimmed, "|")
}

// IsTableSeparatorLine reports whether line is a table delimiter row.
func IsTableSeparatorLine(line string) bool { return isSeparatorLine(line) }

// CountLeadingHashes returns the number of leading # characters (up to 6).
func CountLeadingHashes(line string) int {
	trimmed := strings.TrimSpace(line)
	n := 0
	for n < len(trimmed) && n < 6 && trimmed[n] == '#' {
		n++
	}
	return n
}

// CountBlockquoteDepth returns the nesting depth of blockquote markers.
func CountBlockquoteDepth(line string) int {
	depth, _ := blockquotePrefix(strings.TrimLeft(line, " \t"))
	return depth
}

// ClassifyLine returns the styling class for line, given code-block membership.
func ClassifyLine(line string, isInCodeBlock bool) int {
	if isInCodeBlock {
		return LineCodeContent
	}
	if IsCodeFence(line) {
		return LineCodeFence
	}
	if n := CountLeadingHashes(line); n > 0 && IsHeadingLine(line) {
		return n
	}
	if strings.HasPrefix(strings.TrimSpace(line), ">") {
		return LineBlockQuote
	}
	return LineNormal
}
