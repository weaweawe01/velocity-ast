package parser

import (
	"fmt"
	"strings"

	"github.com/weaweawe01/velocity-ast/internal/ast"
	"github.com/weaweawe01/velocity-ast/internal/lexer"
)

type Parser struct {
	input     string
	tokens    []lexer.Token
	pos       int
	astTokens []ast.Token
	macroDefs map[string]bool
}

const deferredLastIdx = -2

func Parse(template string) (*ast.Node, []ast.Token, error) {
	template = trimEOFNewlines(template)
	lexed, err := lexer.Lex(template)
	if err != nil {
		return nil, nil, err
	}

	p := &Parser{
		input:     template,
		tokens:    lexed,
		pos:       0,
		macroDefs: make(map[string]bool),
	}
	root, err := p.parseProcess()
	if err != nil {
		return nil, nil, err
	}
	return root, p.astTokens, nil
}

func trimEOFNewlines(s string) string {
	// Java demo file loader (BufferedReader.readLine + join with '\n')
	// effectively drops only the final line separator sequence.
	if strings.HasSuffix(s, "\r\n") {
		return s[:len(s)-2]
	}
	if strings.HasSuffix(s, "\n") || strings.HasSuffix(s, "\r") {
		return s[:len(s)-1]
	}
	return s
}

func (p *Parser) parseProcess() (*ast.Node, error) {
	children := make([]*ast.Node, 0)
	for p.peek().Kind != lexer.TokenEOF {
		child, err := p.parseNode()
		if err != nil {
			return nil, err
		}
		children = append(children, child)
	}

	// Java parser emits three control tokens at process tail.
	p.astTokens = append(p.astTokens,
		ast.Token{Image: "\x1c"},
		ast.Token{Image: "\x1c"},
		ast.Token{Image: "\x1c"},
	)

	first, last := rangeFromChildren(children)
	if len(p.astTokens) > 0 {
		last = len(p.astTokens) - 1
	}
	root := ast.NewNode("ASTprocess", ast.NodeIDs["ASTprocess"], first, last)
	root.Children = children
	resolveDeferredLastIdx(root, last)
	return root, nil
}

func (p *Parser) parseNode() (*ast.Node, error) {
	if indentFirst, ok := p.consumeLeadingDirectiveIndent(); ok {
		node, err := p.parseNodeCore()
		if err != nil {
			return nil, err
		}
		if node != nil && indentFirst >= 0 && (node.FirstIdx < 0 || indentFirst < node.FirstIdx) {
			node.FirstIdx = indentFirst
		}
		return node, nil
	}
	return p.parseNodeCore()
}

func (p *Parser) parseNodeCore() (*ast.Node, error) {
	switch p.peek().Kind {
	case lexer.TokenDirectiveSet:
		return p.parseSetDirective()
	case lexer.TokenDirective:
		return p.parseHashDirective()
	case lexer.TokenEscapedDirective:
		return p.parseEscapedDirective()
	case lexer.TokenReference:
		return p.parseReference()
	case lexer.TokenEscape:
		return p.parseEscape()
	case lexer.TokenCommentLine:
		return p.parseLineComment()
	case lexer.TokenComment:
		return p.parseBlockComment()
	case lexer.TokenTextblock:
		return p.parseTextblock()
	default:
		return p.parseText()
	}
}

func (p *Parser) consumeLeadingDirectiveIndent() (int, bool) {
	if !p.canAttachLeadingDirectiveIndent() {
		return -1, false
	}
	first := -1
	for {
		tok := p.peek()
		if tok.Kind != lexer.TokenText || !isWhitespaceOnly(tok.Image) || hasLineBreak(tok.Image) {
			break
		}
		idx, _ := p.consumeAny()
		if first < 0 {
			first = idx
		}
	}
	return first, first >= 0
}

func (p *Parser) canAttachLeadingDirectiveIndent() bool {
	tok := p.peek()
	if tok.Kind != lexer.TokenText || !isWhitespaceOnly(tok.Image) || hasLineBreak(tok.Image) {
		return false
	}
	if tok.Pos > 0 {
		prev := p.input[tok.Pos-1]
		if prev != '\n' && prev != '\r' {
			return false
		}
	}

	offset := 0
	for {
		wsTok := p.peekN(offset)
		if wsTok.Kind == lexer.TokenText && isWhitespaceOnly(wsTok.Image) && !hasLineBreak(wsTok.Image) {
			offset++
			continue
		}
		break
	}
	next := p.peekN(offset)
	if next.Kind == lexer.TokenDirectiveSet {
		return true
	}
	if next.Kind != lexer.TokenDirective {
		return false
	}
	switch canonicalDirective(next.Image) {
	case "#elseif", "#else", "#end":
		return false
	default:
		return true
	}
}

func (p *Parser) parseSetDirective() (*ast.Node, error) {
	first, err := p.expect(lexer.TokenDirectiveSet)
	if err != nil {
		return nil, err
	}
	p.consumeSpaces()

	lhs, err := p.parseReference()
	if err != nil {
		return nil, err
	}
	p.consumeSpaces()
	if _, err = p.expect(lexer.TokenEquals); err != nil {
		return nil, err
	}

	exprStart := len(p.astTokens)

	rhs, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if exprStart >= 0 && exprStart < rhs.FirstIdx {
		rhs.FirstIdx = exprStart
	}
	if p.consumeSpaces() > 0 {
		rhs.LastIdx = len(p.astTokens) - 1
	}

	last, err := p.expect(lexer.TokenRParen)
	if err != nil {
		return nil, err
	}

	node := ast.NewNode("ASTSetDirective", ast.NodeIDs["ASTSetDirective"], first, last)
	node.AddChild(lhs)
	node.AddChild(rhs)
	before := len(p.astTokens)
	if p.consumeDirectivePostfixNewline(true) && len(p.astTokens) > before {
		node.LastIdx = len(p.astTokens) - 1
	}
	return node, nil
}

func (p *Parser) parseText() (*ast.Node, error) {
	first, err := p.consumeTextRun()
	if err != nil {
		return nil, err
	}
	return ast.NewNode("ASTText", ast.NodeIDs["ASTText"], first, first), nil
}

func (p *Parser) parseEscape() (*ast.Node, error) {
	idx, err := p.expect(lexer.TokenEscape)
	if err != nil {
		return nil, err
	}
	return ast.NewNode("ASTEscape", ast.NodeIDs["ASTEscape"], idx, idx), nil
}

func (p *Parser) parseEscapedDirective() (*ast.Node, error) {
	idx, err := p.expect(lexer.TokenEscapedDirective)
	if err != nil {
		return nil, err
	}
	if idx >= 0 && idx < len(p.astTokens) {
		img := p.astTokens[idx].Image
		if strings.HasPrefix(img, "\\#") {
			name := directiveName(img[1:])
			if p.macroDefs[name] {
				p.astTokens[idx].Image = img[1:]
			}
		}
	}
	return ast.NewNode("ASTEscapedDirective", ast.NodeIDs["ASTEscapedDirective"], idx, idx), nil
}

func (p *Parser) parseLineComment() (*ast.Node, error) {
	first, err := p.expect(lexer.TokenCommentLine)
	if err != nil {
		return nil, err
	}
	last := first
	if p.peek().Kind == lexer.TokenText {
		img := p.peek().Image
		if strings.Contains(img, "\n") || strings.Contains(img, "\r") {
			last, _ = p.consumeAny()
		}
	}
	return ast.NewNode("ASTComment", ast.NodeIDs["ASTComment"], first, last), nil
}

func (p *Parser) parseBlockComment() (*ast.Node, error) {
	idx, err := p.expect(lexer.TokenComment)
	if err != nil {
		return nil, err
	}
	return ast.NewNode("ASTComment", ast.NodeIDs["ASTComment"], idx, idx), nil
}

func (p *Parser) parseTextblock() (*ast.Node, error) {
	idx, err := p.expect(lexer.TokenTextblock)
	if err != nil {
		return nil, err
	}
	return ast.NewNode("ASTTextblock", ast.NodeIDs["ASTTextblock"], idx, idx), nil
}

