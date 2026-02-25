# Bunkr Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a CLI tool that hardens a VPS and deploys self-hosted apps in one command, supporting both local and remote execution.

**Architecture:** Two-phase model (Plan local, Execute via Executor interface). Package-per-domain structure. LocalExecutor and RemoteExecutor share a 3-method interface (Run, WriteFile, ReadFile).

**Tech Stack:** Go, Cobra, fatih/color, golang.org/x/crypto/ssh, golang.org/x/term, gopkg.in/yaml.v3, GoReleaser

---

### Task 1: Project Scaffolding

**Files:**
- Create: `go.mod`
- Create: `cmd/bunkr/main.go`
- Create: `cmd/bunkr/root.go`

**Step 1: Initialize Go module and install dependencies**

Run:
```bash
cd /Users/pankajbeniwal/Code/bunkr
go mod init github.com/pankajbeniwal/bunkr
go get github.com/spf13/cobra@latest
go get github.com/fatih/color@latest
go get gopkg.in/yaml.v3@latest
go get golang.org/x/crypto/ssh@latest
go get golang.org/x/term@latest
```

**Step 2: Create main.go**

```go
// cmd/bunkr/main.go
package main

import "os"

var version = "dev"

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
```

**Step 3: Create root.go with --on flag**

```go
// cmd/bunkr/root.go
package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var onFlag string

var rootCmd = &cobra.Command{
	Use:   "bunkr",
	Short: "Harden a VPS and deploy self-hosted apps in one command",
	Long:  "Bunkr takes a fresh VPS and turns it into a hardened server running any self-hosted app.",
}

func init() {
	rootCmd.PersistentFlags().StringVar(&onFlag, "on", "", "remote server to execute on (e.g., root@167.71.50.23)")
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print bunkr version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("bunkr", version)
	},
}
```

**Step 4: Verify it compiles and runs**

Run: `go run ./cmd/bunkr version`
Expected: `bunkr dev`

**Step 5: Commit**

```bash
git add cmd/ go.mod go.sum
git commit -m "feat: scaffold project with cobra root command and --on flag"
```

---

### Task 2: Executor Interface + Mock

**Files:**
- Create: `internal/executor/executor.go`
- Create: `internal/executor/mock.go`
- Create: `internal/executor/executor_test.go`

**Step 1: Write the Executor interface and MockExecutor**

```go
// internal/executor/executor.go
package executor

import (
	"context"
	"os"
)

type Executor interface {
	Run(ctx context.Context, cmd string) (string, error)
	WriteFile(ctx context.Context, path string, content []byte, mode os.FileMode) error
	ReadFile(ctx context.Context, path string) ([]byte, error)
}
```

```go
// internal/executor/mock.go
package executor

import (
	"context"
	"fmt"
	"os"
)

type MockCall struct {
	Method string
	Args   []interface{}
}

type MockExecutor struct {
	Calls       []MockCall
	RunOutputs  map[string]string
	RunErrors   map[string]error
	Files       map[string][]byte
	ReadErrors  map[string]error
	WriteErrors map[string]error
}

func NewMockExecutor() *MockExecutor {
	return &MockExecutor{
		RunOutputs:  make(map[string]string),
		RunErrors:   make(map[string]error),
		Files:       make(map[string][]byte),
		ReadErrors:  make(map[string]error),
		WriteErrors: make(map[string]error),
	}
}

func (m *MockExecutor) Run(_ context.Context, cmd string) (string, error) {
	m.Calls = append(m.Calls, MockCall{Method: "Run", Args: []interface{}{cmd}})
	if err, ok := m.RunErrors[cmd]; ok {
		return "", err
	}
	if out, ok := m.RunOutputs[cmd]; ok {
		return out, nil
	}
	return "", nil
}

func (m *MockExecutor) WriteFile(_ context.Context, path string, content []byte, mode os.FileMode) error {
	m.Calls = append(m.Calls, MockCall{Method: "WriteFile", Args: []interface{}{path, content, mode}})
	if err, ok := m.WriteErrors[path]; ok {
		return err
	}
	m.Files[path] = content
	return nil
}

func (m *MockExecutor) ReadFile(_ context.Context, path string) ([]byte, error) {
	m.Calls = append(m.Calls, MockCall{Method: "ReadFile", Args: []interface{}{path}})
	if err, ok := m.ReadErrors[path]; ok {
		return nil, err
	}
	if data, ok := m.Files[path]; ok {
		return data, nil
	}
	return nil, fmt.Errorf("file not found: %s", path)
}
```

**Step 2: Write test for MockExecutor**

```go
// internal/executor/executor_test.go
package executor

import (
	"context"
	"testing"
)

func TestMockExecutor_Run(t *testing.T) {
	m := NewMockExecutor()
	m.RunOutputs["echo hello"] = "hello\n"

	out, err := m.Run(context.Background(), "echo hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "hello\n" {
		t.Fatalf("expected 'hello\\n', got %q", out)
	}
	if len(m.Calls) != 1 || m.Calls[0].Method != "Run" {
		t.Fatalf("expected 1 Run call, got %v", m.Calls)
	}
}

func TestMockExecutor_WriteReadFile(t *testing.T) {
	m := NewMockExecutor()
	ctx := context.Background()

	err := m.WriteFile(ctx, "/tmp/test", []byte("data"), 0644)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := m.ReadFile(ctx, "/tmp/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != "data" {
		t.Fatalf("expected 'data', got %q", string(data))
	}
}

func TestMockExecutor_ReadFile_NotFound(t *testing.T) {
	m := NewMockExecutor()
	_, err := m.ReadFile(context.Background(), "/nonexistent")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
```

**Step 3: Run tests**

Run: `go test ./internal/executor/ -v`
Expected: All 3 tests PASS

**Step 4: Commit**

```bash
git add internal/executor/
git commit -m "feat: add Executor interface and MockExecutor for testing"
```

---

### Task 3: LocalExecutor

**Files:**
- Create: `internal/executor/local.go`
- Create: `internal/executor/local_test.go`

**Step 1: Write failing test**

```go
// internal/executor/local_test.go
package executor

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalExecutor_Run(t *testing.T) {
	exec := NewLocalExecutor()
	out, err := exec.Run(context.Background(), "echo hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "hello\n" {
		t.Fatalf("expected 'hello\\n', got %q", out)
	}
}

func TestLocalExecutor_Run_Error(t *testing.T) {
	exec := NewLocalExecutor()
	_, err := exec.Run(context.Background(), "false")
	if err == nil {
		t.Fatal("expected error from 'false' command")
	}
}

func TestLocalExecutor_WriteReadFile(t *testing.T) {
	exec := NewLocalExecutor()
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "test.txt")

	err := exec.WriteFile(ctx, path, []byte("hello world"), 0644)
	if err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}

	data, err := exec.ReadFile(ctx, path)
	if err != nil {
		t.Fatalf("unexpected read error: %v", err)
	}
	if string(data) != "hello world" {
		t.Fatalf("expected 'hello world', got %q", string(data))
	}

	info, _ := os.Stat(path)
	if info.Mode().Perm() != 0644 {
		t.Fatalf("expected mode 0644, got %v", info.Mode().Perm())
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/executor/ -run TestLocalExecutor -v`
Expected: FAIL — `NewLocalExecutor` not defined

