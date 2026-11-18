package markdown

import (
	"strings"

	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/extension"
	extast "github.com/yuin/goldmark/v2/extension/ast"
	gparser "github.com/yuin/goldmark/v2/parser"
)

// md is the shared goldmark parser with the extensions md-cli renders.
// It is stateless after construction, so one instance serves every line.
var md = newParser()

func newParser() gparser.Parser {
	return gparser.New(gparser.WithExtensions(
		extension.NewTableParser(),
		extension.NewStrikethroughParser(),
		extension.NewTaskListItemParser(),
		extension.NewLinkifyParser(),
		NewUnderlineParser(),
	))
}

// ParseInlineElements parses a markdown line into inline elements.
func ParseInlineElements(line string) []InlineElement {
	return parseNodes(md.Parse([]byte(line)), line)
}

// parseNodes converts the inline content of root. Markdown wraps inline nodes
// in a block (Paragraph, Heading, ...), so the block itself is descended but
// the inline children become the element list.
func parseNodes(root ast.Node, line string) []InlineElement {
	src := []byte(line)
	for c := root.FirstChild(); c != nil; c = c.NextSibling() {
		if c.FirstChild() == nil {
			continue
		}
		return children(c, src, line)
	}
	return nil
}

// convert maps one goldmark node onto an InlineElement.
func convert(n ast.Node, src []byte, line string) InlineElement {
	switch v := n.(type) {
	case *ast.Text:
		return InlineElement{Type: InlineText, Content: v.Value.Value(src), Start: v.Value.Index().Start}
	case *ast.CodeSpan:
		open := codeOpenStart(line, v.Pos())
		return InlineElement{
			Type:      InlineCode,
			Content:   v.Value.Value(src),
			Delimiter: codeDelimiter(line, v.Pos()),
			Start:     open,
		}
	case *ast.Emphasis:
		typ := InlineItalic
		if run := runAt(line, v.Pos()); len(run) >= 3 {
			typ = InlineBoldItalic
		}
		return InlineElement{Type: typ, Delimiter: runAt(line, v.Pos()), Content: plainText(v, src), Children: children(v, src, line), Start: v.Pos()}
	case *ast.Strong:
		return InlineElement{Type: InlineBold, Delimiter: runAt(line, v.Pos()), Content: plainText(v, src), Children: children(v, src, line), Start: v.Pos()}
	case *extast.Strikethrough:
		return InlineElement{Type: InlineStrikethrough, Delimiter: "~~", Content: plainText(v, src), Children: children(v, src, line), Start: v.Pos()}
	case *Underline:
		return InlineElement{Type: InlineUnderline, Delimiter: runAt(line, v.Pos()), Content: plainText(v, src), Children: children(v, src, line), Start: v.Pos()}
	case *ast.Link:
		return InlineElement{Type: InlineLink, Content: plainText(v, src), URL: v.Destination.Value(src), Start: v.Pos()}
	case *ast.Image:
		return InlineElement{Type: InlineImage, Content: plainText(v, src), URL: v.Destination.Value(src), Start: v.Pos()}
	case *ast.AutoLink:
		return InlineElement{Type: InlineLink, Content: v.Label.Value(src), URL: v.Destination.Value(src), Start: v.Pos()}
	}
	return InlineElement{Type: InlineText, Content: plainText(n, src), Start: n.Pos()}
}

func children(n ast.Node, src []byte, line string) []InlineElement {
	var out []InlineElement
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		out = append(out, convert(c, src, line))
	}
	return out
}

// plainText concatenates the visible text of an inline subtree.
func plainText(n ast.Node, src []byte) string {
	if t, ok := n.(*ast.Text); ok {
		return t.Value.Value(src)
	}
	var b strings.Builder
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		b.WriteString(plainText(c, src))
	}
	return b.String()
}

// runAt returns the delimiter run that opens at pos.
func runAt(line string, pos int) string {
	if pos < 0 || pos >= len(line) {
		return ""
	}
	end := pos
	for end < len(line) && line[end] == line[pos] {
		end++
	}
	return line[pos:end]
}

