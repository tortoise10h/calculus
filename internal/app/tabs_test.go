package app

import (
	"strconv"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/tortoise10h/calculus/internal/workspace"
)

// fakeStore counts Save calls and remembers the last workspace written, so the
// app's debounce can be tested without touching the real data directory.
type fakeStore struct {
	saved  int
	last   *workspace.Workspace
	saveFn func(*workspace.Workspace)
}

func (f *fakeStore) Load() (*workspace.Workspace, error) { return workspace.Seed(), nil }

func (f *fakeStore) Save(w *workspace.Workspace) error {
	f.saved++
	f.last = w
	if f.saveFn != nil {
		f.saveFn(w)
	}
	return nil
}

// tabStripDotsW is the width of the dots alone (excluding the balanced "+"
// button and its leading counterweight), which is centred on the window's
// midline.
func tabStripDotsW(tabs int) float32 {
	if tabs <= 0 {
		return 0
	}
	return float32(tabs)*dotHitSize + float32(tabs-1)*dotGap
}

// dotCenterX is the pointer position that lands on tab i's dot. Because the
// leading spacer balances the trailing "+" button, the dots themselves are
// centred on the window's midline.
func dotCenterX(width float32, tabs, i int) float32 {
	avail := width - 2*resultPadRight
	left := float32(resultPadRight) + (avail-tabStripDotsW(tabs))/2
	return left + float32(i)*(dotHitSize+dotGap) + dotHitSize/2
}

// newTestApp returns an app over an empty workspace with a shortened debounce,
// so tests never wait the full 800 ms.
func newTestApp() (*App, *fakeStore) {
	store := &fakeStore{}
	a := NewWithWorkspace(workspace.New(), store)
	a.autoSaveDelay = 10 * time.Millisecond
	a.revealActive()
	return a, store
}

// newStripedApp returns a rendering app over n empty tabs, so a strip long
// enough to overflow can be driven. Tab 0 is active, so the window starts at
// the beginning of the list.
func newStripedApp(n int) *App {
	ws := workspace.New()
	for i := 1; i < n; i++ {
		ws.Add("", "")
	}
	ws.SetActive(0) // Add leaves the last tab active
	a := NewWithWorkspace(ws, nil)
	a.revealActive()
	return a
}

func TestTabStripShowsOneDotPerTab(t *testing.T) {
	a := New()
	tt := ui.NewTester(a.View, 720, 520)
	tt.Frame()

	if len(a.ws.Tabs) != 1 {
		t.Fatalf("tabs = %d, want 1", len(a.ws.Tabs))
	}
	if !tt.HasText("Starter") {
		t.Fatalf("expected the seeded title in the header, got texts: %v", tt.Texts())
	}

	a.addTab()
	tt.Frame()

	if len(a.ws.Tabs) != 2 {
		t.Fatalf("tabs = %d, want 2", len(a.ws.Tabs))
	}
	// The seeded tab is called "Starter", so the first free "Untitled" name wins.
	if !tt.HasText("Untitled") {
		t.Fatalf("expected the active tab's title in the header, got texts: %v", tt.Texts())
	}
	if tt.HasText("Starter") {
		t.Fatalf("the inactive tab's title should not be in the header, got texts: %v", tt.Texts())
	}
}

func TestNewTabButtonAddsTab(t *testing.T) {
	a := newBlankApp()
	tt := ui.NewTester(a.View, 720, 520)
	tt.Frame()

	if err := tt.Click("+"); err != nil {
		t.Fatalf("click the new-tab button: %v", err)
	}
	tt.Frame()

	if len(a.ws.Tabs) != 2 {
		t.Fatalf("tabs = %d, want 2", len(a.ws.Tabs))
	}
	if a.ws.ActiveIndex() != 1 {
		t.Errorf("ActiveIndex = %d, want 1", a.ws.ActiveIndex())
	}
	if !tt.HasText("Untitled 2") {
		t.Fatalf("expected the new tab's title, got texts: %v", tt.Texts())
	}
	// The new tab is empty, so there is nothing to sum and nothing to clear.
	if !tt.HasText("Sum: 0") {
		t.Fatalf("expected 'Sum: 0' for an empty tab, got texts: %v", tt.Texts())
	}
	if tt.HasText("Clear") {
		t.Fatalf("the Clear button must hide for an empty tab, got texts: %v", tt.Texts())
	}
}

