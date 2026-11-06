package ngapak

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// Lexer breaks NgapakIn source code into tokens.
type Lexer struct {
	source   []rune
	filename string
	pos      int
	line     int
	column   int
	char     rune
}

// NewLexer creates a new Lexer instance.
func NewLexer(source string, filename string) *Lexer {
	if filename == "" {
		filename = "<stdin>"
	}
	r := []rune(source)
	l := &Lexer{
		source:   r,
		filename: filename,
		pos:      0,
		line:     1,
		column:   1,
	}
	if len(r) > 0 {
		l.char = r[0]
	}
	return l
}

func (l *Lexer) advance() {
	if l.char == '\n' {
		l.line++
		l.column = 1
	} else {
		l.column++
	}
	l.pos++
	if l.pos < len(l.source) {
		l.char = l.source[l.pos]
	} else {
		l.char = 0
	}
}

func (l *Lexer) peek() rune {
	peekPos := l.pos + 1
	if peekPos < len(l.source) {
		return l.source[peekPos]
	}
	return 0
}

// Tokenize processes the entire source and returns a slice of Token.
func (l *Lexer) Tokenize() ([]Token, error) {
	var tokens []Token

	for l.char != 0 {
		// Skip spaces and carriage returns
		if l.char == ' ' || l.char == '\t' || l.char == '\r' {
			l.advance()
			continue
		}

		// Newlines: collapse multiple newlines
		if l.char == '\n' {
			if len(tokens) > 0 && tokens[len(tokens)-1].Type != T_NEWLINE {
				tokens = append(tokens, Token{Type: T_NEWLINE, Line: l.line, Column: l.column})
			}
			l.advance()
			continue
		}

		// Comments: skip until newline or EOF
		if l.char == '#' {
			for l.char != 0 && l.char != '\n' {
				l.advance()
			}
			continue
		}

		// Strings
		if l.char == '"' || l.char == '\'' {
			tok, err := l.makeString(l.char)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, tok)
			continue
		}

		// Numbers
		if unicode.IsDigit(l.char) || (l.char == '.' && unicode.IsDigit(l.peek())) {
			tokens = append(tokens, l.makeNumber())
			continue
		}

		// Identifiers and Keywords
		if unicode.IsLetter(l.char) || l.char == '_' {
			tokens = append(tokens, l.makeIdentifier())
			continue
		}

		// Operators & Delimiters
		startCol := l.column
		startLine := l.line

		switch l.char {
		case '=':
			l.advance()
			if l.char == '=' {
				tokens = append(tokens, Token{Type: T_EQ, Line: startLine, Column: startCol})
				l.advance()
			} else {
				tokens = append(tokens, Token{Type: T_ASSIGN, Line: startLine, Column: startCol})
			}
		case '!':
			l.advance()
			if l.char == '=' {
				tokens = append(tokens, Token{Type: T_NEQ, Line: startLine, Column: startCol})
				l.advance()
			} else {
				return nil, fmt.Errorf("%s:%d:%d: karakter '!' kudu diterusna '=' nggo operator ora padha (!=)", l.filename, startLine, startCol)
			}
		case '<':
			l.advance()
			if l.char == '=' {
				tokens = append(tokens, Token{Type: T_LTE, Line: startLine, Column: startCol})
				l.advance()
			} else {
				tokens = append(tokens, Token{Type: T_LT, Line: startLine, Column: startCol})
			}
		case '>':
			l.advance()
			if l.char == '=' {
				tokens = append(tokens, Token{Type: T_GTE, Line: startLine, Column: startCol})
				l.advance()
			} else {
				tokens = append(tokens, Token{Type: T_GT, Line: startLine, Column: startCol})
			}
		case '+':
			tokens = append(tokens, Token{Type: T_PLUS, Line: startLine, Column: startCol})
			l.advance()
		case '-':
			tokens = append(tokens, Token{Type: T_MINUS, Line: startLine, Column: startCol})
			l.advance()
		case '*':
			tokens = append(tokens, Token{Type: T_MUL, Line: startLine, Column: startCol})
			l.advance()
		case '/':
			tokens = append(tokens, Token{Type: T_DIV, Line: startLine, Column: startCol})
			l.advance()
		case '(':
			tokens = append(tokens, Token{Type: T_LPAREN, Line: startLine, Column: startCol})
			l.advance()
		case ')':
			tokens = append(tokens, Token{Type: T_RPAREN, Line: startLine, Column: startCol})
			l.advance()
		case ',':
			tokens = append(tokens, Token{Type: T_COMMA, Line: startLine, Column: startCol})
			l.advance()
		case '.':
			tokens = append(tokens, Token{Type: T_DOT, Line: startLine, Column: startCol})
			l.advance()
		case '{':
			tokens = append(tokens, Token{Type: T_LBRACE, Line: startLine, Column: startCol})
			l.advance()
		case '}':
			tokens = append(tokens, Token{Type: T_RBRACE, Line: startLine, Column: startCol})
			l.advance()
		case '[':
			tokens = append(tokens, Token{Type: T_LBRACKET, Line: startLine, Column: startCol})
			l.advance()
		case ']':
			tokens = append(tokens, Token{Type: T_RBRACKET, Line: startLine, Column: startCol})
			l.advance()
		case ':':
			tokens = append(tokens, Token{Type: T_COLON, Line: startLine, Column: startCol})
			l.advance()
		default:
			return nil, fmt.Errorf("%s:%d:%d: karakter ora dingerteni: %q", l.filename, startLine, startCol, l.char)
		}
	}

	// Remove trailing newline if present
	if len(tokens) > 0 && tokens[len(tokens)-1].Type == T_NEWLINE {
		tokens = tokens[:len(tokens)-1]
	}

	tokens = append(tokens, Token{Type: T_EOF, Line: l.line, Column: l.column})
	return tokens, nil
}

