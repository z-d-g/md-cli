package editor

// UndoEntry is a single undoable operation. Byte offsets index the gap buffer.
// Insertions set inserted and leave deleted empty; deletions do the reverse.
// cursorBefore/cursorAfter restore the caret on undo and redo.
type UndoEntry struct {
	offset       int
	deleted      []byte
	inserted     []byte
	cursorBefore int
	cursorAfter  int
}

// UndoManager stores undo/redo stacks in fixed-size ring buffers, so memory
// stays bounded and no backing array leaks when the oldest entry is dropped.
type UndoManager struct {
	undo ring
	redo ring
}

type ring struct {
	buf   []UndoEntry
	start int
	size  int
}

func newRing(limit int) ring { return ring{buf: make([]UndoEntry, limit)} }

func (r *ring) len() int { return r.size }

func (r *ring) clear() {
	r.start = 0
	r.size = 0
}

func (r *ring) push(e UndoEntry) {
	if len(r.buf) == 0 {
		return
	}
	if r.size == len(r.buf) {
		r.start = (r.start + 1) % len(r.buf)
		r.size--
	}
	r.buf[(r.start+r.size)%len(r.buf)] = e
	r.size++
}

func (r *ring) pop() (UndoEntry, bool) {
	if r.size == 0 {
		return UndoEntry{}, false
	}
	idx := (r.start + r.size - 1) % len(r.buf)
	e := r.buf[idx]
	r.buf[idx] = UndoEntry{}
	r.size--
	return e, true
}

func (r *ring) at(i int) UndoEntry     { return r.buf[(r.start+i)%len(r.buf)] }
func (r *ring) set(i int, e UndoEntry) { r.buf[(r.start+i)%len(r.buf)] = e }

// NewUndoManager creates a manager holding up to limit entries per stack.
func NewUndoManager(limit int) *UndoManager {
	if limit <= 0 {
		limit = 500
	}
	return &UndoManager{undo: newRing(limit), redo: newRing(limit)}
}

// Record pushes a new operation and clears the redo stack.
func (um *UndoManager) Record(entry UndoEntry) {
	um.redo.clear()
	um.undo.push(entry)
}

// Undo reverses the newest operation and returns the cursor to restore.
func (um *UndoManager) Undo(buf *GapBuffer) (int, bool) {
	entry, ok := um.undo.pop()
	if !ok {
		return 0, false
	}
	applyInverse(buf, entry)
	um.redo.push(entry)
	return entry.cursorBefore, true
}

// Redo re-applies the most recently undone operation.
func (um *UndoManager) Redo(buf *GapBuffer) (int, bool) {
	entry, ok := um.redo.pop()
	if !ok {
		return 0, false
	}
	applyForward(buf, entry)
	um.undo.push(entry)
	return entry.cursorAfter, true
}

func applyInverse(buf *GapBuffer, e UndoEntry) {
	if len(e.inserted) > 0 {
		buf.Delete(e.offset, len(e.inserted))
	}
	if len(e.deleted) > 0 {
		buf.Insert(e.offset, e.deleted)
	}
}

func applyForward(buf *GapBuffer, e UndoEntry) {
	if len(e.deleted) > 0 {
		buf.Delete(e.offset, len(e.deleted))
	}
	if len(e.inserted) > 0 {
		buf.Insert(e.offset, e.inserted)
	}
}

// CanUndo reports whether an undo is available.
func (um *UndoManager) CanUndo() bool { return um.undo.len() > 0 }

// CanRedo reports whether a redo is available.
func (um *UndoManager) CanRedo() bool { return um.redo.len() > 0 }

// Len returns the number of undoable operations.
func (um *UndoManager) Len() int { return um.undo.len() }

// Clear drops all undo and redo history.
func (um *UndoManager) Clear() {
	um.undo.clear()
	um.redo.clear()
}

// GroupUndoEntries merges the trailing run of single-rune inserts that are
// adjacent in both the buffer and the timeline into one entry. It returns the
// number of entries collapsed.
func (um *UndoManager) GroupUndoEntries(startIndex int) int {
	if startIndex < 0 || startIndex >= um.undo.len() {
		return 0
	}
	out := make([]UndoEntry, 0, um.undo.len())
	for i := 0; i < um.undo.len(); i++ {
		out = append(out, um.undo.at(i))
	}
	grouped := out[:startIndex]
	for i := startIndex; i < len(out); {
		e := out[i]
		i++
		if len(e.inserted) != 1 {
			grouped = append(grouped, e)
			continue
		}
		merged := 1
		for i < len(out) && len(out[i].inserted) == 1 &&
			e.offset+len(e.inserted) == out[i].offset &&
			e.cursorAfter == out[i].cursorBefore {
			e.inserted = append(e.inserted, out[i].inserted...)
			e.cursorAfter = out[i].cursorAfter
			i++
			merged++
		}
		grouped = append(grouped, e)
	}
	um.undo.clear()
	for _, e := range grouped {
		um.undo.push(e)
	}
	return len(out) - len(grouped)
}
