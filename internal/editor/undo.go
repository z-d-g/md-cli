package editor

// UndoEntry represents a single undoable operation
// It stores the minimal information needed to reverse an edit
// and restore the cursor position
// All positions are byte offsets in the gap buffer
// For insertions: deleted is empty, inserted contains the inserted text
// For deletions: deleted contains the deleted text, inserted is empty
// For replacements: both deleted and inserted contain text
// Cursor positions are stored to restore exact cursor location after undo/redo
type UndoEntry struct {
	offset       int    // byte offset where the edit occurred
	deleted      []byte // text that was deleted (empty for pure insert)
	inserted     []byte // text that was inserted (empty for pure delete)
	cursorBefore int    // cursor position before the edit
	cursorAfter  int    // cursor position after the edit
}

// UndoManager manages undo and redo stacks
// It limits the stack size to prevent memory bloat
// All operations are recorded as UndoEntry instances
type UndoManager struct {
	undoStack []UndoEntry // stack of operations that can be undone
	redoStack []UndoEntry // stack of operations that can be redone
	limit     int         // maximum number of entries to keep
}

// NewUndoManager creates a new undo manager with the specified limit
// If limit is 0, a default of 500 is used
func NewUndoManager(limit int) *UndoManager {
	if limit <= 0 {
		limit = 500
	}
	return &UndoManager{
		undoStack: make([]UndoEntry, 0, limit),
		redoStack: make([]UndoEntry, 0, limit),
		limit:     limit,
	}
}

// Record records a new edit operation
// This should be called after every edit to the buffer
// The entry is pushed to the undo stack and the redo stack is cleared
func (um *UndoManager) Record(entry UndoEntry) {
	// Clear redo stack when new operation is recorded
	um.redoStack = um.redoStack[:0]

	// Add to undo stack
	um.undoStack = append(um.undoStack, entry)

	// Enforce size limit
	if len(um.undoStack) > um.limit {
		um.undoStack = um.undoStack[1:]
	}
}

// Undo reverses the last recorded operation
// Returns the cursor position to restore and whether an undo was performed
// If no undo is available, returns (0, false)
func (um *UndoManager) Undo(buf *GapBuffer) (cursorPos int, ok bool) {
	if len(um.undoStack) == 0 {
		return 0, false
	}

	// Pop the last undo entry
	entry := um.undoStack[len(um.undoStack)-1]
	um.undoStack = um.undoStack[:len(um.undoStack)-1]

	// Apply the reverse operation
	if len(entry.inserted) > 0 {
		// Reverse insert = delete
		buf.Delete(entry.offset, len(entry.inserted))
	}
	if len(entry.deleted) > 0 {
		// Reverse delete = insert
		buf.Insert(entry.offset, entry.deleted)
	}

	// Push to redo stack
	um.redoStack = append(um.redoStack, entry)

	// Return cursor position before the operation
	return entry.cursorBefore, true
}

// Redo re-applies the last undone operation
// Returns the cursor position to restore and whether a redo was performed
// If no redo is available, returns (0, false)
func (um *UndoManager) Redo(buf *GapBuffer) (cursorPos int, ok bool) {
	if len(um.redoStack) == 0 {
		return 0, false
	}

	// Pop the last redo entry
	entry := um.redoStack[len(um.redoStack)-1]
	um.redoStack = um.redoStack[:len(um.redoStack)-1]

	// Apply the original operation
	if len(entry.deleted) > 0 {
		// Original delete
		buf.Delete(entry.offset, len(entry.deleted))
	}
	if len(entry.inserted) > 0 {
		// Original insert
		buf.Insert(entry.offset, entry.inserted)
	}

	// Push back to undo stack
	um.undoStack = append(um.undoStack, entry)

	// Return cursor position after the operation
	return entry.cursorAfter, true
}

// CanUndo returns whether there are operations that can be undone
func (um *UndoManager) CanUndo() bool {
	return len(um.undoStack) > 0
}

// Len returns the number of operations in the undo stack
func (um *UndoManager) Len() int {
	return len(um.undoStack)
}

// CanRedo returns whether there are operations that can be redone
func (um *UndoManager) CanRedo() bool {
	return len(um.redoStack) > 0
}

// Clear clears both undo and redo stacks
func (um *UndoManager) Clear() {
	um.undoStack = um.undoStack[:0]
	um.redoStack = um.redoStack[:0]
}

// GroupUndoEntries groups rapid single-character operations into a single undo entry
// This is called when a typing session ends (e.g., when typing stops for 500ms)
// startIndex is the index in undoStack where grouping should start
// Returns the number of entries grouped
func (um *UndoManager) GroupUndoEntries(startIndex int) int {
	if startIndex < 0 || startIndex >= len(um.undoStack) {
		return 0
	}

	grouped := 0
	for i := startIndex; i < len(um.undoStack)-1; i++ {
		current := um.undoStack[i]
		next := um.undoStack[i+1]

		// Check if these are consecutive single-character inserts
		if len(current.inserted) == 1 && len(next.inserted) == 1 &&
			next.offset == current.offset+len(current.inserted) &&
			next.cursorBefore == current.cursorAfter {

			// Merge the operations
			merged := UndoEntry{
				offset:       current.offset,
				inserted:     append(current.inserted, next.inserted...),
				cursorBefore: current.cursorBefore,
				cursorAfter:  next.cursorAfter,
			}

			// Replace current with merged, remove next
			um.undoStack[i] = merged
			um.undoStack = append(um.undoStack[:i+1], um.undoStack[i+2:]...)
			grouped++
			i-- // Check the same position again
		}
	}

	return grouped
}
