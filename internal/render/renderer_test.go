package render

import (
	"strings"
	"testing"

	"github.com/z-d-g/md-cli/internal/config"

	"charm.land/lipgloss/v2"
)

func testRenderer() LineRenderer {
	styles := config.DefaultTheme().ToEditorStyles()
	return NewLipglossRenderer(&styles, 80)
}

func TestRenderLineHeading(t *testing.T) {
	r := testRenderer()
	tests := []struct {
		line string
	}{
		{"# H1"},
		{"## H2"},
		{"### H3"},
		{"#### H4"},
		{"##### H5"},
		{"###### H6"},
		{"  # one-space-indented"},
		{"  ## two-space-indented"},
		{"   ### three-space-indented"},
	}
	for _, tt := range tests {
		result := r.RenderLine(tt.line, false)
		if result == "" {
			t.Errorf("RenderLine(%q) returned empty", tt.line)
		}
		if strings.Contains(result, "# ") {
			t.Errorf("RenderLine(%q) leaked heading markers: %q", tt.line, result)
		}
	}
}

func TestRenderLineCodeFence(t *testing.T) {
	r := testRenderer()
	result := r.RenderLine("```go", false)
	if result == "" {
		t.Error("RenderLine(code fence) returned empty")
	}

	result = r.RenderLine("code here", true)
	if result == "" {
		t.Error("RenderLine(code content) returned empty")
	}
}

func TestRenderLineBlockquote(t *testing.T) {
	r := testRenderer()
	result := r.RenderLine("> quoted text", false)
	if result == "" {
		t.Error("RenderLine(blockquote) returned empty")
	}
	if !strings.Contains(result, "quoted text") {
		t.Errorf("RenderLine(blockquote) missing content: %q", result)
	}
}

func TestRenderLineHorizontalRule(t *testing.T) {
	r := testRenderer()
	result := r.RenderLine("---", false)
	if result == "" {
		t.Error("RenderLine(hr) returned empty")
	}
}

func TestRenderLineList(t *testing.T) {
	r := testRenderer()
	tests := []struct {
		line string
	}{
		{"- bullet item"},
		{"* bullet item"},
		{"+ bullet item"},
		{"1. numbered"},
		{"42. numbered"},
		{"- [ ] unchecked"},
		{"- [x] checked"},
	}
	for _, tt := range tests {
		result := r.RenderLine(tt.line, false)
		if result == "" {
			t.Errorf("RenderLine(%q) returned empty", tt.line)
		}
	}
}

func TestRenderLineTable(t *testing.T) {
	r := testRenderer()
	r.SetDocument(func() []string {
		return []string{"| A | B |", "|---|---|", "| 1 | 2 |"}
	})
	defer r.SetDocument(nil)

	result := r.RenderLine("| A | B |", false)
	if result == "" {
		t.Error("RenderLine(table header) returned empty")
	}
}

func TestRenderLineNormal(t *testing.T) {
	r := testRenderer()
	result := r.RenderLine("plain text", false)
	if result != "plain text" {
		t.Errorf("RenderLine(plain) = %q, want %q", result, "plain text")
	}
}

func TestRenderInline(t *testing.T) {
	r := testRenderer()
	empty := lipgloss.Style{}

	elements := ParseInlineElements("**bold**")
	result := r.RenderInline(elements, empty)
	if result == "" {
		t.Error("RenderInline(bold) returned empty")
	}

	elements = ParseInlineElements("*italic*")
	result = r.RenderInline(elements, empty)
	if result == "" {
		t.Error("RenderInline(italic) returned empty")
	}

	elements = ParseInlineElements("`code`")
	result = r.RenderInline(elements, empty)
	if result == "" {
		t.Error("RenderInline(code) returned empty")
	}

	elements = ParseInlineElements("[link](url)")
	result = r.RenderInline(elements, empty)
	if result == "" {
		t.Error("RenderInline(link) returned empty")
	}
}

func TestRenderSourceInline(t *testing.T) {
	r := testRenderer()
	empty := lipgloss.Style{}

	elements := ParseInlineElements("**bold**")
	result := r.RenderSourceInline(elements, empty)
	if !strings.Contains(result, "**") {
		t.Errorf("RenderSourceInline(bold) missing delimiters: %q", result)
	}

	elements = ParseInlineElements("*italic*")
	result = r.RenderSourceInline(elements, empty)
	if !strings.Contains(result, "*") {
		t.Errorf("RenderSourceInline(italic) missing delimiters: %q", result)
	}
}

func TestRenderLineNumber(t *testing.T) {
	r := testRenderer()
	result := r.RenderLineNumber(" 42")
	if result == "" {
		t.Error("RenderLineNumber returned empty")
	}
}

func TestRenderCursorChar(t *testing.T) {
	r := testRenderer()
	result := r.RenderCursorChar("A")
	if result == "" {
		t.Error("RenderCursorChar returned empty")
	}
}

func TestRenderSelectionChar(t *testing.T) {
	r := testRenderer()
	result := r.RenderSelectionChar("A")
	if result == "" {
		t.Error("RenderSelectionChar returned empty")
	}
}