**Step 3: Implement LocalExecutor**

```go
// internal/executor/local.go
package executor

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
)

type LocalExecutor struct{}

func NewLocalExecutor() *LocalExecutor {
	return &LocalExecutor{}
}

func (l *LocalExecutor) Run(ctx context.Context, cmd string) (string, error) {
	c := exec.CommandContext(ctx, "sh", "-c", cmd)
	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr
	if err := c.Run(); err != nil {
		return "", fmt.Errorf("%w: %s", err, stderr.String())
	}
	return stdout.String(), nil
}

func (l *LocalExecutor) WriteFile(_ context.Context, path string, content []byte, mode os.FileMode) error {
	return os.WriteFile(path, content, mode)
}

func (l *LocalExecutor) ReadFile(_ context.Context, path string) ([]byte, error) {
	return os.ReadFile(path)
}
```

**Step 4: Run tests**

Run: `go test ./internal/executor/ -v`
Expected: All tests PASS

**Step 5: Commit**

```bash
git add internal/executor/local.go internal/executor/local_test.go
git commit -m "feat: add LocalExecutor implementation"
```

---

### Task 4: UI Package

**Files:**
- Create: `internal/ui/ui.go`

**Step 1: Implement UI helpers**

```go
// internal/ui/ui.go
package ui

import (
	"fmt"

	"github.com/fatih/color"
)

var (
	green  = color.New(color.FgGreen).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
	cyan   = color.New(color.FgCyan).SprintFunc()
	bold   = color.New(color.Bold).SprintFunc()
)

func Success(msg string) {
	fmt.Printf("  %s %s\n", green("✓"), msg)
}

func Skip(msg string) {
	fmt.Printf("  %s %s\n", yellow("⏭"), msg)
}

func Warn(msg string) {
	fmt.Printf("  %s %s\n", yellow("⚠"), msg)
}

func Error(msg string) {
	fmt.Printf("  %s %s\n", red("✗"), msg)
}

func Info(msg string) {
	fmt.Printf("  %s\n", cyan(msg))
}

func Header(msg string) {
	fmt.Printf("\n  %s\n", bold(msg))
}

func Result(msg string) {
	fmt.Printf("\n  %s\n\n", green(msg))
}
```

**Step 2: Verify it compiles**

Run: `go build ./internal/ui/`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/ui/
git commit -m "feat: add UI package with colored output helpers"
```

---

### Task 5: State Management

**Files:**
- Create: `internal/state/state.go`
- Create: `internal/state/state_test.go`

**Step 1: Write failing test**

```go
// internal/state/state_test.go
package state

import (
	"context"
	"testing"
	"time"

	"github.com/pankajbeniwal/bunkr/internal/executor"
)

