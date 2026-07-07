// Package config loads and persists the Pathguard YAML config, applying
// defaults for any missing keys. See ADR-0003 for the schema.
package config

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const appDir = "pathguard" // identifier (ADR-0011)

type Watch struct {
	Root         string `yaml:"root"`
	RemotePrefix string `yaml:"remote_prefix"`
}

type Limits struct {
	NameMax   int     `yaml:"name_max"`
	PathMax   int     `yaml:"path_max"`
	WarnRatio float64 `yaml:"warn_ratio"`
}

type Schedule struct {
	Interval string   `yaml:"interval"`
	At       []string `yaml:"at"`
}

type Thresholds struct {
	Yellow int `yaml:"yellow"`
	Red    int `yaml:"red"`
	Warn   int `yaml:"warn"`
}

type Notify struct {
	TrayWarnIcon bool       `yaml:"tray_warn_icon"`
	NativeBanner bool       `yaml:"native_banner"`
	Thresholds   Thresholds `yaml:"thresholds"`
}

type UI struct {
	Language string `yaml:"language"` // auto | en | ko
}

type Menu struct {
	MaxInline int `yaml:"max_inline"`
}

type Config struct {
	Watch    []Watch  `yaml:"watch"`
	Limits   Limits   `yaml:"limits"`
	Schedule Schedule `yaml:"schedule"`
	Notify   Notify   `yaml:"notify"`
	Exclude  []string `yaml:"exclude"`
	UI       UI       `yaml:"ui"`
	Menu     Menu     `yaml:"menu"`
}

// DefaultExclude are noise/tool/server-generated entries skipped by default (ADR-0008).
var DefaultExclude = []string{
	".git", "node_modules",
	"@eaDir", "#recycle", "#snapshot", // Synology
	".DS_Store", ".Trashes", ".Spotlight-V100", ".fseventsd", // macOS
	"$RECYCLE.BIN", "System Volume Information", // Windows
}

// Default returns the built-in configuration.
func Default() Config {
	return Config{
		Watch:    []Watch{{Root: "~/Documents", RemotePrefix: "/volume1/homes/johndoe/MyDocuments"}},
		Limits:   Limits{NameMax: 255, PathMax: 4096, WarnRatio: 0.80},
		Schedule: Schedule{Interval: "6h", At: []string{}},
		Notify: Notify{
			TrayWarnIcon: true,
			NativeBanner: false,
			Thresholds:   Thresholds{Yellow: 1, Red: 10, Warn: 1},
		},
		Exclude: append([]string(nil), DefaultExclude...),
		UI:      UI{Language: "auto"},
		Menu:    Menu{MaxInline: 10},
	}
}

// Path is the OS-conventional config file location (ADR-0003):
//
//	macOS:   ~/Library/Application Support/pathguard/config.yml
//	Windows: %AppData%\pathguard\config.yml
//	Linux:   ~/.config/pathguard/config.yml
func Path() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, appDir, "config.yml"), nil
}

// Load reads the config file, filling missing keys with defaults. If the file
// does not exist it is created with the defaults. Returns the config and the
// path it was loaded from/created at.
func Load() (Config, string, error) {
	cfg := Default()
	path, err := Path()
	if err != nil {
		return cfg, "", err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return cfg, path, Save(cfg, path)
	}
	if err != nil {
		return cfg, path, err
	}
	// Unmarshal onto the defaults so absent keys keep their default values.
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, path, err
	}
	return cfg, path, nil
}

// LoadFile reads a specific config file onto the defaults (no auto-create).
// Missing keys keep their default values.
func LoadFile(path string) (Config, error) {
	cfg := Default()
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// Save writes cfg to path (creating parent dirs).
func Save(cfg Config, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// ExcludeSet turns the exclude list into a lookup set.
func (c Config) ExcludeSet() map[string]bool {
	m := make(map[string]bool, len(c.Exclude))
	for _, e := range c.Exclude {
		if e = strings.TrimSpace(e); e != "" {
			m[e] = true
		}
	}
	return m
}

// ExpandRoot resolves a leading ~ in a watch root to the user's home dir.
func ExpandRoot(root string) string {
	if root == "~" || strings.HasPrefix(root, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(strings.TrimPrefix(root, "~"), "/"))
		}
	}
	return root
}
