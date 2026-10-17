// Package workspace holds the multi-tab session state and its persistence.
//
// It must not import mygo or mygo/ui: the UI layer builds on top of it, and
// this package stays unit-testable with plain `go test`.
package workspace

import (
	"crypto/rand"
	"encoding/hex"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

// StoreVersion is the schema version this build writes. Files with a higher
// version are still loaded — see Store.Load — but a lower one is never
// downgraded into.
const StoreVersion = 1

// starterDoc is the document a first run (or a repaired/unreadable file)
// starts from: the drink-cost example from the style guide.
const starterDoc = `# Calculus — free-style math, one result per line

20 + 5 - 12
(8 + 2) / 2

# Variables are assigned with = and read on later lines
drink_cost = 10 + 5 + 20
meal_cost = 40
total_cost = drink_cost + meal_cost`

// Tab is one calculation buffer: its text, the variables it defines
// (implicitly, through the document order) and the view state that belongs to
// it.
type Tab struct {
	ID      string  `json:"id"`      // Stable identity; ActiveID points at one.
	Title   string  `json:"title"`   // Shown in the header bar, and as the dot tooltip.
	Content string  `json:"content"` // The one true editor buffer.
	ScrollY float32 `json:"scrollY"`
	// Caret/selection is deliberately not persisted; see the phase plan's
	// Known limitations.
}

// Workspace is the whole session: every tab and which one is active.
type Workspace struct {
	Version  int    `json:"version"`
	ActiveID string `json:"activeId"`
	Tabs     []*Tab `json:"tabs"`
}

// New returns a workspace with a single empty tab named "Untitled".
func New() *Workspace {
	w := &Workspace{Version: StoreVersion}
	w.Tabs = []*Tab{{ID: NewID(), Title: "Untitled"}}
	w.ActiveID = w.Tabs[0].ID
	return w
}

// Seed returns the first-run workspace: a single "Starter" tab holding
// starterDoc.
func Seed() *Workspace {
	w := &Workspace{Version: StoreVersion}
	w.Tabs = []*Tab{{ID: NewID(), Title: "Starter", Content: starterDoc}}
	w.ActiveID = w.Tabs[0].ID
	return w
}

// idFallback counts up when crypto/rand is unavailable, so NewID still returns
// distinct values within one process.
var idFallback atomic.Uint64

// NewID returns 8 hex characters of random data. If the system random source
// fails it falls back to a counter mixed with the clock — still unique enough
// for tab identity, and never an empty string.
func NewID() string {
	var b [4]byte
	if _, err := rand.Read(b[:]); err == nil {
		return hex.EncodeToString(b[:])
	}
	n := idFallback.Add(1)
	return hex.EncodeToString([]byte{
		byte(n >> 24), byte(n >> 16), byte(n >> 8), byte(n),
	}) + strconv.FormatInt(time.Now().UnixNano()&0xffff, 16)
}

// Active returns the tab ActiveID names, falling back to Tabs[0], then nil when
// there are no tabs at all.
func (w *Workspace) Active() *Tab {
	if w == nil {
		return nil
	}
	for _, t := range w.Tabs {
		if t != nil && t.ID != "" && t.ID == w.ActiveID {
			return t
		}
	}
	if len(w.Tabs) > 0 && w.Tabs[0] != nil {
		return w.Tabs[0]
	}
	return nil
}

// ActiveIndex returns the position of the active tab, or -1 when there is none.
func (w *Workspace) ActiveIndex() int {
	if w == nil || len(w.Tabs) == 0 {
		return -1
	}
	for i, t := range w.Tabs {
		if t != nil && t.ID != "" && t.ID == w.ActiveID {
			return i
		}
	}
	return 0
}

// SetActive makes tab i active. The index is clamped to the valid range and an
// out-of-range or nil workspace is a no-op.
func (w *Workspace) SetActive(i int) {
	if w == nil || len(w.Tabs) == 0 {
		return
	}
	if i < 0 {
		i = 0
	}
	if i >= len(w.Tabs) {
		i = len(w.Tabs) - 1
	}
	if t := w.Tabs[i]; t != nil {
		w.ActiveID = t.ID
	}
}

// nextUntitledTitle returns the first unused "Untitled" / "Untitled N" title.
func (w *Workspace) nextUntitledTitle() string {
	const base = "Untitled"
	if w.titleTaken(base) {
		for n := 2; ; n++ {
			candidate := base + " " + strconv.Itoa(n)
			if !w.titleTaken(candidate) {
				return candidate
			}
		}
	}
	return base
}

func (w *Workspace) titleTaken(title string) bool {
	for _, t := range w.Tabs {
		if t != nil && t.Title == title {
			return true
		}
	}
	return false
}

// Add appends a tab with the given title and content and makes it active.
// An empty title falls back to the first unused "Untitled" name. It returns the
// new tab, or nil when the workspace is nil.
func (w *Workspace) Add(title, content string) *Tab {
	if w == nil {
		return nil
	}
	if w.Version == 0 {
		w.Version = StoreVersion
	}
	title = strings.TrimSpace(title)
	if title == "" {
		title = w.nextUntitledTitle()
	}
	t := &Tab{ID: NewID(), Title: title, Content: content}
	w.Tabs = append(w.Tabs, t)
	w.ActiveID = t.ID
	return t
}

// Close removes tab i and fixes ActiveID. It returns false — changing nothing
// — when i is out of range or the workspace holds only one tab: the last tab is
// never closable. When the removed tab was active, the new active tab is
// min(i, len-1).
func (w *Workspace) Close(i int) bool {
	if w == nil || len(w.Tabs) <= 1 || i < 0 || i >= len(w.Tabs) {
		return false
	}
	wasActive := w.Tabs[i] != nil && w.Tabs[i].ID == w.ActiveID

	w.Tabs = append(w.Tabs[:i:i], w.Tabs[i+1:]...)

	if wasActive {
		if i >= len(w.Tabs) {
			i = len(w.Tabs) - 1
		}
		if t := w.Tabs[i]; t != nil {
			w.ActiveID = t.ID
		}
	}
	return true
}

// Rename sets the title of tab i after trimming surrounding whitespace. An
// empty or whitespace-only title is ignored, keeping the previous one, as is
// any out-of-range index.
func (w *Workspace) Rename(i int, title string) {
	if w == nil || i < 0 || i >= len(w.Tabs) || w.Tabs[i] == nil {
		return
	}
	if title = strings.TrimSpace(title); title == "" {
		return
	}
	w.Tabs[i].Title = title
}

// Titles returns the title of every tab, in order.
func (w *Workspace) Titles() []string {
	if w == nil {
		return nil
	}
	out := make([]string, len(w.Tabs))
	for i, t := range w.Tabs {
		if t != nil {
			out[i] = t.Title
		}
	}
	return out
}

// repair fills in whatever the decoder left missing: tab IDs, the active
// pointer and, when there are no usable tabs at all, the seeded starter tab. It
// is called by Store.Load and never fails.
func (w *Workspace) repair() {
	for _, t := range w.Tabs {
		if t != nil && t.ID == "" {
			t.ID = NewID()
		}
	}
	if len(w.Tabs) == 0 || w.Tabs[0] == nil {
		seeded := Seed()
		w.Tabs = seeded.Tabs
		w.ActiveID = seeded.ActiveID
		return
	}
	// An unknown or empty ActiveID falls back to the first tab.
	for _, t := range w.Tabs {
		if t != nil && t.ID != "" && t.ID == w.ActiveID {
			return
		}
	}
	w.ActiveID = w.Tabs[0].ID
}
