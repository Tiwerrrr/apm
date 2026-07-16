package installer

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/apm-cli/apm/internal/console"
)

// extractZip unpacks a zip archive into a destination directory
func extractZip(src string, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip vulnerability
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("%s: illegal file path", fpath)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

// createShim creates a simple .bat file in the apm bin directory
func createShim(binDir string, exePath string, shimName string) error {
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return err
	}

	if !strings.HasSuffix(shimName, ".bat") {
		shimName += ".bat"
	}

	shimPath := filepath.Join(binDir, shimName)
	content := fmt.Sprintf(`@echo off
"%s" %%*
`, exePath)

	console.Info("Создан shim-файл (ярлык CLI): %s -> %s", shimName, exePath)
	return os.WriteFile(shimPath, []byte(content), 0755)
}

// createStartMenuShortcut creates a .lnk shortcut in the Windows Start Menu
func createStartMenuShortcut(exePath string, appName string) error {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return fmt.Errorf("APPDATA environment variable not set")
	}

	startMenuDir := filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs")
	os.MkdirAll(startMenuDir, 0755)

	shortcutPath := filepath.Join(startMenuDir, appName+".lnk")
	
	psCmd := fmt.Sprintf(`$s=(New-Object -COM WScript.Shell).CreateShortcut('%s'); $s.TargetPath='%s'; $s.Save()`, shortcutPath, exePath)
	
	cmd := exec.Command("powershell", "-NoProfile", "-Command", psCmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("powershell failed to create shortcut: %w", err)
	}
	
	console.Info("Добавлен ярлык %s в меню Пуск", appName)
	return nil
}
