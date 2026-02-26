// internal/recipe/recipe.go
package recipe

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type Recipe struct {
	Name        string            `yaml:"name"`
	Version     string            `yaml:"version"`
	Description string            `yaml:"description"`
	Image       string            `yaml:"image"`
	Private     bool              `yaml:"private"`
	Prompts     []Prompt          `yaml:"prompts"`
	Ports       []int             `yaml:"ports"`
	Volumes     []string          `yaml:"volumes"`
	Services    []Service         `yaml:"services"`
	Environment map[string]string `yaml:"environment"`
	HealthCheck *HealthCheck      `yaml:"health_check"`
}

type Prompt struct {
	Key      string `yaml:"key"`
	Label    string `yaml:"label"`
	Required bool   `yaml:"required"`
	Default  string `yaml:"default"`
	Secret   bool   `yaml:"secret"`
}

type Service struct {
	Name        string            `yaml:"name"`
	Image       string            `yaml:"image"`
	Environment map[string]string `yaml:"environment"`
	Volumes     []string          `yaml:"volumes"`
}

type HealthCheck struct {
	URL      string `yaml:"url"`
	Timeout  int    `yaml:"timeout"`
	Interval int    `yaml:"interval"`
}

func Parse(data []byte) (*Recipe, error) {
	r := &Recipe{}
	if err := yaml.Unmarshal(data, r); err != nil {
		return nil, fmt.Errorf("invalid recipe YAML: %w", err)
	}
	if r.HealthCheck != nil {
		if r.HealthCheck.Timeout == 0 {
			r.HealthCheck.Timeout = 30
		}
		if r.HealthCheck.Interval == 0 {
			r.HealthCheck.Interval = 2
		}
	}
	return r, nil
}

func (r *Recipe) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("recipe name is required")
	}
	if r.Version == "" {
		return fmt.Errorf("recipe version is required")
	}
	if r.Image == "" {
		return fmt.Errorf("recipe image is required")
	}
	if len(r.Ports) == 0 {
		return fmt.Errorf("recipe must expose at least one port")
	}
	return nil
}