func (p *Parser) parseHashDirective() (*ast.Node, error) {
	dirTok := p.peek()
	if dirTok.Kind != lexer.TokenDirective {
		return nil, fmt.Errorf("expected hash directive at pos=%d", dirTok.Pos)
	}
	switch canonicalDirective(dirTok.Image) {
	case "#if":
		return p.parseIfDirective()
	case "#foreach":
		return p.parseForeachDirective()
	case "#macro":
		return p.parseMacroDirective()
	case "#define":
		return p.parseDefineDirective()
	case "#elseif", "#else", "#end":
		return nil, fmt.Errorf("unexpected %s at pos=%d", canonicalDirective(dirTok.Image), dirTok.Pos)
	default:
		return p.parseDirectiveCall()
	}
}

func (p *Parser) parseIfDirective() (*ast.Node, error) {
	first, err := p.expectDirectiveCanonical("#if")
	if err != nil {
		return nil, err
	}
	p.consumeSpaces()
	if _, err = p.expect(lexer.TokenLParen); err != nil {
		return nil, err
	}
	condStart := len(p.astTokens)
	cond, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if condStart >= 0 && condStart < cond.FirstIdx {
		cond.FirstIdx = condStart
	}
	if p.consumeSpaces() > 0 {
		cond.LastIdx = len(p.astTokens) - 1
		p.extendTopUnaryExprLast(cond, cond.LastIdx)
	}
	if _, err = p.expect(lexer.TokenRParen); err != nil {
		return nil, err
	}

	block, err := p.parseBlockUntilDirectives(map[string]bool{
		"#elseif": true,
		"#else":   true,
		"#end":    true,
	}, true)
	if err != nil {
		return nil, err
	}

	children := []*ast.Node{cond, block}

	for p.isDirectiveToken("#elseif") {
		elseIfNode, e := p.parseElseIfNode()
		if e != nil {
			return nil, e
		}
		children = append(children, elseIfNode)
	}

	if p.isDirectiveToken("#else") {
		elseNode, e := p.parseElseNode()
		if e != nil {
			return nil, e
		}
		children = append(children, elseNode)
	}

	last, err := p.expectDirectiveCanonical("#end")
	if err != nil {
		return nil, err
	}
	before := len(p.astTokens)
	if p.consumeDirectivePostfixNewline(p.shouldKeepPostfixNewlineForNestedStop()) && len(p.astTokens) > before {
		last = len(p.astTokens) - 1
	}

	node := ast.NewNode("ASTIfStatement", ast.NodeIDs["ASTIfStatement"], first, last)
	for _, child := range children {
		node.AddChild(child)
	}
	return node, nil
}

func (p *Parser) parseForeachDirective() (*ast.Node, error) {
	first, err := p.expectDirectiveCanonical("#foreach")
	if err != nil {
		return nil, err
	}
	p.consumeSpaces()
	if _, err = p.expect(lexer.TokenLParen); err != nil {
		return nil, err
	}
	p.consumeSpaces()
	ref1, err := p.parseReference()
	if err != nil {
		return nil, err
	}
	p.consumeSpaces()
	word, err := p.parseWord()
	if err != nil {
		return nil, err
	}
	p.consumeSpaces()
	iter, err := p.parseDirectiveArg()
	if err != nil {
		iter, err = p.parseExpression()
		if err != nil {
			return nil, err
		}
	}
	p.consumeSpaces()
	if _, err = p.expect(lexer.TokenRParen); err != nil {
		return nil, err
	}
	block, err := p.parseBlockUntilDirectives(map[string]bool{
		"#else": true,
		"#end":  true,
	}, false)
	if err != nil {
		return nil, err
	}
	children := []*ast.Node{ref1, word, iter, block}
	if p.isDirectiveToken("#else") {
		if _, e := p.expectDirectiveCanonical("#else"); e != nil {
			return nil, e
		}
		elseBlock, e := p.parseBlockUntilDirectives(map[string]bool{"#end": true}, true)
		if e != nil {
			return nil, e
		}
		children = append(children, elseBlock)
	}

	last, err := p.expectDirectiveCanonical("#end")
	if err != nil {
		return nil, err
	}
	before := len(p.astTokens)
	if p.consumeDirectivePostfixNewline(p.shouldKeepPostfixNewlineForNestedStop()) && len(p.astTokens) > before {
		last = len(p.astTokens) - 1
	}

	node := ast.NewNode("ASTDirective", ast.NodeIDs["ASTDirective"], first, last)
	node.DirectiveName = "foreach"
	for _, child := range children {
		node.AddChild(child)
	}
	return node, nil
}

func (p *Parser) parseMacroDirective() (*ast.Node, error) {
	first, err := p.expectDirectiveCanonical("#macro")
	if err != nil {
		return nil, err
	}
	p.consumeSpaces()
	if _, err = p.expect(lexer.TokenLParen); err != nil {
		return nil, err
	}
	p.consumeSpaces()
	nameWord, err := p.parseWord()
	if err != nil {
		return nil, err
	}
	if nameWord.FirstIdx >= 0 && nameWord.FirstIdx < len(p.astTokens) {
		p.macroDefs[p.astTokens[nameWord.FirstIdx].Image] = true
	}
	children := []*ast.Node{nameWord}
	for {
		p.consumeSpaces()
		if p.peek().Kind == lexer.TokenRParen {
			break
		}
		if p.peek().Kind == lexer.TokenComma {
			if _, err = p.expect(lexer.TokenComma); err != nil {
				return nil, err
			}
			continue
		}
		if p.peek().Kind == lexer.TokenReference {
			arg, e := p.parseReference()
			if e != nil {
				return nil, e
			}
			p.consumeSpaces()
			if p.peek().Kind == lexer.TokenEquals {
				assign := ast.NewNode("ASTDirectiveAssign", ast.NodeIDs["ASTDirectiveAssign"], arg.FirstIdx, arg.LastIdx)
				assign.AddChild(arg)
				children = append(children, assign)
				if _, e = p.expect(lexer.TokenEquals); e != nil {
					return nil, e
				}
				p.consumeSpaces()
				val, e := p.parseDirectiveArg()
				if e != nil {
					return nil, e
				}
				children = append(children, val)
			} else {
				children = append(children, arg)
			}
			continue
		}
		if p.peek().Kind == lexer.TokenIdentifier {
			argWord, e := p.parseWord()
			if e != nil {
				return nil, e
			}
			children = append(children, argWord)
			continue
		}
		return nil, fmt.Errorf("unexpected token in #macro args at pos=%d: %q", p.peek().Pos, p.peek().Image)
	}
	p.consumeSpaces()
	if _, err = p.expect(lexer.TokenRParen); err != nil {
		return nil, err
	}
	p.consumeDirectivePostfixNewline(true)
	block, err := p.parseBlockUntilDirectives(map[string]bool{"#end": true}, false)
	if err != nil {
		return nil, err
	}
	last, err := p.expectDirectiveCanonical("#end")
	if err != nil {
		return nil, err
	}
	p.consumeDirectivePostfixNewline(false)

	node := ast.NewNode("ASTDirective", ast.NodeIDs["ASTDirective"], first, last)
	node.DirectiveName = "macro"
	for _, child := range children {
		node.AddChild(child)
	}
	node.AddChild(block)
	return node, nil
}

func (p *Parser) parseDefineDirective() (*ast.Node, error) {
	first, err := p.expectDirectiveCanonical("#define")
	if err != nil {
		return nil, err
	}
	p.consumeSpaces()
	if _, err = p.expect(lexer.TokenLParen); err != nil {
		return nil, err
	}
	p.consumeSpaces()
	arg, err := p.parseDirectiveArg()
	if err != nil {
		return nil, err
	}
	if arg.Name == "ASTWord" {
		return nil, fmt.Errorf("the argument to #define is of the wrong type")
	}
	p.consumeSpaces()
	if _, err = p.expect(lexer.TokenRParen); err != nil {
		return nil, err
	}

	block, err := p.parseBlockUntilDirectives(map[string]bool{"#end": true}, false)
	if err != nil {
		return nil, err
	}
	last, err := p.expectDirectiveCanonical("#end")
	if err != nil {
		return nil, err
	}
	p.consumeDirectivePostfixNewline(false)

	node := ast.NewNode("ASTDirective", ast.NodeIDs["ASTDirective"], first, last)
	node.DirectiveName = "define"
	node.AddChild(arg)
	node.AddChild(block)
	return node, nil
}

