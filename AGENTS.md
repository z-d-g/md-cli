# AGENTS.md

Go (Bubble Tea v2 + Lipgloss v2) terminal markdown editor with custom in-editor rendering.

## Module

`github.com/z-d-g/md-cli/v2` — the v2 major line. Imports use the `/v2` suffix.
Releases are tagged `vX.Y.Z`; `make release VERSION=X.Y.Z` then `make publish`.

## Rules
- Go 1.27+. Parallel tools when independent. Read before editing.
- Self-documenting names; comments explain *why*. DRY/KISS/UNIX.
- Imports: stdlib → blank → project → third-party.
- Mutate state via pointer receivers. Cache with explicit invalidation.
- `strings.Builder` + `Grow()` in hot paths. `unicode/utf8`, `x/ansi` for width.
- Library returns errors. UI logs via `slog`.
- Keep `docs/` in sync with any structural change; `docs/INDEX.md` is the TOC.

## Dependencies
- `charm.land/bubbletea/v2` — TUI framework (v2 View carries the hardware cursor)
- `charm.land/lipgloss/v2` — Styling (value-based, `color.Color`); `Canvas`/`Layer` for overlays
- `github.com/charmbracelet/x/ansi` — ANSI-aware truncate/width/sequence decoding
- `github.com/yuin/goldmark/v2` — CommonMark parser (AST + byte offsets)
- `github.com/mattn/go-runewidth` — rune cell width (zero-width/combining)
- `github.com/aymanbagabas/go-nativeclipboard` — System clipboard (CGO=0)
- `golang.org/x/term` — stdout width in print mode

## Package Map

```
internal/
├── app/           Bubble Tea model, CLI, print mode, help dialog
│   ├── cli.go       CLIArgs, ParseCLIArgs, PrintUsage
│   ├── markdown.go  HandlePrintMode, HandlePrintContent
│   ├── model.go     Model, View, Update, notifications, status bar
│   └── help.go      help dialog + animated Canvas overlay
├── config/        Theme → lipgloss styles
│   ├── config.go    Config, EditorStyles, buildConfig, LoadConfig
│   └── theme.go     Theme, DefaultTheme, ToEditorStyles
├── constants/     NotificationType, timing
│   ├── notifications.go  NotificationType, Message()
│   └── timing.go         NotificationDuration, TypingGroupTimeout, AnimationInterval/Step, MouseScrollLines
├── cursor/        Persistent cursor position (~/.cache/md-cli/)
│   └── cursor.go    FileConfig, PositionStore, Get/Set/Remove
├── markdown/      goldmark v2 wrapper, no framework deps
│   ├── types.go       InlineType, InlineElement, SpanType, SyntaxSpan, LineXxx, Fence
│   ├── doc.go         Doc (per-line kind/code/bounds/table/list/quote), NewDoc
│   ├── inline.go      ParseInlineElements, FindSyntaxSpans, MergeSpans
│   ├── delimiter.go   CodeSpanClose, SplitRun, FindClosingDelimiter
│   ├── scan.go        ParseFence, IsCodeFence, IsTableLine, IsListLine, ClassifyLine
│   └── underline.go   ++underline++ goldmark extension
├── render/        LineRenderer interface + lipgloss impl
│   ├── types.go       LineRenderer interface, re-exports from markdown
│   ├── renderer.go    lipglossRenderer, documentLines, terminalWidth, table doc cache, RenderLine
│   ├── inline.go      RenderInline, RenderSourceInline
│   ├── table.go       tableDoc (per-block widths), renderTable(line, idx), alignText
│   ├── list.go        renderCheckbox, renderNumberedList, renderBulletList
│   ├── image.go       renderImageAlt (renders ⊞ + alt text)
│   └── print.go       PrintRenderer, RenderDocument
├── editor/        Buffer, nav, selection, undo, keys, view
│   ├── gapbuffer.go   GapBuffer (byte gap buffer + line index + rune methods)
│   ├── editor.go      Editor, View, caches (renderCache, syntaxCache), frame state
│   ├── keybindings.go KeyBindings dispatch, clipboard ops (errors surfaced), typing groups
│   ├── navigation.go  Navigation (rune-aware cursor movement, desiredCol)
│   ├── selection.go   Selection (anchor/cursor, SelectAll/Word/Line)
│   ├── undo.go        UndoManager (stack + grouping via GroupUndoEntries)
│   ├── activeregion.go FindBlockRegion, syntax-span lookup, cursor-on-syntax
│   └── *_test.go      Tests for editor, navigation, selection, undo, activeregion, rendercache
└── utils/         ReadFile, WriteFile, IsMarkdownFile, FilterMarkdownFiles
    └── file.go
```

## Architecture

