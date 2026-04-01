package markdown

import (
	"strings"
	"unicode/utf8"
)

// ParseInlineElements parses a markdown line into inline elements.
func ParseInlineElements(line string) []InlineElement {
	var elements []InlineElement
	i := 0

	for i < len(line) {
		// Bold+Italic: ***text*** or ___text___
		if i+3 <= len(line) && (line[i:i+3] == "***" || line[i:i+3] == "___") {
			end := FindClosingDelimiter(line, i+3, line[i:i+3])
			if end != -1 {
				content := line[i+3 : end]
				nestedElements := ParseInlineElements(content)
				elements = append(elements, InlineElement{
					Type:      InlineBoldItalic,
					Content:   content,
					Delimiter: line[i : i+3],
					Children:  nestedElements,
				})
				i = end + 3
				continue
			}
		}

		// Bold: **text**
		if i+2 <= len(line) && line[i:i+2] == "**" {
			end := FindClosingDelimiter(line, i+2, "**")
			if end != -1 {
				content := line[i+2 : end]
				nestedElements := ParseInlineElements(content)
				elements = append(elements, InlineElement{
					Type:      InlineBold,
					Content:   content,
					Delimiter: "**",
					Children:  nestedElements,
				})
				i = end + 2
				continue
			}
		}

		// Underline: __text__
		if i+2 <= len(line) && line[i:i+2] == "__" {
			end := FindClosingDelimiter(line, i+2, "__")
			if end != -1 {
				content := line[i+2 : end]
				nestedElements := ParseInlineElements(content)
				elements = append(elements, InlineElement{
					Type:      InlineUnderline,
					Content:   content,
					Delimiter: "__",
					Children:  nestedElements,
				})
				i = end + 2
				continue
			}
		}

		// Italic: *text* or _text_
		if i+1 <= len(line) && (line[i] == '*' || line[i] == '_') {
			end := FindClosingDelimiter(line, i+1, string(line[i]))
			if end != -1 {
				content := line[i+1 : end]
				nestedElements := ParseInlineElements(content)
				elements = append(elements, InlineElement{
					Type:      InlineItalic,
					Content:   content,
					Delimiter: string(line[i]),
					Children:  nestedElements,
				})
				i = end + 1
				continue
			}
		}

		// Strikethrough: ~~text~~
		if i+2 <= len(line) && line[i:i+2] == "~~" {
			end := FindClosingDelimiter(line, i+2, "~~")
			if end != -1 {
				content := line[i+2 : end]
				nestedElements := ParseInlineElements(content)
				elements = append(elements, InlineElement{
					Type:      InlineStrikethrough,
					Content:   content,
					Delimiter: "~~",
					Children:  nestedElements,
				})
				i = end + 2
				continue
			}
		}

		// Code spans: `code`
		if i+1 <= len(line) && line[i] == '`' {
			end := FindClosingDelimiter(line, i+1, "`")
			if end != -1 {
				content := line[i+1 : end]
				elements = append(elements, InlineElement{
					Type:      InlineCode,
					Content:   content,
					Delimiter: "`",
				})
				i = end + 1
				continue
			}
		}

		// Links: [text](url)
		if i+1 <= len(line) && line[i] == '[' {
			textEnd := FindClosingDelimiter(line, i+1, "]")
			if textEnd != -1 && textEnd+1 < len(line) && line[textEnd+1] == '(' {
				urlEnd := FindClosingDelimiter(line, textEnd+2, ")")
				if urlEnd != -1 {
					linkText := line[i+1 : textEnd]
					linkURL := line[textEnd+2 : urlEnd]
					elements = append(elements, InlineElement{
						Type:    InlineLink,
						Content: linkText,
						URL:     linkURL,
					})
					i = urlEnd + 1
					continue
				}
			}
		}

		// Images: ![alt](url)
		if i+2 <= len(line) && line[i:i+2] == "![" {
			altEnd := FindClosingDelimiter(line, i+2, "]")
			if altEnd != -1 && altEnd+1 < len(line) && line[altEnd+1] == '(' {
				urlEnd := FindClosingDelimiter(line, altEnd+2, ")")
				if urlEnd != -1 {
					altText := line[i+2 : altEnd]
					imageURL := line[altEnd+2 : urlEnd]
					elements = append(elements, InlineElement{
						Type:    InlineImage,
						Content: altText,
						URL:     imageURL,
					})
					i = urlEnd + 1
					continue
				}
			}
		}

		// Regular character — handle UTF-8 properly
		if i < len(line) {
			r, size := utf8.DecodeRuneInString(line[i:])
			if r == utf8.RuneError && size == 1 {
				elements = append(elements, InlineElement{
					Type:    InlineText,
					Content: string(line[i]),
				})
				i++
			} else {
				elements = append(elements, InlineElement{
					Type:    InlineText,
					Content: string(r),
				})
				i += size
			}
		}
	}

	return elements
}

