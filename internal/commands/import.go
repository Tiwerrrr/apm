package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/apm-cli/apm/internal/console"
	"github.com/apm-cli/apm/internal/installer"
)

// Import reads a list of packages from a file and installs them
func Import(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read import file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	var packages []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			packages = append(packages, line)
		}
	}

	if len(packages) == 0 {
		console.Warning("No packages found in file.")
		return nil
	}

	console.Info("Found %d packages in %s", len(packages), filePath)
	fmt.Println()

	successCount := 0
	for _, pkgID := range packages {
		if installer.IsInstalled(pkgID) {
			console.Success("✅ %s (already installed)", console.PackageName(pkgID))
			successCount++
			continue
		}

		if err := Install(pkgID); err != nil {
			console.Error("Failed to install %s: %v", pkgID, err)
		} else {
			successCount++
		}
	}

	fmt.Println()
	console.Info("Import complete: %d/%d installed successfully.", successCount, len(packages))
	return nil
}
