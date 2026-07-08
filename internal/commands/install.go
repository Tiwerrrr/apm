package commands

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/apm-cli/apm/internal/config"
	"github.com/apm-cli/apm/internal/console"
	"github.com/apm-cli/apm/internal/downloader"
	"github.com/apm-cli/apm/internal/installer"
	"github.com/apm-cli/apm/internal/registry"
)

// Install downloads and installs a package
func Install(pkgID string) error {
	// Ensure APM directories exist
	if err := config.EnsureDirs(); err != nil {
		return fmt.Errorf("failed to create APM directories: %w", err)
	}

	// Load registry
	reg, err := registry.Load()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	// Find the package
	pkg := reg.Get(pkgID)
	if pkg == nil {
		// Try fuzzy search
		results := reg.Search(pkgID)
		if len(results) > 0 {
			console.Error("Package \"%s\" not found", pkgID)
			console.Info("Did you mean one of these?")
			fmt.Println()
			for _, r := range results {
				fmt.Printf("  • %s  %s%s%s\n", console.PackageName(r.ID), console.Dim, r.Package.Description, console.Reset)
			}
			fmt.Println()
			console.Info("Use: %sapm install %s%s", console.Bold, results[0].ID, console.Reset)
			return nil
		}
		return fmt.Errorf("package '%s' not found in registry", pkgID)
	}

	// Check if already installed
	if installer.IsInstalled(pkgID) {
		console.Warning("%s is already installed", console.PackageName(pkg.Name))
		console.Info("Use %sapm remove %s%s first to reinstall", console.Bold, pkgID, console.Reset)
		return nil
	}

	// Display what we're installing
	fmt.Println()
	console.Step("📦", "Installing %s (%s %s)...\n",
		console.PackageName(pkg.Name),
		pkgID,
		console.VersionStr("v"+pkg.Version),
	)

	// Determine filename from URL
	filename := guessFilename(pkg.URL, pkgID)

	// Download
	filePath, err := downloader.Download(pkg.URL, config.CacheDir, filename)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	// Install
	if err := installer.Install(pkg, pkgID, filePath); err != nil {
		return fmt.Errorf("installation failed: %w", err)
	}

	fmt.Println()
	console.Success("%s installed successfully!", console.PackageName(pkg.Name))
	fmt.Println()

	return nil
}

// guessFilename determines the installer filename from URL and package ID
func guessFilename(url string, pkgID string) string {
	// Try to extract filename from URL
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		last := parts[len(parts)-1]
		// Remove query parameters
		if idx := strings.Index(last, "?"); idx >= 0 {
			last = last[:idx]
		}
		// Check if it has a valid extension
		ext := strings.ToLower(filepath.Ext(last))
		if ext == ".exe" || ext == ".msi" || ext == ".zip" {
			return last
		}
	}

	// Fallback: use package ID
	if strings.Contains(strings.ToLower(url), ".msi") {
		return pkgID + "-setup.msi"
	}
	return pkgID + "-setup.exe"
}
