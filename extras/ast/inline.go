package ast

import (
	gast "github.com/yuin/goldmark/ast"
)

type TagType int

const (
	Superscript TagType = iota + 1
	Subscript
	Insert
	Mark
)

type Tag struct {
	TagType           TagType
	Char              byte
	Number            int
	Html              string
	WhitespaceAllowed bool
	ParsePriority     int
	RenderPriority    int
}

var SuperscriptTag = Tag{
	TagType:           Superscript,
	Char:              '^',
	Number:            1,
	Html:              "sup",
	WhitespaceAllowed: false,
	ParsePriority:     600,
	RenderPriority:    600,
}

var SubscriptTag = Tag{
	TagType:           Subscript,
	Char:              '~',
	Number:            1,
	Html:              "sub",
	WhitespaceAllowed: false,
	ParsePriority:     602,
	RenderPriority:    602,
}

var InsertTag = Tag{
	TagType:           Insert,
	Char:              '+',
	Number:            2,
	Html:              "ins",
	WhitespaceAllowed: true,
	ParsePriority:     501,
	RenderPriority:    501,
}

var MarkTag = Tag{
	TagType:           Mark,
	Char:              '=',
	Number:            2,
	Html:              "mark",
	WhitespaceAllowed: true,
	ParsePriority:     550,
	RenderPriority:    550,
}

type InlineTag struct {
	gast.BaseInline

	Tag
}

func NewInlineTag(tag Tag) *InlineTag {
	return &InlineTag{
		BaseInline: gast.BaseInline{},

		Tag: tag,
	}
}

var KindSuperscript = gast.NewNodeKind("Superscript")
var KindSubscript = gast.NewNodeKind("Subscript")
var KindInsert = gast.NewNodeKind("Insert")
var KindMark = gast.NewNodeKind("Mark")

func NewInlineTagKind(t TagType) gast.NodeKind {
	var kind gast.NodeKind
	switch t {
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

func (n *InlineTag) Kind() gast.NodeKind {
	return NewInlineTagKind(n.TagType)
}

func (n *InlineTag) Dump(source []byte, level int) {
	gast.DumpHelper(n, source, level, nil, nil)
}
