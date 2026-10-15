package eval

import (
	"math"
	"strconv"
	"strings"
)

const (
	// maxDecimals caps the decimal places shown. It hides IEEE-754 noise
	// (0.1 + 0.2 -> 0.3) while keeping typical precision.
	maxDecimals = 10
	// maxSignificant caps the significant digits shown. A float64 holds about
	// 15-17 reliable digits, so without this cap large values such as sums
	// show noise like 1234567056.6333334446.
	maxSignificant = 12
)

// LineResult is the evaluated outcome of one line of the document.
type LineResult struct {
	LineIndex      int     // Zero-based index of the line in the document.
	RawText        string  // The line as typed, without the trailing '\n'.
	Value          float64 // Numeric value, valid only when HasResult is true.
	FormattedValue string  // Display string for Value, empty when HasResult is false.
	HasResult      bool    // True when the line is a complete, valid expression.
	Err            error   // Why the line has no result. Nil for blank and comment-only lines.
	Tokens         []Token // All tokens of the line, including comments, for highlighting.
}

// EvaluateDocument evaluates every line of text independently and returns one
// LineResult per line. Lines are split on '\n', so the slice length is always
// strings.Count(text, "\n") + 1 (an empty text yields a single empty line).
func EvaluateDocument(text string) []LineResult {
	lines := strings.Split(text, "\n")
	results := make([]LineResult, len(lines))
	for i, line := range lines {
		results[i] = EvaluateLine(i, line)
	}
	return results
}

// EvaluateLine lexes, parses and evaluates one line. Blank lines, comment-only
// lines and lines that do not produce a value return HasResult == false; no
// line ever causes a panic.
func EvaluateLine(index int, line string) LineResult {
	res := LineResult{
		LineIndex: index,
		RawText:   line,
		Tokens:    LexLine(line),
	}

	// Strip comments before parsing; they are kept in Tokens for highlighting.
	code := make([]Token, 0, len(res.Tokens))
	for _, tok := range res.Tokens {
		if tok.Kind != TokenComment {
			code = append(code, tok)
		}
	}
	if len(code) == 0 {
		return res // Blank or comment-only line: no result, no error.
	}

	node, err := Parse(code)
	if err != nil {
		res.Err = err
		return res
	}

	value, err := Eval(node)
	if err != nil {
		res.Err = err
		return res
	}

	res.Value = value
	res.FormattedValue = FormatNumber(value)
	res.HasResult = true
	return res
}

// Eval computes the value of an AST. It returns ErrDivideByZero for '/' or '%'
// by zero, and ErrNonFinite if any intermediate or final value is NaN or ±Inf.
func Eval(n Node) (float64, error) {
	switch node := n.(type) {
	case *NumberNode:
		return checkFinite(node.Value)

	case *UnaryNode:
		x, err := Eval(node.X)
		if err != nil {
			return 0, err
		}
		if node.Op == TokenMinus {
			return -x, nil
		}
		return x, nil

	case *BinaryNode:
		left, err := Eval(node.Left)
		if err != nil {
			return 0, err
		}
		right, err := Eval(node.Right)
		if err != nil {
			return 0, err
		}

		var v float64
		switch node.Op {
		case TokenPlus:
			v = left + right
		case TokenMinus:
			v = left - right
		case TokenStar:
			v = left * right
		case TokenSlash:
			if right == 0 {
				return 0, ErrDivideByZero
			}
			v = left / right
		case TokenPercent:
			if right == 0 {
				return 0, ErrDivideByZero
			}
			v = math.Mod(left, right)
		default:
			return 0, ErrUnexpectedToken
		}
		return checkFinite(v)
	}

	return 0, ErrUnexpectedToken
}

func checkFinite(v float64) (float64, error) {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0, ErrNonFinite
	}
	return v, nil
}

// FormatNumber renders a finite value for display. It rounds to at most
// maxDecimals decimal places and at most maxSignificant significant digits,
// strips trailing zeros and the decimal point, and normalises -0 to 0.
//
// Examples: 13 -> "13", 0.1+0.2 -> "0.3", 10/4 -> "2.5", 10/3 -> "3.3333333333",
// 1234567056.6333 -> "1234567056.63".
// Callers must pass a finite value; EvaluateLine guarantees this.
func FormatNumber(v float64) string {
	decimals := maxDecimals
	if v != 0 {
		// Digits before the decimal point, from the magnitude (floor of log10).
		magnitude := int(math.Floor(math.Log10(math.Abs(v))))
		if d := maxSignificant - 1 - magnitude; d < decimals {
			decimals = max(d, 0)
		}
	}

	s := strconv.FormatFloat(v, 'f', decimals, 64)
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimSuffix(s, ".")
	}
	if s == "-0" {
		return "0"
	}
	return s
}

// SumResults adds the values of all lines that have a result. It returns the
// total and how many lines contributed.
func SumResults(results []LineResult) (total float64, count int) {
	for _, res := range results {
		if res.HasResult {
			total += res.Value
			count++
		}
	}
	return total, count
}