func (p *Parser) parseDirectiveCall() (*ast.Node, error) {
	tok := p.peek()
	first, err := p.expect(lexer.TokenDirective)
	if err != nil {
		return nil, err
	}
	name := directiveName(tok.Image)

	node := ast.NewNode("ASTDirective", ast.NodeIDs["ASTDirective"], first, first)
	node.DirectiveName = name

	next, ws := p.peekAfterWhitespace()
	if next.Kind == lexer.TokenLParen {
		p.consumeWhitespaceTokens(ws)
		if _, err = p.expect(lexer.TokenLParen); err != nil {
			return nil, err
		}
		p.consumeSpaces()
		for p.peek().Kind != lexer.TokenRParen {
			if p.peek().Kind == lexer.TokenComma {
				if _, err = p.expect(lexer.TokenComma); err != nil {
					return nil, err
				}
				p.consumeSpaces()
				continue
			}

			arg, e := p.parseDirectiveArg()
			if e != nil {
				return nil, e
			}
			node.AddChild(arg)
			p.consumeSpaces()

			if p.peek().Kind == lexer.TokenComma {
				if _, err = p.expect(lexer.TokenComma); err != nil {
					return nil, err
				}
				p.consumeSpaces()
			}
		}
		p.consumeSpaces()
		last, e := p.expect(lexer.TokenRParen)
		if e != nil {
			return nil, e
		}
		node.LastIdx = last
	}
	if err = p.validateDirectiveArgs(name, node.Children); err != nil {
		return nil, err
	}
	if strings.HasPrefix(name, "@") {
		block, e := p.parseBlockUntilDirectives(map[string]bool{"#end": true}, false)
		if e != nil {
			return nil, e
		}
		node.AddChild(block)
		last, e := p.expectDirectiveCanonical("#end")
		if e != nil {
			return nil, e
		}
		node.LastIdx = last
		p.consumeDirectivePostfixNewline(false)
	} else {
		if lastIdx, _ := p.consumeLineDirectivePostfixNewline(); lastIdx >= 0 {
			node.LastIdx = lastIdx
		}
	}

	return node, nil
}

func (p *Parser) parseDirectiveArg() (*ast.Node, error) {
	p.consumeSpaces()
	switch p.peek().Kind {
	case lexer.TokenReference:
		return p.parseReference()
	case lexer.TokenNumber:
		return p.parseNumericLiteral()
	case lexer.TokenString:
		idx, err := p.expect(lexer.TokenString)
		if err != nil {
			return nil, err
		}
		return ast.NewNode("ASTStringLiteral", ast.NodeIDs["ASTStringLiteral"], idx, idx), nil
	case lexer.TokenLBracket:
		return p.parseArrayOrRange()
	case lexer.TokenLBrace:
		return p.parseMap()
	case lexer.TokenIdentifier:
		if p.peekKeyword("true") || p.peekKeyword("false") {
			return p.parseBooleanLiteral()
		}
		return p.parseWord()
	default:
		tok := p.peek()
		return nil, fmt.Errorf("unsupported directive arg at pos=%d image=%q", tok.Pos, tok.Image)
	}
}

func (p *Parser) parseWord() (*ast.Node, error) {
	idx, err := p.expect(lexer.TokenIdentifier)
	if err != nil {
		return nil, err
	}
	return ast.NewNode("ASTWord", ast.NodeIDs["ASTWord"], idx, idx), nil
}

func (p *Parser) parseElseIfNode() (*ast.Node, error) {
	first, err := p.expectDirectiveCanonical("#elseif")
	if err != nil {
		return nil, err
	}
	p.consumeSpaces()
	if _, err = p.expect(lexer.TokenLParen); err != nil {
		return nil, err
	}
	condStart := len(p.astTokens)
	cond, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if condStart >= 0 && condStart < cond.FirstIdx {
		cond.FirstIdx = condStart
	}
	if p.consumeSpaces() > 0 {
		cond.LastIdx = len(p.astTokens) - 1
		p.extendTopUnaryExprLast(cond, cond.LastIdx)
	}
	if _, err = p.expect(lexer.TokenRParen); err != nil {
		return nil, err
	}
	block, err := p.parseBlockUntilDirectives(map[string]bool{
		"#elseif": true,
		"#else":   true,
		"#end":    true,
	}, true)
	if err != nil {
		return nil, err
	}

	last := block.LastIdx
	if last < 0 {
		last = cond.LastIdx
	}
	node := ast.NewNode("ASTElseIfStatement", ast.NodeIDs["ASTElseIfStatement"], first, last)
	node.AddChild(cond)
	node.AddChild(block)
	return node, nil
}

func (p *Parser) parseElseNode() (*ast.Node, error) {
	first, err := p.expectDirectiveCanonical("#else")
	if err != nil {
		return nil, err
	}
	block, err := p.parseBlockUntilDirectives(map[string]bool{"#end": true}, true)
	if err != nil {
		return nil, err
	}

	last := block.LastIdx
	if last < 0 {
		last = first
	}
	node := ast.NewNode("ASTElseStatement", ast.NodeIDs["ASTElseStatement"], first, last)
	node.AddChild(block)
	return node, nil
}

func (p *Parser) parseBlockUntilDirectives(stop map[string]bool, keepPrefixNewline bool) (*ast.Node, error) {
	children := make([]*ast.Node, 0)
	prefixFirst := -1
	before := len(p.astTokens)
	p.consumeDirectivePostfixNewline(keepPrefixNewline)
	if keepPrefixNewline && len(p.astTokens) > before {
		prefixFirst = before
	}
	for {
		tok := p.peek()
		if tok.Kind == lexer.TokenEOF {
			return nil, fmt.Errorf("unexpected EOF: missing #end")
		}
		if scan, ok := p.scanHorizontalWhitespaceBeforeStop(stop); ok {
			if len(children) == 0 {
				for p.pos < scan {
					idx, _ := p.consumeAny()
					if prefixFirst < 0 {
						prefixFirst = idx
					}
				}
			} else {
				// Java parser drops trailing indentation-only text before stop directives.
				p.pos = scan
			}
			break
		}
		if tok.Kind == lexer.TokenDirective && stop[canonicalDirective(tok.Image)] {
			break
		}

		child, err := p.parseNode()
		if err != nil {
			return nil, err
		}
		children = append(children, child)
	}

	first, last := rangeFromChildren(children)
	if len(children) == 0 {
		if prefixFirst >= 0 {
			first = prefixFirst
		} else {
			// Keep empty block without body text anchored at the upcoming stop directive.
			first = len(p.astTokens)
		}
		last = deferredLastIdx
	} else if prefixFirst >= 0 {
		first = prefixFirst
	}
	block := ast.NewNode("ASTBlock", ast.NodeIDs["ASTBlock"], first, last)
	block.Children = children
	return block, nil
}

func (p *Parser) parseReference() (*ast.Node, error) {
	tok := p.peek()
	if tok.Kind != lexer.TokenReference {
		return nil, fmt.Errorf("expected reference at pos=%d, got %q", tok.Pos, tok.Image)
	}

	first, err := p.expect(lexer.TokenReference)
	if err != nil {
		return nil, err
	}

	node := ast.NewNode("ASTReference", ast.NodeIDs["ASTReference"], first, first)
	if tok.Image == "${" || tok.Image == "$!{" {
		return p.parseFormalReference(node, first)
	}

	last := first
	if err := p.parseReferenceIndexes(node, &last); err != nil {
		return nil, err
	}

	for p.peek().Kind == lexer.TokenDot && p.peekN(1).Kind == lexer.TokenIdentifier {
		if _, err := p.expect(lexer.TokenDot); err != nil {
			return nil, err
		}
		member, err := p.parseReferenceMember()
		if err != nil {
			return nil, err
		}
		node.AddChild(member)
		last = member.LastIdx

		if err := p.parseReferenceIndexes(node, &last); err != nil {
			return nil, err
		}
	}

	node.LastIdx = last
	return node, nil
}

