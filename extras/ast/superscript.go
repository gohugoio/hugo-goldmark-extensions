// Package ast defines AST nodes that represents extension's elements
package ast

import (
	gast "github.com/yuin/goldmark/ast"
)

// A Superscript struct represents a superscript text.
type Superscript struct {
	gast.BaseInline
}

// Dump implements Node.Dump.
func (n *Superscript) Dump(source []byte, level int) {
	gast.DumpHelper(n, source, level, nil, nil)
}

// KindSuperscript is a NodeKind of the Superscript node.
var KindSuperscript = gast.NewNodeKind("Superscript")

// Kind implements Node.Kind.
func (n *Superscript) Kind() gast.NodeKind {
	return KindSuperscript
}

// NewSuperscript returns a new Superscript node.
func NewSuperscript() *Superscript {
	return &Superscript{}
}
