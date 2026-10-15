package eval

import (
	"errors"
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
	LineIndex      int      // Zero-based index of the line in the document.
	RawText        string   // The line as typed, without the trailing '\n'.
	Value          float64  // Numeric value, valid only when HasResult is true.
	FormattedValue string   // Display string for Value, empty when HasResult is false.
	HasResult      bool     // True when the line is a complete, valid expression.
	AssignedVar    string   // Variable name when the line starts with `<ident> =`, else "".
	UnresolvedVars []string // Identifier names that failed scope lookup on this line.
	Err            error    // Why the line has no result. Nil for blank and comment-only lines.
	Tokens         []Token  // All tokens of the line, including comments, for highlighting.
}

// EvaluateDocument evaluates every line of text sequentially with a shared
// variable scope and returns one LineResult per line. Lines are split on '\n',
// so the slice length is always strings.Count(text, "\n") + 1 (an empty text yields a single empty line).
func EvaluateDocument(text string) []LineResult {
	lines := strings.Split(text, "\n")
	results := make([]LineResult, len(lines))
	scope := make(map[string]float64)
	for i, line := range lines {
		results[i] = EvaluateLineWithScope(i, line, scope)
	}
	return results
}

// EvaluateLine lexes, parses and evaluates one line without an ambient scope.
// Blank lines, comment-only lines and lines that do not produce a value return
// HasResult == false; no line ever causes a panic.
func EvaluateLine(index int, line string) LineResult {
	return EvaluateLineWithScope(index, line, nil)
}

// EvaluateLineWithScope lexes, parses and evaluates one line with the given variable scope.
func EvaluateLineWithScope(index int, line string, scope map[string]float64) LineResult {
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

	if len(code) >= 2 && code[0].Kind == TokenIdent && code[1].Kind == TokenAssign {
		res.AssignedVar = code[0].Text
	}

	addUnresolved := func(name string) {
		for _, u := range res.UnresolvedVars {
			if u == name {
				return
			}
		}
		res.UnresolvedVars = append(res.UnresolvedVars, name)
	}

	node, err := Parse(code)
	if err != nil {
		res.Err = err
		if errors.Is(err, ErrIncomplete) {
			start := 0
			if res.AssignedVar != "" {
				start = 2
			}
			for _, tok := range code[start:] {
				if tok.Kind == TokenIdent {
					if scope == nil {
						addUnresolved(tok.Text)
					} else if _, ok := scope[tok.Text]; !ok {
						addUnresolved(tok.Text)
					}
				}
			}
		}
		return res
	}

	var collectUnresolved func(n Node)
	collectUnresolved = func(n Node) {
		if n == nil {
			return
		}
		switch node := n.(type) {
		case *IdentNode:
			if scope == nil {
				addUnresolved(node.Name)
			} else if _, ok := scope[node.Name]; !ok {
				addUnresolved(node.Name)
			}
		case *AssignNode:
			collectUnresolved(node.Expr)
		case *UnaryNode:
			collectUnresolved(node.X)
		case *BinaryNode:
			collectUnresolved(node.Left)
			collectUnresolved(node.Right)
		}
	}
	collectUnresolved(node)

	value, err := EvalWithScope(node, scope)
	if err != nil {
		res.Err = err
		return res
	}

	res.Value = value
	res.FormattedValue = FormatNumber(value)
	res.HasResult = true
	return res
}

// Eval computes the value of an AST with no variables in scope. It returns
// ErrUndefinedVar if any identifier is encountered.
func Eval(n Node) (float64, error) {
	return EvalWithScope(n, nil)
}

// EvalWithScope computes the value of an AST using scope for variable lookup
// and mutation. It returns ErrDivideByZero for '/' or '%' by zero, and
// ErrNonFinite if any intermediate or final value is NaN or ±Inf.
func EvalWithScope(n Node, scope map[string]float64) (float64, error) {
	switch node := n.(type) {
	case *NumberNode:
		return checkFinite(node.Value)

	case *IdentNode:
		if scope == nil {
			return 0, ErrUndefinedVar
		}
		val, ok := scope[node.Name]
		if !ok {
			return 0, ErrUndefinedVar
		}
		return val, nil

	case *AssignNode:
		val, err := EvalWithScope(node.Expr, scope)
		if err != nil {
			return 0, err
		}
		if scope != nil {
			scope[node.Name] = val
		}
		return val, nil

	case *UnaryNode:
		x, err := EvalWithScope(node.X, scope)
		if err != nil {
			return 0, err
		}
		if node.Op == TokenMinus {
			return checkFinite(-x)
		}
		return checkFinite(x)

	case *BinaryNode:
		left, err := EvalWithScope(node.Left, scope)
		if err != nil {
			return 0, err
		}
		right, err := EvalWithScope(node.Right, scope)
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
