package commands

import (
	"fmt"

	"github.com/apm-cli/apm/internal/config"
	"github.com/apm-cli/apm/internal/console"
	"github.com/apm-cli/apm/internal/installer"
	"github.com/apm-cli/apm/internal/registry"
)

// UpgradeAll checks for updates for all installed packages and installs them
func UpgradeAll() error {
	db, err := config.LoadInstalled()
	if err != nil {
		return fmt.Errorf("failed to load installed packages: %w", err)
	}

	if len(db.Packages) == 0 {
		console.Info("No packages installed to upgrade.")
		return nil
	}

	reg, err := registry.Load()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	console.Step("🔄", "Checking for updates for %d packages...", len(db.Packages))

	updatesAvailable := []string{}

	for pkgID, installedPkg := range db.Packages {
		regPkg := reg.Get(pkgID)
		if regPkg == nil {
			continue // Package no longer in registry
		}

		latestVersion := regPkg.Version

		// If it's a GitHub repo and version is "latest", fetch the real latest version tag
		if regPkg.GithubRepo != "" {
			_, ghVersion, _, err := installer.FetchLatestGitHubAsset(regPkg.GithubRepo, regPkg.AssetRegex)
			if err == nil && ghVersion != "" {
				latestVersion = ghVersion
			}
		}

		if latestVersion != installedPkg.Version && latestVersion != "latest" {
			console.Info("Update available for %s: %s -> %s", console.PackageName(pkgID), installedPkg.Version, latestVersion)
			updatesAvailable = append(updatesAvailable, pkgID)
		}
	}

	if len(updatesAvailable) == 0 {
		console.Success("All packages are up to date!")
		return nil
	}

	fmt.Println()
	console.Step("🚀", "Starting upgrade of %d packages...", len(updatesAvailable))

	for _, pkgID := range updatesAvailable {
		fmt.Println()
		// Remove first to avoid "already installed" error
		if err := installer.Uninstall(pkgID); err != nil {
			console.Warning("Failed to uninstall %s during upgrade: %v", pkgID, err)
		}

		if err := Install(pkgID); err != nil {
			console.Error("Failed to upgrade %s: %v", pkgID, err)
		} else {
			console.Success("Successfully upgraded %s", pkgID)
		}
	}

	console.Success("Upgrade complete!")
	return nil
}
