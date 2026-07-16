package commands

import (
	"fmt"
	"path/filepath"
	"strings"

	"sync"

	"github.com/apm-cli/apm/internal/config"
	"github.com/apm-cli/apm/internal/console"
	"github.com/apm-cli/apm/internal/downloader"
	"github.com/apm-cli/apm/internal/installer"
	"github.com/apm-cli/apm/internal/registry"
)

// InstallMultiple resolves all dependencies, downloads in parallel, and installs sequentially
func InstallMultiple(pkgIDs []string) error {
	// Ensure APM directories exist
	if err := config.EnsureDirs(); err != nil {
		return fmt.Errorf("failed to create APM directories: %w", err)
	}

	// Load registry
	reg, err := registry.Load()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	var toInstall []string
	visited := make(map[string]bool)

	// Helper to resolve dependencies
	var resolve func(id string) error
	resolve = func(id string) error {
		if visited[id] {
			return nil
		}
		visited[id] = true

		if installer.IsInstalled(id) {
			return nil
		}

		pkg := reg.Get(id)
		if pkg == nil {
			// Try fuzzy search for the first missing package
			results := reg.Search(id)
			if len(results) > 0 {
				console.Error("Package \"%s\" not found", id)
				console.Info("Did you mean one of these?")
				fmt.Println()
				for _, r := range results {
					fmt.Printf("  • %s  %s%s%s\n", console.PackageName(r.ID), console.Dim, r.Package.Description, console.Reset)
				}
				fmt.Println()
				return fmt.Errorf("package '%s' not found", id)
			}
			return fmt.Errorf("package '%s' not found in registry", id)
		}

		// Resolve dependencies first
		for _, dep := range pkg.Dependencies {
			if err := resolve(dep); err != nil {
				return err
			}
		}

		toInstall = append(toInstall, id)
		return nil
	}

	for _, id := range pkgIDs {
		if err := resolve(id); err != nil {
			return err
		}
	}

	if len(toInstall) == 0 {
		console.Info("All requested packages are already installed.")
		return nil
	}

	console.Step("📦", "Resolved %d package(s) to install.", len(toInstall))

	// Parallel Download Phase
	if len(toInstall) > 1 {
		console.Step("⬇", "Downloading %d packages in parallel...", len(toInstall))
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(toInstall))
	dlMap := make(map[string]string)
	var mu sync.Mutex

	for _, id := range toInstall {
		wg.Add(1)
		go func(pkgID string) {
			defer wg.Done()
			
			pkg := reg.Get(pkgID)
			
			// Dynamic GitHub Fetch
			if pkg.GithubRepo != "" {
				ghUrl, ghVersion, _, err := installer.FetchLatestGitHubAsset(pkg.GithubRepo, pkg.AssetRegex)
				if err != nil {
					errCh <- fmt.Errorf("failed to fetch latest GitHub release for %s: %w", pkgID, err)
					return
				}
				pkg.URL = ghUrl
				pkg.Version = ghVersion
			}

			filename := guessFilename(pkg.URL, pkgID)
			
			// Download (quiet if multiple packages)
			quiet := len(toInstall) > 1
			filePath, err := downloader.Download(pkg.URL, config.CacheDir, filename, pkg.Hash, quiet)
			if err != nil {
				errCh <- fmt.Errorf("failed to download %s: %w", pkgID, err)
				return
			}

			mu.Lock()
			dlMap[pkgID] = filePath
			mu.Unlock()

			if quiet {
				fmt.Printf("  • %sDownloaded %s%s\n", console.BrightGreen, pkgID, console.Reset)
			}
		}(id)
	}

	wg.Wait()
	close(errCh)

	if len(errCh) > 0 {
		return <-errCh // Return first error
	}

	if len(toInstall) > 1 {
		console.Success("All downloads completed.")
		fmt.Println()
	}

	// Sequential Install Phase
	for idx, pkgID := range toInstall {
		pkg := reg.Get(pkgID)
		filePath := dlMap[pkgID]

		if len(toInstall) > 1 {
			fmt.Printf("%s%s [%d/%d] Installing %s%s\n", console.Bold, console.BrightCyan, idx+1, len(toInstall), pkgID, console.Reset)
		} else {
			fmt.Println()
		}

		console.Step("📦", "Installing %s (%s %s)...",
			console.PackageName(pkg.Name),
			pkgID,
			console.VersionStr("v"+pkg.Version),
		)

		if err := installer.Install(pkg, pkgID, filePath); err != nil {
			return fmt.Errorf("installation failed for %s: %w", pkgID, err)
		}

		console.Success("%s installed successfully!", console.PackageName(pkg.Name))
		fmt.Println()
	}

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