func (p *Parser) parseFormalReference(node *ast.Node, first int) (*ast.Node, error) {
	base, err := p.expect(lexer.TokenIdentifier)
	if err != nil {
		return nil, err
	}
	last := base

	if err := p.parseReferenceIndexes(node, &last); err != nil {
		return nil, err
	}

	for p.peek().Kind == lexer.TokenDot && p.peekN(1).Kind == lexer.TokenIdentifier {
		if _, err = p.expect(lexer.TokenDot); err != nil {
			return nil, err
		}
		member, e := p.parseReferenceMember()
		if e != nil {
			return nil, e
		}
		node.AddChild(member)
		last = member.LastIdx

		if e := p.parseReferenceIndexes(node, &last); e != nil {
			return nil, e
		}
	}

	if p.peek().Kind == lexer.TokenPipe {
		if _, err = p.expect(lexer.TokenPipe); err != nil {
			return nil, err
		}
		expr, e := p.parseExpression()
		if e != nil {
			return nil, e
		}
		node.AddChild(expr)
		last = expr.LastIdx
	}

	closeIdx, err := p.expect(lexer.TokenRBrace)
	if err != nil {
		return nil, err
	}
	last = closeIdx

	node.FirstIdx = first
	node.LastIdx = last
	return node, nil
}

func (p *Parser) parseReferenceIndexes(node *ast.Node, last *int) error {
	for p.peek().Kind == lexer.TokenLBracket {
		idxNode, err := p.parseIndex()
		if err != nil {
			return err
		}
		node.AddChild(idxNode)
		*last = idxNode.LastIdx
	}
	return nil
}

func (p *Parser) parseReferenceMember() (*ast.Node, error) {
	if p.peek().Kind != lexer.TokenIdentifier {
		return nil, fmt.Errorf("expected identifier after dot at pos=%d", p.peek().Pos)
	}
	if p.peekN(1).Kind == lexer.TokenLParen {
		// Java parser is fairly lenient in odd templates: if method parsing
		// cannot complete (for example malformed args), treat this as plain
		// identifier member and let following tokens parse as text/directives.
		savedPos := p.pos
		savedLen := len(p.astTokens)
		method, err := p.parseMethod()
		if err == nil {
			return method, nil
		}
		p.pos = savedPos
		p.astTokens = p.astTokens[:savedLen]
	}
	return p.parseIdentifierNode()
}

func (p *Parser) parseIdentifierNode() (*ast.Node, error) {
	idx, err := p.expect(lexer.TokenIdentifier)
	if err != nil {
		return nil, err
	}
	return ast.NewNode("ASTIdentifier", ast.NodeIDs["ASTIdentifier"], idx, idx), nil
}

func (p *Parser) parseMethod() (*ast.Node, error) {
	first, err := p.expect(lexer.TokenIdentifier)
	if err != nil {
		return nil, err
	}
	ident := ast.NewNode("ASTIdentifier", ast.NodeIDs["ASTIdentifier"], first, first)

	if _, err = p.expect(lexer.TokenLParen); err != nil {
		return nil, err
	}

	args := make([]*ast.Node, 0)
	if p.peek().Kind != lexer.TokenRParen {
		arg, err := p.parseMethodArgExpression()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)

		for p.peek().Kind == lexer.TokenComma {
			if _, err = p.expect(lexer.TokenComma); err != nil {
				return nil, err
			}
			arg, err = p.parseMethodArgExpression()
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
		}
	}
	p.consumeSpaces()

	last, err := p.expect(lexer.TokenRParen)
	if err != nil {
		return nil, err
	}

	node := ast.NewNode("ASTMethod", ast.NodeIDs["ASTMethod"], first, last)
	node.AddChild(ident)
	for _, arg := range args {
		node.AddChild(arg)
	}
	return node, nil
}

func (p *Parser) parseMethodArgExpression() (*ast.Node, error) {
	argStart := len(p.astTokens)
	p.consumeSpaces()
	arg, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if argStart >= 0 && argStart < arg.FirstIdx {
		arg.FirstIdx = argStart
	}
	if p.consumeSpaces() > 0 {
		arg.LastIdx = len(p.astTokens) - 1
	}
	return arg, nil
}

func (p *Parser) parseIndex() (*ast.Node, error) {
	openIdx, err := p.expect(lexer.TokenLBracket)
	if err != nil {
		return nil, err
	}
	inner, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	p.consumeSpaces()
	closeIdx, err := p.expect(lexer.TokenRBracket)
	if err != nil {
		return nil, err
	}

	node := ast.NewNode("ASTIndex", ast.NodeIDs["ASTIndex"], openIdx, closeIdx)
	node.AddChild(inner)
	return node, nil
}

func (p *Parser) parseExpression() (*ast.Node, error) {
	p.consumeSpaces()
	exprFirst := len(p.astTokens)
	child, err := p.parseOrExpression()
	if err != nil {
		return nil, err
	}
	if exprFirst < 0 || exprFirst > child.LastIdx {
		exprFirst = child.FirstIdx
	}
	expr := ast.NewNode("ASTExpression", ast.NodeIDs["ASTExpression"], exprFirst, child.LastIdx)
	expr.AddChild(child)
	return expr, nil
}

func (p *Parser) parseOrExpression() (*ast.Node, error) {
	left, err := p.parseAndExpression()
	if err != nil {
		return nil, err
	}
	for {
		next, ws := p.peekAfterWhitespace()
		op := lexer.TokenEOF
		switch {
		case next.Kind == lexer.TokenOrOr:
			p.consumeWhitespaceTokens(ws)
			if _, err = p.expect(lexer.TokenOrOr); err != nil {
				return nil, err
			}
			op = lexer.TokenOrOr
		case next.Kind == lexer.TokenIdentifier && next.Image == "or":
			p.consumeWhitespaceTokens(ws)
			if _, err = p.expectKeyword("or"); err != nil {
				return nil, err
			}
			op = lexer.TokenOrOr
		default:
			return left, nil
		}
		rightStart := len(p.astTokens)
		p.consumeSpaces()
		right, e := p.parseAndExpression()
		if e != nil {
			return nil, e
		}
		rightLast := right.LastIdx
		if p.consumeSpaces() > 0 {
			rightLast = len(p.astTokens) - 1
		}
		left = makeBinaryNode(op, left, right, rightStart, rightLast)
	}
}

func (p *Parser) parseAndExpression() (*ast.Node, error) {
	left, err := p.parseComparisonExpression()
	if err != nil {
		return nil, err
	}
	for {
		next, ws := p.peekAfterWhitespace()
		op := lexer.TokenEOF
		switch {
		case next.Kind == lexer.TokenAndAnd:
			p.consumeWhitespaceTokens(ws)
			if _, err = p.expect(lexer.TokenAndAnd); err != nil {
				return nil, err
			}
			op = lexer.TokenAndAnd
		case next.Kind == lexer.TokenIdentifier && next.Image == "and":
			p.consumeWhitespaceTokens(ws)
			if _, err = p.expectKeyword("and"); err != nil {
				return nil, err
			}
			op = lexer.TokenAndAnd
		default:
			return left, nil
		}
		rightStart := len(p.astTokens)
		p.consumeSpaces()
		right, e := p.parseComparisonExpression()
		if e != nil {
			return nil, e
		}
		rightLast := right.LastIdx
		if p.consumeSpaces() > 0 {
			rightLast = len(p.astTokens) - 1
		}
		left = makeBinaryNode(op, left, right, rightStart, rightLast)
	}
}