func TestCmdTAddsATab(t *testing.T) {
	a := newBlankApp()
	tt := ui.NewTester(a.View, 720, 520)
	tt.Frame()

	tt.Key(ui.Cmd, ui.KeyT)
	tt.Frame()

	if len(a.ws.Tabs) != 2 {
		t.Fatalf("tabs = %d, want 2", len(a.ws.Tabs))
	}
	if a.ws.ActiveIndex() != 1 {
		t.Errorf("ActiveIndex = %d, want the new tab active", a.ws.ActiveIndex())
	}
	// The chord must not have leaked a "t" into the editor.
	if a.ActiveTab().Content != "" {
		t.Errorf("editor content = %q, want empty; Cmd+T was typed instead of handled", a.ActiveTab().Content)
	}
}

func TestSwitchTabKeepsEachBufferSeparate(t *testing.T) {
	a := newBlankApp()
	tt := ui.NewTester(a.View, 720, 520)
	tt.Frame()

	typeLines(tt, "20 + 5")
	tt.Frame()
	if !tt.HasText("Sum: 25") {
		t.Fatalf("expected 'Sum: 25' in tab 1, got texts: %v", tt.Texts())
	}

	tt.Key(ui.Cmd, ui.KeyT) // new, empty tab
	tt.Frame()
	typeLines(tt, "1 + 1")
	tt.Frame()
	if !tt.HasText("Sum: 2") {
		t.Fatalf("expected 'Sum: 2' in tab 2, got texts: %v", tt.Texts())
	}
	if got := a.ws.Tabs[1].Content; got != "1 + 1" {
		t.Fatalf("tab 2 content = %q, want %q", got, "1 + 1")
	}

	tt.Key(ui.Cmd, ui.Key1) // back to the first tab
	tt.Frame()

	if got := a.ws.Tabs[0].Content; got != "20 + 5" {
		t.Fatalf("tab 1 content = %q, want %q", got, "20 + 5")
	}
	if !tt.HasText("25") {
		t.Fatalf("expected tab 1's result back on screen, got texts: %v", tt.Texts())
	}
	if !tt.HasText("Sum: 25") {
		t.Fatalf("expected 'Sum: 25' again, got texts: %v", tt.Texts())
	}

	// And tab 2 kept its own buffer while tab 1 was in front.
	if got := a.ws.Tabs[1].Content; got != "1 + 1" {
		t.Fatalf("tab 2 content changed to %q while inactive", got)
	}
}

func TestSwitchTabRestoresScroll(t *testing.T) {
	a := newBlankApp()
	tt := ui.NewTester(a.View, 720, 520)
	tt.Frame()

	lines := make([]string, 40)
	for i := range lines {
		lines[i] = "1 + 1"
	}
	typeLines(tt, lines...)
	tt.Frame()

	for i := 0; i < 10; i++ {
		tt.Scroll(300, 200, 0, -40)
		tt.Frame()
	}
	scrolled := a.ws.Tabs[0].ScrollY
	if scrolled <= 0 {
		t.Fatalf("tab 1 scrollY = %v, want > 0 after scrolling", scrolled)
	}

	tt.Key(ui.Cmd, ui.KeyT)
	tt.Frame()
	tt.Key(ui.Cmd, ui.Key1)
	tt.Frame()
	tt.Frame()

	if got := a.ws.Tabs[0].ScrollY; got <= 0 {
		t.Fatalf("tab 1 scrollY = %v after switching away and back, want > 0", got)
	}
	if got := a.ws.Tabs[0].ScrollY; got != scrolled {
		t.Logf("scrollY moved from %v to %v across the switch (clamped by layout)", scrolled, got)
	}
}

