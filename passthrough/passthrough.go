package passthrough

import (
	"bytes"
	"io"

	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/extension"
	extast "github.com/yuin/goldmark/v2/extension/ast"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer"
	"github.com/yuin/goldmark/v2/renderer/html"
	"github.com/yuin/goldmark/v2/text"
	"github.com/yuin/goldmark/v2/util"
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
	n := &PassthroughInline{
		Segment:    segment,
		Delimiters: delimiters,
	}
	n.Init(n)
	return n
}

// Dump implements Node.Dump.
func (n *PassthroughInline) Dump(source []byte) *ast.NodeDump {
	return ast.NewNodeDump(n, map[string]any{"Segment": string(n.Segment.Bytes(source))})
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

// isInlineContainerNode returns true if n is a block-level node that contains
// inline content.
func isInlineContainerNode(n ast.Node) bool {
	if n == nil {
		return false
	}

	return n.Kind() == ast.KindParagraph
}

// newInlineContainer creates a fresh inline container.
func newInlineContainer() ast.Node {
	return ast.NewParagraph()
}

// isTightParagraph reports whether the paragraph n is rendered without <p> tags
// because it lives in a tight block. This mirrors the HTML renderer's tight
// detection: a paragraph directly inside a tight definition description, or a
// paragraph in a list item whose list is tight.
func isTightParagraph(n ast.Node) bool {
	parent := n.Parent()
	if parent == nil {
		return false
	}
	if desc, ok := parent.(*extast.DefinitionDescription); ok {
		return desc.IsTight
	}
	return html.IsInTightBlock(n)
}

// trimContainerSpace removes leading or trailing whitespace from text nodes
// in a container, modifying the text segment boundaries in place.
func trimContainerSpace(container ast.Node, source []byte, leading bool) {
	if container.ChildCount() == 0 {
		return
	}

	var textNode *ast.Text
	if leading {
		child := container.FirstChild()
		if child.Kind() != ast.KindText {
			return
		}
		textNode = child.(*ast.Text)
	} else {
		child := container.LastChild()
		if child.Kind() != ast.KindText {
			return
		}
		textNode = child.(*ast.Text)
	}

	// Text nodes produced by the block parser are backed by a source position;
	// owned (literal string) values can't be trimmed by adjusting an index.
	if textNode.Value.IsOwned() {
		return
	}
	idx := textNode.Value.Index()
	value := textNode.Value.Bytes(source)
	origLen := len(value)

	if leading {
		value = bytes.TrimLeft(value, " \t")
		newLen := len(value)
		if newLen < origLen {
			textNode.Value = text.NewIndexValue(text.NewIndex(idx.Start+(origLen-newLen), idx.Stop))
		}
	} else {
		value = bytes.TrimRight(value, " \t")
		newLen := len(value)
		if newLen < origLen {
			textNode.Value = text.NewIndexValue(text.NewIndex(idx.Start, idx.Stop-(origLen-newLen)))
		}
	}
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
				return ast.NewSegmentText(startSegment.WithStop(startSegment.Start + len(fencePair.Open) + 2))
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
			return ast.NewSegmentText(startSegment.WithStop(startSegment.Start + openerSize))
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

// renderRawInline renders a PassthroughInline node, emitting its raw segment
// unchanged. It is wrapped with html.NodeRendererFunc when registered.
func renderRawInline(w io.Writer, source []byte, n ast.Node, entering bool, _ renderer.Context) (ast.WalkStatus, error) {
	bw := w.(util.BufWriter)
	if entering {
		n, ok := n.(*PassthroughInline)
		if !ok {
			return ast.WalkContinue, nil
		}
		bw.WriteString(string(n.Segment.Bytes(source)))
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
func (n *PassthroughBlock) Dump(source []byte) *ast.NodeDump {
	return ast.NewNodeDump(n, nil)
}

// KindPassthroughBlock is a NodeKind of the PassthroughBlock node.
var KindPassthroughBlock = ast.NewNodeKind("PassthroughBlock")

// Kind implements Node.Kind.
func (n *PassthroughBlock) Kind() ast.NodeKind {
	return KindPassthroughBlock
}

// newPassthroughBlock return a new PassthroughBlock node.
func newPassthroughBlock(delimiters *Delimiters) *PassthroughBlock {
	n := &PassthroughBlock{
		Delimiters: delimiters,
		BaseBlock:  ast.BaseBlock{},
	}
	n.Init(n)
	return n
}

// renderRawBlock renders a PassthroughBlock node, emitting its raw source lines
// unchanged. It is wrapped with html.NodeRendererFunc when registered.
func renderRawBlock(w io.Writer, source []byte, n ast.Node, entering bool, _ renderer.Context) (ast.WalkStatus, error) {
	bw := w.(util.BufWriter)
	if entering {
		block, ok := n.(ast.BlockNode)
		if !ok {
			return ast.WalkContinue, nil
		}
		lines := block.Source()
		for i := 0; i < len(lines); i++ {
			line := lines[i]
			bw.WriteString(string(line.Bytes(source)))
		}
		bw.WriteString("\n")
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
	doc *ast.Document, reader text.Reader, pc parser.Context,
) {
	source := reader.Source()
	processed := map[ast.Node]bool{}
	markedForDeletion := map[ast.Node]bool{}

	// Goldmark's walking algorithm is simplistic, and doesn't handle the
	// possibility of replacing the current node being walked with a new node. So
	// as a workaround, we split the walk in two. The first walk inserts new
	// nodes, and marks the original nodes for deletion. The second walk deletes
	// the marked nodes. To avoid an infinite loop, we also need to mark the
	// newly inserted nodes as "processed" so that they are not re-processed as
	// the walk continues.
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering || !isInlineContainerNode(n) {
			return ast.WalkContinue, nil
		}

		if processed[n] {
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
		// isTight reproduces the v1 KindTextBlock special-casing: in tight
		// contexts the surrounding text is trimmed around the split-out block.
		isTight := isTightParagraph(n)
		// In Goldmark v2 the task-list checkbox is emitted by the paragraph
		// renderer for the first child paragraph of a task list item (in v1 it was
		// an inline node). If such a paragraph begins directly with a block
		// passthrough, splitting it out would leave the list item with no leading
		// paragraph and the checkbox would be lost; detect that case so we can
		// keep an (empty) leading paragraph to carry the checkbox.
		_, isTask := extension.TaskStatusOf(parent)
		isTaskFirstChildParagraph := isTask && parent.FirstChild() == n
		currentContainer := newInlineContainer()
		// AppendChild breaks the link between the node and its siblings, so we
		// need to manually track the current and next node.
		currentNode := n.FirstChild()
		insertionPoint := n

		for currentNode != nil {
			nextNode := currentNode.NextSibling()
			if currentNode.Kind() != KindPassthroughInline {
				currentContainer.AppendChild(currentNode)
				currentNode = nextNode
			} else if currentNode.Kind() == KindPassthroughInline {
				inline := currentNode.(*PassthroughInline)

				// Only split into a new block if the delimiters are block delimiters
				if !containsDelimiters(p.BlockDelimiters, inline.Delimiters) {
					currentContainer.AppendChild(currentNode)
					currentNode = nextNode
					continue
				}

				newBlock := newPassthroughBlock(inline.Delimiters)
				newBlock.SetPos(inline.Pos())
				newBlock.AppendSource(inline.Segment)
				if currentContainer.ChildCount() > 0 {
					// Trim trailing whitespace from text preceding the block in tight lists
					if isTight {
						trimContainerSpace(currentContainer, source, false)
					}
					// Trim leading whitespace from text following a previous block in tight lists
					if isTight && insertionPoint.Kind() == KindPassthroughBlock {
						trimContainerSpace(currentContainer, source, true)
					}
					parent.InsertAfter(insertionPoint, currentContainer)
					// Since we're not removing the original paragraph, we need to ensure
					// that this paragraph is not re-processed as the walk continues
					processed[currentContainer] = true
					insertionPoint = currentContainer
				} else if isTaskFirstChildParagraph && insertionPoint == n {
					// The list item's content begins with a block passthrough. Keep an
					// empty leading paragraph (with an empty text child so the renderer
					// emits the trailing newline) so the task-list checkbox still renders.
					placeholder := newInlineContainer()
					placeholder.AppendChild(ast.NewText())
					parent.InsertAfter(insertionPoint, placeholder)
					processed[placeholder] = true
					insertionPoint = placeholder
				}
				parent.InsertAfter(insertionPoint, newBlock)
				insertionPoint = newBlock
				currentContainer = newInlineContainer()
				currentNode = nextNode
			}
		}

		if currentContainer.ChildCount() > 0 {
			// Trim leading whitespace from text following a block in tight lists
			if isTight && insertionPoint.Kind() == KindPassthroughBlock {
				trimContainerSpace(currentContainer, source, true)
			}
			parent.InsertAfter(insertionPoint, currentContainer)
			// Since we're not removing the original paragraph, we need to ensure
			// that this paragraph is not re-processed as the walk continues
			processed[currentContainer] = true
		}

		// At this point, we don't remove the original paragraph, but mark it
		// for removal in the second walk.
		markedForDeletion[n] = true
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
			if isInlineContainerNode(c) && markedForDeletion[c] {
				n.RemoveChild(c)
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

// ---- Extension and config ----

// Config configures this extension.
type Config struct {
	InlineDelimiters []Delimiters
	BlockDelimiters  []Delimiters
}

// passthroughParser is the parser.Extension for passthrough delimiters.
type passthroughParser struct {
	InlineDelimiters []Delimiters
	BlockDelimiters  []Delimiters
}

// NewParser returns a parser.Extension that parses passthrough delimiters. Add
// it with parser.WithExtensions.
//
// The parser executes in two phases:
//
// Phase 1: parse the input with all delimiters treated as inline, and block
// delimiters taking precedence over inline delimiters.
//
// Phase 2: transform the parsed AST to split paragraphs at the point of inline
// passthroughs with matching block delimiters.
func NewParser(c Config) parser.Extension {
	combinedDelimiters := make([]Delimiters, len(c.InlineDelimiters)+len(c.BlockDelimiters))
	copy(combinedDelimiters, c.BlockDelimiters)
	copy(combinedDelimiters[len(c.BlockDelimiters):], c.InlineDelimiters)
	return &passthroughParser{
		InlineDelimiters: combinedDelimiters,
		BlockDelimiters:  c.BlockDelimiters,
	}
}

// ParserOptions implements parser.Extension.
func (e *passthroughParser) ParserOptions(_ *parser.Config) []parser.Option {
	return []parser.Option{
		parser.WithInlineParsers(
			util.Prioritized(newInlinePassthroughParser(e.InlineDelimiters), 201),
		),
		parser.WithASTTransformers(
			util.Prioritized(newPassthroughInlineTransformer(e.BlockDelimiters), 0),
		),
	}
}

// passthroughHTMLRenderer is the html.Extension for passthrough nodes.
type passthroughHTMLRenderer struct{}

// NewHTMLRenderer returns an html.Extension that renders passthrough nodes. Add
// it with html.WithExtensions.
func NewHTMLRenderer(_ Config) html.Extension {
	return &passthroughHTMLRenderer{}
}

// RendererOptions implements html.Extension.
func (r *passthroughHTMLRenderer) RendererOptions(_ *html.Config) []html.Option {
	return []html.Option{
		html.WithNodeRenderers(map[ast.NodeKind]html.NodeRenderer{
			KindPassthroughInline: html.NodeRendererFunc(renderRawInline),
			KindPassthroughBlock:  html.NodeRendererFunc(renderRawBlock),
		}),
	}
}
