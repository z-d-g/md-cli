package editor

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"
	"github.com/z-d-g/md-cli/v2/internal/constants"
	"github.com/z-d-g/md-cli/v2/internal/markdown"
	"github.com/z-d-g/md-cli/v2/internal/render"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mattn/go-runewidth"
)

type Editor struct {
	// Core components
	buf       *GapBuffer
	nav       *Navigation
	selection *Selection
	undo      *UndoManager

	// State
	scroll  int // first visible line
	width   int
	height  int
	focused bool
	dirty   bool // tracks if content has been modified

	// Rendering — the only framework dependency
	renderer render.LineRenderer

	// Caches
	renderCache map[int]cacheEntry // line number → cached render output
	syntaxCache map[int][]SyntaxSpan

	// Key handling
	keyBindings *KeyBindings

	// Typing group state
	typingGroupCount int

	// viewScroll is the first document row drawn by the last View call.
	viewScroll int

	// Frame state for caching per-frame computations
	frame struct {
		doc   *markdown.Doc
		dirty bool
	}
}

// cacheEntry stores a rendered line with its source for cache validation.
type cacheEntry struct {
	raw          string // raw line content when cached
	rendered     string // styled output
	inCodeBlock  bool   // whether line was in a code block
	tableVersion int    // renderer table version when cached (for invalidation)
}

// NewEditor creates an editor with the given renderer for styled output.
// The renderer is the only TUI framework dependency — swap it to change frameworks.
func NewEditor(initialContent string, r render.LineRenderer) *Editor {
	var initialBytes []byte
	if initialContent != "" {
		initialBytes = []byte(initialContent)
	}
	buf := NewGapBuffer(initialBytes)

	editor := &Editor{
		buf:         buf,
		nav:         NewNavigation(buf, 0),
		selection:   NewSelection(),
		undo:        NewUndoManager(500),
		scroll:      0,
		focused:     true,
		renderer:    r,
		renderCache: make(map[int]cacheEntry),
		syntaxCache: make(map[int][]SyntaxSpan),
		keyBindings: nil,
	}

	editor.keyBindings = NewKeyBindings(editor)
	editor.frame.dirty = true
	return editor
}

// LineCount returns the number of lines in the buffer.
func (e *Editor) LineCount() int {
	return e.buf.LineCount()
}

// LineAt returns the raw text of line n.
func (e *Editor) LineAt(n int) string {
	return e.buf.LineAt(n)
}

func (e *Editor) computeFrameState() {
	lineCount := e.buf.LineCount()
	if !e.frame.dirty && e.frame.doc != nil && e.frame.doc.Len() == lineCount {
		return
	}
	var b strings.Builder
	b.Grow(256)
	for i := 0; i < lineCount; i++ {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(e.buf.LineAt(i))
	}
	e.frame.doc = markdown.NewDoc(b.String())
	e.frame.dirty = false
}

// Update applies one message to the editor and returns the command to run, if
// any. The editor is mutated in place.
func (e *Editor) Update(msg tea.Msg) tea.Cmd {
	if !e.focused {
		return nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		e.SetSize(msg.Width, msg.Height)

	case tea.KeyPressMsg:
		return e.keyBindings.HandleKey(msg)

	case tea.MouseWheelMsg:
		if msg.Button == tea.MouseWheelUp {
			for range constants.MouseScrollLines {
				e.keyBindings.handleUp(false)
			}
		} else if msg.Button == tea.MouseWheelDown {
			for range constants.MouseScrollLines {
				e.keyBindings.handleDown(false)
			}
		}

	case tea.PasteMsg:
		return e.keyBindings.HandlePaste(msg.Content)
	}

	return nil
}