// FindSyntaxSpans converts inline elements to syntax spans for source-mode rendering.
func FindSyntaxSpans(line string) []SyntaxSpan {
	var spans []SyntaxSpan

	// Block-level markers first
	if has, sp := findHeadingMarker(line); has {
		spans = append(spans, sp)
	}
	if has, sp := findBlockquoteMarker(line); has {
		spans = append(spans, sp)
	}
	if has, sp := findListMarker(line); has {
		spans = append(spans, sp)
	}
	if IsHorizontalRule(line) {
		spans = append(spans, SyntaxSpan{Start: 0, End: len(line), SpanType: SpanHR})
	}

	// Inline spans from element parsing
	spans = append(spans, inlineToSpans(line)...)
	return spans
}

func findHeadingMarker(line string) (bool, SyntaxSpan) {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "#") {
		return false, SyntaxSpan{}
	}
	hashCount := 0
	for i := 0; i < len(trimmed) && i < 6; i++ {
		if trimmed[i] == '#' {
			hashCount++
		} else {
			break
		}
	}
	if hashCount > 0 && hashCount < len(trimmed) && trimmed[hashCount] == ' ' {
		idx := strings.Index(line, "#")
		return true, SyntaxSpan{
			Start:    idx,
			End:      idx + hashCount,
			SpanType: SpanHeadingMarker,
		}
	}
	return false, SyntaxSpan{}
}

func findBlockquoteMarker(line string) (bool, SyntaxSpan) {
	trimmed := strings.TrimSpace(line)
	rest := trimmed
	depth := 0
	for strings.HasPrefix(rest, ">") {
		depth++
		rest = strings.TrimPrefix(rest, ">")
		rest = strings.TrimPrefix(rest, " ")
	}
	if depth == 0 {
		return false, SyntaxSpan{}
	}
	idx := strings.Index(line, ">")
	// Find byte end of all leading > and spaces
	end := idx
	scan := trimmed
	for strings.HasPrefix(scan, ">") {
		end++
		scan = strings.TrimPrefix(scan, ">")
		if strings.HasPrefix(scan, " ") {
			end++
			scan = strings.TrimPrefix(scan, " ")
		}
	}
	return true, SyntaxSpan{
		Start:    idx,
		End:      end,
		SpanType: SpanBlockquoteMarker,
	}
}

func findListMarker(line string) (bool, SyntaxSpan) {
	if !IsListLine(line) {
		return false, SyntaxSpan{}
	}
	trimmed := strings.TrimSpace(line)
	marker := trimmed[0]
	idx := strings.IndexByte(line, marker)
	if idx >= 0 {
		return true, SyntaxSpan{
			Start:    idx,
			End:      idx + 1,
			SpanType: SpanListMarker,
		}
	}
	return false, SyntaxSpan{}
}

func inlineToSpans(line string) []SyntaxSpan {
	var spans []SyntaxSpan
	elements := ParseInlineElements(line)
	offset := 0
	collectSpans(elements, line, &offset, &spans)
	return spans
}

