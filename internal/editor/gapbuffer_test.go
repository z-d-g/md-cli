package editor

import "testing"

func TestDecodeLastRuneAtRightSegment(t *testing.T) {
	gb := NewGapBuffer([]byte("café world"))
	gb.Insert(0, []byte("X")) // gap sits at byte 1; rune before logical 6 is 'é' (2 bytes)

	r, size := gb.decodeLastRuneAt(6)
	if r != 'é' || size != 2 {
		t.Fatalf("decodeLastRuneAt(6) = %q/%d, want é/2", r, size)
	}

	gb.Delete(6-size, size)
	if got := string(gb.Contents()); got != "Xcaf world" {
		t.Fatalf("after backspace-multibyte contents = %q, want %q", got, "Xcaf world")
	}
}

func TestDecodeRuneAtRoundTrip(t *testing.T) {
	gb := NewGapBuffer([]byte("héllo wörld"))
	for i := 0; i < gb.Len(); i++ {
		_, sz := gb.decodeRuneAt(i)
		if sz <= 0 {
			t.Fatalf("decodeRuneAt(%d) size=%d", i, sz)
		}
		i += sz - 1
	}
}

func TestLineCountEmpty(t *testing.T) {
	if c := NewGapBuffer(nil).LineCount(); c != 1 {
		t.Fatalf("empty buffer line count = %d, want 1", c)
	}
	if c := NewGapBuffer([]byte("\n")).LineCount(); c != 2 {
		t.Fatalf("single-newline buffer line count = %d, want 2", c)
	}
}
