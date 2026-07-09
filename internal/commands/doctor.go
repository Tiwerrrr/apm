package commands

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/apm-cli/apm/internal/config"
	"github.com/apm-cli/apm/internal/console"
)

// Doctor runs diagnostics on the APM installation
func Doctor() error {
	fmt.Println()
	console.Step("🩺", "Running APM diagnostics...\n")

	allOk := true

	// 1. Check APM version
	printCheck(true, "APM version: v%s", config.Version)

	// 2. Check APM directories
	dirs := map[string]string{
		"Root":  config.RootDir,
		"Apps":  config.AppsDir,
		"Cache": config.CacheDir,
	}
	for name, dir := range dirs {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			printCheck(true, "%s directory exists: %s", name, dir)
		} else {
			printCheck(false, "%s directory missing: %s", name, dir)
			allOk = false
		}
	}

	// 3. Check APM in PATH
	apmPath, err := exec.LookPath("apm")
	if err == nil {
		printCheck(true, "APM found in PATH: %s", apmPath)
	} else {
		printCheck(false, "APM not found in PATH")
		console.Info("  Add %s to your PATH", filepath.Join(config.RootDir, "bin"))
		allOk = false
	}

	// 4. Check installed.json
	db, err := config.LoadInstalled()
	if err != nil {
		printCheck(false, "installed.json is corrupted: %v", err)
		allOk = false
	} else {
		printCheck(true, "installed.json is valid (%d packages tracked)", len(db.Packages))
	}

	// 5. Check GitHub API access
	ghOk := checkGitHub()
	if ghOk {
		printCheck(true, "GitHub API is accessible")
	} else {
		printCheck(false, "GitHub API is unreachable (updates won't work)")
		allOk = false
	}

	// 6. Check for orphaned portable installs
	if db != nil {
		orphaned := 0
		for pkgID, pkg := range db.Packages {
			if pkg.Type == "portable" && pkg.InstallPath != "" {
				if _, err := os.Stat(pkg.InstallPath); os.IsNotExist(err) {
					orphaned++
					printCheck(false, "Orphaned install: %s (path missing: %s)", pkgID, pkg.InstallPath)
				}
			}
		}
		if orphaned == 0 {
			printCheck(true, "No orphaned portable installations")
		} else {
			allOk = false
		}
	}

	// 7. Check cache size
	cacheSize, cacheFiles := getCacheStats()
	if cacheFiles > 0 {
		printCheck(true, "Cache: %d file(s), %s (run %sapm cleanup%s to free space)",
			cacheFiles, console.FormatBytes(cacheSize), console.Bold, console.Reset)
	} else {
		printCheck(true, "Cache is empty")
	}

	fmt.Println()
	if allOk {
		console.Success("All checks passed! APM is healthy.")
	} else {
		console.Warning("Some issues were found. See above for details.")
	}
	fmt.Println()

	return nil
}

func printCheck(ok bool, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if ok {
		fmt.Printf("  %s%s✓%s %s\n", console.BrightGreen, console.Bold, console.Reset, msg)
	} else {
		fmt.Printf("  %s%s✗%s %s\n", console.BrightRed, console.Bold, console.Reset, msg)
	}
}

func checkGitHub() bool {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://api.github.com/rate_limit")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func getCacheStats() (int64, int) {
	entries, err := os.ReadDir(config.CacheDir)
	if err != nil {
		return 0, 0
	}

	var totalSize int64
	var count int
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		totalSize += info.Size()
		count++
	}
	return totalSize, count
}

// CheckPATH checks if a directory is in the system PATH
func isInPATH(dir string) bool {
	pathEnv := os.Getenv("PATH")
	paths := strings.Split(pathEnv, string(os.PathListSeparator))
	for _, p := range paths {
		if strings.EqualFold(filepath.Clean(p), filepath.Clean(dir)) {
			return true
		}
	}
	return false
}
