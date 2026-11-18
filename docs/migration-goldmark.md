# goldmark v2 migration

Status: **complete**. `internal/markdown/` parses with goldmark v2.

## Why

The previous parser was hand-written (~2,570 LOC), not CommonMark-compliant,
and accumulated edge-case bugs: fence length, delimiter runs, escapes, table
separators, span offsets. goldmark is fuzzed, CommonMark 0.31.2 compliant, and
exposes byte offsets per node.

Stayed in-house: cursor-aware, per-line source/rendered switching. No markdown
library can express "show raw markdown because the cursor is on line 7".

## Layout

```
markdown/
├── doc.go        Doc: parse once, answer per-line questions (kind, bounds, code)
├── inline.go     ParseInlineElements, FindSyntaxSpans (goldmark AST → elements)
├── delimiter.go  delimiter-run helpers used for spans and source mode
├── scan.go       line predicates (fence, heading, table, list, hr)
├── underline.go  ++underline++ extension (not in CommonMark)
└── types.go      InlineElement, SyntaxSpan, LineXxx, Fence
```

## API notes

```go
p := parser.New(parser.WithExtensions(...))
root := p.Parse(src) // ast.Node
```

- `goldmark.New`, `goldmark.Convert`, `WithExtensions` do not exist in v2.
- Extensions are constructors: `extension.NewTableParser()`,
  `NewStrikethroughParser()`, `NewTaskListItemParser()`, `NewLinkifyParser()`.
  `extension.GFM` is HTML-only.
- `Lines()` is not on `ast.Node`. Use `ast.BlockNode.Source()`.
- `*ast.Text.Value` is a `text.SingleLineValue` field, not a method.
- `*ast.CodeBlock.Value` is `text.Lines`; use `.Segments()`.
- `*ast.CodeSpan.Pos()` is the **content** offset, not the opening fence.
- Inline nodes live under a block (`Paragraph`/`Heading`), never at the root.

## Semantics that changed

- **Lazy continuation.** A line directly after a list item with no blank line is
  part of that item (CommonMark). Sepárate blocks with a blank line.
- **`***x***`** parses as `Emphasis(Strong(x))`, not a single bold-italic node.

## Performance

`Doc` parses once per content change and answers per-line queries in O(1).
`Editor.computeFrameState` rebuilds it only when the frame is dirty.
