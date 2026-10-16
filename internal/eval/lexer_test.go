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

func TestLexLineNumberShorthand(t *testing.T) {
	tests := []struct {
		input string
		want  []Token
	}{
		{"100k", []Token{{Kind: TokenNumber, Text: "100k", Value: 100000, Start: 0, End: 4}}},
		{"10m", []Token{{Kind: TokenNumber, Text: "10m", Value: 10000000, Start: 0, End: 3}}},
		{"10tr", []Token{{Kind: TokenNumber, Text: "10tr", Value: 10000000, Start: 0, End: 4}}},
		{"1b", []Token{{Kind: TokenNumber, Text: "1b", Value: 1000000000, Start: 0, End: 2}}},
		{"1t", []Token{{Kind: TokenNumber, Text: "1t", Value: 1000000000, Start: 0, End: 2}}},
		{"5TR", []Token{{Kind: TokenNumber, Text: "5TR", Value: 5000000, Start: 0, End: 3}}},
		{"2M", []Token{{Kind: TokenNumber, Text: "2M", Value: 2000000, Start: 0, End: 2}}},
		{"-2k", []Token{
			{Kind: TokenMinus, Text: "-", Start: 0, End: 1},
			{Kind: TokenNumber, Text: "2k", Value: 2000, Start: 1, End: 3},
		}},
		{"1.5k", []Token{{Kind: TokenNumber, Text: "1.5k", Value: 1500, Start: 0, End: 4}}},
		{".5m", []Token{{Kind: TokenNumber, Text: ".5m", Value: 500000, Start: 0, End: 3}}},
		{"1.k", []Token{{Kind: TokenNumber, Text: "1.k", Value: 1000, Start: 0, End: 3}}},
		{"5train", []Token{
			{Kind: TokenNumber, Text: "5", Value: 5, Start: 0, End: 1},
			{Kind: TokenIdent, Text: "train", Start: 1, End: 6},
		}},
		{"5tr2", []Token{
			{Kind: TokenNumber, Text: "5", Value: 5, Start: 0, End: 1},
			{Kind: TokenIdent, Text: "tr2", Start: 1, End: 4},
		}},
		{"5 k", []Token{
			{Kind: TokenNumber, Text: "5", Value: 5, Start: 0, End: 1},
			{Kind: TokenIdent, Text: "k", Start: 2, End: 3},
		}},
		{"m", []Token{{Kind: TokenIdent, Text: "m", Start: 0, End: 1}}},
		{"k = 5", []Token{
			{Kind: TokenIdent, Text: "k", Start: 0, End: 1},
			{Kind: TokenAssign, Text: "=", Start: 2, End: 3},
			{Kind: TokenNumber, Text: "5", Value: 5, Start: 4, End: 5},
		}},
		{"100k + 5", []Token{
			{Kind: TokenNumber, Text: "100k", Value: 100000, Start: 0, End: 4},
			{Kind: TokenPlus, Text: "+", Start: 5, End: 6},
			{Kind: TokenNumber, Text: "5", Value: 5, Start: 7, End: 8},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := LexLine(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("LexLine(%q) = %+v, want %+v", tt.input, got, tt.want)
			}
			for i := range got {
				w := tt.want[i]
				g := got[i]
				if g.Kind != w.Kind || g.Text != w.Text || g.Start != w.Start || g.End != w.End {
					t.Fatalf("LexLine(%q)[%d] = %+v, want %+v", tt.input, i, g, w)
				}
				if w.Kind == TokenNumber && g.Value != w.Value {
					t.Fatalf("LexLine(%q)[%d].Value = %v, want %v", tt.input, i, g.Value, w.Value)
				}
			}
		})
	}

	t.Run("3.14k float rounding", func(t *testing.T) {
		got := LexLine("3.14k")
		if len(got) != 1 || got[0].Kind != TokenNumber || got[0].Text != "3.14k" || got[0].Start != 0 || got[0].End != 5 {
			t.Fatalf("LexLine(\"3.14k\") = %+v", got)
		}
		if diff := got[0].Value - 3140; diff < -1e-6 || diff > 1e-6 {
			t.Fatalf("Value = %v, want ~3140", got[0].Value)
		}
	})
}
