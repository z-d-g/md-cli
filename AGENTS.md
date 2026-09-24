# AGENTS.md

Go (Bubble Tea v2 + Lipgloss v2) terminal markdown editor with custom in-editor rendering.

## Rules
- Parallel tools when independent. Read files before editing. Bullet-point summaries only.
- Self-documenting names, no comments. DRY/KISS/UNIX.
- Imports: stdlib → blank → project → third-party.
- Mutate state via pointer receivers. Cache with explicit invalidation.
- `strings.Builder` + `Grow()` in hot paths. `unicode/utf8` where needed.
- Library returns errors. UI logs via `slog`.

## Dependencies
- `charm.land/bubbletea/v2` — TUI framework
- `charm.land/lipgloss/v2` — Styling (value-based, `color.Color`)
- `github.com/aymanbagabas/go-nativeclipboard` — System clipboard (CGO=0)
- `github.com/mattn/go-runewidth` — Unicode-aware cell width

## Package Map

```
internal/
├── app/           Bubble Tea model, CLI, print mode
│   ├── cli.go       CLIArgs, ParseCLIArgs, PrintUsage
│   ├── markdown.go  HandlePrintMode, HandlePrintContent
│   └── model.go     Model, View, Update, help dialog, notifications
├── config/        Theme → lipgloss styles
│   ├── config.go    Config, EditorStyles, buildConfig, LoadConfig
│   └── theme.go     Theme, DefaultTheme, ToEditorStyles
├── constants/     NotificationType, timing
│   ├── notifications.go  NotificationType, Message()
│   └── timing.go         NotificationDuration, TypingGroupTimeout, AnimationInterval/Step, MouseScrollLines
├── cursor/        Persistent cursor position (~/.cache/md-cli/)
│   └── cursor.go    FileConfig, PositionStore, Get/Set/Remove
├── markdown/      Framework-agnostic parsing, zero deps
│   ├── types.go       InlineType, InlineElement, SpanType, SyntaxSpan, LineXxx consts
│   ├── classify.go    IsCodeFence, CodeFenceChar, IsBlockquoteLine, IsHorizontalRule, IsListLine, IsHeadingLine, IsTableLine, ClassifyLine, CountBlockquoteDepth, CountLeadingHashes
│   ├── codetracker.go CodeFenceTracker (code-block membership + bounds, char-aware)
│   ├── delimiter.go   FindClosingDelimiter
│   └── inline.go      ParseInlineElements, FindSyntaxSpans, collectSpans
├── render/        LineRenderer interface + lipgloss impl
│   ├── types.go       LineRenderer interface, StyleFunc, re-exports from markdown
│   ├── renderer.go    lipglossRenderer, styleCache, documentLines, terminalWidth, SetDocument, SetTerminalWidth, RenderLine
│   ├── inline.go      RenderInline, RenderSourceInline
│   ├── table.go       tableContext, renderTable, parseTableCells, alignText
│   ├── list.go        renderCheckbox, renderNumberedList, renderBulletList
│   ├── image.go       renderImageAlt (renders ⊞ + alt text)
│   └── print.go       PrintRenderer, RenderDocument
├── editor/        Buffer, nav, selection, undo, keys, view
│   ├── gapbuffer.go   GapBuffer (byte gap buffer + line index + rune methods)
│   ├── editor.go      Editor, View, caches (renderCache, syntaxCache), frame state
│   ├── keybindings.go KeyBindings dispatch, clipboard ops, typing groups
│   ├── navigation.go  Navigation (rune-aware cursor movement, desiredCol)
│   ├── selection.go   Selection (anchor/cursor, SelectAll/Word/Line)
│   ├── undo.go        UndoManager (stack + grouping via GroupUndoEntries)
│   ├── activeregion.go FindBlockRegion, newCodeFenceTracker, bounds detection
│   └── *_test.go      Tests for editor, navigation, selection, undo, activeregion, rendercache
└── utils/         ReadFile, WriteFile, IsMarkdownFile, FilterMarkdownFiles
    └── file.go
```

## Architecture

```
markdown/  pure Go parsing, no deps
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

print: ReadFile → PrintRenderer → stdout
interactive: NewModel → ReadFile → NewEditor(content, LineRenderer)

Model.Update → Editor.Update → KeyBindings.HandleKey
                                  ↓ gap buffer + undo + afterEdit()

Model.View → computeFrameState → visible lines
             getCachedLine / stylizeSourceLine → selection/cursor overlay
```

## Rendering Pipeline

1. `render.RenderLine(line, inCodeBlock)` — dispatches to block detectors (`IsCodeFence`, `IsListLine`, `IsHeadingLine`, `IsTableLine`, etc.).
2. `markdown.ParseInlineElements(line)` → `[]InlineElement`. `FindSyntaxSpans(line)` → `[]SyntaxSpan`.
3. `LineRenderer.RenderInline(elements)` / `RenderSourceInline(elements)` — styled output.
4. `editor.FindBlockRegion(buf, cursorRow, tracker)` — raw source for code blocks, tables, lists, headings.
5. `editor.View()` — rendered vs source per line, cursor/selection overlay.
6. Cache: `renderCache map[int]cacheEntry`, `syntaxCache map[int][]SyntaxSpan`. Invalidated via `afterEdit(row)` / `afterMultiLineEdit()`.
7. `computeFrameState()` builds `frame.tracker *markdown.CodeFenceTracker` once per dirty frame (skipped if clean), then resolves active-region source/raw mode.

## Key Types & Interfaces

```go
// markdown/ — framework-agnostic
type InlineType int; type InlineElement struct { Type, Content, URL, Delimiter, Children }
type SpanType int; type SyntaxSpan struct { Start, End, SpanType }
const LineNormal, LineHeading1..6, LineCodeFence, LineCodeContent, LineBlockQuote

// render/ — abstraction over terminal styling
type StyleFunc func(text string) string
type LineRenderer interface {
    SetDocument(lines func() []string)
    SetTerminalWidth(w int)
    TerminalWidth() int
    RenderLine(line string, isInCodeBlock bool) string
    RenderStyled(text string, lineType int) string
    RenderInline(elements []InlineElement, base lipgloss.Style) string
    RenderSourceInline(elements []InlineElement, base lipgloss.Style) string
    RenderLineNumber(text string) string
    RenderCursorChar(ch string) string
    RenderSelectionChar(ch string) string
    RenderLink(text, url string) string
    TableVersion() int
}

// editor/ — core editing
type GapBuffer struct { data, gapStart, gapEnd, capacity, lineStarts }
type Editor struct { buf, nav, selection, undo, scroll, width, height, renderer, renderCache, syntaxCache, keyBindings, frame }
type Navigation struct { buf, cursor, desiredCol }
type Selection struct { active, anchor, cursor }
type UndoManager struct { undoStack, redoStack, limit }
type UndoEntry struct { offset, deleted, inserted, cursorBefore, cursorAfter }
```

## Undo Grouping

Typing groups: rapid single-character inserts are merged into one `UndoEntry` lazily.
- `typingGroupCount` tracks entries in current group
- On next keypress after `TypingGroupTimeout` (500ms), `GroupUndoEntries(startIdx)` merges them
- Non-typing operations (arrows, backspace, enter) call `endTypingGroup()` which resets counters
- Limit: 500 undo entries

## Active Region (Source Mode)

Cursor position determines which lines show raw markdown source vs rendered output:
- Inside code fence → entire fence region is raw
- On table line → entire table block is raw
- On list line → list block (with blank lines) is raw
- On heading → heading line is raw
- On blockquote → blockquote block is raw
- On line with inline syntax (bold, italic, etc.) → that line is raw
