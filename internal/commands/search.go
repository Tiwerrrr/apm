package commands

import (
	"fmt"

	"github.com/apm-cli/apm/internal/console"
	"github.com/apm-cli/apm/internal/registry"
)

// Search finds packages matching the query and displays results
func Search(query string) error {
	reg, err := registry.Load()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	console.Step("🔍", "Searching for %s\"%s\"%s...\n", console.BrightYellow, query, console.Reset)

	results := reg.Search(query)

	if len(results) == 0 {
		console.Warning("No packages found for \"%s\"", query)
		console.Info("Try a different search term or run %sapm list-all%s to see all available packages", console.Bold, console.Reset)
		return nil
	}

	// Build table
	headers := []string{"Package", "Description", "Version"}
	rows := make([][]string, len(results))
	for i, r := range results {
		rows[i] = []string{r.ID, r.Package.Description, r.Package.Version}
	}

	console.Table(headers, rows)
	fmt.Printf("\n  %s%d package(s) found%s\n", console.Dim, len(results), console.Reset)

	return nil
}
