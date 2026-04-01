package render

import (
	"strings"

	"github.com/z-d-g/md-cli/internal/markdown"
)

// PrintRenderer renders all lines of a markdown document using a LineRenderer.
// It tracks fenced-code-block state across lines so RenderLine receives the
// same isInCodeBlock flag the editor computes per frame.
type PrintRenderer struct {
	renderer LineRenderer
}

// NewPrintRenderer creates a print renderer that reuses the given LineRenderer.
func NewPrintRenderer(r LineRenderer) *PrintRenderer {
	return &PrintRenderer{renderer: r}
}

// RenderDocument renders the full markdown content to a styled string.
// Output matches the editor's rendered view line-for-line.
func (p *PrintRenderer) RenderDocument(content string) string {
	lines := strings.Split(content, "\n")

	// Provide document lines for table width pre-computation
	SetTableLines(func() []string { return lines })
	defer SetTableLines(nil)

	var b strings.Builder
	b.Grow(len(content) * 2)

	inCodeBlock := false
	for i, line := range lines {
		b.WriteString(p.renderer.RenderLine(line, inCodeBlock))
		if markdown.IsCodeFence(line) {
			inCodeBlock = !inCodeBlock
		}
		if i < len(lines)-1 {
			b.WriteByte('\n')
		}
	}

	return b.String()
}
