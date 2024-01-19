package passthrough

import (
	"bytes"
	"fmt"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"strings"
)

type delimiters struct {
	Open  string
	Close string
}

// Determine if a byte array starts with a given string
func startsWith(b []byte, s string) bool {
	if len(b) < len(s) {
		return false
	}

	return string(b[:len(s)]) == s
}

type PassthroughInline struct {
	ast.BaseInline

	// The segment of text that this inline passthrough represents.
	Segment text.Segment
}

func NewPassthroughInline(segment text.Segment) *PassthroughInline {
	return &PassthroughInline{
		Segment: segment,
	}
}

// Text implements Node.Text.
func (n *PassthroughInline) Text(source []byte) []byte {
	return n.Segment.Value(source)
}

// Dump implements Node.Dump.
func (n *PassthroughInline) Dump(source []byte, level int) {
	indent := strings.Repeat("    ", level)
	fmt.Printf("%sPassthroughInline {\n", indent)
	indent2 := strings.Repeat("    ", level+1)
	fmt.Printf("%sSegment: \"%s\"\n", indent2, n.Text(source))
	fmt.Printf("%s}\n", indent)
}

// KindPassthroughInline is a NodeKind of the PassthroughInline node.
var KindPassthroughInline = ast.NewNodeKind("PassthroughInline")

// Kind implements Node.Kind.
func (n *PassthroughInline) Kind() ast.NodeKind {
	return KindPassthroughInline
}

type inlinePassthroughParser struct {
	PassthroughDelimiters []delimiters
}

func NewInlinePassthroughParser(ds []delimiters) parser.InlineParser {
	return &inlinePassthroughParser{
		PassthroughDelimiters: ds,
	}
}

// Determine if the input slice starts with a full valid opening delimiter.
// If so, returns the delimiter struct, otherwise returns nil.
func GetFullOpeningDelimiter(delims []delimiters, line []byte) *delimiters {
	for _, d := range delims {
		if startsWith(line, d.Open) {
			return &d
		}
	}

	return nil
}

// Return an array of bytes containing the first byte of each opening
// delimiter. Used to populate the trigger list for inline and block parsers.
// `Parse` will be executed once for each character that is in this list of
// allowed trigger characters. Our parse function needs to do some additional
// checks because Trigger only works for single-byte delimiters.
func OpenersFirstByte(delims []delimiters) []byte {
	var firstBytes []byte
	for _, d := range delims {
		firstBytes = append(firstBytes, d.Open[0])
	}
	return firstBytes
}

func (s *inlinePassthroughParser) Trigger() []byte {
	return OpenersFirstByte(s.PassthroughDelimiters)
}

func (s *inlinePassthroughParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	// In order to prevent other parser extensions from operating on the text
	// between passthrough delimiters, we must process the entire inline
	// passthrough in one execution of Parse. This means we can't use the style
	// of multiple triggers with parser.Context state saved between calls.
	line, startSegment := block.PeekLine()

	fencePair := GetFullOpeningDelimiter(s.PassthroughDelimiters, line)
	// fencePair == nil can happen if only the first byte of an opening delimiter
	// matches, but it is not the complete opening delimiter. The trigger causes
	// this Parse function to execute, but the trigger interface is limited to
	// matching single bytes.
	if fencePair == nil {
		return nil
	}

	// This roughly follows goldmark/parser/code_span.go
	block.Advance(len(fencePair.Open))
	openerSize := len(fencePair.Open)
	l, pos := block.Position()

	for {
		line, lineSegment := block.PeekLine()
		if line == nil {
			block.SetPosition(l, pos)
			return ast.NewTextSegment(startSegment.WithStop(startSegment.Start + openerSize))
		}

		closingDelimiterPos := bytes.Index(line, []byte(fencePair.Close))
		if closingDelimiterPos == -1 { // no closer on this line
			block.AdvanceLine()
			continue
		}

		// This segment spans from the original starting trigger (including the delimiter)
		// up to and including the closing delimiter.
		seg := startSegment.WithStop(lineSegment.Start + closingDelimiterPos + len(fencePair.Close))
		if seg.Len() == len(fencePair.Open)+len(fencePair.Close) {
			return nil
		}

		block.Advance(closingDelimiterPos + len(fencePair.Close))
		return NewPassthroughInline(seg)
	}
}

type passthroughInlineRenderer struct {
}

func (r *passthroughInlineRenderer) renderRawInline(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.WriteString(string(n.Text(source)))
	}
	return ast.WalkContinue, nil
}

// ---- Block Parser ----

// A PassthroughBlock struct represents a fenced block of raw text to pass
// through unchanged.
// There is no built-in "raw text block" node in goldmark, and the closest
// thing is a code block, which emits `<pre><code>` tags. So we need a new
// node and a new block renderer.
type PassthroughBlock struct {
	ast.BaseBlock
}

// Dump implements Node.Dump.
func (n *PassthroughBlock) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

