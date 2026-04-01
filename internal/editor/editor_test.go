package editor

import (
	"strings"
	"testing"

	"github.com/z-d-g/md-cli/internal/config"
	"github.com/z-d-g/md-cli/internal/render"

	tea "charm.land/bubbletea/v2"
)

func testRenderer() render.LineRenderer {
	theme := config.DefaultTheme()
	styles := theme.ToEditorStyles()
	return render.NewLipglossRenderer(&styles)
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
	_, cmd := editor.Update(msg)
	if cmd != nil {
		t.Fatalf("expected no command, got %v", cmd)
	}

	row, col = editor.Cursor()
	if row != 0 || col != 1 {
		t.Fatalf("expected cursor at (0, 1) after right arrow, got (%d, %d)", row, col)
	}

	// Test down arrow key
	msg = tea.KeyPressMsg{Text: "down"}
	_, cmd = editor.Update(msg)
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

	t.Run("cursor at end", func(t *testing.T) {
		var b strings.Builder
		rawLine := "Hello"
		styledLine := "Hello"

		editor.renderWithCursor(&b, rawLine, styledLine, 5)

		result := b.String()
		if len(result) <= len("Hello") {
			t.Errorf("Expected cursor styling at end, got: %q", result)
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
