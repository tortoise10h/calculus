package workspace

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// newStore returns a FileStore rooted in a temp dir, plus that dir.
func newStore(t *testing.T) (FileStore, string) {
	t.Helper()
	dir := t.TempDir()
	return NewFileStore(dir), dir
}

func TestFileStoreRoundTripsEveryTabField(t *testing.T) {
	s, _ := newStore(t)

	want := Seed()
	want.Add("Groceries", "1 + 1")
	want.Add("Rent", "2 + 2")
	want.Rename(0, "Renamed")
	want.Tabs[1].ScrollY = 123.5
	want.Tabs[2].ScrollY = 42
	want.SetActive(2)

	if err := s.Save(want); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := s.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(got.Tabs) != 3 {
		t.Fatalf("tabs = %d, want 3", len(got.Tabs))
	}
	for i, tab := range got.Tabs {
		w := want.Tabs[i]
		if tab.ID != w.ID {
			t.Errorf("tab %d ID = %q, want %q", i, tab.ID, w.ID)
		}
		if tab.Title != w.Title {
			t.Errorf("tab %d title = %q, want %q", i, tab.Title, w.Title)
		}
		if tab.Content != w.Content {
			t.Errorf("tab %d content = %q, want %q", i, tab.Content, w.Content)
		}
		if tab.ScrollY != w.ScrollY {
			t.Errorf("tab %d scrollY = %v, want %v", i, tab.ScrollY, w.ScrollY)
		}
	}
	if got.ActiveIndex() != 2 {
		t.Errorf("ActiveIndex = %d, want 2", got.ActiveIndex())
	}
	if got.Version != StoreVersion {
		t.Errorf("version = %d, want %d", got.Version, StoreVersion)
	}
}

