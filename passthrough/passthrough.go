package passthrough

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
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

// PassthroughInline is a node representing a inline passthrough.
type PassthroughInline struct {
	ast.BaseInline

	// The segment of text that this inline passthrough represents.
	Segment text.Segment

	// The matched delimiters
	Delimiters *Delimiters
}

func newPassthroughInline(segment text.Segment, delimiters *Delimiters) *PassthroughInline {
	return &PassthroughInline{
		Segment:    segment,
		Delimiters: delimiters,
	}
}

// Text implements Node.Text.
// Deprecated: Goldmark v1.7.8 deprecates Node.Text
func (n *PassthroughInline) Text(source []byte) []byte {
	return n.Segment.Value(source)
}

// Dump implements Node.Dump.
func (n *PassthroughInline) Dump(source []byte, level int) {
	indent := strings.Repeat("    ", level)
	fmt.Printf("%sPassthroughInline {\n", indent)
	indent2 := strings.Repeat("    ", level+1)
	fmt.Printf("%sSegment: \"%s\"\n", indent2, n.Segment.Value(source))
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

func newInlinePassthroughParser(ds []Delimiters) parser.InlineParser {
	return &inlinePassthroughParser{
		PassthroughDelimiters: ds,
	}
}

// Determine if the input slice starts with a full valid opening delimiter.
// If so, returns the delimiter struct, otherwise returns nil.
func getFullOpeningDelimiter(delims []Delimiters, line []byte) *Delimiters {
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
func openersFirstByte(delims []Delimiters) []byte {
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
func containsDelimiters(delims []Delimiters, toFind *Delimiters) bool {
	for _, d := range delims {
		if d.Open == toFind.Open && d.Close == toFind.Close {
			return true
		}
	}

	return false
}

func (s *inlinePassthroughParser) Trigger() []byte {
	return openersFirstByte(s.PassthroughDelimiters)
}

func (s *inlinePassthroughParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	// In order to prevent other parser extensions from operating on the text
	// between passthrough delimiters, we must process the entire inline
	// passthrough in one execution of Parse. This means we can't use the style
	// of multiple triggers with parser.Context state saved between calls.
	line, startSegment := block.PeekLine()

	fencePair := getFullOpeningDelimiter(s.PassthroughDelimiters, line)
	// fencePair == nil can happen if only the first byte of an opening delimiter
	// matches, but it is not the complete opening delimiter. The trigger causes
	// this Parse function to execute, but the trigger interface is limited to
	// matching single bytes.
	// It can also be because the opening delimiter is escaped with a
	// double-backslash. In this case, we advance and return nil.
	if fencePair == nil {
		if len(line) > 2 && line[0] == '\\' && line[1] == '\\' {
			fencePair = getFullOpeningDelimiter(s.PassthroughDelimiters, line[2:])
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
		return newPassthroughInline(seg, fencePair)
	}
}

type passthroughInlineRenderer struct{}

func (r *passthroughInlineRenderer) renderRawInline(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		n, ok := n.(*PassthroughInline)
		if !ok {
			return ast.WalkContinue, nil
		}
		w.WriteString(string(n.Segment.Value(source)))
	}
	return ast.WalkContinue, nil
}

// A PassthroughBlock struct represents a fenced block of raw text to pass
// through unchanged. This is not parsed directly, but emitted by an
// ASTTransformer that splits a paragraph at the point of an inline passthrough
// with the matching block delimiters.
type PassthroughBlock struct {
	ast.BaseBlock
	// The matched delimiters
	Delimiters *Delimiters
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

// newPassthroughBlock return a new PassthroughBlock node.
func newPassthroughBlock(delimiters *Delimiters) *PassthroughBlock {
	return &PassthroughBlock{
		Delimiters: delimiters,
		BaseBlock:  ast.BaseBlock{},
	}
}

type passthroughBlockRenderer struct{}

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

const (
	passthroughMarkedForDeletion = "passthrough_marked_for_deletion"
	passthroughProcessed         = "passthrough_processed"
)

// Note, this transformer destroys the RawText attributes of the paragraph
// nodes that it transforms. However, this does not seem to have an impact on
// rendering.
func (p *passthroughInlineTransformer) Transform(
	doc *ast.Document, reader text.Reader, pc parser.Context,
) {
	// Goldmark's walking algorithm is simplistic, and doesn't handle the
	// possibility of replacing the current node being walked with a new node. So
	// as a workaround, we split the walk in two. The first walk inserts new
	// nodes, and marks the original nodes for deletion. The second walk deletes
	// the marked nodes. To avoid an infinite loop, we also need to mark the
	// newly inserted nodes as "processed" so that they are not re-processed as
	// the walk continues.
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		// Anchor on paragraphs
		if n.Kind() != ast.KindParagraph || !entering {
			return ast.WalkContinue, nil
		}

		val, found := n.AttributeString(passthroughProcessed)
		if found && val == "true" {
			return ast.WalkContinue, nil
		}

		// If no direct children are passthroughs, skip it.
		foundInlinePassthrough := false
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			if c.Kind() == KindPassthroughInline {
				foundInlinePassthrough = true
				break
			}
		}
		if !foundInlinePassthrough {
			return ast.WalkContinue, nil
		}

		parent := n.Parent()
		currentParagraph := ast.NewParagraph()
		// AppendChild breaks the link between the node and its siblings, so we
		// need to manually track the current and next node.
		currentNode := n.FirstChild()
		insertionPoint := n

		for currentNode != nil {
			nextNode := currentNode.NextSibling()
			if currentNode.Kind() != KindPassthroughInline {
				currentParagraph.AppendChild(currentParagraph, currentNode)
				currentNode = nextNode
			} else if currentNode.Kind() == KindPassthroughInline {
				inline := currentNode.(*PassthroughInline)

				// Only split into a new block if the delimiters are block delimiters
				if !containsDelimiters(p.BlockDelimiters, inline.Delimiters) {
					currentParagraph.AppendChild(currentParagraph, currentNode)
					currentNode = nextNode
					continue
				}

				newBlock := newPassthroughBlock(inline.Delimiters)
				newBlock.Lines().Append(inline.Segment)
				if currentParagraph.ChildCount() > 0 {
					parent.InsertAfter(parent, insertionPoint, currentParagraph)
					// Since we're not removing the original paragraph, we need to ensure
					// that this paragraph is not re-processed as the walk continues
					currentParagraph.SetAttributeString(passthroughProcessed, "true")
					insertionPoint = currentParagraph
				}
				parent.InsertAfter(parent, insertionPoint, newBlock)
				insertionPoint = newBlock
				currentParagraph = ast.NewParagraph()
				currentNode = nextNode
			}
		}

		if currentParagraph.ChildCount() > 0 {
			parent.InsertAfter(parent, insertionPoint, currentParagraph)
			// Since we're not removing the original paragraph, we need to ensure
			// that this paragraph is not re-processed as the walk continues
			currentParagraph.SetAttributeString(passthroughProcessed, "true")
		}

		// At this point, we don't remove the original paragraph, but mark it
		// for removal in the second walk.
		n.SetAttributeString(passthroughMarkedForDeletion, "true")
		return ast.WalkContinue, nil
	})

	// Now delete any marked nodes
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		for c := n.FirstChild(); c != nil; {
			// Have to eagerly fetch this because `c` may be removed from the tree,
			// destroying its link to the next sibling.
			next := c.NextSibling()
			if c.Kind() == ast.KindParagraph {
				val, found := c.AttributeString(passthroughMarkedForDeletion)
				if found && val == "true" {
					n.RemoveChild(n, c)
				}
			}
			c = next
		}

		return ast.WalkContinue, nil
	})
}

