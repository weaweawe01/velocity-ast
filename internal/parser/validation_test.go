package parser

import (
	"strings"
	"testing"
)

func TestDirectiveValidationParse(t *testing.T) {
	_, _, err := Parse(`#parse "a.vm"`)
	if err == nil || !strings.Contains(err.Error(), "#parse directive requires one argument") {
		t.Fatalf("expected #parse validation error, got: %v", err)
	}
}

func TestDirectiveValidationStopBreak(t *testing.T) {
	_, _, err := Parse(`#stop($a,$b)`)
	if err == nil || !strings.Contains(err.Error(), "#stop directive only accepts") {
		t.Fatalf("expected #stop validation error, got: %v", err)
	}

	_, _, err = Parse(`#break($a,$b)`)
	if err == nil || !strings.Contains(err.Error(), "#break directive takes only") {
		t.Fatalf("expected #break validation error, got: %v", err)
	}
}

func TestDirectiveValidationDefine(t *testing.T) {
	_, _, err := Parse(`#define(foo)bar#end`)
	if err == nil || !strings.Contains(err.Error(), "argument to #define is of the wrong type") {
		t.Fatalf("expected #define validation error, got: %v", err)
	}
}

func TestHyphenReferenceSplitsToText(t *testing.T) {
	root, tokens, err := Parse(`$foo-bar`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(root.Children) != 2 {
		t.Fatalf("expected 2 root children, got %d", len(root.Children))
	}
	if root.Children[0].Name != "ASTReference" {
		t.Fatalf("expected first child ASTReference, got %s", root.Children[0].Name)
	}
	if root.Children[1].Name != "ASTText" {
		t.Fatalf("expected second child ASTText, got %s", root.Children[1].Name)
	}
	if got := tokens[root.Children[1].FirstIdx].Image; got != "-bar" {
		t.Fatalf("expected trailing text token '-bar', got %q", got)
	}
}

func TestLineDirectivePostfixNewline(t *testing.T) {
	cases := []struct {
		name     string
		template string
	}{
		{name: "parse", template: `#parse("a.vm")
X`},
		{name: "include", template: `#include("a.vm")
X`},
		{name: "stop", template: `#stop
X`},
		{name: "break", template: `#break
X`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root, tokens, err := Parse(tc.template)
			if err != nil {
				t.Fatalf("parse failed: %v", err)
			}
			if len(root.Children) != 2 {
				t.Fatalf("expected directive + text children, got %d", len(root.Children))
			}
			dir := root.Children[0]
			txt := root.Children[1]
			if dir.Name != "ASTDirective" {
				t.Fatalf("expected first child ASTDirective, got %s", dir.Name)
			}
			if txt.Name != "ASTText" {
				t.Fatalf("expected second child ASTText, got %s", txt.Name)
			}
			if got := tokens[dir.LastIdx].Image; got != "\n" {
				t.Fatalf("expected directive last token to be newline, got %q", got)
			}
			if got := tokens[txt.FirstIdx].Image; got != "X" {
				t.Fatalf("expected trailing text token 'X', got %q", got)
			}
		})
	}
}

func TestIfBlockLeadingNewlineNotTextChild(t *testing.T) {
	root, tokens, err := Parse(`#if($x)
X
#end`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(root.Children) != 1 {
		t.Fatalf("expected one root child, got %d", len(root.Children))
	}
	ifNode := root.Children[0]
	if ifNode.Name != "ASTIfStatement" {
		t.Fatalf("expected ASTIfStatement, got %s", ifNode.Name)
	}
	if len(ifNode.Children) != 2 {
		t.Fatalf("expected condition + block children, got %d", len(ifNode.Children))
	}
	block := ifNode.Children[1]
	if block.Name != "ASTBlock" {
		t.Fatalf("expected ASTBlock, got %s", block.Name)
	}
	if len(block.Children) != 1 {
		t.Fatalf("expected one block child, got %d", len(block.Children))
	}
	text := block.Children[0]
	if text.Name != "ASTText" {
		t.Fatalf("expected block child ASTText, got %s", text.Name)
	}
	if got := tokens[block.FirstIdx].Image; got != "\n" {
		t.Fatalf("expected block leading token to be newline, got %q", got)
	}
	if got := tokens[text.FirstIdx].Image; got != "X\n" {
		t.Fatalf("expected block text token to be 'X\\n', got %q", got)
	}
	if !(block.FirstIdx < text.FirstIdx) {
		t.Fatalf("expected block prefix token before text token, got block.FirstIdx=%d text.FirstIdx=%d", block.FirstIdx, text.FirstIdx)
	}
}

func TestEscapedDirectiveAndReferenceNodes(t *testing.T) {
	root, tokens, err := Parse(`\#if \$a`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(root.Children) != 3 {
		t.Fatalf("expected 3 root children, got %d", len(root.Children))
	}
	if root.Children[0].Name != "ASTEscapedDirective" {
		t.Fatalf("expected first child ASTEscapedDirective, got %s", root.Children[0].Name)
	}
	if root.Children[0].ID != 3 {
		t.Fatalf("expected escaped directive node id 3, got %d", root.Children[0].ID)
	}
	if got := tokens[root.Children[0].FirstIdx].Image; got != "#if" {
		t.Fatalf("expected escaped directive token '#if', got %q", got)
	}
	if root.Children[1].Name != "ASTText" {
		t.Fatalf("expected second child ASTText, got %s", root.Children[1].Name)
	}
	if got := tokens[root.Children[1].FirstIdx].Image; got != " " {
		t.Fatalf("expected middle text token ' ', got %q", got)
	}
	if root.Children[2].Name != "ASTReference" {
		t.Fatalf("expected third child ASTReference, got %s", root.Children[2].Name)
	}
	if got := tokens[root.Children[2].FirstIdx].Image; got != `\$a` {
		t.Fatalf("expected escaped reference token '\\\\$a', got %q", got)
	}
}

func TestParseTrimsTrailingEOFNewline(t *testing.T) {
	root, tokens, err := Parse("a\n")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(root.Children) != 1 || root.Children[0].Name != "ASTText" {
		t.Fatalf("expected one ASTText child, got %d nodes", len(root.Children))
	}
	if got := tokens[root.Children[0].FirstIdx].Image; got != "a" {
		t.Fatalf("expected trailing newline trimmed from text token, got %q", got)
	}

	root, _, err = Parse("\n")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(root.Children) != 0 {
		t.Fatalf("expected empty process for newline-only template, got %d children", len(root.Children))
	}

	root, tokens, err = Parse("\n\n")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(root.Children) != 1 || root.Children[0].Name != "ASTText" {
		t.Fatalf("expected one ASTText child for double newline input, got %d nodes", len(root.Children))
	}
	if got := tokens[root.Children[0].FirstIdx].Image; got != "\n" {
		t.Fatalf("expected single trailing newline preserved for double newline input, got %q", got)
	}

	root, tokens, err = Parse("$foo\n")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(root.Children) != 1 || root.Children[0].Name != "ASTReference" {
		t.Fatalf("expected one ASTReference child, got %d nodes", len(root.Children))
	}
	if got := tokens[root.Children[0].FirstIdx].Image; got != "$foo" {
		t.Fatalf("expected trailing newline trimmed after reference, got %q", got)
	}
}
