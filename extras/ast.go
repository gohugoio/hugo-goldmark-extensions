package extras

import (
	gast "github.com/yuin/goldmark/ast"
)

type inlineTagType int

const (
	superscript inlineTagType = iota + 1
	subscript
	insert
	mark
)

type inlineTag struct {
	TagType           inlineTagType
	Char              byte
	Number            int
	Html              string
	WhitespaceAllowed bool
	ParsePriority     int
	RenderPriority    int
}

var superscriptTag = inlineTag{
	TagType:           superscript,
	Char:              '^',
	Number:            1,
	Html:              "sup",
	WhitespaceAllowed: false,
	ParsePriority:     600,
	RenderPriority:    600,
}

var subscriptTag = inlineTag{
	TagType:           subscript,
	Char:              '~',
	Number:            1,
	Html:              "sub",
	WhitespaceAllowed: false,
	ParsePriority:     602,
	RenderPriority:    602,
}

var insertTag = inlineTag{
	TagType:           insert,
	Char:              '+',
	Number:            2,
	Html:              "ins",
	WhitespaceAllowed: true,
	ParsePriority:     501,
	RenderPriority:    501,
}

var markTag = inlineTag{
	TagType:           mark,
	Char:              '=',
	Number:            2,
	Html:              "mark",
	WhitespaceAllowed: true,
	ParsePriority:     550,
	RenderPriority:    550,
}

type inlineTagNode struct {
	gast.BaseInline

	inlineTag
}

func newInlineTag(tag inlineTag) *inlineTagNode {
	return &inlineTagNode{
		BaseInline: gast.BaseInline{},

		inlineTag: tag,
	}
}

var (
	kindSuperscript = gast.NewNodeKind("Superscript")
	kindSubscript   = gast.NewNodeKind("Subscript")
	kindInsert      = gast.NewNodeKind("Insert")
	kindMark        = gast.NewNodeKind("Mark")
)

func newInlineTagNodeKind(tag inlineTagType) gast.NodeKind {
	var kind gast.NodeKind
	switch tag {
	case superscript:
		kind = kindSuperscript
	case subscript:
		kind = kindSubscript
	case insert:
		kind = kindInsert
	case mark:
		kind = kindMark
	}
	return kind
}

func (n *inlineTagNode) Kind() gast.NodeKind {
	return newInlineTagNodeKind(n.TagType)
}

func (n *inlineTagNode) Dump(source []byte, level int) {
	gast.DumpHelper(n, source, level, nil, nil)
}
