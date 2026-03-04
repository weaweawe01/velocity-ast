// Package velocityast provides a public API for parsing Apache Velocity
// templates into an AST and rendering the tree as a human-readable dump.
package velocityast

import (
	"github.com/weaweawe01/velocity-ast/internal/ast"
	"github.com/weaweawe01/velocity-ast/internal/dump"
	"github.com/weaweawe01/velocity-ast/internal/parser"
)

// Node is the AST node type returned by Parse.
type Node = ast.Node

// Token is the token type returned by Parse.
type Token = ast.Token

// Parse parses a Velocity template string and returns the root AST node,
// the flat token slice, and any parse error.
func Parse(template string) (*Node, []Token, error) {
	return parser.Parse(template)
}

// Render walks the AST rooted at root and returns a Java-compatible
// tree-dump string. tokens must be the slice returned by Parse.
func Render(root *Node, tokens []Token) string {
	return dump.Render(root, tokens)
}