// View renders the editor.
func (e *Editor) View() tea.View {
	if e.width <= 0 || e.height <= 0 {
		return tea.NewView("")
	}

	e.computeFrameState()

	var b strings.Builder
	visibleLines := e.visibleLines()
	e.viewScroll = e.scroll

	maxLineNum := e.buf.LineCount()
	lineNumWidth := e.getLineNumWidth() - 1

	cursorRow, cursorCol := e.buf.CursorToRowCol(e.nav.Cursor())
	activeStart, activeEnd, isBlockRegion := e.getActiveRegion()

	for i, rawLine := range visibleLines {
		absLine := e.scroll + i

		// Line number
		lineNumStr := fmt.Sprintf("%*d", lineNumWidth, absLine+1)
		b.WriteString(e.renderLineNumber(lineNumStr))
		b.WriteString(" ")

		isActive := absLine >= activeStart && absLine <= activeEnd
		isCursorLine := absLine == cursorRow

		var line strings.Builder
		if isActive {
			if isBlockRegion {
				if e.selection.IsActive() {
					e.renderWithSelection(&line, rawLine, rawLine, absLine, cursorCol, isCursorLine)
				} else if isCursorLine {
					e.renderWithCursor(&line, rawLine, rawLine, cursorCol)
				} else {
					line.WriteString(rawLine)
				}
			} else {
				if e.selection.IsActive() {
					styled := e.stylizeSourceLine(rawLine, absLine)
					e.renderWithSelection(&line, rawLine, styled, absLine, cursorCol, isCursorLine)
				} else {
					e.renderActiveLine(&line, rawLine, absLine, cursorCol, isCursorLine)
				}
			}
		} else {
			styled := e.getCachedLine(absLine, rawLine)
			if e.selection.IsActive() {
				e.renderWithSelection(&line, rawLine, styled, absLine, cursorCol, isCursorLine)
			} else if isCursorLine {
				e.renderWithCursor(&line, rawLine, styled, cursorCol)
			} else {
				line.WriteString(styled)
			}
		}
		b.WriteString(clipLine(line.String(), e.textWidth()))

		if i < len(visibleLines)-1 {
			b.WriteByte('\n')
		}
	}

	// Fill remaining height
	remainingLines := e.height - len(visibleLines)
	for i := range remainingLines {
		lineNum := e.scroll + len(visibleLines) + i + 1
		if lineNum <= maxLineNum {
			lineNumStr := fmt.Sprintf("%*d", lineNumWidth, lineNum)
			b.WriteString("\n")
			b.WriteString(e.renderLineNumber(lineNumStr))
		} else {
			b.WriteString("\n")
			b.WriteString(strings.Repeat(" ", lineNumWidth+1))
		}
	}

	return tea.NewView(b.String())
}

// getCachedLine returns the rendered line from cache or re-renders.
func (e *Editor) getCachedLine(lineNum int, rawLine string) string {
	isInCodeBlock := e.frame.doc.InCode(lineNum)
	tableVersion := e.renderer.TableVersion()

	if entry, exists := e.renderCache[lineNum]; exists {
		if entry.raw == rawLine && entry.inCodeBlock == isInCodeBlock && entry.tableVersion == tableVersion {
			return entry.rendered
		}
	}

	rendered := e.renderer.RenderLine(rawLine, lineNum, isInCodeBlock)
	e.renderCache[lineNum] = cacheEntry{
		raw:          rawLine,
		rendered:     rendered,
		inCodeBlock:  isInCodeBlock,
		tableVersion: tableVersion,
	}
	return rendered
}

// renderLineNumber renders a line number string with appropriate dimming.
func (e *Editor) renderLineNumber(text string) string {
	return e.renderer.RenderLineNumber(text)
}

// getLineNumWidth returns the width of the line number column.
func (e *Editor) getLineNumWidth() int {
	return len(fmt.Sprintf("%d", e.buf.LineCount())) + 1
}

// textWidth returns the number of cells available for line content after the
// line-number gutter and its trailing space.
func (e *Editor) textWidth() int {
	return max(e.width-e.getLineNumWidth(), 1)
}

// visibleLines returns the lines visible in the viewport.
func (e *Editor) visibleLines() []string {
	if e.height <= 0 {
		return []string{}
	}

	cursorRow, _ := e.buf.CursorToRowCol(e.nav.Cursor())
	if cursorRow < e.scroll {
		e.scroll = cursorRow
	} else if cursorRow >= e.scroll+e.height {
		e.scroll = cursorRow - e.height + 1
	}

	start := e.scroll
	end := min(start+e.height, e.buf.LineCount())

	if start >= end {
		return []string{}
	}

	lines := make([]string, end-start)
	for i := start; i < end; i++ {
		lines[i-start] = e.buf.LineAt(i)
	}

	return lines
}

// document returns the parsed document for the current buffer contents,
// computing it on first use so callers never rebuild the AST themselves.
func (e *Editor) document() *markdown.Doc {
	e.computeFrameState()
	return e.frame.doc
}

// getActiveRegion determines which lines should be shown as raw markdown source.
func (e *Editor) getActiveRegion() (int, int, bool) {
	cursorRow, _ := e.buf.CursorToRowCol(e.nav.Cursor())
	return FindBlockRegion(e.frame.doc, cursorRow)
}

// stylizeSourceLine applies styling to a line showing raw markdown syntax.
func (e *Editor) stylizeSourceLine(line string, lineNum int) string {
	lineType := e.frame.doc.LineKind(lineNum)
	if lineType == markdown.LineNormal || lineType == markdown.LineBlockQuote {
		elements := render.ParseInlineElements(line)
		return e.renderer.RenderSourceInline(elements, lipgloss.Style{})
	}
	return e.renderer.RenderStyled(line, lineType)
}

