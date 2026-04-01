package editor

import (
	"sort"
	"unicode/utf8"
)

const defaultGapBufferSize = 4096

// GapBuffer implements a byte-oriented gap buffer with a line index overlay.
type GapBuffer struct {
	data       []byte
	gapStart   int
	gapEnd     int
	capacity   int
	lineStarts []int
}

// NewGapBuffer creates a gap buffer with optional initial content.
// If initial is nil, the buffer starts empty with the default capacity.
func NewGapBuffer(initial []byte) *GapBuffer {
	if initial == nil {
		initial = []byte{}
	}

	capacity := defaultGapBufferSize
	needed := len(initial) + 1 // ensure we always have some gap space
	for capacity < needed {
		capacity *= 2
	}

	data := make([]byte, capacity)
	copy(data, initial)

	gb := &GapBuffer{
		data:     data,
		gapStart: len(initial),
		gapEnd:   capacity,
		capacity: capacity,
	}
	gb.rebuildLineIndex()
	return gb
}

// Len returns the length of the buffer content (excluding the gap).
func (g *GapBuffer) Len() int {
	return g.capacity - (g.gapEnd - g.gapStart)
}

// Insert inserts text at the given byte offset.
// Offsets are clamped to the current content length.
func (g *GapBuffer) Insert(pos int, text []byte) {
	if len(text) == 0 {
		return
	}

	if pos < 0 {
		pos = 0
	}
	if pos > g.Len() {
		pos = g.Len()
	}

	g.moveGap(pos)
	if g.gapSize() < len(text) {
		g.grow(len(text))
	}

	copy(g.data[g.gapStart:], text)
	g.gapStart += len(text)

	g.rebuildLineIndex()
}

// Delete removes count bytes starting at pos. The range is clamped to the buffer length.
func (g *GapBuffer) Delete(pos, count int) {
	if count <= 0 {
		return
	}

	if pos < 0 {
		pos = 0
	}
	length := g.Len()
	if pos >= length {
		return
	}
	if pos+count > length {
		count = length - pos
	}
	if count <= 0 {
		return
	}

	g.moveGap(pos)
	g.gapEnd += count

	g.rebuildLineIndex()
}

// Contents returns a copy of the buffer contents without the gap.
func (g *GapBuffer) Contents() []byte {
	out := make([]byte, g.Len())
	copy(out, g.data[:g.gapStart])
	copy(out[g.gapStart:], g.data[g.gapEnd:])
	return out
}

// LineAt returns the content of the nth line (0-indexed) without the trailing newline.
// Returns an empty string for out-of-range lines.
func (g *GapBuffer) LineAt(n int) string {
	g.ensureLineIndex()
	if n < 0 || n >= len(g.lineStarts) {
		return ""
	}

	start := g.lineStarts[n]
	end := g.Len()
	if n+1 < len(g.lineStarts) {
		end = g.lineStarts[n+1] - 1 // exclude the newline character
	}
	if end < start {
		end = start
	}

	return string(g.slice(start, end))
}

// LineCount returns the number of lines in the buffer.
func (g *GapBuffer) LineCount() int {
	g.ensureLineIndex()
	return len(g.lineStarts)
}

// ByteOffsetOfLine returns the byte offset of the start of the nth line.
// Returns -1 if the line is out of range.
func (g *GapBuffer) ByteOffsetOfLine(n int) int {
	g.ensureLineIndex()
	if n < 0 || n >= len(g.lineStarts) {
		return -1
	}
	return g.lineStarts[n]
}

// CursorToRowCol converts a byte offset into (row, col) where col is in runes.
func (g *GapBuffer) CursorToRowCol(byteOffset int) (int, int) {
	g.ensureLineIndex()
	length := g.Len()
	if byteOffset < 0 {
		byteOffset = 0
	}
	if byteOffset > length {
		byteOffset = length
	}

	row := len(g.lineStarts) - 1
	idx := sort.Search(len(g.lineStarts), func(i int) bool {
		return g.lineStarts[i] > byteOffset
	})
	if idx > 0 {
		row = idx - 1
	} else {
		row = 0
	}

	lineStart := g.lineStarts[row]
	lineEnd := length
	if row+1 < len(g.lineStarts) {
		lineEnd = g.lineStarts[row+1] - 1
	}
	if lineEnd < lineStart {
		lineEnd = lineStart
	}

	if byteOffset > lineEnd {
		byteOffset = lineEnd
	}

	col := g.runeCountAt(lineStart, byteOffset)
	return row, col
}

// RowColToByteOffset converts a (row, col) where col is in runes to a byte offset.
// Out-of-range rows are clamped to the last line, and cols are clamped to the line length.
func (g *GapBuffer) RowColToByteOffset(row, col int) int {
	g.ensureLineIndex()
	if len(g.lineStarts) == 0 {
		return 0
	}

	if row < 0 {
		row = 0
	}
	if row >= len(g.lineStarts) {
		row = len(g.lineStarts) - 1
	}

	lineStart := g.lineStarts[row]
	lineEnd := g.Len()
	if row+1 < len(g.lineStarts) {
		lineEnd = g.lineStarts[row+1] - 1
	}
	if lineEnd < lineStart {
		lineEnd = lineStart
	}

	if col <= 0 {
		return lineStart
	}

	lineLen := g.runeCountAt(lineStart, lineEnd)
	if col >= lineLen {
		return lineEnd
	}
	return g.byteOffsetOfRune(lineStart, lineEnd, col)
}