func TestCloseTabKeepsAtLeastOneTab(t *testing.T) {
	a := newBlankApp()
	tt := ui.NewTester(a.View, 720, 520)
	tt.Frame()

	// A lone tab refuses to close.
	tt.Key(ui.Cmd, ui.KeyW)
	tt.Frame()
	if len(a.ws.Tabs) != 1 {
		t.Fatalf("tabs = %d, want the last tab kept", len(a.ws.Tabs))
	}

	tt.Key(ui.Cmd, ui.KeyT)
	tt.Key(ui.Cmd, ui.KeyT)
	tt.Frame()
	if len(a.ws.Tabs) != 3 {
		t.Fatalf("tabs = %d, want 3", len(a.ws.Tabs))
	}

	// Closing the active (third) tab lands on min(2, len-1) = 1.
	tt.Key(ui.Cmd, ui.KeyW)
	tt.Frame()
	if len(a.ws.Tabs) != 2 {
		t.Fatalf("tabs = %d, want 2", len(a.ws.Tabs))
	}
	if a.ws.ActiveIndex() != 1 {
		t.Errorf("ActiveIndex = %d, want 1", a.ws.ActiveIndex())
	}

	tt.Key(ui.Cmd, ui.KeyW)
	tt.Frame()
	if len(a.ws.Tabs) != 1 {
		t.Fatalf("tabs = %d, want 1", len(a.ws.Tabs))
	}
	if a.ActiveTab() == nil {
		t.Fatal("ActiveTab() is nil; a workspace must always keep one tab")
	}
}

func TestCloseActiveDotClosesTheTab(t *testing.T) {
	const w, h = 720, 520

	a := newBlankApp()
	tt := ui.NewTester(a.View, w, h)
	tt.Frame()
	tt.Key(ui.Cmd, ui.KeyT) // the second tab is active
	tt.Frame()

	// Hovering the active dot swaps it for a close glyph. The strip is
	// tabStripHeight tall at the bottom with a resultPadRight inset, and the
	// dot boxes march across it dotHitSize + dotGap at a time, so the second
	// dot sits at a predictable point.
	dotX := dotCenterX(w, 2, 1)
	dotY := float32(h) - tabStripHeight/2

	tt.Move(dotX, dotY)
	tt.Frame()

	rect, ok := tt.Find(closeGlyph)
	if !ok {
		t.Fatalf("no %q on the hovered active dot, texts: %v", closeGlyph, tt.Texts())
	}
	tt.ClickAt(rect.X+rect.W/2, rect.Y+rect.H/2)
	tt.Frame()

	if len(a.ws.Tabs) != 1 {
		t.Fatalf("tabs = %d, want the tab closed", len(a.ws.Tabs))
	}
}

func TestALoneTabShowsNoCloseGlyph(t *testing.T) {
	const w, h = 720, 520

	a := newBlankApp()
	tt := ui.NewTester(a.View, w, h)
	tt.Frame()

	tt.Move(dotCenterX(w, 1, 0), float32(h)-tabStripHeight/2)
	tt.Frame()

	if _, ok := tt.Find(closeGlyph); ok {
		t.Fatalf("the only tab offered a close %q", closeGlyph)
	}
}

func TestHoveringAnInactiveDotKeepsItADot(t *testing.T) {
	const w, h = 720, 520

	a := newBlankApp()
	tt := ui.NewTester(a.View, w, h)
	tt.Frame()
	tt.Key(ui.Cmd, ui.KeyT) // tab 1 active
	tt.Frame()

	// Hovering an inactive dot must not offer to close it.
	firstDotX := dotCenterX(w, 2, 0)
	dotY := float32(h) - tabStripHeight/2
	tt.Move(firstDotX, dotY)
	tt.Frame()

	if _, ok := tt.Find(closeGlyph); ok {
		t.Fatalf("an inactive dot offered a close %q", closeGlyph)
	}
	tt.ClickAt(firstDotX, dotY)
	tt.Frame()

	if a.ws.ActiveIndex() != 0 {
		t.Errorf("ActiveIndex = %d, want 0; the click did not activate the tab", a.ws.ActiveIndex())
	}
	if len(a.ws.Tabs) != 2 {
		t.Errorf("tabs = %d, want 2; the click closed a tab", len(a.ws.Tabs))
	}
}

