package markdown

import (
	"strings"
	"testing"
)

func TestCodeBlockLinesBasic(t *testing.T) {
	lines := strings.Split("Line 1\n```\ncode\n```\nLine 5", "\n")
	tr := NewCodeFenceTracker(lines)
	code := tr.CodeBlockLines()
	want := []bool{false, false, true, false, false}
	if len(code) != len(want) {
		t.Fatalf("expected %d lines, got %d", len(want), len(code))
	}
	for i := range want {
		if code[i] != want[i] {
			t.Errorf("code[%d] = %v, want %v", i, code[i], want[i])
		}
	}
}

func TestCodeBlockLinesMismatchedCharStaysOpen(t *testing.T) {
	lines := strings.Split("```\n~~~\n```\n", "\n")
	tr := NewCodeFenceTracker(lines)
	code := tr.CodeBlockLines()
	if !code[1] {
		t.Errorf("expected ~~~ line inside ``` block to stay code, got %v", code)
	}
}

func TestCodeBlockBounds(t *testing.T) {
	lines := strings.Split("Line 1\n```\ncode block\n```\nLine 5", "\n")
	tr := NewCodeFenceTracker(lines)
	tests := []struct {
		row    int
		wantS  int
		wantE  int
		wantIn bool
	}{
		{0, 0, 0, false},
		{1, 1, 3, false},
		{2, 1, 3, true},
		{3, 1, 3, false},
		{4, 4, 4, false},
	}
	for _, tt := range tests {
		gotS, gotE := tr.CodeBlockBounds(tt.row)
		gotIn := tr.IsInside(tt.row)
		if gotS != tt.wantS || gotE != tt.wantE || gotIn != tt.wantIn {
			t.Errorf("row %d: bounds=(%d,%d) inside=%v, want (%d,%d) inside=%v",
				tt.row, gotS, gotE, gotIn, tt.wantS, tt.wantE, tt.wantIn)
		}
	}
}

func TestCodeBlockBoundsUnclosed(t *testing.T) {
	lines := strings.Split("```\ncode\n", "\n")
	tr := NewCodeFenceTracker(lines)
	s, e := tr.CodeBlockBounds(1)
	if s != 0 || e != 2 {
		t.Errorf("unclosed block bounds = (%d,%d), want (0,2)", s, e)
	}
}
