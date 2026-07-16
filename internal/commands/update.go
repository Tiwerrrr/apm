package commands

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/apm-cli/apm/internal/config"
	"github.com/apm-cli/apm/internal/console"
)

const registryURL = "https://raw.githubusercontent.com/Tiwerrrr/apm/main/data/registry.json"

// Update fetches the latest registry from GitHub and saves it locally
func Update() error {
	console.Step("🔄", "Updating package registry from GitHub...")

	if err := config.EnsureDirs(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	resp, err := http.Get(registryURL)
	if err != nil {
		return fmt.Errorf("failed to download registry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub returned status: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read registry data: %w", err)
	}

	// Write to local cache
	if err := os.WriteFile(config.RegistryFile, data, 0644); err != nil {
		return fmt.Errorf("failed to save registry: %w", err)
	}

	// Fetch custom registries
	repos, err := LoadRepos()
	if err == nil && len(repos) > 0 {
		console.Step("🔄", "Updating custom repositories...")
		for name, rURL := range repos {
			cResp, err := http.Get(rURL)
			if err != nil {
				console.Warning("Failed to update repo '%s': %v", name, err)
				continue
			}
			if cResp.StatusCode != http.StatusOK {
				cResp.Body.Close()
				console.Warning("Repo '%s' returned status: %s", name, cResp.Status)
				continue
			}
			cData, err := io.ReadAll(cResp.Body)
			cResp.Body.Close()
			if err != nil {
				console.Warning("Failed to read repo '%s': %v", name, err)
				continue
			}
			cPath := filepath.Join(config.ReposDir, name+".json")
			if err := os.WriteFile(cPath, cData, 0644); err != nil {
				console.Warning("Failed to save repo '%s': %v", name, err)
				continue
			}
		}
	}

	console.Success("Registry updated successfully! (%s)", console.FormatBytes(int64(len(data))))
	return nil
}
