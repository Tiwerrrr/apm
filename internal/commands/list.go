package commands

import (
	"fmt"

	"github.com/apm-cli/apm/internal/config"
	"github.com/apm-cli/apm/internal/console"
	"github.com/apm-cli/apm/internal/registry"
)

// List shows all installed packages
func List() error {
	db, err := config.LoadInstalled()
	if err != nil {
		return fmt.Errorf("failed to load installed packages: %w", err)
	}

	if len(db.Packages) == 0 {
		console.Info("No packages installed via APM yet")
		console.Info("Use %sapm install <package>%s to install a package", console.Bold, console.Reset)
		return nil
	}

	fmt.Println()
	console.Step("📋", "Installed packages:\n")

	headers := []string{"Package", "Name", "Version", "Installed"}
	rows := make([][]string, 0, len(db.Packages))
	for id, pkg := range db.Packages {
		installedAt := pkg.InstalledAt
		if len(installedAt) > 10 {
			installedAt = installedAt[:10] // Just the date
		}
		rows = append(rows, []string{id, pkg.DisplayName, pkg.Version, installedAt})
	}

	console.Table(headers, rows)
	fmt.Printf("\n  %s%d package(s) installed%s\n\n", console.Dim, len(db.Packages), console.Reset)

	return nil
}

// ListAll shows all available packages in the registry
func ListAll() error {
	reg, err := registry.Load()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	fmt.Println()
	console.Step("📦", "Available packages:\n")

	results := reg.ListAll()

	headers := []string{"Package", "Description", "Version"}
	rows := make([][]string, len(results))
	for i, r := range results {
		rows[i] = []string{r.ID, r.Package.Description, r.Package.Version}
	}

	console.Table(headers, rows)
	fmt.Printf("\n  %s%d package(s) available%s\n\n", console.Dim, len(results), console.Reset)

	return nil
}
