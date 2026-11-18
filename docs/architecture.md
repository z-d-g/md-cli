# Architecture

md-cli is a terminal markdown editor built on Bubble Tea v2, Lipgloss v2 and
goldmark v2.

## Packages

```
markdown/  goldmark v2 → Doc (per-line AST index)
    ↓
render/    markdown/ + lipgloss → styled strings
    ↓
editor/    render/ + gap buffer → buffer, nav, selection, view
    ↓
app/       editor/ + config/ + cursor/ → Bubble Tea model
```

## Data flow

```
CLI → ParseCLIArgs

print: ReadFile → PrintRenderer → stdout
interactive: NewModel → ReadFile → NewEditor(content, LineRenderer)

Model.Update → Editor.Update → KeyBindings.HandleKey
                                  ↓ gap buffer + undo + invalidateEdit()

Model.View → computeFrameState → visible lines
             getCachedLine / stylizeSourceLine → selection overlay
             → clipLine (no wrapping)
             → tea.View{Cursor: hardware caret}
```

## Source of truth

- `markdown.Doc` — the goldmark AST, indexed into per-line kind, code
  membership and block bounds. Built once per dirty frame
  (`Editor.computeFrameState`).
- `render.tableDoc` — table index (per-block column widths, alignments, header
  rows). Cached on the renderer by line-count and content compare. Lookups are
  by document row, so two tables never share state.
- `Editor.renderCache` / `syntaxCache` — keyed by line number, validated by raw
  text, code-block flag and table version. Invalidated by byte range in
  `invalidateEdit`.

## Conventions

- Errors are returned by libraries; the UI logs via `slog` or shows a
  notification.
- Mutate state through pointer receivers; caches invalidate explicitly.
- No comments explaining *what*; names carry intent.
