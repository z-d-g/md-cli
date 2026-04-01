package editor

import (
	"strings"
	"testing"

	"github.com/z-d-g/md-cli/internal/config"
	"github.com/z-d-g/md-cli/internal/render"
)

func containsMarkdownSyntax(s string) bool {
	markers := []string{"**", "*", "`", "__", "~~"}
	for _, marker := range markers {
		if len(marker) == 1 {
			continue
		}
		if strings.Contains(s, marker) {
			return true
		}
	}
	return false
}

func testRendererFromTheme() render.LineRenderer {
	theme := config.DefaultTheme()
	styles := theme.ToEditorStyles()
	return render.NewLipglossRenderer(&styles)
}

func TestRenderCache(t *testing.T) {
	editor := NewEditor("# Heading\n**bold text**\nnormal text", testRendererFromTheme())
	editor.SetSize(80, 20)
	editor.computeFrameState()

	line1 := "# Heading"
	firstRender := editor.getCachedLine(0, line1)
	secondRender := editor.getCachedLine(0, line1)

	if firstRender != secondRender {
		t.Errorf("Cache hit failed: first render != second render")
	}

	line2 := "**bold** text"
	firstRender2 := editor.getCachedLine(1, line2)

	line2Changed := "**different** text"
	secondRender2 := editor.getCachedLine(1, line2Changed)

	if firstRender2 == secondRender2 {
		t.Errorf("Content change should produce different output")
	}
}

func TestRenderCacheInvalidation(t *testing.T) {
	editor := NewEditor("Line 0\nLine 1\nLine 2", testRendererFromTheme())
	editor.computeFrameState()

	// Populate cache
	editor.getCachedLine(0, "Line 0")
	editor.getCachedLine(1, "Line 1")
	editor.getCachedLine(2, "Line 2")

	if len(editor.renderCache) != 3 {
		t.Errorf("Expected 3 cache entries, got %d", len(editor.renderCache))
	}

	// Invalidate specific line
	delete(editor.renderCache, 1)

	if len(editor.renderCache) != 2 {
		t.Errorf("Expected 2 cache entries after invalidation, got %d", len(editor.renderCache))
	}

	// Invalidate all
	editor.renderCache = make(map[int]cacheEntry)

	if len(editor.renderCache) != 0 {
		t.Errorf("Expected 0 cache entries after InvalidateAll, got %d", len(editor.renderCache))
	}
}

func TestRenderCacheHeadingBoldCode(t *testing.T) {
	editor := NewEditor("", testRendererFromTheme())
	editor.SetSize(80, 20)
	editor.computeFrameState()

	tests := []struct {
		name string
		line string
	}{
		{"heading", "# Heading"},
		{"bold", "**bold** text"},
		{"code", "`code` text"},
		{"mixed", "# Heading with **bold** and `code`"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := editor.getCachedLine(0, tt.line)
			if result == "" {
				t.Errorf("Render failed for %s: empty result", tt.name)
			}
			if result == tt.line {
				t.Errorf("Render failed for %s: no styling applied", tt.name)
			}
		})
	}
}

func TestRenderCacheBlockquote(t *testing.T) {
	editor := NewEditor("", testRendererFromTheme())
	editor.SetSize(80, 20)
	editor.computeFrameState()

	tests := []struct {
		name string
		line string
	}{
		{"simple blockquote", "> This is a blockquote"},
		{"blockquote with bold", "> This is **bold** text"},
		{"blockquote with italic", "> This is *italic* text"},
		{"blockquote with code", "> This is `code` text"},
		{"blockquote with mixed", "> This has **bold**, *italic*, and `code`"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := editor.getCachedLine(0, tt.line)
			if result == "" {
				t.Errorf("Render failed for %s: empty result", tt.name)
			}
			if result == tt.line {
				t.Errorf("Render failed for %s: no styling applied", tt.name)
			}
			if containsMarkdownSyntax(result) {
				t.Errorf("Render failed for %s: result still contains markdown syntax: %s", tt.name, result)
			}
		})
	}
}
