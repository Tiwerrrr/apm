package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/apm-cli/apm/internal/commands"
	"github.com/apm-cli/apm/internal/console"
)

const version = "1.0.0"

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
			console.Info("Usage: %sapm install <package>%s", console.Bold, console.Reset)
			os.Exit(1)
		}
		pkgID := strings.ToLower(os.Args[2])
		if err := commands.Install(pkgID); err != nil {
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

	case "update", "u":
		if err := commands.Update(); err != nil {
			console.Error("%v", err)
			os.Exit(1)
		}

	case "version", "v", "--version", "-v":
		fmt.Printf("APM (Awesome Package Manager) v%s\n", version)

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
	fmt.Printf("  %s%sAPM v%s%s — Awesome Package Manager for Windows\n\n", console.Bold, console.BrightWhite, version, console.Reset)
	fmt.Printf("  %s%sUSAGE:%s\n", console.Bold, console.BrightYellow, console.Reset)
	fmt.Printf("    apm <command> [arguments]\n\n")
	fmt.Printf("  %s%sCOMMANDS:%s\n", console.Bold, console.BrightYellow, console.Reset)
	fmt.Printf("    %s%sinstall%s <package>    Install a package\n", console.Bold, console.BrightCyan, console.Reset)
	fmt.Printf("    %s%sremove%s  <package>    Remove an installed package\n", console.Bold, console.BrightCyan, console.Reset)
	fmt.Printf("    %s%supdate%s               Update the package registry from GitHub\n", console.Bold, console.BrightCyan, console.Reset)
	fmt.Printf("    %s%ssearch%s  <query>      Search for packages\n", console.Bold, console.BrightCyan, console.Reset)
	fmt.Printf("    %s%slist%s                 Show installed packages\n", console.Bold, console.BrightCyan, console.Reset)
	fmt.Printf("    %s%slist-all%s             Show all available packages\n", console.Bold, console.BrightCyan, console.Reset)
	fmt.Printf("    %s%sversion%s              Show APM version\n", console.Bold, console.BrightCyan, console.Reset)
	fmt.Printf("    %s%shelp%s                 Show this help message\n", console.Bold, console.BrightCyan, console.Reset)
	fmt.Println()
	fmt.Printf("  %s%sALIASES:%s\n", console.Bold, console.BrightYellow, console.Reset)
	fmt.Printf("    %si%s → install,  %srm%s → remove,  %su%s → update,  %ss%s → search,  %sls%s → list\n",
		console.BrightCyan, console.Reset,
		console.BrightCyan, console.Reset,
		console.BrightCyan, console.Reset,
		console.BrightCyan, console.Reset,
		console.BrightCyan, console.Reset,
	)
	fmt.Println()
	fmt.Printf("  %s%sEXAMPLES:%s\n", console.Bold, console.BrightYellow, console.Reset)
	fmt.Printf("    apm search obs           %s# Find packages related to OBS%s\n", console.Dim, console.Reset)
	fmt.Printf("    apm install obs-studio   %s# Install OBS Studio%s\n", console.Dim, console.Reset)
	fmt.Printf("    apm remove obs-studio    %s# Remove OBS Studio%s\n", console.Dim, console.Reset)
	fmt.Printf("    apm search google        %s# Find Google packages%s\n", console.Dim, console.Reset)
	fmt.Println()
}
