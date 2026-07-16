package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/apm-cli/apm/internal/config"
	"github.com/apm-cli/apm/internal/console"
)

// LoadRepos loads the custom repositories from repos.json
func LoadRepos() (map[string]string, error) {
	data, err := os.ReadFile(config.ReposFile)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]string), nil
		}
		return nil, err
	}
	var repos map[string]string
	if err := json.Unmarshal(data, &repos); err != nil {
		return nil, err
	}
	if repos == nil {
		repos = make(map[string]string)
	}
	return repos, nil
}

// SaveRepos saves the custom repositories to repos.json
func SaveRepos(repos map[string]string) error {
	data, err := json.MarshalIndent(repos, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(config.ReposFile, data, 0644)
}

// RepoAdd adds a custom repository
func RepoAdd(name, url string) error {
	name = strings.ToLower(name)
	repos, err := LoadRepos()
	if err != nil {
		return err
	}
	repos[name] = url
	if err := SaveRepos(repos); err != nil {
		return err
	}
	console.Success("Repository '%s' added successfully!", name)
	console.Info("Run 'apm update' to fetch the latest packages.")
	return nil
}

// RepoRemove removes a custom repository
func RepoRemove(name string) error {
	name = strings.ToLower(name)
	repos, err := LoadRepos()
	if err != nil {
		return err
	}
	if _, ok := repos[name]; !ok {
		return fmt.Errorf("repository '%s' not found", name)
	}
	delete(repos, name)
	if err := SaveRepos(repos); err != nil {
		return err
	}
	// Also remove the cached file
	os.Remove(filepath.Join(config.ReposDir, name+".json"))
	console.Success("Repository '%s' removed.", name)
	return nil
}

// RepoList lists custom repositories
func RepoList() error {
	repos, err := LoadRepos()
	if err != nil {
		return err
	}
	if len(repos) == 0 {
		console.Info("No custom repositories added.")
		return nil
	}
	console.Step("📋", "Custom Repositories:")
	for name, url := range repos {
		fmt.Printf("  • %s%s%s - %s\n", console.Bold, name, console.Reset, url)
	}
	return nil
}
