package editor

import (
	"strings"
	"time"

	"github.com/z-d-g/md-cli/internal/constants"

	"golang.design/x/clipboard"

	tea "charm.land/bubbletea/v2"
)

type NotificationMsg struct {
	MessageType constants.NotificationType
}

type KeyBindings struct {
	editor      *Editor
	lastKeyTime time.Time
	typingGroup bool
}

func NewKeyBindings(editor *Editor) *KeyBindings {
	return &KeyBindings{
		editor:      editor,
		lastKeyTime: time.Now(),
		typingGroup: false,
	}
}

func (kb *KeyBindings) HandleKey(msg tea.KeyPressMsg) tea.Cmd {
	// Check if we should start/end a typing group
	currentTime := time.Now()
	if kb.typingGroup && currentTime.Sub(kb.lastKeyTime) > constants.TypingGroupTimeout {
		startIdx := max(kb.editor.undo.Len()-kb.editor.typingGroupCount, 0)
		kb.editor.undo.GroupUndoEntries(startIdx)
		kb.typingGroup = false
		kb.editor.typingGroupCount = 0
	}

	kb.lastKeyTime = currentTime

	// Handle key based on string representation
	switch msg.String() {
	case "space":
		return kb.handleInsert(' ')
	case "backspace":
		return kb.handleBackspace()
	case "delete":
		return kb.handleDelete()
	case "enter":
		return kb.handleEnter()
	case "tab":
		return kb.handleTab()
	case "shift+tab":
		return kb.handleShiftTab()
	case "up":
		return kb.handleUp(false)
	case "shift+up":
		return kb.handleUp(true)
	case "down":
		return kb.handleDown(false)
	case "shift+down":
		return kb.handleDown(true)
	case "left":
		return kb.handleLeft(false)
	case "shift+left":
		return kb.handleLeft(true)
	case "ctrl+left":
		return kb.handleWordLeft(false)
	case "ctrl+shift+left":
		return kb.handleWordLeft(true)
	case "right":
		return kb.handleRight(false)
	case "shift+right":
		return kb.handleRight(true)
	case "ctrl+right":
		return kb.handleWordRight(false)
	case "ctrl+shift+right":
		return kb.handleWordRight(true)
	case "home":
		return kb.handleHome(false)
	case "shift+home":
		return kb.handleHome(true)
	case "end":
		return kb.handleEnd(false)
	case "shift+end":
		return kb.handleEnd(true)
	case "pgup":
		return kb.handlePageUp(false)
	case "shift+pgup":
		return kb.handlePageUp(true)
	case "pgdown":
		return kb.handlePageDown(false)
	case "shift+pgdown":
		return kb.handlePageDown(true)
	case "ctrl+home":
		return kb.handleDocStart(false)
	case "ctrl+shift+home":
		return kb.handleDocStart(true)
	case "ctrl+end":
		return kb.handleDocEnd(false)
	case "ctrl+shift+end":
		return kb.handleDocEnd(true)
	case "ctrl+z":
		return kb.handleUndo()
	case "ctrl+y":
		return kb.handleRedo()
	case "ctrl+c":
		return kb.handleCopy()
	case "ctrl+x":
		return kb.handleCut()
	case "ctrl+v":
		return kb.handlePaste()
	case "ctrl+a":
		return kb.handleSelectAll()
	case "ctrl+d":
		return kb.handleSelectWord()
	case "ctrl+l":
		return kb.handleSelectLine()
	case "alt+backspace":
		return kb.handleDeleteWordBackward()
	case "alt+delete":
		return kb.handleDeleteWordForward()

	default:
		// Handle character input
		if msg.Text != "" {
			return kb.handleRunes(msg)
		}
	}

	return nil
}

func (kb *KeyBindings) handleRunes(msg tea.KeyPressMsg) tea.Cmd {
	if kb.editor.selection.IsActive() {
		kb.deleteSelection()
	}

	for _, r := range msg.Text {
		kb.insertRune(r)
	}

	// Start or continue typing group
	if !kb.typingGroup {
		kb.typingGroup = true
		kb.editor.typingGroupCount = 0
	}
	kb.editor.typingGroupCount++

	return nil
}

