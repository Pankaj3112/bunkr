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
	if len(mock.Calls) != 2 {
		t.Fatalf("expected 2 calls (pull + up), got %d", len(mock.Calls))
	}
	pullCmd := mock.Calls[0].Args[0].(string)
	if pullCmd != "docker compose -f /opt/bunkr/ghost/docker-compose.yml pull 2>&1" {
		t.Fatalf("unexpected pull command: %s", pullCmd)
	}
	upCmd := mock.Calls[1].Args[0].(string)
	if upCmd != "docker compose -f /opt/bunkr/ghost/docker-compose.yml up -d" {
		t.Fatalf("unexpected up command: %s", upCmd)
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
