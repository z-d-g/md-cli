package markdown

import "testing"

func TestParseInlineElements(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		wantLen  int
		wantType InlineType
	}{
		{"plain text", "hello", 1, InlineText},
		{"bold", "**bold**", 1, InlineBold},
		{"italic", "*italic*", 1, InlineItalic},
		{"bold italic", "***both***", 1, InlineBoldItalic},
		{"code", "`code`", 1, InlineCode},
		{"link", "[text](url)", 1, InlineLink},
		{"image", "![alt](img.png)", 1, InlineImage},
		{"strikethrough", "~~strike~~", 1, InlineStrikethrough},
		{"underline", "++under++", 1, InlineUnderline},
		{"underscore bold", "__bold__", 1, InlineBold},
		{"mixed", "a **b** c", 3, InlineText},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			elems := ParseInlineElements(tt.line)
			if len(elems) != tt.wantLen {
				t.Errorf("ParseInlineElements(%q) has %d elements, want %d",
					tt.line, len(elems), tt.wantLen)
			}
			if len(elems) > 0 && elems[0].Type != tt.wantType {
				t.Errorf("first element type = %d, want %d", elems[0].Type, tt.wantType)
			}
		})
	}
}

func TestParseInlineElementsContent(t *testing.T) {
	elems := ParseInlineElements("**hello**")
	if len(elems) != 1 {
		t.Fatalf("expected 1 element, got %d", len(elems))
	}
	if elems[0].Content != "hello" {
		t.Errorf("content = %q, want %q", elems[0].Content, "hello")
	}

	elems = ParseInlineElements("[click](https://example.com)")
	if len(elems) != 1 {
		t.Fatalf("expected 1 element, got %d", len(elems))
	}
	if elems[0].Type != InlineLink {
		t.Errorf("type = %d, want InlineLink", elems[0].Type)
	}
	if elems[0].Content != "click" {
		t.Errorf("content = %q, want %q", elems[0].Content, "click")
	}
	if elems[0].URL != "https://example.com" {
		t.Errorf("url = %q, want %q", elems[0].URL, "https://example.com")
	}
}

func TestParseInlineElementsNested(t *testing.T) {
	elems := ParseInlineElements("***bold italic***")
	if len(elems) != 1 {
		t.Fatalf("expected 1 element, got %d", len(elems))
	}
	if elems[0].Type != InlineBoldItalic {
		t.Errorf("type = %d, want InlineBoldItalic", elems[0].Type)
	}
	if len(elems[0].Children) == 0 {
		t.Error("expected nested children for bold-italic")
	}
}

func TestFindSyntaxSpans(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		minSpans int
	}{
		{"heading", "# Hello", 1},
		{"blockquote", "> quote", 1},
		{"list", "- item", 1},
		{"hr", "---", 1},
		{"bold", "**bold**", 2},
		{"plain", "hello", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spans := FindSyntaxSpans(tt.line)
			if len(spans) < tt.minSpans {
				t.Errorf("FindSyntaxSpans(%q) returned %d spans, want >= %d",
					tt.line, len(spans), tt.minSpans)
			}
		})
	}
}

func TestFindSyntaxSpansHeading(t *testing.T) {
	spans := FindSyntaxSpans("## Hello")
	if len(spans) < 1 {
		t.Fatal("expected at least 1 span")
	}
	if spans[0].SpanType != SpanHeadingMarker {
		t.Errorf("span type = %d, want SpanHeadingMarker", spans[0].SpanType)
	}
	if spans[0].End-spans[0].Start != 2 {
		t.Errorf("heading marker length = %d, want 2", spans[0].End-spans[0].Start)
	}
}

func TestFindSyntaxSpansCode(t *testing.T) {
	spans := FindSyntaxSpans("`code`")
	codeSpans := 0
	for _, s := range spans {
		if s.SpanType == SpanCode {
			codeSpans++
		}
	}
	if codeSpans != 2 {
		t.Errorf("expected 2 code spans (open+close), got %d", codeSpans)
	}
}
