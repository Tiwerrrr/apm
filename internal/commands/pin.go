package commands

import (
	"fmt"

	"github.com/apm-cli/apm/internal/config"
	"github.com/apm-cli/apm/internal/console"
	"github.com/apm-cli/apm/internal/installer"
)

// Pin marks a package as pinned so upgrade-all skips it
func Pin(pkgID string) error {
	if !installer.IsInstalled(pkgID) {
		return fmt.Errorf("package '%s' is not installed", pkgID)
	}

	db, err := config.LoadInstalled()
	if err != nil {
		return fmt.Errorf("failed to load installed packages: %w", err)
	}

	pkg, ok := db.Packages[pkgID]
	if !ok {
		return fmt.Errorf("package '%s' not found in installed database", pkgID)
	}

	if pkg.Pinned {
		console.Warning("%s is already pinned", console.PackageName(pkgID))
		return nil
	}

	pkg.Pinned = true
	db.Packages[pkgID] = pkg

	if err := config.SaveInstalled(db); err != nil {
		return fmt.Errorf("failed to save: %w", err)
	}

	fmt.Println()
	console.Success("📌 %s pinned at v%s", console.PackageName(pkgID), pkg.Version)
	console.Info("This package will be skipped by %sapm upgrade-all%s", console.Bold, console.Reset)
	fmt.Println()

	return nil
}

// Unpin removes the pinned status from a package
func Unpin(pkgID string) error {
	if !installer.IsInstalled(pkgID) {
		return fmt.Errorf("package '%s' is not installed", pkgID)
	}

	db, err := config.LoadInstalled()
	if err != nil {
		return fmt.Errorf("failed to load installed packages: %w", err)
	}

	pkg, ok := db.Packages[pkgID]
	if !ok {
		return fmt.Errorf("package '%s' not found in installed database", pkgID)
	}

	if !pkg.Pinned {
		console.Warning("%s is not pinned", console.PackageName(pkgID))
		return nil
	}

	pkg.Pinned = false
	db.Packages[pkgID] = pkg

	if err := config.SaveInstalled(db); err != nil {
		return fmt.Errorf("failed to save: %w", err)
	}

	fmt.Println()
	console.Success("🔓 %s unpinned", console.PackageName(pkgID))
	console.Info("This package will be updated by %sapm upgrade-all%s", console.Bold, console.Reset)
	fmt.Println()

	return nil
}
