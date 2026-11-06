package ngapak

import "fmt"

// TokenType represents the type of a lexical token in NgapakIn.
type TokenType string

const (
	// Keywords
	T_TULIS   TokenType = "TULIS"
	T_NEK     TokenType = "NEK"
	T_YA      TokenType = "YA"
	T_LIYANE  TokenType = "LIYANE"
	T_RAMPUNG TokenType = "RAMPUNG"
	T_BALENI  TokenType = "BALENI"
	T_SAKA    TokenType = "SAKA"
	T_NGANTI  TokenType = "NGANTI"
	T_GAWE    TokenType = "GAWE"
	T_BALEKNA TokenType = "BALEKNA"
	T_BENER   TokenType = "BENER"
	T_SALAH   TokenType = "SALAH"
	T_KOSONG  TokenType = "KOSONG"
	T_KELAS   TokenType = "KELAS"
	T_EXTENDS TokenType = "EXTENDS"
	T_LAN     TokenType = "LAN"
	T_UTAWA   TokenType = "UTAWA"
	T_ORA     TokenType = "ORA"
	T_NENG    TokenType = "NENG"
	T_MANDEG  TokenType = "MANDEG"
	T_TERUSNA TokenType = "TERUSNA"

	// Literals & Identifiers
	T_IDENTIFIER TokenType = "IDENTIFIER"
	T_NUMBER     TokenType = "NUMBER"
	T_STRING     TokenType = "STRING"

	// Operators
	T_PLUS   TokenType = "PLUS"
	T_MINUS  TokenType = "MINUS"
	T_MUL    TokenType = "MUL"
	T_DIV    TokenType = "DIV"
	T_ASSIGN TokenType = "ASSIGN"
	T_EQ     TokenType = "EQ"
	T_NEQ    TokenType = "NEQ"
	T_LT     TokenType = "LT"
	T_GT     TokenType = "GT"
	T_LTE    TokenType = "LTE"
	T_GTE    TokenType = "GTE"

	// Punctuation & Delimiters
	T_LPAREN   TokenType = "LPAREN"
	T_RPAREN   TokenType = "RPAREN"
	T_LBRACE   TokenType = "LBRACE"
	T_RBRACE   TokenType = "RBRACE"
	T_COMMA    TokenType = "COMMA"
	T_NEWLINE  TokenType = "NEWLINE"
	T_DOT      TokenType = "DOT"
	T_LBRACKET TokenType = "LBRACKET"
	T_RBRACKET TokenType = "RBRACKET"
	T_COLON    TokenType = "COLON"
	T_EOF      TokenType = "EOF"
)

// Keywords map maps keyword identifiers to their TokenType.
var Keywords = map[string]TokenType{
	"tulis":   T_TULIS,
	"nek":     T_NEK,
	"ya":      T_YA,
	"liyane":  T_LIYANE,
	"rampung": T_RAMPUNG,
	"baleni":  T_BALENI,
	"saka":    T_SAKA,
	"nganti":  T_NGANTI,
	"gawe":    T_GAWE,
	"balekna": T_BALEKNA,
	"bener":   T_BENER,
	"salah":   T_SALAH,
	"kosong":  T_KOSONG,
	"kelas":   T_KELAS,
	"extends": T_EXTENDS,
	"lan":     T_LAN,
	"utawa":   T_UTAWA,
	"ora":     T_ORA,
	"neng":    T_NENG,
	"mandeg":  T_MANDEG,
	"terusna": T_TERUSNA,
}

// Token holds information for a single scanned token.
type Token struct {
	Type   TokenType
	Value  any
	Line   int
	Column int
}

func (t Token) String() string {
	if t.Value != nil {
		return fmt.Sprintf("Token(%s, %v, %d:%d)", t.Type, t.Value, t.Line, t.Column)
	}
	return fmt.Sprintf("Token(%s, %d:%d)", t.Type, t.Line, t.Column)
}

