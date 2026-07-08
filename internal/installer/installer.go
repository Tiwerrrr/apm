package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/apm-cli/apm/internal/config"
	"github.com/apm-cli/apm/internal/console"
	"github.com/apm-cli/apm/internal/registry"
)

func Install(pkg *registry.Package, pkgID string, filePath string) error {
	ext := strings.ToLower(filepath.Ext(filePath))

	// Handle Portable ZIPs
	if ext == ".zip" {
		console.Step("📦", "Extracting portable package...")
		appsDir := filepath.Join(config.RootDir, "apps")
		installDir := filepath.Join(appsDir, pkgID)
		
		// Remove previous install dir if exists
		os.RemoveAll(installDir)
		
		if err := extractZip(filePath, installDir); err != nil {
			return fmt.Errorf("failed to extract zip: %w", err)
		}
		
		// Create Shim if bin is specified
		if pkg.Bin != "" {
			exePath := filepath.Join(installDir, pkg.Bin)
			binDir := filepath.Join(config.RootDir, "bin")
			if err := createShim(binDir, exePath, pkgID); err != nil {
				console.Warning("Failed to create shim: %v", err)
			}
		}

		// Record the installation
		if err := recordInstall(pkgID, pkg, installDir); err != nil {
			console.Warning("Package installed but failed to record: %v", err)
		}
		return nil
	}

	console.Step("🔧", "Running installer (silent mode)...")

	var cmd *exec.Cmd

	switch ext {
	case ".msi":
		// MSI packages use msiexec
		args := []string{"/i", filePath}
		if pkg.SilentArgs != "" {
			args = append(args, splitArgs(pkg.SilentArgs)...)
		}
		cmd = exec.Command("msiexec.exe", args...)

	case ".exe":
		// EXE packages use the silent args directly
		args := splitArgs(pkg.SilentArgs)
		cmd = exec.Command(filePath, args...)

	default:
		return fmt.Errorf("unsupported installer format: %s", ext)
	}

	// Run the installer
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		errStr := err.Error()
		// Проверяем, требует ли установщик прав администратора (exit code 740)
		if strings.Contains(strings.ToLower(errStr), "requires elevation") || strings.Contains(errStr, "740") || strings.Contains(strings.ToLower(errStr), "повышения прав") {
			console.Warning("Установщик требует прав Администратора. Подтвердите запрос UAC...")
			
			var psCmd string
			if ext == ".msi" {
				psCmd = fmt.Sprintf(`Start-Process -FilePath "msiexec.exe" -ArgumentList '/i "%s" %s' -Verb RunAs -Wait`, filePath, pkg.SilentArgs)
			} else {
				psCmd = fmt.Sprintf(`Start-Process -FilePath "%s" -ArgumentList '%s' -Verb RunAs -Wait`, filePath, pkg.SilentArgs)
			}

			elevCmd := exec.Command("powershell", "-NoProfile", "-Command", psCmd)
			elevCmd.Stdout = os.Stdout
			elevCmd.Stderr = os.Stderr
			if elevErr := elevCmd.Run(); elevErr != nil {
				return fmt.Errorf("elevated installer failed: %w", elevErr)
			}
		} else {
			return fmt.Errorf("installer failed: %w", err)
		}
	}

	// Record the installation
	if err := recordInstall(pkgID, pkg, ""); err != nil {
		console.Warning("Package installed but failed to record: %v", err)
	}

	return nil
}

// Uninstall removes a package
func Uninstall(pkgID string) error {
	db, err := config.LoadInstalled()
	if err != nil {
		return fmt.Errorf("failed to load installed packages: %w", err)
	}

	installed, ok := db.Packages[pkgID]
	if !ok {
		return fmt.Errorf("package '%s' is not installed via APM", pkgID)
	}

	console.Step("🗑", "Removing %s...", console.PackageName(installed.DisplayName))

	if installed.Type == "portable" && installed.InstallPath != "" {
		// Portable: remove directory
		if err := os.RemoveAll(installed.InstallPath); err != nil {
			return fmt.Errorf("failed to remove portable app: %w", err)
		}
	} else {
		// Installer: try to find and run uninstaller from registry
		if err := runWindowsUninstall(installed.DisplayName); err != nil {
			console.Warning("Automatic uninstall failed: %v", err)
			console.Info("Try uninstalling manually via Windows Settings > Apps")
			// Still remove from our tracking
		}
	}

	// Remove from installed database
	delete(db.Packages, pkgID)
	if err := config.SaveInstalled(db); err != nil {
		return fmt.Errorf("failed to update installed database: %w", err)
	}

	return nil
}

