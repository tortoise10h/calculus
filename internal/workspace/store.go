package workspace

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// maxStoreBytes caps how large a workspace file may be before Load refuses it.
// 4 MB is far more than any real session and keeps a corrupt or hostile file
// from being slurped into memory wholesale.
const maxStoreBytes = 4 << 20

// FileName is the workspace file inside the user-data directory.
const FileName = "workspace.json"

// ErrNoWorkspace is returned by Store.Load when the file does not exist yet —
// the normal first-run case, not a failure.
var ErrNoWorkspace = errors.New("workspace: no saved workspace")

// Store reads and writes a Workspace. The UI layer depends on the interface, not
// on the filesystem, so app tests can inject a fake and never touch the user's
// real data directory.
type Store interface {
	Load() (*Workspace, error)
	Save(w *Workspace) error
}

// FileStore keeps the workspace as indented JSON at Path.
type FileStore struct {
	Path string
}

// NewFileStore returns a FileStore writing to workspace.json inside dir.
func NewFileStore(dir string) FileStore {
	return FileStore{Path: filepath.Join(dir, FileName)}
}

// Load reads and decodes the workspace.
//
// It returns ErrNoWorkspace when the file is absent. Anything else — an
// oversized, corrupt or unreadable file — is reported as an error, but Load
// still returns whatever it managed to decode so a caller can choose to use it.
// A file written by a newer build (version > StoreVersion) is kept as-is and
// reported through a wrapped error, so the app can log it without discarding
// data it does not understand.
func (s FileStore) Load() (*Workspace, error) {
	info, err := os.Stat(s.Path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrNoWorkspace
		}
		return nil, err
	}
	if info.Size() > maxStoreBytes {
		return nil, fmt.Errorf("workspace: %s is %d bytes, over the %d byte limit", s.Path, info.Size(), maxStoreBytes)
	}

	data, err := os.ReadFile(s.Path)
	if err != nil {
		return nil, err
	}

	var w Workspace
	if err := json.Unmarshal(data, &w); err != nil {
		return nil, fmt.Errorf("workspace: decode %s: %w", s.Path, err)
	}
	w.repair()

	if w.Version > StoreVersion {
		return &w, fmt.Errorf("workspace: %s has version %d, newer than %d; loaded as-is",
			s.Path, w.Version, StoreVersion)
	}
	if w.Version == 0 {
		w.Version = StoreVersion
	}
	return &w, nil
}

// Save writes the workspace atomically: marshal, write a sibling .tmp, fsync
// it, then rename it over the target. A crash therefore leaves either the old
// file or the new one, never a half-written mix, and never a stray .tmp behind
// in the steady state.
func (s FileStore) Save(w *Workspace) error {
	if w == nil {
		return errors.New("workspace: cannot save a nil workspace")
	}
	if w.Version == 0 {
		w.Version = StoreVersion
	}

	data, err := json.MarshalIndent(w, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	dir := filepath.Dir(s.Path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, FileName+".*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // No-op once the rename below succeeded.

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	// CreateTemp already uses 0600; Chmod makes the mode explicit and survives
	// a different umask on the way in.
	if err := os.Chmod(tmpName, 0o600); err != nil {
		return err
	}
	return os.Rename(tmpName, s.Path)
}

// LoadOrSeed loads the workspace, or returns Seed() when the file is missing or
// unreadable. A non-nil error can come back together with a usable workspace —
// callers log it and carry on, so a corrupt file never blocks startup.
func LoadOrSeed(s Store) (*Workspace, error) {
	if s == nil {
		return Seed(), nil
	}
	w, err := s.Load()
	switch {
	case err == nil && w != nil:
		return w, nil
	case errors.Is(err, ErrNoWorkspace):
		return Seed(), nil
	default:
		return Seed(), err
	}
}
