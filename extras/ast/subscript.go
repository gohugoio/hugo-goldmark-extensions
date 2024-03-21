// Package ast defines AST nodes that represents extension's elements
package ast

import (
	gast "github.com/yuin/goldmark/ast"
)

// A Subscript struct represents subscript text.
type Subscript struct {
	gast.BaseInline
}

// Dump implements Node.Dump.
func (n *Subscript) Dump(source []byte, level int) {
	gast.DumpHelper(n, source, level, nil, nil)
}

// KindSubscript is a NodeKind of the Subscript node.
var KindSubscript = gast.NewNodeKind("Subscript")

// Kind implements Node.Kind.
func (n *Subscript) Kind() gast.NodeKind {
	return KindSubscript
}

// NewSubscript returns a new Subscript node.
func NewSubscript() *Subscript {
	return &Subscript{}
}
