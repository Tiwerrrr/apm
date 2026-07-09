package commands

import (
	"fmt"

	"github.com/apm-cli/apm/internal/config"
	"github.com/apm-cli/apm/internal/console"
	"github.com/apm-cli/apm/internal/installer"
	"github.com/apm-cli/apm/internal/registry"
)

// Outdated shows installed packages that have newer versions available
func Outdated() error {
	db, err := config.LoadInstalled()
	if err != nil {
		return fmt.Errorf("failed to load installed packages: %w", err)
	}

	if len(db.Packages) == 0 {
		console.Info("No packages installed via APM yet")
		return nil
	}

	reg, err := registry.Load()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	fmt.Println()
	console.Step("🔍", "Checking for updates for %d packages...\n", len(db.Packages))

	type outdatedPkg struct {
		id        string
		name      string
		installed string
		available string
		pinned    bool
	}

	var outdatedList []outdatedPkg

	for pkgID, installedPkg := range db.Packages {
		regPkg := reg.Get(pkgID)
		if regPkg == nil {
			continue // Package no longer in registry
		}

		latestVersion := regPkg.Version

		// If it's a GitHub repo, fetch the real latest version tag
		if regPkg.GithubRepo != "" {
			_, ghVersion, _, err := installer.FetchLatestGitHubAsset(regPkg.GithubRepo, regPkg.AssetRegex)
			if err == nil && ghVersion != "" {
				latestVersion = ghVersion
			}
		}

		if latestVersion != installedPkg.Version && latestVersion != "latest" {
			outdatedList = append(outdatedList, outdatedPkg{
				id:        pkgID,
				name:      installedPkg.DisplayName,
				installed: installedPkg.Version,
				available: latestVersion,
				pinned:    installedPkg.Pinned,
			})
		}
	}

	if len(outdatedList) == 0 {
		console.Success("All packages are up to date!")
		fmt.Println()
		return nil
	}

	// Build table
	headers := []string{"Package", "Installed", "Available", "Status"}
	rows := make([][]string, len(outdatedList))
	for i, pkg := range outdatedList {
		status := "update available"
		if pkg.pinned {
			status = "📌 pinned"
		}
		rows[i] = []string{pkg.id, pkg.installed, pkg.available, status}
	}

	console.Table(headers, rows)

	pinnedCount := 0
	for _, pkg := range outdatedList {
		if pkg.pinned {
			pinnedCount++
		}
	}

	fmt.Printf("\n  %s%d package(s) can be updated%s", console.Dim, len(outdatedList)-pinnedCount, console.Reset)
	if pinnedCount > 0 {
		fmt.Printf("  %s(%d pinned)%s", console.BrightYellow, pinnedCount, console.Reset)
	}
	fmt.Println()
	console.Info("Run %sapm upgrade-all%s to update all packages", console.Bold, console.Reset)
	fmt.Println()

	return nil
}
