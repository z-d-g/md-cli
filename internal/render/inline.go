package render

import (
	"strings"

	"github.com/z-d-g/md-cli/internal/markdown"
)

// ParseInlineElements delegates to the shared markdown package implementation.
func ParseInlineElements(line string) []InlineElement {
	return markdown.ParseInlineElements(line)
}

// RenderInline renders inline elements with styling, hiding markdown delimiters.
func (renderer *lipglossRenderer) RenderInline(elements []InlineElement, base StyleFunc) string {
	var result strings.Builder
	result.Grow(len(elements) * 10)

	for _, elem := range elements {
		switch elem.Type {
		case InlineText:
			if base != nil {
				result.WriteString(base(elem.Content))
			} else {
				result.WriteString(elem.Content)
			}
		case InlineBold:
			boldBase := base.Compose(renderer.styleCache.boldFunc)
			content := renderer.RenderInline(elem.Children, boldBase)
			result.WriteString(content)
		case InlineItalic:
			italicBase := base.Compose(renderer.styleCache.italicFunc)
			content := renderer.RenderInline(elem.Children, italicBase)
			result.WriteString(content)
		case InlineBoldItalic:
			composedBase := base.Compose(renderer.styleCache.boldFunc).Compose(renderer.styleCache.italicFunc)
			content := renderer.RenderInline(elem.Children, composedBase)
			result.WriteString(content)
		case InlineUnderline:
			underlineBase := base.Compose(renderer.styleCache.underlineFunc)
			content := renderer.RenderInline(elem.Children, underlineBase)
			result.WriteString(content)
		case InlineCode:
			codeBase := base.Compose(renderer.styleCache.codeSpanFunc)
			result.WriteString(codeBase(elem.Content))
		case InlineLink:
			content := renderer.RenderInline(ParseInlineElements(elem.Content), nil)
			result.WriteString(renderer.RenderLink(content, elem.URL))
		case InlineImage:
			result.WriteString(renderImageAlt(elem.Content, &renderer.styleCache))
		case InlineStrikethrough:
			strikeBase := base.Compose(renderer.styleCache.strikethroughFunc)
			content := renderer.RenderInline(elem.Children, strikeBase)
			result.WriteString(content)
		}
	}

	return result.String()
}

// RenderSourceInline renders inline elements showing markdown syntax with styling.
func (renderer *lipglossRenderer) RenderSourceInline(elements []InlineElement, base StyleFunc) string {
	var result strings.Builder
	result.Grow(len(elements) * 10)

	for _, elem := range elements {
		switch elem.Type {
		case InlineText:
			if base != nil {
				result.WriteString(base(elem.Content))
			} else {
				result.WriteString(elem.Content)
			}
		case InlineBold:
			delim := base("**")
			result.WriteString(delim)
			boldBase := base.Compose(renderer.styleCache.boldFunc)
			content := renderer.RenderSourceInline(elem.Children, boldBase)
			result.WriteString(content)
			result.WriteString(delim)
		case InlineItalic:
			delim := base(elem.Delimiter)
			result.WriteString(delim)
			italicBase := base.Compose(renderer.styleCache.italicFunc)
			content := renderer.RenderSourceInline(elem.Children, italicBase)
			result.WriteString(content)
			result.WriteString(delim)
		case InlineBoldItalic:
			delim := base(elem.Delimiter)
			result.WriteString(delim)
			composedBase := base.Compose(renderer.styleCache.boldFunc).Compose(renderer.styleCache.italicFunc)
			content := renderer.RenderSourceInline(elem.Children, composedBase)
			result.WriteString(content)
			result.WriteString(delim)
		case InlineUnderline:
			delim := base("__")
			result.WriteString(delim)
			underlineBase := base.Compose(renderer.styleCache.underlineFunc)
			content := renderer.RenderSourceInline(elem.Children, underlineBase)
			result.WriteString(content)
			result.WriteString(delim)
		case InlineCode:
			result.WriteString(base("`"))
			codeBase := base.Compose(renderer.styleCache.codeSpanFunc)
			result.WriteString(codeBase(elem.Content))
			result.WriteString(base("`"))
		case InlineLink:
			result.WriteString(base("["))
			linkBase := base.Compose(renderer.styleCache.linkFunc)
			content := renderer.RenderSourceInline(ParseInlineElements(elem.Content), linkBase)
			result.WriteString(content)
			result.WriteString(base("]("))
			urlBase := base.Compose(renderer.styleCache.linkURLFunc)
			result.WriteString(urlBase(elem.URL))
			result.WriteString(base(")"))
		case InlineImage:
			result.WriteString(base("!["))
			imageBase := base.Compose(renderer.styleCache.imageFunc)
			content := renderer.RenderSourceInline(ParseInlineElements(elem.Content), imageBase)
			result.WriteString(content)
			result.WriteString(base("]("))
			urlBase := base.Compose(renderer.styleCache.linkURLFunc)
			result.WriteString(urlBase(elem.URL))
			result.WriteString(base(")"))
		case InlineStrikethrough:
			delim := base("~~")
			result.WriteString(delim)
			strikeBase := base.Compose(renderer.styleCache.strikethroughFunc)
			content := renderer.RenderSourceInline(elem.Children, strikeBase)
			result.WriteString(content)
			result.WriteString(delim)
		}
	}

	return result.String()
}
