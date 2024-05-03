package ast

import (
	gast "github.com/yuin/goldmark/ast"
)

type InlineTagType int

const (
	Superscript InlineTagType = iota + 1
	Subscript
	Insert
	Mark
)

type InlineTag struct {
	TagType           InlineTagType
	Char              byte
	Number            int
	Html              string
	WhitespaceAllowed bool
	ParsePriority     int
	RenderPriority    int
}

var SuperscriptTag = InlineTag{
	TagType:           Superscript,
	Char:              '^',
	Number:            1,
	Html:              "sup",
	WhitespaceAllowed: false,
	ParsePriority:     600,
	RenderPriority:    600,
}

var SubscriptTag = InlineTag{
	TagType:           Subscript,
	Char:              '~',
	Number:            1,
	Html:              "sub",
	WhitespaceAllowed: false,
	ParsePriority:     602,
	RenderPriority:    602,
}

var InsertTag = InlineTag{
	TagType:           Insert,
	Char:              '+',
	Number:            2,
	Html:              "ins",
	WhitespaceAllowed: true,
	ParsePriority:     501,
	RenderPriority:    501,
}

var MarkTag = InlineTag{
	TagType:           Mark,
	Char:              '=',
	Number:            2,
	Html:              "mark",
	WhitespaceAllowed: true,
	ParsePriority:     550,
	RenderPriority:    550,
}

type InlineTagNode struct {
	gast.BaseInline

	InlineTag
}

func NewInlineTag(tag InlineTag) *InlineTagNode {
	return &InlineTagNode{
		BaseInline: gast.BaseInline{},

		InlineTag: tag,
	}
}

var KindSuperscript = gast.NewNodeKind("Superscript")
var KindSubscript = gast.NewNodeKind("Subscript")
var KindInsert = gast.NewNodeKind("Insert")
var KindMark = gast.NewNodeKind("Mark")

func NewInlineTagNodeKind(tag InlineTagType) gast.NodeKind {
	var kind gast.NodeKind
	switch tag {
	case Superscript:
		kind = KindSuperscript
	case Subscript:
		kind = KindSubscript
	case Insert:
		kind = KindInsert
	case Mark:
		kind = KindMark
	}
	return kind
}

func (n *InlineTagNode) Kind() gast.NodeKind {
	return NewInlineTagNodeKind(n.TagType)
}

func (n *InlineTagNode) Dump(source []byte, level int) {
	gast.DumpHelper(n, source, level, nil, nil)
}
