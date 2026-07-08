package commands

import (
	"fmt"
	"io"
	"net/http"
	"os"

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

	console.Success("Registry updated successfully! (%s)", console.FormatBytes(int64(len(data))))
	return nil
}
