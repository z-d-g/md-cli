package render

import (
	"strings"

	"github.com/z-d-g/md-cli/v2/internal/markdown"

	"charm.land/lipgloss/v2"
)

// ParseInlineElements delegates to the shared markdown package implementation.
func ParseInlineElements(line string) []InlineElement {
	return markdown.ParseInlineElements(line)
}

// RenderInline renders inline elements with styling, hiding markdown delimiters.
// Styles compose via Inherit — one render call, clean ANSI output.
func (renderer *lipglossRenderer) RenderInline(elements []InlineElement, base lipgloss.Style) string {
	var result strings.Builder
	result.Grow(len(elements) * 10)

	for _, elem := range elements {
		switch elem.Type {
		case InlineText:
			result.WriteString(base.Render(elem.Content))
		case InlineBold:
			content := renderer.RenderInline(elem.Children, base.Inherit(renderer.styles.Bold))
			result.WriteString(content)
		case InlineItalic:
			content := renderer.RenderInline(elem.Children, base.Inherit(renderer.styles.Italic))
			result.WriteString(content)
		case InlineBoldItalic:
			biBase := base.Inherit(renderer.styles.Bold).Inherit(renderer.styles.Italic)
			content := renderer.RenderInline(elem.Children, biBase)
			result.WriteString(content)
		case InlineUnderline:
			content := renderer.RenderInline(elem.Children, base.Inherit(renderer.styles.Underline))
			result.WriteString(content)
		case InlineCode:
			codeBase := base.Inherit(renderer.styles.CodeSpan)
			result.WriteString(codeBase.Render(elem.Content))
		case InlineLink:
			content := renderer.RenderInline(ParseInlineElements(elem.Content), lipgloss.Style{})
			result.WriteString(renderer.RenderLink(content, elem.URL))
		case InlineImage:
			result.WriteString(renderImageAlt(elem.Content, renderer.styles))
		case InlineStrikethrough:
			content := renderer.RenderInline(elem.Children, base.Inherit(renderer.styles.Strikethrough))
			result.WriteString(content)
		}
	}

	return result.String()
}

// RenderSourceInline renders inline elements showing markdown syntax with styling.
func (renderer *lipglossRenderer) RenderSourceInline(elements []InlineElement, base lipgloss.Style) string {
	var result strings.Builder
	result.Grow(len(elements) * 10)

	for _, elem := range elements {
		switch elem.Type {
		case InlineText:
			result.WriteString(base.Render(elem.Content))
		case InlineBold:
			delim := base.Render(elem.Delimiter)
			result.WriteString(delim)
			content := renderer.RenderSourceInline(elem.Children, base.Inherit(renderer.styles.Bold))
			result.WriteString(content)
			result.WriteString(delim)
		case InlineItalic:
			delim := base.Render(elem.Delimiter)
			result.WriteString(delim)
			content := renderer.RenderSourceInline(elem.Children, base.Inherit(renderer.styles.Italic))
			result.WriteString(content)
			result.WriteString(delim)
		case InlineBoldItalic:
			delim := base.Render(elem.Delimiter)
			result.WriteString(delim)
			biBase := base.Inherit(renderer.styles.Bold).Inherit(renderer.styles.Italic)
			content := renderer.RenderSourceInline(elem.Children, biBase)
			result.WriteString(content)
			result.WriteString(delim)
		case InlineUnderline:
			delim := base.Render(elem.Delimiter)
			result.WriteString(delim)
			content := renderer.RenderSourceInline(elem.Children, base.Inherit(renderer.styles.Underline))
			result.WriteString(content)
			result.WriteString(delim)
		case InlineCode:
			result.WriteString(base.Render("`"))
			codeBase := base.Inherit(renderer.styles.CodeSpan)
			result.WriteString(codeBase.Render(elem.Content))
			result.WriteString(base.Render("`"))
		case InlineLink:
			result.WriteString(base.Render("["))
			linkBase := base.Inherit(renderer.styles.Link)
			content := renderer.RenderSourceInline(ParseInlineElements(elem.Content), linkBase)
			result.WriteString(content)
			result.WriteString(base.Render("]("))
			urlBase := base.Inherit(renderer.styles.LinkURL)
			result.WriteString(urlBase.Render(elem.URL))
			result.WriteString(base.Render(")"))
		case InlineImage:
			result.WriteString(base.Render("!["))
			imageBase := base.Inherit(renderer.styles.Image)
			content := renderer.RenderSourceInline(ParseInlineElements(elem.Content), imageBase)
			result.WriteString(content)
			result.WriteString(base.Render("]("))
			urlBase := base.Inherit(renderer.styles.LinkURL)
			result.WriteString(urlBase.Render(elem.URL))
			result.WriteString(base.Render(")"))
		case InlineStrikethrough:
			delim := base.Render(elem.Delimiter)
			result.WriteString(delim)
			content := renderer.RenderSourceInline(elem.Children, base.Inherit(renderer.styles.Strikethrough))
			result.WriteString(content)
			result.WriteString(delim)
		}
	}

	return result.String()
}
