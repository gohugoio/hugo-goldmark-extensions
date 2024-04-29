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

type inlineTagDelimiterProcessor struct {
	ast.InlineTag
}

func newInlineTagDelimiterProcessor(tag ast.InlineTag) parser.DelimiterProcessor {
	return &inlineTagDelimiterProcessor{tag}
}

func (p *inlineTagDelimiterProcessor) IsDelimiter(b byte) bool {
	return b == p.Char
}

func (p *inlineTagDelimiterProcessor) CanOpenCloser(opener, closer *parser.Delimiter) bool {
	return opener.Char == closer.Char
}

func (p *inlineTagDelimiterProcessor) OnMatch(_ int) gast.Node {
	return ast.NewInlineTag(p.InlineTag)
}

type inlineTagParser struct {
	ast.InlineTag
}

func newInlineTagParser(tag ast.InlineTag) parser.InlineParser {
	return &inlineTagParser{InlineTag: tag}
}

// Trigger implements parser.InlineParser.
func (s *inlineTagParser) Trigger() []byte {
	return []byte{s.Char}
}

// Parse implements the parser.InlineParser for all types of InlineTags.
func (s *inlineTagParser) Parse(_ gast.Node, block text.Reader, pc parser.Context) gast.Node {
	before := block.PrecendingCharacter()
	line, segment := block.PeekLine()
	node := parser.ScanDelimiter(line, before, s.Number, newInlineTagDelimiterProcessor(s.InlineTag))
	if node == nil {
		return nil
	}
	if !s.WhitespaceAllowed && node.CanOpen && hasSpace(line) {
		if !(node.CanClose && pc.LastDelimiter() != nil && pc.LastDelimiter().Char == node.Char) {
			return nil
		}
	}
	node.Segment = segment.WithStop(segment.Start + node.OriginalLength)
	block.Advance(node.OriginalLength)
	pc.PushDelimiter(node)
	return node
}

// Check if there is an ordinary white space in the line before the next marker
func hasSpace(line []byte) bool {
	marker := line[0]
	for i := 1; i < len(line); i++ {
		c := line[i]
		if c == marker {
			break
		}
		if util.IsSpace(c) {
			return true
		}
	}
	return false
}

type inlineTagHTMLRenderer struct {
	htmlTag string
	tagType ast.InlineTagType
	html.Config
}

// newInlineTagHTMLRenderer returns a new NodeRenderer that renders InlineTagNode nodes to HTML.
func newInlineTagHTMLRenderer(tag ast.InlineTag, opts ...html.Option) renderer.NodeRenderer {
	r := &inlineTagHTMLRenderer{
		htmlTag: tag.Html,
		tagType: tag.TagType,
		Config:  html.NewConfig(),
	}
	for _, opt := range opts {
		opt.SetHTMLOption(&r.Config)
	}
	return r
}

// RegisterFuncs registers rendering functions to the given NodeRendererFuncRegisterer.
func (r *inlineTagHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.NewInlineTagNodeKind(r.tagType), r.renderInlineTag)
}

// inlineTagAttributeFilter is a global filter for attributes.
var inlineTagAttributeFilter = html.GlobalAttributeFilter

// renderInlineTag renders an inline tag.
func (r *inlineTagHTMLRenderer) renderInlineTag(
	w util.BufWriter, _ []byte, n gast.Node, entering bool) (gast.WalkStatus, error) {
	if entering {
		_ = w.WriteByte('<')
		_, _ = w.WriteString(r.htmlTag)
		if n.Attributes() != nil {
			html.RenderAttributes(w, n, inlineTagAttributeFilter)
		}
	} else {
		_, _ = w.WriteString("</")
		_, _ = w.WriteString(r.htmlTag)
	}
	_ = w.WriteByte('>')
	return gast.WalkContinue, nil
}

// inlineTag is a general inline tag parser and renderer.
type inlineTag struct {
	ast.InlineTag
}

// Superscript is an inline tag parser and renderer for superscript text.
var Superscript = &inlineTag{ast.SuperscriptTag}

// Subscript is an inline tag parser and renderer for subscript text.
var Subscript = &inlineTag{ast.SubscriptTag}

// Insert is an inline tag parser and renderer for inserted text.
var Insert = &inlineTag{ast.InsertTag}

// Mark is an inline tag parser and renderer for marked text.
var Mark = &inlineTag{ast.MarkTag}

// Extend adds inline tags to the Markdown parser and renderer.
func (n *inlineTag) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(newInlineTagParser(n.InlineTag), n.ParsePriority),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(newInlineTagHTMLRenderer(n.InlineTag), n.RenderPriority),
	))
}
