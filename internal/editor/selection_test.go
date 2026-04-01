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

func TestSelectionCut(t *testing.T) {
	content := "HelloWorld"
	editor := NewEditor(content, testRenderer())

	editor.selection.Activate(5)
	editor.selection.Extend(10)

	cutContent := editor.selection.Cut(editor.buf)

	if string(cutContent) != "World" {
		t.Errorf("Cut content should be 'World', got '%s'", string(cutContent))
	}

	if editor.buf.Len() != 5 {
		t.Errorf("Buffer length should be 5 after cut, got %d", editor.buf.Len())
	}

	if string(editor.buf.Contents()) != "Hello" {
		t.Errorf("Buffer content should be 'Hello', got '%s'", string(editor.buf.Contents()))
	}

	if editor.selection.IsActive() {
		t.Error("Selection should not be active after Cut()")
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

func TestSelectionWithUndo(t *testing.T) {
	content := "HelloWorld"
	editor := NewEditor(content, testRenderer())

	editor.selection.Activate(5)
	editor.selection.Extend(10)

	kb := editor.keyBindings
	kb.handleCut()

	if string(editor.buf.Contents()) != "Hello" {
		t.Errorf("After cut, buffer content should be 'Hello', got '%s'", string(editor.buf.Contents()))
	}

	editor.undo.Undo(editor.buf)

	if string(editor.buf.Contents()) != "HelloWorld" {
		t.Errorf("After undo, buffer content should be 'HelloWorld', got '%s'", string(editor.buf.Contents()))
	}
}
