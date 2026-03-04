package lexer

import (
	"fmt"
	"strings"
	"unicode"
)

type Kind int

const (
	TokenEOF Kind = iota
	TokenDirectiveSet
	TokenDirective
	TokenReference
	TokenIdentifier
	TokenNumber
	TokenString
	TokenDot
	TokenDoubleDot
	TokenLParen
	TokenRParen
	TokenLBracket
	TokenRBracket
	TokenLBrace
	TokenRBrace
	TokenComma
	TokenColon
	TokenPipe
	TokenOrOr
	TokenAndAnd
	TokenPlus
	TokenMinus
	TokenMul
	TokenDiv
	TokenMod
	TokenEqEq
	TokenNotEq
	TokenNot
	TokenGt
	TokenLt
	TokenGe
	TokenLe
	TokenEquals
	TokenEscape
	TokenEscapedDirective
	TokenComment
	TokenCommentLine
	TokenTextblock
	TokenText
)

type Token struct {
	Kind  Kind
	Image string
	Pos   int
}

func Lex(input string) ([]Token, error) {
	tokens := make([]Token, 0, len(input)/2)
	i := 0
	exprMode := false
	exprParenDepth := 0
	pendingDirectiveArgs := false
	pendingMethodArgs := false
	for i < len(input) {
		if img, next, ok := scanSetDirective(input, i); ok && !isEscapedByBackslash(input, i) {
			tokens = append(tokens, Token{Kind: TokenDirectiveSet, Image: img, Pos: i})
			exprMode = true
			exprParenDepth = 1
			pendingDirectiveArgs = false
			pendingMethodArgs = false
			i = next
			continue
		}

		switch {
		case backslashRunBeforeDollar(input, i):
			j := i
			for j < len(input) && input[j] == '\\' {
				j++
			}
			if ref, next, ok := scanReference(input, j); ok {
				tokens = append(tokens, Token{Kind: TokenReference, Image: input[i:j] + ref, Pos: i})
				pendingDirectiveArgs = false
				pendingMethodArgs = false
				i = next
				continue
			}
			// Preserve progress for malformed escape+reference prefixes.
			tokens = append(tokens, Token{Kind: TokenText, Image: "\\", Pos: i})
			pendingDirectiveArgs = false
			pendingMethodArgs = false
			i++

		case input[i] == '\\' && i+1 < len(input) && input[i+1] == '#':
			if dir, next, ok := scanDirective(input, i+1); ok {
				img := dir
				if !isKnownDirective(dir) {
					img = "\\" + dir
				}
				tokens = append(tokens, Token{Kind: TokenEscapedDirective, Image: img, Pos: i})
				pendingDirectiveArgs = false
				pendingMethodArgs = false
				i = next
				continue
			}
			tokens = append(tokens, Token{Kind: TokenText, Image: "\\", Pos: i})
			tokens = append(tokens, Token{Kind: TokenText, Image: "#", Pos: i + 1})
			pendingDirectiveArgs = false
			pendingMethodArgs = false
			i += 2

		case hasPrefixAt(input, i, "#[["):
			end := i + 3
			for end+2 < len(input) && input[end:end+3] != "]]#" {
				end++
			}
			if end+2 >= len(input) {
				return nil, fmt.Errorf("unterminated textblock at %d", i)
			}
			end += 3
			tokens = append(tokens, Token{Kind: TokenTextblock, Image: input[i:end], Pos: i})
			i = end

		case hasPrefixAt(input, i, "$##"):
			tokens = append(tokens, Token{Kind: TokenCommentLine, Image: "$##", Pos: i})
			i += 3
			for i < len(input) && input[i] != '\n' && input[i] != '\r' {
				i++
			}
			newlineStart := i
			if i < len(input) {
				if input[i] == '\r' && i+1 < len(input) && input[i+1] == '\n' {
					i += 2
				} else {
					i++
				}
			}
			if i > newlineStart {
				tokens = append(tokens, Token{Kind: TokenText, Image: input[newlineStart:i], Pos: newlineStart})
			}

		case hasPrefixAt(input, i, "##"):
			tokens = append(tokens, Token{Kind: TokenCommentLine, Image: "##", Pos: i})
			i += 2
			lineStart := i
			for i < len(input) && input[i] != '\n' && input[i] != '\r' {
				i++
			}
			lineEnd := i
			newlineStart := i
			if i < len(input) {
				if input[i] == '\r' && i+1 < len(input) && input[i+1] == '\n' {
					i += 2
				} else {
					i++
				}
			}
			lineBody := input[lineStart:lineEnd]
			if i > newlineStart {
				if strings.TrimSpace(lineBody) == "##" {
					idx := strings.Index(lineBody, "##")
					tokens = append(tokens, Token{Kind: TokenText, Image: lineBody[idx:idx+2] + input[newlineStart:i], Pos: lineStart + idx})
				} else {
					tokens = append(tokens, Token{Kind: TokenText, Image: input[newlineStart:i], Pos: newlineStart})
				}
			}

		case hasPrefixAt(input, i, "#*"):
			end := i + 2
			for end+1 < len(input) && input[end:end+2] != "*#" {
				end++
			}
			if end+1 >= len(input) {
				return nil, fmt.Errorf("unterminated block comment at %d", i)
			}
			end += 2
			// Java AST dumps comment node with token image "*#" for block/formal comments.
			tokens = append(tokens, Token{Kind: TokenComment, Image: "*#", Pos: i})
			i = end

		case hasPrefixAt(input, i, "${"):
			tokens = append(tokens, Token{Kind: TokenReference, Image: "${", Pos: i})
			i += 2

		case hasPrefixAt(input, i, "$!{"):
			tokens = append(tokens, Token{Kind: TokenReference, Image: "$!{", Pos: i})
			i += 3

		case hasPrefixAt(input, i, "\\\\") && isEscapableStart(input, i+2):
			tokens = append(tokens, Token{Kind: TokenEscape, Image: "\\\\", Pos: i})
			i += 2

		case input[i] == '#':
			if isEscapedByBackslash(input, i) {
				tokens = append(tokens, Token{Kind: TokenText, Image: "#", Pos: i})
				pendingDirectiveArgs = false
				pendingMethodArgs = false
				i++
				continue
			}
			if dir, next, ok := scanDirective(input, i); ok {
				tokens = append(tokens, Token{Kind: TokenDirective, Image: dir, Pos: i})
				pendingDirectiveArgs = false
				if !exprMode {
					k := skipWhitespace(input, next)
					if k < len(input) && input[k] == '(' {
						pendingDirectiveArgs = true
					}
				}
				pendingMethodArgs = false
				i = next
				continue
			}
			tokens = append(tokens, Token{Kind: TokenText, Image: "#", Pos: i})
			if pendingDirectiveArgs {
				pendingDirectiveArgs = false
			}
			if pendingMethodArgs {
				pendingMethodArgs = false
			}
			i++

		case input[i] == '$':
			if isEscapedByBackslash(input, i) {
				tokens = append(tokens, Token{Kind: TokenText, Image: "$", Pos: i})
				pendingDirectiveArgs = false
				pendingMethodArgs = false
				i++
				continue
			}
			if ref, next, ok := scanReference(input, i); ok {
				tokens = append(tokens, Token{Kind: TokenReference, Image: ref, Pos: i})
				i = next
				continue
			}
			// Keep invalid/incomplete references as plain text, matching Java's lenient text handling.
			if i+1 < len(input) && input[i+1] == '!' {
				tokens = append(tokens, Token{Kind: TokenText, Image: "$!", Pos: i})
				i += 2
			} else {
				tokens = append(tokens, Token{Kind: TokenText, Image: "$", Pos: i})
				i++
			}

		case input[i] == '"' || input[i] == '\'':
			if !exprMode {
				tokens = append(tokens, Token{Kind: TokenText, Image: input[i : i+1], Pos: i})
				i++
				continue
			}
			start := i
			quote := input[i]
			i++
			escaped := false
			for i < len(input) {
				ch := input[i]
				if escaped {
					escaped = false
					i++
					continue
				}
				if ch == '\\' {
					escaped = true
					i++
					continue
				}
				if ch == quote {
					i++
					break
				}
				i++
			}
			if i > len(input) || input[i-1] != quote {
				return nil, fmt.Errorf("unterminated string literal at %d", start)
			}
			tokens = append(tokens, Token{Kind: TokenString, Image: input[start:i], Pos: start})
			pendingDirectiveArgs = false
			pendingMethodArgs = false

		case isSignedNumberStart(input, i):
			start := i
			i = scanNumberEnd(input, i)
			tokens = append(tokens, Token{Kind: TokenNumber, Image: input[start:i], Pos: start})

		case isUnsignedNumberStart(input, i):
			start := i
			i = scanNumberEnd(input, i)
			tokens = append(tokens, Token{Kind: TokenNumber, Image: input[start:i], Pos: start})

		case isIdentifierStart(rune(input[i])):
			start := i
			i++
			for i < len(input) && isIdentifierPart(rune(input[i])) {
				i++
			}
			tok := Token{Kind: TokenIdentifier, Image: input[start:i], Pos: start}
			tokens = append(tokens, tok)
			if !exprMode && len(tokens) >= 2 && tokens[len(tokens)-2].Kind == TokenDot {
				k := skipWhitespace(input, i)
				if k < len(input) && input[k] == '(' {
					pendingMethodArgs = true
				}
			}
			pendingDirectiveArgs = false

		case hasPrefixAt(input, i, ".."):
			tokens = append(tokens, Token{Kind: TokenDoubleDot, Image: "..", Pos: i})
			pendingDirectiveArgs = false
			pendingMethodArgs = false
			i += 2

		case input[i] == '.':
			tokens = append(tokens, Token{Kind: TokenDot, Image: ".", Pos: i})
			pendingDirectiveArgs = false
			i++
		case input[i] == '(':
			tokens = append(tokens, Token{Kind: TokenLParen, Image: "(", Pos: i})
			if exprMode {
				exprParenDepth++
			} else if pendingDirectiveArgs || pendingMethodArgs {
				exprMode = true
				exprParenDepth = 1
			}
			pendingDirectiveArgs = false
			pendingMethodArgs = false
			i++
		case input[i] == ')':
			tokens = append(tokens, Token{Kind: TokenRParen, Image: ")", Pos: i})
			if exprMode {
				exprParenDepth--
				if exprParenDepth <= 0 {
					exprMode = false
					exprParenDepth = 0
				}
			}
			pendingDirectiveArgs = false
			pendingMethodArgs = false
			i++
		case input[i] == '[':
			tokens = append(tokens, Token{Kind: TokenLBracket, Image: "[", Pos: i})
			pendingDirectiveArgs = false
			pendingMethodArgs = false
			i++
		case input[i] == ']':
			tokens = append(tokens, Token{Kind: TokenRBracket, Image: "]", Pos: i})
			pendingDirectiveArgs = false
			pendingMethodArgs = false
			i++
		case input[i] == '{':
			tokens = append(tokens, Token{Kind: TokenLBrace, Image: "{", Pos: i})
			pendingDirectiveArgs = false
			pendingMethodArgs = false
			i++
		case input[i] == '}':
			tokens = append(tokens, Token{Kind: TokenRBrace, Image: "}", Pos: i})
			pendingDirectiveArgs = false
			pendingMethodArgs = false
			i++
		case input[i] == ',':
			tokens = append(tokens, Token{Kind: TokenComma, Image: ",", Pos: i})
			pendingDirectiveArgs = false
			pendingMethodArgs = false
			i++
		case input[i] == ':':
			tokens = append(tokens, Token{Kind: TokenColon, Image: ":", Pos: i})
			pendingDirectiveArgs = false
			pendingMethodArgs = false
			i++
		case hasPrefixAt(input, i, "||"):
			tokens = append(tokens, Token{Kind: TokenOrOr, Image: "||", Pos: i})
			pendingDirectiveArgs = false
			pendingMethodArgs = false
			i += 2
		case hasPrefixAt(input, i, "&&"):
			tokens = append(tokens, Token{Kind: TokenAndAnd, Image: "&&", Pos: i})
			pendingDirectiveArgs = false
			pendingMethodArgs = false
			i += 2
		case input[i] == '|':
			tokens = append(tokens, Token{Kind: TokenPipe, Image: "|", Pos: i})
			pendingDirectiveArgs = false
			pendingMethodArgs = false
			i++
		case hasPrefixAt(input, i, ">="):
			tokens = append(tokens, Token{Kind: TokenGe, Image: ">=", Pos: i})
			pendingDirectiveArgs = false
			pendingMethodArgs = false
			i += 2
		case hasPrefixAt(input, i, "<="):
			tokens = append(tokens, Token{Kind: TokenLe, Image: "<=", Pos: i})
			pendingDirectiveArgs = false
			pendingMethodArgs = false
			i += 2
		case hasPrefixAt(input, i, "!="):
			tokens = append(tokens, Token{Kind: TokenNotEq, Image: "!=", Pos: i})
			pendingDirectiveArgs = false
			pendingMethodArgs = false
			i += 2
		case input[i] == '!':
			tokens = append(tokens, Token{Kind: TokenNot, Image: "!", Pos: i})
			pendingDirectiveArgs = false
			pendingMethodArgs = false
			i++
		case hasPrefixAt(input, i, "=="):
			tokens = append(tokens, Token{Kind: TokenEqEq, Image: "==", Pos: i})
			pendingDirectiveArgs = false
			pendingMethodArgs = false
			i += 2
		case input[i] == '>':
			tokens = append(tokens, Token{Kind: TokenGt, Image: ">", Pos: i})
			pendingDirectiveArgs = false
			pendingMethodArgs = false
			i++
		case input[i] == '<':
			tokens = append(tokens, Token{Kind: TokenLt, Image: "<", Pos: i})
			pendingDirectiveArgs = false
			pendingMethodArgs = false
			i++
		case input[i] == '+':
			tokens = append(tokens, Token{Kind: TokenPlus, Image: "+", Pos: i})
			pendingDirectiveArgs = false
			pendingMethodArgs = false
			i++
		case input[i] == '-':
			tokens = append(tokens, Token{Kind: TokenMinus, Image: "-", Pos: i})
			pendingDirectiveArgs = false
			pendingMethodArgs = false
			i++
		case input[i] == '*':
			tokens = append(tokens, Token{Kind: TokenMul, Image: "*", Pos: i})
			pendingDirectiveArgs = false
			pendingMethodArgs = false
			i++
		case input[i] == '/':
			tokens = append(tokens, Token{Kind: TokenDiv, Image: "/", Pos: i})
			pendingDirectiveArgs = false
			pendingMethodArgs = false
			i++
		case input[i] == '%':
			tokens = append(tokens, Token{Kind: TokenMod, Image: "%", Pos: i})
			pendingDirectiveArgs = false
			pendingMethodArgs = false
			i++
		case input[i] == '=':
			tokens = append(tokens, Token{Kind: TokenEquals, Image: "=", Pos: i})
			pendingDirectiveArgs = false
			pendingMethodArgs = false
			i++
		case input[i] == ';':
			tokens = append(tokens, Token{Kind: TokenText, Image: ";", Pos: i})
			pendingDirectiveArgs = false
			pendingMethodArgs = false
			i++

		default:
			start := i
			for i < len(input) &&
				!isTokenBoundary(input[i]) &&
				input[i] != '\n' &&
				input[i] != '\r' {
				i++
			}
			// Guard against unmatched boundary chars (for example '\' or '&'):
			// if no lexer branch consumed them, we must still make progress.
			if i == start {
				if input[i] == '\r' && i+1 < len(input) && input[i+1] == '\n' {
					i += 2
				} else {
					i++
				}
			}
			tok := Token{Kind: TokenText, Image: input[start:i], Pos: start}
			tokens = append(tokens, tok)
			if !isWhitespaceOnlyText(tok.Image) {
				pendingDirectiveArgs = false
				pendingMethodArgs = false
			}
		}
	}

	tokens = append(tokens, Token{Kind: TokenEOF, Image: "", Pos: len(input)})
	return tokens, nil
}

