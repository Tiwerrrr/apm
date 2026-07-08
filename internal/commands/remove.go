package commands

import (
	"fmt"

	"github.com/apm-cli/apm/internal/console"
	"github.com/apm-cli/apm/internal/installer"
)

// Remove uninstalls a package
func Remove(pkgID string) error {
	if !installer.IsInstalled(pkgID) {
		console.Error("Package \"%s\" is not installed via APM", pkgID)
		console.Info("Use %sapm list%s to see installed packages", console.Bold, console.Reset)
		return nil
	}

	fmt.Println()
	if err := installer.Uninstall(pkgID); err != nil {
		return fmt.Errorf("removal failed: %w", err)
	}

	fmt.Println()
	console.Success("%s removed successfully!", console.PackageName(pkgID))
	fmt.Println()

	return nil
}
