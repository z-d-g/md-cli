package editor

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/z-d-g/md-cli/internal/constants"
	"github.com/z-d-g/md-cli/internal/markdown"
	"github.com/z-d-g/md-cli/internal/render"

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

	// Frame state for caching per-frame computations
	frame struct {
		tracker *markdown.CodeFenceTracker
		dirty   bool
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

func (e *Editor) Init() tea.Cmd {
	return nil
}

func (e *Editor) computeFrameState() {
	lineCount := e.buf.LineCount()
	if !e.frame.dirty && e.frame.tracker != nil && len(e.frame.tracker.CodeBlockLines()) == lineCount {
		return
	}
	lines := make([]string, lineCount)
	for i := 0; i < lineCount; i++ {
		lines[i] = e.buf.LineAt(i)
	}
	e.frame.tracker = markdown.NewCodeFenceTracker(lines)
	e.frame.dirty = false
}

func (e *Editor) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !e.focused {
		return e, nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		e.SetSize(msg.Width, msg.Height)
		return e, nil

	case tea.KeyPressMsg:
		cmd := e.keyBindings.HandleKey(msg)
		return e, cmd

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
		return e, nil

	case tea.PasteMsg:
		cmd := e.keyBindings.HandlePaste(msg.Content)
		return e, cmd
	}

	return e, nil
}

// View renders the editor.
func (e *Editor) View() tea.View {
	if e.width <= 0 || e.height <= 0 {
		return tea.NewView("")
	}

	e.computeFrameState()

	var b strings.Builder
	visibleLines := e.visibleLines()

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

		if isActive {
			if isBlockRegion {
				if e.selection.IsActive() {
					e.renderWithSelection(&b, rawLine, rawLine, absLine, cursorCol, isCursorLine)
				} else if isCursorLine {
					e.renderWithCursor(&b, rawLine, rawLine, cursorCol)
				} else {
					b.WriteString(rawLine)
				}
			} else {
				if e.selection.IsActive() {
					styled := e.stylizeSourceLine(rawLine, absLine)
					e.renderWithSelection(&b, rawLine, styled, absLine, cursorCol, isCursorLine)
				} else {
					e.renderActiveLine(&b, rawLine, absLine, cursorCol, isCursorLine)
				}
			}
		} else {
			styled := e.getCachedLine(absLine, rawLine)
			if e.selection.IsActive() {
				e.renderWithSelection(&b, rawLine, styled, absLine, cursorCol, isCursorLine)
			} else if isCursorLine {
				e.renderWithCursor(&b, rawLine, styled, cursorCol)
			} else {
				b.WriteString(styled)
			}
		}

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
	isInCodeBlock := e.frame.tracker.IsInside(lineNum)
	tableVersion := e.renderer.TableVersion()

	if entry, exists := e.renderCache[lineNum]; exists {
		if entry.raw == rawLine && entry.inCodeBlock == isInCodeBlock && entry.tableVersion == tableVersion {
			return entry.rendered
		}
	}

	rendered := e.renderer.RenderLine(rawLine, isInCodeBlock)
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

// getActiveRegion determines which lines should be shown as raw markdown source.
func (e *Editor) getActiveRegion() (int, int, bool) {
	cursorRow, _ := e.buf.CursorToRowCol(e.nav.Cursor())
	return FindBlockRegion(e.buf, cursorRow, e.frame.tracker)
}

// stylizeSourceLine applies styling to a line showing raw markdown syntax.
func (e *Editor) stylizeSourceLine(line string, lineNum int) string {
	lineType := markdown.ClassifyLine(line, e.frame.tracker.IsInside(lineNum))
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

	cursorDisp := displayOffset(rawLine, cursorCol)
	selStartDisp := displayOffset(rawLine, lineSelStart)
	selEndDisp := displayOffset(rawLine, lineSelEnd)
	writeStyledRunes(b, styledLine, selStartDisp, selEndDisp, cursorDisp, isCursorLine, e.renderCursorChar, e.renderSelectionChar)
}

func (e *Editor) renderWithCursor(b *strings.Builder, rawLine, styledLine string, cursorCol int) {
	cursorDisp := displayOffset(rawLine, cursorCol)
	lineWidth := displayWidth(rawLine)
	if cursorDisp >= lineWidth {
		b.WriteString(styledLine)
		b.WriteString(e.renderCursorChar(" "))
		return
	}

	writeStyledRunes(b, styledLine, -1, -1, cursorDisp, true, e.renderCursorChar, e.renderSelectionChar)
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

func displayOffset(line string, runeCol int) int {
	var n, col int
	for col < runeCol {
		r, sz := utf8.DecodeRuneInString(line[n:])
		if r == utf8.RuneError {
			break
		}
		n += sz
		col++
	}
	return runewidth.StringWidth(line[:n])
}

func displayWidth(line string) int {
	n := 0
	for i := 0; i < len(line); {
		r, size := utf8.DecodeRuneInString(line[i:])
		if r == utf8.RuneError {
			break
		}
		n += runewidth.RuneWidth(r)
		i += size
	}
	return n
}

func writeStyledRunes(b *strings.Builder, styledLine string, selStart, selEnd, cursorDisp int, isCursorLine bool, renderCursor, renderSelection func(string) string) {
	var displayCol int
	for i := 0; i < len(styledLine); {
		if styledLine[i] == '\x1b' {
			next := skipEscape(styledLine, i)
			b.WriteString(styledLine[i:next])
			i = next
			continue
		}

		r, size := utf8.DecodeRuneInString(styledLine[i:])
		if r == utf8.RuneError {
			i++
			continue
		}
		w := runewidth.RuneWidth(r)

		isCursorChar := isCursorLine && displayCol == cursorDisp
		isSelected := selStart >= 0 && selEnd >= 0 && displayCol >= selStart && displayCol < selEnd

		switch {
		case isCursorChar:
			b.WriteString(renderCursor(styledLine[i : i+size]))
		case isSelected:
			b.WriteString(renderSelection(styledLine[i : i+size]))
		default:
			b.WriteString(styledLine[i : i+size])
		}

		displayCol += w
		i += size
	}

	if isCursorLine && displayCol <= cursorDisp {
		b.WriteString(renderCursor(" "))
	}
}

// renderCursorChar renders a character with cursor styling via the renderer.
func (e *Editor) renderCursorChar(ch string) string {
	if e.renderer != nil {
		return e.renderer.RenderCursorChar(ch)
	}
	return ch
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
	e.frame.tracker = nil
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

func (e *Editor) SetCursor(row, col int) {
	cursorPos := e.buf.RowColToByteOffset(row, col)
	e.nav.SetCursor(cursorPos)
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
