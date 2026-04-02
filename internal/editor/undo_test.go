package editor

import "testing"

func TestNewUndoManager(t *testing.T) {
	um := NewUndoManager(100)
	if um == nil {
		t.Fatal("NewUndoManager returned nil")
	}
	if um.limit != 100 {
		t.Errorf("limit = %d, want 100", um.limit)
	}
}

func TestNewUndoManagerDefault(t *testing.T) {
	um := NewUndoManager(0)
	if um.limit != 500 {
		t.Errorf("default limit = %d, want 500", um.limit)
	}
	um = NewUndoManager(-1)
	if um.limit != 500 {
		t.Errorf("negative limit = %d, want 500", um.limit)
	}
}

func TestUndoManagerRecord(t *testing.T) {
	um := NewUndoManager(10)
	um.Record(UndoEntry{offset: 0, inserted: []byte("hello"), cursorAfter: 5})

	if !um.CanUndo() {
		t.Error("CanUndo should be true after recording")
	}
	if um.Len() != 1 {
		t.Errorf("Len = %d, want 1", um.Len())
	}
	if um.CanRedo() {
		t.Error("CanRedo should be false after recording")
	}
}

func TestUndoManagerUndo(t *testing.T) {
	buf := NewGapBuffer([]byte("hello"))
	um := NewUndoManager(10)

	um.Record(UndoEntry{
		offset:       0,
		inserted:     []byte("hello"),
		cursorBefore: 0,
		cursorAfter:  5,
	})

	cursorPos, ok := um.Undo(buf)
	if !ok {
		t.Fatal("Undo returned false")
	}
	if cursorPos != 0 {
		t.Errorf("cursorPos = %d, want 0", cursorPos)
	}
	if buf.Len() != 0 {
		t.Errorf("buffer len after undo = %d, want 0", buf.Len())
	}
	if !um.CanRedo() {
		t.Error("CanRedo should be true after undo")
	}
}

func TestUndoManagerUndoEmpty(t *testing.T) {
	buf := NewGapBuffer([]byte(""))
	um := NewUndoManager(10)

	_, ok := um.Undo(buf)
	if ok {
		t.Error("Undo on empty stack should return false")
	}
}

func TestUndoManagerRedo(t *testing.T) {
	buf := NewGapBuffer([]byte(""))
	um := NewUndoManager(10)

	entry := UndoEntry{
		offset:       0,
		inserted:     []byte("hello"),
		cursorBefore: 0,
		cursorAfter:  5,
	}
	um.Record(entry)
	um.Undo(buf)

	cursorPos, ok := um.Redo(buf)
	if !ok {
		t.Fatal("Redo returned false")
	}
	if cursorPos != 5 {
		t.Errorf("cursorPos = %d, want 5", cursorPos)
	}
	if string(buf.Contents()) != "hello" {
		t.Errorf("buffer = %q, want %q", buf.Contents(), "hello")
	}
}

func TestUndoManagerRedoEmpty(t *testing.T) {
	buf := NewGapBuffer([]byte(""))
	um := NewUndoManager(10)

	_, ok := um.Redo(buf)
	if ok {
		t.Error("Redo on empty redo stack should return false")
	}
}

func TestUndoManagerClear(t *testing.T) {
	um := NewUndoManager(10)
	um.Record(UndoEntry{inserted: []byte("x")})
	um.Clear()

	if um.CanUndo() {
		t.Error("CanUndo should be false after clear")
	}
	if um.CanRedo() {
		t.Error("CanRedo should be false after clear")
	}
	if um.Len() != 0 {
		t.Errorf("Len = %d, want 0", um.Len())
	}
}

func TestUndoManagerRecordClearsRedo(t *testing.T) {
	buf := NewGapBuffer([]byte("ab"))
	um := NewUndoManager(10)

	um.Record(UndoEntry{offset: 0, inserted: []byte("a"), cursorAfter: 1})
	um.Record(UndoEntry{offset: 1, inserted: []byte("b"), cursorAfter: 2})
	um.Undo(buf)
	if !um.CanRedo() {
		t.Fatal("should have redo")
	}

	um.Record(UndoEntry{offset: 0, inserted: []byte("c"), cursorAfter: 1})
	if um.CanRedo() {
		t.Error("recording should clear redo stack")
	}
}

func TestUndoManagerLimit(t *testing.T) {
	um := NewUndoManager(3)
	for i := range 5 {
		um.Record(UndoEntry{offset: i, inserted: []byte("x"), cursorAfter: i + 1})
	}
	if um.Len() != 3 {
		t.Errorf("Len = %d, want 3 (limit)", um.Len())
	}
}

func TestUndoManagerDeleteEntry(t *testing.T) {
	buf := NewGapBuffer([]byte(""))
	um := NewUndoManager(10)

	um.Record(UndoEntry{
		offset:       0,
		deleted:      []byte("hello"),
		cursorBefore: 5,
		cursorAfter:  0,
	})

	_, ok := um.Undo(buf)
	if !ok {
		t.Fatal("Undo returned false")
	}
	if string(buf.Contents()) != "hello" {
		t.Errorf("buffer = %q, want %q", buf.Contents(), "hello")
	}
}

func TestGroupUndoEntries(t *testing.T) {
	um := NewUndoManager(100)

	um.Record(UndoEntry{offset: 0, inserted: []byte("a"), cursorBefore: 0, cursorAfter: 1})
	um.Record(UndoEntry{offset: 1, inserted: []byte("b"), cursorBefore: 1, cursorAfter: 2})
	um.Record(UndoEntry{offset: 2, inserted: []byte("c"), cursorBefore: 2, cursorAfter: 3})

	grouped := um.GroupUndoEntries(0)
	if grouped != 2 {
		t.Errorf("grouped = %d, want 2", grouped)
	}
	if um.Len() != 1 {
		t.Errorf("Len after grouping = %d, want 1", um.Len())
	}
}

func TestGroupUndoEntriesInvalid(t *testing.T) {
	um := NewUndoManager(10)
	if got := um.GroupUndoEntries(-1); got != 0 {
		t.Errorf("GroupEntries(-1) = %d, want 0", got)
	}
	if got := um.GroupUndoEntries(5); got != 0 {
		t.Errorf("GroupEntries(5) on empty = %d, want 0", got)
	}
}
