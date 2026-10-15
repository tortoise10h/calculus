package eval

import "strconv"

// LexLine splits a single line into tokens. It never fails: any rune it does
// not recognise becomes a TokenUnknown, so arbitrary user text is safe.
//
// Whitespace is skipped. A '#' or '//' starts a comment that runs to the end
// of the line and is emitted as a single TokenComment.
func LexLine(line string) []Token {
	// Use rune to ensure if text contain characte line chinese then each character will stay as 1 item per index
	runes := []rune(line)
	var toks []Token

	for i := 0; i < len(runes); {
		r := runes[i]
		switch {
		case r == ' ' || r == '\t' || r == '\r':
			i++

		// Comments parsing, if we face # -> mean the rest of this line is just comment
		case r == '#' || (r == '/' && i+1 < len(runes) && runes[i+1] == '/'):
			toks = append(toks, Token{
				Kind:  TokenComment,
				Text:  string(runes[i:]),
				Start: i,
				End:   len(runes),
			})
			return toks

		case isDigit(r) || (r == '.' && i+1 < len(runes) && isDigit(runes[i+1])):
			// This one to be sure if we face decimal number like 1.245678 we then can get the whole string from 1->8 => "1.2345678"
			end := scanNumber(runes, i)
			text := string(runes[i:end])

			// ParseFloat returns ±Inf with ErrRange for absurdly large literals;
			// the evaluator rejects non-finite values, so the error is ignored here.
			value, _ := strconv.ParseFloat(text, 64)
			toks = append(toks, Token{Kind: TokenNumber, Text: text, Value: value, Start: i, End: end})
			i = end

		// For checking both variable `total` `sub_total` or also a built-in functions like `sqrt`
		case isIdentStart(r):
			end := i + 1
			for end < len(runes) && isIdentPart(runes[end]) {
				end++
			}
			toks = append(toks, Token{Kind: TokenIdent, Text: string(runes[i:end]), Start: i, End: end})
			i = end

		default:
			toks = append(toks, Token{Kind: singleRuneKind(r), Text: string(r), Start: i, End: i + 1})
			i++
		}
	}

	return toks
}

// scanNumber returns the exclusive end index of the numeric literal starting
// at runes[start]. Accepted forms: "123", "123.", "123.45", ".45".
func scanNumber(runes []rune, start int) int {
	i := start
	for i < len(runes) && isDigit(runes[i]) {
		i++
	}
	if i < len(runes) && runes[i] == '.' {
		i++
		for i < len(runes) && isDigit(runes[i]) {
			i++
		}
	}
	return i
}

func singleRuneKind(r rune) TokenKind {
	switch r {
	case '+':
		return TokenPlus
	case '-':
		return TokenMinus
	case '*':
		return TokenStar
	case '/':
		return TokenSlash
	case '%':
		return TokenPercent
	case '(':
		return TokenLParen
	case ')':
		return TokenRParen
	case '=':
		return TokenAssign
	default:
		return TokenUnknown
	}
}

func isDigit(r rune) bool { return r >= '0' && r <= '9' }

func isIdentStart(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_'
}

func isIdentPart(r rune) bool { return isIdentStart(r) || isDigit(r) }
