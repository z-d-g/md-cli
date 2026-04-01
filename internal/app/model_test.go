package app

import (
	"strings"
	"testing"

	"github.com/z-d-g/md-cli/internal/config"
	"github.com/z-d-g/md-cli/internal/editor"
	"github.com/z-d-g/md-cli/internal/render"

	tea "charm.land/bubbletea/v2"
)

func lineCount(s string) int {
	if s == "" {
		return 0
	}
	return strings.Count(s, "\n") + 1
}

func testEditor(content string) *editor.Editor {
	cfg := config.LoadConfig()
	r := render.NewLipglossRenderer(&cfg.EditorStyles)
	return editor.NewEditor(content, r)
}

func TestWindowSizeRespectsReservedRows(t *testing.T) {
	cfg := config.LoadConfig()
	ed := testEditor("a\nb\nc\nd\ne\nf\ng\nh\ni\nj\n")
	m := Model{
		Editor:       ed,
		Config:       cfg,
		FilePath:     "test.md",
		SavedContent: ed.Value(),
	}

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 10})
	m2 := updated.(Model)

	editorLines := lineCount(m2.Editor.View().Content)
	expectedEditorLines := 10 - m2.reservedRows()
	if editorLines != expectedEditorLines {
		t.Errorf("Editor viewport height mismatch: got %d lines, want %d lines", editorLines, expectedEditorLines)
	}

	appLines := lineCount(m2.View().Content)
	expectedAppLines := 10
	if appLines != expectedAppLines {
		t.Errorf("App view height mismatch: got %d lines, want %d lines", appLines, expectedAppLines)
	}
}

func TestWindowSizeClampsToZero(t *testing.T) {
	cfg := config.LoadConfig()
	ed := testEditor("a\nb\nc\n")
	m := Model{
		Editor:       ed,
		Config:       cfg,
		FilePath:     "test.md",
		SavedContent: ed.Value(),
	}

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 1})
	m2 := updated.(Model)

	editorLines := lineCount(m2.Editor.View().Content)
	expectedEditorLines := 0
	if editorLines != expectedEditorLines {
		t.Errorf("Editor viewport height not clamped: got %d lines, want %d lines", editorLines, expectedEditorLines)
	}
}