func (p *Parser) parseComparisonExpression() (*ast.Node, error) {
	left, err := p.parseAdditiveExpression()
	if err != nil {
		return nil, err
	}
	for {
		next, ws := p.peekAfterWhitespace()
		if _, ok := comparisonKindForToken(next); !ok {
			return left, nil
		}
		p.consumeWhitespaceTokens(ws)
		op, consumed, e := p.consumeComparisonOperator()
		if e != nil {
			return nil, e
		}
		if !consumed {
			return left, nil
		}
		rightStart := len(p.astTokens)
		p.consumeSpaces()
		right, e := p.parseAdditiveExpression()
		if e != nil {
			return nil, e
		}
		rightLast := right.LastIdx
		if p.consumeSpaces() > 0 {
			rightLast = len(p.astTokens) - 1
		}
		left = makeBinaryNode(op, left, right, rightStart, rightLast)
	}
}

func (p *Parser) consumeComparisonOperator() (lexer.Kind, bool, error) {
	switch p.peek().Kind {
	case lexer.TokenEqEq:
		_, err := p.expect(lexer.TokenEqEq)
		return lexer.TokenEqEq, true, err
	case lexer.TokenNotEq:
		_, err := p.expect(lexer.TokenNotEq)
		return lexer.TokenNotEq, true, err
	case lexer.TokenLt:
		_, err := p.expect(lexer.TokenLt)
		return lexer.TokenLt, true, err
	case lexer.TokenGt:
		_, err := p.expect(lexer.TokenGt)
		return lexer.TokenGt, true, err
	case lexer.TokenLe:
		_, err := p.expect(lexer.TokenLe)
		return lexer.TokenLe, true, err
	case lexer.TokenGe:
		_, err := p.expect(lexer.TokenGe)
		return lexer.TokenGe, true, err
	}

	switch {
	case p.peekKeyword("eq"):
		_, err := p.expectKeyword("eq")
		return lexer.TokenEqEq, true, err
	case p.peekKeyword("ne"):
		_, err := p.expectKeyword("ne")
		return lexer.TokenNotEq, true, err
	case p.peekKeyword("lt"):
		_, err := p.expectKeyword("lt")
		return lexer.TokenLt, true, err
	case p.peekKeyword("gt"):
		_, err := p.expectKeyword("gt")
		return lexer.TokenGt, true, err
	case p.peekKeyword("le"):
		_, err := p.expectKeyword("le")
		return lexer.TokenLe, true, err
	case p.peekKeyword("ge"):
		_, err := p.expectKeyword("ge")
		return lexer.TokenGe, true, err
	default:
		return lexer.TokenEOF, false, nil
	}
}

func comparisonKindForToken(tok lexer.Token) (lexer.Kind, bool) {
	switch tok.Kind {
	case lexer.TokenEqEq, lexer.TokenNotEq, lexer.TokenLt, lexer.TokenGt, lexer.TokenLe, lexer.TokenGe:
		return tok.Kind, true
	case lexer.TokenIdentifier:
		switch tok.Image {
		case "eq":
			return lexer.TokenEqEq, true
		case "ne":
			return lexer.TokenNotEq, true
		case "lt":
			return lexer.TokenLt, true
		case "gt":
			return lexer.TokenGt, true
		case "le":
			return lexer.TokenLe, true
		case "ge":
			return lexer.TokenGe, true
		}
	}
	return lexer.TokenEOF, false
}

func (p *Parser) parseAdditiveExpression() (*ast.Node, error) {
	left, err := p.parseMultiplicativeExpression()
	if err != nil {
		return nil, err
	}
	for {
		next, ws := p.peekAfterWhitespace()
		op := next.Kind
		if op != lexer.TokenPlus && op != lexer.TokenMinus {
			return left, nil
		}
		p.consumeWhitespaceTokens(ws)
		if _, err = p.expect(op); err != nil {
			return nil, err
		}
		rightStart := len(p.astTokens)
		p.consumeSpaces()
		right, e := p.parseMultiplicativeExpression()
		if e != nil {
			return nil, e
		}
		rightLast := right.LastIdx
		if p.consumeSpaces() > 0 {
			rightLast = len(p.astTokens) - 1
		}
		left = makeBinaryNode(op, left, right, rightStart, rightLast)
	}
}

func (p *Parser) parseMultiplicativeExpression() (*ast.Node, error) {
	left, err := p.parseUnaryExpression()
	if err != nil {
		return nil, err
	}
	for {
		next, ws := p.peekAfterWhitespace()
		op := next.Kind
		if op != lexer.TokenMul && op != lexer.TokenDiv && op != lexer.TokenMod {
			return left, nil
		}
		p.consumeWhitespaceTokens(ws)
		if _, err = p.expect(op); err != nil {
			return nil, err
		}
		rightStart := len(p.astTokens)
		p.consumeSpaces()
		right, e := p.parseUnaryExpression()
		if e != nil {
			return nil, e
		}
		rightLast := right.LastIdx
		if p.consumeSpaces() > 0 {
			rightLast = len(p.astTokens) - 1
		}
		left = makeBinaryNode(op, left, right, rightStart, rightLast)
	}
}

func binaryNodeMeta(kind lexer.Kind) (string, int, error) {
	switch kind {
	case lexer.TokenOrOr:
		return "ASTOrNode", ast.NodeIDs["ASTOrNode"], nil
	case lexer.TokenAndAnd:
		return "ASTAndNode", ast.NodeIDs["ASTAndNode"], nil
	case lexer.TokenEqEq:
		return "ASTEQNode", ast.NodeIDs["ASTEQNode"], nil
	case lexer.TokenNotEq:
		return "ASTNENode", ast.NodeIDs["ASTNENode"], nil
	case lexer.TokenLt:
		return "ASTLTNode", ast.NodeIDs["ASTLTNode"], nil
	case lexer.TokenGt:
		return "ASTGTNode", ast.NodeIDs["ASTGTNode"], nil
	case lexer.TokenLe:
		return "ASTLENode", ast.NodeIDs["ASTLENode"], nil
	case lexer.TokenGe:
		return "ASTGENode", ast.NodeIDs["ASTGENode"], nil
	case lexer.TokenPlus:
		return "ASTAddNode", ast.NodeIDs["ASTAddNode"], nil
	case lexer.TokenMinus:
		return "ASTSubtractNode", ast.NodeIDs["ASTSubtractNode"], nil
	case lexer.TokenMul:
		return "ASTMulNode", ast.NodeIDs["ASTMulNode"], nil
	case lexer.TokenDiv:
		return "ASTDivNode", ast.NodeIDs["ASTDivNode"], nil
	case lexer.TokenMod:
		return "ASTModNode", ast.NodeIDs["ASTModNode"], nil
	default:
		return "", 0, fmt.Errorf("unsupported binary token kind: %v", kind)
	}
}

func makeBinaryNode(op lexer.Kind, left, right *ast.Node, rightStart, rightLast int) *ast.Node {
	name, id, err := binaryNodeMeta(op)
	if err != nil {
		// Keep parser behavior deterministic; panic here indicates programmer bug.
		panic(err)
	}

	first := rightStart
	if right == nil || first > right.LastIdx || first < 0 {
		first = right.FirstIdx
	}
	last := rightLast
	if right == nil || last < right.FirstIdx {
		last = right.LastIdx
	}
	node := ast.NewNode(name, id, first, last)
	node.AddChild(left)
	node.AddChild(right)
	return node
}

