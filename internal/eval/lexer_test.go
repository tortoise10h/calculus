package eval

import (
	"reflect"
	"testing"
)

func TestLexLineKinds(t *testing.T) {
	toks := LexLine("(20 + 5.5) * x_1 % 2 / 4 - 1 = .5")
	want := []TokenKind{
		TokenLParen, TokenNumber, TokenPlus, TokenNumber, TokenRParen,
		TokenStar, TokenIdent, TokenPercent, TokenNumber, TokenSlash,
		TokenNumber, TokenMinus, TokenNumber, TokenAssign, TokenNumber,
	}
	if got := kinds(toks); !reflect.DeepEqual(got, want) {
		t.Fatalf("kinds = %v, want %v", got, want)
	}
}

func TestLexLineNumberValues(t *testing.T) {
	toks := LexLine("3.14 .5 5.")
	wantValues := []float64{3.14, 0.5, 5}
	var got []float64
	for _, tok := range toks {
		if tok.Kind == TokenNumber {
			got = append(got, tok.Value)
		}
	}
	if !reflect.DeepEqual(got, wantValues) {
		t.Fatalf("number values = %v, want %v", got, wantValues)
	}
}

func TestLexLineSpans(t *testing.T) {
	// Spans are rune offsets (End exclusive), so they index a []rune of the line.
	line := "ab + 12"
	toks := LexLine(line)
	runes := []rune(line)

	for _, tok := range toks {
		if got := string(runes[tok.Start:tok.End]); got != tok.Text {
			t.Errorf("token %v span [%d:%d] = %q, want %q", tok.Kind, tok.Start, tok.End, got, tok.Text)
		}
	}
}

func TestLexLineMultibyteSpans(t *testing.T) {
	// "é" is one rune but two bytes; spans must count runes.
	line := "é + 1"
	toks := LexLine(line)
	runes := []rune(line)
	last := toks[len(toks)-1]
	if last.Kind != TokenNumber || last.Start != 4 || last.End != 5 {
		t.Fatalf("last token = %+v, want number at rune span [4:5]", last)
	}
	if got := string(runes[last.Start:last.End]); got != "1" {
		t.Fatalf("slice = %q, want %q", got, "1")
	}
}

func TestLexLineCommentConsumesRest(t *testing.T) {
	toks := LexLine("20 # + 5 * ( ")
	if len(toks) != 2 {
		t.Fatalf("got %d tokens, want 2: %+v", len(toks), toks)
	}
	if toks[1].Kind != TokenComment || toks[1].Text != "# + 5 * ( " {
		t.Fatalf("comment token = %+v", toks[1])
	}
}

func TestLexLineDoubleSlashCommentVersusDivision(t *testing.T) {
	if k := kinds(LexLine("10 / 2")); !reflect.DeepEqual(k, []TokenKind{TokenNumber, TokenSlash, TokenNumber}) {
		t.Fatalf("single slash lexed as %v", k)
	}
	if k := kinds(LexLine("10 // 2")); !reflect.DeepEqual(k, []TokenKind{TokenNumber, TokenComment}) {
		t.Fatalf("double slash lexed as %v", k)
	}
}

func TestLexLineUnknownRunesDoNotPanic(t *testing.T) {
	toks := LexLine("$ ? : 🙂 ^ .")
	for _, tok := range toks {
		if tok.Kind != TokenUnknown {
			t.Errorf("token %q kind = %v, want Unknown", tok.Text, tok.Kind)
		}
	}
	if len(toks) != 6 {
		t.Errorf("got %d unknown tokens, want 6", len(toks))
	}
}

func TestLexLineEmpty(t *testing.T) {
	if toks := LexLine(""); len(toks) != 0 {
		t.Fatalf("LexLine(\"\") = %+v, want none", toks)
	}
}

func kinds(toks []Token) []TokenKind {
	out := make([]TokenKind, 0, len(toks))
	for _, tok := range toks {
		out = append(out, tok.Kind)
	}
	return out
}
