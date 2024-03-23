package ast

import (
	gast "github.com/yuin/goldmark/ast"
)

// A Insert struct represents a insert text segment.
type Insert struct {
	gast.BaseInline
}

// Dump implements Node.Dump.
func (n *Insert) Dump(source []byte, level int) {
	gast.DumpHelper(n, source, level, nil, nil)
}

// KindInsert is a NodeKind of the Insert node.
var KindInsert = gast.NewNodeKind("Insert")

// Kind implements Node.Kind.
func (n *Insert) Kind() gast.NodeKind {
	return KindInsert
}

// NewInsert returns a new Insert node.
func NewInsert() *Insert {
	return &Insert{}
}
