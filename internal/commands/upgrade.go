package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"unsafe"

	"github.com/apm-cli/apm/internal/config"
	"github.com/apm-cli/apm/internal/console"
	"github.com/apm-cli/apm/internal/downloader"
)

type githubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Size               int64  `json:"size"`
	} `json:"assets"`
}

func askYesNo(title, text string) bool {
	user32 := syscall.NewLazyDLL("user32.dll")
	procMessageBoxW := user32.NewProc("MessageBoxW")

	// MB_YESNO | MB_ICONQUESTION | MB_TOPMOST
	const mbYesNo = 0x00000004
	const mbIconQuestion = 0x00000020
	const mbTopMost = 0x00040000
	const idYes = 6

	titlePtr, _ := syscall.UTF16PtrFromString(title)
	textPtr, _ := syscall.UTF16PtrFromString(text)

	ret, _, _ := procMessageBoxW.Call(
		0,
		uintptr(unsafe.Pointer(textPtr)),
		uintptr(unsafe.Pointer(titlePtr)),
		uintptr(mbYesNo|mbIconQuestion|mbTopMost),
	)

	return ret == idYes
}

// UpgradeSelf checks GitHub for a new release and updates APM if the user agrees
func UpgradeSelf() error {
	console.Step("🔍", "Checking for APM updates...")

	resp, err := http.Get("https://api.github.com/repos/Tiwerrrr/apm/releases/latest")
	if err != nil {
		return fmt.Errorf("failed to check updates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		console.Warning("No releases found on GitHub yet.")
		return nil
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub returned status: %s", resp.Status)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("failed to parse release info: %w", err)
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	currentVersion := strings.TrimPrefix(config.Version, "v")

	if latestVersion == currentVersion {
		console.Success("APM is already up-to-date (v%s)!", config.Version)
		return nil
	}

	console.Info("New version found: v%s (Current: v%s)", latestVersion, currentVersion)

	var installerURL string
	for _, asset := range release.Assets {
		if strings.HasSuffix(strings.ToLower(asset.Name), ".exe") {
			installerURL = asset.BrowserDownloadURL
			break
		}
	}

	if installerURL == "" {
		return fmt.Errorf("no executable installer found in the latest release assets")
	}

	// Prompt user with GUI MessageBox
	msg := fmt.Sprintf("Доступна новая версия APM (v%s)!\n\nТекущая версия: v%s\n\nХотите обновить пакетный менеджер сейчас?", latestVersion, currentVersion)
	if !askYesNo("Обновление APM", msg) {
		console.Warning("Обновление отменено пользователем.")
		return nil
	}

	console.Step("⬇", "Downloading apm-installer.exe...")

	tempDir := os.TempDir()
	
	downloadedPath, err := downloader.Download(installerURL, tempDir, "apm-installer-new.exe")
	if err != nil {
		return fmt.Errorf("failed to download installer: %w", err)
	}
	tempInstaller := downloadedPath

	// Rename current executable to .old so we can overwrite it
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable path: %w", err)
	}

	oldExePath := exePath + ".old"
	os.Remove(oldExePath) // Ignore error if it doesn't exist

	if err := os.Rename(exePath, oldExePath); err != nil {
		return fmt.Errorf("failed to rename current executable (is it locked?): %w", err)
	}

	console.Step("🚀", "Starting installer...")

	// Start the installer and exit
	cmd := exec.Command(tempInstaller)
	if err := cmd.Start(); err != nil {
		// Rollback rename if failed to start
		os.Rename(oldExePath, exePath)
		return fmt.Errorf("failed to start new installer: %w", err)
	}

	os.Exit(0)
	return nil
}