func newPassthroughInlineTransformer(ds []Delimiters) parser.ASTTransformer {
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

// ---- Extension and config ----

type passthrough struct {
	InlineDelimiters []Delimiters
	BlockDelimiters  []Delimiters
}

// Config configures this extension.
type Config struct {
	InlineDelimiters []Delimiters
	BlockDelimiters  []Delimiters
}

func New(c Config) goldmark.Extender {
	// The parser executes in two phases:
	//
	// Phase 1: parse the input with all delimiters treated as inline, and block delimiters
	// taking precedence over inline delimiters.
	//
	// Phase 2: transform the parsed AST to split paragraphs at the point of
	// inline passthroughs with matching block delimiters.
	combinedDelimiters := make([]Delimiters, len(c.InlineDelimiters)+len(c.BlockDelimiters))
	copy(combinedDelimiters, c.BlockDelimiters)
	copy(combinedDelimiters[len(c.BlockDelimiters):], c.InlineDelimiters)
	return &passthrough{
		InlineDelimiters: combinedDelimiters,
		BlockDelimiters:  c.BlockDelimiters,
	}
}

func (e *passthrough) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithInlineParsers(
			util.Prioritized(newInlinePassthroughParser(e.InlineDelimiters), 201),
		),
		parser.WithASTTransformers(
			util.Prioritized(newPassthroughInlineTransformer(e.BlockDelimiters), 0),
		),
	)

	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&passthroughInlineRenderer{}, 101),
		util.Prioritized(&passthroughBlockRenderer{}, 99),
	))
}
