# Rendering

Package `render` turns markdown lines into styled strings with Lipgloss v2.

## LineRenderer

`LineRenderer` is the boundary between markdown and the terminal:

```go
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
```

`RenderLine` takes the document row so table state can be resolved without
matching line text. Pass `-1` for a context-free line.

## Styling

Styles come from `config.EditorStyles`; the renderer only reads them. Inline
rendering composes with `Style.Inherit`, one `Render` call per element, which
avoids nested ANSI corruption. Links are rendered in one pass with
`Hyperlink(url)` plus curly underline.

## Tables

Table rendering is pure. `tableDoc` indexes the document once per change into
per-block column widths, alignments and header flags. `renderTable(line, idx)`
looks the row up directly, so render order never matters and separate tables
never inherit each other's widths.

## Print mode

`PrintRenderer` renders a whole document with the same `LineRenderer` used by
the editor, tracking code-fence state. Output width comes from the terminal,
or 80 columns when stdout is not a TTY.