```
markdown/  goldmark v2 → Doc (AST + byte offsets)
    ↓
render/    markdown/ + lipgloss v2 → styled strings
    ↓
editor/    render/ → buffer, nav, selection, view
    ↓
app/       editor/ + config/ + cursor/ → Bubble Tea model
```

## Data Flow

```
CLI → ParseCLIArgs()

print: ReadFile → PrintRenderer → stdout (exit 1 on read failure)
interactive: NewModel → ReadFile → NewEditor(content, LineRenderer)

Model.Update → Editor.Update (tea.Cmd) → KeyBindings.HandleKey
                                  ↓ gap buffer + undo + invalidateEdit()

Model.View → computeFrameState → visible lines
             getCachedLine / stylizeSourceLine → selection overlay
             → clipLine (never wraps)
             → tea.View{Cursor: hardware caret}
```

## Rendering Pipeline

1. `render.RenderLine(line, idx, inCodeBlock)` — idx is the document row; dispatches on the line's shape.
2. `markdown.ParseInlineElements(line)` → `[]InlineElement`. `FindSyntaxSpans(line)` → `[]SyntaxSpan`.
3. `LineRenderer.RenderInline(elements)` / `RenderSourceInline(elements)` — styled output.
4. `editor.FindBlockRegion(doc, cursorRow)` — raw source for code blocks, tables, lists, headings, inline syntax.
5. `editor.View()` — rendered vs source per line, cursor/selection overlay.
6. Cache: `renderCache map[int]cacheEntry`, `syntaxCache map[int][]SyntaxSpan`. Invalidated via `invalidateEdit(start, end)` / `afterMultiLineEdit()`.
7. `computeFrameState()` builds `frame.doc *markdown.Doc` once per dirty frame (skipped if clean), then resolves active-region source/raw mode.
8. `invalidateEdit(start, end)` drops cache entries for every line the edit
   touched and all later lines (byte offsets may have shifted).
9. `clipLine` truncates each rendered row to `width - lineNumberColumn` cells
   with `ansi.Truncate`, so long lines never wrap.
10. The caret is the terminal cursor: `Editor.CursorScreen()` gives
    (row, cell) and `Model.View` sets `tea.View.Cursor`.

## Key Types & Interfaces

```go
// markdown/ — goldmark-backed
type InlineType int; type InlineElement struct { Type, Content, URL, Delimiter, Children, Start }
type SpanType int; type SyntaxSpan struct { Start, End, SpanType }
type Fence struct { Char byte; Len int; Info string; Valid bool }
type Doc struct { source, lines, code, bounds, kind, inCodeBlock, table, sep, list, quote }
const LineNormal, LineHeading1..6, LineCodeFence, LineCodeContent, LineBlockQuote

// render/ — abstraction over terminal styling
type LineRenderer interface {
    RenderLine(line string, idx int, isInCodeBlock bool) string
    RenderStyled(text string, lineType int) string
    RenderInline(elements []InlineElement, base lipgloss.Style) string
    RenderSourceInline(elements []InlineElement, base lipgloss.Style) string
    RenderLineNumber(text string) string
    RenderCursorChar(ch string) string
    RenderSelectionChar(ch string) string
    RenderLink(text, url string) string
    SetDocument(lines func() []string)
    SetTerminalWidth(w int)
    TerminalWidth() int
    TableVersion() int
    ResetTable()
}

// editor/ — core editing
type GapBuffer struct { data, gapStart, gapEnd, capacity, lineStarts, lineDirty }
type Editor struct { buf, nav, selection, undo, scroll, viewScroll, width, height, renderer, renderCache, syntaxCache, keyBindings, frame }
type Navigation struct { buf, cursor, desiredCol }
type Selection struct { active, anchor, cursor }
type UndoManager struct { undo, redo ring }
type ring struct { buf []UndoEntry; start, size int }
type UndoEntry struct { offset, deleted, inserted, cursorBefore, cursorAfter }
```

## Undo Grouping

Typing groups: rapid single-character inserts merge into one `UndoEntry` lazily.
- `KeyBindings.beginTypingGroup()` opens a window; `typingGroupCount` tracks it
- On the next keypress after `TypingGroupTimeout` (500ms), or any non-insert op,
  `endTypingGroup()` calls `GroupUndoEntries(start)` and resets the counter
- `UndoManager` stores entries in fixed-size ring buffers (500 per stack), so
  memory is bounded and no array leaks when the oldest entry is dropped
- Grouping requires adjacency in both the buffer (`offset`) and the timeline
  (`cursorAfter == next.cursorBefore`)

## Active Region (Source Mode)

Cursor position determines which lines show raw markdown source vs rendered output:
- Inside code fence → entire fence region is raw
- On table line → entire table block is raw
- On list line → list block (with blank lines) is raw
- On heading → heading line is raw
- On blockquote → blockquote block is raw
- On line with inline syntax (bold, italic, etc.) → that line is raw
