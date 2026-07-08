package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/apm-cli/apm/internal/config"
	"github.com/apm-cli/apm/internal/console"
	"github.com/apm-cli/apm/internal/installer"
)

// SelfDestruct permanently uninstalls APM and its data
func SelfDestruct() error {
	console.Warning("ВНИМАНИЕ! Вы собираетесь полностью удалить Awesome Package Manager.")
	fmt.Println()

	keepApps := console.AskYesNoConsole("Оставить программы, которые были установлены через APM?")
	keepData := console.AskYesNoConsole("Оставить кэш и список программ (ускорит переустановку в будущем)?")

	if !keepApps {
		console.Step("🗑", "Удаление установленных программ...")
		db, err := config.LoadInstalled()
		if err == nil {
			for pkgID := range db.Packages {
				if err := installer.Uninstall(pkgID); err != nil {
					console.Warning("Не удалось удалить %s: %v", pkgID, err)
				}
			}
		}
	}

	console.Step("🔥", "Инициализация самоуничтожения...")

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("ошибка получения пути: %w", err)
	}

	var dataDir string
	if !keepData {
		dataDir = config.RootDir
	}

	// PowerShell script to wait 2 seconds, delete APM and its folders, then exit
	psCmd := fmt.Sprintf(`
Start-Sleep -Seconds 2
Remove-Item -Path '%s' -Force -ErrorAction SilentlyContinue
if ('%s' -ne '') {
    Remove-Item -Path '%s' -Recurse -Force -ErrorAction SilentlyContinue
}
`, filepath.ToSlash(exePath), filepath.ToSlash(dataDir), filepath.ToSlash(dataDir))

	cmd := exec.Command("powershell", "-WindowStyle", "Hidden", "-Command", psCmd)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("не удалось запустить процесс удаления: %w", err)
	}

	fmt.Println()
	console.Success("APM удален! Программа завершает работу. До свидания!")
	os.Exit(0)
	return nil
}
