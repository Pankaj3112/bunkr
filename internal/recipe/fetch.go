// internal/recipe/fetch.go
package recipe

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const defaultBaseURL = "https://raw.githubusercontent.com/pankajbeniwal/bunkr/main/recipes"

func BuildRecipeURL(name string, baseURL string) string {
	base := defaultBaseURL
	if baseURL != "" {
		base = strings.TrimRight(baseURL, "/")
	}
	return fmt.Sprintf("%s/%s.yaml", base, name)
}

func BuildIndexURL(baseURL string) string {
	base := defaultBaseURL
	if baseURL != "" {
		base = strings.TrimRight(baseURL, "/")
	}
	return fmt.Sprintf("%s/index.yaml", base)
}

func getBaseURL() string {
	if url := os.Getenv("BUNKR_RECIPES_URL"); url != "" {
		return url
	}
	return ""
}

func Fetch(name string) (*Recipe, error) {
	url := BuildRecipeURL(name, getBaseURL())
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recipe %s: %w", name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("recipe %s not found (HTTP %d)", name, resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read recipe %s: %w", name, err)
	}

	r, err := Parse(data)
	if err != nil {
		return nil, err
	}
	if err := r.Validate(); err != nil {
		return nil, fmt.Errorf("invalid recipe %s: %w", name, err)
	}
	return r, nil
}

type IndexEntry struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Version     string `yaml:"version"`
}

func FetchIndex() ([]IndexEntry, error) {
	url := BuildIndexURL(getBaseURL())
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recipe index: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("recipe index not found (HTTP %d)", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read recipe index: %w", err)
	}

	var index []IndexEntry
	if err := yaml.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("invalid recipe index: %w", err)
	}
	return index, nil
}
