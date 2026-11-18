package editor

import (
	"strings"
	"testing"
)

func TestSkipEscape(t *testing.T) {
	tests := []struct {
		name string
		in   string
		from int
		want int
	}{
		{"csi", "\x1b[31mrest", 0, 5},
		{"osc bel", "\x1b]8;;http://x\x07rest", 0, 14},
		{"osc st", "\x1b]8;;x\x1b\\rest", 0, 8},
		{"dcs st", "\x1bP0data\x1b\\rest", 0, 9},
	}
	for _, tt := range tests {
		if got := skipEscape(tt.in, tt.from); got != tt.want {
			t.Errorf("%s: skipEscape = %d, want %d (in=%q)", tt.name, got, tt.want, tt.in)
		}
	}
}

func TestDisplayWidthAndOffset(t *testing.T) {
	if w := displayWidth("ａb"); w != 3 {
		t.Errorf("displayWidth(ａb) = %d, want 3", w)
	}
	if d := displayOffset("ａb", 1); d != 2 {
		t.Errorf("displayOffset(ａb,1) = %d, want 2", d)
	}
	if d := displayOffset("ａb", 0); d != 0 {
		t.Errorf("displayOffset(ａb,0) = %d, want 0", d)
	}
}

func TestWriteStyledRunesPreservesOsc(t *testing.T) {
	styled := "\x1b]8;;http://x\x07hi\x1b]8;;\x1b\\"
	var b strings.Builder
	writeStyledRunes(&b, styled, -1, -1, func(s string) string { return s })
	if b.String() != styled {
		t.Errorf("OSC hyperlink corrupted: got %q, want %q", b.String(), styled)
	}
}

func TestWriteStyledRunesPreservesCsi(t *testing.T) {
	styled := "\x1b[31mhi\x1b[0m"
	var b strings.Builder
	writeStyledRunes(&b, styled, -1, -1, func(s string) string { return s })
	if b.String() != styled {
		t.Errorf("CSI styling stripped: got %q, want %q", b.String(), styled)
	}
}

func TestWriteStyledRunesWideChar(t *testing.T) {
	styled := "ａb"
	var b strings.Builder
	writeStyledRunes(&b, styled, -1, -1, func(s string) string { return s })
	if b.String() != styled {
		t.Errorf("wide char mangled: got %q, want %q", b.String(), styled)
	}
}

func TestSelectionOverlayOnWideChar(t *testing.T) {
	styled := "ａb"
	selection := func(s string) string { return "{" + s + "}" }

	var b strings.Builder
	writeStyledRunes(&b, styled, 0, 3, selection)
	if b.String() != "{ａ}{b}" {
		t.Errorf("selection over 2-col-wide + 1-col: got %q, want %q", b.String(), "{ａ}{b}")
	}

	b.Reset()
	writeStyledRunes(&b, styled, 2, 3, selection)
	if b.String() != "ａ{b}" {
		t.Errorf("selection over trailing cell: got %q, want %q", b.String(), "ａ{b}")
	}
}