func TestSaveIsAtomicAndLeavesOnlyOneFile(t *testing.T) {
	s, dir := newStore(t)

	for i := 0; i < 3; i++ {
		if err := s.Save(Seed()); err != nil {
			t.Fatalf("Save %d: %v", i, err)
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 1 {
		names := make([]string, len(entries))
		for i, e := range entries {
			names[i] = e.Name()
		}
		t.Fatalf("directory holds %v, want just the workspace file", names)
	}
	if entries[0].Name() != FileName {
		t.Errorf("file = %q, want %q", entries[0].Name(), FileName)
	}

	// An existing file keeps its 0600 mode across the rename.
	info, err := os.Stat(s.Path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("mode = %o, want 600", perm)
	}
}

func TestSaveCreatesMissingDirectories(t *testing.T) {
	root := t.TempDir()
	s := NewFileStore(filepath.Join(root, "nested", "deeper"))

	if err := s.Save(Seed()); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if _, err := os.Stat(s.Path); err != nil {
		t.Fatalf("Stat after Save: %v", err)
	}
}

func TestSaveRejectsNilWorkspace(t *testing.T) {
	s, _ := newStore(t)
	if err := s.Save(nil); err == nil {
		t.Fatal("Save(nil) returned nil error")
	}
}

func TestLoadMissingFileReportsErrNoWorkspace(t *testing.T) {
	s, _ := newStore(t)

	w, err := s.Load()
	if !errors.Is(err, ErrNoWorkspace) {
		t.Fatalf("err = %v, want ErrNoWorkspace", err)
	}
	if w != nil {
		t.Errorf("workspace = %+v, want nil", w)
	}
}

// writeRaw drops arbitrary bytes at the store path.
func writeRaw(t *testing.T, s FileStore, data string) {
	t.Helper()
	if err := os.WriteFile(s.Path, []byte(data), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

func TestLoadCorruptJSONReportsAnErrorWithoutPanicking(t *testing.T) {
	s, _ := newStore(t)
	writeRaw(t, s, "{not json at all")

	w, err := s.Load()
	if err == nil {
		t.Fatal("Load of a corrupt file returned no error")
	}
	if w != nil {
		t.Errorf("workspace = %+v, want nil for undecodable JSON", w)
	}
}

func TestLoadRepairsEmptyTabsAndUnknownActiveID(t *testing.T) {
	s, _ := newStore(t)

	// No tabs at all.
	writeRaw(t, s, `{"version":1,"activeId":"x","tabs":[]}`)
	w, err := s.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(w.Tabs) != 1 || w.Tabs[0].Title != "Starter" {
		t.Fatalf("empty-tabs workspace not seeded: %+v", w.Tabs)
	}
	if w.ActiveIndex() != 0 {
		t.Errorf("ActiveIndex = %d, want 0", w.ActiveIndex())
	}

	// Unknown activeId, and a tab with no ID of its own.
	writeRaw(t, s, `{"version":1,"activeId":"ghost","tabs":[{"title":"One"},{"title":"Two","id":"two"}]}`)
	w, err = s.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got := w.Active().Title; got != "One" {
		t.Errorf("Active = %q, want %q", got, "One")
	}
	if w.Tabs[0].ID == "" {
		t.Error("tab with an empty ID was not repaired")
	}
	if w.Tabs[1].ID != "two" {
		t.Errorf("existing ID = %q, want %q", w.Tabs[1].ID, "two")
	}
}

func TestLoadKeepsANewerVersionAndReportsIt(t *testing.T) {
	s, _ := newStore(t)
	writeRaw(t, s, `{"version":99,"activeId":"a","tabs":[{"id":"a","title":"From the future"}]}`)

	w, err := s.Load()
	if err == nil {
		t.Fatal("Load of a newer version returned no error")
	}
	if !strings.Contains(err.Error(), "newer") {
		t.Errorf("err = %v, want it to mention a newer version", err)
	}
	if w == nil {
		t.Fatal("Load returned no workspace for a newer version")
	}
	if got := w.Tabs[0].Title; got != "From the future" {
		t.Errorf("title = %q, want %q", got, "From the future")
	}
	if w.Version != 99 {
		t.Errorf("version = %d, want the file's 99", w.Version)
	}
}

func TestLoadRejectsOversizedFiles(t *testing.T) {
	s, _ := newStore(t)

	f, err := os.Create(s.Path)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	// Write one byte past the limit.
	if _, err := f.Write(make([]byte, maxStoreBytes+1)); err != nil {
		t.Fatalf("Write: %v", err)
	}
	f.Close()

	w, err := s.Load()
	if err == nil {
		t.Fatal("Load of an oversized file returned no error")
	}
	if w != nil {
		t.Errorf("workspace = %+v, want nil", w)
	}
}

func TestLoadAcceptsTheVersionWrittenBySave(t *testing.T) {
	s, _ := newStore(t)
	if err := s.Save(Seed()); err != nil {
		t.Fatalf("Save: %v", err)
	}

	raw, err := os.ReadFile(s.Path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	var doc struct {
		Version  int    `json:"version"`
		ActiveID string `json:"activeId"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if doc.Version != StoreVersion {
		t.Errorf("file version = %d, want %d", doc.Version, StoreVersion)
	}
	if doc.ActiveID == "" {
		t.Error("file has an empty activeId")
	}
}

func TestLoadOrSeed(t *testing.T) {
	t.Run("missing file seeds silently", func(t *testing.T) {
		s, _ := newStore(t)
		w, err := LoadOrSeed(s)
		if err != nil {
			t.Fatalf("err = %v, want nil on a first run", err)
		}
		if w.Tabs[0].Title != "Starter" {
			t.Errorf("title = %q, want %q", w.Tabs[0].Title, "Starter")
		}
	})

	t.Run("existing file loads", func(t *testing.T) {
		s, _ := newStore(t)
		save := Seed()
		save.Tabs[0].Title = "Restored"
		if err := s.Save(save); err != nil {
			t.Fatalf("Save: %v", err)
		}

		w, err := LoadOrSeed(s)
		if err != nil {
			t.Fatalf("err = %v", err)
		}
		if w.Tabs[0].Title != "Restored" {
			t.Errorf("title = %q, want %q", w.Tabs[0].Title, "Restored")
		}
	})

	t.Run("corrupt file seeds and reports", func(t *testing.T) {
		s, _ := newStore(t)
		writeRaw(t, s, "}{ broken")

		w, err := LoadOrSeed(s)
		if err == nil {
			t.Fatal("err = nil, want the corruption reported")
		}
		if w == nil {
			t.Fatal("workspace = nil; a corrupt file must still yield a usable workspace")
		}
		if w.Tabs[0].Title != "Starter" {
			t.Errorf("title = %q, want the seeded %q", w.Tabs[0].Title, "Starter")
		}

		// And the next save overwrites the corrupt file.
		if err := s.Save(w); err != nil {
			t.Fatalf("Save over a corrupt file: %v", err)
		}
		if _, err := s.Load(); err != nil {
			t.Fatalf("Load after overwrite: %v", err)
		}
	})

	t.Run("nil store seeds", func(t *testing.T) {
		w, err := LoadOrSeed(nil)
		if err != nil {
			t.Fatalf("err = %v", err)
		}
		if w.Tabs[0].Title != "Starter" {
			t.Errorf("title = %q, want %q", w.Tabs[0].Title, "Starter")
		}
	})
}

// TestSessionSurvivesQuitAndRelaunch walks the lifecycle the app depends on:
// seed a first run, work in it, save on quit, relaunch and get the session back
// exactly — then corrupt the file and confirm the next launch still starts
// clean instead of crashing.
func TestSessionSurvivesQuitAndRelaunch(t *testing.T) {
	store := NewFileStore(t.TempDir())

	// First launch: no file yet, so it seeds.
	ws, err := LoadOrSeed(store)
	if err != nil {
		t.Fatalf("first launch: %v", err)
	}
	if ws.Tabs[0].Title != "Starter" {
		t.Fatalf("first launch title = %q, want %q", ws.Tabs[0].Title, "Starter")
	}

	// A session: rename, edit, scroll, add a tab, switch.
	ws.Rename(0, "Groceries")
	ws.Tabs[0].Content = "drink = 10 + 5\ndrink * 2"
	ws.Tabs[0].ScrollY = 64.25
	ws.Add("Rent", "900 + 100")
	ws.SetActive(1)

	// Quit: one atomic save.
	if err := store.Save(ws); err != nil {
		t.Fatalf("save on quit: %v", err)
	}

	// Relaunch.
	got, err := LoadOrSeed(store)
	if err != nil {
		t.Fatalf("relaunch: %v", err)
	}
	if len(got.Tabs) != 2 {
		t.Fatalf("relaunch tabs = %d, want 2", len(got.Tabs))
	}
	if got.Tabs[0].Title != "Groceries" || got.Tabs[0].Content != "drink = 10 + 5\ndrink * 2" {
		t.Errorf("tab 0 not restored: %+v", got.Tabs[0])
	}
	if got.Tabs[0].ScrollY != 64.25 {
		t.Errorf("tab 0 scrollY = %v, want 64.25", got.Tabs[0].ScrollY)
	}
	if got.Tabs[1].Title != "Rent" || got.Tabs[1].Content != "900 + 100" {
		t.Errorf("tab 1 not restored: %+v", got.Tabs[1])
	}
	if got.ActiveIndex() != 1 {
		t.Errorf("relaunch active = %d, want 1", got.ActiveIndex())
	}

	// A corrupt file must not block the next launch.
	if err := os.WriteFile(store.Path, []byte("{{{ not json"), 0o600); err != nil {
		t.Fatalf("corrupt: %v", err)
	}
	recovered, err := LoadOrSeed(store)
	if err == nil {
		t.Error("corruption was not reported")
	}
	if recovered.Tabs[0].Title != "Starter" {
		t.Errorf("after corruption title = %q, want the seeded %q", recovered.Tabs[0].Title, "Starter")
	}
}