// collectSpans walks elements and emits syntax spans at delimiter boundaries.
func collectSpans(elems []InlineElement, line string, offset *int, spans *[]SyntaxSpan) {
	for _, elem := range elems {
		switch elem.Type {
		case InlineText:
			*offset += len(elem.Content)
		case InlineBold, InlineBoldItalic:
			dLen := len(elem.Delimiter)
			*spans = append(*spans,
				SyntaxSpan{Start: *offset, End: *offset + dLen, SpanType: SpanBold},
			)
			*offset += dLen
			collectSpans(elem.Children, line, offset, spans)
			*spans = append(*spans,
				SyntaxSpan{Start: *offset, End: *offset + dLen, SpanType: SpanBold},
			)
			*offset += dLen
		case InlineItalic:
			dLen := len(elem.Delimiter)
			*spans = append(*spans,
				SyntaxSpan{Start: *offset, End: *offset + dLen, SpanType: SpanItalic},
			)
			*offset += dLen
			collectSpans(elem.Children, line, offset, spans)
			*spans = append(*spans,
				SyntaxSpan{Start: *offset, End: *offset + dLen, SpanType: SpanItalic},
			)
			*offset += dLen
		case InlineUnderline:
			dLen := len(elem.Delimiter)
			*spans = append(*spans,
				SyntaxSpan{Start: *offset, End: *offset + dLen, SpanType: SpanBold},
			)
			*offset += dLen
			collectSpans(elem.Children, line, offset, spans)
			*spans = append(*spans,
				SyntaxSpan{Start: *offset, End: *offset + dLen, SpanType: SpanBold},
			)
			*offset += dLen
		case InlineCode:
			*spans = append(*spans,
				SyntaxSpan{Start: *offset, End: *offset + 1, SpanType: SpanCode},
			)
			*offset += 1 + len(elem.Content)
			*spans = append(*spans,
				SyntaxSpan{Start: *offset, End: *offset + 1, SpanType: SpanCode},
			)
			*offset += 1
		case InlineLink:
			*spans = append(*spans,
				SyntaxSpan{Start: *offset, End: *offset + 1, SpanType: SpanLink},
			)
			*offset += 1 + len(elem.Content)
			*spans = append(*spans,
				SyntaxSpan{Start: *offset, End: *offset + 1, SpanType: SpanLink},
			)
			*offset++ // ]
			*spans = append(*spans,
				SyntaxSpan{Start: *offset, End: *offset + 1, SpanType: SpanLink},
			)
			*offset++ // (
			*offset += len(elem.URL)
			*spans = append(*spans,
				SyntaxSpan{Start: *offset, End: *offset + 1, SpanType: SpanLink},
			)
			*offset++ // )
		case InlineImage:
			*spans = append(*spans,
				SyntaxSpan{Start: *offset, End: *offset + 2, SpanType: SpanImage},
			)
			*offset += 2 + len(elem.Content)
			*spans = append(*spans,
				SyntaxSpan{Start: *offset, End: *offset + 1, SpanType: SpanImage},
			)
			*offset++ // ]
			*spans = append(*spans,
				SyntaxSpan{Start: *offset, End: *offset + 1, SpanType: SpanImage},
			)
			*offset++ // (
			*offset += len(elem.URL)
			*spans = append(*spans,
				SyntaxSpan{Start: *offset, End: *offset + 1, SpanType: SpanImage},
			)
			*offset++ // )
		case InlineStrikethrough:
			dLen := len(elem.Delimiter)
			*spans = append(*spans,
				SyntaxSpan{Start: *offset, End: *offset + dLen, SpanType: SpanStrikethrough},
			)
			*offset += dLen
			collectSpans(elem.Children, line, offset, spans)
			*spans = append(*spans,
				SyntaxSpan{Start: *offset, End: *offset + dLen, SpanType: SpanStrikethrough},
			)
			*offset += dLen
		}
	}
}
