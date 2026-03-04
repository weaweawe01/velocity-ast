package dump

import (
	"strings"

	"github.com/weaweawe01/velocity-ast/internal/ast"
)

const (
	BRANCH = "├── "
	LAST   = "└── "
	PIPE   = "│   "
	SPACE  = "    "
)

func Render(root *ast.Node, tokens []ast.Token) string {
	var b strings.Builder
	writeNode(&b, root, nil, tokens, "", true, true)
	return b.String()
}

func writeNode(b *strings.Builder, node *ast.Node, prevSibling *ast.Node, tokens []ast.Token, prefix string, tail bool, root bool) {
	linePrefix := ""
	if !root {
		if tail {
			linePrefix = prefix + LAST
		} else {
			linePrefix = prefix + BRANCH
		}
	}

	b.WriteString(linePrefix)
	b.WriteString(node.VerboseString(tokens))
	if first := node.FirstToken(tokens); first != nil {
		b.WriteString(" -> ")
		b.WriteString(inferSpecialPrefix(node, prevSibling, tokens, first.Image))
		b.WriteString(first.Image)
	}
	b.WriteString("\n")

	childPrefix := prefix
	if !root {
		if tail {
			childPrefix += SPACE
		} else {
			childPrefix += PIPE
		}
	}

	for i, child := range node.Children {
		var prev *ast.Node
		if i > 0 {
			prev = node.Children[i-1]
		}
		writeNode(b, child, prev, tokens, childPrefix, i == len(node.Children)-1, false)
	}
}

func inferSpecialPrefix(node *ast.Node, prevSibling *ast.Node, tokens []ast.Token, image string) string {
	if node == nil || prevSibling == nil || node.Name != "ASTText" || prevSibling.Name != "ASTReference" {
		return ""
	}
	if image == "" {
		return ""
	}
	if isFormalReference(prevSibling, tokens) {
		return ""
	}
	switch image[0] {
	case ' ', '\t', '\n', '\r', '.', '"', '\'', ')', '\\', '!', ',', ';', '<', '>', '-':
		return image[:1]
	default:
		return ""
	}
}

func isFormalReference(node *ast.Node, tokens []ast.Token) bool {
	if node == nil {
		return false
	}
	first := node.FirstToken(tokens)
	if first == nil {
		return false
	}
	return first.Image == "${" || first.Image == "$!{"
}
