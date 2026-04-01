package render

import (
	"strings"
)

// isCheckbox reports whether line is a task list item (e.g., "- [ ] " or "- [x]").
func isCheckbox(line string) bool {
	return strings.HasPrefix(line, "- [ ]") || strings.HasPrefix(line, "- [x]") ||
		strings.HasPrefix(line, "- [X]") ||
		strings.HasPrefix(line, "* [ ]") || strings.HasPrefix(line, "* [x]") ||
		strings.HasPrefix(line, "* [X]") ||
		strings.HasPrefix(line, "+ [ ]") || strings.HasPrefix(line, "+ [x]") ||
		strings.HasPrefix(line, "+ [X]")
}

// isNumberedList reports whether line starts a numbered list item (e.g., "1. ").
func isNumberedList(line string) bool {
	if len(line) >= 3 && line[1] == '.' && line[2] == ' ' {
		if line[0] >= '0' && line[0] <= '9' {
			return true
		}
	}
	if len(line) >= 4 && line[2] == '.' && line[3] == ' ' {
		if line[0] >= '0' && line[0] <= '9' && line[1] >= '0' && line[1] <= '9' {
			return true
		}
	}
	return false
}

// renderCheckbox renders a task list item with checkmark styling.
func renderCheckbox(line string, r *lipglossRenderer) string {
	var checkboxChar string
	var content string

	if strings.HasPrefix(line, "- [x]") || strings.HasPrefix(line, "- [X]") ||
		strings.HasPrefix(line, "* [x]") || strings.HasPrefix(line, "* [X]") ||
		strings.HasPrefix(line, "+ [x]") || strings.HasPrefix(line, "+ [X]") {
		checkboxChar = "✓"
		content = strings.TrimSpace(line[5:])
	} else {
		checkboxChar = "☐"
		content = strings.TrimSpace(line[5:])
	}

	elements := ParseInlineElements(content)
	styledContent := r.RenderInline(elements, nil)
	return r.styleCache.bulletFunc(checkboxChar+" ") + styledContent
}

// renderNumberedList renders a numbered list item with inline styling.
func renderNumberedList(line string, r *lipglossRenderer) string {
	before, after, ok := strings.Cut(line, ". ")
	if !ok {
		return line
	}

	content := strings.TrimSpace(after)
	elements := ParseInlineElements(content)
	styledContent := r.RenderInline(elements, nil)
	return r.styleCache.bulletFunc(before+". ") + styledContent
}

// renderBulletList renders an unordered list item with inline styling.
func renderBulletList(line string, r *lipglossRenderer) string {
	content := strings.TrimSpace(line[1:])
	elements := ParseInlineElements(content)
	styledContent := r.RenderInline(elements, nil)
	return r.styleCache.bulletFunc("• ") + styledContent
}
