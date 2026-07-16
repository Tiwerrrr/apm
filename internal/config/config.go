package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// APM directories and file paths
var (
	// RootDir is the APM root directory: %LOCALAPPDATA%\apm
	RootDir string
	// AppsDir is where portable apps are installed
	AppsDir string
	// CacheDir is where downloaded files are cached
	CacheDir string
	// InstalledFile tracks installed packages
	InstalledFile string
	// RegistryFile is the locally cached registry from GitHub
	RegistryFile string
)

const (
	// Version is the current APM version
	Version = "1.0.8"
)

func init() {
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		localAppData = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local")
	}

	RootDir = filepath.Join(localAppData, "apm")
	AppsDir = filepath.Join(RootDir, "apps")
	CacheDir = filepath.Join(RootDir, "cache")
	InstalledFile = filepath.Join(RootDir, "installed.json")
	RegistryFile = filepath.Join(RootDir, "registry.json")
}

// EnsureDirs creates all necessary APM directories
func EnsureDirs() error {
	dirs := []string{RootDir, AppsDir, CacheDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// Cleanup old exe files from self-update
	oldExe := filepath.Join(RootDir, "bin", "apm.exe.old")
	if _, err := os.Stat(oldExe); err == nil {
		os.Remove(oldExe)
	}

	return nil
}

// InstalledPackage represents a package that has been installed
type InstalledPackage struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Version     string `json:"version"`
	Type        string `json:"type"` // "installer" or "portable"
	InstallPath string `json:"install_path,omitempty"`
	InstalledAt string `json:"installed_at"`
	Pinned      bool   `json:"pinned,omitempty"`
}

// InstalledDB is the database of installed packages
type InstalledDB struct {
	Packages map[string]InstalledPackage `json:"packages"`
}

// LoadInstalled loads the installed packages database
func LoadInstalled() (*InstalledDB, error) {
	db := &InstalledDB{
		Packages: make(map[string]InstalledPackage),
	}

	data, err := os.ReadFile(InstalledFile)
	if err != nil {
		if os.IsNotExist(err) {
			return db, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, db); err != nil {
		return nil, err
	}

	return db, nil
}

// SaveInstalled saves the installed packages database
func SaveInstalled(db *InstalledDB) error {
	if err := EnsureDirs(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(InstalledFile, data, 0644)
}
