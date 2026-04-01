package editor

import (
	"testing"
)

func TestFindSyntaxSpans(t *testing.T) {
	tests := []struct {
		name string
		line string
		want []SyntaxSpan
	}{
		{
			name: "bold",
			line: "Hello **bold** world",
			want: []SyntaxSpan{
				{Start: 6, End: 8, SpanType: SpanBold},
				{Start: 12, End: 14, SpanType: SpanBold},
			},
		},
		{
			name: "italic",
			line: "Hello *italic* world",
			want: []SyntaxSpan{
				{Start: 6, End: 7, SpanType: SpanItalic},
				{Start: 13, End: 14, SpanType: SpanItalic},
			},
		},
		{
			name: "code",
			line: "Hello `code` world",
			want: []SyntaxSpan{
				{Start: 6, End: 7, SpanType: SpanCode},
				{Start: 11, End: 12, SpanType: SpanCode},
			},
		},
		{
			name: "link",
			line: "Hello [text](url) world",
			want: []SyntaxSpan{
				{Start: 6, End: 7, SpanType: SpanLink},
				{Start: 11, End: 12, SpanType: SpanLink},
				{Start: 12, End: 13, SpanType: SpanLink},
				{Start: 16, End: 17, SpanType: SpanLink},
			},
		},
		{
			name: "heading",
			line: "## Heading",
			want: []SyntaxSpan{
				{Start: 0, End: 2, SpanType: SpanHeadingMarker},
			},
		},
		{
			name: "list",
			line: "- item",
			want: []SyntaxSpan{
				{Start: 0, End: 1, SpanType: SpanListMarker},
			},
		},
		{
			name: "HR",
			line: "---",
			want: []SyntaxSpan{
				{Start: 0, End: 3, SpanType: SpanHR},
			},
		},
		{
			name: "empty line",
			line: "",
			want: []SyntaxSpan{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindSyntaxSpans(tt.line)
			if len(got) != len(tt.want) {
				t.Errorf("FindSyntaxSpans() got %d spans, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i].Start != tt.want[i].Start || got[i].End != tt.want[i].End || got[i].SpanType != tt.want[i].SpanType {
					t.Errorf("span[%d] got %+v, want %+v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestIsCursorOnSyntax(t *testing.T) {
	line := "Hello **bold** world"
	tests := []struct {
		cursorCol int
		want      bool
	}{
		{0, false},  // H
		{5, false},  // o
		{6, true},   // * (first)
		{7, true},   // * (second)
		{8, false},  // b
		{11, false}, // d
		{12, true},  // * (first closing)
		{13, true},  // * (second closing)
		{14, false}, // space
		{20, false}, // d
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got, _ := IsCursorOnSyntax(line, tt.cursorCol)
			if got != tt.want {
				t.Errorf("IsCursorOnSyntax(%d) = %v, want %v", tt.cursorCol, got, tt.want)
			}
		})
	}
}

func TestFindBlockRegion(t *testing.T) {
	t.Run("code block", func(t *testing.T) {
		// Create test buffer with code fence
		content := "Line 1\n```\ncode block\n```\nLine 5"
		buf := NewGapBuffer([]byte(content))

		tests := []struct {
			name       string
			cursorRow  int
			wantStart  int
			wantEnd    int
			wantActive bool
		}{
			{
				name:       "before code block",
				cursorRow:  0,
				wantStart:  0,
				wantEnd:    0,
				wantActive: false,
			},
			{
				name:       "on opening fence",
				cursorRow:  1,
				wantStart:  1,
				wantEnd:    3,
				wantActive: true,
			},
			{
				name:       "inside code block",
				cursorRow:  2,
				wantStart:  1,
				wantEnd:    3,
				wantActive: true,
			},
			{
				name:       "on closing fence",
				cursorRow:  3,
				wantStart:  1,
				wantEnd:    3,
				wantActive: true,
			},
			{
				name:       "after code block",
				cursorRow:  4,
				wantStart:  4,
				wantEnd:    4,
				wantActive: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				gotStart, gotEnd, gotActive := FindBlockRegion(buf, tt.cursorRow)
				if gotStart != tt.wantStart || gotEnd != tt.wantEnd || gotActive != tt.wantActive {
					t.Errorf("FindBlockRegion() = (%d, %d, %v), want (%d, %d, %v)",
						gotStart, gotEnd, gotActive, tt.wantStart, tt.wantEnd, tt.wantActive)
				}
			})
		}
	})

	t.Run("lists", func(t *testing.T) {
		content := "Regular text\n- item 1\n- item 2\n  - subitem\n- item 3\nMore text"
		buf := NewGapBuffer([]byte(content))

		tests := []struct {
			name       string
			cursorRow  int
			wantStart  int
			wantEnd    int
			wantActive bool
		}{
			{
				name:       "before list",
				cursorRow:  0,
				wantStart:  0,
				wantEnd:    0,
				wantActive: false,
			},
			{
				name:       "first list item",
				cursorRow:  1,
				wantStart:  1,
				wantEnd:    4,
				wantActive: true,
			},
			{
				name:       "middle list item",
				cursorRow:  2,
				wantStart:  1,
				wantEnd:    4,
				wantActive: true,
			},
			{
				name:       "subitem",
				cursorRow:  3,
				wantStart:  1,
				wantEnd:    4,
				wantActive: true,
			},
			{
				name:       "last list item",
				cursorRow:  4,
				wantStart:  1,
				wantEnd:    4,
				wantActive: true,
			},
			{
				name:       "after list",
				cursorRow:  5,
				wantStart:  5,
				wantEnd:    5,
				wantActive: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				gotStart, gotEnd, gotActive := FindBlockRegion(buf, tt.cursorRow)
				if gotStart != tt.wantStart || gotEnd != tt.wantEnd || gotActive != tt.wantActive {
					t.Errorf("FindBlockRegion() = (%d, %d, %v), want (%d, %d, %v)",
						gotStart, gotEnd, gotActive, tt.wantStart, tt.wantEnd, tt.wantActive)
				}
			})
		}
	})

	t.Run("headings", func(t *testing.T) {
		content := "Regular text\n# Heading 1\n## Heading 2\nMore text"
		buf := NewGapBuffer([]byte(content))

		tests := []struct {
			name       string
			cursorRow  int
			wantStart  int
			wantEnd    int
			wantActive bool
		}{
			{
				name:       "before heading",
				cursorRow:  0,
				wantStart:  0,
				wantEnd:    0,
				wantActive: false,
			},
			{
				name:       "on heading 1",
				cursorRow:  1,
				wantStart:  1,
				wantEnd:    1,
				wantActive: true,
			},
			{
				name:       "on heading 2",
				cursorRow:  2,
				wantStart:  2,
				wantEnd:    2,
				wantActive: true,
			},
			{
				name:       "after heading",
				cursorRow:  3,
				wantStart:  3,
				wantEnd:    3,
				wantActive: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				gotStart, gotEnd, gotActive := FindBlockRegion(buf, tt.cursorRow)
				if gotStart != tt.wantStart || gotEnd != tt.wantEnd || gotActive != tt.wantActive {
					t.Errorf("FindBlockRegion() = (%d, %d, %v), want (%d, %d, %v)",
						gotStart, gotEnd, gotActive, tt.wantStart, tt.wantEnd, tt.wantActive)
				}
			})
		}
	})

	t.Run("inline syntax paragraph", func(t *testing.T) {
		content := "First paragraph\nThis has **bold** text\nThird paragraph"
		buf := NewGapBuffer([]byte(content))

		tests := []struct {
			name       string
			cursorRow  int
			wantStart  int
			wantEnd    int
			wantActive bool
		}{
			{
				name:       "before paragraph",
				cursorRow:  0,
				wantStart:  0,
				wantEnd:    0,
				wantActive: false,
			},
			{
				name:       "on line with inline syntax",
				cursorRow:  1,
				wantStart:  1,
				wantEnd:    1,
				wantActive: true,
			},
			{
				name:       "after paragraph",
				cursorRow:  2,
				wantStart:  2,
				wantEnd:    2,
				wantActive: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				gotStart, gotEnd, gotActive := FindBlockRegion(buf, tt.cursorRow)
				if gotStart != tt.wantStart || gotEnd != tt.wantEnd || gotActive != tt.wantActive {
					t.Errorf("FindBlockRegion() = (%d, %d, %v), want (%d, %d, %v)",
						gotStart, gotEnd, gotActive, tt.wantStart, tt.wantEnd, tt.wantActive)
				}
			})
		}
	})
}
