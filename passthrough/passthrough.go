package passthrough

import (
	"bytes"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type delimiters struct {
	Open  string
	Close string
}

type inlinePassthroughParser struct {
	PassthroughDelimiters []delimiters
}

// NewInlinePassthroughParser return a new InlineParser that parses inline pass through.
func NewInlinePassthroughParser(ds []delimiters) parser.InlineParser {
	return &inlinePassthroughParser{
		PassthroughDelimiters: ds,
	}
}

// `Parse` will be executed once for each character that is in this list of
// allowed trigger characters. Our parse function needs to do some additional
// checks because Trigger only works for single-byte delimiters.
func (s *inlinePassthroughParser) Trigger() []byte {
	var firstBytes []byte

	for _, d := range s.PassthroughDelimiters {
		firstBytes = append(firstBytes, d.Open[0])
		if d.Open[0] != d.Close[0] {
			firstBytes = append(firstBytes, d.Close[0])
		}
	}

	return firstBytes
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
func (s *inlinePassthroughParser) GetFullOpeningDelimiter(line []byte) *delimiters {
	for _, d := range s.PassthroughDelimiters {
		if startsWith(line, d.Open) {
			return &d
		}
	}

	return nil
}

func (s *inlinePassthroughParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	// In order to prevent other parser extensions from operating on the text
	// between passthrough delimiters, we must process the entire inline
	// passthrough in one execution of Parse. This means we can't use the style
	// of multiple triggers with parser.Context state saved between calls.
	line, startSegment := block.PeekLine()

	fencePair := s.GetFullOpeningDelimiter(line)
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

		return ast.NewRawTextSegment(seg)
	}
}

type passthrough struct {
	InlineDelimiters []delimiters
	BlockDelimiters  []delimiters
}

func NewPassthroughWithDelimiters(Inlines []delimiters, Blocks []delimiters) goldmark.Extender {
	return &passthrough{
		InlineDelimiters: Inlines,
		BlockDelimiters:  Blocks,
	}
}

func (e *passthrough) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(NewInlinePassthroughParser(e.InlineDelimiters), 501),
	))
}
