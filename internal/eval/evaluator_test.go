package eval

import (
	"errors"
	"math"
	"strings"
	"testing"
)

func TestEvaluateLineValues(t *testing.T) {
	tests := []struct {
		name string
		line string
		want string // expected FormattedValue; HasResult is implied true
	}{
		{"addition and subtraction", "20 + 5 - 12", "13"},
		{"multiplication", "10 * 5", "50"},
		{"division", "20 / 4", "5"},
		{"modulo", "17 % 5", "2"},
		{"precedence mul over add", "2 + 3 * 4", "14"},
		{"parentheses override precedence", "(2 + 3) * 4", "20"},
		{"unary minus on group", "-(10 + 5) * -2", "30"},
		{"unary plus and minus", "+5 + -3", "2"},
		{"double unary", "--5", "5"},
		{"unary after binary operator", "2 * -3", "-6"},
		{"left associativity of division", "100 / 10 / 2", "5"},
		{"left associativity of subtraction", "10 - 3 - 2", "5"},
		{"decimal", "3.14 * 2", "6.28"},
		{"leading dot decimal", ".5 * 4", "2"},
		{"trailing dot number", "5. + 1", "6"},
		{"float noise removed", "0.1 + 0.2", "0.3"},
		{"repeating decimal rounded", "10 / 3", "3.3333333333"},
		{"negative result", "3 - 10", "-7"},
		{"negative zero normalised", "-0", "0"},
		{"negative tiny value rounds to zero", "-0.0000000001 * 0.001", "0"},
		{"large integer keeps digits", "1000000000000 * 10", "10000000000000"},
		{"whitespace and tabs", "  20\t+ 5  ", "25"},
		{"trailing comment", "20 + 5 # lunch", "25"},
		{"trailing double-slash comment", "20 + 5 // lunch", "25"},
		{"CRLF line ending", "20 + 5\r", "25"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := EvaluateLine(0, tt.line)
			if !res.HasResult {
				t.Fatalf("EvaluateLine(%q): HasResult = false, Err = %v", tt.line, res.Err)
			}
			if res.Err != nil {
				t.Fatalf("EvaluateLine(%q): unexpected Err %v", tt.line, res.Err)
			}
			if res.FormattedValue != tt.want {
				t.Fatalf("EvaluateLine(%q) = %q, want %q", tt.line, res.FormattedValue, tt.want)
			}
		})
	}
}

func TestEvaluateLineNoResult(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		wantErr error // nil means no error is expected (blank or comment-only)
	}{
		{"empty line", "", nil},
		{"whitespace only", "   \t ", nil},
		{"comment only hash", "# Grocery budget", nil},
		{"comment only slashes", "// items", nil},
		{"trailing operator mid-typing", "20 + ", ErrIncomplete},
		{"trailing minus mid-typing", "20 + 5 - ", ErrIncomplete},
		{"unary without operand", "-", ErrIncomplete},
		{"unclosed paren", "(10 + 5", ErrIncomplete},
		{"plain text header", "Lunch expenses:", ErrUnexpectedToken},
		{"plain text note with number", "Note: buy milk", ErrUnexpectedToken},
		{"two adjacent numbers", "2 apples", ErrUnexpectedToken},
		{"stray close paren", "10 + 5)", ErrUnexpectedToken},
		{"empty parens", "()", ErrUnexpectedToken},
		{"assignment not yet supported", "x = 10 + 5", ErrUnexpectedToken},
		{"unknown symbol", "$20 + 5", ErrUnexpectedToken},
		{"divide by zero", "10 / 0", ErrDivideByZero},
		{"modulo by zero", "10 % 0", ErrDivideByZero},
		{"divide by zero inside expression", "1 + 10 / (5 - 5)", ErrDivideByZero},
		{"literal overflows to infinity", "1" + strings.Repeat("0", 400), ErrNonFinite},
		{"multiplication overflows", "1" + strings.Repeat("0", 200) + " * 1" + strings.Repeat("0", 200), ErrNonFinite},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := EvaluateLine(0, tt.line)
			if res.HasResult {
				t.Fatalf("EvaluateLine(%q): HasResult = true, want false (value %q)", tt.line, res.FormattedValue)
			}
			if res.FormattedValue != "" {
				t.Fatalf("EvaluateLine(%q): FormattedValue = %q, want empty", tt.line, res.FormattedValue)
			}
			if tt.wantErr == nil {
				if res.Err != nil {
					t.Fatalf("EvaluateLine(%q): Err = %v, want nil", tt.line, res.Err)
				}
				return
			}
			if !errors.Is(res.Err, tt.wantErr) {
				t.Fatalf("EvaluateLine(%q): Err = %v, want %v", tt.line, res.Err, tt.wantErr)
			}
		})
	}
}

