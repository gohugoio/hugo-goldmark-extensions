package subscript

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// A Subscript struct represents a subscript text segment.
type Subscript struct {
	ast.BaseInline
}

// Dump implements Node.Dump.
func (n *Subscript) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

// Kind is a NodeKind of the Subscript node.
var Kind = ast.NewNodeKind("Subscript")

// Kind implements Node.Kind.
func (n *Subscript) Kind() ast.NodeKind {
	return Kind
}

// New returns a new Subscript node.
func New() *Subscript {
	return &Subscript{}
}

type delimiterProcessor struct {
}

func (p *delimiterProcessor) IsDelimiter(b byte) bool {
	return b == '~'
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

// NewParser returns a new InlineParser that parses subscript expressions.
func NewParser() parser.InlineParser {
	return defaultParser
}

func (s *Parser) Trigger() []byte {
	return []byte{'~'}
}

func (s *Parser) Parse(_ ast.Node, block text.Reader, pc parser.Context) ast.Node {
	before := block.PrecendingCharacter()
	line, segment := block.PeekLine()
	node := parser.ScanDelimiter(line, before, 1, defaultDelimiterProcessor)
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

func (s *Parser) CloseBlock() {
	// nothing to do
}

// HTMLRenderer is a renderer.NodeRenderer implementation that renders
// Subscript nodes.
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
			_, _ = w.WriteString("<sub")
			html.RenderAttributes(w, n, AttributeFilter)
			_ = w.WriteByte('>')
		} else {
			_, _ = w.WriteString("<sub>")
		}
	} else {
		_, _ = w.WriteString("</sub>")
	}
	return ast.WalkContinue, nil
}

// The Extension allows you to use a subscript expression like 'H~2~O'.
var Extension = &Subscript{}

func (n *Subscript) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(NewParser(), 501),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewHTMLRenderer(), 501),
	))
}
