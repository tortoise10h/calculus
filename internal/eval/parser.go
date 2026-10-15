package eval

import "errors"

// Sentinel errors returned by the parser and evaluator. They are stored on
// LineResult.Err for inspection only; the UI never displays them, so that a
// half-typed expression does not produce error noise.
var (
	// ErrIncomplete means the expression ended where an operand or ')' was expected,
	// e.g. "20 +" or "(10 + 5".
	ErrIncomplete = errors.New("incomplete expression")
	// ErrUnexpectedToken means the line contains something that is not part of
	// a math expression, e.g. plain text or a stray operator.
	ErrUnexpectedToken = errors.New("unexpected token")
	// ErrDivideByZero means a '/' or '%' had a zero right-hand operand.
	ErrDivideByZero = errors.New("division by zero")
	// ErrNonFinite means the result overflowed to Inf or is NaN.
	ErrNonFinite = errors.New("result is not a finite number")
)

// Node is an AST node for a single-line expression.
type Node interface {
	node()
}

// NumberNode is a numeric literal.
type NumberNode struct {
	Value float64
}

// UnaryNode is a prefix '+' or '-' applied to X.
type UnaryNode struct {
	Op TokenKind
	X  Node
}

// BinaryNode is an infix arithmetic operation.
type BinaryNode struct {
	Op          TokenKind
	Left, Right Node
}

func (*NumberNode) node() {}
func (*UnaryNode) node()  {}
func (*BinaryNode) node() {}

// Grammar (lowest to highest precedence, all binary operators left-associative):
//
//	expr    := term   { ('+' | '-') term }
//	term    := unary  { ('*' | '/' | '%') unary }
//	unary   := ('+' | '-') unary | primary
//	primary := NUMBER | '(' expr ')'
//
// Parse consumes the whole token slice; any leftover token is an error.
// Comment tokens must be removed by the caller.
func Parse(tokens []Token) (Node, error) {
	p := &parser{tokens: tokens}
	if len(tokens) == 0 {
		return nil, ErrIncomplete
	}
	node, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if p.pos < len(p.tokens) {
		return nil, ErrUnexpectedToken
	}
	return node, nil
}

type parser struct {
	tokens []Token
	pos    int
}

func (p *parser) peekKind() TokenKind {
	if p.pos >= len(p.tokens) {
		return TokenEOF
	}
	return p.tokens[p.pos].Kind
}

func (p *parser) parseExpr() (Node, error) {
	left, err := p.parseTerm()
	if err != nil {
		return nil, err
	}
	for {
		switch op := p.peekKind(); op {
		case TokenPlus, TokenMinus:
			p.pos++
			right, err := p.parseTerm()
			if err != nil {
				return nil, err
			}
			left = &BinaryNode{Op: op, Left: left, Right: right}
		default:
			return left, nil
		}
	}
}

func (p *parser) parseTerm() (Node, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for {
		switch op := p.peekKind(); op {
		case TokenStar, TokenSlash, TokenPercent:
			p.pos++
			right, err := p.parseUnary()
			if err != nil {
				return nil, err
			}
			left = &BinaryNode{Op: op, Left: left, Right: right}
		default:
			return left, nil
		}
	}
}

func (p *parser) parseUnary() (Node, error) {
	switch op := p.peekKind(); op {
	case TokenPlus, TokenMinus:
		p.pos++
		x, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &UnaryNode{Op: op, X: x}, nil
	default:
		return p.parsePrimary()
	}
}

func (p *parser) parsePrimary() (Node, error) {
	if p.pos >= len(p.tokens) {
		return nil, ErrIncomplete
	}
	tok := p.tokens[p.pos]
	switch tok.Kind {
	case TokenNumber:
		p.pos++
		return &NumberNode{Value: tok.Value}, nil
	case TokenLParen:
		p.pos++
		inner, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if p.peekKind() != TokenRParen {
			if p.pos >= len(p.tokens) {
				return nil, ErrIncomplete
			}
			return nil, ErrUnexpectedToken
		}
		p.pos++
		return inner, nil
	default:
		return nil, ErrUnexpectedToken
	}
}
