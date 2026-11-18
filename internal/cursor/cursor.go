package cursor

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// FileConfig is the persisted cursor position for one file.
type FileConfig struct {
	CursorLine int `json:"cursor_line"`
	CursorCol  int `json:"cursor_col"`
}

// PositionStore persists cursor positions in a per-user cache directory.
type PositionStore struct {
	cacheDir string
}

// NewPositionStore creates a store under ~/.cache/md-cli, falling back to
// ~/.md-cli-cache when the XDG directory cannot be created.
func NewPositionStore() (*PositionStore, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(home, ".cache", "md-cli")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		dir = filepath.Join(home, ".md-cli-cache")
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return nil, err
		}
	}
	return &PositionStore{cacheDir: dir}, nil
}

// key returns a stable, collision-free file name for path. Hashing avoids the
// collisions and traversal hazards of encoding paths into file names.
func key(path string) string {
	sum := sha256.Sum256([]byte(filepath.Clean(path)))
	return hex.EncodeToString(sum[:])
}

func (ps *PositionStore) configPath(filePath string) string {
	return filepath.Join(ps.cacheDir, key(filePath)+".json")
}

// GetPosition returns the saved cursor position for filePath.
func (ps *PositionStore) GetPosition(filePath string) (FileConfig, bool) {
	data, err := os.ReadFile(ps.configPath(filePath))
	if err != nil {
		return FileConfig{}, false
	}
	var cfg FileConfig
	if err := json.Unmarshal(data, &cfg); err != nil || cfg.CursorLine < 0 || cfg.CursorCol < 0 {
		return FileConfig{}, false
	}
	return cfg, true
}

// SetPosition writes cfg atomically so a crash cannot truncate the cache file.
func (ps *PositionStore) SetPosition(filePath string, cfg FileConfig) error {
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	return writeAtomic(ps.configPath(filePath), data)
}

// RemovePosition deletes the saved position for filePath.
func (ps *PositionStore) RemovePosition(filePath string) error {
	err := os.Remove(ps.configPath(filePath))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func writeAtomic(path string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(name, path)
}
