package app

import (
	"github.com/egoist/mygo/ui"
	"github.com/tortoise10h/calculus/internal/eval"
)

// computeTextRanges converts the lexed tokens from all line results into a flat
// slice of ui.TextRange items covering the entire document buffer.
func computeTextRanges(results []eval.LineResult) []ui.TextRange {
	var ranges []ui.TextRange
	lineOffset := 0

	for _, res := range results {
		var codeTokens []eval.Token
		for _, tok := range res.Tokens {
			if tok.Kind != eval.TokenComment {
				codeTokens = append(codeTokens, tok)
			}
		}

		isLHSStart := -1
		if res.AssignedVar != "" && len(codeTokens) >= 2 &&
			codeTokens[0].Kind == eval.TokenIdent && codeTokens[1].Kind == eval.TokenAssign {
			isLHSStart = codeTokens[0].Start
		}

		for _, tok := range res.Tokens {
			var col ui.Color
			weight := 400

			switch tok.Kind {
			case eval.TokenNumber:
				col = SyntaxNumber
			case eval.TokenPlus, eval.TokenMinus, eval.TokenStar, eval.TokenSlash,
				eval.TokenPercent, eval.TokenLParen, eval.TokenRParen:
				col = SyntaxOperator
			case eval.TokenComment:
				col = SyntaxComment
			case eval.TokenAssign:
				col = FullContrast
			case eval.TokenIdent:
				if tok.Start == isLHSStart {
					col = Accent
				} else if len(codeTokens) > 1 && isUnresolved(tok.Text, res.UnresolvedVars) {
					col = Muted
				} else {
					col = FG
				}
			default:
				continue
			}

			ranges = append(ranges, ui.TextRange{
				Start:  lineOffset + tok.Start,
				End:    lineOffset + tok.End,
				Color:  col,
				Weight: weight,
			})
		}
		// Each logical line in the editor buffer ends with a '\n' (1 rune)
		lineOffset += len([]rune(res.RawText)) + 1
	}

	return ranges
}

func isUnresolved(name string, unresolved []string) bool {
	for _, u := range unresolved {
		if u == name {
			return true
		}
	}
	return false
}
