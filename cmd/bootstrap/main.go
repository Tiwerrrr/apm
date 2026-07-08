package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

// Вшиваем файл apm.exe прямо внутрь нашего инсталлера!
//go:embed apm.exe
var apmBinary []byte

func main() {
	// 1. Определяем пути
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		localAppData = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local")
	}

	installDir := filepath.Join(localAppData, "apm", "bin")
	exePath := filepath.Join(installDir, "apm.exe")

	// 2. Проверяем, установлен ли уже APM
	if _, err := os.Stat(exePath); err == nil {
		if !askYesNo("APM уже установлен", "Хотите ли вы переустановить APM (без потери списков установленных программ)?") {
			os.Exit(0)
		}
	}

	// 3. Создаем директорию
	if err := os.MkdirAll(installDir, 0755); err != nil {
		showError("Ошибка создания директории", err.Error())
		os.Exit(1)
	}

	// 3. Сохраняем вшитый файл на диск пользователя
	if err := os.WriteFile(exePath, apmBinary, 0755); err != nil {
		showError("Ошибка распаковки", err.Error())
		os.Exit(1)
	}

	// 4. Добавляем в PATH пользователя
	if err := addToPath(installDir); err != nil {
		showError("Ошибка обновления PATH", err.Error())
		os.Exit(1)
	}

	// 5. Показываем системное уведомление об успехе
	showInfo("Установка завершена", "APM успешно установлен!\n\nПерезапустите терминал и введите 'apm help' для начала работы.")
}

func addToPath(newPath string) error {
	// Получаем текущий PATH пользователя через PowerShell
	cmdGet := exec.Command("powershell", "-NoProfile", "-Command", "[Environment]::GetEnvironmentVariable('Path', 'User')")
	out, err := cmdGet.Output()
	if err != nil {
		return fmt.Errorf("не удалось получить PATH: %v", err)
	}

	currentPath := strings.TrimSpace(string(out))
	
	// Проверяем, есть ли уже наш путь в PATH
	paths := strings.Split(currentPath, ";")
	for _, p := range paths {
		if strings.EqualFold(strings.TrimSpace(p), newPath) {
			return nil // Уже есть в PATH
		}
	}

	// Добавляем наш путь
	var updatedPath string
	if currentPath == "" {
		updatedPath = newPath
	} else {
		if strings.HasSuffix(currentPath, ";") {
			updatedPath = currentPath + newPath
		} else {
			updatedPath = currentPath + ";" + newPath
		}
	}

	// Сохраняем новый PATH
	psCmd := fmt.Sprintf(`[Environment]::SetEnvironmentVariable('Path', '%s', 'User')`, updatedPath)
	cmdSet := exec.Command("powershell", "-NoProfile", "-Command", psCmd)
	if err := cmdSet.Run(); err != nil {
		return fmt.Errorf("не удалось установить PATH: %v", err)
	}

	return nil
}

// WinAPI константы для MessageBox
const (
	MB_OK                = 0x00000000
	MB_ICONINFORMATION   = 0x00000040
	MB_ICONERROR         = 0x00000010
	MB_SETFOREGROUND     = 0x00010000
)

func showInfo(title, message string) {
	showMessageBox(title, message, MB_OK|MB_ICONINFORMATION|MB_SETFOREGROUND)
}

func showError(title, message string) {
	showMessageBox(title, message, MB_OK|MB_ICONERROR|MB_SETFOREGROUND)
}

func showMessageBox(title, message string, flags uint32) uintptr {
	user32 := syscall.NewLazyDLL("user32.dll")
	messageBox := user32.NewProc("MessageBoxW")

	t, _ := syscall.UTF16PtrFromString(title)
	m, _ := syscall.UTF16PtrFromString(message)

	ret, _, _ := messageBox.Call(
		0,
		uintptr(unsafe.Pointer(m)),
		uintptr(unsafe.Pointer(t)),
		uintptr(flags),
	)
	return ret
}

func askYesNo(title, message string) bool {
	// MB_YESNO | MB_ICONQUESTION | MB_SETFOREGROUND
	ret := showMessageBox(title, message, 0x00000004|0x00000020|MB_SETFOREGROUND)
	return ret == 6 // IDYES
}