func (kb *KeyBindings) handleInsert(r rune) tea.Cmd {
	if kb.editor.selection.IsActive() {
		kb.deleteSelection()
	}

	kb.insertRune(r)

	// Start or continue typing group
	if !kb.typingGroup {
		kb.typingGroup = true
		kb.editor.typingGroupCount = 0
	}
	kb.editor.typingGroupCount++

	return nil
}

func (kb *KeyBindings) handleBackspace() tea.Cmd {
	currentCursor := kb.editor.nav.Cursor()

	if kb.editor.selection.IsActive() {
		kb.deleteSelection()
		kb.endTypingGroup()
		return nil
	}

	// Check if we're at start of line and need to merge with previous line
	row, col := kb.editor.buf.CursorToRowCol(currentCursor)
	if col == 0 && row > 0 {
		// Merge with previous line
		prevLineEnd := kb.editor.buf.ByteOffsetOfLine(row) - 1
		if prevLineEnd >= 0 {
			kb.recordUndo(UndoEntry{
				offset:       prevLineEnd,
				deleted:      []byte{'\n'},
				cursorBefore: currentCursor,
				cursorAfter:  prevLineEnd,
			})
			kb.editor.buf.Delete(prevLineEnd, 1)
			kb.editor.nav.SetCursor(prevLineEnd)

			// Invalidate all cache entries since line numbers may have shifted
			kb.editor.afterMultiLineEdit()
		}
	} else {
		// Delete character before cursor
		if currentCursor > 0 {
			rowBefore, _ := kb.editor.buf.CursorToRowCol(currentCursor)

			_, size := kb.editor.buf.decodeLastRuneAt(currentCursor)

			// Copy deleted bytes for undo
			deleted := kb.editor.buf.slice(currentCursor-size, currentCursor)
			deletedCopy := make([]byte, len(deleted))
			copy(deletedCopy, deleted)

			kb.recordUndo(UndoEntry{
				offset:       currentCursor - size,
				deleted:      deletedCopy,
				inserted:     nil,
				cursorBefore: currentCursor,
				cursorAfter:  currentCursor - size,
			})

			kb.editor.buf.Delete(currentCursor-size, size)
			kb.editor.nav.SetCursor(currentCursor - size)

			// Invalidate render and syntax cache for affected lines
			rowAfter, _ := kb.editor.buf.CursorToRowCol(kb.editor.nav.Cursor())
			kb.editor.afterEdit(rowBefore)
			if rowAfter != rowBefore {
				kb.editor.afterEdit(rowAfter)
			}
		}
	}

	kb.endTypingGroup()
	kb.editor.nav.updateDesiredCol()
	return nil
}

func (kb *KeyBindings) handleDelete() tea.Cmd {
	currentCursor := kb.editor.nav.Cursor()

	if kb.editor.selection.IsActive() {
		kb.deleteSelection()
		kb.endTypingGroup()
		return nil
	}

	// Delete character after cursor
	if currentCursor < kb.editor.buf.Len() {
		// Get current line before deletion
		rowBefore, _ := kb.editor.buf.CursorToRowCol(currentCursor)

		_, size := kb.editor.buf.decodeRuneAt(currentCursor)

		// Copy deleted bytes for undo
		deleted := kb.editor.buf.slice(currentCursor, currentCursor+size)
		deletedCopy := make([]byte, len(deleted))
		copy(deletedCopy, deleted)

		kb.recordUndo(UndoEntry{
			offset:       currentCursor,
			deleted:      deletedCopy,
			inserted:     nil,
			cursorBefore: currentCursor,
			cursorAfter:  currentCursor,
		})

		kb.editor.buf.Delete(currentCursor, size)

		// If we deleted a newline character, invalidate everything
		if string(deleted) == "\n" {
			kb.editor.afterMultiLineEdit()
		}

		// Invalidate render and syntax cache for affected lines
		rowAfter, _ := kb.editor.buf.CursorToRowCol(kb.editor.nav.Cursor())
		kb.editor.afterEdit(rowBefore)
		if rowAfter != rowBefore {
			kb.editor.afterEdit(rowAfter)
		}
	}

	kb.endTypingGroup()
	return nil
}