func scanSetDirective(s string, i int) (string, int, bool) {
	if hasPrefixAt(s, i, "#set") {
		j := i + len("#set")
		for j < len(s) && isSetWhitespace(s[j]) {
			j++
		}
		if j < len(s) && s[j] == '(' {
			j++
			return s[i:j], j, true
		}
	}

	if hasPrefixAt(s, i, "#{set}") {
		j := i + len("#{set}")
		for j < len(s) && isSetWhitespace(s[j]) {
			j++
		}
		if j < len(s) && s[j] == '(' {
			j++
			return s[i:j], j, true
		}
	}

	return "", i, false
}

func scanDirective(s string, i int) (string, int, bool) {
	if i >= len(s) || s[i] != '#' {
		return "", i, false
	}

	if i+1 < len(s) && s[i+1] == '{' {
		j := i + 2
		if j >= len(s) || !isIdentifierStart(rune(s[j])) {
			return "", i, false
		}
		j++
		for j < len(s) && isIdentifierPart(rune(s[j])) {
			j++
		}
		if j >= len(s) || s[j] != '}' {
			return "", i, false
		}
		j++
		return s[i:j], j, true
	}

	j := i + 1
	if j < len(s) && s[j] == '@' {
		j++
	}
	if j >= len(s) || !isIdentifierStart(rune(s[j])) {
		return "", i, false
	}
	j++
	for j < len(s) && isIdentifierPart(rune(s[j])) {
		j++
	}
	return s[i:j], j, true
}

