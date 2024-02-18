package insert

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// A Insert struct represents a insert text segment.
type Insert struct {
	ast.BaseInline
}

// Dump implements Node.Dump.
func (n *Insert) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

// Kind is a NodeKind of the Insert node.
var Kind = ast.NewNodeKind("Insert")

// Kind implements Node.Kind.
func (n *Insert) Kind() ast.NodeKind {
	return Kind
}

// New returns a new Insert node.
func New() *Insert {
	return &Insert{}
}

type delimiterProcessor struct {
}

func (p *delimiterProcessor) IsDelimiter(b byte) bool {
	return b == '+'
}

func (p *delimiterProcessor) CanOpenCloser(opener, closer *parser.Delimiter) bool {
	return opener.Char == closer.Char
}

func (p *delimiterProcessor) OnMatch(int) ast.Node {
	return New()
}

var defaultDelimiterProcessor = &delimiterProcessor{}

type Parser struct {
}

var defaultParser = &Parser{}

// NewParser returns a new InlineParser that parses
// insert expressions.
func NewParser() parser.InlineParser {
	return defaultParser
}

func (s *Parser) Trigger() []byte {
	return []byte{'+'}
}

func (s *Parser) Parse(_ ast.Node, block text.Reader, pc parser.Context) ast.Node {
	before := block.PrecendingCharacter()
	line, segment := block.PeekLine()
	node := parser.ScanDelimiter(line, before, 2, defaultDelimiterProcessor)
	if node == nil {
		return nil
	}
	node.Segment = segment.WithStop(segment.Start + node.OriginalLength)
	block.Advance(node.OriginalLength)
	pc.PushDelimiter(node)
	return node
}

func (s *Parser) CloseBlock() {
	// nothing to do
}

// HTMLRenderer is a renderer.NodeRenderer implementation that renders
// Insert nodes.
type HTMLRenderer struct {
	html.Config
}

// NewHTMLRenderer returns a new HTMLRenderer.
func NewHTMLRenderer(opts ...html.Option) renderer.NodeRenderer {
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
	reg.Register(Kind, r.render)
}

// AttributeFilter defines attribute names which dd elements can have.
var AttributeFilter = html.GlobalAttributeFilter

func (r *HTMLRenderer) render(
	w util.BufWriter, _ []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		if n.Attributes() != nil {
			_, _ = w.WriteString("<ins")
			html.RenderAttributes(w, n, AttributeFilter)
			_ = w.WriteByte('>')
		} else {
			_, _ = w.WriteString("<ins>")
		}
	} else {
		_, _ = w.WriteString("</ins>")
	}
	return ast.WalkContinue, nil
}

// The Extension allows you to use an insert expression like '++insertion++'.
var Extension = &Insert{}

func (n *Insert) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(NewParser(), 501),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewHTMLRenderer(), 501),
	))
}