// KindPassthroughBlock is a NodeKind of the PassthroughBlock node.
var KindPassthroughBlock = ast.NewNodeKind("PassthroughBlock")

// Kind implements Node.Kind.
func (n *PassthroughBlock) Kind() ast.NodeKind {
	return KindPassthroughBlock
}

// NewPassthroughBlock return a new PassthroughBlock node.
func NewPassthroughBlock() *PassthroughBlock {
	return &PassthroughBlock{
		BaseBlock: ast.BaseBlock{},
	}
}

// Block parsing happens across different interface methods, and the initial
// Open detects the fence pair to use by its opening delimiter. This needs to
// be preserved for the Continue and Close methods to have access to the
// corresponding closing delimiter.
var passthroughParserStateKey = parser.NewContextKey()

type passthroughParserState struct {
	DetectedDelimiters *delimiters
}

type blockPassthroughParser struct {
	PassthroughDelimiters []delimiters
}

// Implements parserWithDelimiters for blockPassthroughParser
func (b *blockPassthroughParser) delimiters() []delimiters {
	return b.PassthroughDelimiters
}

func NewBlockPassthroughParser(ds []delimiters) parser.BlockParser {
	return &blockPassthroughParser{
		PassthroughDelimiters: ds,
	}
}

func (b *blockPassthroughParser) Trigger() []byte {
	return OpenersFirstByte(b.PassthroughDelimiters)
}

func (b *blockPassthroughParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, segment := reader.PeekLine()
	fencePair := GetFullOpeningDelimiter(b.PassthroughDelimiters, line)
	// fencePair == nil can happen if only the first byte of an opening delimiter
	// matches, but it is not the complete opening delimiter.
	if fencePair == nil {
		return nil, parser.NoChildren
	}
	node := NewPassthroughBlock()
	pc.Set(passthroughParserStateKey, &passthroughParserState{DetectedDelimiters: fencePair})

	node.Lines().Append(segment)
	reader.Advance(segment.Len() - 1)
	return node, parser.NoChildren
}

func (b *blockPassthroughParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	// currentState cannot be nil or else Continue was triggered without Open
	// successfully creating a new node.
	currentState := pc.Get(passthroughParserStateKey).(*passthroughParserState)
	fencePair := currentState.DetectedDelimiters
	line, segment := reader.PeekLine()

	closingDelimiterPos := bytes.Index(line, []byte(fencePair.Close))
	if closingDelimiterPos == -1 { // no closer on this line
		node.Lines().Append(segment)
		reader.Advance(segment.Len() - 1)
		return parser.Continue | parser.NoChildren
	}

	// This segment spans up to and including the closing delimiter.
	seg := segment.WithStop(segment.Start + closingDelimiterPos + len(fencePair.Close))
	node.Lines().Append(seg)
	reader.Advance(closingDelimiterPos + len(fencePair.Close))

	return parser.Close
}

func (b *blockPassthroughParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
}

func (b *blockPassthroughParser) CanInterruptParagraph() bool {
	return false
}

func (b *blockPassthroughParser) CanAcceptIndentedLine() bool {
	return true
}

type passthroughBlockRenderer struct {
}

func (r *passthroughBlockRenderer) renderRawBlock(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		l := n.Lines().Len()
		for i := 0; i < l; i++ {
			line := n.Lines().At(i)
			w.WriteString(string(line.Value(source)))
		}
		w.WriteString("\n")
	}
	return ast.WalkSkipChildren, nil
}

// RegisterFuncs implements renderer.NodeRenderer.RegisterFuncs.
func (r *passthroughInlineRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindPassthroughInline, r.renderRawInline)
}

// RegisterFuncs implements renderer.NodeRenderer.RegisterFuncs.
func (r *passthroughBlockRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindPassthroughBlock, r.renderRawBlock)
}

func NewPassthroughInlineRenderer() renderer.NodeRenderer {
	return &passthroughInlineRenderer{}
}

func NewPassthroughBlockRenderer() renderer.NodeRenderer {
	return &passthroughBlockRenderer{}
}

// ---- Extension and config ----

type passthrough struct {
	InlineDelimiters []delimiters
	BlockDelimiters  []delimiters
}

func NewPassthroughWithDelimiters(
	InlineDelimiters []delimiters,
	BlockDelimiters []delimiters) goldmark.Extender {
	return &passthrough{
		InlineDelimiters: InlineDelimiters,
		BlockDelimiters:  BlockDelimiters,
	}
}

func (e *passthrough) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithInlineParsers(
			util.Prioritized(NewInlinePassthroughParser(e.InlineDelimiters), 201),
		),
		parser.WithBlockParsers(
			util.Prioritized(NewBlockPassthroughParser(e.BlockDelimiters), 99),
		),
	)

	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewPassthroughInlineRenderer(), 101),
		util.Prioritized(NewPassthroughBlockRenderer(), 99),
	))
}
