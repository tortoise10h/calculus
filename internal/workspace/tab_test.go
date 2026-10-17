package workspace

import (
	"strings"
	"testing"
)

func TestNewStartsWithOneUntitledTab(t *testing.T) {
	w := New()

	if len(w.Tabs) != 1 {
		t.Fatalf("tabs = %d, want 1", len(w.Tabs))
	}
	if w.Tabs[0].Title != "Untitled" {
		t.Errorf("title = %q, want %q", w.Tabs[0].Title, "Untitled")
	}
	if w.Tabs[0].Content != "" {
		t.Errorf("content = %q, want empty", w.Tabs[0].Content)
	}
	if w.Tabs[0].ID == "" {
		t.Error("tab ID is empty")
	}
	if w.ActiveID != w.Tabs[0].ID {
		t.Errorf("ActiveID = %q, want %q", w.ActiveID, w.Tabs[0].ID)
	}
	if w.Version != StoreVersion {
		t.Errorf("version = %d, want %d", w.Version, StoreVersion)
	}
}

func TestSeedHoldsTheStarterDocument(t *testing.T) {
	w := Seed()

	if len(w.Tabs) != 1 {
		t.Fatalf("tabs = %d, want 1", len(w.Tabs))
	}
	tab := w.Tabs[0]
	if tab.Title != "Starter" {
		t.Errorf("title = %q, want %q", tab.Title, "Starter")
	}
	for _, want := range []string{
		"20 + 5 - 12",
		"drink_cost = 10 + 5 + 20",
		"total_cost = drink_cost + meal_cost",
	} {
		if !strings.Contains(tab.Content, want) {
			t.Errorf("seeded content is missing %q:\n%s", want, tab.Content)
		}
	}
	if w.ActiveIndex() != 0 {
		t.Errorf("ActiveIndex = %d, want 0", w.ActiveIndex())
	}
}

func TestAddAppendsAndActivates(t *testing.T) {
	w := New()
	second := w.Add("", "1 + 1")
	third := w.Add("Groceries", "2 + 2")

	if len(w.Tabs) != 3 {
		t.Fatalf("tabs = %d, want 3", len(w.Tabs))
	}
	if second.Title != "Untitled 2" {
		t.Errorf("second title = %q, want %q", second.Title, "Untitled 2")
	}
	if third.Title != "Groceries" {
		t.Errorf("third title = %q, want %q", third.Title, "Groceries")
	}
	if w.Active() != third {
		t.Error("Add did not make the new tab active")
	}
	if w.ActiveIndex() != 2 {
		t.Errorf("ActiveIndex = %d, want 2", w.ActiveIndex())
	}
}

func TestUntitledNamingSkipsTakenNames(t *testing.T) {
	w := New() // already holds "Untitled"

	if got := w.Add("", "").Title; got != "Untitled 2" {
		t.Fatalf("second tab title = %q, want %q", got, "Untitled 2")
	}
	if got := w.Add("", "").Title; got != "Untitled 3" {
		t.Fatalf("third tab title = %q, want %q", got, "Untitled 3")
	}

	// A closed name becomes available again.
	w.Close(1)
	if got := w.Add("", "").Title; got != "Untitled 2" {
		t.Errorf("reused title = %q, want %q", got, "Untitled 2")
	}
}

func TestSetActiveClampsOutOfRange(t *testing.T) {
	w := Seed()
	w.Add("Second", "")
	w.Add("Third", "")

	w.SetActive(-5)
	if w.ActiveIndex() != 0 {
		t.Errorf("ActiveIndex after SetActive(-5) = %d, want 0", w.ActiveIndex())
	}
	w.SetActive(99)
	if w.ActiveIndex() != 2 {
		t.Errorf("ActiveIndex after SetActive(99) = %d, want 2", w.ActiveIndex())
	}
	if got := w.Active().Title; got != "Third" {
		t.Errorf("Active().Title = %q, want %q", got, "Third")
	}
}

func TestCloseKeepsTheLastTab(t *testing.T) {
	w := Seed()
	if w.Close(0) {
		t.Fatal("Close removed the only tab; the last tab must be unclosable")
	}
	if len(w.Tabs) != 1 {
		t.Fatalf("tabs = %d, want 1", len(w.Tabs))
	}
	if w.Active() == nil {
		t.Fatal("Active() is nil after a rejected close")
	}
}

func TestCloseOutOfRangeIsANoOp(t *testing.T) {
	w := Seed()
	w.Add("Second", "")

	for _, i := range []int{-1, 2, 100} {
		if w.Close(i) {
			t.Errorf("Close(%d) returned true, want false", i)
		}
	}
	if len(w.Tabs) != 2 {
		t.Fatalf("tabs = %d, want 2", len(w.Tabs))
	}
}

