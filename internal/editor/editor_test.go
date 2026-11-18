package editor

import (
	"strings"
	"testing"

	"github.com/z-d-g/md-cli/v2/internal/config"
	"github.com/z-d-g/md-cli/v2/internal/render"

	tea "charm.land/bubbletea/v2"
)

func testRenderer() render.LineRenderer {
	theme := config.DefaultTheme()
	styles := theme.ToEditorStyles()
	return render.NewLipglossRenderer(&styles, 80)
}

func TestCursorMovement(t *testing.T) {
	editor := NewEditor("hello world", testRenderer())

	// Test initial cursor position
	row, col := editor.Cursor()
	if row != 0 || col != 0 {
		t.Fatalf("expected cursor at (0, 0), got (%d, %d)", row, col)
	}

	// Test right arrow key
	msg := tea.KeyPressMsg{Text: "right"}
	cmd := editor.Update(msg)
	if cmd != nil {
		t.Fatalf("expected no command, got %v", cmd)
	}

	row, col = editor.Cursor()
	if row != 0 || col != 1 {
		t.Fatalf("expected cursor at (0, 1) after right arrow, got (%d, %d)", row, col)
	}

	// Test down arrow key
	msg = tea.KeyPressMsg{Text: "down"}
	cmd = editor.Update(msg)
	if cmd != nil {
		t.Fatalf("expected no command, got %v", cmd)
	}

	row, col = editor.Cursor()
	if row != 0 || col != 1 {
		t.Fatalf("expected cursor at (0, 1) after down arrow, got (%d, %d)", row, col)
	}
}

func TestCursorMovementInvalidatesRenderCache(t *testing.T) {
	content := "Line 1\n**bold text**\nLine 3\n"

	editor := NewEditor(content, testRenderer())

	// Set size and render
	editor.SetSize(80, 20)
	view := editor.View()

	if view.Content == "" {
		t.Fatal("Expected non-empty view")
	}

	// Move cursor to line 1 (with bold syntax)
	editor.nav.SetCursor(editor.buf.RowColToByteOffset(1, 0))

	start, end, isBlock := editor.getActiveRegion()
	if start != 1 || end != 1 || !isBlock {
		t.Errorf("Expected active region (1, 1, true), got (%d, %d, %v)", start, end, isBlock)
	}

	editor.nav.MoveDown()

	start, end, isBlock = editor.getActiveRegion()
	if start != 2 || end != 2 || isBlock {
		t.Errorf("Expected active region (2, 2, false), got (%d, %d, %v)", start, end, isBlock)
	}
}

func TestCursorBarRendering(t *testing.T) {
	editor := &Editor{
		buf:         NewGapBuffer([]byte("Hello world")),
		nav:         NewNavigation(NewGapBuffer([]byte("Hello world")), 0),
		renderCache: make(map[int]cacheEntry),
		syntaxCache: make(map[int][]SyntaxSpan),
		renderer:    testRenderer(),
	}

	t.Run("cursor on character", func(t *testing.T) {
		var b strings.Builder
		rawLine := "Hello world"
		styledLine := "Hello world"

		editor.renderWithCursor(&b, rawLine, styledLine, 5)

		result := b.String()
		if !strings.Contains(result, "Hello") {
			t.Errorf("Expected 'Hello' before cursor, got: %q", result)
		}
		if !strings.Contains(result, "orld") {
			t.Errorf("Expected 'orld' after cursor, got: %q", result)
		}
		if len(result) < len("Hello world") {
			t.Errorf("Result length %d is less than expected, got: %q", len(result), result)
		}
	})

	t.Run("cursor at end renders content", func(t *testing.T) {
		var b strings.Builder
		rawLine := "Hello"
		styledLine := "Hello"

		editor.renderWithCursor(&b, rawLine, styledLine, 5)

		if got := b.String(); got != "Hello" {
			t.Errorf("renderWithCursor at end = %q, want %q", got, "Hello")
		}
	})

	t.Run("cursor screen position", func(t *testing.T) {
		editor.buf = NewGapBuffer([]byte("abcdef\nghijkl"))
		editor.nav = NewNavigation(editor.buf, 0)
		editor.scroll = 0
		editor.SetSize(80, 10)

		row, col := editor.CursorScreen()
		if row != 0 || col != editor.getLineNumWidth() {
			t.Errorf("CursorScreen() = (%d,%d), want (0,%d)", row, col, editor.getLineNumWidth())
		}

		editor.nav.SetCursor(7) // start of second line
		row, col = editor.CursorScreen()
		if row != 1 || col != editor.getLineNumWidth() {
			t.Errorf("CursorScreen() row 1 = (%d,%d), want (1,%d)", row, col, editor.getLineNumWidth())
		}
	})

	t.Run("cursor on space", func(t *testing.T) {
		var b strings.Builder
		rawLine := "Hello world"
		styledLine := "Hello world"

		editor.renderWithCursor(&b, rawLine, styledLine, 5)

		result := b.String()
		if !strings.Contains(result, "Hello") {
			t.Errorf("Expected 'Hello' before cursor, got: %q", result)
		}
		if !strings.Contains(result, "world") {
			t.Errorf("Expected 'world' after cursor, got: %q", result)
		}
	})
}
