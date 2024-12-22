package extras

import (
	"slices"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type inlineTagDelimiterProcessor struct {
	inlineTag
}

func newInlineTagDelimiterProcessor(tag inlineTag) parser.DelimiterProcessor {
	return &inlineTagDelimiterProcessor{tag}
}

func (p *inlineTagDelimiterProcessor) IsDelimiter(b byte) bool {
	return b == p.Char
}

func (p *inlineTagDelimiterProcessor) CanOpenCloser(opener, closer *parser.Delimiter) bool {
	return opener.Char == closer.Char
}

func (p *inlineTagDelimiterProcessor) OnMatch(_ int) ast.Node {
	return newInlineTag(p.inlineTag)
}

type inlineTagParser struct {
	inlineTag
}

func newInlineTagParser(tag inlineTag) parser.InlineParser {
	return &inlineTagParser{inlineTag: tag}
}

// Trigger implements parser.InlineParser.
func (s *inlineTagParser) Trigger() []byte {
	return []byte{s.Char}
}

// Parse implements the parser.InlineParser for all types of InlineTags.
func (s *inlineTagParser) Parse(_ ast.Node, block text.Reader, pc parser.Context) ast.Node {
	before := block.PrecendingCharacter()
	line, segment := block.PeekLine()

	// Issue 30
	modifiedLine := slices.Clone(line)
	if s.inlineTag.TagKind == kindSuperscript && len(line) > s.Number {
		symbols := []byte{'+', '-', '\''}
		if slices.Contains(symbols, line[s.Number]) {
			modifiedLine[s.Number] = 'z' // replace with any letter or number
		}
	}

	node := parser.ScanDelimiter(modifiedLine, before, s.Number, newInlineTagDelimiterProcessor(s.inlineTag))
	if node == nil || node.OriginalLength > 2 || before == rune(s.Char) {
		return nil
	}
	node.Segment = segment.WithStop(segment.Start + node.OriginalLength)
	block.Advance(node.OriginalLength)
	pc.PushDelimiter(node)
	return node
}

type inlineTagHTMLRenderer struct {
	htmlTag string
	tagKind ast.NodeKind
	html.Config
}

// newInlineTagHTMLRenderer returns a new NodeRenderer that renders Inline nodes to HTML.
func newInlineTagHTMLRenderer(tag inlineTag, opts ...html.Option) renderer.NodeRenderer {
	r := &inlineTagHTMLRenderer{
		htmlTag: tag.Html,
		tagKind: tag.TagKind,
		Config:  html.NewConfig(),
	}
	for _, opt := range opts {
		opt.SetHTMLOption(&r.Config)
	}
	return r
}

// RegisterFuncs registers rendering functions to the given NodeRendererFuncRegisterer.
func (r *inlineTagHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(r.tagKind, r.renderInlineTag)
}

// inlineTagAttributeFilter is a global filter for attributes.
var inlineTagAttributeFilter = html.GlobalAttributeFilter

// renderInlineTag renders an inline tag.
func (r *inlineTagHTMLRenderer) renderInlineTag(
	w util.BufWriter, _ []byte, n ast.Node, entering bool,
) (ast.WalkStatus, error) {
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
	return ast.WalkContinue, nil
}

// inlineExtension is an extension that adds inline tags to the Markdown parser and renderer.
type inlineExtension struct {
	conf Config
}

// Config configures the extras extension.
type Config struct {
	Superscript SuperscriptConfig
	Subscript   SubscriptConfig
	Insert      InsertConfig
	Mark        MarkConfig
	Delete      DeleteConfig
}

// SuperscriptConfig configures the superscript extension.
type SuperscriptConfig struct {
	Enable bool
}

// SubscriptConfig configures the subscript extension.
type SubscriptConfig struct {
	Enable bool
}

// InsertConfig configures the insert extension.
type InsertConfig struct {
	Enable bool
}

// MarkConfig configures the mark extension.
type MarkConfig struct {
	Enable bool
}

type DeleteConfig struct {
	Enable bool
}

// New returns a new inline tag extension.

func New(config Config) goldmark.Extender {
	return &inlineExtension{
		conf: config,
	}
}

// Extend adds inline tags to the Markdown parser and renderer.
func (tag *inlineExtension) Extend(md goldmark.Markdown) {
	addTag := func(tag inlineTag) {
		md.Parser().AddOptions(parser.WithInlineParsers(
			util.Prioritized(newInlineTagParser(tag), tag.ParsePriority),
		))
		md.Renderer().AddOptions(renderer.WithNodeRenderers(
			util.Prioritized(newInlineTagHTMLRenderer(tag), tag.RenderPriority),
		))
	}
	if tag.conf.Superscript.Enable {
		addTag(superscriptTag)
	}
	if tag.conf.Subscript.Enable {
		addTag(subscriptTag)
	}
	if tag.conf.Insert.Enable {
		addTag(insertTag)
	}
	if tag.conf.Mark.Enable {
		addTag(markTag)
	}
	if tag.conf.Delete.Enable {
		addTag(deleteTag)
	}
}