func TestCloseActivePicksMinIndexAndLen(t *testing.T) {
	// Closing the active middle tab lands on the tab that shifted into its slot.
	w := Seed()
	w.Add("Second", "")
	w.Add("Third", "")
	w.SetActive(1)
	if !w.Close(1) {
		t.Fatal("Close(1) returned false")
	}
	if got := w.Titles(); got[0] != "Starter" || got[1] != "Third" {
		t.Fatalf("titles = %v", got)
	}
	if got := w.Active().Title; got != "Third" {
		t.Errorf("Active = %q, want %q", got, "Third")
	}

	// Closing the last tab when it is active lands on the new last tab.
	if !w.Close(1) {
		t.Fatal("Close(1) returned false")
	}
	if got := w.Active().Title; got != "Starter" {
		t.Errorf("Active = %q, want %q", got, "Starter")
	}
}

func TestCloseInactiveLeavesActiveAlone(t *testing.T) {
	w := Seed()
	w.Add("Second", "")
	w.Add("Third", "")
	w.SetActive(2)

	if !w.Close(0) {
		t.Fatal("Close(0) returned false")
	}
	if got := w.Active().Title; got != "Third" {
		t.Errorf("Active = %q, want %q", got, "Third")
	}
	if w.ActiveIndex() != 1 {
		t.Errorf("ActiveIndex = %d, want 1", w.ActiveIndex())
	}
}

func TestRenameTrimsAndIgnoresEmpty(t *testing.T) {
	w := Seed()

	w.Rename(0, "  Groceries  ")
	if got := w.Tabs[0].Title; got != "Groceries" {
		t.Errorf("title = %q, want %q", got, "Groceries")
	}

	for _, blank := range []string{"", "   ", "\t\n"} {
		w.Rename(0, blank)
		if got := w.Tabs[0].Title; got != "Groceries" {
			t.Errorf("blank rename %q changed the title to %q", blank, got)
		}
	}

	// Out-of-range indices must not panic.
	w.Rename(-1, "x")
	w.Rename(9, "x")
}

func TestTitlesAndNilSafety(t *testing.T) {
	w := Seed()
	w.Add("Second", "")

	got := w.Titles()
	if len(got) != 2 || got[0] != "Starter" || got[1] != "Second" {
		t.Fatalf("Titles = %v", got)
	}

	var nilWS *Workspace
	if nilWS.Active() != nil {
		t.Error("Active() on a nil workspace should be nil")
	}
	if nilWS.ActiveIndex() != -1 {
		t.Errorf("ActiveIndex() on a nil workspace = %d, want -1", nilWS.ActiveIndex())
	}
	if nilWS.Add("x", "") != nil {
		t.Error("Add on a nil workspace should return nil")
	}
	if nilWS.Close(0) {
		t.Error("Close on a nil workspace should return false")
	}
	// None of these may panic.
	nilWS.SetActive(3)
	nilWS.Rename(0, "x")
	if nilWS.Titles() != nil {
		t.Error("Titles on a nil workspace should be nil")
	}

	// An empty (but non-nil) workspace reports no active tab.
	empty := &Workspace{}
	if empty.ActiveIndex() != -1 {
		t.Errorf("empty ActiveIndex = %d, want -1", empty.ActiveIndex())
	}
	empty.SetActive(0)
	if empty.Active() != nil {
		t.Error("empty Active() should be nil")
	}
}

func TestNewIDIsUniqueOverManyDraws(t *testing.T) {
	seen := make(map[string]struct{}, 1000)
	for i := 0; i < 1000; i++ {
		id := NewID()
		if id == "" {
			t.Fatal("NewID returned an empty string")
		}
		if _, dup := seen[id]; dup {
			t.Fatalf("NewID returned the duplicate %q on draw %d", id, i)
		}
		seen[id] = struct{}{}
	}
}

func TestRepairFixesEmptyAndUnknownActiveID(t *testing.T) {
	// No tabs at all -> the seeded starter tab.
	empty := &Workspace{}
	empty.repair()
	if len(empty.Tabs) != 1 || empty.Tabs[0].Title != "Starter" {
		t.Fatalf("repaired empty workspace = %+v", empty.Tabs)
	}

	// Unknown ActiveID -> Tabs[0]; missing IDs get fresh ones.
	w := &Workspace{
		ActiveID: "nope",
		Tabs: []*Tab{
			{Title: "First"},
			{Title: "Second", ID: "keep-me"},
		},
	}
	w.repair()
	if w.ActiveIndex() != 0 {
		t.Errorf("ActiveIndex = %d, want 0", w.ActiveIndex())
	}
	if w.Tabs[0].ID == "" {
		t.Error("empty tab ID was not repaired")
	}
	if w.Tabs[1].ID != "keep-me" {
		t.Errorf("existing tab ID was changed to %q", w.Tabs[1].ID)
	}
}