func (l *Lexer) makeString(quote rune) (Token, error) {
	startLine := l.line
	startCol := l.column
	l.advance() // skip quote

	var b strings.Builder
	for l.char != 0 && l.char != quote {
		if l.char == '\\' {
			l.advance()
			switch l.char {
			case 'n':
				b.WriteRune('\n')
			case 't':
				b.WriteRune('\t')
			case 'r':
				b.WriteRune('\r')
			case '\\':
				b.WriteRune('\\')
			case quote:
				b.WriteRune(quote)
			case 0:
				return Token{}, fmt.Errorf("%s:%d:%d: teks string ora ditutup", l.filename, startLine, startCol)
			default:
				b.WriteRune('\\')
				b.WriteRune(l.char)
			}
		} else {
			b.WriteRune(l.char)
		}
		l.advance()
	}

	if l.char == 0 {
		return Token{}, fmt.Errorf("%s:%d:%d: teks string ora ditutup", l.filename, startLine, startCol)
	}

	l.advance() // skip closing quote
	return Token{Type: T_STRING, Value: b.String(), Line: startLine, Column: startCol}, nil
}

func (l *Lexer) makeNumber() Token {
	startCol := l.column
	var b strings.Builder
	dotCount := 0

	for l.char != 0 && (unicode.IsDigit(l.char) || l.char == '.') {
		if l.char == '.' {
			if dotCount == 1 {
				break
			}
			dotCount++
		}
		b.WriteRune(l.char)
		l.advance()
	}

	numStr := b.String()
	if dotCount == 0 {
		val, _ := strconv.ParseInt(numStr, 10, 64)
		return Token{Type: T_NUMBER, Value: val, Line: l.line, Column: startCol}
	}
	val, _ := strconv.ParseFloat(numStr, 64)
	return Token{Type: T_NUMBER, Value: val, Line: l.line, Column: startCol}
}

func (l *Lexer) makeIdentifier() Token {
	startCol := l.column
	var b strings.Builder

	for l.char != 0 && (unicode.IsLetter(l.char) || unicode.IsDigit(l.char) || l.char == '_') {
		b.WriteRune(l.char)
		l.advance()
	}

	id := b.String()
	tokType, isKeyword := Keywords[id]
	if !isKeyword {
		tokType = T_IDENTIFIER
	}
	return Token{Type: tokType, Value: id, Line: l.line, Column: startCol}
}