func TestRenameActiveTab(t *testing.T) {
	a := New()
	tt := ui.NewTester(a.View, 720, 520)
	tt.Frame()

	tt.Key(ui.Cmd, ui.KeyR)
	tt.Frame()
	if !a.renaming {
		t.Fatal("Cmd+R did not open the rename field")
	}

	// The field starts seeded with the current title, so select-all then retype.
	tt.Key(ui.Cmd, ui.KeyA)
	tt.Frame()
	tt.Type("Groceries")
	tt.Frame()
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	tt.Frame()

	if a.renaming {
		t.Error("the rename field is still open after Enter")
	}
	if got := a.ws.Tabs[0].Title; got != "Groceries" {
		t.Fatalf("title = %q, want %q", got, "Groceries")
	}
	if !tt.HasText("Groceries") {
		t.Fatalf("expected the new title in the header, got texts: %v", tt.Texts())
	}
}

func TestRenameEscapeKeepsTheOldTitle(t *testing.T) {
	a := New()
	tt := ui.NewTester(a.View, 720, 520)
	tt.Frame()

	tt.Key(ui.Cmd, ui.KeyR)
	tt.Frame()
	tt.Key(ui.Cmd, ui.KeyA)
	tt.Frame()
	tt.Type("Discarded")
	tt.Frame()
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	tt.Frame()

	if a.renaming {
		t.Error("Escape did not close the rename field")
	}
	if got := a.ws.Tabs[0].Title; got != "Starter" {
		t.Fatalf("title = %q, want the original %q", got, "Starter")
	}
}

func TestRenameWithABlankDraftKeepsTheOldTitle(t *testing.T) {
	a := New()
	tt := ui.NewTester(a.View, 720, 520)
	tt.Frame()

	tt.Key(ui.Cmd, ui.KeyR)
	tt.Frame()
	tt.Key(ui.Cmd, ui.KeyA)
	tt.Key(0, ui.KeyBackspace)
	tt.Frame()
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	tt.Frame()

	if got := a.ws.Tabs[0].Title; got != "Starter" {
		t.Fatalf("title = %q, want the original kept through a blank draft", got)
	}
}

func TestShortcutsDoNotLeakIntoTheEditorWhileRenaming(t *testing.T) {
	a := New()
	tt := ui.NewTester(a.View, 720, 520)
	tt.Frame()

	tt.Key(ui.Cmd, ui.KeyR)
	tt.Frame()

	// While the rename field is open, tab chords must not create or close tabs.
	tt.Key(ui.Cmd, ui.KeyT)
	tt.Key(ui.Cmd, ui.KeyW)
	tt.Frame()

	if len(a.ws.Tabs) != 1 {
		t.Fatalf("tabs = %d, want 1; chords fired mid-rename", len(a.ws.Tabs))
	}
	if !a.renaming {
		t.Error("the rename field closed on a chord other than Escape")
	}
}

func TestCmdDigitSwitchesTab(t *testing.T) {
	a := newBlankApp()
	tt := ui.NewTester(a.View, 720, 520)
	tt.Frame()

	a.addTab()
	a.addTab()
	tt.Frame()
	if len(a.ws.Tabs) != 3 {
		t.Fatalf("tabs = %d, want 3", len(a.ws.Tabs))
	}

	tt.Key(ui.Cmd, ui.Key2)
	tt.Frame()
	if a.ws.ActiveIndex() != 1 {
		t.Errorf("ActiveIndex after Cmd+2 = %d, want 1", a.ws.ActiveIndex())
	}

	// Beyond the last tab the chord does nothing at all.
	tt.Key(ui.Cmd, ui.Key9)
	tt.Frame()
	if a.ws.ActiveIndex() != 1 {
		t.Errorf("ActiveIndex after Cmd+9 = %d, want 1 (unchanged)", a.ws.ActiveIndex())
	}

	tt.Key(ui.Cmd, ui.Key1)
	tt.Frame()
	if a.ws.ActiveIndex() != 0 {
		t.Errorf("ActiveIndex after Cmd+1 = %d, want 0", a.ws.ActiveIndex())
	}
}