// handleDeleteWordBackward deletes the word before the cursor (alt+backspace).
func (kb *KeyBindings) handleDeleteWordBackward() tea.Cmd {
	currentCursor := kb.editor.nav.Cursor()
	if currentCursor == 0 {
		return nil
	}

	if kb.editor.selection.IsActive() {
		kb.deleteSelection()
		kb.endTypingGroup()
		return nil
	}

	// Find word start (same logic as MoveWordLeft)
	target := currentCursor
	for target > 0 {
		r, size := kb.editor.buf.decodeLastRuneAt(target)
		if r != ' ' && r != '\t' && r != '\n' {
			break
		}
		target -= size
	}
	for target > 0 {
		r, size := kb.editor.buf.decodeLastRuneAt(target)
		if r == ' ' || r == '\t' || r == '\n' {
			break
		}
		target -= size
	}

	// Delete bytes between target and cursor
	deleteLen := currentCursor - target
	deleted := kb.editor.buf.slice(target, currentCursor)
	deletedCopy := make([]byte, len(deleted))
	copy(deletedCopy, deleted)

	kb.recordUndo(UndoEntry{
		offset:       target,
		deleted:      deletedCopy,
		inserted:     nil,
		cursorBefore: currentCursor,
		cursorAfter:  target,
	})

	kb.editor.buf.Delete(target, deleteLen)
	kb.editor.nav.SetCursor(target)

	rowBefore, _ := kb.editor.buf.CursorToRowCol(currentCursor)
	rowAfter, _ := kb.editor.buf.CursorToRowCol(target)
	kb.editor.afterEdit(rowBefore)
	if rowAfter != rowBefore {
		kb.editor.afterEdit(rowAfter)
	}

	kb.endTypingGroup()
	kb.editor.nav.updateDesiredCol()
	return nil
}

// handleDeleteWordForward deletes the word after the cursor (alt+delete).
func (kb *KeyBindings) handleDeleteWordForward() tea.Cmd {
	currentCursor := kb.editor.nav.Cursor()
	if currentCursor >= kb.editor.buf.Len() {
		return nil
	}

	if kb.editor.selection.IsActive() {
		kb.deleteSelection()
		kb.endTypingGroup()
		return nil
	}

	// Find word end (same logic as MoveWordRight)
	target := currentCursor
	// Skip non-whitespace
	for target < kb.editor.buf.Len() {
		r, size := kb.editor.buf.decodeRuneAt(target)
		if r == ' ' || r == '\t' || r == '\n' {
			break
		}
		target += size
	}
	// Skip whitespace
	for target < kb.editor.buf.Len() {
		r, size := kb.editor.buf.decodeRuneAt(target)
		if r != ' ' && r != '\t' && r != '\n' {
			break
		}
		target += size
	}

	// Delete bytes between cursor and target
	deleteLen := target - currentCursor
	deleted := kb.editor.buf.slice(currentCursor, target)
	deletedCopy := make([]byte, len(deleted))
	copy(deletedCopy, deleted)

	kb.recordUndo(UndoEntry{
		offset:       currentCursor,
		deleted:      deletedCopy,
		inserted:     nil,
		cursorBefore: currentCursor,
		cursorAfter:  currentCursor,
	})

	kb.editor.buf.Delete(currentCursor, deleteLen)

	rowBefore, _ := kb.editor.buf.CursorToRowCol(currentCursor)
	kb.editor.afterEdit(rowBefore)

	kb.endTypingGroup()
	kb.editor.nav.updateDesiredCol()
	return nil
}

func (kb *KeyBindings) handleEnter() tea.Cmd {
	currentCursor := kb.editor.nav.Cursor()

	if kb.editor.selection.IsActive() {
		kb.deleteSelection()
	}

	kb.recordUndo(UndoEntry{
		offset:       currentCursor,
		inserted:     []byte{'\n'},
		cursorBefore: currentCursor,
		cursorAfter:  currentCursor + 1,
	})

	kb.editor.buf.Insert(currentCursor, []byte{'\n'})
	kb.editor.nav.SetCursor(currentCursor + 1)

	// Invalidate all cache entries since line numbers may have shifted
	kb.editor.afterMultiLineEdit()

	kb.endTypingGroup()
	kb.editor.nav.updateDesiredCol()
	return nil
}