// IsInstalled checks if a package is already installed
func IsInstalled(pkgID string) bool {
	db, err := config.LoadInstalled()
	if err != nil {
		return false
	}
	_, ok := db.Packages[pkgID]
	return ok
}

// recordInstall adds a package to the installed database
func recordInstall(pkgID string, pkg *registry.Package, installDir string) error {
	db, err := config.LoadInstalled()
	if err != nil {
		return err
	}

	db.Packages[pkgID] = config.InstalledPackage{
		Name:        pkgID,
		DisplayName: pkg.Name,
		Version:     pkg.Version,
		Type:        pkg.Type,
		InstallPath: installDir,
		InstalledAt: time.Now().Format(time.RFC3339),
	}

	return config.SaveInstalled(db)
}

// runWindowsUninstall attempts to find and run the Windows uninstaller
func runWindowsUninstall(displayName string) error {
	// Search in Windows registry for the uninstall command
	// We check both HKLM and HKCU uninstall registry keys
	regPaths := []string{
		`HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`,
		`HKLM\SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`,
		`HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`,
	}

	for _, regPath := range regPaths {
		cmd := exec.Command("reg", "query", regPath, "/s", "/f", displayName, "/d")
		output, err := cmd.Output()
		if err != nil {
			continue
		}

		// Parse the output to find the UninstallString
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "HKEY_") {
				// Found a registry key, query for QuietUninstallString or UninstallString
				for _, valueName := range []string{"QuietUninstallString", "UninstallString"} {
					qCmd := exec.Command("reg", "query", line, "/v", valueName)
					qOutput, err := qCmd.Output()
					if err != nil {
						continue
					}
					// Extract the uninstall command
					uninstallCmd := parseRegValue(string(qOutput), valueName)
					if uninstallCmd != "" {
						console.Step("🔧", "Running uninstaller...")
						return runUninstallCommand(uninstallCmd)
					}
				}
			}
		}
	}

	return fmt.Errorf("uninstaller not found for '%s'", displayName)
}

// parseRegValue extracts a value from reg query output
func parseRegValue(output string, valueName string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, valueName) {
			parts := strings.SplitN(line, "REG_SZ", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
			parts = strings.SplitN(line, "REG_EXPAND_SZ", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}

// runUninstallCommand runs an uninstall command string
func runUninstallCommand(cmdStr string) error {
	// Try to add silent flags
	cmdStr = strings.TrimSpace(cmdStr)

	// Check if it's an msiexec command
	if strings.Contains(strings.ToLower(cmdStr), "msiexec") {
		if !strings.Contains(strings.ToLower(cmdStr), "/quiet") {
			cmdStr += " /quiet /norestart"
		}
	} else {
		// For exe uninstallers, try common silent flags
		if !strings.Contains(cmdStr, "/S") && !strings.Contains(cmdStr, "/silent") {
			cmdStr += " /S"
		}
	}

	args := splitArgs(cmdStr)
	if len(args) == 0 {
		return fmt.Errorf("empty uninstall command")
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		errStr := err.Error()
		if strings.Contains(strings.ToLower(errStr), "requires elevation") || strings.Contains(errStr, "740") || strings.Contains(strings.ToLower(errStr), "повышения прав") {
			console.Warning("Удаление требует прав Администратора. Подтвердите запрос UAC...")
			
			argsListStr := ""
			if len(args) > 1 {
				argsListStr = strings.Join(args[1:], " ")
			}
			
			psCmd := fmt.Sprintf(`Start-Process -FilePath "%s" -ArgumentList '%s' -Verb RunAs -Wait`, args[0], argsListStr)
			elevCmd := exec.Command("powershell", "-NoProfile", "-Command", psCmd)
			elevCmd.Stdout = os.Stdout
			elevCmd.Stderr = os.Stderr
			if elevErr := elevCmd.Run(); elevErr != nil {
				return fmt.Errorf("elevated uninstaller failed: %w", elevErr)
			}
		} else {
			return err
		}
	}
	return nil
}

// splitArgs splits a command line arguments string respecting quotes
func splitArgs(s string) []string {
	var args []string
	var current strings.Builder
	inQuote := false
	quoteChar := byte(0)

	for i := 0; i < len(s); i++ {
		c := s[i]
		if inQuote {
			if c == quoteChar {
				inQuote = false
			} else {
				current.WriteByte(c)
			}
		} else if c == '"' || c == '\'' {
			inQuote = true
			quoteChar = c
		} else if c == ' ' {
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		} else {
			current.WriteByte(c)
		}
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}