func (g *GapBuffer) moveGap(pos int) {
	length := g.Len()
	if pos < 0 {
		pos = 0
	}
	if pos > length {
		pos = length
	}

	if pos < g.gapStart {
		shift := g.gapStart - pos
		copy(g.data[g.gapEnd-shift:g.gapEnd], g.data[pos:g.gapStart])
		g.gapStart -= shift
		g.gapEnd -= shift
	} else if pos > g.gapStart {
		shift := pos - g.gapStart
		copy(g.data[g.gapStart:g.gapStart+shift], g.data[g.gapEnd:g.gapEnd+shift])
		g.gapStart += shift
		g.gapEnd += shift
	}
}

func (g *GapBuffer) grow(minGap int) {
	contentLen := g.Len()
	required := contentLen + minGap
	newCap := g.capacity
	if newCap == 0 {
		newCap = defaultGapBufferSize
	}
	for required > newCap {
		newCap *= 2
	}

	leftLen := g.gapStart
	rightLen := g.capacity - g.gapEnd

	newData := make([]byte, newCap)
	copy(newData, g.data[:leftLen])
	newGapStart := leftLen
	newGapEnd := newCap - rightLen
	copy(newData[newGapEnd:], g.data[g.gapEnd:])

	g.data = newData
	g.gapStart = newGapStart
	g.gapEnd = newGapEnd
	g.capacity = newCap
}

func (g *GapBuffer) gapSize() int {
	return g.gapEnd - g.gapStart
}

func (g *GapBuffer) ensureLineIndex() {
	if len(g.lineStarts) == 0 {
		g.rebuildLineIndex()
	}
}

func (g *GapBuffer) rebuildLineIndex() {
	g.lineStarts = g.lineStarts[:0]
	g.lineStarts = append(g.lineStarts, 0)

	contentLen := g.Len()
	if contentLen == 0 {
		return
	}

	for i := 0; i < g.gapStart; i++ {
		if g.data[i] == '\n' {
			g.lineStarts = append(g.lineStarts, i+1)
		}
	}

	for i := g.gapEnd; i < g.capacity; i++ {
		if g.data[i] == '\n' {
			pos := g.gapStart + (i - g.gapEnd)
			g.lineStarts = append(g.lineStarts, pos+1)
		}
	}
}

// decodeRuneAt decodes the rune starting at byte offset pos.
// Returns the rune and its byte size. No allocation.
func (g *GapBuffer) decodeRuneAt(pos int) (rune, int) {
	if pos < 0 || pos >= g.Len() {
		return utf8.RuneError, 0
	}
	if pos < g.gapStart {
		return utf8.DecodeRune(g.data[pos:])
	}
	return utf8.DecodeRune(g.data[pos+g.gapSize():])
}

// decodeLastRuneAt decodes the last rune ending at byte offset pos.
// Returns the rune and its byte size. No allocation.
func (g *GapBuffer) decodeLastRuneAt(pos int) (rune, int) {
	if pos <= 0 || pos > g.Len() {
		return utf8.RuneError, 0
	}
	if pos <= g.gapStart {
		return utf8.DecodeLastRune(g.data[:pos])
	}
	gapSize := g.gapSize()
	if pos-gapSize >= g.gapStart {
		return utf8.DecodeLastRune(g.data[g.gapEnd : pos-gapSize+g.gapEnd])
	}
	// pos straddles the gap: read from before gap start
	return utf8.DecodeLastRune(g.data[:g.gapStart])
}

// runeCountAt returns the number of runes in the segment [start, end).
func (g *GapBuffer) runeCountAt(start, end int) int {
	count := 0
	pos := start
	for pos < end {
		_, size := g.decodeRuneAt(pos)
		if size == 0 {
			break
		}
		count++
		pos += size
	}
	return count
}

// byteOffsetOfRune returns the byte offset of the runeIndex-th rune in [start, end).
func (g *GapBuffer) byteOffsetOfRune(start, end, runeIndex int) int {
	pos := start
	for i := 0; i < runeIndex && pos < end; i++ {
		_, size := g.decodeRuneAt(pos)
		if size == 0 {
			break
		}
		pos += size
	}
	return pos
}

// slice returns a copy of the content between start and end (exclusive of end).
func (g *GapBuffer) slice(start, end int) []byte {
	if start < 0 {
		start = 0
	}
	contentLen := g.Len()
	if end > contentLen {
		end = contentLen
	}
	if start > end {
		start = end
	}

	gapSize := g.gapSize()
	res := make([]byte, end-start)

	switch {
	case end <= g.gapStart:
		copy(res, g.data[start:end])
	case start >= g.gapStart:
		copy(res, g.data[start+gapSize:end+gapSize])
	default:
		leftLen := g.gapStart - start
		copy(res, g.data[start:g.gapStart])
		copy(res[leftLen:], g.data[g.gapEnd:g.gapEnd+(end-g.gapStart)])
	}

	return res
}
