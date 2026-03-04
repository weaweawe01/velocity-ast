package lexer

import "testing"

func TestLexStandaloneBoundaryCharsProgress(t *testing.T) {
	tokens, err := Lex(`\&`)
	if err != nil {
		t.Fatalf("lex failed: %v", err)
	}
	if len(tokens) != 3 {
		t.Fatalf("expected 3 tokens (text, text, eof), got %d", len(tokens))
	}
	if tokens[0].Kind != TokenText || tokens[0].Image != `\` {
		t.Fatalf("expected first token text '\\\\', got kind=%v image=%q", tokens[0].Kind, tokens[0].Image)
	}
	if tokens[1].Kind != TokenText || tokens[1].Image != "&" {
		t.Fatalf("expected second token text '&', got kind=%v image=%q", tokens[1].Kind, tokens[1].Image)
	}
	if tokens[2].Kind != TokenEOF {
		t.Fatalf("expected last token EOF, got kind=%v", tokens[2].Kind)
	}
}

func TestLexEscapedDirectivePrefixNoHang(t *testing.T) {
	tokens, err := Lex(`\#if`)
	if err != nil {
		t.Fatalf("lex failed: %v", err)
	}
	if len(tokens) != 2 {
		t.Fatalf("expected escaped directive + eof, got %d tokens", len(tokens))
	}
	if tokens[0].Kind != TokenEscapedDirective || tokens[0].Image != "#if" {
		t.Fatalf("expected first token escaped directive '#if', got kind=%v image=%q", tokens[0].Kind, tokens[0].Image)
	}
	if tokens[1].Kind != TokenEOF {
		t.Fatalf("expected eof token, got kind=%v image=%q", tokens[1].Kind, tokens[1].Image)
	}
}

func TestLexReferenceInsideQuotedPlainText(t *testing.T) {
	tokens, err := Lex(`<input value="$!email">`)
	if err != nil {
		t.Fatalf("lex failed: %v", err)
	}
	foundRef := false
	for _, tok := range tokens {
		if tok.Kind == TokenReference && tok.Image == "$!email" {
			foundRef = true
			break
		}
	}
	if !foundRef {
		t.Fatalf("expected to lex $!email as reference in plain text context")
	}
}

func TestLexStringLiteralInDirectiveExpression(t *testing.T) {
	tokens, err := Lex(`#set($x = "abc")`)
	if err != nil {
		t.Fatalf("lex failed: %v", err)
	}
	foundString := false
	for _, tok := range tokens {
		if tok.Kind == TokenString && tok.Image == `"abc"` {
			foundString = true
			break
		}
	}
	if !foundString {
		t.Fatalf("expected to lex \"abc\" as string literal in expression context")
	}
}

func TestLexEscapedDirectiveAndReferenceAsText(t *testing.T) {
	tokens, err := Lex(`\#if \$a`)
	if err != nil {
		t.Fatalf("lex failed: %v", err)
	}
	gotDirective := false
	gotReference := false
	gotEscapedReference := false
	for _, tok := range tokens {
		if tok.Kind == TokenDirective {
			gotDirective = true
		}
		if tok.Kind == TokenReference {
			gotReference = true
			if tok.Image == `\$a` {
				gotEscapedReference = true
			}
		}
	}
	if gotDirective {
		t.Fatalf("escaped directive should not be lexed as TokenDirective")
	}
	if !gotReference || !gotEscapedReference {
		t.Fatalf("escaped reference should be lexed as TokenReference with image \\\\$a")
	}
}

func TestLexEscapedSetDirectiveAsText(t *testing.T) {
	tokens, err := Lex(`\#set($x = 1)`)
	if err != nil {
		t.Fatalf("lex failed: %v", err)
	}
	for _, tok := range tokens {
		if tok.Kind == TokenDirectiveSet {
			t.Fatalf("escaped #set should not be lexed as TokenDirectiveSet")
		}
		if tok.Kind == TokenDirective {
			t.Fatalf("escaped #set should not be lexed as TokenDirective")
		}
	}
}

func TestLexEscapedUnknownDirectiveKeepsBackslash(t *testing.T) {
	tokens, err := Lex(`\#foo`)
	if err != nil {
		t.Fatalf("lex failed: %v", err)
	}
	if len(tokens) != 2 {
		t.Fatalf("expected escaped directive + eof, got %d tokens", len(tokens))
	}
	if tokens[0].Kind != TokenEscapedDirective || tokens[0].Image != `\#foo` {
		t.Fatalf("expected escaped unknown directive image '\\\\#foo', got kind=%v image=%q", tokens[0].Kind, tokens[0].Image)
	}
}

