// internal/state/state.go
package state

import (
	"context"
	"encoding/json"
	"time"

	"github.com/pankajbeniwal/bunkr/internal/executor"
)

const StatePath = "/etc/bunkr/state.json"

type State struct {
	Hardening HardeningState         `json:"hardening"`
	Tailscale TailscaleState         `json:"tailscale"`
	Recipes   map[string]RecipeState `json:"recipes"`
}

type TailscaleState struct {
	Installed bool   `json:"installed"`
	Connected bool   `json:"connected"`
	Hostname  string `json:"hostname"`
}

type HardeningState struct {
	Applied   bool            `json:"applied"`
	Steps     map[string]bool `json:"steps"`
	AppliedAt time.Time       `json:"applied_at"`
	SSHPort   int             `json:"ssh_port"`
}

type RecipeState struct {
	Version       string    `json:"version"`
	Domain        string    `json:"domain"`
	Private       bool      `json:"private"`
	InstalledAt   time.Time `json:"installed_at"`
	Port          int       `json:"port"`
	ContainerPort int       `json:"container_port"`
}

func New() *State {
	return &State{
		Hardening: HardeningState{
			Steps: make(map[string]bool),
		},
		Recipes: make(map[string]RecipeState),
	}
}

func Load(ctx context.Context, exec executor.Executor) (*State, error) {
	data, err := exec.ReadFile(ctx, StatePath)
	if err != nil {
		return New(), nil
	}
	s := New()
	if err := json.Unmarshal(data, s); err != nil {
		return nil, err
	}
	if s.Recipes == nil {
		s.Recipes = make(map[string]RecipeState)
	}
	if s.Hardening.Steps == nil {
		s.Hardening.Steps = make(map[string]bool)
	}
	return s, nil
}

func Save(ctx context.Context, exec executor.Executor, s *State) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return exec.WriteFile(ctx, StatePath, data, 0644)
}

func (s *State) AllocatePort(desired int) int {
	taken := make(map[int]bool)
	for _, r := range s.Recipes {
		taken[r.Port] = true
	}
	port := desired
	for taken[port] {
		port++
	}
	return port
}

func (s *State) IsPortTaken(port int) bool {
	for _, r := range s.Recipes {
		if r.Port == port {
			return true
		}
	}
	return false
}
