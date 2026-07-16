package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/apm-cli/apm/internal/commands"
	"github.com/apm-cli/apm/internal/config"
	"github.com/apm-cli/apm/internal/console"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	command := strings.ToLower(os.Args[1])

	switch command {
	case "install", "i":
		if len(os.Args) < 3 {
			console.Error("Missing package name")
			console.Info("Usage: %sapm install <package> [package2] [package3]...%s", console.Bold, console.Reset)
			os.Exit(1)
		}
		// Multi-install: download in parallel, install sequentially
		packages := os.Args[2:]
		if err := commands.InstallMultiple(packages); err != nil {
			console.Error("%v", err)
			os.Exit(1)
		}

	case "remove", "uninstall", "rm":
		if len(os.Args) < 3 {
			console.Error("Missing package name")
			console.Info("Usage: %sapm remove <package>%s", console.Bold, console.Reset)
			os.Exit(1)
		}
		pkgID := strings.ToLower(os.Args[2])
		if err := commands.Remove(pkgID); err != nil {
			console.Error("%v", err)
			os.Exit(1)
		}

	case "reinstall", "ri":
		if len(os.Args) < 3 {
			console.Error("Missing package name")
			console.Info("Usage: %sapm reinstall <package>%s", console.Bold, console.Reset)
			os.Exit(1)
		}
		pkgID := strings.ToLower(os.Args[2])
		if err := commands.Reinstall(pkgID); err != nil {
			console.Error("%v", err)
			os.Exit(1)
		}

	case "search", "find", "s":
		if len(os.Args) < 3 {
			console.Error("Missing search query")
			console.Info("Usage: %sapm search <query>%s", console.Bold, console.Reset)
			os.Exit(1)
		}
		query := strings.Join(os.Args[2:], " ")
		if err := commands.Search(query); err != nil {
			console.Error("%v", err)
			os.Exit(1)
		}

	case "info":
		if len(os.Args) < 3 {
			console.Error("Missing package name")
			console.Info("Usage: %sapm info <package>%s", console.Bold, console.Reset)
			os.Exit(1)
		}
		pkgID := strings.ToLower(os.Args[2])
		if err := commands.Info(pkgID); err != nil {
			console.Error("%v", err)
			os.Exit(1)
		}

	case "list", "ls":
		if err := commands.List(); err != nil {
			console.Error("%v", err)
			os.Exit(1)
		}

	case "list-all", "available":
		if err := commands.ListAll(); err != nil {
			console.Error("%v", err)
			os.Exit(1)
		}

	case "outdated":
		if err := commands.Outdated(); err != nil {
			console.Error("%v", err)
			os.Exit(1)
		}

	case "update", "u":
		if err := commands.Update(); err != nil {
			console.Error("%v", err)
			os.Exit(1)
		}

	case "upgrade":
		if err := commands.UpgradeSelf(); err != nil {
			console.Error("%v", err)
			os.Exit(1)
		}

	case "selfdestruct":
		if err := commands.SelfDestruct(); err != nil {
			console.Error("%v", err)
			os.Exit(1)
		}

	case "upgrade-all":
		if err := commands.UpgradeAll(); err != nil {
			console.Error("%v", err)
			os.Exit(1)
		}

	case "export":
		if len(os.Args) < 3 {
			console.Error("Usage: apm export <file.txt>")
			os.Exit(1)
		}
		if err := commands.Export(os.Args[2]); err != nil {
			console.Error("%v", err)
			os.Exit(1)
		}

	case "import":
		if len(os.Args) < 3 {
			console.Error("Usage: apm import <file.txt>")
			os.Exit(1)
		}
		if err := commands.Import(os.Args[2]); err != nil {
			console.Error("%v", err)
			os.Exit(1)
		}

	case "cleanup", "cache":
		if err := commands.Cleanup(); err != nil {
			console.Error("%v", err)
			os.Exit(1)
		}

	case "doctor":
		if err := commands.Doctor(); err != nil {
			console.Error("%v", err)
			os.Exit(1)
		}

	case "pin":
		if len(os.Args) < 3 {
			console.Error("Missing package name")
			console.Info("Usage: %sapm pin <package>%s", console.Bold, console.Reset)
			os.Exit(1)
		}
		pkgID := strings.ToLower(os.Args[2])
		if err := commands.Pin(pkgID); err != nil {
			console.Error("%v", err)
			os.Exit(1)
		}

	case "unpin":
		if len(os.Args) < 3 {
			console.Error("Missing package name")
			console.Info("Usage: %sapm unpin <package>%s", console.Bold, console.Reset)
			os.Exit(1)
		}
		pkgID := strings.ToLower(os.Args[2])
		if err := commands.Unpin(pkgID); err != nil {
			console.Error("%v", err)
			os.Exit(1)
		}

	case "version", "v", "--version", "-v":
		fmt.Printf("APM (Awesome Package Manager) v%s\n", config.Version)

	case "repo":
		if len(os.Args) < 3 {
			console.Error("Missing repo subcommand")
			console.Info("Usage:")
			console.Info("  apm repo add <name> <url>")
			console.Info("  apm repo remove <name>")
			console.Info("  apm repo list")
			os.Exit(1)
		}
		subCmd := strings.ToLower(os.Args[2])
		switch subCmd {
		case "add":
			if len(os.Args) < 5 {
				console.Error("Missing repo name or URL")
				console.Info("Usage: apm repo add <name> <url>")
				os.Exit(1)
			}
			name := os.Args[3]
			url := os.Args[4]
			if err := commands.RepoAdd(name, url); err != nil {
				console.Error("%v", err)
				os.Exit(1)
			}
		case "remove", "rm":
			if len(os.Args) < 4 {
				console.Error("Missing repo name")
				console.Info("Usage: apm repo remove <name>")
				os.Exit(1)
			}
			name := os.Args[3]
			if err := commands.RepoRemove(name); err != nil {
				console.Error("%v", err)
				os.Exit(1)
			}
		case "list", "ls":
			if err := commands.RepoList(); err != nil {
				console.Error("%v", err)
				os.Exit(1)
			}
		default:
			console.Error("Unknown repo subcommand: %s", subCmd)
			os.Exit(1)
		}

	case "help", "h", "--help", "-h":
		printUsage()

	default:
		console.Error("Unknown command: %s", command)
		fmt.Println()
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	console.Logo()
	fmt.Printf("  %s%sAPM v%s%s — Awesome Package Manager for Windows\n\n", console.Bold, console.BrightWhite, config.Version, console.Reset)
	fmt.Printf("  %s%sUSAGE:%s\n", console.Bold, console.BrightYellow, console.Reset)
	fmt.Printf("    apm <command> [arguments]\n\n")
	fmt.Printf("  %s%sCOMMANDS:%s\n", console.Bold, console.BrightYellow, console.Reset)
	fmt.Printf("    %s%sinstall%s   <pkg> [pkg2]...  Install one or more packages\n", console.Bold, console.BrightCyan, console.Reset)
	fmt.Printf("    %s%sremove%s    <package>        Remove an installed package\n", console.Bold, console.BrightCyan, console.Reset)
	fmt.Printf("    %s%sreinstall%s <package>        Reinstall a package (remove + install)\n", console.Bold, console.BrightCyan, console.Reset)
	fmt.Printf("    %s%sinfo%s      <package>        Show detailed package information\n", console.Bold, console.BrightCyan, console.Reset)
	fmt.Printf("    %s%ssearch%s    <query>          Search for packages\n", console.Bold, console.BrightCyan, console.Reset)
	fmt.Printf("    %s%slist%s                       Show installed packages\n", console.Bold, console.BrightCyan, console.Reset)
	fmt.Printf("    %s%slist-all%s                   Show all available packages\n", console.Bold, console.BrightCyan, console.Reset)
	fmt.Printf("    %s%soutdated%s                   Show packages with available updates\n", console.Bold, console.BrightCyan, console.Reset)
	fmt.Printf("    %s%supdate%s                     Update the package registry from GitHub\n", console.Bold, console.BrightCyan, console.Reset)
	fmt.Printf("    %s%supgrade%s                    Upgrade APM itself to the latest release\n", console.Bold, console.BrightCyan, console.Reset)
	fmt.Printf("    %s%supgrade-all%s                Upgrade all installed packages\n", console.Bold, console.BrightCyan, console.Reset)
	fmt.Printf("    %s%spin%s       <package>        Pin a package version (skip in upgrade-all)\n", console.Bold, console.BrightCyan, console.Reset)
	fmt.Printf("    %s%sunpin%s     <package>        Unpin a package version\n", console.Bold, console.BrightCyan, console.Reset)
	fmt.Printf("    %s%scleanup%s                    Clear the download cache\n", console.Bold, console.BrightCyan, console.Reset)
	fmt.Printf("    %s%sdoctor%s                     Run diagnostics on APM installation\n", console.Bold, console.BrightCyan, console.Reset)
	fmt.Printf("    %s%sexport%s    <file>           Export installed packages list\n", console.Bold, console.BrightCyan, console.Reset)
	fmt.Printf("    %s%simport%s    <file>           Install packages from a list\n", console.Bold, console.BrightCyan, console.Reset)
	fmt.Printf("    %s%sselfdestruct%s               Permanently uninstall APM and its data\n", console.Bold, console.BrightCyan, console.Reset)
	fmt.Printf("    %s%sversion%s                    Show APM version\n", console.Bold, console.BrightCyan, console.Reset)
	fmt.Printf("    %s%shelp%s                       Show this help message\n", console.Bold, console.BrightCyan, console.Reset)
	fmt.Println()
	fmt.Printf("  %s%sALIASES:%s\n", console.Bold, console.BrightYellow, console.Reset)
	fmt.Printf("    %si%s → install,  %srm%s → remove,  %sri%s → reinstall,  %su%s → update,  %ss%s → search,  %sls%s → list\n",
		console.BrightCyan, console.Reset,
		console.BrightCyan, console.Reset,
		console.BrightCyan, console.Reset,
		console.BrightCyan, console.Reset,
		console.BrightCyan, console.Reset,
		console.BrightCyan, console.Reset,
	)
	fmt.Println()
	fmt.Printf("  %s%sEXAMPLES:%s\n", console.Bold, console.BrightYellow, console.Reset)
	fmt.Printf("    apm install firefox obs-studio   %s# Install multiple packages%s\n", console.Dim, console.Reset)
	fmt.Printf("    apm info telegram                %s# Show package details%s\n", console.Dim, console.Reset)
	fmt.Printf("    apm search browser               %s# Find browser packages%s\n", console.Dim, console.Reset)
	fmt.Printf("    apm outdated                     %s# Check for updates%s\n", console.Dim, console.Reset)
	fmt.Printf("    apm pin obs-studio               %s# Lock OBS version%s\n", console.Dim, console.Reset)
	fmt.Printf("    apm doctor                       %s# Run diagnostics%s\n", console.Dim, console.Reset)
	fmt.Println()
}
