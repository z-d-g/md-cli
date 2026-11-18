package markdown

import (
	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/text"
	"github.com/yuin/goldmark/v2/util"
)

// KindUnderline is the node kind for ++underline++ spans.
var KindUnderline = ast.NewNodeKind("Underline")

// Underline is the AST node for ++text++.
type Underline struct {
	ast.BaseInline
}

// Kind implements ast.Node.
func (n *Underline) Kind() ast.NodeKind { return KindUnderline }

// Dump implements ast.Node.
func (n *Underline) Dump(_ []byte) *ast.NodeDump { return ast.NewNodeDump(n, nil) }

// NewUnderline returns a new Underline node.
func NewUnderline() *Underline {
	n := &Underline{}
	n.Init(n)
	return n
}

type underlineDelimiterProcessor struct{}

func (p *underlineDelimiterProcessor) IsDelimiter(b byte) bool { return b == '+' }

func (p *underlineDelimiterProcessor) CanOpenCloser(opener, closer *parser.Delimiter) bool {
	return opener.Char == closer.Char
}

func (p *underlineDelimiterProcessor) OnMatch(int) ast.Node { return NewUnderline() }

var underlineProcessor = &underlineDelimiterProcessor{}

type underlineParser struct{}

func (s *underlineParser) Trigger() []byte { return []byte{'+'} }

func (s *underlineParser) Parse(_ ast.Node, block text.Reader, pc parser.Context) ast.Node {
	if block.PrecedingCharacter() == '+' {
		return nil
	}
	line, _ := block.PeekLine()
	n := 0
	for n < len(line) && line[n] == '+' {
		n++
	}
	if n != 2 {
		return nil
	}
	return parser.ParseDelimiter(block, 1, underlineProcessor, pc)
}

// NewUnderlineParser returns a parser.Extension enabling ++underline++.
func NewUnderlineParser() parser.Extension { return &underlineExtension{} }

type underlineExtension struct{}

func (e *underlineExtension) ParserOptions(*parser.Config) []parser.Option {
	return []parser.Option{
		parser.WithInlineParsers(util.Prioritized[parser.InlineParser](&underlineParser{}, 200)),
	}
}