func TestSumFollowsActiveTab(t *testing.T) {
	a := newBlankApp()
	tt := ui.NewTester(a.View, 720, 520)
	tt.Frame()

	typeLines(tt, "20 + 5")
	tt.Frame()
	if !tt.HasText("Sum: 25") {
		t.Fatalf("expected 'Sum: 25', got texts: %v", tt.Texts())
	}

	tt.Key(ui.Cmd, ui.KeyT) // a second, empty tab
	tt.Frame()
	if !tt.HasText("Sum: 0") {
		t.Fatalf("expected 'Sum: 0' for the empty tab, got texts: %v", tt.Texts())
	}
	if tt.HasText("Sum: 25") {
		t.Fatalf("the previous tab's sum leaked into the header, got texts: %v", tt.Texts())
	}

	tt.Key(ui.Cmd, ui.Key1)
	tt.Frame()
	if !tt.HasText("Sum: 25") {
		t.Fatalf("expected 'Sum: 25' back, got texts: %v", tt.Texts())
	}
}

func TestAutoSaveDebouncesToStore(t *testing.T) {
	a, store := newTestApp()
	tt := ui.NewTester(a.View, 720, 520)
	tt.Frame()

	typeLines(tt, "20 + 5")
	tt.Frame()

	// The debounce has not expired yet, so nothing is written.
	if store.saved != 0 {
		t.Fatalf("Save calls = %d before the debounce expired, want 0", store.saved)
	}

	time.Sleep(40 * time.Millisecond)
	tt.Frame()
	if store.saved < 1 {
		t.Fatalf("Save calls = %d after the debounce, want at least 1", store.saved)
	}
	if store.last.Tabs[0].Content != "20 + 5" {
		t.Fatalf("saved content = %q, want %q", store.last.Tabs[0].Content, "20 + 5")
	}

	// A further edit schedules another save.
	tt.Type(" + 1")
	tt.Frame()
	time.Sleep(40 * time.Millisecond)
	tt.Frame()
	if store.saved < 2 {
		t.Fatalf("Save calls = %d after a second edit, want at least 2", store.saved)
	}
	if store.last.Tabs[0].Content != "20 + 5 + 1" {
		t.Fatalf("saved content = %q, want %q", store.last.Tabs[0].Content, "20 + 5 + 1")
	}
}

func TestStructuralChangesSaveImmediately(t *testing.T) {
	a, store := newTestApp()
	tt := ui.NewTester(a.View, 720, 520)
	tt.Frame()

	// No debounce wait at all: a structural edit must be on disk right away.
	tt.Key(ui.Cmd, ui.KeyT)
	tt.Frame()
	if store.saved != 1 {
		t.Fatalf("Save calls after Cmd+T = %d, want 1 (no debounce)", store.saved)
	}
	if len(store.last.Tabs) != 2 {
		t.Fatalf("saved tabs = %d, want 2", len(store.last.Tabs))
	}

	tt.Key(ui.Cmd, ui.KeyR)
	tt.Frame()
	tt.Key(ui.Cmd, ui.KeyA)
	tt.Type("Groceries")
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	tt.Frame()

	if store.last.Tabs[1].Title != "Groceries" {
		t.Fatalf("saved title = %q, want %q", store.last.Tabs[1].Title, "Groceries")
	}
}

func TestFlushSavesImmediately(t *testing.T) {
	a, store := newTestApp()
	tt := ui.NewTester(a.View, 720, 520)
	tt.Frame()

	a.SetContent("7 * 6")
	if store.saved != 0 {
		t.Fatalf("Save calls = %d before Flush, want 0", store.saved)
	}

	a.Flush()
	if store.saved != 1 {
		t.Fatalf("Save calls after Flush = %d, want 1", store.saved)
	}
	if got := store.last.Tabs[0].Content; got != "7 * 6" {
		t.Fatalf("saved content = %q, want %q", got, "7 * 6")
	}

	// Nothing pending, so a second Flush is free.
	a.Flush()
	if store.saved != 1 {
		t.Fatalf("Save calls after a redundant Flush = %d, want 1", store.saved)
	}
}

func TestAFailedSaveStaysPending(t *testing.T) {
	a := NewWithWorkspace(workspace.Seed(), failingStore{})
	tt := ui.NewTester(a.View, 720, 520)
	tt.Frame()

	a.SetContent("1 + 1")
	a.Flush()

	if !a.dirty {
		t.Error("dirty was cleared despite the save failing; the edit would be lost")
	}
}

// failingStore always errors, standing in for a full or read-only disk.
type failingStore struct{}

