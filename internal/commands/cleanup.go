package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/apm-cli/apm/internal/config"
	"github.com/apm-cli/apm/internal/console"
)

// Cleanup removes all cached download files
func Cleanup() error {
	cacheDir := config.CacheDir

	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		if os.IsNotExist(err) {
			console.Info("Cache directory is already empty")
			return nil
		}
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	if len(entries) == 0 {
		console.Info("Cache directory is already empty")
		return nil
	}

	var totalSize int64
	var fileCount int

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		totalSize += info.Size()
		fileCount++
	}

	fmt.Println()
	console.Step("🧹", "Found %d cached file(s) (%s)", fileCount, console.FormatBytes(totalSize))

	// Remove all files
	removed := 0
	var freedBytes int64
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, _ := entry.Info()
		path := filepath.Join(cacheDir, entry.Name())
		if err := os.Remove(path); err != nil {
			console.Warning("Failed to remove %s: %v", entry.Name(), err)
		} else {
			removed++
			if info != nil {
				freedBytes += info.Size()
			}
		}
	}

	fmt.Println()
	console.Success("Removed %d file(s), freed %s", removed, console.FormatBytes(freedBytes))
	fmt.Println()

	return nil
}