func (p *Parser) parseUnaryExpression() (*ast.Node, error) {
	p.consumeSpaces()
	if p.peek().Kind == lexer.TokenNot || p.peekKeyword("not") {
		if p.peek().Kind == lexer.TokenNot {
			if _, err := p.expect(lexer.TokenNot); err != nil {
				return nil, err
			}
		} else {
			if _, err := p.expectKeyword("not"); err != nil {
				return nil, err
			}
		}

		rightStart := len(p.astTokens)
		p.consumeSpaces()

		if p.peek().Kind == lexer.TokenLParen {
			inner, openIdx, closeIdx, err := p.parseParenthesizedExpression()
			if err != nil {
				return nil, err
			}
			first := rightStart
			if first < 0 || first > closeIdx {
				first = openIdx
			}
			node := ast.NewNode("ASTNotNode", ast.NodeIDs["ASTNotNode"], first, closeIdx)
			node.AddChild(inner)
			return node, nil
		}

		operand, err := p.parseUnaryExpression()
		if err != nil {
			return nil, err
		}
		first := rightStart
		if first < 0 || first > operand.LastIdx {
			first = operand.FirstIdx
		}
		node := ast.NewNode("ASTNotNode", ast.NodeIDs["ASTNotNode"], first, operand.LastIdx)
		node.AddChild(operand)
		return node, nil
	}

	if p.peek().Kind == lexer.TokenMinus {
		if _, err := p.expect(lexer.TokenMinus); err != nil {
			return nil, err
		}

		rightStart := len(p.astTokens)
		p.consumeSpaces()

		if p.peek().Kind == lexer.TokenLParen {
			inner, openIdx, closeIdx, err := p.parseParenthesizedExpression()
			if err != nil {
				return nil, err
			}
			first := rightStart
			if first < 0 || first > closeIdx {
				first = openIdx
			}
			node := ast.NewNode("ASTNegateNode", ast.NodeIDs["ASTNegateNode"], first, closeIdx)
			node.AddChild(inner)
			return node, nil
		}

		operand, err := p.parsePrimaryExpression()
		if err != nil {
			return nil, err
		}
		first := rightStart
		if first < 0 || first > operand.LastIdx {
			first = operand.FirstIdx
		}
		node := ast.NewNode("ASTNegateNode", ast.NodeIDs["ASTNegateNode"], first, operand.LastIdx)
		node.AddChild(operand)
		return node, nil
	}

	return p.parsePrimaryExpression()
}

func (p *Parser) parsePrimaryExpression() (*ast.Node, error) {
	tok := p.peek()
	switch tok.Kind {
	case lexer.TokenNumber:
		return p.parseNumericLiteral()
	case lexer.TokenString:
		idx, err := p.expect(lexer.TokenString)
		if err != nil {
			return nil, err
		}
		return ast.NewNode("ASTStringLiteral", ast.NodeIDs["ASTStringLiteral"], idx, idx), nil
	case lexer.TokenReference:
		return p.parseReference()
	case lexer.TokenLBracket:
		return p.parseArrayOrRange()
	case lexer.TokenLBrace:
		return p.parseMap()
	case lexer.TokenLParen:
		inner, _, _, err := p.parseParenthesizedExpression()
		if err != nil {
			return nil, err
		}
		return inner, nil
	case lexer.TokenIdentifier:
		if p.peekKeyword("true") || p.peekKeyword("false") {
			return p.parseBooleanLiteral()
		}
		// Java parser maps bare identifier/null as ASTReference in this path.
		idx, err := p.expect(lexer.TokenIdentifier)
		if err != nil {
			return nil, err
		}
		return ast.NewNode("ASTReference", ast.NodeIDs["ASTReference"], idx, idx), nil
	default:
		return nil, fmt.Errorf("unsupported expression token at pos=%d: %q", tok.Pos, tok.Image)
	}
}

func (p *Parser) extendTopUnaryExprLast(expr *ast.Node, last int) {
	if expr == nil || len(expr.Children) == 0 {
		return
	}
	node := expr.Children[0]
	for node != nil && (node.Name == "ASTNotNode" || node.Name == "ASTNegateNode") {
		node.LastIdx = last
		if len(node.Children) == 0 {
			break
		}
		node = node.Children[0]
	}
}

func (p *Parser) parseParenthesizedExpression() (*ast.Node, int, int, error) {
	openIdx, err := p.expect(lexer.TokenLParen)
	if err != nil {
		return nil, -1, -1, err
	}
	p.consumeSpaces()
	inner, err := p.parseExpression()
	if err != nil {
		return nil, -1, -1, err
	}
	p.consumeSpaces()
	closeIdx, err := p.expect(lexer.TokenRParen)
	if err != nil {
		return nil, -1, -1, err
	}
	return inner, openIdx, closeIdx, nil
}

func (p *Parser) parseNumericLiteral() (*ast.Node, error) {
	idx, err := p.expect(lexer.TokenNumber)
	if err != nil {
		return nil, err
	}
	img := p.astTokens[idx].Image
	if strings.Contains(img, ".") || strings.ContainsAny(img, "eE") {
		return ast.NewNode("ASTFloatingPointLiteral", ast.NodeIDs["ASTFloatingPointLiteral"], idx, idx), nil
	}
	return ast.NewNode("ASTIntegerLiteral", ast.NodeIDs["ASTIntegerLiteral"], idx, idx), nil
}

func (p *Parser) parseBooleanLiteral() (*ast.Node, error) {
	if p.peekKeyword("true") {
		idx, err := p.expectKeyword("true")
		if err != nil {
			return nil, err
		}
		return ast.NewNode("ASTTrue", ast.NodeIDs["ASTTrue"], idx, idx), nil
	}
	if p.peekKeyword("false") {
		idx, err := p.expectKeyword("false")
		if err != nil {
			return nil, err
		}
		return ast.NewNode("ASTFalse", ast.NodeIDs["ASTFalse"], idx, idx), nil
	}
	return nil, fmt.Errorf("expected boolean literal at pos=%d", p.peek().Pos)
}

func (p *Parser) parseArrayOrRange() (*ast.Node, error) {
	openIdx, err := p.expect(lexer.TokenLBracket)
	if err != nil {
		return nil, err
	}

	p.consumeSpaces()
	if p.peek().Kind == lexer.TokenRBracket {
		closeIdx, e := p.expect(lexer.TokenRBracket)
		if e != nil {
			return nil, e
		}
		return ast.NewNode("ASTObjectArray", ast.NodeIDs["ASTObjectArray"], openIdx, closeIdx), nil
	}

	firstParam, err := p.parseParameter()
	if err != nil {
		return nil, err
	}
	p.consumeSpaces()

	if p.peek().Kind == lexer.TokenDoubleDot {
		if _, err = p.expect(lexer.TokenDoubleDot); err != nil {
			return nil, err
		}
		p.consumeSpaces()
		second, e := p.parseRangeBound()
		if e != nil {
			return nil, e
		}
		p.consumeSpaces()
		closeIdx, e := p.expect(lexer.TokenRBracket)
		if e != nil {
			return nil, e
		}
		node := ast.NewNode("ASTIntegerRange", ast.NodeIDs["ASTIntegerRange"], openIdx, closeIdx)
		node.AddChild(firstParam)
		node.AddChild(second)
		return node, nil
	}

	items := []*ast.Node{firstParam}
	for p.peek().Kind == lexer.TokenComma {
		if _, err = p.expect(lexer.TokenComma); err != nil {
			return nil, err
		}
		p.consumeSpaces()
		item, e := p.parseParameter()
		if e != nil {
			return nil, e
		}
		items = append(items, item)
		p.consumeSpaces()
	}

	closeIdx, err := p.expect(lexer.TokenRBracket)
	if err != nil {
		return nil, err
	}
	node := ast.NewNode("ASTObjectArray", ast.NodeIDs["ASTObjectArray"], openIdx, closeIdx)
	for _, item := range items {
		node.AddChild(item)
	}
	return node, nil
}