// renderActiveLine renders a line with mixed raw and rendered content.
func (e *Editor) renderActiveLine(b *strings.Builder, rawLine string, lineNum int, cursorCol int, isCursorLine bool) {
	spans, exists := e.syntaxCache[lineNum]
	if !exists {
		spans = FindSyntaxSpans(rawLine)
		e.syntaxCache[lineNum] = spans
	}

	cursorOnSyntax := false
	if isCursorLine {
		cursorOnSyntax, _ = IsCursorOnSyntax(rawLine, cursorCol)
	}

	if cursorOnSyntax {
		styled := e.stylizeSourceLine(rawLine, lineNum)
		if isCursorLine {
			e.renderWithCursor(b, rawLine, styled, cursorCol)
		} else {
			b.WriteString(styled)
		}
	} else {
		styled := e.getCachedLine(lineNum, rawLine)
		if isCursorLine {
			e.renderWithCursor(b, rawLine, styled, cursorCol)
		} else {
			b.WriteString(styled)
		}
	}
}

func (e *Editor) renderWithSelection(b *strings.Builder, rawLine, styledLine string, lineNum int, cursorCol int, isCursorLine bool) {
	if !e.selection.IsActive() {
		if isCursorLine {
			e.renderWithCursor(b, rawLine, styledLine, cursorCol)
		} else {
			b.WriteString(styledLine)
		}
		return
	}

	selectionStart := e.selection.Start()
	selectionEnd := e.selection.End()

	selectionStartRow, selectionStartCol := e.buf.CursorToRowCol(selectionStart)
	selectionEndRow, selectionEndCol := e.buf.CursorToRowCol(selectionEnd)

	if lineNum < selectionStartRow || lineNum > selectionEndRow {
		if isCursorLine {
			e.renderWithCursor(b, rawLine, styledLine, cursorCol)
		} else {
			b.WriteString(styledLine)
		}
		return
	}

	lineSelStart := 0
	lineSelEnd := utf8.RuneCountInString(rawLine)

	if lineNum == selectionStartRow {
		lineSelStart = selectionStartCol
	}
	if lineNum == selectionEndRow {
		lineSelEnd = selectionEndCol
	}

	selStartDisp := displayOffset(rawLine, lineSelStart)
	selEndDisp := displayOffset(rawLine, lineSelEnd)
	writeStyledRunes(b, styledLine, selStartDisp, selEndDisp, e.renderSelectionChar)
}

// renderWithCursor renders a line. The hardware cursor is positioned by the
// model from CursorScreen, so no block cursor is drawn here.
func (e *Editor) renderWithCursor(b *strings.Builder, rawLine, styledLine string, cursorCol int) {
	b.WriteString(styledLine)
}

func skipEscape(s string, i int) int {
	i++
	if i >= len(s) {
		return i
	}
	switch s[i] {
	case '[':
		i++
		for i < len(s) {
			c := s[i]
			i++
			if c >= 0x40 && c <= 0x7e {
				return i
			}
		}
	case ']', 'P', 'X', '^', '_':
		for i < len(s) {
			if s[i] == '\x07' {
				return i + 1
			}
			if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '\\' {
				return i + 2
			}
			i++
		}
	case '(', ')', '#', '"':
		if i < len(s) {
			i++
		}
	}
	return i
}

// displayOffset converts a rune column into a terminal cell offset.
func displayOffset(line string, runeCol int) int {
	var n, col int
	for col < runeCol && n < len(line) {
		r, sz := utf8.DecodeRuneInString(line[n:])
		if r == utf8.RuneError && sz <= 1 {
			break
		}
		n += sz
		col++
	}
	return runewidth.StringWidth(line[:n])
}

// displayWidth returns the terminal cell width of a styled string.
func displayWidth(line string) int { return ansi.StringWidth(line) }

// writeStyledRunes applies selection styling to the styled line. The cursor is
// drawn by the terminal, so only selection needs overlaying.
func writeStyledRunes(b *strings.Builder, styledLine string, selStart, selEnd int, renderSelection func(string) string) {
	col := 0
	for i := 0; i < len(styledLine); {
		if styledLine[i] == '\x1b' {
			seq, width, n, _ := ansi.DecodeSequence(styledLine[i:], 0, nil)
			if n == 0 {
				b.WriteString(styledLine[i:])
				return
			}
			if width == 0 {
				b.WriteString(seq)
				i += n
				continue
			}
			b.WriteString(overlay(styledLine[i:i+width], col, col+width, selStart, selEnd, renderSelection))
			col += width
			i += width
			continue
		}

		r, size := utf8.DecodeRuneInString(styledLine[i:])
		if r == utf8.RuneError && size <= 1 {
			b.WriteString(styledLine[i : i+1])
			col++
			i++
			continue
		}
		w := max(runewidth.RuneWidth(r), 1)
		b.WriteString(overlay(styledLine[i:i+size], col, col+w, selStart, selEnd, renderSelection))
		col += w
		i += size
	}

}

