package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/apm-cli/apm/internal/config"
	"github.com/apm-cli/apm/internal/console"
	"github.com/apm-cli/apm/internal/installer"
)

// Reinstall removes and re-installs a package
func Reinstall(pkgID string) error {
	if !installer.IsInstalled(pkgID) {
		console.Warning("Package \"%s\" is not currently installed, performing fresh install...", pkgID)
		return Install(pkgID)
	}

	fmt.Println()
	console.Step("🔄", "Reinstalling %s...", console.PackageName(pkgID))

	// 1. Uninstall
	console.Step("🗑", "Removing current installation...")
	if err := installer.Uninstall(pkgID); err != nil {
		console.Warning("Failed to cleanly uninstall: %v (continuing anyway)", err)
	}

	// 2. Clear cached installer for this package
	clearCacheForPackage(pkgID)

	// 3. Reinstall
	if err := Install(pkgID); err != nil {
		return fmt.Errorf("reinstall failed: %w", err)
	}

	return nil
}

// clearCacheForPackage removes cached files for a specific package
func clearCacheForPackage(pkgID string) {
	entries, err := os.ReadDir(config.CacheDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		name := entry.Name()
		// Match files that start with the package ID
		if len(name) >= len(pkgID) && name[:len(pkgID)] == pkgID {
			path := filepath.Join(config.CacheDir, name)
			os.Remove(path)
		}
	}
}
