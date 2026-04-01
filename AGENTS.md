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
- `golang.design/x/clipboard` — System clipboard

## Package Map

```
internal/
├── app/           Bubble Tea model, CLI, print mode
│   ├── cli.go       CLIArgs, PrintUsage
│   ├── markdown.go  Print-mode via render.PrintRenderer
│   └── model.go     Model, View, Update, help dialog, notifications
├── config/        Theme → lipgloss styles
│   ├── config.go    Config, EditorStyles, buildConfig
│   └── theme.go     Theme, DefaultTheme, ToEditorStyles
├── constants/     NotificationType, timing
│   ├── notifications.go
│   └── timing.go
├── cursor/        Persistent cursor position (~/.cache/md-cli/)
│   └── cursor.go    PositionStore
├── markdown/      Framework-agnostic parsing
│   ├── types.go       InlineType, InlineElement, SpanType, SyntaxSpan, line consts
│   ├── classify.go    IsCodeFence, IsListLine, IsHeadingLine, ClassifyLine
│   ├── delimiter.go   FindClosingDelimiter
│   └── inline.go      ParseInlineElements, FindSyntaxSpans
├── render/        LineRenderer interface + lipgloss impl
│   ├── types.go       LineRenderer, StyleFunc, re-exports
│   ├── renderer.go    lipglossRenderer, styleCache, RenderLine
│   ├── inline.go      RenderInline, RenderSourceInline
│   ├── table.go       Table rendering
│   ├── list.go        Bullet/numbered/checkbox
│   ├── image.go       Image placeholder
│   └── print.go       PrintRenderer
├── editor/        Buffer, nav, selection, undo, keys, view
│   ├── gapbuffer.go   Byte gap buffer + line index + rune methods
│   ├── editor.go      Editor model, View, caches
│   ├── keybindings.go KeyBindings dispatch
│   ├── navigation.go  Rune-aware cursor movement
│   ├── selection.go   Selection model
│   ├── undo.go        Undo/redo stacks + grouping
│   ├── activeregion.go FindBlockRegion, bounds detection
│   └── editor_test.go + *_test.go
└── utils/         ReadFile, WriteFile, FilterMarkdownFiles
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

1. `markdown.ClassifyLine` → line type. Block detectors: `IsCodeFence`, `IsListLine`, `IsHeadingLine`, `IsTableLine`.
2. `markdown.ParseInlineElements` → `[]InlineElement`. `FindSyntaxSpans` → `[]SyntaxSpan`.
3. `render.LineRenderer.RenderLine(line, inCodeBlock)` — dispatches by line type.
4. `editor.FindBlockRegion(buf, cursorRow)` — raw source for code blocks, tables, lists.
5. `editor.View()` — rendered vs source per line, cursor/selection overlay.
6. Cache: `renderCache map[int]cacheEntry`, `syntaxCache map[int][]SyntaxSpan`. Invalidated via `afterEdit(row)` / `afterMultiLineEdit()`.
