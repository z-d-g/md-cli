package editor

type Navigation struct {
	buf        *GapBuffer
	cursor     int // current byte offset
	desiredCol int // desired column for vertical movement
}

func NewNavigation(buf *GapBuffer, initialCursor int) *Navigation {
	return &Navigation{
		buf:        buf,
		cursor:     initialCursor,
		desiredCol: 0,
	}
}

func (n *Navigation) Cursor() int {
	return n.cursor
}

func (n *Navigation) SetCursor(pos int) {
	if pos < 0 {
		pos = 0
	}
	if pos > n.buf.Len() {
		pos = n.buf.Len()
	}
	n.cursor = pos
	n.updateDesiredCol()
}

func (n *Navigation) MoveLeft() {
	if n.cursor == 0 {
		return
	}
	_, size := n.buf.decodeLastRuneAt(n.cursor)
	n.cursor -= size
	n.updateDesiredCol()
}

func (n *Navigation) MoveRight() {
	if n.cursor >= n.buf.Len() {
		return
	}
	_, size := n.buf.decodeRuneAt(n.cursor)
	n.cursor += size
	n.updateDesiredCol()
}

func (n *Navigation) MoveUp() {
	row, _ := n.buf.CursorToRowCol(n.cursor)
	if row == 0 {
		return // Already at top
	}

	prevLineStart := n.buf.ByteOffsetOfLine(row - 1)
	prevLineEnd := max(n.buf.ByteOffsetOfLine(row)-1, prevLineStart)
	prevLineRuneCount := n.buf.runeCountAt(prevLineStart, prevLineEnd)
	targetCol := min(n.desiredCol, prevLineRuneCount)
	n.cursor = n.buf.RowColToByteOffset(row-1, targetCol)
}

func (n *Navigation) MoveDown() {
	row, _ := n.buf.CursorToRowCol(n.cursor)
	if row >= n.buf.LineCount()-1 {
		return // Already at bottom
	}

	// Calculate new position
	nextLineStart := n.buf.ByteOffsetOfLine(row + 1)
	nextLineEnd := n.buf.Len()
	if row+2 < n.buf.LineCount() {
		nextLineEnd = n.buf.ByteOffsetOfLine(row+2) - 1
	}
	if nextLineEnd < nextLineStart {
		nextLineEnd = nextLineStart
	}

	nextLineRuneCount := n.buf.runeCountAt(nextLineStart, nextLineEnd)
	targetCol := min(n.desiredCol, nextLineRuneCount)
	n.cursor = n.buf.RowColToByteOffset(row+1, targetCol)
}

func (n *Navigation) MoveWordLeft() {
	if n.cursor == 0 {
		return
	}

	// Skip whitespace backward
	for n.cursor > 0 {
		r, size := n.buf.decodeLastRuneAt(n.cursor)
		if r != ' ' && r != '\t' && r != '\n' {
			break
		}
		n.cursor -= size
	}

	// Skip non-whitespace backward
	for n.cursor > 0 {
		r, size := n.buf.decodeLastRuneAt(n.cursor)
		if r == ' ' || r == '\t' || r == '\n' {
			break
		}
		n.cursor -= size
	}

	n.updateDesiredCol()
}

func (n *Navigation) MoveWordRight() {
	if n.cursor >= n.buf.Len() {
		return
	}

	// Skip non-whitespace forward
	for n.cursor < n.buf.Len() {
		r, size := n.buf.decodeRuneAt(n.cursor)
		if r == ' ' || r == '\t' || r == '\n' {
			break
		}
		n.cursor += size
	}

	// Skip whitespace forward
	for n.cursor < n.buf.Len() {
		r, size := n.buf.decodeRuneAt(n.cursor)
		if r != ' ' && r != '\t' && r != '\n' {
			break
		}
		n.cursor += size
	}

	n.updateDesiredCol()
}

func (n *Navigation) MoveHome() {
	row, _ := n.buf.CursorToRowCol(n.cursor)
	n.cursor = n.buf.ByteOffsetOfLine(row)
	n.updateDesiredCol()
}

func (n *Navigation) MoveEnd() {
	row, _ := n.buf.CursorToRowCol(n.cursor)
	lineStart := n.buf.ByteOffsetOfLine(row)
	lineEnd := n.buf.Len()
	if row+1 < n.buf.LineCount() {
		lineEnd = n.buf.ByteOffsetOfLine(row+1) - 1
	}
	if lineEnd < lineStart {
		lineEnd = lineStart
	}
	n.cursor = lineEnd
	n.updateDesiredCol()
}

func (n *Navigation) MoveDocStart() {
	n.cursor = 0
	n.desiredCol = 0
}

func (n *Navigation) MoveDocEnd() {
	n.cursor = n.buf.Len()
	n.updateDesiredCol()
}

func (n *Navigation) PageUp(viewportHeight int) {
	if viewportHeight <= 0 {
		return
	}

	row, _ := n.buf.CursorToRowCol(n.cursor)
	newRow := max(row-viewportHeight, 0)

	n.cursor = n.buf.RowColToByteOffset(newRow, n.desiredCol)
}

func (n *Navigation) PageDown(viewportHeight int) {
	if viewportHeight <= 0 {
		return
	}

	row, _ := n.buf.CursorToRowCol(n.cursor)
	newRow := row + viewportHeight
	if newRow >= n.buf.LineCount() {
		newRow = n.buf.LineCount() - 1
	}

	n.cursor = n.buf.RowColToByteOffset(newRow, n.desiredCol)
}

func (n *Navigation) updateDesiredCol() {
	row, col := n.buf.CursorToRowCol(n.cursor)
	n.desiredCol = col

	// Special case: if we're at the end of a line that's not the last line,
	// our desired column should be the actual column, not the line length
	lineStart := n.buf.ByteOffsetOfLine(row)
	lineEnd := n.buf.Len()
	if row+1 < n.buf.LineCount() {
		lineEnd = n.buf.ByteOffsetOfLine(row+1) - 1
	}
	if lineEnd < lineStart {
		lineEnd = lineStart
	}

	if n.cursor == lineEnd && row < n.buf.LineCount()-1 {
		n.desiredCol = n.buf.runeCountAt(lineStart, lineEnd)
	}
}
