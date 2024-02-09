package superscript

import (
	"github.com/gohugoio/hugo-goldmark-extensions/superscript/ast"
	"github.com/yuin/goldmark"
	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type superscriptDelimiterProcessor struct {
}

func (p *superscriptDelimiterProcessor) IsDelimiter(b byte) bool {
	return b == '^'
}

func (p *superscriptDelimiterProcessor) CanOpenCloser(opener, closer *parser.Delimiter) bool {
	return opener.Char == closer.Char
}

func (p *superscriptDelimiterProcessor) OnMatch(int) gast.Node {
	return ast.NewSuperscript()
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

func (s *superscriptParser) Parse(_ gast.Node, block text.Reader, pc parser.Context) gast.Node {
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

// HTMLRenderer is a renderer.NodeRenderer implementation that
// renders Superscript nodes.
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
	reg.Register(ast.KindSuperscript, r.renderSuperscript)
}

// AttributeFilter defines attribute names which dd elements can have.
var AttributeFilter = html.GlobalAttributeFilter

func (r *HTMLRenderer) renderSuperscript(
	w util.BufWriter, _ []byte, n gast.Node, entering bool) (gast.WalkStatus, error) {
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
	return gast.WalkContinue, nil
}

type superscript struct {
}

// Superscript is an extension that allows you to use a superscript expression like 'x^2^'.
var Superscript = &superscript{}

func (e *superscript) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(NewSuperscriptParser(), 600),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewSuperscriptHTMLRenderer(), 600),
	))
}
