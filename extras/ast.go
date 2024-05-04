package extras

import (
	"github.com/yuin/goldmark/ast"
)

type inlineTag struct {
	TagKind           ast.NodeKind
	Char              byte
	Number            int
	Html              string
	WhitespaceAllowed bool
	ParsePriority     int
	RenderPriority    int
}

var superscriptTag = inlineTag{
	TagKind:           kindSuperscript,
	Char:              '^',
	Number:            1,
	Html:              "sup",
	WhitespaceAllowed: false,
	ParsePriority:     600,
	RenderPriority:    600,
}

var subscriptTag = inlineTag{
	TagKind:           kindSubscript,
	Char:              '~',
	Number:            1,
	Html:              "sub",
	WhitespaceAllowed: false,
	ParsePriority:     602,
	RenderPriority:    602,
}

var insertTag = inlineTag{
	TagKind:           kindInsert,
	Char:              '+',
	Number:            2,
	Html:              "ins",
	WhitespaceAllowed: true,
	ParsePriority:     501,
	RenderPriority:    501,
}

var markTag = inlineTag{
	TagKind:           kindMark,
	Char:              '=',
	Number:            2,
	Html:              "mark",
	WhitespaceAllowed: true,
	ParsePriority:     550,
	RenderPriority:    550,
}

type inlineTagNode struct {
	ast.BaseInline

	inlineTag
}

func newInlineTag(tag inlineTag) *inlineTagNode {
	return &inlineTagNode{
		BaseInline: ast.BaseInline{},

		inlineTag: tag,
	}
}

var (
	kindSuperscript = ast.NewNodeKind("Superscript")
	kindSubscript   = ast.NewNodeKind("Subscript")
	kindInsert      = ast.NewNodeKind("Insert")
	kindMark        = ast.NewNodeKind("Mark")
)

func (n *inlineTagNode) Kind() ast.NodeKind {
	return n.TagKind
}

func (n *inlineTagNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}
