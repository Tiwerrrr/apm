package registry

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/apm-cli/apm/internal/config"
)

// Вшиваем базу данных пакетов прямо в бинарник!
//go:embed registry.json
var registryData []byte

// Package represents a single package in the registry
type Package struct {
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Version       string   `json:"version"`
	Type          string   `json:"type"` // "installer" or "portable"
	URL           string   `json:"url"`
	SilentArgs    string   `json:"silent_args"`
	UninstallName string   `json:"uninstall_name"`
	Homepage      string   `json:"homepage"`
	Tags          []string `json:"tags"`
	GithubRepo    string   `json:"github_repo,omitempty"`
	AssetRegex    string   `json:"asset_regex,omitempty"`
	Dependencies  []string `json:"dependencies,omitempty"`
	Bin           string   `json:"bin,omitempty"`
}

// Registry holds all available packages
type Registry struct {
	Packages map[string]Package `json:"packages"`
}

// Load reads and parses the registry from local cache, falling back to embedded data
func Load() (*Registry, error) {
	var data []byte
	var err error

	// Try reading from the local GitHub-cached registry
	data, err = os.ReadFile(config.RegistryFile)
	if err != nil || len(data) == 0 {
		// Fallback to embedded data if cache doesn't exist
		data = registryData
	}

	var reg Registry
	if err := json.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("failed to parse registry: %w", err)
	}
	return &reg, nil
}

// Get returns a package by its ID, or nil if not found
func (r *Registry) Get(id string) *Package {
	id = strings.ToLower(strings.TrimSpace(id))
	if pkg, ok := r.Packages[id]; ok {
		return &pkg
	}
	return nil
}

// SearchResult holds a search result with its relevance score
type SearchResult struct {
	ID      string
	Package Package
	Score   int // higher = more relevant
}

// Search finds packages matching the query by name, ID, and tags
func (r *Registry) Search(query string) []SearchResult {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return nil
	}

	var results []SearchResult

	for id, pkg := range r.Packages {
		score := 0
		idLower := strings.ToLower(id)
		nameLower := strings.ToLower(pkg.Name)
		descLower := strings.ToLower(pkg.Description)

		// Exact ID match — highest priority
		if idLower == query {
			score += 100
		}

		// ID contains query
		if strings.Contains(idLower, query) {
			score += 50
		}

		// Name contains query
		if strings.Contains(nameLower, query) {
			score += 40
		}

		// Tag exact match
		for _, tag := range pkg.Tags {
			if strings.ToLower(tag) == query {
				score += 30
				break
			}
		}

		// Tag contains query
		for _, tag := range pkg.Tags {
			if strings.Contains(strings.ToLower(tag), query) {
				score += 15
				break
			}
		}

		// Description contains query
		if strings.Contains(descLower, query) {
			score += 10
		}

		// Query contains the ID or parts of ID
		if strings.Contains(query, idLower) {
			score += 20
		}

		if score > 0 {
			results = append(results, SearchResult{
				ID:      id,
				Package: pkg,
				Score:   score,
			})
		}
	}

	// Sort by score descending, then by ID ascending
	sort.Slice(results, func(i, j int) bool {
		if results[i].Score != results[j].Score {
			return results[i].Score > results[j].Score
		}
		return results[i].ID < results[j].ID
	})

	return results
}

// ListAll returns all packages sorted by ID
func (r *Registry) ListAll() []SearchResult {
	var results []SearchResult
	for id, pkg := range r.Packages {
		results = append(results, SearchResult{
			ID:      id,
			Package: pkg,
		})
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].ID < results[j].ID
	})
	return results
}
