package markdown

// CodeFenceTracker resolves fenced-code-block membership and bounds for a
// snapshot of document lines. It replaces the per-package re-walks that used
// to live in both the editor frame state and FindBlockRegion, giving a single
// char-aware source of truth. A tracker is bound to a fixed line snapshot and
// built lazily, so it is safe to hold for the lifetime of one rendered frame.
type CodeFenceTracker struct {
	lines []string
	code  []bool
	built bool
}

// NewCodeFenceTracker binds the tracker to a snapshot of document lines. The
// per-line membership array is computed on the first call to CodeBlockLines.
func NewCodeFenceTracker(lines []string) *CodeFenceTracker {
	return &CodeFenceTracker{lines: lines}
}

// CodeBlockLines returns the per-line "inside a code block" array, building it
// on the first call. A fence delimiter line itself is not treated as code
// content (code[fenceLine] is false). Fences are matched by character: a ``` block
// is not closed by a ~~~ line.
func (t *CodeFenceTracker) CodeBlockLines() []bool {
	if t.built {
		return t.code
	}
	t.code = make([]bool, len(t.lines))
	inside := false
	fenceChar := byte(0)
	for i := 0; i < len(t.lines); i++ {
		if !IsCodeFence(t.lines[i]) {
			t.code[i] = inside
			continue
		}
		fc := CodeFenceChar(t.lines[i])
		if inside {
			if fc != 0 && fenceChar != fc {
				t.code[i] = true
				continue
			}
			t.code[i] = false
			inside = false
			fenceChar = 0
		} else {
			t.code[i] = false
			inside = true
			fenceChar = fc
		}
	}
	t.built = true
	return t.code
}

// IsInside reports whether row is strictly inside a code block (not on a fence
// delimiter line).
func (t *CodeFenceTracker) IsInside(row int) bool {
	code := t.CodeBlockLines()
	return row >= 0 && row < len(code) && code[row]
}

// CodeBlockBounds returns the inclusive [start, end] line range of the code
// block that contains row, or (row, row) if row is not part of a block. A row
// sitting on a fence delimiter is considered part of its block.
func (t *CodeFenceTracker) CodeBlockBounds(row int) (int, int) {
	n := len(t.lines)
	if row < 0 || row >= n {
		return row, row
	}
	inside := false
	openStart := 0
	fenceChar := byte(0)
	for i := 0; i <= row; i++ {
		if !IsCodeFence(t.lines[i]) {
			continue
		}
		fc := CodeFenceChar(t.lines[i])
		if inside && fc != 0 && fenceChar != fc {
			continue
		}
		if inside {
			if i == row {
				return openStart, row
			}
			inside = false
			fenceChar = 0
		} else {
			inside = true
			openStart = i
			fenceChar = fc
			if i == row {
				return openStart, t.findClose(openStart, fenceChar)
			}
		}
	}
	if inside {
		return openStart, t.findClose(openStart, fenceChar)
	}
	return row, row
}

func (t *CodeFenceTracker) findClose(from int, fenceChar byte) int {
	n := len(t.lines)
	for i := from + 1; i < n; i++ {
		if !IsCodeFence(t.lines[i]) {
			continue
		}
		fc := CodeFenceChar(t.lines[i])
		if fc == 0 || fenceChar == 0 || fc == fenceChar {
			return i
		}
	}
	return n - 1
}
