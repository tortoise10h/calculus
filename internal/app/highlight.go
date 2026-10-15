package app

import (
	"github.com/egoist/mygo/ui"
	"github.com/tortoise10h/go-calc/internal/eval"
)

// computeTextRanges converts the lexed tokens from all line results into a flat
// slice of ui.TextRange items covering the entire document buffer.
func computeTextRanges(results []eval.LineResult) []ui.TextRange {
	var ranges []ui.TextRange
	lineOffset := 0

	for _, res := range results {
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
				weight = 500
			case eval.TokenIdent:
				col = FG
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