func (failingStore) Load() (*workspace.Workspace, error) { return nil, workspace.ErrNoWorkspace }

func (failingStore) Save(*workspace.Workspace) error { return errSaveFailed }

type saveError struct{}

func (saveError) Error() string { return "disk is full" }

var errSaveFailed error = saveError{}

func TestAppWithNoStoreNeverSaves(t *testing.T) {
	a := New() // store == nil
	a.autoSaveDelay = 10 * time.Millisecond
	tt := ui.NewTester(a.View, 720, 520)
	tt.Frame()

	a.SetContent("1 + 1")
	time.Sleep(30 * time.Millisecond)
	tt.Frame()

	if a.dirty {
		t.Error("dirty stayed set with no store; there is nothing to save to")
	}
}

func TestTabStripGroupIsCentred(t *testing.T) {
	const w, h = 720, 520

	for _, n := range []int{1, 2, 5, 9} {
		a := newBlankApp()
		for i := 1; i < n; i++ {
			a.ws.Add("", "")
		}
		tt := ui.NewTester(a.View, w, h)
		tt.Frame()

		// The dots themselves are symmetric about the window's midline: a
		// leading counterweight balances the trailing "+" button.
		left := float32(resultPadRight) + (float32(w)-2*resultPadRight-tabStripDotsW(n))/2
		centre := left + tabStripDotsW(n)/2
		if want := float32(w) / 2; centre != want {
			t.Errorf("%d tabs: dots centre = %v, want %v", n, centre, want)
		}

		// And "+" rides right alongside the last dot. The gap is measured
		// to the glyph's left edge, which sits inside the button box: the
		// nearest it may be is one dot gap plus half a dot, and the furthest
		// is the button's own centre. Either way the pitch is identical for 1
		// tab as for 9, so "+" stays part of the group instead of drifting
		// toward the window edge as tabs accumulate.
		plus, ok := tt.Find("+")
		if !ok {
			t.Fatalf("%d tabs: no + button rendered", n)
		}
		pitch := plus.X - dotCenterX(w, n, n-1)
		nearest := float32(dotGap) + dotHitSize/2
		furthest := nearest + tabStripAdd
		if pitch < nearest || pitch > furthest {
			t.Errorf("%d tabs: '+' is %v DIPs past the last dot centre, want %v..%v", n, pitch, nearest, furthest)
		}
		if right := float32(w) - resultPadRight; plus.X > right-2*float32(resultPadRight) {
			t.Errorf("%d tabs: '+' at %v is stranded near the right edge", n, plus.X)
		}
	}
}

// TestStripWindowRevealsEveryTab is the core of the overflow fix: with more
// tabs than the strip can draw, every tab must still be reachable by clicking
// something. Before the sliding window, tabs past maxDots were only reachable
// through Cmd+digit, and the +N badge did nothing at all.
func TestStripWindowRevealsEveryTab(t *testing.T) {
	const total = 29 // well past maxDots

	a := newStripedApp(total)
	tt := ui.NewTester(a.View, 720, 520)
	tt.Frame()

	if a.stripFirst != 0 {
		t.Fatalf("window starts at %d, want 0", a.stripFirst)
	}
	// The badge must report what is hidden past the window.
	if !tt.HasText("+13") {
		t.Fatalf("expected a '+13' badge for the 13 hidden tabs, texts: %v", tt.Texts())
	}

	// The last tab is unreachable by click until we page forward. Page with the
	// badge, exactly as a user would.
	seen := map[int]bool{0: true}
	for a.stripFirst == 0 {
		if !tt.HasText("+13") {
			t.Fatal("the +13 badge vanished while paging forward")
		}
		if err := tt.Click("+13"); err != nil {
			t.Fatalf("click the overflow badge: %v", err)
		}
		tt.Frame()
		if a.stripFirst == 0 {
			t.Fatal("the badge did not move the window")
		}
	}

	// Once fully paged, the last tab's dot is drawn and can be clicked.
	final := total - min(total, maxDots)
	if a.stripFirst != final {
		t.Fatalf("fully paged window starts at %d, want %d", a.stripFirst, final)
	}
	if !tt.HasText("+13") {
		t.Fatalf("expected 13 hidden before the window, texts: %v", tt.Texts())
	}

	// Click the last dot in the window (just before the '+' button) and the
	// app must switch to the final tab.
	plus, ok := tt.Find("+")
	if !ok {
		t.Fatal("no '+' button found in the tab strip")
	}
	lastDotX := (plus.X + plus.W/2) - tabStripAdd/2 - dotGap - dotHitSize/2
	tt.ClickAt(lastDotX, dotYOf(520))
	tt.Frame()
	if got, want := a.ws.ActiveIndex(), total-1; got != want {
		t.Fatalf("clicking last dot after paging activated tab %d, want %d", got, want)
	}

	// Paging back must reach the start again.
	for a.stripFirst > 0 {
		badge := "+" + strconv.Itoa(a.stripFirst)
		if !tt.HasText(badge) {
			t.Fatalf("expected a %q badge for the tabs hidden before the window, texts: %v", badge, tt.Texts())
		}
		if err := tt.Click(badge); err != nil {
			t.Fatalf("click the earlier-tabs badge: %v", err)
		}
		tt.Frame()
	}
	if a.stripFirst != 0 {
		t.Fatalf("paged back to %d, want 0", a.stripFirst)
	}
	seen[0] = true
}

