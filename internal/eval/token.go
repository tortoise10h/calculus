package eval

// TokenKind identifies the lexical class of a Token.
type TokenKind int

const (
	TokenEOF     TokenKind = iota // Reserved for parser end-of-input; never emitted by LexLine.
	TokenNumber                   // 20, 3.14, .5, and shorthand: 100k==100000, 10m/10tr==10000000, 1b/1t==1000000000
	TokenIdent                    // drink_cost, x1 (variables arrive in Phase 3)
	TokenPlus                     // +
	TokenMinus                    // -
	TokenStar                     // *
	TokenSlash                    // /
	TokenPercent                  // %
	TokenLParen                   // (
	TokenRParen                   // )
	TokenAssign                   // = (assignments arrive in Phase 3)
	TokenComment                  // # ... or // ... to end of line
	TokenUnknown                  // any other rune
)

var tokenKindNames = [...]string{
	TokenEOF:     "EOF",
	TokenNumber:  "Number",
	TokenIdent:   "Ident",
	TokenPlus:    "+",
	TokenMinus:   "-",
	TokenStar:    "*",
	TokenSlash:   "/",
	TokenPercent: "%",
	TokenLParen:  "(",
	TokenRParen:  ")",
	TokenAssign:  "=",
	TokenComment: "Comment",
	TokenUnknown: "Unknown",
}

// String returns a human-readable name for the kind, mainly for tests and errors.
func (k TokenKind) String() string {
	if int(k) < 0 || int(k) >= len(tokenKindNames) {
		return "Invalid"
	}
	return tokenKindNames[k]
}

// Token is one lexical unit of a single line.
//
// Start and End are rune offsets relative to the start of the line (End is
// exclusive), so they match the rune indexing used by ui.TextRange.
type Token struct {
	Kind  TokenKind
	Text  string
	Value float64 // Set when Kind == TokenNumber (already multiplied by any shorthand suffix).
	Start int
	End   int
}
