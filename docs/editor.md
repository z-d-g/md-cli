# Editor

Package `editor` owns the buffer, navigation, selection, undo and the
rendered view.

## Gap buffer

`GapBuffer` is a byte slice with a gap plus a line-start index.

- `Insert` / `Delete` move the gap and mark the line index dirty; the index is
  rebuilt lazily on the next line-aware call.
- `LineAt`, `LineCount`, `CursorToRowCol`, `RowColToByteOffset` operate on
  runes; `decodeRuneAt` and `decodeLastRuneAt` avoid allocations.
- Offsets are clamped, so out-of-range edits and stale positions are safe.

## Navigation and selection

`Navigation` tracks a byte cursor and a `desiredCol` so vertical movement keeps
the preferred column. `Selection` stores anchor/cursor byte offsets; `Start`
and `End` normalise direction.

## Undo

`UndoManager` keeps undo and redo in fixed-size ring buffers (`ring`), so memory
is bounded and no backing array leaks. Typing groups merge consecutive
single-rune inserts that are adjacent in both the buffer and the timeline:

- `KeyBindings.beginTypingGroup` opens a window.
- `endTypingGroup` calls `GroupUndoEntries(start)` when the window closes.
- A 500 ms gap or any non-insert operation closes the window.

## View and caches

`Editor.View` renders the visible rows and clips each line to the text width
with `ansi.Truncate`, so long lines never wrap and desync the viewport.

- `computeFrameState` rebuilds the `markdown.Doc` once per dirty frame.
- `getCachedLine` caches rendered output per line; `invalidateEdit(start, end)`
  drops entries for every touched line and everything after it.
- The hardware cursor is positioned by the model from `Editor.CursorScreen`;
  the editor itself draws only selection.

## Width handling

Cell widths use `github.com/charmbracelet/x/ansi` and `go-runewidth`. Cursor and
selection overlays walk styled strings while decoding ANSI sequences, so wide
and combining characters map to the correct terminal cell.
