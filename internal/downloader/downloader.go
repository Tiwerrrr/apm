package downloader

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/apm-cli/apm/internal/console"
)

// progressReader wraps an io.Reader and reports progress with speed and ETA
type progressReader struct {
	reader    io.Reader
	total     int64
	current   int64
	barWidth  int
	lastPct   int
	startTime time.Time
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.current += int64(n)

	// Update progress bar
	if pr.total > 0 {
		pct := int(float64(pr.current) / float64(pr.total) * 100)
		if pct != pr.lastPct {
			pr.lastPct = pct
			bar := console.ProgressBar(pr.current, pr.total, pr.barWidth)

			// Calculate speed and ETA
			elapsed := time.Since(pr.startTime).Seconds()
			speedStr := ""
			etaStr := ""
			if elapsed > 0.5 {
				speed := float64(pr.current) / elapsed
				speedStr = console.FormatBytes(int64(speed)) + "/s"

				if pr.current < pr.total && speed > 0 {
					remaining := float64(pr.total-pr.current) / speed
					if remaining < 60 {
						etaStr = fmt.Sprintf("ETA %ds", int(remaining))
					} else {
						etaStr = fmt.Sprintf("ETA %dm%ds", int(remaining)/60, int(remaining)%60)
					}
				}
			}

			fmt.Printf("\r  %s  ⬇  %s / %s  %s  %s%s",
				bar,
				console.FormatBytes(pr.current),
				console.FormatBytes(pr.total),
				speedStr,
				etaStr,
				strings.Repeat(" ", 5), // Clear leftover chars
			)
		}
	} else {
		// Unknown size — show bytes downloaded with speed
		elapsed := time.Since(pr.startTime).Seconds()
		speedStr := ""
		if elapsed > 0.5 {
			speed := float64(pr.current) / elapsed
			speedStr = "  " + console.FormatBytes(int64(speed)) + "/s"
		}
		fmt.Printf("\r  ⬇  Downloaded: %s%s%s",
			console.FormatBytes(pr.current),
			speedStr,
			strings.Repeat(" ", 10),
		)
	}

	return n, err
}

// Download downloads a file from url and saves it to the cache directory.
// Returns the path to the downloaded file.
func Download(url string, cacheDir string, filename string) (string, error) {
	// Create cache directory
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache dir: %w", err)
	}

	destPath := filepath.Join(cacheDir, filename)

	// Check if already cached
	if info, err := os.Stat(destPath); err == nil && info.Size() > 0 {
		console.Info("Using cached file: %s", console.FormatBytes(info.Size()))
		return destPath, nil
	}

	// Start download
	console.Step("⬇", "Downloading from %s%s%s...", console.Dim, truncateURL(url, 60), console.Reset)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed: HTTP %d %s", resp.StatusCode, resp.Status)
	}

	// Create temp file
	tmpPath := destPath + ".tmp"
	file, err := os.Create(tmpPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		file.Close()
		// Clean up temp file on error
		if _, err := os.Stat(tmpPath); err == nil {
			os.Remove(tmpPath)
		}
	}()

	// Copy with progress
	pr := &progressReader{
		reader:    resp.Body,
		total:     resp.ContentLength,
		barWidth:  30,
		lastPct:   -1,
		startTime: time.Now(),
	}

	_, err = io.Copy(file, pr)
	if err != nil {
		return "", fmt.Errorf("download interrupted: %w", err)
	}
	file.Close()

	// Move temp file to final destination
	if err := os.Rename(tmpPath, destPath); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	// Final stats
	elapsed := time.Since(pr.startTime).Seconds()
	avgSpeed := ""
	if elapsed > 0 {
		speed := float64(pr.current) / elapsed
		avgSpeed = fmt.Sprintf(" (avg %s/s)", console.FormatBytes(int64(speed)))
	}

	fmt.Println() // New line after progress bar
	console.Success("Download complete: %s%s", console.FormatBytes(pr.current), avgSpeed)

	return destPath, nil
}

// truncateURL shortens a URL for display
func truncateURL(url string, maxLen int) string {
	if len(url) <= maxLen {
		return url
	}
	return url[:maxLen-3] + "..."
}
