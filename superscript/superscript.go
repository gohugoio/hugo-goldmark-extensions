package superscript

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// A Superscript struct represents a superscript text.
type Superscript struct {
	ast.BaseInline
}

// Dump implements Node.Dump.
func (n *Superscript) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

// KindSuperscript is a NodeKind of the Superscript node.
var KindSuperscript = ast.NewNodeKind("Superscript")

// Kind implements Node.Kind.
func (n *Superscript) Kind() ast.NodeKind {
	return KindSuperscript
}

// NewSuperscript returns a new Superscript node.
func NewSuperscript() *Superscript {
	return &Superscript{}
}

type superscriptDelimiterProcessor struct {
}

func (p *superscriptDelimiterProcessor) IsDelimiter(b byte) bool {
	return b == '^'
}

func (p *superscriptDelimiterProcessor) CanOpenCloser(opener, closer *parser.Delimiter) bool {
	return opener.Char == closer.Char
}

func (p *superscriptDelimiterProcessor) OnMatch(int) ast.Node {
	return NewSuperscript()
}

var defaultSuperscriptDelimiterProcessor = &superscriptDelimiterProcessor{}

type superscriptParser struct {
}

var defaultSuperscriptParser = &superscriptParser{}

// NewSuperscriptParser return a new InlineParser that parses
// superscript expressions.
func NewSuperscriptParser() parser.InlineParser {
	return defaultSuperscriptParser
}

func (s *superscriptParser) Trigger() []byte {
	return []byte{'^'}
}

func (s *superscriptParser) Parse(_ ast.Node, block text.Reader, pc parser.Context) ast.Node {
	before := block.PrecendingCharacter()
	line, segment := block.PeekLine()
	node := parser.ScanDelimiter(line, before, 1, defaultSuperscriptDelimiterProcessor)
	if node == nil {
		return nil
	}
	if node.CanOpen {
		for i := 1; i < len(line); i++ {
			c := line[i]
			if c == line[0] { // Found closing match
				break
			}
			if util.IsSpace(c) {
				return nil
			} // No ordinary whitespaces allowed
		}
	}
	node.Segment = segment.WithStop(segment.Start + node.OriginalLength)
	block.Advance(node.OriginalLength)
	pc.PushDelimiter(node)
	return node
}

func (s *superscriptParser) CloseBlock() {
	// nothing to do
}

// HTMLRenderer is a renderer.NodeRenderer implementation that renders
// Superscript nodes.
type HTMLRenderer struct {
	html.Config
}

// NewSuperscriptHTMLRenderer returns a new SuperscriptHTMLRenderer.
func NewSuperscriptHTMLRenderer(opts ...html.Option) renderer.NodeRenderer {
	r := &HTMLRenderer{
		Config: html.NewConfig(),
	}
	for _, opt := range opts {
		opt.SetHTMLOption(&r.Config)
	}
	return r
}

// RegisterFuncs implements renderer.NodeRenderer.RegisterFuncs.
func (r *HTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindSuperscript, r.renderSuperscript)
}

// AttributeFilter defines attribute names which dd elements can have.
var AttributeFilter = html.GlobalAttributeFilter

func (r *HTMLRenderer) renderSuperscript(
	w util.BufWriter, _ []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		if n.Attributes() != nil {
			_, _ = w.WriteString("<sup")
			html.RenderAttributes(w, n, AttributeFilter)
			_ = w.WriteByte('>')
		} else {
			_, _ = w.WriteString("<sup>")
		}
	} else {
		_, _ = w.WriteString("</sup>")
	}
	return ast.WalkContinue, nil
}

// Extension allows you to use a superscript expression like 'x^2^'.
var Extension = &Superscript{}

func (n *Superscript) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(NewSuperscriptParser(), 600),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewSuperscriptHTMLRenderer(), 600),
	))
}