func (p *Parser) parseMap() (*ast.Node, error) {
	openIdx, err := p.expect(lexer.TokenLBrace)
	if err != nil {
		return nil, err
	}

	p.consumeSpaces()
	if p.peek().Kind == lexer.TokenRBrace {
		closeIdx, e := p.expect(lexer.TokenRBrace)
		if e != nil {
			return nil, e
		}
		return ast.NewNode("ASTMap", ast.NodeIDs["ASTMap"], openIdx, closeIdx), nil
	}

	children := make([]*ast.Node, 0)
	for {
		key, e := p.parseParameter()
		if e != nil {
			return nil, e
		}
		p.consumeSpaces()
		if _, e = p.expect(lexer.TokenColon); e != nil {
			return nil, e
		}
		p.consumeSpaces()
		val, e := p.parseParameter()
		if e != nil {
			return nil, e
		}
		children = append(children, key, val)

		p.consumeSpaces()
		if p.peek().Kind != lexer.TokenComma {
			break
		}
		if _, e = p.expect(lexer.TokenComma); e != nil {
			return nil, e
		}
		p.consumeSpaces()
	}

	closeIdx, err := p.expect(lexer.TokenRBrace)
	if err != nil {
		return nil, err
	}

	node := ast.NewNode("ASTMap", ast.NodeIDs["ASTMap"], openIdx, closeIdx)
	for _, child := range children {
		node.AddChild(child)
	}
	return node, nil
}

func (p *Parser) parseRangeBound() (*ast.Node, error) {
	p.consumeSpaces()
	switch p.peek().Kind {
	case lexer.TokenReference:
		return p.parseReference()
	case lexer.TokenIdentifier:
		idx, err := p.expect(lexer.TokenIdentifier)
		if err != nil {
			return nil, err
		}
		return ast.NewNode("ASTReference", ast.NodeIDs["ASTReference"], idx, idx), nil
	case lexer.TokenNumber:
		node, err := p.parseNumericLiteral()
		if err != nil {
			return nil, err
		}
		if node.Name != "ASTIntegerLiteral" {
			return nil, fmt.Errorf("integer range bound must be integer at pos=%d", p.peek().Pos)
		}
		return node, nil
	default:
		return nil, fmt.Errorf("unsupported range bound at pos=%d image=%q", p.peek().Pos, p.peek().Image)
	}
}

func (p *Parser) parseParameter() (*ast.Node, error) {
	p.consumeSpaces()
	switch p.peek().Kind {
	case lexer.TokenString:
		idx, err := p.expect(lexer.TokenString)
		if err != nil {
			return nil, err
		}
		return ast.NewNode("ASTStringLiteral", ast.NodeIDs["ASTStringLiteral"], idx, idx), nil
	case lexer.TokenNumber:
		return p.parseNumericLiteral()
	case lexer.TokenReference:
		return p.parseReference()
	case lexer.TokenLBracket:
		return p.parseArrayOrRange()
	case lexer.TokenLBrace:
		return p.parseMap()
	case lexer.TokenIdentifier:
		if p.peekKeyword("true") || p.peekKeyword("false") {
			return p.parseBooleanLiteral()
		}
		// Java parser treats bare identifiers (for example null) as references.
		idx, err := p.expect(lexer.TokenIdentifier)
		if err != nil {
			return nil, err
		}
		return ast.NewNode("ASTReference", ast.NodeIDs["ASTReference"], idx, idx), nil
	default:
		return nil, fmt.Errorf("unsupported parameter token at pos=%d image=%q", p.peek().Pos, p.peek().Image)
	}
}

func (p *Parser) validateDirectiveArgs(name string, args []*ast.Node) error {
	switch name {
	case "parse":
		if len(args) != 1 {
			return fmt.Errorf("the #parse directive requires one argument")
		}
		if args[0].Name == "ASTWord" {
			return fmt.Errorf("the argument to #parse is of the wrong type")
		}
	case "stop":
		if len(args) > 1 {
			return fmt.Errorf("the #stop directive only accepts a single message parameter")
		}
	case "break":
		if len(args) > 1 {
			return fmt.Errorf("the #break directive takes only a single, optional Scope argument")
		}
	}
	return nil
}

func (p *Parser) consumeDirectivePostfixNewline(keepTokens bool) bool {
	scan := p.pos
	sawNewline := false
	for scan < len(p.tokens) {
		tok := p.tokens[scan]
		if tok.Kind != lexer.TokenText || !isWhitespaceOnly(tok.Image) {
			break
		}
		if strings.Contains(tok.Image, "\n") || strings.Contains(tok.Image, "\r") {
			sawNewline = true
			scan++
			break
		}
		scan++
	}
	if !sawNewline {
		return false
	}
	for p.pos < scan {
		if keepTokens {
			_, _ = p.consumeAny()
		} else {
			p.pos++
		}
	}
	return true
}

func (p *Parser) shouldKeepPostfixNewlineForNestedStop() bool {
	scan := p.pos
	sawNewline := false
	for scan < len(p.tokens) {
		tok := p.tokens[scan]
		if tok.Kind != lexer.TokenText || !isWhitespaceOnly(tok.Image) {
			break
		}
		if strings.Contains(tok.Image, "\n") || strings.Contains(tok.Image, "\r") {
			sawNewline = true
			scan++
			break
		}
		scan++
	}
	if !sawNewline || scan >= len(p.tokens) {
		return false
	}
	next := p.tokens[scan]
	if next.Kind != lexer.TokenDirective {
		if next.Kind == lexer.TokenText && isWhitespaceOnly(next.Image) && hasLineBreak(next.Image) {
			// Keep one newline on control nodes when a blank line follows.
			return true
		}
		return false
	}
	switch canonicalDirective(next.Image) {
	case "#end", "#else", "#elseif":
		return true
	default:
		return false
	}
}

func (p *Parser) consumeLineDirectivePostfixNewline() (int, bool) {
	scan := p.pos
	sawNewline := false
	for scan < len(p.tokens) {
		tok := p.tokens[scan]
		if tok.Kind != lexer.TokenText || !isWhitespaceOnly(tok.Image) {
			break
		}
		if strings.Contains(tok.Image, "\n") || strings.Contains(tok.Image, "\r") {
			sawNewline = true
			scan++
			break
		}
		scan++
	}
	if !sawNewline {
		return -1, false
	}

	keepTokens := true
	lastConsumed := -1
	for p.pos < scan {
		if keepTokens {
			var err error
			lastConsumed, err = p.consumeAny()
			if err != nil {
				return -1, false
			}
		} else {
			p.pos++
		}
	}
	return lastConsumed, true
}

func isNodeBoundaryToken(tok lexer.Token) bool {
	switch tok.Kind {
	case lexer.TokenEOF,
		lexer.TokenDirectiveSet,
		lexer.TokenDirective,
		lexer.TokenEscapedDirective,
		lexer.TokenReference,
		lexer.TokenEscape,
		lexer.TokenCommentLine,
		lexer.TokenComment,
		lexer.TokenTextblock:
		return true
	default:
		return false
	}
}

func canonicalDirective(image string) string {
	if strings.HasPrefix(image, "#{") && strings.HasSuffix(image, "}") {
		inner := image[2 : len(image)-1]
		return "#" + inner
	}
	return image
}

func directiveName(image string) string {
	if strings.HasPrefix(image, "#{") && strings.HasSuffix(image, "}") {
		return image[2 : len(image)-1]
	}
	if strings.HasPrefix(image, "#") {
		return image[1:]
	}
	return image
}

func (p *Parser) isDirectiveToken(canonical string) bool {
	tok := p.peek()
	return tok.Kind == lexer.TokenDirective && canonicalDirective(tok.Image) == canonical
}

func (p *Parser) expectDirectiveCanonical(image string) (int, error) {
	tok := p.peek()
	if tok.Kind != lexer.TokenDirective || canonicalDirective(tok.Image) != image {
		return -1, fmt.Errorf("expected directive %q at pos=%d, got %q", image, tok.Pos, canonicalDirective(tok.Image))
	}
	return p.consumeAny()
}

func (p *Parser) consumeSpaces() int {
	count := 0
	for {
		tok := p.peek()
		if tok.Kind != lexer.TokenText {
			break
		}
		if !isWhitespaceOnly(tok.Image) {
			break
		}
		_, _ = p.consumeAny()
		count++
	}
	return count
}

func isWhitespaceOnly(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r != ' ' && r != '\t' && r != '\n' && r != '\r' {
			return false
		}
	}
	return true
}

