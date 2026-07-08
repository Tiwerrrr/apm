package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

type githubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

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

	// 4. Скачиваем самую новую версию apm.exe с GitHub
	if err := downloadLatestAPM(exePath); err != nil {
		showError("Ошибка загрузки APM", err.Error())
		os.Exit(1)
	}

	// 5. Добавляем в PATH пользователя
	if err := addToPath(installDir); err != nil {
		showError("Ошибка обновления PATH", err.Error())
		os.Exit(1)
	}

	// 6. Показываем системное уведомление об успехе
	showInfo("Установка завершена", "Самая свежая версия APM успешно скачана и установлена!\n\nПерезапустите терминал и введите 'apm help' для начала работы.")
}

func downloadLatestAPM(destPath string) error {
	resp, err := http.Get("https://api.github.com/repos/Tiwerrrr/apm/releases/latest")
	if err != nil {
		return fmt.Errorf("ошибка проверки обновлений: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub вернул статус: %s", resp.Status)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("ошибка разбора ответа GitHub: %v", err)
	}

	var downloadURL string
	for _, asset := range release.Assets {
		if strings.EqualFold(asset.Name, "apm.exe") {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("файл apm.exe не найден в последнем релизе")
	}

	// Скачиваем файл
	fileResp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("ошибка скачивания файла: %v", err)
	}
	defer fileResp.Body.Close()

	if fileResp.StatusCode != http.StatusOK {
		return fmt.Errorf("ошибка при скачивании файла, статус: %s", fileResp.Status)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("ошибка создания файла: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, fileResp.Body)
	if err != nil {
		return fmt.Errorf("ошибка записи файла: %v", err)
	}

	return nil
}

func addToPath(newPath string) error {
	cmdGet := exec.Command("powershell", "-NoProfile", "-Command", "[Environment]::GetEnvironmentVariable('Path', 'User')")
	out, err := cmdGet.Output()
	if err != nil {
		return fmt.Errorf("не удалось получить PATH: %v", err)
	}

	currentPath := strings.TrimSpace(string(out))
	paths := strings.Split(currentPath, ";")
	for _, p := range paths {
		if strings.EqualFold(strings.TrimSpace(p), newPath) {
			return nil
		}
	}

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

	psCmd := fmt.Sprintf(`[Environment]::SetEnvironmentVariable('Path', '%s', 'User')`, updatedPath)
	cmdSet := exec.Command("powershell", "-NoProfile", "-Command", psCmd)
	if err := cmdSet.Run(); err != nil {
		return fmt.Errorf("не удалось установить PATH: %v", err)
	}

	return nil
}

const (
	MB_OK              = 0x00000000
	MB_ICONINFORMATION = 0x00000040
	MB_ICONERROR       = 0x00000010
	MB_SETFOREGROUND   = 0x00010000
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
	ret := showMessageBox(title, message, 0x00000004|0x00000020|MB_SETFOREGROUND)
	return ret == 6 // IDYES
}
