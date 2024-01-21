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

type Delimiters struct {
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

	// The matched delimiters
	Delimiters *Delimiters
}

func NewPassthroughInline(segment text.Segment, delimiters *Delimiters) *PassthroughInline {
	return &PassthroughInline{
		Segment:    segment,
		Delimiters: delimiters,
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
	PassthroughDelimiters []Delimiters
}

func NewInlinePassthroughParser(ds []Delimiters) parser.InlineParser {
	return &inlinePassthroughParser{
		PassthroughDelimiters: ds,
	}
}

// Determine if the input slice starts with a full valid opening delimiter.
// If so, returns the delimiter struct, otherwise returns nil.
func GetFullOpeningDelimiter(delims []Delimiters, line []byte) *Delimiters {
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
func OpenersFirstByte(delims []Delimiters) []byte {
	var firstBytes []byte
	containsBackslash := false
	for _, d := range delims {
		if d.Open[0] == '\\' {
			containsBackslash = true
		}
		firstBytes = append(firstBytes, d.Open[0])
	}

	if !containsBackslash {
		// always trigger on backslash because it can be used to escape the opening
		// delimiter.
		firstBytes = append(firstBytes, '\\')
	}
	return firstBytes
}

// Determine if the input list of delimiters contains the given delimiter pair
func ContainsDelimiters(delims []Delimiters, toFind *Delimiters) bool {
	for _, d := range delims {
		if d.Open == toFind.Open && d.Close == toFind.Close {
			return true
		}
	}

	return false
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
	// It can also be because the opening delimiter is escaped with a
	// double-backslash. In this case, we advance and return nil.
	if fencePair == nil {
		if len(line) > 2 && line[0] == '\\' && line[1] == '\\' {
			fencePair = GetFullOpeningDelimiter(s.PassthroughDelimiters, line[2:])
			if fencePair != nil {
				// Opening delimiter is escaped, return the escaped opener as plain text
				// So that the characters are not processed again.
				block.Advance(2 + len(fencePair.Open))
				return ast.NewTextSegment(startSegment.WithStop(startSegment.Start + len(fencePair.Open) + 2))
			}
		}
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
		return NewPassthroughInline(seg, fencePair)
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

// A PassthroughBlock struct represents a fenced block of raw text to pass
// through unchanged. This is not parsed directly, but emitted by an
// ASTTransformer that splits a paragraph at the point of an inline passthrough
// with the matching block delimiters.
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

// To support the use of passthrough block delimiters in inline contexts, I
// wasn't able to get the normal block parser to work. Goldmark seems to only
// trigger the inline parser when the trigger is not the first characters in a
// block. So instead we hook into the transformer interface, and process an
// inline passthrough after it's parsed, looking for nodes whose delimiters
// match the block delimiters, and splitting the paragraph at that point.
type passthroughInlineTransformer struct {
	BlockDelimiters []Delimiters
}

var PassthroughInlineTransformer = &passthroughInlineTransformer{}

func (p *passthroughInlineTransformer) Transform(
	doc *ast.Document, reader text.Reader, pc parser.Context) {

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		// Only match inline passthroughs that are direct descendants of
		// paragraphs. It's not clear what it would mean to have a block equation
		// rendered in, say, a list item, so in that case we just leave it as an
		// inline passthrough.
		if n.Kind() != KindPassthroughInline {
			return ast.WalkContinue, nil
		}
		if n.Parent() == nil || n.Parent().Kind() != ast.KindParagraph {
			return ast.WalkContinue, nil
		}

		inline := n.(*PassthroughInline)
		if !ContainsDelimiters(p.BlockDelimiters, inline.Delimiters) {
			return ast.WalkContinue, nil
		}
		paragraph := n.Parent().(*ast.Paragraph)
		parent := paragraph.Parent()
		var insertionPoint ast.Node
		insertionPoint = paragraph

		// Split the paragraph at this point
		precedingParagraph := ast.NewParagraph()
		for c := paragraph.FirstChild(); c != n && c != nil; c = c.NextSibling() {
			precedingParagraph.AppendChild(precedingParagraph, c)
		}
		for i := 0; i < paragraph.Lines().Len(); i++ {
			seg := paragraph.Lines().At(i)
			if seg.Stop > inline.Segment.Start {
				newSeg := seg.WithStop(inline.Segment.Start)
				if newSeg.Len() > 0 {
					precedingParagraph.Lines().Append(newSeg)
				}
				break
			}
			precedingParagraph.Lines().Append(seg)
		}
		if precedingParagraph.ChildCount() > 0 || precedingParagraph.Lines().Len() > 0 {
			parent.InsertAfter(parent, insertionPoint, precedingParagraph)
			insertionPoint = precedingParagraph
		}

		newBlock := NewPassthroughBlock()
		newBlock.Lines().Append(inline.Segment)
		parent.InsertAfter(parent, insertionPoint, newBlock)
		insertionPoint = newBlock

		succeedingParagraph := ast.NewParagraph()

		for c := n.NextSibling(); c != nil; c = c.NextSibling() {
			succeedingParagraph.AppendChild(succeedingParagraph, c)
		}
		for i := 0; i < paragraph.Lines().Len(); i++ {
			seg := paragraph.Lines().At(i)
			if seg.Start <= inline.Segment.Start {
				// We haven't passed the inline passthrough
				continue
			}
			if seg.Start >= inline.Segment.Stop {
				// We have completely passed the inline passthrough
				precedingParagraph.Lines().Append(seg)
				continue
			}
			newSeg := seg.WithStart(inline.Segment.Stop)
			if newSeg.Len() > 0 {
				precedingParagraph.Lines().Append(newSeg)
			}
		}
		if succeedingParagraph.ChildCount() > 0 || succeedingParagraph.Lines().Len() > 0 {
			parent.InsertAfter(parent, insertionPoint, succeedingParagraph)
		}

		parent.RemoveChild(parent, paragraph)
		return ast.WalkSkipChildren, nil
	})
}

func NewPassthroughInlineTransformer(ds []Delimiters) parser.ASTTransformer {
	return &passthroughInlineTransformer{
		BlockDelimiters: ds,
	}
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
	InlineDelimiters []Delimiters
	BlockDelimiters  []Delimiters
}

func NewPassthroughWithDelimiters(
	InlineDelimiters []Delimiters,
	BlockDelimiters []Delimiters) goldmark.Extender {
	// The parser executes in two phases:
	//
	// Phase 1: parse the input with all delimiters treated as inline, and block delimiters
	// taking precedence over inline delimiters.
	//
	// Phase 2: transform the parsed AST to split paragraphs at the point of
	// inline passthroughs with matching block delimiters.
	combinedDelimiters := make([]Delimiters, len(InlineDelimiters)+len(BlockDelimiters))
	copy(combinedDelimiters, BlockDelimiters)
	copy(combinedDelimiters[len(BlockDelimiters):], InlineDelimiters)
	return &passthrough{
		InlineDelimiters: combinedDelimiters,
		BlockDelimiters:  BlockDelimiters,
	}
}

func (e *passthrough) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithInlineParsers(
			util.Prioritized(NewInlinePassthroughParser(e.InlineDelimiters), 201),
		),
		parser.WithASTTransformers(
			util.Prioritized(NewPassthroughInlineTransformer(e.BlockDelimiters), 0),
		),
	)

	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewPassthroughInlineRenderer(), 101),
		util.Prioritized(NewPassthroughBlockRenderer(), 99),
	))
}
