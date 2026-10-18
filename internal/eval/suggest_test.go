package eval

import (
	"reflect"
	"testing"
)

func TestCaretContext(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		caret     int
		wantIndex int
		wantStart int
		wantLine  string
	}{
		{"empty text", "", 0, 0, 0, ""},
		{"single line", "meal = 1", 4, 0, 0, "meal = 1"},
		{"second line", "a = 1\nm", 6, 1, 6, "m"},
		{"caret on newline stays on its line", "ab\ncd", 2, 0, 0, "ab"},
		{"caret after newline starts the next line", "ab\ncd", 3, 1, 3, "cd"},
		{"blank middle line", "a\n\nb", 2, 1, 2, ""},
		{"caret past end clamps", "ab\ncd", 99, 1, 3, "cd"},
		{"negative caret clamps", "ab\ncd", -5, 0, 0, "ab"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			index, start, line := CaretContext(tt.text, tt.caret)
			if index != tt.wantIndex || start != tt.wantStart || line != tt.wantLine {
				t.Fatalf("CaretContext(%q, %d) = (%d, %d, %q), want (%d, %d, %q)",
					tt.text, tt.caret, index, start, line, tt.wantIndex, tt.wantStart, tt.wantLine)
			}
		})
	}
}

func TestIdentAt(t *testing.T) {
	tests := []struct {
		name       string
		line       string
		caret      int
		wantStart  int
		wantEnd    int
		wantPrefix string
		wantOK     bool
	}{
		{"prefix at end of word", "x = m", 5, 4, 5, "m", true},
		{"caret mid-word", "meal", 2, 0, 4, "me", true},
		{"caret at word start is nothing typed", "meal", 0, 0, 0, "", false},
		{"caret at word start after space", "x = meal", 4, 0, 0, "", false},
		{"whole word at its end", "meal", 4, 0, 4, "meal", true},
		{"caret on operator", "m + 1", 2, 0, 0, "", false},
		{"caret after space", "m ", 2, 0, 0, "", false},
		{"caret in number", "123", 2, 0, 0, "", false},
		{"empty line", "", 0, 0, 0, "", false},
		{"suffix number is not an ident", "5m", 2, 0, 0, "", false},
		{"underscores and digits", "sub_1x", 6, 0, 6, "sub_1x", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, prefix, ok := IdentAt(tt.line, tt.caret)
			if start != tt.wantStart || end != tt.wantEnd || prefix != tt.wantPrefix || ok != tt.wantOK {
				t.Fatalf("IdentAt(%q, %d) = (%d, %d, %q, %v), want (%d, %d, %q, %v)",
					tt.line, tt.caret, start, end, prefix, ok, tt.wantStart, tt.wantEnd, tt.wantPrefix, tt.wantOK)
			}
		})
	}
}

func TestSuggestVars(t *testing.T) {
	// Line layout: 0 drink, 1 meal, 2 mosh, 3 use, 4 bad (incomplete), 5 use again.
	text := "drink = 50 + 10 + 100\nmeal = 100 + 70\nmosh = 50 + 1\nm + 1\nbad = \nmeal = 2"
	results := EvaluateDocument(text)

	tests := []struct {
		name   string
		before int
		prefix string
		want   []string
	}{
		{"prefix m matches two", 3, "m", []string{"meal", "mosh"}},
		{"prefix me matches one", 3, "me", []string{"meal"}},
		{"empty prefix lists all in order", 3, "", []string{"drink", "meal", "mosh"}},
		{"unknown prefix matches none", 3, "z", nil},
		{"nothing defined before the first line", 0, "", nil},
		{"caret line itself is excluded", 2, "m", []string{"meal"}},
		{"incomplete assignment is not defined", 5, "b", nil},
		{"redefinition keeps first order", 6, "", []string{"drink", "meal", "mosh"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SuggestVars(results, tt.before, tt.prefix)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("SuggestVars(before %d, %q) = %v, want %v", tt.before, tt.prefix, got, tt.want)
			}
		})
	}
}
