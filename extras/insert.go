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

type insertDelimiterProcessor struct {
}

func (p *insertDelimiterProcessor) IsDelimiter(b byte) bool {
	return b == '+'
}

func (p *insertDelimiterProcessor) CanOpenCloser(opener, closer *parser.Delimiter) bool {
	return opener.Char == closer.Char
}

func (p *insertDelimiterProcessor) OnMatch(int) gast.Node {
	return ast.NewInsert()
}

var defaultInsertDelimiterProcessor = &insertDelimiterProcessor{}

type insertParser struct {
}

var defaultInsertParser = &insertParser{}

// NewInsertParser returns a new InlineParser that parses
// insert expressions.
func NewInsertParser() parser.InlineParser {
	return defaultInsertParser
}

func (s *insertParser) Trigger() []byte {
	return []byte{'+'}
}

func (s *insertParser) Parse(_ gast.Node, block text.Reader, pc parser.Context) gast.Node {
	before := block.PrecendingCharacter()
	line, segment := block.PeekLine()
	node := parser.ScanDelimiter(line, before, 2, defaultInsertDelimiterProcessor)
	if node == nil {
		return nil
	}
	node.Segment = segment.WithStop(segment.Start + node.OriginalLength)
	block.Advance(node.OriginalLength)
	pc.PushDelimiter(node)
	return node
}

func (s *insertParser) CloseBlock() {
	// nothing to do
}

// InsertHTMLRenderer is a renderer.NodeRenderer implementation that renders Insert nodes.
type InsertHTMLRenderer struct {
	html.Config
}

// NewHTMLRenderer returns a new InsertHTMLRenderer.
func NewHTMLRenderer(opts ...html.Option) renderer.NodeRenderer {
	r := &InsertHTMLRenderer{
		Config: html.NewConfig(),
	}
	for _, opt := range opts {
		opt.SetHTMLOption(&r.Config)
	}
	return r
}

// RegisterFuncs implements renderer.NodeRenderer.RegisterFuncs.
func (r *InsertHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindInsert, r.render)
}

// InsertAttributeFilter defines attribute names.
var InsertAttributeFilter = html.GlobalAttributeFilter

func (r *InsertHTMLRenderer) render(
	w util.BufWriter, _ []byte, n gast.Node, entering bool) (gast.WalkStatus, error) {
	if entering {
		if n.Attributes() != nil {
			_, _ = w.WriteString("<ins")
			html.RenderAttributes(w, n, InsertAttributeFilter)
			_ = w.WriteByte('>')
		} else {
			_, _ = w.WriteString("<ins>")
		}
	} else {
		_, _ = w.WriteString("</ins>")
	}
	return gast.WalkContinue, nil
}

type insert struct{}

// Insert is an extension that allows you to use an insert expression like '++insertion++'.
var Insert = &insert{}

func (n *insert) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(NewInsertParser(), 501),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewHTMLRenderer(), 501),
	))
}
