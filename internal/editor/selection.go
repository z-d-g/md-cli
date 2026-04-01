package editor

// Selection represents a text selection in the buffer
// It tracks the anchor (start) and cursor (end) positions as byte offsets
// When active is false, there is no selection
type Selection struct {
	active bool
	anchor int // byte offset of selection start
	cursor int // byte offset of selection end
}

// NewSelection creates a new selection manager
func NewSelection() *Selection {
	return &Selection{
		active: false,
		anchor: 0,
		cursor: 0,
	}
}

// Start returns the start position of the selection (minimum of anchor and cursor)
func (s *Selection) Start() int {
	if s.anchor <= s.cursor {
		return s.anchor
	}
	return s.cursor
}

// End returns the end position of the selection (maximum of anchor and cursor)
func (s *Selection) End() int {
	if s.anchor >= s.cursor {
		return s.anchor
	}
	return s.cursor
}

// Length returns the length of the selection in bytes
func (s *Selection) Length() int {
	if !s.active {
		return 0
	}
	return s.End() - s.Start()
}

// IsActive returns whether there is an active selection
func (s *Selection) IsActive() bool {
	return s.active
}

// Activate starts a selection at the given anchor position
func (s *Selection) Activate(anchor int) {
	s.active = true
	s.anchor = anchor
	s.cursor = anchor
}

// Extend extends the selection to the given cursor position
func (s *Selection) Extend(cursor int) {
	if !s.active {
		s.Activate(cursor)
		return
	}
	s.cursor = cursor
}

// Clear clears the selection
func (s *Selection) Clear() {
	s.active = false
	s.anchor = 0
	s.cursor = 0
}

// GetSelectedText returns the selected text from the buffer
// Returns empty slice if no selection or if selection is invalid
func (s *Selection) GetSelectedText(buf *GapBuffer) []byte {
	if !s.active || s.Start() == s.End() {
		return nil
	}

	start := s.Start()
	end := s.End()

	// Clamp to buffer bounds
	if start < 0 {
		start = 0
	}
	if end > buf.Len() {
		end = buf.Len()
	}
	if start >= end {
		return nil
	}

	return buf.slice(start, end)
}

// Cut deletes the selected text and returns it
// Also clears the selection
func (s *Selection) Cut(buf *GapBuffer) []byte {
	if !s.active {
		return nil
	}

	selected := s.GetSelectedText(buf)
	if selected == nil {
		s.Clear()
		return nil
	}

	start := s.Start()
	end := s.End()

	buf.Delete(start, end-start)
	s.Clear()

	return selected
}

// SelectAll selects the entire buffer content
func (s *Selection) SelectAll(buf *GapBuffer) {
	s.Activate(0)
	s.Extend(buf.Len())
}

// SelectWord selects the word at the given cursor position
// Word boundaries are defined by whitespace
func (s *Selection) SelectWord(buf *GapBuffer, cursor int) {
	if buf.Len() == 0 {
		return
	}

	// Find word boundaries
	start := cursor
	end := cursor

	// Move left to find word start
	for start > 0 {
		r, size := buf.decodeLastRuneAt(start)
		if r == ' ' || r == '\t' || r == '\n' {
			break
		}
		start -= size
	}

	// Move right to find word end
	for end < buf.Len() {
		r, size := buf.decodeRuneAt(end)
		if r == ' ' || r == '\t' || r == '\n' {
			break
		}
		end += size
	}

	s.Activate(start)
	s.Extend(end)
}

// SelectLine selects the entire line containing the given cursor position
func (s *Selection) SelectLine(buf *GapBuffer, cursor int) {
	row, _ := buf.CursorToRowCol(cursor)

	lineStart := buf.ByteOffsetOfLine(row)
	lineEnd := buf.Len()
	if row+1 < buf.LineCount() {
		lineEnd = buf.ByteOffsetOfLine(row+1) - 1 // exclude newline
	}
	if lineEnd < lineStart {
		lineEnd = lineStart
	}

	s.Activate(lineStart)
	s.Extend(lineEnd)
}

// MoveCursor moves the cursor position while maintaining selection
// This is used for shift+arrow key operations
func (s *Selection) MoveCursor(cursor int) {
	if !s.active {
		s.Activate(cursor)
	} else {
		s.Extend(cursor)
	}
}