// codeOpenStart returns the byte offset of the backtick run opening a code span
// whose content begins at pos.
func codeOpenStart(line string, pos int) int {
	if pos <= 0 || pos > len(line) {
		return -1
	}
	i := pos - 1
	if line[i] != '`' {
		return -1
	}
	for i > 0 && line[i-1] == '`' {
		i--
	}
	return i
}

// codeDelimiter returns the backtick run for a code span whose content begins
// at pos. CodeSpan content excludes the fences, so the opener is the nearest run
// before the content.
func codeDelimiter(line string, pos int) string {
	start := codeOpenStart(line, pos)
	if start < 0 {
		return "`"
	}
	return runAt(line, start)
}

// FindSyntaxSpans returns the byte ranges of every syntax marker on line.
func FindSyntaxSpans(line string) []SyntaxSpan {
	spans := blockMarkerSpans(line)
	spans = append(spans, inlineSpans(line)...)
	SortSpans(spans)
	return spans
}

// inlineSpans walks the goldmark AST for one line and emits delimiter ranges.
func inlineSpans(line string) []SyntaxSpan {
	src := []byte(line)
	root := md.Parse(src)
	var spans []SyntaxSpan
	var walk func(n ast.Node)
	walk = func(n ast.Node) {
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			switch v := c.(type) {
			case *ast.CodeSpan:
				spans = appendCodeSpan(spans, line, v)
			case *ast.Emphasis:
				spans = appendEmphasis(spans, line, v.Pos(), SpanItalic)
			case *ast.Strong:
				spans = appendEmphasis(spans, line, v.Pos(), SpanBold)
			case *extast.Strikethrough:
				spans = appendEmphasis(spans, line, v.Pos(), SpanStrikethrough)
			case *Underline:
				spans = appendEmphasis(spans, line, v.Pos(), SpanUnderline)
			case *ast.Link:
				spans = appendLink(spans, line, v.Pos(), SpanLink)
			case *ast.Image:
				spans = appendLink(spans, line, v.Pos(), SpanImage)
			}
			walk(c)
		}
	}
	walk(root)
	return spans
}

func appendCodeSpan(spans []SyntaxSpan, line string, v *ast.CodeSpan) []SyntaxSpan {
	contentStart := v.Pos()
	start := codeOpenStart(line, contentStart)
	if start < 0 {
		return spans
	}
	run := SplitRun(line, start, "`")
	if run == 0 {
		return spans
	}
	end := CodeSpanClose(line, contentStart, run)
	if end < 0 {
		return append(spans, SyntaxSpan{Start: start, End: start + run, SpanType: SpanCode})
	}
	return append(spans,
		SyntaxSpan{Start: start, End: start + run, SpanType: SpanCode},
		SyntaxSpan{Start: end, End: end + run, SpanType: SpanCode})
}

// appendEmphasis emits the opening and closing delimiter ranges. goldmark does
// not expose emphasis delimiters, so the run length is read from the source and
// the closer is the matching run at the end of the node.
func appendEmphasis(spans []SyntaxSpan, line string, pos int, st SpanType) []SyntaxSpan {
	run := SplitRun(line, pos, "*_~+")
	if run == 0 {
		return spans
	}
	switch {
	case run == 3 && st == SpanItalic:
		st = SpanBoldItalic
	case run == 3 && st == SpanBold:
		st = SpanBoldItalic
	}
	if close := FindClosingDelimiter(line, pos+run, line[pos:pos+run]); close >= 0 {
		return append(spans,
			SyntaxSpan{Start: pos, End: pos + run, SpanType: st},
			SyntaxSpan{Start: close, End: close + run, SpanType: st})
	}
	return append(spans, SyntaxSpan{Start: pos, End: pos + run, SpanType: st})
}

