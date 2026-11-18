package editor

import (
	"testing"
)

func TestSelectionVisualization(t *testing.T) {
	content := "Line 1\nLine 2\nLine 3\nLine 4"
	editor := NewEditor(content, testRenderer())

	editor.selection.Activate(0)
	editor.selection.Extend(4)

	if !editor.selection.IsActive() {
		t.Error("Selection should be active")
	}

	if editor.selection.Start() != 0 {
		t.Errorf("Selection start should be 0, got %d", editor.selection.Start())
	}

	if editor.selection.End() != 4 {
		t.Errorf("Selection end should be 4, got %d", editor.selection.End())
	}

	editor.selection.Activate(0)
	editor.selection.Extend(12)

	if editor.selection.Start() != 0 {
		t.Errorf("Multi-line selection start should be 0, got %d", editor.selection.Start())
	}

	if editor.selection.End() != 12 {
		t.Errorf("Multi-line selection end should be 12, got %d", editor.selection.End())
	}

	editor.selection.Clear()

	if editor.selection.IsActive() {
		t.Error("Selection should not be active after Clear()")
	}
}

func TestSelectionGetSelectedText(t *testing.T) {
	content := "Hello World"
	editor := NewEditor(content, testRenderer())

	editor.selection.Activate(6)
	editor.selection.Extend(11)

	selected := editor.selection.GetSelectedText(editor.buf)

	if string(selected) != "World" {
		t.Errorf("Selected text should be 'World', got '%s'", string(selected))
	}

	if !editor.selection.IsActive() {
		t.Error("Selection should remain active after GetSelectedText()")
	}
}

func TestSelectionSelectAll(t *testing.T) {
	content := "Hello World"
	editor := NewEditor(content, testRenderer())

	editor.selection.SelectAll(editor.buf)

	if !editor.selection.IsActive() {
		t.Error("selection should be active")
	}
	if editor.selection.Start() != 0 {
		t.Errorf("start = %d, want 0", editor.selection.Start())
	}
	if editor.selection.End() != len(content) {
		t.Errorf("end = %d, want %d", editor.selection.End(), len(content))
	}
}

func TestSelectionSelectWord(t *testing.T) {
	content := "hello world"
	editor := NewEditor(content, testRenderer())

	editor.selection.SelectWord(editor.buf, 2) // inside "hello"

	if editor.selection.Start() != 0 {
		t.Errorf("word start = %d, want 0", editor.selection.Start())
	}
	if editor.selection.End() != 5 {
		t.Errorf("word end = %d, want 5", editor.selection.End())
	}

	// Select second word
	editor.selection.SelectWord(editor.buf, 7) // inside "world"
	if editor.selection.Start() != 6 {
		t.Errorf("word start = %d, want 6", editor.selection.Start())
	}
	if editor.selection.End() != 11 {
		t.Errorf("word end = %d, want 11", editor.selection.End())
	}
}

func TestSelectionSelectWordEmpty(t *testing.T) {
	editor := NewEditor("", testRenderer())
	editor.selection.SelectWord(editor.buf, 0)
	if editor.selection.IsActive() {
		t.Error("selecting word in empty buffer should not activate")
	}
}

func TestSelectionSelectLine(t *testing.T) {
	content := "line1\nline2\nline3"
	editor := NewEditor(content, testRenderer())

	editor.selection.SelectLine(editor.buf, 2) // inside first line

	if editor.selection.Start() != 0 {
		t.Errorf("line start = %d, want 0", editor.selection.Start())
	}
	if editor.selection.End() != 5 {
		t.Errorf("line end = %d, want 5", editor.selection.End())
	}
}

func TestSelectionMoveCursor(t *testing.T) {
	sel := NewSelection()

	// First move activates
	sel.MoveCursor(5)
	if !sel.IsActive() {
		t.Error("MoveCursor should activate selection")
	}
	if sel.Start() != sel.End() {
		t.Errorf("empty selection = [%d,%d), want equal", sel.Start(), sel.End())
	}

	// Second move extends
	sel.MoveCursor(10)
	if sel.End() != 10 {
		t.Errorf("end = %d, want 10", sel.End())
	}
}

func TestSelectionGetSelectedTextClamp(t *testing.T) {
	content := "short"
	editor := NewEditor(content, testRenderer())

	editor.selection.Activate(0)
	editor.selection.Extend(100) // beyond buffer

	selected := editor.selection.GetSelectedText(editor.buf)
	if string(selected) != "short" {
		t.Errorf("clamped text = %q, want %q", selected, "short")
	}
}