func TestLexBackslashRunBeforeReferenceStaysReference(t *testing.T) {
	tokens, err := Lex(`\\$a`)
	if err != nil {
		t.Fatalf("lex failed: %v", err)
	}
	if len(tokens) != 2 {
		t.Fatalf("expected reference + eof, got %d tokens", len(tokens))
	}
	if tokens[0].Kind != TokenReference || tokens[0].Image != `\\$a` {
		t.Fatalf("expected token reference '\\\\\\\\$a', got kind=%v image=%q", tokens[0].Kind, tokens[0].Image)
	}

	tokens, err = Lex(`\\\$a`)
	if err != nil {
		t.Fatalf("lex failed: %v", err)
	}
	if len(tokens) != 2 {
		t.Fatalf("expected reference + eof, got %d tokens", len(tokens))
	}
	if tokens[0].Kind != TokenReference || tokens[0].Image != `\\\$a` {
		t.Fatalf("expected token reference '\\\\\\\\\\\\$a', got kind=%v image=%q", tokens[0].Kind, tokens[0].Image)
	}
}

func TestLexNegativeRangeEndAsNumber(t *testing.T) {
	tokens, err := Lex(`[-4..-5]`)
	if err != nil {
		t.Fatalf("lex failed: %v", err)
	}
	found := false
	for i := 0; i+2 < len(tokens); i++ {
		if tokens[i].Kind == TokenNumber && tokens[i].Image == "-4" &&
			tokens[i+1].Kind == TokenDoubleDot &&
			tokens[i+2].Kind == TokenNumber && tokens[i+2].Image == "-5" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected to lex negative range end as number -5, got %#v", tokens)
	}
}

func TestLexLineCommentKeepsRemainderWithNewline(t *testing.T) {
	tokens, err := Lex("####\n")
	if err != nil {
		t.Fatalf("lex failed: %v", err)
	}
	if len(tokens) != 3 {
		t.Fatalf("expected comment + tail + eof, got %d tokens", len(tokens))
	}
	if tokens[0].Kind != TokenCommentLine || tokens[0].Image != "##" {
		t.Fatalf("expected first token line comment '##', got kind=%v image=%q", tokens[0].Kind, tokens[0].Image)
	}
	if tokens[1].Kind != TokenText || tokens[1].Image != "##\n" {
		t.Fatalf("expected second token '##\\n', got kind=%v image=%q", tokens[1].Kind, tokens[1].Image)
	}
}

func TestLexDollarLineComment(t *testing.T) {
	tokens, err := Lex("$##\n")
	if err != nil {
		t.Fatalf("lex failed: %v", err)
	}
	if len(tokens) != 3 {
		t.Fatalf("expected comment + newline + eof, got %d tokens", len(tokens))
	}
	if tokens[0].Kind != TokenCommentLine || tokens[0].Image != "$##" {
		t.Fatalf("expected first token line comment '$##', got kind=%v image=%q", tokens[0].Kind, tokens[0].Image)
	}
	if tokens[1].Kind != TokenText || tokens[1].Image != "\n" {
		t.Fatalf("expected second token '\\n', got kind=%v image=%q", tokens[1].Kind, tokens[1].Image)
	}
}

func TestLexSplitsConsecutiveNewlines(t *testing.T) {
	tokens, err := Lex("\n\n")
	if err != nil {
		t.Fatalf("lex failed: %v", err)
	}
	if len(tokens) != 3 {
		t.Fatalf("expected 2 text tokens + eof, got %d", len(tokens))
	}
	if tokens[0].Kind != TokenText || tokens[0].Image != "\n" {
		t.Fatalf("expected first newline token, got kind=%v image=%q", tokens[0].Kind, tokens[0].Image)
	}
	if tokens[1].Kind != TokenText || tokens[1].Image != "\n" {
		t.Fatalf("expected second newline token, got kind=%v image=%q", tokens[1].Kind, tokens[1].Image)
	}
}

func TestLexDirectiveMustStartWithIdentifier(t *testing.T) {
	tokens, err := Lex("#333333")
	if err != nil {
		t.Fatalf("lex failed: %v", err)
	}
	if len(tokens) != 3 {
		t.Fatalf("expected text + number + eof, got %d tokens", len(tokens))
	}
	if tokens[0].Kind != TokenText || tokens[0].Image != "#" {
		t.Fatalf("expected first token '#', got kind=%v image=%q", tokens[0].Kind, tokens[0].Image)
	}
	if tokens[1].Kind != TokenNumber || tokens[1].Image != "333333" {
		t.Fatalf("expected second token number '333333', got kind=%v image=%q", tokens[1].Kind, tokens[1].Image)
	}
}

func TestLexReferenceMustStartWithIdentifier(t *testing.T) {
	tokens, err := Lex("$100")
	if err != nil {
		t.Fatalf("lex failed: %v", err)
	}
	foundReference := false
	for _, tok := range tokens {
		if tok.Kind == TokenReference {
			foundReference = true
			break
		}
	}
	if foundReference {
		t.Fatalf("did not expect $100 to be tokenized as reference: %#v", tokens)
	}
}

func TestLexDoubleDollarReference(t *testing.T) {
	tokens, err := Lex("$$provider")
	if err != nil {
		t.Fatalf("lex failed: %v", err)
	}
	if len(tokens) != 2 {
		t.Fatalf("expected reference + eof, got %d tokens", len(tokens))
	}
	if tokens[0].Kind != TokenReference || tokens[0].Image != "$$provider" {
		t.Fatalf("expected reference '$$provider', got kind=%v image=%q", tokens[0].Kind, tokens[0].Image)
	}
}
