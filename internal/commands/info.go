package commands

import (
	"fmt"
	"strings"

	"github.com/apm-cli/apm/internal/config"
	"github.com/apm-cli/apm/internal/console"
	"github.com/apm-cli/apm/internal/installer"
	"github.com/apm-cli/apm/internal/registry"
)

// Info displays detailed information about a package
func Info(pkgID string) error {
	reg, err := registry.Load()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	pkg := reg.Get(pkgID)
	if pkg == nil {
		// Try fuzzy search
		results := reg.Search(pkgID)
		if len(results) > 0 {
			console.Error("Package \"%s\" not found", pkgID)
			console.Info("Did you mean one of these?")
			fmt.Println()
			for _, r := range results {
				if len(results) > 5 {
					break
				}
				fmt.Printf("  • %s  %s%s%s\n", console.PackageName(r.ID), console.Dim, r.Package.Description, console.Reset)
			}
			return nil
		}
		return fmt.Errorf("package '%s' not found in registry", pkgID)
	}

	fmt.Println()

	// Package name and type
	typeBadge := fmt.Sprintf("%s%s %s %s", console.Bold, console.BgCyan, strings.ToUpper(pkg.Type), console.Reset)
	fmt.Printf("  %s  %s\n", console.PackageName(pkg.Name), typeBadge)

	// Description
	fmt.Printf("  %s%s%s\n\n", console.Dim, pkg.Description, console.Reset)

	// Details table
	printField := func(label, value string) {
		if value != "" {
			fmt.Printf("  %s%-14s%s %s\n", console.BrightYellow, label, console.Reset, value)
		}
	}

	printField("ID:", pkgID)
	printField("Version:", pkg.Version)
	printField("Type:", pkg.Type)

	if pkg.Homepage != "" {
		printField("Homepage:", pkg.Homepage)
	}

	if pkg.GithubRepo != "" {
		printField("GitHub:", fmt.Sprintf("https://github.com/%s", pkg.GithubRepo))
	}

	if pkg.URL != "" {
		url := pkg.URL
		if len(url) > 70 {
			url = url[:67] + "..."
		}
		printField("Download:", url)
	}

	if pkg.SilentArgs != "" {
		printField("Silent args:", pkg.SilentArgs)
	}

	if len(pkg.Dependencies) > 0 {
		printField("Dependencies:", strings.Join(pkg.Dependencies, ", "))
	}

	if pkg.Bin != "" {
		printField("Binary:", pkg.Bin)
	}

	// Tags
	if len(pkg.Tags) > 0 {
		tags := make([]string, len(pkg.Tags))
		for i, t := range pkg.Tags {
			tags[i] = fmt.Sprintf("%s%s#%s%s", console.Dim, console.BrightBlue, t, console.Reset)
		}
		fmt.Printf("  %s%-14s%s %s\n", console.BrightYellow, "Tags:", console.Reset, strings.Join(tags, "  "))
	}

	fmt.Println()

	// Installation status
	if installer.IsInstalled(pkgID) {
		db, _ := config.LoadInstalled()
		if db != nil {
			if installed, ok := db.Packages[pkgID]; ok {
				fmt.Printf("  %s%s ✓ Installed%s  v%s  (%s)\n", console.BrightGreen, console.Bold, console.Reset, installed.Version, installed.InstalledAt[:10])
				if installed.Pinned {
					fmt.Printf("  %s%s 📌 Pinned%s — this package will be skipped by upgrade-all\n", console.BrightYellow, console.Bold, console.Reset)
				}
			}
		}
	} else {
		fmt.Printf("  %s%s ○ Not installed%s\n", console.Dim, console.Bold, console.Reset)
		fmt.Printf("  %sInstall with: %sapm install %s%s\n", console.Dim, console.Bold, pkgID, console.Reset)
	}

	fmt.Println()

	return nil
}
