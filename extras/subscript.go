package extras

import (
	"github.com/gohugoio/hugo-goldmark-extensions/extras/ast"
	"github.com/yuin/goldmark"
	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type subscriptDelimiterProcessor struct {
}

func (p *subscriptDelimiterProcessor) IsDelimiter(b byte) bool {
	return b == '~'
}

func (p *subscriptDelimiterProcessor) CanOpenCloser(opener, closer *parser.Delimiter) bool {
	return opener.Char == closer.Char
}

func (p *subscriptDelimiterProcessor) OnMatch(consumes int) gast.Node {
	return ast.NewSubscript()
}

var defaultSubscriptDelimiterProcessor = &subscriptDelimiterProcessor{}

type subscriptParser struct {
}

var defaultSubscriptParser = &subscriptParser{}

// NewSubscriptParser returns a new InlineParser that parses subscript expressions.
func NewSubscriptParser() parser.InlineParser {
	return defaultSubscriptParser
}

func (s *subscriptParser) Trigger() []byte {
	return []byte{'~'}
}

func (s *subscriptParser) Parse(parent gast.Node, block text.Reader, pc parser.Context) gast.Node {
	before := block.PrecendingCharacter()
	line, segment := block.PeekLine()
	node := parser.ScanDelimiter(line, before, 1, defaultSubscriptDelimiterProcessor)
	if node == nil || (node.CanOpen && hasSpace(line)) {
		return nil
	}
	node.Segment = segment.WithStop(segment.Start + node.OriginalLength)
	block.Advance(node.OriginalLength)
	pc.PushDelimiter(node)
	return node
}

func (s *subscriptParser) CloseBlock(parent gast.Node, pc parser.Context) {
	// nothing to do
}

// SubscriptHTMLRenderer is a renderer.NodeRenderer implementation that renders Subscript nodes.
type SubscriptHTMLRenderer struct {
	html.Config
}

// NewSubscriptHTMLRenderer returns a new SubscriptHTMLRenderer.
func NewSubscriptHTMLRenderer(opts ...html.Option) renderer.NodeRenderer {
	r := &SubscriptHTMLRenderer{
		Config: html.NewConfig(),
	}
	for _, opt := range opts {
		opt.SetHTMLOption(&r.Config)
	}
	return r
}

// RegisterFuncs implements renderer.NodeRenderer.RegisterFuncs.
func (r *SubscriptHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindSubscript, r.renderSubscript)
}

// SubscriptAttributeFilter defines attribute names.
var SubscriptAttributeFilter = html.GlobalAttributeFilter

func (r *SubscriptHTMLRenderer) renderSubscript(
	w util.BufWriter, source []byte, n gast.Node, entering bool) (gast.WalkStatus, error) {
	if entering {
		if n.Attributes() != nil {
			_, _ = w.WriteString("<sub")
			html.RenderAttributes(w, n, SubscriptAttributeFilter)
			_ = w.WriteByte('>')
		} else {
			_, _ = w.WriteString("<sub>")
		}
	} else {
		_, _ = w.WriteString("</sub>")
	}
	return gast.WalkContinue, nil
}

type subscript struct {
}

// Subscript is an extension that allows you to use a subscript expression like 'H~2~O'.
var Subscript = &subscript{}

func (e *subscript) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(NewSubscriptParser(), 600),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewSubscriptHTMLRenderer(), 600),
	))
}
