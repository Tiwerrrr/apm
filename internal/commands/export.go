package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/apm-cli/apm/internal/config"
	"github.com/apm-cli/apm/internal/console"
)

// Export saves the list of installed packages to a text file
func Export(filePath string) error {
	db, err := config.LoadInstalled()
	if err != nil {
		return fmt.Errorf("failed to load installed packages: %w", err)
	}

	if len(db.Packages) == 0 {
		console.Warning("No packages installed via APM to export.")
		return nil
	}

	var packages []string
	for pkgID := range db.Packages {
		packages = append(packages, pkgID)
	}

	content := strings.Join(packages, "\n") + "\n"

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write export file: %w", err)
	}

	console.Success("Exported %d packages to %s", len(packages), filePath)
	return nil
}
