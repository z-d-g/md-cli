package markdown

import (
	"strings"

	"github.com/yuin/goldmark/v2/ast"
	extast "github.com/yuin/goldmark/v2/extension/ast"
)

// Doc is a parsed markdown document. Build one per content change and read it
// for the lifetime of a frame; every query is a lookup, not a re-parse.
type Doc struct {
	source string
	lines  []string
	// code[i] is true when line i is a code-block content line (not the fence).
	code []bool
	// bounds[i] is the inclusive line range of the block containing line i.
	bounds [][2]int
	// kind[i] is true when line i starts a block that renders as raw source
	// when the cursor is on it.
	kind []int
	// inCodeBlock is kept for callers that need "is this line code" semantics.
	inCodeBlock []bool
	// table, sep, list and quote are per-line block memberships.
	table []bool
	sep   []bool
	list  []bool
	quote []bool
}

// NewDoc parses content into a Doc.
func NewDoc(content string) *Doc {
	d := &Doc{source: content, lines: strings.Split(content, "\n")}
	d.index()
	return d
}

// Lines returns the document lines.
func (d *Doc) Lines() []string { return d.lines }

// Len returns the number of lines.
func (d *Doc) Len() int { return len(d.lines) }

// IsCodeFence reports whether line i is a fence delimiter.
func (d *Doc) IsCodeFence(i int) bool {
	return i >= 0 && i < len(d.kind) && d.kind[i] == LineCodeFence
}

// IsHeading returns the heading level for line i, or 0 when it is not a heading.
func (d *Doc) HeadingLevel(i int) int {
	if i < 0 || i >= len(d.kind) {
		return 0
	}
	if k := d.kind[i]; k >= LineHeading1 && k <= LineHeading6 {
		return k
	}
	return 0
}

// IsTable reports whether line i belongs to a table.
func (d *Doc) IsTable(i int) bool {
	if i < 0 || i >= len(d.table) {
		return false
	}
	return d.table[i]
}

// IsTableSeparator reports whether line i is a table delimiter row.
func (d *Doc) IsTableSeparator(i int) bool {
	if i < 0 || i >= len(d.sep) {
		return false
	}
	return d.sep[i]
}

// IsList reports whether line i is a list item.
func (d *Doc) IsList(i int) bool {
	if i < 0 || i >= len(d.list) {
		return false
	}
	return d.list[i]
}

// IsBlockquote reports whether line i is inside a blockquote.
func (d *Doc) IsBlockquote(i int) bool {
	if i < 0 || i >= len(d.quote) {
		return false
	}
	return d.quote[i]
}

// LineKind returns the styling class for line i.
func (d *Doc) LineKind(i int) int {
	if i < 0 || i >= len(d.kind) {
		return LineNormal
	}
	return d.kind[i]
}

// InCode reports whether line i is code-block content.
func (d *Doc) InCode(i int) bool {
	if i < 0 || i >= len(d.inCodeBlock) {
		return false
	}
	return d.inCodeBlock[i]
}

// Bounds returns the inclusive line range of the block containing i.
func (d *Doc) Bounds(i int) (int, int) {
	if i < 0 || i >= len(d.bounds) {
		return i, i
	}
	return d.bounds[i][0], d.bounds[i][1]
}

func (d *Doc) index() {
	n := len(d.lines)
	d.kind = make([]int, n)
	d.bounds = make([][2]int, n)
	d.inCodeBlock = make([]bool, n)
	d.code = make([]bool, n)
	d.table = make([]bool, n)
	d.sep = make([]bool, n)
	d.list = make([]bool, n)
	d.quote = make([]bool, n)
	for i := range d.bounds {
		d.bounds[i] = [2]int{i, i}
	}

	root := md.Parse([]byte(d.source))
	collectBlocks(root, d, 0)
}

// collectBlocks walks block nodes, filling classification and bounds.
func collectBlocks(n ast.Node, d *Doc, depth int) {
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		start, end := nodeLineRange(c, d.lines)
		switch v := c.(type) {
		case *ast.CodeBlock:
			d.markCode(start, end, v.CodeBlockKind == ast.CodeBlockKindFenced)
		case *ast.Heading:
			d.kind[start] = v.Level
		case *ast.Blockquote:
			d.markBlock(start, end, LineBlockQuote)
			for i := start; i <= end && i < len(d.quote); i++ {
				d.quote[i] = true
			}
		case *ast.List:
			d.markBlock(start, end, LineNormal)
			for i := start; i <= end && i < len(d.list); i++ {
				d.list[i] = true
			}
		case *extast.Table:
			d.markBlock(start, end, LineNormal)
			d.markTable(start, end)
		case *ast.ThematicBreak:
			d.kind[start] = LineNormal
		}
		collectBlocks(c, d, depth+1)
	}
}