// appendLink emits the [ ] ( ) ranges for a link or image.
func appendLink(spans []SyntaxSpan, line string, pos int, st SpanType) []SyntaxSpan {
	rest := line[pos:]
	open := strings.IndexByte(rest, '[')
	if open < 0 {
		return spans
	}
	start := pos + open
	closeRel := strings.IndexByte(rest[open:], ']')
	if closeRel < 0 {
		return spans
	}
	closeIdx := start + closeRel
	if closeIdx+1 >= len(line) || line[closeIdx+1] != '(' {
		return spans
	}
	urlEnd := strings.IndexByte(line[closeIdx+2:], ')')
	if urlEnd < 0 {
		return spans
	}
	urlEnd += closeIdx + 2

	spans = append(spans,
		SyntaxSpan{Start: start, End: start + 1, SpanType: st},
		SyntaxSpan{Start: closeIdx, End: closeIdx + 1, SpanType: st},
		SyntaxSpan{Start: closeIdx + 1, End: closeIdx + 2, SpanType: st},
		SyntaxSpan{Start: urlEnd, End: urlEnd + 1, SpanType: st})
	if pos < start {
		spans = append(spans, SyntaxSpan{Start: pos, End: start + 1, SpanType: st})
	}
	return spans
}

func blockMarkerSpans(line string) []SyntaxSpan {
	trimmed := strings.TrimLeft(line, " \t")
	indent := len(line) - len(trimmed)
	var spans []SyntaxSpan

	if n := leadingRun(trimmed, '#'); n > 0 && n < len(trimmed) && trimmed[n] == ' ' {
		spans = append(spans, SyntaxSpan{Start: indent, End: indent + n, SpanType: SpanHeadingMarker})
	}
	if depth, width := blockquotePrefix(trimmed); depth > 0 {
		spans = append(spans, SyntaxSpan{Start: indent, End: indent + width, SpanType: SpanBlockquoteMarker})
	}
	if m := listMarkerStart(line); m >= 0 {
		spans = append(spans, SyntaxSpan{Start: m, End: m + listMarkerLen(line, m), SpanType: SpanListMarker})
	}
	if IsHorizontalRule(line) {
		spans = append(spans, SyntaxSpan{Start: 0, End: len(line), SpanType: SpanHR})
	}
	return spans
}

func blockquotePrefix(s string) (int, int) {
	depth, i := 0, 0
	for i < len(s) {
		switch {
		case s[i] == '>':
			depth++
			i++
		case s[i] == ' ' || s[i] == '\t':
			i++
		default:
			if depth == 0 {
				return 0, 0
			}
			return depth, i
		}
	}
	if depth == 0 {
		return 0, 0
	}
	return depth, i
}

func leadingRun(s string, c byte) int {
	n := 0
	for n < len(s) && s[n] == c {
		n++
	}
	return n
}

func listMarkerStart(line string) int {
	for i := 0; i < len(line); i++ {
		switch {
		case line[i] == '-' || line[i] == '+' || line[i] == '*':
			if i+1 < len(line) && (line[i+1] == ' ' || line[i+1] == '\t') {
				return i
			}
		case line[i] >= '0' && line[i] <= '9':
			if _, ok := numberMarkerEnd(line, i); ok {
				return i
			}
		default:
			return -1
		}
	}
	return -1
}

func numberMarkerEnd(line string, i int) (int, bool) {
	j := i
	for j < len(line) && line[j] >= '0' && line[j] <= '9' {
		j++
	}
	if j < len(line) && line[j] == '.' && j+1 < len(line) && (line[j+1] == ' ' || line[j+1] == '\t') {
		return j + 1, true
	}
	return 0, false
}

func listMarkerLen(line string, m int) int {
	if line[m] >= '0' && line[m] <= '9' {
		if end, ok := numberMarkerEnd(line, m); ok {
			return end - m + 1
		}
		return 1
	}
	return 1
}

// SortSpans orders spans by start offset, then by end offset.
func SortSpans(spans []SyntaxSpan) {
	for i := 1; i < len(spans); i++ {
		for j := i; j > 0 && (spans[j].Start < spans[j-1].Start ||
			(spans[j].Start == spans[j-1].Start && spans[j].End < spans[j-1].End)); j-- {
			spans[j], spans[j-1] = spans[j-1], spans[j]
		}
	}
}

// MergeSpans unions overlapping spans so each byte belongs to at most one span.
func MergeSpans(spans []SyntaxSpan) []SyntaxSpan {
	if len(spans) < 2 {
		return spans
	}
	SortSpans(spans)
	out := spans[:1]
	for _, s := range spans[1:] {
		last := &out[len(out)-1]
		if s.Start <= last.End {
			if s.End > last.End {
				last.End = s.End
			}
			continue
		}
		out = append(out, s)
	}
	return out
}