func TestRenderStyled(t *testing.T) {
	r := testRenderer()
	tests := []struct {
		name     string
		lineType int
	}{
		{"H1", LineHeading1},
		{"H2", LineHeading2},
		{"H3", LineHeading3},
		{"H4", LineHeading4},
		{"H5", LineHeading5},
		{"H6", LineHeading6},
		{"CodeFence", LineCodeFence},
		{"CodeContent", LineCodeContent},
		{"BlockQuote", LineBlockQuote},
		{"Normal", LineNormal},
	}
	for _, tt := range tests {
		result := r.RenderStyled("text", tt.lineType)
		if result == "" {
			t.Errorf("RenderStyled(%s) returned empty", tt.name)
		}
	}
}

func TestRenderInlineImage(t *testing.T) {
	r := testRenderer()
	empty := lipgloss.Style{}

	elements := ParseInlineElements("![alt text](image.png)")
	result := r.RenderInline(elements, empty)
	if result == "" {
		t.Error("RenderInline(image) returned empty")
	}

	elements = ParseInlineElements("![](image.png)")
	result = r.RenderInline(elements, empty)
	if result == "" {
		t.Error("RenderInline(image no alt) returned empty")
	}
}

func TestRenderInlineStrikethrough(t *testing.T) {
	r := testRenderer()
	empty := lipgloss.Style{}

	elements := ParseInlineElements("~~strike~~")
	result := r.RenderInline(elements, empty)
	if result == "" {
		t.Error("RenderInline(strikethrough) returned empty")
	}
}

func TestRenderInlineUnderline(t *testing.T) {
	r := testRenderer()
	empty := lipgloss.Style{}

	elements := ParseInlineElements("++under++")
	result := r.RenderInline(elements, empty)
	if result == "" {
		t.Error("RenderInline(underline) returned empty")
	}
	if !strings.Contains(result, "u") || !strings.Contains(result, "n") {
		t.Errorf("RenderInline(underline) missing content chars: %q", result)
	}
}

func TestRenderTableFull(t *testing.T) {
	r := testRenderer()
	lines := []string{"| H1 | H2 |", "|---|---|", "| R1 | R2 |"}
	r.SetDocument(func() []string { return lines })
	defer r.SetDocument(nil)

	renderer := r.(*lipglossRenderer)

	header := r.RenderLine(lines[0], false)
	if header == "" {
		t.Error("table header empty")
	}

	sep := r.RenderLine(lines[1], false)
	if sep == "" {
		t.Error("table separator empty")
	}

	row := r.RenderLine(lines[2], false)
	if row == "" {
		t.Error("table row empty")
	}

	renderer.ResetTable()
	if renderer.table.active {
		t.Error("ResetTable should deactivate table")
	}
}

func TestTableVersion(t *testing.T) {
	r := testRenderer()
	if v := r.TableVersion(); v != 0 {
		t.Errorf("TableVersion() = %d, want 0", v)
	}
}

func TestParseTableCells(t *testing.T) {
	cells := parseTableCells("| a | b | c |")
	if len(cells) != 3 {
		t.Fatalf("expected 3 cells, got %d: %v", len(cells), cells)
	}
	if cells[0] != "a" || cells[1] != "b" || cells[2] != "c" {
		t.Errorf("cells = %v, want [a b c]", cells)
	}
}

func TestParseSeparator(t *testing.T) {
	tests := []struct {
		cells []string
		want  []tableAlign
	}{
		{[]string{"---", "---"}, []tableAlign{alignLeft, alignLeft}},
		{[]string{":---:", ":---"}, []tableAlign{alignCenter, alignLeft}},
		{[]string{"---:"}, []tableAlign{alignRight}},
	}
	for _, tt := range tests {
		got, _ := parseSeparator(tt.cells)
		if len(got) != len(tt.want) {
			t.Errorf("parseSeparator(%v) length = %d, want %d", tt.cells, len(got), len(tt.want))
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("parseSeparator(%v)[%d] = %d, want %d", tt.cells, i, got[i], tt.want[i])
			}
		}
	}
}

func TestAlignText(t *testing.T) {
	tests := []struct {
		text  string
		width int
		align tableAlign
		want  string
	}{
		{"ab", 5, alignLeft, "ab   "},
		{"ab", 5, alignRight, "   ab"},
		{"ab", 5, alignCenter, " ab  "},
		{"abcde", 3, alignLeft, "abcde"},
	}
	for _, tt := range tests {
		got := alignText(tt.text, tt.width, tt.align)
		if got != tt.want {
			t.Errorf("alignText(%q, %d, %d) = %q, want %q", tt.text, tt.width, tt.align, got, tt.want)
		}
	}
}

func TestPrintRenderer(t *testing.T) {
	r := testRenderer()
	p := NewPrintRenderer(r)

	content := "# Hello\n\nSome **bold** text.\n\n```go\nfmt.Println()\n```\n"
	result := p.RenderDocument(content)
	if result == "" {
		t.Error("RenderDocument returned empty")
	}
	if !strings.Contains(result, "Hello") {
		t.Error("RenderDocument missing heading content")
	}
}