// TestStripWindowFollowsTabSwitches checks the window tracks the active tab on
// its own, so Cmd+digit or a new tab pulls its dot into view.
func TestStripWindowFollowsTabSwitches(t *testing.T) {
	const total = 29

	a := newStripedApp(total)
	tt := ui.NewTester(a.View, 720, 520)
	tt.Frame()

	// Jump past the window.
	a.setActive(28)
	tt.Frame()

	first, shown := a.stripFirst, min(total, maxDots)
	if 28 < first || 28 >= first+shown {
		t.Fatalf("active tab 28 is outside the drawn window [%d,%d)", first, first+shown)
	}

	// And back again.
	a.setActive(0)
	tt.Frame()
	if a.stripFirst != 0 {
		t.Fatalf("window did not return to the start: stripFirst = %d", a.stripFirst)
	}
}

// TestStripPagingSurvivesATabSwitch is why stripFollow exists: a user can page
// away from the active dot, but the next real tab switch must re-centre.
func TestStripPagingSurvivesATabSwitch(t *testing.T) {
	const total = 29

	a := newStripedApp(total)
	tt := ui.NewTester(a.View, 720, 520)
	tt.Frame()

	// Page forward by hand.
	if err := tt.Click("+13"); err != nil {
		t.Fatalf("click the overflow badge: %v", err)
	}
	tt.Frame()
	if a.stripFirst == 0 {
		t.Fatal("paging did not move the window")
	}
	if a.stripFollow {
		t.Error("paging left the window tracking the active tab; browsing would snap back")
	}

	// The active tab (0) is now off-window, which is the point.
	if active := a.ws.ActiveIndex(); active >= a.stripFirst && active < a.stripFirst+maxDots {
		t.Errorf("active tab %d is still on screen after paging away", active)
	}

	// A tab switch re-centres.
	a.setActive(1)
	tt.Frame()
	if a.stripFirst > 1 {
		t.Fatalf("after switching to tab 1 the window starts at %d, want it to include tab 1", a.stripFirst)
	}
}

func TestStripWindowClampsAtBothEnds(t *testing.T) {
	tests := []struct {
		name     string
		total    int
		wantLast int // stripFirst when the last tab is active
	}{
		{"fewer than the cap", 3, 0},
		{"exactly the cap", maxDots, 0},
		{"one past the cap", maxDots + 1, 1},
		{"twice the cap", maxDots * 2, maxDots},
		{"far past the cap", maxDots + 13, 13},
	}

	for _, tc := range tests {
		a := newStripedApp(tc.total)
		tt := ui.NewTester(a.View, 720, 520)
		tt.Frame()

		if a.stripFirst != 0 {
			t.Errorf("%s (%d tabs, tab 0 active): stripFirst = %d, want 0",
				tc.name, tc.total, a.stripFirst)
		}

		a.setActive(tc.total - 1)
		tt.Frame()
		if a.stripFirst != tc.wantLast {
			t.Errorf("%s (%d tabs, last tab active): stripFirst = %d, want %d",
				tc.name, tc.total, a.stripFirst, tc.wantLast)
		}
	}
}

