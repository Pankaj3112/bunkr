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
