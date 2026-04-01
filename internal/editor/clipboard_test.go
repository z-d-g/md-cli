package editor

import (
	"testing"
)

func TestPasteHandling(t *testing.T) {
	ed := NewEditor("Hello, world!", testRenderer())
	kb := NewKeyBindings(ed)

	currentCursor := ed.nav.Cursor()
	kb.handlePaste()

	newCursor := ed.nav.Cursor()
	if newCursor < currentCursor {
		t.Errorf("Cursor should not move backward after paste, got %d -> %d", currentCursor, newCursor)
	}

	ed.selection.Activate(0)
	ed.selection.MoveCursor(5)

	currentCursor = ed.nav.Cursor()
	kb.handlePaste()

	newCursor = ed.nav.Cursor()
	if newCursor < 0 {
		t.Errorf("Cursor should not be negative after paste, got %d", newCursor)
	}
}
