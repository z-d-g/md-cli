package render

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// isCheckbox reports whether line is a task list item (e.g., "- [ ] " or "- [x]").
func isCheckbox(line string) bool {
	trimmed := strings.TrimSpace(line)
	if !isBulletStart(trimmed) {
		return false
	}
	rest := trimmed[1:]
	rest = strings.TrimSpace(rest)
	return strings.HasPrefix(rest, "[ ]") || strings.HasPrefix(rest, "[x]") || strings.HasPrefix(rest, "[X]")
}

func isBulletStart(line string) bool {
	return len(line) > 0 && (line[0] == '-' || line[0] == '*' || line[0] == '+')
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
	trimmed := strings.TrimSpace(line)
	rest := trimmed[1:]
	rest = strings.TrimSpace(rest)

	var checkboxChar string
	contentStart := 3

	if strings.HasPrefix(rest, "[x]") || strings.HasPrefix(rest, "[X]") {
		checkboxChar = "✓"
		if len(rest) > 3 && rest[3] == ' ' {
			contentStart = 4
		}
	} else {
		checkboxChar = "☐"
		if len(rest) > 3 && rest[3] == ' ' {
			contentStart = 4
		}
	}

	content := strings.TrimSpace(rest[contentStart:])
	elements := ParseInlineElements(content)
	styledContent := r.RenderInline(elements, lipgloss.Style{})
	return r.styles.Bullet.Render(checkboxChar+" ") + styledContent
}

// renderNumberedList renders a numbered list item with inline styling.
func renderNumberedList(line string, r *lipglossRenderer) string {
	before, after, ok := strings.Cut(line, ". ")
	if !ok {
		return line
	}

	content := strings.TrimSpace(after)
	elements := ParseInlineElements(content)
	styledContent := r.RenderInline(elements, lipgloss.Style{})
	return r.styles.Bullet.Render(before+". ") + styledContent
}

// renderBulletList renders an unordered list item with inline styling.
func renderBulletList(line string, r *lipglossRenderer) string {
	content := strings.TrimSpace(line[1:])
	elements := ParseInlineElements(content)
	styledContent := r.RenderInline(elements, lipgloss.Style{})
	return r.styles.Bullet.Render("• ") + styledContent
}