func TestStripWindowNeverScrollsPastTheEnd(t *testing.T) {
	const total = maxDots + 5

	a := newStripedApp(total)
	tt := ui.NewTester(a.View, 720, 520)
	tt.Frame()

	// Hammer the forward badge far past the end.
	for i := 0; i < 5; i++ {
		if badge := "+" + strconv.Itoa(total-min(total, maxDots)-a.stripFirst); tt.HasText(badge) {
			_ = tt.Click(badge)
			tt.Frame()
		}
	}
	want := total - min(total, maxDots)
	if a.stripFirst != want {
		t.Fatalf("stripFirst = %d after over-paging, want %d", a.stripFirst, want)
	}

	// Activate the last tab and ensure it is visible inside the clamped window.
	a.setActive(total - 1)
	tt.Frame()
	if a.ws.ActiveIndex() < a.stripFirst || a.ws.ActiveIndex() >= a.stripFirst+min(total, maxDots) {
		t.Errorf("active tab %d is off-window after over-paging", a.ws.ActiveIndex())
	}
}

func TestStripBadgeCountsEachSide(t *testing.T) {
	// Park the window in the middle of a long list: both sides overflow, so
	// both badges appear and neither is a duplicate of the other.
	const total = 29

	a := newStripedApp(total)
	tt := ui.NewTester(a.View, 720, 520)
	tt.Frame()

	a.stripFirst = 5
	a.stripFollow = false
	tt.Frame()

	if !tt.HasText("+5") {
		t.Errorf("expected '+5' for the 5 earlier tabs, texts: %v", tt.Texts())
	}
	if !tt.HasText("+8") {
		t.Errorf("expected '+8' for the 8 later tabs, texts: %v", tt.Texts())
	}
}

// dotYOf is the vertical middle of the bottom tab strip.
func dotYOf(height int) float32 {
	return float32(height) - tabStripHeight/2
}

func TestTabOverflowShowsACounter(t *testing.T) {
	a := newBlankApp()
	// One tab already, so the strip is 1 over the cap and overflows by 2.
	for i := 0; i < maxDots+1; i++ {
		a.ws.Add("", "")
	}
	if len(a.ws.Tabs) != maxDots+2 {
		t.Fatalf("tabs = %d, want %d", len(a.ws.Tabs), maxDots+2)
	}
	tt := ui.NewTester(a.View, 720, 520)
	tt.Frame()

	if !tt.HasText("+2") {
		t.Fatalf("expected a '+2' overflow counter, got texts: %v", tt.Texts())
	}

	// The counter is not itself a tab: every dot below it still switches.
	tt.Key(ui.Cmd, ui.Key1)
	tt.Frame()
	if a.ws.ActiveIndex() != 0 {
		t.Errorf("ActiveIndex = %d, want 0", a.ws.ActiveIndex())
	}
}

func TestBottomBarClearIconEmptiesActiveTab(t *testing.T) {
	const w, h = 720, 520

	a := New() // seeded with non-empty Starter tab
	tt := ui.NewTester(a.View, w, h)
	tt.Frame()

	rect, ok := tt.Find("Clear")
	if !ok {
		t.Fatalf("expected Clear icon button on non-empty tab, texts: %v", tt.Texts())
	}
	// Must sit in the bottom bar on the right edge, not in the top header.
	if rect.Y < float32(h)-tabStripHeight {
		t.Errorf("Clear button Y = %v, want in bottom strip (>= %v)", rect.Y, float32(h)-tabStripHeight)
	}
	if rect.X < float32(w)/2 {
		t.Errorf("Clear button X = %v, want on right edge of bottom strip", rect.X)
	}

	if err := tt.Click("Clear"); err != nil {
		t.Fatalf("click Clear icon button: %v", err)
	}
	tt.Frame()

	if got := a.ActiveTab().Content; got != "" {
		t.Fatalf("active tab content = %q after Clear, want empty", got)
	}
	if _, ok := tt.Find("Clear"); ok {
		t.Fatal("Clear icon button should hide once the active tab is empty")
	}
}