func (kb *KeyBindings) handleTab() tea.Cmd {
	currentCursor := kb.editor.nav.Cursor()

	if kb.editor.selection.IsActive() {
		kb.deleteSelection()
	}

	row, _ := kb.editor.buf.CursorToRowCol(currentCursor)
	lineStart := kb.editor.buf.ByteOffsetOfLine(row)

	indent := []byte("    ")

	kb.recordUndo(UndoEntry{
		offset:       lineStart,
		inserted:     indent,
		cursorBefore: currentCursor,
		cursorAfter:  currentCursor + len(indent),
	})

	kb.editor.buf.Insert(lineStart, indent)
	kb.editor.nav.SetCursor(currentCursor + len(indent))
	kb.editor.afterEdit(row)
	kb.endTypingGroup()
	return nil
}

func (kb *KeyBindings) handleShiftTab() tea.Cmd {
	cursor := kb.editor.nav.Cursor()
	row, col := kb.editor.buf.CursorToRowCol(cursor)
	lineStart := kb.editor.buf.ByteOffsetOfLine(row)
	line := kb.editor.buf.LineAt(row)

	// Count leading spaces (up to 4)
	spaces := 0
	for i := 0; i < len(line) && i < 4 && line[i] == ' '; i++ {
		spaces++
	}
	if spaces == 0 {
		kb.endTypingGroup()
		kb.editor.nav.updateDesiredCol()
		return nil
	}

	kb.recordUndo(UndoEntry{
		offset:       lineStart,
		deleted:      []byte(strings.Repeat(" ", spaces)),
		cursorBefore: cursor,
		cursorAfter:  kb.editor.buf.RowColToByteOffset(row, max(col-spaces, 0)),
	})
	kb.editor.buf.Delete(lineStart, spaces)
	kb.editor.nav.SetCursor(kb.editor.buf.RowColToByteOffset(row, max(col-spaces, 0)))
	kb.editor.afterEdit(row)
	kb.endTypingGroup()
	kb.editor.nav.updateDesiredCol()
	return nil
}

func (kb *KeyBindings) handleMove(shift bool, moveFn func()) tea.Cmd {
	currentCursor := kb.editor.nav.Cursor()

	if shift {
		// Extend selection
		if !kb.editor.selection.IsActive() {
			kb.editor.selection.Activate(currentCursor)
		}
		moveFn()
		kb.editor.selection.MoveCursor(kb.editor.nav.Cursor())
	} else {
		// Clear selection and move
		kb.editor.selection.Clear()
		moveFn()
	}
	return nil
}

func (kb *KeyBindings) handleUp(shift bool) tea.Cmd {
	return kb.handleMove(shift, kb.editor.nav.MoveUp)
}

func (kb *KeyBindings) handleDown(shift bool) tea.Cmd {
	return kb.handleMove(shift, kb.editor.nav.MoveDown)
}

func (kb *KeyBindings) handleLeft(shift bool) tea.Cmd {
	return kb.handleMove(shift, kb.editor.nav.MoveLeft)
}

func (kb *KeyBindings) handleRight(shift bool) tea.Cmd {
	return kb.handleMove(shift, kb.editor.nav.MoveRight)
}

func (kb *KeyBindings) handleWordLeft(shift bool) tea.Cmd {
	return kb.handleMove(shift, kb.editor.nav.MoveWordLeft)
}

func (kb *KeyBindings) handleWordRight(shift bool) tea.Cmd {
	return kb.handleMove(shift, kb.editor.nav.MoveWordRight)
}

func (kb *KeyBindings) handleHome(shift bool) tea.Cmd {
	return kb.handleMove(shift, kb.editor.nav.MoveHome)
}

func (kb *KeyBindings) handleEnd(shift bool) tea.Cmd {
	return kb.handleMove(shift, kb.editor.nav.MoveEnd)
}

func (kb *KeyBindings) handlePageUp(shift bool) tea.Cmd {
	return kb.handleMove(shift, func() { kb.editor.nav.PageUp(kb.editor.height) })
}

