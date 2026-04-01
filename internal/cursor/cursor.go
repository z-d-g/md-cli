package cursor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type FileConfig struct {
	CursorLine int `json:"cursor_line"`
	CursorCol  int `json:"cursor_col"`
}

type PositionStore struct {
	cacheDir string
}

func NewPositionStore() (*PositionStore, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	cacheDir := filepath.Join(homeDir, ".cache", "md-cli")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		cacheDir = filepath.Join(homeDir, ".md-cli-cache")
		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			return nil, err
		}
	}

	return &PositionStore{cacheDir: cacheDir}, nil
}

func (ps *PositionStore) configPath(filePath string) string {
	return filepath.Join(ps.cacheDir, sanitize(filePath)+".json")
}

func sanitize(path string) string {
	r := strings.NewReplacer("/", "_", "\\", "_", ":", "_", "|", "_", "?", "_", "*", "_", "\"", "_", "<", "_", ">", "_")
	path = r.Replace(path)
	if len(path) > 0 && path[0] == '.' {
		path = "_" + path[1:]
	}
	return path
}

func (ps *PositionStore) GetPosition(filePath string) (FileConfig, bool) {
	data, err := os.ReadFile(ps.configPath(filePath))
	if err != nil {
		return FileConfig{}, false
	}
	var cfg FileConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return FileConfig{}, false
	}
	return cfg, true
}

func (ps *PositionStore) SetPosition(filePath string, cfg FileConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ps.configPath(filePath), data, 0644)
}

func (ps *PositionStore) RemovePosition(filePath string) error {
	return os.Remove(ps.configPath(filePath))
}
