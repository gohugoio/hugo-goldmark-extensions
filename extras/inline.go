package extras

import (
	"io"

	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer"
	"github.com/yuin/goldmark/v2/renderer/html"
	"github.com/yuin/goldmark/v2/text"
	"github.com/yuin/goldmark/v2/util"
)

type inlineTagDelimiterProcessor struct {
	InlineTag
}

func newInlineTagDelimiterProcessor(tag InlineTag) parser.DelimiterProcessor {
	return &inlineTagDelimiterProcessor{tag}
}

func (p *inlineTagDelimiterProcessor) IsDelimiter(b byte) bool {
	return b == p.Char
}

func (p *inlineTagDelimiterProcessor) CanOpenCloser(opener, closer *parser.Delimiter) bool {
	return opener.Char == closer.Char
}

func (p *inlineTagDelimiterProcessor) OnMatch(_ int) ast.Node {
	return newInlineTag(p.InlineTag)
}

type inlineTagParser struct {
	InlineTag
	proc parser.DelimiterProcessor
}

func newInlineTagParser(tag InlineTag) parser.InlineParser {
	return &inlineTagParser{InlineTag: tag, proc: newInlineTagDelimiterProcessor(tag)}
}

// Trigger implements parser.InlineParser.
func (s *inlineTagParser) Trigger() []byte {
	return []byte{s.Char}
}

// Parse implements the parser.InlineParser for all types of InlineTags.
func (s *inlineTagParser) Parse(_ ast.Node, block text.Reader, pc parser.Context) ast.Node {
	before := block.PrecedingCharacter()
	if before == rune(s.Char) {
		return nil
	}

	line, _ := block.PeekLine()
	// Length of the delimiter run: at least Number, at most 2 (these tags are
	// one- or two-character delimiters; a longer run is not one of ours).
	n := 0
	for n < len(line) && line[n] == s.Char {
		n++
	}
	if n < s.Number || n > 2 {
		return nil
	}

	// Issue 30: a superscript whose content begins with one of + - ' — the
	// punctuation right after the caret makes the run non-left-flanking, so
	// ParseDelimiter would refuse to open it. Force it open in that case (this is
	// the v2 equivalent of the old trick of swapping that byte for a letter before
	// scanning; here CanOpen is exported so we can set it directly instead).
	forceOpen := s.TagKind == KindSuperscript && s.Number < len(line) &&
		(line[s.Number] == '+' || line[s.Number] == '-' || line[s.Number] == '\'')

	node := parser.ParseDelimiter(block, s.Number, s.proc, pc)
	if node == nil {
		return nil
	}
	if forceOpen {
		node.CanOpen = true
	}
	return node
}

// inlineTagAttributeFilter is a global filter for attributes.
var inlineTagAttributeFilter = html.GlobalAttributeFilter

// renderInlineTag renders any inline tag node: it reads the HTML tag name off the
// node itself, so one function serves every configured tag kind.
func renderInlineTag(writer io.Writer, _ []byte, n ast.Node, entering bool, _ renderer.Context) (ast.WalkStatus, error) {
	w := writer.(util.BufWriter)
	tag := n.(*inlineTagNode)
	if entering {
		_ = w.WriteByte('<')
		_, _ = w.WriteString(tag.Html)
		if n.Attributes() != nil {
			html.RenderAttributes(w, n, inlineTagAttributeFilter)
		}
	} else {
		_, _ = w.WriteString("</")
		_, _ = w.WriteString(tag.Html)
	}
	_ = w.WriteByte('>')
	return ast.WalkContinue, nil
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

// enabledTags returns the InlineTags enabled by the config, in registration order.
func (c Config) enabledTags() []InlineTag {
	var tags []InlineTag
	if c.Superscript.Enable {
		tags = append(tags, SuperscriptTag)
	}
	if c.Subscript.Enable {
		tags = append(tags, SubscriptTag)
	}
	if c.Insert.Enable {
		tags = append(tags, InsertTag)
	}
	if c.Mark.Enable {
		tags = append(tags, MarkTag)
	}
	if c.Delete.Enable {
		tags = append(tags, DeleteTag)
	}
	return tags
}

// inlineParserExtension adds the configured inline tag parsers to a parser.
type inlineParserExtension struct {
	conf Config
}

// NewParser returns a parser.Extension that parses the configured inline tags.
// Add it with parser.WithExtensions.
func NewParser(config Config) parser.Extension {
	return &inlineParserExtension{conf: config}
}

// ParserOptions implements parser.Extension.
func (e *inlineParserExtension) ParserOptions(_ *parser.Config) []parser.Option {
	var opts []parser.Option
	for _, tag := range e.conf.enabledTags() {
		opts = append(opts, parser.WithInlineParsers(
			util.Prioritized(newInlineTagParser(tag), tag.ParsePriority),
		))
	}
	return opts
}

// inlineHTMLRenderer renders the configured inline tags to HTML.
type inlineHTMLRenderer struct {
	conf Config
}

// NewHTMLRenderer returns an html.Extension that renders the configured inline
// tags. Add it with html.WithExtensions.
func NewHTMLRenderer(config Config) html.Extension {
	return &inlineHTMLRenderer{conf: config}
}

// RendererOptions implements html.Extension.
func (r *inlineHTMLRenderer) RendererOptions(_ *html.Config) []html.Option {
	renderers := map[ast.NodeKind]html.NodeRenderer{}
	for _, tag := range r.conf.enabledTags() {
		renderers[tag.TagKind] = html.NodeRendererFunc(renderInlineTag)
	}
	return []html.Option{html.WithNodeRenderers(renderers)}
}
