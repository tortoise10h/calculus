package eval

import "strings"

// This file powers the Tab-triggered variable autocomplete. It is pure text
// and result-list logic with no UI dependencies: the app layer owns keys and
// pixels, this layer owns "what can be suggested under this caret".

// CaretContext splits text at the newline before caret, a rune offset into
// text. It returns the zero-based index of the line the caret sits on, the
// rune offset where that line starts, and the line's text without a trailing
// newline. A caret out of range is clamped; text without newlines is line 0.
func CaretContext(text string, caret int) (lineIndex, lineStart int, line string) {
	runes := []rune(text)
	caret = min(max(caret, 0), len(runes))

	for i := 0; i < caret; i++ {
		if runes[i] == '\n' {
			lineIndex++
			lineStart = i + 1
		}
	}

	lineEnd := len(runes)
	for i := caret; i < len(runes); i++ {
		if runes[i] == '\n' {
			lineEnd = i
			break
		}
	}
	return lineIndex, lineStart, string(runes[lineStart:lineEnd])
}

// IdentAt finds the identifier being typed at caret, a rune offset into line.
// It returns the identifier's rune range [start, end) and the partial prefix
// runes[start:caret]. ok is false when the caret is not typing an identifier:
// on whitespace or an operator, on a number, or at the very start of a word
// (nothing of it typed yet).
func IdentAt(line string, caret int) (start, end int, prefix string, ok bool) {
	runes := []rune(line)
	caret = min(max(caret, 0), len(runes))

	for _, tok := range LexLine(line) {
		if tok.Kind != TokenIdent || tok.Start >= caret || caret > tok.End {
			continue
		}
		return tok.Start, tok.End, string(runes[tok.Start:caret]), true
	}
	return 0, 0, "", false
}

// SuggestVars returns the variables defined before line beforeLine whose names
// start with prefix, in first-definition order and without repeats. A variable
// counts as defined only when its line assigned it and produced a value, so a
// suggestion never names something that would read as unresolved at the caret.
// The prefix normally comes from IdentAt; an empty one lists every variable.
func SuggestVars(results []LineResult, beforeLine int, prefix string) []string {
	var out []string
	seen := make(map[string]bool)
	for _, res := range results {
		if res.LineIndex >= beforeLine || res.AssignedVar == "" || !res.HasResult {
			continue
		}
		if seen[res.AssignedVar] || !strings.HasPrefix(res.AssignedVar, prefix) {
			continue
		}
		seen[res.AssignedVar] = true
		out = append(out, res.AssignedVar)
	}
	return out
}
