package app

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestAppViewShell(t *testing.T) {
	a := New()
	tt := ui.NewTester(a.View, 720, 520)

	if !tt.HasText("Go-Calc") {
		t.Fatalf("expected title text 'Go-Calc', got texts: %v", tt.Texts())
	}
	if !tt.HasText("Ready") {
		t.Fatalf("expected status bar text 'Ready', got texts: %v", tt.Texts())
	}
}