// markCode records fence lines and code content lines.
func (d *Doc) markCode(start, end int, fenced bool) {
	if start < 0 || start >= len(d.lines) {
		return
	}
	if fenced {
		d.kind[start] = LineCodeFence
		d.kind[end] = LineCodeFence
		for i := start + 1; i < end; i++ {
			d.inCodeBlock[i] = true
			d.code[i] = true
		}
	} else {
		for i := start; i <= end; i++ {
			d.inCodeBlock[i] = true
			d.code[i] = true
		}
	}
	d.markBlock(start, end, LineNormal)
}

func (d *Doc) markBlock(start, end, kind int) {
	for i := start; i <= end && i < len(d.bounds); i++ {
		// Keep the outermost block: a nested list must not shrink the range of
		// the list that contains it.
		if cur := d.bounds[i]; cur[0] == cur[1] {
			d.bounds[i] = [2]int{start, end}
		}
	}
	if start >= 0 && start < len(d.kind) && d.kind[start] == LineNormal {
		d.kind[start] = kind
	}
}

// markTable records table membership and separator rows.
func (d *Doc) markTable(start, end int) {
	for i := start; i <= end; i++ {
		if i >= len(d.table) {
			break
		}
		d.table[i] = true
		if isSeparatorLine(d.lines[i]) {
			d.sep[i] = true
		}
	}
}

// isSeparatorLine reports whether line is a GFM table delimiter row.
func isSeparatorLine(line string) bool {
	if !strings.Contains(line, "-") || !strings.Contains(line, "|") {
		return false
	}
	for _, cell := range strings.Split(line, "|") {
		cell = strings.TrimSpace(cell)
		if cell == "" {
			continue
		}
		if !isSeparatorCell(cell) {
			return false
		}
	}
	return true
}

func isSeparatorCell(cell string) bool {
	cell = strings.TrimPrefix(cell, ":")
	cell = strings.TrimSuffix(cell, ":")
	if cell == "" {
		return false
	}
	for i := 0; i < len(cell); i++ {
		if cell[i] != '-' {
			return false
		}
	}
	return true
}

// nodeLineRange maps a node's source extent onto inclusive zero-based lines.
// It uses the node position plus every descendant offset and source segment, so
// nested lists and multi-line paragraphs get their full span.
func nodeLineRange(n ast.Node, lines []string) (int, int) {
	lo, hi := nodeExtent(n)
	if lo < 0 {
		return 0, 0
	}
	start := lineAtOffset(lines, lo)
	end := lineAtOffset(lines, hi)
	if cb, ok := n.(*ast.CodeBlock); ok && cb.CodeBlockKind == ast.CodeBlockKindFenced {
		// The closing fence is not part of the block's content segments.
		if end+1 < len(lines) && IsCodeFence(lines[end+1]) {
			end++
		}
	}
	if end < start {
		end = start
	}
	return start, end
}

// nodeExtent returns the smallest and largest source offsets covered by n.
func nodeExtent(n ast.Node) (int, int) {
	lo, hi := -1, -1
	take := func(p int) {
		if p < 0 {
			return
		}
		if lo < 0 || p < lo {
			lo = p
		}
		if p > hi {
			hi = p
		}
	}
	take(n.Pos())
	if b, ok := n.(ast.BlockNode); ok {
		for _, seg := range b.Source() {
			take(seg.Start)
			if seg.Stop > 0 {
				take(seg.Stop - 1)
			}
		}
	}
	if cb, ok := n.(*ast.CodeBlock); ok {
		for _, seg := range cb.Value.Segments() {
			take(seg.Start)
			if seg.Stop > 0 {
				take(seg.Stop - 1)
			}
		}
	}
	var walk func(ast.Node)
	walk = func(x ast.Node) {
		for c := x.FirstChild(); c != nil; c = c.NextSibling() {
			take(c.Pos())
			walk(c)
		}
	}
	walk(n)
	return lo, hi
}

func lineAtOffset(lines []string, offset int) int {
	if offset <= 0 {
		return 0
	}
	pos := 0
	for i, l := range lines {
		pos += len(l) + 1
		if offset < pos {
			return i
		}
	}
	return max(len(lines)-1, 0)
}

// Fence describes a fenced-code delimiter line.
type Fence struct {
	Char  byte
	Len   int
	Info  string
	Valid bool
}