func TestEvaluateLineKeepsMetadata(t *testing.T) {
	res := EvaluateLine(3, "20 + 5 # note")

	if res.LineIndex != 3 {
		t.Errorf("LineIndex = %d, want 3", res.LineIndex)
	}
	if res.RawText != "20 + 5 # note" {
		t.Errorf("RawText = %q", res.RawText)
	}
	// Tokens include the comment so highlighting can colour it later.
	last := res.Tokens[len(res.Tokens)-1]
	if last.Kind != TokenComment {
		t.Errorf("last token kind = %v, want Comment", last.Kind)
	}
}

func TestEvaluateDocument(t *testing.T) {
	text := "20 + 5 - 12\n10 * 4\n# comment\n(8 + 2) / 2\n\nLunch expenses:\n20 + "
	results := EvaluateDocument(text)

	if got, want := len(results), 7; got != want {
		t.Fatalf("len(results) = %d, want %d", got, want)
	}

	wantValues := []string{"13", "40", "", "5", "", "", ""}
	wantHas := []bool{true, true, false, true, false, false, false}
	for i := range results {
		if results[i].LineIndex != i {
			t.Errorf("results[%d].LineIndex = %d", i, results[i].LineIndex)
		}
		if results[i].HasResult != wantHas[i] {
			t.Errorf("results[%d].HasResult = %v, want %v", i, results[i].HasResult, wantHas[i])
		}
		if results[i].FormattedValue != wantValues[i] {
			t.Errorf("results[%d].FormattedValue = %q, want %q", i, results[i].FormattedValue, wantValues[i])
		}
	}

	if sum, n := SumResults(results); sum != 58 || n != 3 {
		t.Errorf("SumResults = (%v, %d), want (58, 3)", sum, n)
	}
}

func TestEvaluateDocumentLineCount(t *testing.T) {
	tests := []struct {
		text string
		want int
	}{
		{"", 1},
		{"1", 1},
		{"1\n", 2},
		{"1\n2\n3", 3},
	}
	for _, tt := range tests {
		if got := len(EvaluateDocument(tt.text)); got != tt.want {
			t.Errorf("len(EvaluateDocument(%q)) = %d, want %d", tt.text, got, tt.want)
		}
	}
}

func TestEvaluateDocumentLinesAreIndependent(t *testing.T) {
	// A broken line must not affect the results of the lines around it.
	results := EvaluateDocument("10 / 0\n5 + 5\n20 +")
	if results[0].HasResult || results[2].HasResult {
		t.Fatalf("broken lines should have no result: %+v", results)
	}
	if !results[1].HasResult || results[1].FormattedValue != "10" {
		t.Fatalf("line 1 = %+v, want value 10", results[1])
	}
}

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		in   float64
		want string
	}{
		{13, "13"},
		{-7, "-7"},
		{0, "0"},
		{math.Copysign(0, -1), "0"},
		{2.5, "2.5"},
		{0.1 + 0.2, "0.3"},
		{10.0 / 3.0, "3.3333333333"},
		{2.0 / 3.0, "0.6666666667"},
		{1e15, "1000000000000000"},
		{-0.00000000001, "0"},
		{1234567056.6333334446, "1234567056.63"},
		{123456789012345, "123456789012345"},
		{0.1234567890123, "0.123456789"},
		{-42.5, "-42.5"},
	}
	for _, tt := range tests {
		if got := FormatNumber(tt.in); got != tt.want {
			t.Errorf("FormatNumber(%v) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestSumResultsIgnoresNoResultLines(t *testing.T) {
	results := []LineResult{
		{HasResult: true, Value: 1.5},
		{HasResult: false, Value: 999},
		{HasResult: true, Value: 2.5},
	}
	sum, n := SumResults(results)
	if sum != 4 || n != 2 {
		t.Fatalf("SumResults = (%v, %d), want (4, 2)", sum, n)
	}
}