func isKnownDirective(dir string) bool {
	switch canonicalDirectiveName(dir) {
	case "set", "if", "elseif", "else", "end", "foreach", "macro", "define", "parse", "include", "evaluate", "break", "stop":
		return true
	default:
		return false
	}
}

func canonicalDirectiveName(dir string) string {
	if len(dir) >= 3 && dir[0] == '#' && dir[1] == '{' && dir[len(dir)-1] == '}' {
		return dir[2 : len(dir)-1]
	}
	if len(dir) >= 1 && dir[0] == '#' {
		return dir[1:]
	}
	return dir
}

func scanReference(s string, i int) (string, int, bool) {
	if i >= len(s) || s[i] != '$' {
		return "", i, false
	}
	start := i
	i++

	if i < len(s) && s[i] == '!' {
		i++
		if i < len(s) && s[i] == '{' {
			return "$!{", i + 1, true
		}
	}

	if i < len(s) && s[i] == '{' {
		return "${", i + 1, true
	}

	// Java parser keeps repeated '$' prefix as part of a reference image.
	for i < len(s) && s[i] == '$' {
		i++
	}
	if i >= len(s) || !isIdentifierStart(rune(s[i])) {
		return "", start, false
	}

	j := i
	for j < len(s) && isIdentifierPart(rune(s[j])) {
		j++
	}
	return s[start:j], j, true
}