// overlay applies selection styling to one grapheme occupying cells [start,end).
func overlay(text string, start, end, selStart, selEnd int, renderSelection func(string) string) string {
	if selStart >= 0 && selEnd > selStart && start < selEnd && end > selStart {
		return renderSelection(text)
	}
	return text
}

// CursorScreen returns the 0-based (row, col) of the caret within the rendered
// view, including the line-number gutter and the top bar offset.
func (e *Editor) CursorScreen() (int, int) {
	row, col := e.buf.CursorToRowCol(e.nav.Cursor())
	return row - e.viewScroll, e.getLineNumWidth() + displayOffset(e.LineAt(row), col)
}

// clipLine truncates a styled line to width cells without cutting escape
// sequences, so long lines never wrap and desync the viewport.
func clipLine(styled string, width int) string {
	if width <= 0 || ansi.StringWidth(styled) <= width {
		return styled
	}
	return ansi.Truncate(styled, width, "")
}

// renderSelectionChar renders a character with selection styling via the renderer.
func (e *Editor) renderSelectionChar(ch string) string {
	if e.renderer != nil {
		return e.renderer.RenderSelectionChar(ch)
	}
	return ch
}

func (e *Editor) Value() string {
	return string(e.buf.Contents())
}

func (e *Editor) SetValue(content string) {
	e.buf = NewGapBuffer([]byte(content))
	e.nav = NewNavigation(e.buf, 0)
	e.scroll = 0
	e.renderCache = make(map[int]cacheEntry)
	e.syntaxCache = make(map[int][]SyntaxSpan)
	e.undo.Clear()
	e.selection.Clear()
	e.frame.doc = nil
	e.markDirty()
}

// markDirty flags the buffer as modified and forces recomputation of the
// per-frame code-block cache on the next View. Any content change invalidates
// both, so callers must use this instead of touching the fields directly.
func (e *Editor) markDirty() {
	e.dirty = true
	e.frame.dirty = true
}

func (e *Editor) afterEdit(affectedRow int) {
	e.markDirty()
	delete(e.renderCache, affectedRow)
	delete(e.syntaxCache, affectedRow)
}

// invalidateEdit drops cache entries for every line touched by an edit that
// changed bytes between start and end (or inserted at start). Byte offsets may
// be from before or after the edit; the affected line range is clamped to the
// current buffer so stale line numbers cannot leak.
func (e *Editor) invalidateEdit(start, end int) {
	e.markDirty()
	lo, hi := start, end
	if lo > hi {
		lo, hi = hi, lo
	}
	lineLo, _ := e.buf.CursorToRowCol(lo)
	lineHi, _ := e.buf.CursorToRowCol(min(hi, e.buf.Len()))
	for row := lineLo; row <= lineHi; row++ {
		delete(e.renderCache, row)
		delete(e.syntaxCache, row)
	}
	// The edit may have shifted every later line's bytes: drop anything past it.
	for row := range e.renderCache {
		if row > lineHi {
			delete(e.renderCache, row)
		}
	}
	for row := range e.syntaxCache {
		if row > lineHi {
			delete(e.syntaxCache, row)
		}
	}
}

func (e *Editor) afterMultiLineEdit() {
	e.markDirty()
	e.renderCache = make(map[int]cacheEntry)
	e.syntaxCache = make(map[int][]SyntaxSpan)
}

func (e *Editor) SetSize(width, height int) {
	e.width = width
	e.height = height
	e.renderer.SetTerminalWidth(width)
}

func (e *Editor) Focus() {
	e.focused = true
}

func (e *Editor) Blur() {
	e.focused = false
}

func (e *Editor) Cursor() (int, int) {
	return e.buf.CursorToRowCol(e.nav.Cursor())
}

// SetCursor places the caret at (row, col). Out-of-range positions are clamped
// to the buffer, matching every other navigation entry point.
func (e *Editor) SetCursor(row, col int) {
	e.nav.SetCursor(e.buf.RowColToByteOffset(row, col))
}

func (e *Editor) IsDirty() bool {
	return e.dirty
}

func (e *Editor) MarkClean() {
	e.dirty = false
}

func (e *Editor) MarkDirty() {
	e.dirty = true
}
