package eval

import (
	"strconv"
	"strings"
)

// numberSuffixes maps an attached, case-insensitive literal suffix to its
// magnitude multiplier. See plans/04-number-shorthand-suffixes.md.
// Matching is whole-run: "5tr" folds, but "5train" and "5tr2" do not.
var numberSuffixes = map[string]float64{
	"k":  1e3, // thousand
	"m":  1e6, // million
	"tr": 1e6, // triệu (Vietnamese: million)
	"b":  1e9, // billion
	"t":  1e9, // tỷ (Vietnamese: billion)
}

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

			// ParseFloat returns ±Inf with ErrRange for absurdly large literals;
			// the evaluator rejects non-finite values, so the error is ignored here.
			value, _ := strconv.ParseFloat(string(runes[i:end]), 64)

			// Number shorthand: an attached suffix like 100k / 10m / 10tr / 1b / 1t
			// folds into this same token (case-insensitive, whole-run match).
			literEnd := scanNumberSuffix(runes, end)
			if literEnd > end {
				if mult, ok := suffixMultiplier(string(runes[end:literEnd])); ok {
					value *= mult
					end = literEnd
				}
			}

			toks = append(toks, Token{Kind: TokenNumber, Text: string(runes[i:end]), Value: value, Start: i, End: end})
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
	// Collect until finding no number character
	for i < len(runes) && isDigit(runes[i]) {
		i++
	}

	// If after the last number character is "." -> decimal -> collect the rest of decimal part as well
	if i < len(runes) && runes[i] == '.' {
		i++
		for i < len(runes) && isDigit(runes[i]) {
			i++
		}
	}
	return i
}

// scanNumberSuffix returns the exclusive end of the maximal ident-like run
// starting at runes[start] (immediately after a numeric literal), or start
// when the next rune cannot begin a suffix (digit stop, operator, EOF, ...).
// The caller must verify the run exactly equals a known suffix, but this
// function's "maximal run" property is what rejects e.g. "5train" or "5tr2".
func scanNumberSuffix(runes []rune, start int) int {
	i := start
	if i < len(runes) && isIdentStart(runes[i]) {
		i++
	}
	for i < len(runes) && isIdentPart(runes[i]) {
		i++
	}
	return i
}

// suffixMultiplier resolves a lowercased ident-style run to its magnitude
// multiplier, using the maximal-run rule: only entire-run matches fold.
func suffixMultiplier(run string) (float64, bool) {
	mult, ok := numberSuffixes[strings.ToLower(run)]
	return mult, ok
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
