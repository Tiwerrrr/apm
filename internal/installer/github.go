package installer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
)

type githubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Size               int64  `json:"size"`
	} `json:"assets"`
}

// FetchLatestGitHubAsset queries the GitHub API for the latest release of a repo
// and finds the best matching asset URL based on assetRegex.
func FetchLatestGitHubAsset(repo string, assetRegex string) (url string, version string, size int64, err error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	resp, err := http.Get(apiURL)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to contact GitHub API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", 0, fmt.Errorf("GitHub API returned status: %s", resp.Status)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", "", 0, fmt.Errorf("failed to parse GitHub response: %w", err)
	}

	var re *regexp.Regexp
	if assetRegex != "" {
		re, err = regexp.Compile("(?i)" + assetRegex)
		if err != nil {
			return "", "", 0, fmt.Errorf("invalid asset_regex in registry: %w", err)
		}
	}

	for _, asset := range release.Assets {
		if re != nil {
			if re.MatchString(asset.Name) {
				return asset.BrowserDownloadURL, release.TagName, asset.Size, nil
			}
		} else {
			// Fallback: just take the first .exe or .msi or .zip
			if len(asset.Name) > 4 {
				ext := asset.Name[len(asset.Name)-4:]
				if ext == ".exe" || ext == ".msi" || ext == ".zip" {
					return asset.BrowserDownloadURL, release.TagName, asset.Size, nil
				}
			}
		}
	}

	return "", "", 0, fmt.Errorf("no matching asset found in latest release for %s", repo)
}