func (p *Parser) peek() lexer.Token {
	if p.pos >= len(p.tokens) {
		return lexer.Token{Kind: lexer.TokenEOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) peekN(n int) lexer.Token {
	idx := p.pos + n
	if idx < 0 || idx >= len(p.tokens) {
		return lexer.Token{Kind: lexer.TokenEOF}
	}
	return p.tokens[idx]
}

func (p *Parser) peekAfterWhitespace() (lexer.Token, int) {
	offset := 0
	for {
		tok := p.peekN(offset)
		if tok.Kind != lexer.TokenText || !isWhitespaceOnly(tok.Image) {
			return tok, offset
		}
		offset++
	}
}

func (p *Parser) consumeWhitespaceTokens(count int) {
	for i := 0; i < count; i++ {
		_, _ = p.consumeAny()
	}
}

func (p *Parser) peekKeyword(word string) bool {
	tok := p.peek()
	return tok.Kind == lexer.TokenIdentifier && tok.Image == word
}

func (p *Parser) expectKeyword(word string) (int, error) {
	tok := p.peek()
	if tok.Kind != lexer.TokenIdentifier || tok.Image != word {
		return -1, fmt.Errorf("expected keyword %q at pos=%d, got %q", word, tok.Pos, tok.Image)
	}
	return p.consumeAny()
}

func (p *Parser) expect(k lexer.Kind) (int, error) {
	tok := p.peek()
	if tok.Kind != k {
		return -1, fmt.Errorf("expected kind=%v at pos=%d, got kind=%v image=%q", k, tok.Pos, tok.Kind, tok.Image)
	}
	return p.consumeAny()
}

func (p *Parser) consumeAny() (int, error) {
	if p.pos >= len(p.tokens) {
		return -1, fmt.Errorf("unexpected end of tokens")
	}
	tok := p.tokens[p.pos]
	p.pos++
	if tok.Kind == lexer.TokenEOF {
		return -1, nil
	}
	p.astTokens = append(p.astTokens, ast.Token{Image: tok.Image})
	return len(p.astTokens) - 1, nil
}

func (p *Parser) consumeTextRun() (int, error) {
	if isNodeBoundaryToken(p.peek()) {
		return -1, fmt.Errorf("expected text token at pos=%d, got kind=%v", p.peek().Pos, p.peek().Kind)
	}
	// Keep standalone whitespace chunks (especially newlines) as one token.
	// Java keeps consecutive line-break chunks split in this scenario.
	if p.peek().Kind == lexer.TokenText &&
		isWhitespaceOnly(p.peek().Image) &&
		hasLineBreak(p.peek().Image) {
		img := p.peek().Image
		p.pos++
		p.astTokens = append(p.astTokens, ast.Token{Image: img})
		return len(p.astTokens) - 1, nil
	}

	var b strings.Builder
	for {
		tok := p.peek()
		if isNodeBoundaryToken(tok) {
			break
		}
		if b.Len() > 0 &&
			tok.Kind == lexer.TokenMul &&
			strings.HasSuffix(b.String(), "\n") &&
			p.peekN(1).Kind == lexer.TokenText &&
			p.peekN(1).Image == "#" {
			// Keep text before "\n*#" separate to match Java text-node splits.
			break
		}
		if b.Len() > 0 &&
			tok.Kind == lexer.TokenText &&
			isWhitespaceOnly(tok.Image) &&
			hasLineBreak(tok.Image) &&
			isWhitespaceOnly(b.String()) &&
			isNodeBoundaryToken(p.peekN(1)) {
			break
		}
		b.WriteString(tok.Image)
		p.pos++
		if tok.Kind == lexer.TokenMul && p.peek().Kind == lexer.TokenText && p.peek().Image == "#" {
			// Keep "*" split from following "#" in edge templates like settest.vm.
			break
		}
		if tok.Kind == lexer.TokenText &&
			hasLineBreak(tok.Image) &&
			strings.HasPrefix(b.String(), "#") &&
			p.peek().Kind == lexer.TokenText &&
			p.peek().Image == "\n" {
			// Keep "#\\n" and the next standalone "\\n" as separate text nodes.
			break
		}
		if tok.Kind == lexer.TokenText && hasLineBreak(tok.Image) && p.shouldSplitAfterLineBreak(tok) {
			// Keep indentation before a boundary node as a separate text node.
			break
		}
	}
	if b.Len() == 0 {
		return -1, fmt.Errorf("empty text run at pos=%d", p.peek().Pos)
	}
	p.astTokens = append(p.astTokens, ast.Token{Image: b.String()})
	return len(p.astTokens) - 1, nil
}

func (p *Parser) shouldSplitAfterLineBreak(tok lexer.Token) bool {
	if p.shouldSplitBeforeBoundaryIndent() {
		return true
	}
	if !p.remainingRunHasLineBreakBeforeBoundary() {
		return true
	}
	next := p.peek()
	if next.Kind == lexer.TokenIdentifier {
		line := p.lineBeforeLastBreak(tok)
		hasAlnum := false
		hasNonSpace := false
		for i := 0; i < len(line); i++ {
			r := line[i]
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
				hasAlnum = true
			}
			if r != ' ' && r != '\t' && r != '\r' && r != '\n' {
				hasNonSpace = true
			}
		}
		return hasNonSpace && !hasAlnum
	}
	return false
}

func (p *Parser) shouldSplitBeforeBoundaryIndent() bool {
	next := p.peek()
	if next.Kind != lexer.TokenText || !isWhitespaceOnly(next.Image) || hasLineBreak(next.Image) {
		return false
	}
	after := p.peekN(1)
	return isNodeBoundaryToken(after)
}

func (p *Parser) lineBeforeLastBreak(tok lexer.Token) string {
	offset := lastLineBreakOffset(tok.Image)
	if offset < 0 {
		return ""
	}
	abs := tok.Pos + offset
	if abs < 0 || abs > len(p.input) {
		return ""
	}
	start := abs - 1
	for start >= 0 {
		ch := p.input[start]
		if ch == '\n' || ch == '\r' {
			break
		}
		start--
	}
	return p.input[start+1 : abs]
}

func lastLineBreakOffset(s string) int {
	last := -1
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' || s[i] == '\r' {
			last = i
		}
	}
	return last
}

func (p *Parser) remainingRunHasLineBreakBeforeBoundary() bool {
	for offset := 0; ; offset++ {
		tok := p.peekN(offset)
		if isNodeBoundaryToken(tok) {
			return false
		}
		if tok.Kind == lexer.TokenText && hasLineBreak(tok.Image) {
			return true
		}
	}
}

func (p *Parser) scanHorizontalWhitespaceBeforeStop(stop map[string]bool) (int, bool) {
	scan := p.pos
	if scan >= len(p.tokens) {
		return 0, false
	}
	for scan < len(p.tokens) {
		tok := p.tokens[scan]
		if tok.Kind != lexer.TokenText || !isWhitespaceOnly(tok.Image) || hasLineBreak(tok.Image) {
			break
		}
		scan++
	}
	if scan == p.pos || scan >= len(p.tokens) {
		return 0, false
	}
	tok := p.tokens[scan]
	if tok.Kind != lexer.TokenDirective {
		return 0, false
	}
	if !stop[canonicalDirective(tok.Image)] {
		return 0, false
	}
	return scan, true
}

func rangeFromChildren(children []*ast.Node) (int, int) {
	if len(children) == 0 {
		return -1, -1
	}
	first := children[0].FirstIdx
	last := children[0].LastIdx
	for _, child := range children[1:] {
		if child.FirstIdx < first || first < 0 {
			first = child.FirstIdx
		}
		if child.LastIdx > last {
			last = child.LastIdx
		}
	}
	return first, last
}

func hasLineBreak(s string) bool {
	return strings.Contains(s, "\n") || strings.Contains(s, "\r")
}

func resolveDeferredLastIdx(node *ast.Node, tail int) {
	if node == nil {
		return
	}
	if node.LastIdx == deferredLastIdx {
		node.LastIdx = tail
	}
	for _, child := range node.Children {
		resolveDeferredLastIdx(child, tail)
	}
}