func isEscapableStart(s string, i int) bool {
	if i >= len(s) {
		return false
	}
	return s[i] == '#' || s[i] == '$'
}

func backslashRunBeforeDollar(s string, i int) bool {
	if i >= len(s) || s[i] != '\\' {
		return false
	}
	j := i
	for j < len(s) && s[j] == '\\' {
		j++
	}
	return j < len(s) && s[j] == '$'
}

func isSetWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

func hasPrefixAt(s string, i int, prefix string) bool {
	if i+len(prefix) > len(s) {
		return false
	}
	return s[i:i+len(prefix)] == prefix
}

func skipWhitespace(s string, i int) int {
	for i < len(s) {
		if s[i] != ' ' && s[i] != '\t' && s[i] != '\n' && s[i] != '\r' {
			return i
		}
		i++
	}
	return i
}

func isWhitespaceOnlyText(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] != ' ' && s[i] != '\t' && s[i] != '\n' && s[i] != '\r' {
			return false
		}
	}
	return true
}

func isEscapedByBackslash(s string, i int) bool {
	if i <= 0 {
		return false
	}
	count := 0
	for j := i - 1; j >= 0 && s[j] == '\\'; j-- {
		count++
	}
	return count%2 == 1
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isIdentifierStart(r rune) bool {
	return unicode.IsLetter(r) || r == '_'
}

func isIdentifierPart(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

func isTokenBoundary(ch byte) bool {
	switch ch {
	case '#', '$', '"', '\'', '.', '(', ')', '[', ']', '{', '}', ',', ':', '=', ';', '!', '>', '<', '+', '-', '*', '/', '%', '&', '|', '\\':
		return true
	default:
		return isIdentifierStart(rune(ch)) || isDigit(rune(ch))
	}
}

func isSignedNumberStart(s string, i int) bool {
	if s[i] != '-' || i+1 >= len(s) {
		return false
	}
	if !isDigit(rune(s[i+1])) && !(s[i+1] == '.' && i+2 < len(s) && isDigit(rune(s[i+2]))) {
		return false
	}
	if i == 0 {
		return true
	}
	prev := s[i-1]
	return prev == ' ' || prev == '\t' || prev == '\n' || prev == '\r' ||
		prev == '(' || prev == '[' || prev == '{' || prev == ',' || prev == ':' ||
		prev == '=' || prev == '+' || prev == '-' || prev == '*' || prev == '/' || prev == '%' ||
		prev == '!' || prev == '<' || prev == '>' || prev == '|' || prev == '&' || prev == '.'
}

func isUnsignedNumberStart(s string, i int) bool {
	if isDigit(rune(s[i])) {
		return true
	}
	return s[i] == '.' && i+1 < len(s) && isDigit(rune(s[i+1]))
}

func scanNumberEnd(s string, start int) int {
	i := start
	if s[i] == '-' {
		i++
	}

	for i < len(s) && isDigit(rune(s[i])) {
		i++
	}

	// Keep integer token when this is a range start like 1..3.
	if i+1 < len(s) && s[i] == '.' && s[i+1] == '.' {
		return i
	}

	if i < len(s) && s[i] == '.' {
		i++
		for i < len(s) && isDigit(rune(s[i])) {
			i++
		}
	}

	if i < len(s) && (s[i] == 'e' || s[i] == 'E') {
		j := i + 1
		if j < len(s) && (s[j] == '+' || s[j] == '-') {
			j++
		}
		if j < len(s) && isDigit(rune(s[j])) {
			i = j + 1
			for i < len(s) && isDigit(rune(s[i])) {
				i++
			}
		}
	}

	return i
}