func (kb *KeyBindings) handlePageDown(shift bool) tea.Cmd {
	return kb.handleMove(shift, func() { kb.editor.nav.PageDown(kb.editor.height) })
}

func (kb *KeyBindings) handleDocStart(shift bool) tea.Cmd {
	return kb.handleMove(shift, kb.editor.nav.MoveDocStart)
}

func (kb *KeyBindings) handleDocEnd(shift bool) tea.Cmd {
	return kb.handleMove(shift, kb.editor.nav.MoveDocEnd)
}

func (kb *KeyBindings) handleUndo() tea.Cmd {
	kb.endTypingGroup()
	cursorPos, ok := kb.editor.undo.Undo(kb.editor.buf)
	if ok {
		kb.editor.nav.SetCursor(cursorPos)
		kb.editor.selection.Clear()
		// Invalidate all cache entries since line numbers may have shifted
		kb.editor.afterMultiLineEdit()
	}
	return nil
}

func (kb *KeyBindings) handleRedo() tea.Cmd {
	kb.endTypingGroup()
	cursorPos, ok := kb.editor.undo.Redo(kb.editor.buf)
	if ok {
		kb.editor.nav.SetCursor(cursorPos)
		kb.editor.selection.Clear()
		// Invalidate all cache entries since line numbers may have shifted
		kb.editor.afterMultiLineEdit()
	}
	return nil
}

func (kb *KeyBindings) handleCopy() tea.Cmd {
	// First check if we're inside a code block and copy the entire block
	if codeBlock := kb.detectCodeBlockAtCursor(); codeBlock != nil {
		clipboard.Write(clipboard.FmtText, codeBlock)
		return kb.showNotification(constants.CodeBlockCopied)
	}

	// Otherwise, handle normal selection copy
	if kb.editor.selection.IsActive() {
		selected := kb.editor.selection.GetSelectedText(kb.editor.buf)
		if selected != nil {
			clipboard.Write(clipboard.FmtText, selected)
			kb.editor.selection.Clear()
			return kb.showNotification(constants.TextCopied)
		}
	}
	return nil
}

func (kb *KeyBindings) handleCut() tea.Cmd {
	if kb.editor.selection.IsActive() {
		selected := kb.editor.selection.GetSelectedText(kb.editor.buf)
		if selected != nil {
			start := kb.editor.selection.Start()
			end := kb.editor.selection.End()
			currentCursor := kb.editor.nav.Cursor()

			kb.recordUndo(UndoEntry{
				offset:       start,
				deleted:      selected,
				inserted:     nil,
				cursorBefore: currentCursor,
				cursorAfter:  start,
			})

			kb.editor.buf.Delete(start, end-start)
			kb.editor.nav.SetCursor(start)
			kb.editor.selection.Clear()

			clipboard.Write(clipboard.FmtText, selected)

			// Invalidate all cache entries since line numbers may have shifted
			kb.editor.afterMultiLineEdit()
			return kb.showNotification(constants.CutToClipboard)
		}
	}
	return nil
}

func (kb *KeyBindings) handlePaste() tea.Cmd {
	currentCursor := kb.editor.nav.Cursor()

	// Get from clipboard
	pasteContent := clipboard.Read(clipboard.FmtText)
	// If clipboard returns nil, use empty string
	if pasteContent == nil {
		pasteContent = []byte("")
	}

	if kb.editor.selection.IsActive() {
		// Replace selection with pasted content
		selected := kb.editor.selection.GetSelectedText(kb.editor.buf)
		if selected != nil {
			start := kb.editor.selection.Start()
			end := kb.editor.selection.End()

			kb.recordUndo(UndoEntry{
				offset:       start,
				deleted:      selected,
				inserted:     pasteContent,
				cursorBefore: currentCursor,
				cursorAfter:  start + len(pasteContent),
			})

			kb.editor.buf.Delete(start, end-start)
			kb.editor.buf.Insert(start, pasteContent)
			kb.editor.nav.SetCursor(start + len(pasteContent))
			kb.editor.selection.Clear()

			// Invalidate all cache entries since line numbers may have shifted
			kb.editor.afterMultiLineEdit()

			kb.endTypingGroup()

			if len(pasteContent) > 0 {
				return kb.showNotification(constants.PastedFromClipboard)
			}
			return nil
		}
	}

	// No selection - paste at cursor position
	kb.recordUndo(UndoEntry{
		offset:       currentCursor,
		inserted:     pasteContent,
		cursorBefore: currentCursor,
		cursorAfter:  currentCursor + len(pasteContent),
	})

	kb.editor.buf.Insert(currentCursor, pasteContent)
	kb.editor.nav.SetCursor(currentCursor + len(pasteContent))

	// Invalidate all cache entries since line numbers may have shifted
	kb.editor.afterMultiLineEdit()

	kb.endTypingGroup()

	if len(pasteContent) > 0 {
		return kb.showNotification(constants.PastedFromClipboard)
	}
	return nil
}