func TestLoadState_Empty(t *testing.T) {
	mock := executor.NewMockExecutor()
	s, err := Load(context.Background(), mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Hardening.Applied {
		t.Fatal("expected hardening not applied on fresh state")
	}
	if len(s.Recipes) != 0 {
		t.Fatal("expected no recipes on fresh state")
	}
}

func TestSaveAndLoadState(t *testing.T) {
	mock := executor.NewMockExecutor()
	ctx := context.Background()

	s := New()
	s.Hardening.Applied = true
	s.Hardening.Steps["ssh_hardening"] = true
	s.Hardening.SSHPort = 2222
	s.Hardening.AppliedAt = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	s.Recipes["ghost"] = RecipeState{
		Version:       "5.82.2",
		Domain:        "blog.example.com",
		InstalledAt:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Port:          2368,
		ContainerPort: 2368,
	}

	err := Save(ctx, mock, s)
	if err != nil {
		t.Fatalf("unexpected save error: %v", err)
	}

	loaded, err := Load(ctx, mock)
	if err != nil {
		t.Fatalf("unexpected load error: %v", err)
	}
	if !loaded.Hardening.Applied {
		t.Fatal("expected hardening applied")
	}
	if loaded.Hardening.SSHPort != 2222 {
		t.Fatalf("expected SSH port 2222, got %d", loaded.Hardening.SSHPort)
	}
	r, ok := loaded.Recipes["ghost"]
	if !ok {
		t.Fatal("expected ghost recipe in state")
	}
	if r.Domain != "blog.example.com" {
		t.Fatalf("expected domain blog.example.com, got %s", r.Domain)
	}
	if r.ContainerPort != 2368 {
		t.Fatalf("expected container port 2368, got %d", r.ContainerPort)
	}
}

func TestAllocatePort(t *testing.T) {
	s := New()
	s.Recipes["app1"] = RecipeState{Port: 3000}

	port := s.AllocatePort(3000)
	if port != 3001 {
		t.Fatalf("expected 3001 (3000 taken), got %d", port)
	}

	port2 := s.AllocatePort(8080)
	if port2 != 8080 {
		t.Fatalf("expected 8080 (free), got %d", port2)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/state/ -v`
Expected: FAIL — package doesn't exist

**Step 3: Implement state package**

```go
// internal/state/state.go
package state

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/pankajbeniwal/bunkr/internal/executor"
)

const StatePath = "/etc/bunkr/state.json"

type State struct {
	Hardening HardeningState          `json:"hardening"`
	Recipes   map[string]RecipeState  `json:"recipes"`
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
```

**Step 4: Run tests**

Run: `go test ./internal/state/ -v`
Expected: All 3 tests PASS

**Step 5: Commit**

```bash
git add internal/state/
git commit -m "feat: add state management with port allocation"
```

---

### Task 6: Recipe Parsing

**Files:**
- Create: `internal/recipe/recipe.go`
- Create: `internal/recipe/recipe_test.go`

**Step 1: Write failing test**

```go
// internal/recipe/recipe_test.go
package recipe

import (
	"testing"
)

const testRecipeYAML = `
name: uptime-kuma
version: "1.23.11"
description: Self-hosted uptime monitoring
image: louislam/uptime-kuma:1.23.11

prompts:
  - key: DOMAIN
    label: "Domain for Uptime Kuma"
    required: true

ports:
  - 3001

volumes:
  - uptime_kuma_data:/app/data

health_check:
  url: "http://localhost:3001"
  timeout: 30
  interval: 2
`

func TestParseRecipe(t *testing.T) {
	r, err := Parse([]byte(testRecipeYAML))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Name != "uptime-kuma" {
		t.Fatalf("expected name uptime-kuma, got %s", r.Name)
	}
	if r.Version != "1.23.11" {
		t.Fatalf("expected version 1.23.11, got %s", r.Version)
	}
	if r.Image != "louislam/uptime-kuma:1.23.11" {
		t.Fatalf("expected image louislam/uptime-kuma:1.23.11, got %s", r.Image)
	}
	if len(r.Prompts) != 1 {
		t.Fatalf("expected 1 prompt, got %d", len(r.Prompts))
	}
	if r.Prompts[0].Key != "DOMAIN" || !r.Prompts[0].Required {
		t.Fatalf("unexpected prompt: %+v", r.Prompts[0])
	}
	if len(r.Ports) != 1 || r.Ports[0] != 3001 {
		t.Fatalf("expected ports [3001], got %v", r.Ports)
	}
	if r.HealthCheck == nil || r.HealthCheck.Timeout != 30 {
		t.Fatalf("unexpected health check: %+v", r.HealthCheck)
	}
}

const testMultiServiceYAML = `
name: plausible
version: "2.0.0"
description: Privacy-friendly web analytics
image: ghcr.io/plausible/community-edition:v2.0.0

prompts:
  - key: DOMAIN
    label: "Domain for Plausible"
    required: true
  - key: ADMIN_PASSWORD
    label: "Admin password"
    required: true
    secret: true

ports:
  - 8000

services:
  - name: plausible_db
    image: postgres:16-alpine
    environment:
      POSTGRES_PASSWORD: "postgres"
    volumes:
      - plausible_db:/var/lib/postgresql/data

environment:
  BASE_URL: "https://${DOMAIN}"
  SECRET_KEY_BASE: "auto_generate_64"
`

func TestParseRecipe_MultiService(t *testing.T) {
	r, err := Parse([]byte(testMultiServiceYAML))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(r.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(r.Services))
	}
	if r.Services[0].Name != "plausible_db" {
		t.Fatalf("expected service name plausible_db, got %s", r.Services[0].Name)
	}
	if r.Environment["SECRET_KEY_BASE"] != "auto_generate_64" {
		t.Fatalf("expected auto_generate_64, got %s", r.Environment["SECRET_KEY_BASE"])
	}
}

func TestValidateRecipe(t *testing.T) {
	r := &Recipe{Name: "", Version: "1.0", Image: "img", Ports: []int{80}}
	if err := r.Validate(); err == nil {
		t.Fatal("expected validation error for empty name")
	}

	r.Name = "test"
	r.Image = ""
	if err := r.Validate(); err == nil {
		t.Fatal("expected validation error for empty image")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/recipe/ -v`
Expected: FAIL — package doesn't exist

**Step 3: Implement recipe parsing**

```go
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
```

**Step 4: Run tests**

Run: `go test ./internal/recipe/ -v`
Expected: All tests PASS

**Step 5: Commit**

```bash
git add internal/recipe/recipe.go internal/recipe/recipe_test.go
git commit -m "feat: add recipe YAML parsing and validation"
```

---

### Task 7: Recipe Fetching

**Files:**
- Create: `internal/recipe/fetch.go`
- Create: `internal/recipe/fetch_test.go`

**Step 1: Write failing test**

```go
// internal/recipe/fetch_test.go
package recipe

import (
	"testing"
)

func TestBuildRecipeURL(t *testing.T) {
	url := BuildRecipeURL("uptime-kuma", "")
	expected := "https://raw.githubusercontent.com/pankajbeniwal/bunkr/main/recipes/uptime-kuma.yaml"
	if url != expected {
		t.Fatalf("expected %s, got %s", expected, url)
	}
}

func TestBuildRecipeURL_Override(t *testing.T) {
	url := BuildRecipeURL("ghost", "https://example.com/recipes")
	expected := "https://example.com/recipes/ghost.yaml"
	if url != expected {
		t.Fatalf("expected %s, got %s", expected, url)
	}
}

func TestBuildIndexURL(t *testing.T) {
	url := BuildIndexURL("")
	expected := "https://raw.githubusercontent.com/pankajbeniwal/bunkr/main/recipes/index.yaml"
	if url != expected {
		t.Fatalf("expected %s, got %s", expected, url)
	}
}
```

**Step 2: Implement fetching**

```go
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
```

**Step 3: Run tests**

Run: `go test ./internal/recipe/ -v -run TestBuild`
Expected: All 3 URL builder tests PASS

**Step 4: Commit**

```bash
git add internal/recipe/fetch.go internal/recipe/fetch_test.go
git commit -m "feat: add recipe fetching from GitHub raw content"
```

---

### Task 8: Compose Generation

**Files:**
- Create: `internal/recipe/compose.go`
- Create: `internal/recipe/compose_test.go`

**Step 1: Write failing test**

```go
// internal/recipe/compose_test.go
package recipe

import (
	"strings"
	"testing"
)

func TestGenerateCompose_Simple(t *testing.T) {
	r := &Recipe{
		Name:  "uptime-kuma",
		Image: "louislam/uptime-kuma:1.23.11",
		Ports: []int{3001},
		Volumes: []string{"uptime_kuma_data:/app/data"},
	}
	values := map[string]string{"DOMAIN": "status.example.com"}

	out, err := GenerateCompose(r, values, 3001)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := string(out)

	if !strings.Contains(s, "uptime-kuma:") {
		t.Fatal("expected service name uptime-kuma")
	}
	if !strings.Contains(s, "louislam/uptime-kuma:1.23.11") {
		t.Fatal("expected image in output")
	}
	if !strings.Contains(s, "127.0.0.1:3001:3001") {
		t.Fatal("expected port mapping 127.0.0.1:3001:3001")
	}
	if !strings.Contains(s, "uptime_kuma_data:/app/data") {
		t.Fatal("expected volume mount")
	}
}

func TestGenerateCompose_MultiService(t *testing.T) {
	r := &Recipe{
		Name:  "plausible",
		Image: "ghcr.io/plausible/community-edition:v2.0.0",
		Ports: []int{8000},
		Services: []Service{
			{
				Name:  "plausible_db",
				Image: "postgres:16-alpine",
				Environment: map[string]string{"POSTGRES_PASSWORD": "postgres"},
				Volumes: []string{"plausible_db:/var/lib/postgresql/data"},
			},
		},
		Environment: map[string]string{
			"BASE_URL": "https://${DOMAIN}",
		},
	}
	values := map[string]string{"DOMAIN": "analytics.example.com"}

	out, err := GenerateCompose(r, values, 8000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := string(out)

	if !strings.Contains(s, "plausible_db:") {
		t.Fatal("expected plausible_db service")
	}
	if !strings.Contains(s, "postgres:16-alpine") {
		t.Fatal("expected postgres image")
	}
}

func TestGenerateCompose_DifferentHostPort(t *testing.T) {
	r := &Recipe{
		Name:  "app",
		Image: "app:latest",
		Ports: []int{3000},
	}

	out, err := GenerateCompose(r, nil, 3005)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(out), "127.0.0.1:3005:3000") {
		t.Fatal("expected port mapping 127.0.0.1:3005:3000")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/recipe/ -v -run TestGenerateCompose`
Expected: FAIL — `GenerateCompose` not defined

**Step 3: Implement compose generation**

```go
// internal/recipe/compose.go
package recipe

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type composeFile struct {
	Services map[string]composeService `yaml:"services"`
	Volumes  map[string]interface{}    `yaml:"volumes,omitempty"`
}

type composeService struct {
	Image       string            `yaml:"image"`
	Ports       []string          `yaml:"ports,omitempty"`
	Volumes     []string          `yaml:"volumes,omitempty"`
	Environment map[string]string `yaml:"environment,omitempty"`
	DependsOn   []string          `yaml:"depends_on,omitempty"`
	Restart     string            `yaml:"restart"`
}

func GenerateCompose(r *Recipe, values map[string]string, hostPort int) ([]byte, error) {
	cf := composeFile{
		Services: make(map[string]composeService),
		Volumes:  make(map[string]interface{}),
	}

	// Primary service
	primary := composeService{
		Image:   r.Image,
		Restart: "unless-stopped",
	}

	if len(r.Ports) > 0 {
		primary.Ports = []string{fmt.Sprintf("127.0.0.1:%d:%d", hostPort, r.Ports[0])}
	}

	if len(r.Volumes) > 0 {
		primary.Volumes = r.Volumes
		for _, v := range r.Volumes {
			volName := strings.Split(v, ":")[0]
			cf.Volumes[volName] = nil
		}
	}

	if len(r.Environment) > 0 {
		env := make(map[string]string)
		for k, v := range r.Environment {
			env[k] = expandEnvValue(v, values)
		}
		primary.Environment = env
	}

	// Add env_file reference
	var dependsOn []string
	for _, svc := range r.Services {
		dependsOn = append(dependsOn, svc.Name)
	}
	if len(dependsOn) > 0 {
		primary.DependsOn = dependsOn
	}

	cf.Services[r.Name] = primary

	// Additional services
	for _, svc := range r.Services {
		s := composeService{
			Image:   svc.Image,
			Restart: "unless-stopped",
		}
		if len(svc.Environment) > 0 {
			s.Environment = svc.Environment
		}
		if len(svc.Volumes) > 0 {
			s.Volumes = svc.Volumes
			for _, v := range svc.Volumes {
				volName := strings.Split(v, ":")[0]
				cf.Volumes[volName] = nil
			}
		}
		cf.Services[svc.Name] = s
	}

	if len(cf.Volumes) == 0 {
		cf.Volumes = nil
	}

	data, err := yaml.Marshal(cf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate compose file: %w", err)
	}
	return data, nil
}

func expandEnvValue(value string, values map[string]string) string {
	result := value
	for k, v := range values {
		result = strings.ReplaceAll(result, "${"+k+"}", v)
	}
	return result
}
```

**Step 4: Run tests**

Run: `go test ./internal/recipe/ -v -run TestGenerateCompose`
Expected: All 3 tests PASS

**Step 5: Commit**

```bash
git add internal/recipe/compose.go internal/recipe/compose_test.go
git commit -m "feat: add docker-compose.yml generation from recipes"
```

---

### Task 9: Env Generation

**Files:**
- Create: `internal/recipe/env.go`
- Create: `internal/recipe/env_test.go`

**Step 1: Write failing test**

```go
// internal/recipe/env_test.go
package recipe

import (
	"strings"
	"testing"
)

func TestGenerateEnv(t *testing.T) {
	values := map[string]string{
		"DOMAIN":     "example.com",
		"ADMIN_EMAIL": "admin@example.com",
	}

	out := GenerateEnv(values)
	s := string(out)

	if !strings.Contains(s, "DOMAIN=example.com") {
		t.Fatal("expected DOMAIN=example.com")
	}
	if !strings.Contains(s, "ADMIN_EMAIL=admin@example.com") {
		t.Fatal("expected ADMIN_EMAIL=admin@example.com")
	}
}

func TestExpandAutoGenerate(t *testing.T) {
	env := map[string]string{
		"KEY":     "auto_generate_32",
		"SECRET":  "auto_generate_64",
		"NORMAL":  "hello",
	}

	expanded := ExpandAutoGenerate(env)

	if len(expanded["KEY"]) != 32 {
		t.Fatalf("expected 32-char key, got %d", len(expanded["KEY"]))
	}
	if len(expanded["SECRET"]) != 64 {
		t.Fatalf("expected 64-char secret, got %d", len(expanded["SECRET"]))
	}
	if expanded["NORMAL"] != "hello" {
		t.Fatalf("expected 'hello', got %s", expanded["NORMAL"])
	}
}
```

**Step 2: Implement env generation**

```go
// internal/recipe/env.go
package recipe

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

func GenerateEnv(values map[string]string) []byte {
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var lines []string
	for _, k := range keys {
		lines = append(lines, fmt.Sprintf("%s=%s", k, values[k]))
	}
	return []byte(strings.Join(lines, "\n") + "\n")
}

func ExpandAutoGenerate(env map[string]string) map[string]string {
	result := make(map[string]string, len(env))
	for k, v := range env {
		switch v {
		case "auto_generate_32":
			result[k] = randomHex(16) // 16 bytes = 32 hex chars
		case "auto_generate_64":
			result[k] = randomHex(32) // 32 bytes = 64 hex chars
		default:
			result[k] = v
		}
	}
	return result
}

func randomHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}
```

**Step 3: Run tests**

Run: `go test ./internal/recipe/ -v -run "TestGenerateEnv|TestExpandAuto"`
Expected: All tests PASS

**Step 4: Commit**

```bash
git add internal/recipe/env.go internal/recipe/env_test.go
git commit -m "feat: add .env generation and auto_generate secret expansion"
```

---

### Task 10: Caddy Management

**Files:**
- Create: `internal/caddy/caddy.go`
- Create: `internal/caddy/caddy_test.go`

**Step 1: Write failing test**

```go
// internal/caddy/caddy_test.go
package caddy

import (
	"context"
	"strings"
	"testing"

	"github.com/pankajbeniwal/bunkr/internal/executor"
)

func TestAddBlock(t *testing.T) {
	mock := executor.NewMockExecutor()
	mock.Files[CaddyfilePath] = []byte("# Existing config\n")

	err := AddBlock(context.Background(), mock, "uptime-kuma", "status.example.com", 3001)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := string(mock.Files[CaddyfilePath])
	if !strings.Contains(content, "# bunkr:uptime-kuma") {
		t.Fatal("expected bunkr marker")
	}
	if !strings.Contains(content, "status.example.com") {
		t.Fatal("expected domain")
	}
	if !strings.Contains(content, "reverse_proxy localhost:3001") {
		t.Fatal("expected reverse_proxy directive")
	}
	if !strings.Contains(content, "# /bunkr:uptime-kuma") {
		t.Fatal("expected closing marker")
	}
}

func TestRemoveBlock(t *testing.T) {
	mock := executor.NewMockExecutor()
	mock.Files[CaddyfilePath] = []byte(`# Existing
# bunkr:ghost
blog.example.com {
    reverse_proxy localhost:2368
}
# /bunkr:ghost
# Other config
`)

	err := RemoveBlock(context.Background(), mock, "ghost")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := string(mock.Files[CaddyfilePath])
	if strings.Contains(content, "bunkr:ghost") {
		t.Fatal("expected ghost block to be removed")
	}
	if !strings.Contains(content, "# Existing") {
		t.Fatal("expected other config to remain")
	}
	if !strings.Contains(content, "# Other config") {
		t.Fatal("expected trailing config to remain")
	}
}

func TestAddBlock_NoCaddyfile(t *testing.T) {
	mock := executor.NewMockExecutor()
	// No existing Caddyfile — ReadFile will return error

	err := AddBlock(context.Background(), mock, "ghost", "blog.example.com", 2368)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := string(mock.Files[CaddyfilePath])
	if !strings.Contains(content, "blog.example.com") {
		t.Fatal("expected domain in new Caddyfile")
	}
}
```

**Step 2: Implement Caddy management**

```go
// internal/caddy/caddy.go
package caddy

import (
	"context"
	"fmt"
	"strings"

	"github.com/pankajbeniwal/bunkr/internal/executor"
)

const CaddyfilePath = "/etc/caddy/Caddyfile"

func AddBlock(ctx context.Context, exec executor.Executor, name string, domain string, hostPort int) error {
	existing, err := exec.ReadFile(ctx, CaddyfilePath)
	if err != nil {
		existing = []byte{}
	}

	block := fmt.Sprintf(
		"\n# bunkr:%s\n%s {\n    reverse_proxy localhost:%d\n}\n# /bunkr:%s\n",
		name, domain, hostPort, name,
	)

	content := string(existing) + block
	return exec.WriteFile(ctx, CaddyfilePath, []byte(content), 0644)
}

func RemoveBlock(ctx context.Context, exec executor.Executor, name string) error {
	data, err := exec.ReadFile(ctx, CaddyfilePath)
	if err != nil {
		return fmt.Errorf("failed to read Caddyfile: %w", err)
	}

	startMarker := fmt.Sprintf("# bunkr:%s", name)
	endMarker := fmt.Sprintf("# /bunkr:%s", name)

	lines := strings.Split(string(data), "\n")
	var result []string
	inside := false
	for _, line := range lines {
		if strings.TrimSpace(line) == startMarker {
			inside = true
			continue
		}
		if strings.TrimSpace(line) == endMarker {
			inside = false
			continue
		}
		if !inside {
			result = append(result, line)
		}
	}

	return exec.WriteFile(ctx, CaddyfilePath, []byte(strings.Join(result, "\n")), 0644)
}

func Reload(ctx context.Context, exec executor.Executor) error {
	_, err := exec.Run(ctx, "caddy reload --config /etc/caddy/Caddyfile")
	return err
}

func EnsureInstalled(ctx context.Context, exec executor.Executor) error {
	_, err := exec.Run(ctx, "which caddy")
	if err == nil {
		return nil
	}

	commands := []string{
		"apt-get install -y debian-keyring debian-archive-keyring apt-transport-https curl",
		"curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg",
		"curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list",
		"apt-get update",
		"apt-get install -y caddy",
	}

	for _, cmd := range commands {
		if _, err := exec.Run(ctx, cmd); err != nil {
			return fmt.Errorf("failed to install Caddy: %w", err)
		}
	}
	return nil
}
```

**Step 3: Run tests**

Run: `go test ./internal/caddy/ -v`
Expected: All 3 tests PASS

**Step 4: Commit**

```bash
git add internal/caddy/
git commit -m "feat: add Caddy management (add/remove blocks, install, reload)"
```

---

### Task 11: Docker Operations

**Files:**
- Create: `internal/docker/docker.go`
- Create: `internal/docker/docker_test.go`

**Step 1: Write failing test**

```go
// internal/docker/docker_test.go
package docker

import (
	"context"
	"testing"

	"github.com/pankajbeniwal/bunkr/internal/executor"
)

func TestEnsureInstalled_AlreadyPresent(t *testing.T) {
	mock := executor.NewMockExecutor()
	mock.RunOutputs["docker --version"] = "Docker version 24.0.7"

	err := EnsureInstalled(context.Background(), mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should only call docker --version, not install
	if len(mock.Calls) != 1 {
		t.Fatalf("expected 1 call (version check), got %d", len(mock.Calls))
	}
}

func TestComposeUp(t *testing.T) {
	mock := executor.NewMockExecutor()
	ctx := context.Background()

	err := ComposeUp(ctx, mock, "ghost")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.Calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(mock.Calls))
	}
	cmd := mock.Calls[0].Args[0].(string)
	if cmd != "docker compose -f /opt/bunkr/ghost/docker-compose.yml up -d" {
		t.Fatalf("unexpected command: %s", cmd)
	}
}

func TestComposeDown(t *testing.T) {
	mock := executor.NewMockExecutor()

	err := ComposeDown(context.Background(), mock, "ghost", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cmd := mock.Calls[0].Args[0].(string)
	if cmd != "docker compose -f /opt/bunkr/ghost/docker-compose.yml down" {
		t.Fatalf("unexpected command: %s", cmd)
	}
}

func TestComposeDown_Purge(t *testing.T) {
	mock := executor.NewMockExecutor()

	err := ComposeDown(context.Background(), mock, "ghost", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cmd := mock.Calls[0].Args[0].(string)
	if cmd != "docker compose -f /opt/bunkr/ghost/docker-compose.yml down -v" {
		t.Fatalf("unexpected command: %s", cmd)
	}
}
```

**Step 2: Implement docker operations**

```go
// internal/docker/docker.go
package docker

import (
	"context"
	"fmt"
	"strings"

	"github.com/pankajbeniwal/bunkr/internal/executor"
)

const basePath = "/opt/bunkr"

func composePath(recipe string) string {
	return fmt.Sprintf("%s/%s/docker-compose.yml", basePath, recipe)
}

func EnsureInstalled(ctx context.Context, exec executor.Executor) error {
	_, err := exec.Run(ctx, "docker --version")
	if err == nil {
		return nil
	}
	_, err = exec.Run(ctx, "curl -fsSL https://get.docker.com | sh")
	if err != nil {
		return fmt.Errorf("failed to install Docker: %w", err)
	}
	return nil
}

func ComposeUp(ctx context.Context, exec executor.Executor, recipe string) error {
	cmd := fmt.Sprintf("docker compose -f %s up -d", composePath(recipe))
	_, err := exec.Run(ctx, cmd)
	return err
}

func ComposeDown(ctx context.Context, exec executor.Executor, recipe string, purge bool) error {
	cmd := fmt.Sprintf("docker compose -f %s down", composePath(recipe))
	if purge {
		cmd += " -v"
	}
	_, err := exec.Run(ctx, cmd)
	return err
}

func ComposePull(ctx context.Context, exec executor.Executor, recipe string) error {
	cmd := fmt.Sprintf("docker compose -f %s pull", composePath(recipe))
	_, err := exec.Run(ctx, cmd)
	return err
}

type ServiceStatus struct {
	Name   string
	Status string // "running" or "exited"
}

func ComposeStatus(ctx context.Context, exec executor.Executor, recipe string) ([]ServiceStatus, error) {
	cmd := fmt.Sprintf("docker compose -f %s ps --format '{{.Name}} {{.State}}'", composePath(recipe))
	out, err := exec.Run(ctx, cmd)
	if err != nil {
		return nil, err
	}

	var statuses []ServiceStatus
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			statuses = append(statuses, ServiceStatus{Name: parts[0], Status: parts[1]})
		}
	}
	return statuses, nil
}

func HealthCheck(ctx context.Context, exec executor.Executor, url string, timeout, interval int) error {
	cmd := fmt.Sprintf(
		"for i in $(seq 1 %d); do curl -sf %s > /dev/null 2>&1 && exit 0; sleep %d; done; exit 1",
		timeout/interval, url, interval,
	)
	_, err := exec.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("health check failed after %ds: %w", timeout, err)
	}
	return nil
}
```

**Step 3: Run tests**

Run: `go test ./internal/docker/ -v`
Expected: All 4 tests PASS

**Step 4: Commit**

```bash
git add internal/docker/
git commit -m "feat: add Docker operations (install, compose up/down/pull, health check)"
```

---

### Task 12: Hardening Orchestrator + Steps

**Files:**
- Create: `internal/hardening/hardening.go`
- Create: `internal/hardening/user.go`
- Create: `internal/hardening/ssh.go`
- Create: `internal/hardening/firewall.go`
- Create: `internal/hardening/fail2ban.go`
- Create: `internal/hardening/upgrades.go`
- Create: `internal/hardening/sysctl.go`
- Create: `internal/hardening/swap.go`
- Create: `internal/hardening/hardening_test.go`

**Step 1: Write failing test for orchestrator**

```go
// internal/hardening/hardening_test.go
package hardening

import (
	"context"
	"fmt"
	"testing"

	"github.com/pankajbeniwal/bunkr/internal/executor"
	"github.com/pankajbeniwal/bunkr/internal/state"
)

func TestRunSteps_AllNew(t *testing.T) {
	mock := executor.NewMockExecutor()
	s := state.New()
	ctx := context.Background()

	// Make all checks return "not applied"
	mock.RunErrors["id bunkr"] = fmt.Errorf("no such user")
	mock.RunErrors["grep -q 'PermitRootLogin no' /etc/ssh/sshd_config"] = fmt.Errorf("not found")
	mock.RunErrors["ufw status | grep -q 'Status: active'"] = fmt.Errorf("inactive")
	mock.RunErrors["systemctl is-active fail2ban"] = fmt.Errorf("inactive")
	mock.RunErrors["dpkg -l | grep -q unattended-upgrades"] = fmt.Errorf("not installed")
	mock.RunErrors["test -f /etc/sysctl.d/99-bunkr.conf"] = fmt.Errorf("not found")
	mock.RunErrors["swapon --show | grep -q /"] = fmt.Errorf("no swap")

	results, err := Run(ctx, mock, s, 2222)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 7 {
		t.Fatalf("expected 7 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Skipped {
			t.Fatalf("expected step %s to not be skipped", r.Name)
		}
	}
}

func TestRunSteps_AllSkipped(t *testing.T) {
	mock := executor.NewMockExecutor()
	s := state.New()
	s.Hardening.Applied = true
	for _, step := range Steps(2222) {
		s.Hardening.Steps[step.Name] = true
	}

	results, err := Run(context.Background(), mock, s, 2222)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, r := range results {
		if !r.Skipped {
			t.Fatalf("expected step %s to be skipped", r.Name)
		}
	}
}
```

**Step 2: Implement the orchestrator and all steps**

Due to the size, I'll provide the orchestrator and step signatures. Each step file follows the same pattern.

```go
// internal/hardening/hardening.go
package hardening

import (
	"context"

	"github.com/pankajbeniwal/bunkr/internal/executor"
	"github.com/pankajbeniwal/bunkr/internal/state"
	"github.com/pankajbeniwal/bunkr/internal/ui"
)

type Step struct {
	Name  string
	Label string
	Check func(ctx context.Context, exec executor.Executor) (bool, error)
	Apply func(ctx context.Context, exec executor.Executor) error
}

type StepResult struct {
	Name    string
	Skipped bool
	Error   error
}

func Steps(sshPort int) []Step {
	return []Step{
		UserStep(),
		SSHStep(sshPort),
		FirewallStep(sshPort),
		Fail2banStep(),
		UpgradesStep(),
		SysctlStep(),
		SwapStep(),
	}
}

func Run(ctx context.Context, exec executor.Executor, s *state.State, sshPort int) ([]StepResult, error) {
	steps := Steps(sshPort)
	var results []StepResult

	for _, step := range steps {
		if s.Hardening.Steps[step.Name] {
			ui.Skip(step.Label + " already configured")
			results = append(results, StepResult{Name: step.Name, Skipped: true})
			continue
		}

		applied, err := step.Check(ctx, exec)
		if err == nil && applied {
			ui.Skip(step.Label + " already configured")
			s.Hardening.Steps[step.Name] = true
			results = append(results, StepResult{Name: step.Name, Skipped: true})
			continue
		}

		if err := step.Apply(ctx, exec); err != nil {
			ui.Error(step.Label + " failed: " + err.Error())
			results = append(results, StepResult{Name: step.Name, Error: err})
			return results, err
		}

		ui.Success(step.Label)
		s.Hardening.Steps[step.Name] = true
		results = append(results, StepResult{Name: step.Name})
	}

	s.Hardening.Applied = true
	return results, nil
}
```

Each step file (`user.go`, `ssh.go`, `firewall.go`, `fail2ban.go`, `upgrades.go`, `sysctl.go`, `swap.go`) returns a `Step` struct. Example for `user.go`:

```go
// internal/hardening/user.go
package hardening

import (
	"context"

	"github.com/pankajbeniwal/bunkr/internal/executor"
)

func UserStep() Step {
	return Step{
		Name:  "sudo_user",
		Label: "Sudo user created",
		Check: func(ctx context.Context, exec executor.Executor) (bool, error) {
			_, err := exec.Run(ctx, "id bunkr")
			return err == nil, err
		},
		Apply: func(ctx context.Context, exec executor.Executor) error {
			cmds := []string{
				"adduser --disabled-password --gecos '' bunkr",
				"usermod -aG sudo bunkr",
				"echo 'bunkr ALL=(ALL) NOPASSWD:ALL' > /etc/sudoers.d/bunkr",
				"mkdir -p /home/bunkr/.ssh",
				"cp /root/.ssh/authorized_keys /home/bunkr/.ssh/authorized_keys",
				"chown -R bunkr:bunkr /home/bunkr/.ssh",
				"chmod 700 /home/bunkr/.ssh",
				"chmod 600 /home/bunkr/.ssh/authorized_keys",
			}
			for _, cmd := range cmds {
				if _, err := exec.Run(ctx, cmd); err != nil {
					return err
				}
			}
			// Verify the user works
			if _, err := exec.Run(ctx, "su - bunkr -c 'whoami'"); err != nil {
				return err
			}
			return nil
		},
	}
}
```

Follow the same pattern for all other steps. Each checks its condition and applies the necessary commands.

**Step 3: Run tests**

Run: `go test ./internal/hardening/ -v`
Expected: All tests PASS

**Step 4: Commit**

```bash
git add internal/hardening/
git commit -m "feat: add hardening orchestrator and all 7 hardening steps"
```

---

### Task 13: Interactive Prompts

**Files:**
- Create: `internal/recipe/prompt.go`

**Step 1: Implement prompt handling**

```go
// internal/recipe/prompt.go
package recipe

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

func PromptUser(prompts []Prompt) (map[string]string, error) {
	reader := bufio.NewReader(os.Stdin)
	values := make(map[string]string)

	for _, p := range prompts {
		var value string
		var err error

		if p.Secret {
			value, err = promptSecret(p)
		} else {
			value, err = promptText(reader, p)
		}
		if err != nil {
			return nil, err
		}

		if value == "" && p.Default != "" {
			value = p.Default
		}

		if value == "" && p.Required {
			return nil, fmt.Errorf("%s is required", p.Key)
		}

		values[p.Key] = value
	}

	return values, nil
}

func promptText(reader *bufio.Reader, p Prompt) (string, error) {
	label := p.Label
	if p.Default != "" {
		label = fmt.Sprintf("%s [%s]", label, p.Default)
	}
	fmt.Printf("  ? %s: ", label)

	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

func promptSecret(p Prompt) (string, error) {
	fmt.Printf("  ? %s: ", p.Label)
	data, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println() // newline after hidden input
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}
```

**Step 2: Verify it compiles**

Run: `go build ./internal/recipe/`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/recipe/prompt.go
git commit -m "feat: add interactive prompt handling for recipe config"
```

---

### Task 14: CLI Commands — init and install

**Files:**
- Create: `cmd/bunkr/init.go`
- Create: `cmd/bunkr/install.go`

**Step 1: Implement init command**

```go
// cmd/bunkr/init.go
package main

import (
	"context"
	"os"
	"time"

	"github.com/pankajbeniwal/bunkr/internal/executor"
	"github.com/pankajbeniwal/bunkr/internal/hardening"
	"github.com/pankajbeniwal/bunkr/internal/state"
	"github.com/pankajbeniwal/bunkr/internal/ui"
	"github.com/spf13/cobra"
)

var sshPortFlag int

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Harden the server (no app install)",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		exec, err := newExecutor()
		if err != nil {
			return err
		}

		ui.Header("Hardening VPS...")

		s, err := state.Load(ctx, exec)
		if err != nil {
			return err
		}

		_, err = hardening.Run(ctx, exec, s, sshPortFlag)
		if err != nil {
			return err
		}

		s.Hardening.AppliedAt = time.Now()
		s.Hardening.SSHPort = sshPortFlag

		if err := state.Save(ctx, exec, s); err != nil {
			return err
		}

		ui.Result("Server hardened successfully!")
		return nil
	},
}

func init() {
	initCmd.Flags().IntVar(&sshPortFlag, "ssh-port", 2222, "SSH port to configure")
	rootCmd.AddCommand(initCmd)
}

func newExecutor() (executor.Executor, error) {
	if onFlag != "" {
		return executor.NewRemoteExecutor(onFlag)
	}
	return executor.NewLocalExecutor(), nil
}
```

**Step 2: Implement install command**

```go
// cmd/bunkr/install.go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/pankajbeniwal/bunkr/internal/caddy"
	"github.com/pankajbeniwal/bunkr/internal/docker"
	"github.com/pankajbeniwal/bunkr/internal/hardening"
	"github.com/pankajbeniwal/bunkr/internal/recipe"
	"github.com/pankajbeniwal/bunkr/internal/state"
	"github.com/pankajbeniwal/bunkr/internal/ui"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install <recipe> [<recipe>...]",
	Short: "Harden server and install app(s)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// === PLAN PHASE (always local) ===

		// Fetch and validate all recipes
		type planned struct {
			recipe  *recipe.Recipe
			values  map[string]string
		}
		var plans []planned

		for _, name := range args {
			ui.Header(fmt.Sprintf("Fetching %s...", name))
			r, err := recipe.Fetch(name)
			if err != nil {
				return fmt.Errorf("failed to fetch recipe %s: %w", name, err)
			}

			ui.Header(fmt.Sprintf("Configuring %s...", name))
			values, err := recipe.PromptUser(r.Prompts)
			if err != nil {
				return err
			}

			// Expand auto_generate values in environment
			if r.Environment != nil {
				r.Environment = recipe.ExpandAutoGenerate(r.Environment)
			}

			plans = append(plans, planned{recipe: r, values: values})
		}

		// === EXECUTE PHASE (via executor) ===

		exec, err := newExecutor()
		if err != nil {
			return err
		}

		s, err := state.Load(ctx, exec)
		if err != nil {
			return err
		}

		// Hardening
		if !s.Hardening.Applied {
			ui.Header("Hardening VPS...")
			if _, err := hardening.Run(ctx, exec, s, sshPortFlag); err != nil {
				return err
			}
			s.Hardening.AppliedAt = time.Now()
			s.Hardening.SSHPort = sshPortFlag
		}

		// Docker
		ui.Info("Checking Docker...")
		if err := docker.EnsureInstalled(ctx, exec); err != nil {
			return err
		}
		ui.Success("Docker ready")

		// Caddy
		ui.Info("Checking Caddy...")
		if err := caddy.EnsureInstalled(ctx, exec); err != nil {
			return err
		}
		ui.Success("Caddy ready")

		// Install each recipe
		for _, p := range plans {
			r := p.recipe
			ui.Header(fmt.Sprintf("Installing %s...", r.Name))

			hostPort := s.AllocatePort(r.Ports[0])

			// Generate files
			composeData, err := recipe.GenerateCompose(r, p.values, hostPort)
			if err != nil {
				return err
			}
			envData := recipe.GenerateEnv(p.values)

			// Create directory
			dir := fmt.Sprintf("/opt/bunkr/%s", r.Name)
			if _, err := exec.Run(ctx, fmt.Sprintf("mkdir -p %s", dir)); err != nil {
				return err
			}

			// Write files
			if err := exec.WriteFile(ctx, dir+"/docker-compose.yml", composeData, 0644); err != nil {
				return err
			}
			ui.Success("Compose file generated")

			if err := exec.WriteFile(ctx, dir+"/.env", envData, 0600); err != nil {
				return err
			}

			// Caddy
			domain := p.values["DOMAIN"]
			if err := caddy.AddBlock(ctx, exec, r.Name, domain, hostPort); err != nil {
				return err
			}
			ui.Success("Caddy configured")

			// Start containers
			if err := docker.ComposeUp(ctx, exec, r.Name); err != nil {
				ui.Error("Failed to start containers")
				ui.Info(fmt.Sprintf("  Run: docker compose -f %s/docker-compose.yml logs", dir))
				return err
			}
			ui.Success("Containers started")

			// Health check
			if r.HealthCheck != nil {
				if err := docker.HealthCheck(ctx, exec, r.HealthCheck.URL, r.HealthCheck.Timeout, r.HealthCheck.Interval); err != nil {
					ui.Warn("Health check failed — app may still be starting")
				} else {
					ui.Success("Health check passed")
				}
			}

			// Update state
			s.Recipes[r.Name] = state.RecipeState{
				Version:       r.Version,
				Domain:        domain,
				InstalledAt:   time.Now(),
				Port:          hostPort,
				ContainerPort: r.Ports[0],
			}
		}

		// Reload Caddy once
		if err := caddy.Reload(ctx, exec); err != nil {
			ui.Warn("Caddy reload failed — you may need to run 'caddy reload' manually")
		}

		// Save state
		if err := state.Save(ctx, exec, s); err != nil {
			return err
		}

		// Print results
		for _, p := range plans {
			domain := p.values["DOMAIN"]
			ui.Result(fmt.Sprintf("%s is running at https://%s", p.recipe.Name, domain))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
```

**Step 3: Verify it compiles**

Run: `go build ./cmd/bunkr/`
Expected: Compile error for `NewRemoteExecutor` (not implemented yet) — stub it.

Create a stub in `internal/executor/remote.go`:

```go
// internal/executor/remote.go
package executor

import "fmt"

type RemoteExecutor struct{}

func NewRemoteExecutor(target string) (*RemoteExecutor, error) {
	return nil, fmt.Errorf("remote execution not yet implemented (target: %s)", target)
}
```

Run: `go build ./cmd/bunkr/`
Expected: Success

**Step 4: Commit**

```bash
git add cmd/bunkr/init.go cmd/bunkr/install.go internal/executor/remote.go
git commit -m "feat: add init and install CLI commands"
```

---

### Task 15: CLI Commands — uninstall, list, status, update, self-update

**Files:**
- Create: `cmd/bunkr/uninstall.go`
- Create: `cmd/bunkr/list.go`
- Create: `cmd/bunkr/status.go`
- Create: `cmd/bunkr/update.go`
- Create: `cmd/bunkr/selfupdate.go`

These follow the same pattern as init/install but simpler. Each command:
- Creates executor
- Loads state
- Performs its operation
- Saves state

I'll provide the key implementation for each. See the design doc section "Command Flows" for the exact logic each should follow.

**Step 1: Implement all 5 remaining commands**

Each file registers its command via `init()` calling `rootCmd.AddCommand()`.

Key points:
- `uninstall`: uses `--purge` flag, confirms before purge, calls `docker.ComposeDown`, `caddy.RemoveBlock`, removes directory
- `list`: calls `recipe.FetchIndex()`, prints table with `fmt.Printf` using aligned columns
- `status`: loads state, calls `docker.ComposeStatus` per recipe, prints table
- `update`: fetches latest recipe, compares version, if newer pulls and recreates
- `selfupdate`: calls GitHub API `https://api.github.com/repos/<owner>/<repo>/releases/latest`, downloads binary

**Step 2: Verify compilation**

Run: `go build ./cmd/bunkr/`
Expected: Success

**Step 3: Commit**

```bash
git add cmd/bunkr/uninstall.go cmd/bunkr/list.go cmd/bunkr/status.go cmd/bunkr/update.go cmd/bunkr/selfupdate.go
git commit -m "feat: add uninstall, list, status, update, and self-update commands"
```

---

### Task 16: RemoteExecutor Implementation

**Files:**
- Modify: `internal/executor/remote.go`
- Create: `internal/executor/remote_test.go`

**Step 1: Implement RemoteExecutor**

Replace the stub with full implementation using `golang.org/x/crypto/ssh`. The executor:
- Parses `user@host:port` or SSH config alias from the `--on` flag
- Connects using SSH agent or default key files (`~/.ssh/id_rsa`, `~/.ssh/id_ed25519`)
- `Run()` executes commands via `session.CombinedOutput()`
- `WriteFile()` writes via `cat > path` over SSH stdin
- `ReadFile()` reads via `cat path` over SSH

Key: Parse the target string to extract user, host, port. If no `@`, treat as SSH config alias and use `ssh_config` to resolve.

**Step 2: Write unit test for target parsing**

```go
func TestParseTarget(t *testing.T) {
	tests := []struct {
		input string
		user  string
		host  string
	}{
		{"root@167.71.50.23", "root", "167.71.50.23"},
		{"bunkr@example.com", "bunkr", "example.com"},
	}
	for _, tt := range tests {
		user, host := parseTarget(tt.input)
		if user != tt.user || host != tt.host {
			t.Errorf("parseTarget(%q) = %q, %q; want %q, %q", tt.input, user, host, tt.user, tt.host)
		}
	}
}
```

**Step 3: Run tests**

Run: `go test ./internal/executor/ -v -run TestParseTarget`
Expected: PASS

**Step 4: Commit**

```bash
git add internal/executor/remote.go internal/executor/remote_test.go
git commit -m "feat: implement RemoteExecutor with SSH support"
```

---

### Task 17: Recipe YAML Files

**Files:**
- Create: `recipes/index.yaml`
- Create: `recipes/uptime-kuma.yaml`
- Create: `recipes/ghost.yaml`
- Create: `recipes/plausible.yaml`

**Step 1: Create all recipe files**

Use the exact YAML from the design doc. The `index.yaml` lists all three:

```yaml
- name: uptime-kuma
  description: Self-hosted uptime monitoring
  version: "1.23.11"
- name: ghost
  description: Professional publishing platform
  version: "5.82.2"
- name: plausible
  description: Privacy-friendly web analytics
  version: "2.0.0"
```

**Step 2: Commit**

```bash
git add recipes/
git commit -m "feat: add initial recipe YAML files (uptime-kuma, ghost, plausible)"
```

---

### Task 18: Install Script

**Files:**
- Create: `scripts/install.sh`

**Step 1: Write the install script**

Detects Linux + arch, downloads from GitHub Releases, places at `/usr/local/bin/bunkr`, creates `/etc/bunkr/`.

**Step 2: Commit**

```bash
git add scripts/
git commit -m "feat: add curl | sh install script"
```

---

### Task 19: GoReleaser + GitHub Actions

**Files:**
- Create: `.goreleaser.yaml`
- Create: `.github/workflows/release.yaml`

**Step 1: Create GoReleaser config**

Cross-compile for `linux/amd64` and `linux/arm64`, set ldflags for version.

**Step 2: Create GitHub Actions workflow**

Trigger on tag push, run GoReleaser.

**Step 3: Commit**

```bash
git add .goreleaser.yaml .github/
git commit -m "feat: add GoReleaser config and release GitHub Action"
```

---

### Task 20: Final — Run All Tests + Verify Build

**Step 1: Run full test suite**

Run: `go test ./... -v`
Expected: All tests PASS

**Step 2: Build binary**

Run: `go build -o bunkr ./cmd/bunkr/`
Expected: Binary created

**Step 3: Smoke test**

Run: `./bunkr version`
Expected: `bunkr dev`

Run: `./bunkr --help`
Expected: Shows all commands (init, install, uninstall, list, status, update, self-update, version)

**Step 4: Commit any final fixes**

```bash
git commit -m "chore: final cleanup and test fixes"
```
