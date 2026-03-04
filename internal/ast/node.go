package ast

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// Token stores parser token image in source order.
type Token struct {
	Image string
}

// Node mirrors the subset of Java SimpleNode fields used by AST dump.
type Node struct {
	Name     string
	ID       int
	Info     int
	Invalid  bool
	FirstIdx int
	LastIdx  int
	Children []*Node
	// DirectiveName is used by ASTDirective formatting.
	DirectiveName string
}

func NewNode(name string, id int, firstIdx, lastIdx int) *Node {
	return &Node{
		Name:     name,
		ID:       id,
		Info:     0,
		Invalid:  false,
		FirstIdx: firstIdx,
		LastIdx:  lastIdx,
	}
}

func (n *Node) AddChild(child *Node) {
	n.Children = append(n.Children, child)
}

func (n *Node) FirstToken(tokens []Token) *Token {
	if n.FirstIdx < 0 || n.FirstIdx >= len(tokens) {
		return nil
	}
	return &tokens[n.FirstIdx]
}

func (n *Node) TokensString(tokens []Token) string {
	if n.FirstIdx < 0 || n.LastIdx < n.FirstIdx || n.FirstIdx >= len(tokens) {
		return ""
	}
	last := n.LastIdx
	if last >= len(tokens) {
		last = len(tokens) - 1
	}

	parts := make([]string, 0, last-n.FirstIdx+1)
	for i := n.FirstIdx; i <= last; i++ {
		img := strings.ReplaceAll(tokens[i].Image, "\n", "\\n")
		parts = append(parts, "["+img+"]")
	}

	tok := strings.Join(parts, ", ")
	if utf8.RuneCountInString(tok) > 50 {
		runes := []rune(tok)
		tok = string(runes[:50]) + "..."
	}
	return tok
}

func (n *Node) VerboseString(tokens []Token) string {
	base := fmt.Sprintf("%s [id=%d, info=%d, invalid=%t, tokens=%s]", n.Name, n.ID, n.Info, n.Invalid, n.TokensString(tokens))
	if n.Name == "ASTDirective" {
		return fmt.Sprintf("ASTDirective [%s, directiveName=%s]", base, n.DirectiveName)
	}
	return base
}

var NodeIDs = map[string]int{
	"ASTprocess":              0,
	"ASTText":                 2,
	"ASTEscapedDirective":     3,
	"ASTEscape":               4,
	"ASTComment":              5,
	"ASTTextblock":            6,
	"ASTFloatingPointLiteral": 7,
	"ASTIntegerLiteral":       8,
	"ASTStringLiteral":        9,
	"ASTIdentifier":           10,
	"ASTWord":                 11,
	"ASTDirectiveAssign":      12,
	"ASTDirective":            13,
	"ASTBlock":                14,
	"ASTMap":                  15,
	"ASTObjectArray":          16,
	"ASTIntegerRange":         17,
	"ASTMethod":               18,
	"ASTIndex":                19,
	"ASTReference":            20,
	"ASTTrue":                 21,
	"ASTFalse":                22,
	"ASTIfStatement":          23,
	"ASTElseStatement":        24,
	"ASTElseIfStatement":      25,
	"ASTSetDirective":         26,
	"ASTExpression":           27,
	"ASTOrNode":               29,
	"ASTAndNode":              30,
	"ASTEQNode":               31,
	"ASTNENode":               32,
	"ASTLTNode":               33,
	"ASTGTNode":               34,
	"ASTLENode":               35,
	"ASTGENode":               36,
	"ASTAddNode":              37,
	"ASTSubtractNode":         38,
	"ASTMulNode":              39,
	"ASTDivNode":              40,
	"ASTModNode":              41,
	"ASTNotNode":              42,
	"ASTNegateNode":           43,
}
