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

type parserWithDelimiters interface {
	delimiters() []delimiters
}

// Determine if a byte array starts with a given string
func startsWith(b []byte, s string) bool {
	if len(b) < len(s) {
		return false
	}

	return string(b[:len(s)]) == s
}

// Determine if the input slice starts with a full valid opening delimiter.
// If so, returns the delimiter struct, otherwise returns nil.
func GetFullOpeningDelimiter(s parserWithDelimiters, line []byte) *delimiters {
	for _, d := range s.delimiters() {
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
func OpenersFirstByte(s parserWithDelimiters) []byte {
	var firstBytes []byte
	for _, d := range s.delimiters() {
		firstBytes = append(firstBytes, d.Open[0])
	}
	return firstBytes
}

// ---- Inline Parser ----

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

// Implements parserWithDelimiters for inlinePassthroughParser
func (s *inlinePassthroughParser) delimiters() []delimiters {
	return s.PassthroughDelimiters
}

func NewInlinePassthroughParser(ds []delimiters) parser.InlineParser {
	return &inlinePassthroughParser{
		PassthroughDelimiters: ds,
	}
}

func (s *inlinePassthroughParser) Trigger() []byte {
	return OpenersFirstByte(s)
}

func (s *inlinePassthroughParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	// In order to prevent other parser extensions from operating on the text
	// between passthrough delimiters, we must process the entire inline
	// passthrough in one execution of Parse. This means we can't use the style
	// of multiple triggers with parser.Context state saved between calls.
	line, startSegment := block.PeekLine()

	fencePair := GetFullOpeningDelimiter(s, line)
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

// RegisterFuncs implements renderer.NodeRenderer.RegisterFuncs.
func (r *passthroughInlineRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindPassthroughInline, r.renderRawInline)
}

func NewPassthroughInlineRenderer() renderer.NodeRenderer {
	return &passthroughInlineRenderer{}
}

// ---- Extension and config ----

type passthrough struct {
	InlineDelimiters []delimiters
}

func NewPassthroughWithDelimiters(Delimiters []delimiters) goldmark.Extender {
	return &passthrough{
		InlineDelimiters: Delimiters,
	}
}

func (e *passthrough) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithInlineParsers(
			util.Prioritized(NewInlinePassthroughParser(e.InlineDelimiters), 201),
		),
	)

	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewPassthroughInlineRenderer(), 101),
	))
}
