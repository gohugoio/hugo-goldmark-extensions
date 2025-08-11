package extras

import (
	"github.com/yuin/goldmark/ast"
)

type InlineTag struct {
	TagKind        ast.NodeKind
	Char           byte
	Number         int
	Html           string
	ParsePriority  int
	RenderPriority int
}

var SuperscriptTag = InlineTag{
	TagKind:        KindSuperscript,
	Char:           '^',
	Number:         1,
	Html:           "sup",
	ParsePriority:  600,
	RenderPriority: 600,
}

var SubscriptTag = InlineTag{
	TagKind:        KindSubscript,
	Char:           '~',
	Number:         1,
	Html:           "sub",
	ParsePriority:  602,
	RenderPriority: 602,
}

var InsertTag = InlineTag{
	TagKind:        KindInsert,
	Char:           '+',
	Number:         2,
	Html:           "ins",
	ParsePriority:  501,
	RenderPriority: 501,
}

var MarkTag = InlineTag{
	TagKind:        KindMark,
	Char:           '=',
	Number:         2,
	Html:           "mark",
	ParsePriority:  550,
	RenderPriority: 550,
}

var DeleteTag = InlineTag{
	TagKind:        KindDelete,
	Char:           '~',
	Number:         2,
	Html:           "del",
	ParsePriority:  400,
	RenderPriority: 400,
}

type inlineTagNode struct {
	ast.BaseInline

	InlineTag
}

func newInlineTag(tag InlineTag) *inlineTagNode {
	return &inlineTagNode{
		BaseInline: ast.BaseInline{},

		InlineTag: tag,
	}
}

var (
	KindSuperscript = ast.NewNodeKind("Superscript")
	KindSubscript   = ast.NewNodeKind("Subscript")
	KindInsert      = ast.NewNodeKind("Insert")
	KindMark        = ast.NewNodeKind("Mark")
	KindDelete      = ast.NewNodeKind("Delete")
)

func (n *inlineTagNode) Kind() ast.NodeKind {
	return n.TagKind
}

func (n *inlineTagNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}
