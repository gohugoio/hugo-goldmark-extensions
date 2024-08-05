package extras

import (
	"github.com/yuin/goldmark/ast"
)

type inlineTag struct {
	TagKind        ast.NodeKind
	Char           byte
	Number         int
	Html           string
	ParsePriority  int
	RenderPriority int
}

var superscriptTag = inlineTag{
	TagKind:        kindSuperscript,
	Char:           '^',
	Number:         1,
	Html:           "sup",
	ParsePriority:  600,
	RenderPriority: 600,
}

var subscriptTag = inlineTag{
	TagKind:        kindSubscript,
	Char:           '~',
	Number:         1,
	Html:           "sub",
	ParsePriority:  602,
	RenderPriority: 602,
}

var insertTag = inlineTag{
	TagKind:        kindInsert,
	Char:           '+',
	Number:         2,
	Html:           "ins",
	ParsePriority:  501,
	RenderPriority: 501,
}

var markTag = inlineTag{
	TagKind:        kindMark,
	Char:           '=',
	Number:         2,
	Html:           "mark",
	ParsePriority:  550,
	RenderPriority: 550,
}

var deleteTag = inlineTag{
	TagKind:        kindDelete,
	Char:           '~',
	Number:         2,
	Html:           "del",
	ParsePriority:  400,
	RenderPriority: 400,
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
	kindDelete      = ast.NewNodeKind("Delete")
)

func (n *inlineTagNode) Kind() ast.NodeKind {
	return n.TagKind
}

func (n *inlineTagNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}
