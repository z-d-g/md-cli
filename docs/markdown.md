# Markdown parsing

Package `markdown` wraps [goldmark v2](https://github.com/yuin/goldmark) and
adds the few pieces md-cli needs: per-line classification, byte-accurate syntax
spans, and a `++underline++` extension.

## Inline parsing

Parsing is goldmark's. `ParseInlineElements` converts its AST into
`InlineElement` values, and `FindSyntaxSpans` reports the byte ranges of every
delimiter so the editor can show raw source.

- `**bold**`, `*italic*`, `***both***`, `` `code` ``, `[text](url)`,
  `![alt](url)`, `~~strike~~` come from goldmark.
- `++underline++` is a local goldmark extension (`underline.go`).
- Escapes follow CommonMark: `\*` is a literal asterisk.
- Code spans close only on a run of exactly the same length.

## Spans

`FindSyntaxSpans` returns byte ranges for source-mode styling: block markers
(heading, blockquote, list, horizontal rule) plus inline delimiters. Spans are
sorted by `(Start, End)`; `MergeSpans` unions overlaps.

## Code fences

goldmark's `*ast.CodeBlock` carries the fence kind, info string and content
segments. `Doc.markCode` maps that onto per-line membership and bounds, so a
shorter inner run stays literal content.

## Lines

`Doc` answers per-line questions without re-parsing:

```go
doc := markdown.NewDoc(content)
doc.LineKind(i)   // LineNormal, LineHeading1..6, LineCodeFence, LineBlockQuote
doc.InCode(i)     // code-block content line
doc.Bounds(i)     // inclusive [start,end] of the containing block
doc.IsTable(i)    // table membership
doc.IsTableSeparator(i)
doc.IsList(i)
doc.IsBlockquote(i)
```

## Line predicates

`scan.go` holds the cheap string checks used before any AST lookup:

```go
ParseFence(line) Fence      // Char, Len, Info, Valid
IsCodeFence(line)
IsHorizontalRule(line)
IsListLine(line)
IsHeadingLine(line)
IsTableLine(line)
IsTableSeparatorLine(line)
ClassifyLine(line, inCode)
CountLeadingHashes(line)
CountBlockquoteDepth(line)
```

## Documents

Parse once per content change; the editor caches the `Doc` in its frame state
and rebuilds it only when the buffer is dirty.