func (kb *KeyBindings) handleSelectAll() tea.Cmd {
	kb.editor.selection.SelectAll(kb.editor.buf)
	return nil
}

func (kb *KeyBindings) handleSelectWord() tea.Cmd {
	kb.editor.selection.SelectWord(kb.editor.buf, kb.editor.nav.Cursor())
	return nil
}

func (kb *KeyBindings) handleSelectLine() tea.Cmd {
	kb.editor.selection.SelectLine(kb.editor.buf, kb.editor.nav.Cursor())
	return nil
}

func (kb *KeyBindings) deleteSelection() int {
	currentCursor := kb.editor.nav.Cursor()
	selected := kb.editor.selection.GetSelectedText(kb.editor.buf)
	if selected == nil {
		return currentCursor
	}

	start := kb.editor.selection.Start()
	end := kb.editor.selection.End()

	kb.recordUndo(UndoEntry{
		offset:       start,
		deleted:      selected,
		inserted:     nil,
		cursorBefore: currentCursor,
		cursorAfter:  start,
	})

	kb.editor.buf.Delete(start, end-start)
	kb.editor.nav.SetCursor(start)
	kb.editor.selection.Clear()

	// Invalidate all cache entries since line numbers may have shifted
	kb.editor.afterMultiLineEdit()

	return start
}

func (kb *KeyBindings) insertRune(r rune) {
	currentCursor := kb.editor.nav.Cursor()

	char := string(r)
	inserted := []byte(char)

	kb.recordUndo(UndoEntry{
		offset:       currentCursor,
		inserted:     inserted,
		cursorBefore: currentCursor,
		cursorAfter:  currentCursor + len(inserted),
	})

	// Get current line before insertion
	rowBefore, _ := kb.editor.buf.CursorToRowCol(currentCursor)

	kb.editor.buf.Insert(currentCursor, inserted)
	kb.editor.nav.SetCursor(currentCursor + len(inserted))

	// Invalidate render and syntax cache for affected lines
	rowAfter, _ := kb.editor.buf.CursorToRowCol(kb.editor.nav.Cursor())
	kb.editor.afterEdit(rowBefore)
	if rowAfter != rowBefore {
		kb.editor.afterEdit(rowAfter)
	}
}

func (kb *KeyBindings) recordUndo(entry UndoEntry) {
	kb.editor.undo.Record(entry)
}

func (kb *KeyBindings) endTypingGroup() {
	if kb.typingGroup {
		kb.typingGroup = false
		kb.editor.typingGroupCount = 0
	}
}

// detectCodeBlockAtCursor detects if the cursor is inside a code block and returns its content
func (kb *KeyBindings) detectCodeBlockAtCursor() []byte {
	currentRow, _ := kb.editor.buf.CursorToRowCol(kb.editor.nav.Cursor())
	if !isInCodeBlock(kb.editor.buf, currentRow) {
		return nil
	}
	start, end := findCodeBlockBounds(kb.editor.buf, currentRow)
	// Extract content between fences (excluding fence lines)
	var lines []string
	for row := start + 1; row < end; row++ {
		lines = append(lines, kb.editor.buf.LineAt(row))
	}
	return []byte(strings.Join(lines, "\n"))
}

// showNotification creates a command to show a notification in the UI
func (kb *KeyBindings) showNotification(notificationType constants.NotificationType) tea.Cmd {
	return func() tea.Msg {
		return NotificationMsg{MessageType: notificationType}
	}
}
